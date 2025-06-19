package operations

import (
	"fmt"
	"regexp"

	"github.com/nitayStain/x-aio/internal/utils"
)

const baseURL = "https://x.com"

// This function simply retrieves the content of the main X page.
func getMainPage() (string, error) {
	return utils.GetPageContent(baseURL)
}

/*
This function uses regex to extract the main.js script from the x.com homepage.
The main.js contains the GraphQL operations that X uses in their GraphQL api
*/
func getMainScriptHref(html string) (string, error) {
	re := regexp.MustCompile(`<link[^>]+rel=["']preload["'][^>]+as=["']script["'][^>]+href=["']([^"']*main\.[^"']*\.js)["']`)
	match := re.FindStringSubmatch(html)
	if len(match) < 2 {
		return "", fmt.Errorf("main script not found")
	}
	return match[1], nil
}

// This function retrieves the content of the main.js script
func getMainScript(html string) (string, error) {
	mainScriptUrl, err := getMainScriptHref(html)
	if err != nil {
		return "", err
	}

	return utils.GetPageContent(mainScriptUrl)
}
