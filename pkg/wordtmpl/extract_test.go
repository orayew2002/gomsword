package wordtmpl

import (
	"context"
	"os"
	"testing"
)

// TestExtractKeysDocx ensures key extraction works on the sample .docx template.
func TestExtractKeysDocx(t *testing.T) {
	filePath := "testdata/template.docx"

	keys, err := ExtractKeys(context.Background(), filePath)
	if err != nil {
		t.Fatalf("Failed to extract keys: %v", err)
	}

	t.Log("keys", keys)
}

// TestExtractKeysDocxBytes ensures key extraction works from in-memory bytes.
func TestExtractKeysDocxBytes(t *testing.T) {
	filePath := "testdata/template.docx"

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read template: %v", err)
	}

	keys, err := ExtractKeysFromBytes(context.Background(), data)
	if err != nil {
		t.Fatalf("Failed to extract keys: %v", err)
	}

	t.Log("keys", keys)
}
