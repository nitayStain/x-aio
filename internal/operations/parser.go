package operations

import (
	"bytes"
	"regexp"
)

// GetOperations retrieves and parses all GraphQL operations from x.com's main script.
func GetOperations() ([]Operation, error) {
	mainPageContent, err := getMainPage()
	if err != nil {
		return nil, err
	}

	mainScriptContent, err := getMainScript(mainPageContent)
	if err != nil {
		return nil, err
	}

	content := []byte(mainScriptContent)
	re := regexp.MustCompile(`{queryId:"([^"]+)",operationName:"([^"]+)",operationType:"([^"]+)",metadata:`)
	matches := re.FindAllIndex(content, -1)

	var ops []Operation

	for _, m := range matches {
		start := m[0]
		section := content[start:]
		metaStart := bytes.Index(section, []byte("metadata:")) + len("metadata:")
		metaStr, consumed := extractBalancedBraces(section[metaStart:])
		if consumed == 0 {
			continue
		}

		reHeader := regexp.MustCompile(`{queryId:"([^"]+)",operationName:"([^"]+)",operationType:"([^"]+)",`)
		fields := reHeader.FindSubmatch(section)
		if len(fields) < 4 {
			continue
		}

		meta := parseMetadata(metaStr)
		ops = append(ops, Operation{
			QueryID:         string(fields[1]),
			OperationName:   string(fields[2]),
			OperationType:   string(fields[3]),
			FeatureSwitches: meta.FeatureSwitches,
			FieldToggles:    meta.FieldToggles,
		})
	}

	return ops, nil
}

// extractBalancedBraces returns a full {...} block and how many bytes it consumed.
func extractBalancedBraces(data []byte) (string, int) {
	if len(data) == 0 || data[0] != '{' {
		return "", 0
	}
	depth := 0
	for i := 0; i < len(data); i++ {
		switch data[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return string(data[:i+1]), i + 1
			}
		}
	}
	return "", 0
}
