package wordtmpl

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

// openDocx resolves a .docx path, converting .doc when required, and returns
// a cleanup function for temporary files. Used by extraction and save flows
// to keep conversion logic in one place.
func openDocx(ctx context.Context, path string, o options) (string, func(), error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".docx":
		return path, func() {}, nil
	case ".doc":
		if o.converter == nil {
			return "", func() {}, fmt.Errorf(".doc conversion is not configured")
		}
		return o.converter.ConvertDocToDocx(ctx, path)
	default:
		return "", func() {}, fmt.Errorf("unsupported file extension %q", ext)
	}
}
