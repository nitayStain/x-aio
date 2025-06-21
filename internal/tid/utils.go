package tid

import (
	"encoding/base64"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func handleXMigration(client *http.Client) (*goquery.Document, error) {
	migrationRegex := regexp.MustCompile(`https?://(?:www\.)?(twitter|x)\.com(/x)?/migrate([/?])?tok=[a-zA-Z0-9%\-_]+`)

	resp, err := client.Get("https://x.com")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check meta refresh
	meta := doc.Find(`meta[http-equiv="refresh"]`).First()
	if content, exists := meta.Attr("content"); exists {
		if loc := migrationRegex.FindString(content); loc != "" {
			res, err := client.Get(loc)
			if err != nil {
				return nil, err
			}
			defer res.Body.Close()
			return goquery.NewDocumentFromReader(res.Body)
		}
	}

	// Check migration form
	var form *goquery.Selection
	form = doc.Find(`form[name="f"]`).First()
	if form.Length() == 0 {
		form = doc.Find(`form[action="https://x.com/x/migrate"]`).First()
	}
	if form.Length() > 0 {
		action, _ := form.Attr("action")
		if action == "" {
			action = "https://x.com/x/migrate"
		}
		method := strings.ToUpper(strings.TrimSpace(getAttr(form, "method", "POST")))

		data := url.Values{}
		form.Find("input").Each(func(i int, s *goquery.Selection) {
			name, nameExists := s.Attr("name")
			value, valueExists := s.Attr("value")
			if nameExists && valueExists {
				data.Set(name, value)
			}
		})

		var res *http.Response
		if method == "POST" {
			res, err = client.PostForm(action, data)
		} else {
			reqURL := action + "?" + data.Encode()
			res, err = client.Get(reqURL)
		}
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		return goquery.NewDocumentFromReader(res.Body)
	}

	return doc, nil
}

// implementation of js' number rounding
func JsRound(num float64) float64 {
	dec := num - math.Trunc(num)
	if dec == -0.5 {
		return math.Ceil(num)
	}

	return math.Round(num)
}

// Some kind of weird implementation of a bool (used in twitter's reverse engineered code)
func IsOdd(num int32) float64 {
	if num%2 == 1 {
		return -1.0
	}

	return 0.0
}

// implementation of js' number float to hex
func JsFloatToHex(num float64) string {
	if num == 0.0 {
		return "0"
	}

	var result string
	quotient := int64(math.Floor(num))
	fraction := num - float64(quotient)

	if quotient == 0 {
		result += "0"
	} else {
		var intPart []rune
		for quotient > 0 {
			remainder := quotient % 16
			quotient /= 16
			intPart = append([]rune{parseDigit(remainder)}, intPart...)
		}
		result += string(intPart)
	}

	// Fractional part
	if fraction > 0.0 {
		result += "."
		loopLimit := 20
		for fraction > 0.0 && loopLimit > 0 {
			fraction *= 16.0
			integer := int64(math.Floor(fraction))
			fraction -= float64(integer)
			result += string(parseDigit(integer))
			loopLimit--
		}
	}

	return result
}

// Encode a byte slice to base64 string
func base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// Decode a base64 string to byte slice
func base64Decode(input string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(input)
}

// helper method to extract a digit
func parseDigit(value int64) rune {
	if value > 9 {
		return rune('A' + value - 10)
	} else {
		return rune('0' + value)
	}
}

func getAttr(sel *goquery.Selection, attr, fallback string) string {
	if val, exists := sel.Attr(attr); exists {
		return val
	}
	return fallback
}
