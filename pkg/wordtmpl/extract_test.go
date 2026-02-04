package wordtmpl

import (
	"context"
	"testing"
)

func TestExtractKeysDocx(t *testing.T) {
	filePath := "testdata/template.docx"

	keys, err := ExtractKeys(context.Background(), filePath)
	if err != nil {
		t.Fatalf("Failed to extract keys: %v", err)
	}

	t.Log("keys", keys)
}
