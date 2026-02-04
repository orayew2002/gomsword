package wordtmpl

import (
	"context"
	"fmt"

	"github.com/orayew2002/gomsword/internal/docx"
)

// ExtractKeys scans a Word template and returns a sorted, de-duplicated list
// of placeholder keys. Call it with the template path and optional settings
// like WithPlaceholderRegex or WithDocConverter.
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
