package wordtmpl

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/orayew2002/gomsword/internal/docx"
)

// MsWord represents a Word template plus the values that should replace its keys.
// Create it with Open, set values with Val, then call Save to write a new file.
type MsWord struct {
	path   string
	data   []byte
	opts   options
	keys   []string
	values map[string]string
}

// Open loads a template, extracts keys, and prepares a document for value replacement.
// Use it once per template before calling Val and Save.
//
// Example:
//
//	doc, err := Open(ctx, "template.docx")
//	if err != nil {
//		// handle error
//	}
//	doc.Val("first_name", "Jordan")
//	_ = doc.Save(ctx, "output.docx")
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

// OpenBytes loads a .docx from memory, extracts keys, and prepares a document
// for value replacement. Use it when templates are fetched from object storage.
//
// Example:
//
//	doc, err := OpenBytes(ctx, data)
//	if err != nil {
//		// handle error
//	}
//	doc.Val("first_name", "Jordan")
//	out, _ := doc.SaveBytes(ctx)
//	_ = out
func OpenBytes(ctx context.Context, data []byte, opts ...Option) (*MsWord, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("docx data is empty")
	}

	o, err := buildOptions(opts...)
	if err != nil {
		return nil, err
	}

	keys, err := docx.ExtractKeysFromBytes(ctx, data, o.placeholder)
	if err != nil {
		return nil, err
	}

	copied := make([]byte, len(data))
	copy(copied, data)

	return &MsWord{
		data:   copied,
		opts:   o,
		keys:   keys,
		values: make(map[string]string),
	}, nil
}

// Keys returns a copy of the extracted placeholder keys.
// Use it to inspect which placeholders exist in the template.
//
// Example:
//
//	for _, key := range doc.Keys() {
//		_ = key
//	}
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
//
// Example:
//
//	doc.Val("first_name", "Jordan")
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
// Use SaveBytes when you need to control filename, permissions, or storage.
//
// Example:
//
//	if err := doc.Save(ctx, "filled.docx"); err != nil {
//		// handle error
//	}
func (m *MsWord) Save(ctx context.Context, outputPath string) error {
	if m == nil {
		return fmt.Errorf("MsWord is nil")
	}
	if outputPath == "" {
		return fmt.Errorf("output path is required")
	}
	if m.path != "" && samePath(m.path, outputPath) {
		return fmt.Errorf("output path must be different from template path")
	}
	if m.opts.placeholder == nil {
		return fmt.Errorf("placeholder regex must not be nil")
	}

	if len(m.data) > 0 {
		out, err := m.SaveBytes(ctx)
		if err != nil {
			return err
		}
		return os.WriteFile(outputPath, out, 0o644)
	}

	docxPath, cleanup, err := openDocx(ctx, m.path, m.opts)
	if err != nil {
		return err
	}
	defer cleanup()

	return docx.ReplaceKeys(ctx, docxPath, m.opts.placeholder, m.values, outputPath)
}

// SaveBytes returns a new .docx as bytes with placeholders replaced.
// Use it when you want to decide the filename and permissions yourself.
//
// Example:
//
//	out, err := doc.SaveBytes(ctx)
//	if err != nil {
//		// handle error
//	}
//	_ = out
func (m *MsWord) SaveBytes(ctx context.Context) ([]byte, error) {
	if m == nil {
		return nil, fmt.Errorf("MsWord is nil")
	}
	if m.opts.placeholder == nil {
		return nil, fmt.Errorf("placeholder regex must not be nil")
	}

	if len(m.data) > 0 {
		return docx.ReplaceKeysFromBytes(ctx, m.data, m.opts.placeholder, m.values)
	}
	if m.path == "" {
		return nil, fmt.Errorf("template source is missing")
	}

	docxPath, cleanup, err := openDocx(ctx, m.path, m.opts)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	return docx.ReplaceKeysToBytes(ctx, docxPath, m.opts.placeholder, m.values)
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
