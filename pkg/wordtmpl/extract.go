package wordtmpl

import (
	"context"
	"fmt"

	"github.com/orayew2002/gomsword/internal/docx"
)

// ExtractKeys scans a Word template and returns a sorted, de-duplicated list
// of placeholder keys. Call it with the template path and optional settings
// like WithPlaceholderRegex or WithDocConverter.
//
// Example:
//
//	keys, err := ExtractKeys(ctx, "template.docx")
//	if err != nil {
//		// handle error
//	}
//	_ = keys
func ExtractKeys(ctx context.Context, path string, opts ...Option) ([]string, error) {
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

	return docx.ExtractKeys(ctx, docxPath, o.placeholder)
}

// ExtractKeysFromBytes scans a .docx byte slice and returns a sorted,
// de-duplicated list of placeholder keys. Use it when your template is already
// in memory (S3, MinIO, HTTP uploads, etc.). Placeholder options such as
// WithPlaceholderRegex are supported. .doc conversion is not available for
// in-memory data.
//
// Example:
//
//	keys, err := ExtractKeysFromBytes(ctx, data)
//	if err != nil {
//		// handle error
//	}
//	_ = keys
func ExtractKeysFromBytes(ctx context.Context, data []byte, opts ...Option) ([]string, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("docx data is required")
	}

	o, err := buildOptions(opts...)
	if err != nil {
		return nil, err
	}

	return docx.ExtractKeysFromBytes(ctx, data, o.placeholder)
}
