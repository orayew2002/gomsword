package wordtmpl

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/orayew2002/gomsword/internal/docx"
)

// MsWord represents a Word template plus the values that should replace its keys.
// Create it with Open, set values with Val, then call Save to write a new file.
type MsWord struct {
	path   string
	opts   options
	keys   []string
	values map[string]string
}

// Open loads a template, extracts keys, and prepares a document for value replacement.
// Use it once per template before calling Val and Save.
func Open(ctx context.Context, path string, opts ...Option) (*MsWord, error) {
	if path == "" {
		return nil, fmt.Errorf("path is required")
	}

	o, err := buildOptions(opts...)
	if err != nil {
		return nil, err
	}

	docxPath, cleanup, err := openDocx(ctx, path, o)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	keys, err := docx.ExtractKeys(ctx, docxPath, o.placeholder)
	if err != nil {
		return nil, err
	}

	return &MsWord{
		path:   path,
		opts:   o,
		keys:   keys,
		values: make(map[string]string),
	}, nil
}

// Keys returns a copy of the extracted placeholder keys.
// Use it to inspect which placeholders exist in the template.
func (m *MsWord) Keys() []string {
	if m == nil {
		return nil
	}
	out := make([]string, len(m.keys))
	copy(out, m.keys)
	return out
}

// Val assigns a replacement value for a placeholder key.
// Call it with a key name (with or without `{}`) before Save.
func (m *MsWord) Val(key, value string) {
	if m == nil {
		return
	}
	normalized := normalizeKey(key)
	if normalized == "" {
		return
	}
	if m.values == nil {
		m.values = make(map[string]string)
	}
	m.values[normalized] = value
}

// Save writes a new .docx with placeholders replaced using values in the MsWord.
// The output path must differ from the template; the original file is untouched.
func (m *MsWord) Save(ctx context.Context, outputPath string) error {
	if m == nil {
		return fmt.Errorf("MsWord is nil")
	}
	if outputPath == "" {
		return fmt.Errorf("output path is required")
	}
	if samePath(m.path, outputPath) {
		return fmt.Errorf("output path must be different from template path")
	}
	if m.opts.placeholder == nil {
		return fmt.Errorf("placeholder regex must not be nil")
	}

	docxPath, cleanup, err := openDocx(ctx, m.path, m.opts)
	if err != nil {
		return err
	}
	defer cleanup()

	return docx.ReplaceKeys(ctx, docxPath, m.opts.placeholder, m.values, outputPath)
}

// normalizeKey trims whitespace and optional `{}` wrappers for consistent lookups.
func normalizeKey(key string) string {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") && len(trimmed) >= 2 {
		trimmed = strings.TrimSpace(trimmed[1 : len(trimmed)-1])
	}
	return trimmed
}

// samePath reports whether two file paths point to the same location.
func samePath(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	absA, errA := filepath.Abs(a)
	absB, errB := filepath.Abs(b)
	if errA == nil && errB == nil {
		return absA == absB
	}
	return a == b
}
