package models

import "testing"

func TestNormalizeImagesJSON_Empty(t *testing.T) {
	if got := normalizeImagesJSON(""); got != "[]" {
		t.Fatalf("expected [], got %s", got)
	}
}

func TestNormalizeImagesJSON_JSONArrayWithBackticks(t *testing.T) {
	raw := "[\" `http://a.com/1.webp\\` \",\" `http://a.com/2.webp\\` \"]"
	got := normalizeImagesJSON(raw)
	want := "[\"http://a.com/1.webp\",\"http://a.com/2.webp\"]"
	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestNormalizeImagesJSON_FallbackRegex(t *testing.T) {
	raw := "xxx `http://a.com/1.webp` yyy http://a.com/2.webp zzz"
	got := normalizeImagesJSON(raw)
	want := "[\"http://a.com/1.webp\",\"http://a.com/2.webp\"]"
	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

