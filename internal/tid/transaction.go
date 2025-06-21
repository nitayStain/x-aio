package tid

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	onDemandRegex = regexp.MustCompile(`['|"]ondemand\\.s['|"]:\s*['|"]([\\w]*)['|"]`)
	indicesRegex  = regexp.MustCompile(`\(\w{1}\[(\d{1,2})\],\s*16\)`)
)

type ClientTransaction struct {
	AdditionalRandomNumber byte
	DefaultKeyword         string
	KeyBytes               []byte
	AnimationKey           string
}

func NewClientTransaction(client *http.Client) (*ClientTransaction, error) {
	homePage, err := handleXMigration(client)
	if err != nil {
		return nil, err
	}

	rowIndex, keyByteIndices, err := getIndices(homePage, client)
	if err != nil {
		return nil, err
	}

	key, err := getKey(homePage)
	if err != nil {
		return nil, err
	}

	keyBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, err
	}

	animationKey, err := getAnimationKey(keyBytes, homePage, rowIndex, keyByteIndices)
	if err != nil {
		return nil, err
	}

	return &ClientTransaction{
		AdditionalRandomNumber: 3,
		DefaultKeyword:         "obfiowerehiring",
		KeyBytes:               keyBytes,
		AnimationKey:           animationKey,
	}, nil
}

func getIndices(doc *goquery.Document, client *http.Client) (int, []int, error) {
	html, err := doc.Html()
	if err != nil {
		return 0, nil, err
	}
	matches := onDemandRegex.FindStringSubmatch(html)
	if len(matches) < 2 {
		return 0, nil, errors.New("ondemand.js not found")
	}

	url := fmt.Sprintf("https://abs.twimg.com/responsive-web/client-web/ondemand.s.%sa.js", matches[1])
	resp, err := client.Get(url)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	indices := []int{}
	for _, match := range indicesRegex.FindAllStringSubmatch(string(body), -1) {
		if len(match) >= 2 {
			if idx, err := strconv.Atoi(match[1]); err == nil {
				indices = append(indices, idx)
			}
		}
	}

	if len(indices) < 2 {
		return 0, nil, errors.New("key byte indices missing")
	}

	return indices[0], indices[1:], nil
}

func getKey(doc *goquery.Document) (string, error) {
	meta := doc.Find(`meta[name="twitter-site-verification"]`).First()
	content, exists := meta.Attr("content")
	if !exists {
		return "", errors.New("twitter-site-verification meta tag not found")
	}
	return content, nil
}

func (c *ClientTransaction) GenerateTransactionID(method, path string) (string, error) {
	now := uint32(time.Now().Unix()) - 1682924400
	timeNowBytes := []byte{
		byte(now & 0xFF),
		byte((now >> 8) & 0xFF),
		byte((now >> 16) & 0xFF),
		byte((now >> 24) & 0xFF),
	}

	hashInput := fmt.Sprintf("%s!%s!%d%s%s", method, path, now, c.DefaultKeyword, c.AnimationKey)
	hash := sha256.Sum256([]byte(hashInput))

	hashBytes := hash[:16]
	randomByte := byte(rand.Intn(256))

	data := append([]byte{}, c.KeyBytes...)
	data = append(data, timeNowBytes...)
	data = append(data, hashBytes...)
	data = append(data, c.AdditionalRandomNumber)

	out := []byte{randomByte}
	for _, b := range data {
		out = append(out, b^randomByte)
	}

	encoded := base64.StdEncoding.EncodeToString(out)
	return strings.TrimRight(encoded, "="), nil
}

func getAnimationKey(keyBytes []byte, page *goquery.Document, rowIndex int, keyByteIndices []int) (string, error) {
	totalTime := 4096.0

	rowIndexValue := int(keyBytes[rowIndex] % 16)

	frameTime := 1.0
	for _, index := range keyByteIndices {
		frameTime *= float64(keyBytes[index] % 16)
	}

	frameTime = JsRound(frameTime/10.0) * 10.0
	targetTime := frameTime / totalTime

	arr, err := get2DArray(keyBytes, page, nil)
	if err != nil {
		return "", err
	}
	if rowIndexValue >= len(arr) {
		return "", errors.New("invalid row index")
	}

	frameRow := arr[rowIndexValue]
	animationKey := animate(frameRow, targetTime)

	return animationKey, nil
}

