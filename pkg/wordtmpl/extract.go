package wordtmpl

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/orayew2002/gomsword/internal/convert"
	"github.com/orayew2002/gomsword/internal/docx"
)

func ExtractKeys(ctx context.Context, path string, opts ...Option) ([]string, error) {
	if path == "" {
		return nil, fmt.Errorf("path is required")
	}

	o := options{
		placeholder: defaultPlaceholder,
		converter:   docConverterFunc(convert.ConvertDocToDocx),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	if o.placeholder == nil {
		return nil, fmt.Errorf("placeholder regex must not be nil")
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".docx":
		return docx.ExtractKeys(ctx, path, o.placeholder)
	case ".doc":
		if o.converter == nil {
			return nil, fmt.Errorf(".doc conversion is not configured")
		}
		docxPath, cleanup, err := o.converter.ConvertDocToDocx(ctx, path)
		if err != nil {
			return nil, err
		}
		defer cleanup()
		return docx.ExtractKeys(ctx, docxPath, o.placeholder)
	default:
		return nil, fmt.Errorf("unsupported file extension %q", ext)
	}
}
