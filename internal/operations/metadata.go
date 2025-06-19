package operations

import (
	"regexp"
	"strings"
)

type metadataRaw struct {
	FeatureSwitches []string
	FieldToggles    []string
}

// parseMetadata extracts featureSwitches and fieldToggles from a metadata string.
func parseMetadata(meta string) metadataRaw {
	var m metadataRaw

	reFS := regexp.MustCompile(`featureSwitches:\[([^\]]*)\]`)
	fs := reFS.FindStringSubmatch(meta)
	if len(fs) == 2 {
		m.FeatureSwitches = splitList(fs[1])
	}

	reFT := regexp.MustCompile(`fieldToggles:\[([^\]]*)\]`)
	ft := reFT.FindStringSubmatch(meta)
	if len(ft) == 2 {
		m.FieldToggles = splitList(ft[1])
	}

	return m
}

// splitList splits a comma-separated string of quoted items into a string slice.
func splitList(s string) []string {
	parts := strings.Split(s, ",")
	var res []string
	for _, p := range parts {
		p = strings.Trim(p, `"`)
		if p != "" {
			res = append(res, p)
		}
	}
	return res
}
