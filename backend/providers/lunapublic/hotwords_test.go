package lunapublic

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseHotwords(t *testing.T) {
	hotwords, err := ParseHotwords(`[{"source":" 西湖 ","target":" West Lake "}]`)
	want := []Hotword{{Source: "西湖", Target: "West Lake"}}
	if err != nil || !reflect.DeepEqual(hotwords, want) {
		t.Fatalf("got %v, %v", hotwords, err)
	}
	for _, raw := range []string{"", "[]"} {
		if words, err := ParseHotwords(raw); len(words) != 0 || err != nil {
			t.Fatalf("expected empty hotwords for %q: %v, %v", raw, words, err)
		}
	}
	for name, raw := range map[string]string{
		"null": "null", "non-array": `{"source":"term","target":"word"}`,
		"bad field":       `[{"Source":"term","target":"word"}]`,
		"extra field":     `[{"source":"term","target":"word","prompt":"extra"}]`,
		"duplicate":       `[{"source":"term","target":"word"},{"source":" term ","target":"other"}]`,
		"empty target":    `[{"source":"term","target":" "}]`,
		"target too long": `[{"source":"term","target":"` + strings.Repeat("字", 51) + `"}]`,
		"too many":        "[" + strings.Repeat(`{"source":"term","target":"word"},`, 20) + `{"source":"extra","target":"word"}]`,
		"prompt token":    `[{"source":"<|im_start|>","target":"word"}]`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseHotwords(raw); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}