func getFrames(doc *goquery.Document) []*goquery.Selection {
	var frames []*goquery.Selection
	doc.Find("[id^='loading-x-anim']").Each(func(i int, s *goquery.Selection) {
		frames = append(frames, s)
	})
	return frames
}

// get2DArray extracts a 2D array of int from the SVG path data embedded in frames
func get2DArray(keyBytes []byte, page *goquery.Document, frames []*goquery.Selection) ([][]int, error) {
	if frames == nil {
		frames = getFrames(page)
	}

	frameIndex := int(keyBytes[5] % 4)
	if frameIndex >= len(frames) {
		return nil, errors.New("invalid frame index")
	}

	frame := frames[frameIndex]

	// Get first child element
	firstChild := frame.Children().First()
	if firstChild == nil {
		return nil, errors.New("no first child in frame")
	}

	// Get second child of firstChild
	secondChild := firstChild.Children().Eq(1)
	if secondChild == nil {
		return nil, errors.New("no second child in inner group")
	}

	dAttr, exists := secondChild.Attr("d")
	if !exists {
		return nil, errors.New("missing 'd' attribute")
	}

	if len(dAttr) < 10 { // get substring from 9 (0-based index 9 means 10th char)
		return nil, errors.New("path data too short")
	}
	dContent := dAttr[9:]

	segments := strings.Split(dContent, "C")

	var result [][]int
	nonDigitOrMinus := regexp.MustCompile(`[^0-9\-]+`)

	for _, segment := range segments {
		// Replace all non-digit and non-minus chars with space
		cleaned := nonDigitOrMinus.ReplaceAllString(segment, " ")
		fields := strings.Fields(cleaned)

		var numbers []int
		for _, f := range fields {
			n, err := strconv.Atoi(f)
			if err == nil {
				numbers = append(numbers, n)
			}
		}

		result = append(result, numbers)
	}

	return result, nil
}

// solve computes a mapped value with optional rounding
func solve(value, minVal, maxVal float64, rounding bool) float64 {
	result := value*(maxVal-minVal)/255.0 + minVal
	if rounding {
		return math.Floor(result)
	}
	return math.Round(result*100) / 100
}

// animate generates animation key string from frames and a target time
func animate(frames []int, targetTime float64) string {
	fromColor := []float64{float64(frames[0]), float64(frames[1]), float64(frames[2]), 1.0}
	toColor := []float64{float64(frames[3]), float64(frames[4]), float64(frames[5]), 1.0}
	fromRotation := []float64{0.0}
	toRotation := []float64{solve(float64(frames[6]), 60.0, 360.0, true)}

	// Calculate curves applying isOdd for min_val in solve
	curves := make([]float64, len(frames)-7)
	for i, val := range frames[7:] {
		curves[i] = solve(float64(val), float64(IsOdd(int32(i))), 1.0, false)
	}

	cubic := NewCubic(curves) // You need to implement this spline interpolation struct with GetValue method
	val := cubic.GetValue(targetTime)

	color, _ := Interpolate(fromColor, toColor, val) // You need interpolate implementation returning []float64
	for i := range color {
		if color[i] < 0 {
			color[i] = 0
		} else if color[i] > 255 {
			color[i] = 255
		}
	}

	rotation, _ := Interpolate(fromRotation, toRotation, val)
	matrix := ConvertRotationToMatrix(rotation[0]) // Implement matrix conversion based on rotation angle

	strArr := []string{}

	// Color values as hex (skip alpha)
	for _, v := range color[:len(color)-1] {
		strArr = append(strArr, strconv.FormatInt(int64(math.Round(v)), 16))
	}

	// Matrix values as hex, using floatToHex (implement floatToHex)
	for _, v := range matrix {
		rounded := math.Round(v*100) / 100
		absVal := math.Abs(rounded)
		hexVal := JsFloatToHex(absVal)

		if strings.HasPrefix(hexVal, ".") {
			strArr = append(strArr, "0"+strings.ToLower(hexVal))
		} else if hexVal == "" {
			strArr = append(strArr, "0")
		} else {
			strArr = append(strArr, strings.ToLower(hexVal))
		}
	}

	// Append final zeros
	strArr = append(strArr, "0", "0")

	animationKey := strings.Join(strArr, "")
	animationKey = strings.ReplaceAll(animationKey, ".", "")
	animationKey = strings.ReplaceAll(animationKey, "-", "")
	return animationKey
}
