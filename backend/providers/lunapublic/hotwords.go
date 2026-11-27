package lunapublic

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Hotword struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

// ParseHotwords reads deployment-wide terminology in the public API's pair format.
func ParseHotwords(value string) ([]Hotword, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	var pairs []map[string]string
	if err := json.Unmarshal([]byte(value), &pairs); err != nil || pairs == nil {
		return nil, errors.New("expected a JSON array containing source/target objects")
	}
	if len(pairs) > 20 {
		return nil, errors.New("expected at most 20 source/target pairs")
	}
	hotwords := make([]Hotword, 0, len(pairs))
	seen := make(map[string]bool, len(pairs))
	for i, pair := range pairs {
		source, hasSource := pair["source"]
		target, hasTarget := pair["target"]
		if len(pair) != 2 || !hasSource || !hasTarget {
			return nil, fmt.Errorf("pair %d must contain only source and target", i+1)
		}
		source, target = strings.TrimSpace(source), strings.TrimSpace(target)
		if !validHotwordTerm(source) || !validHotwordTerm(target) {
			return nil, fmt.Errorf("pair %d needs terms of 1–50 characters without control characters or prompt markers", i+1)
		}
		if seen[source] {
			return nil, fmt.Errorf("pair %d repeats a source term", i+1)
		}
		seen[source] = true
		hotwords = append(hotwords, Hotword{Source: source, Target: target})
	}
	return hotwords, nil
}

func validHotwordTerm(term string) bool {
	return term != "" && utf8.RuneCountInString(term) <= 50 &&
		!strings.ContainsFunc(term, unicode.IsControl) && !strings.Contains(term, "<|") && !strings.Contains(term, "{target}")
}
