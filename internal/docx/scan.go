package docx

import (
	"archive/zip"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
)

func ExtractKeys(ctx context.Context, path string, re *regexp.Regexp) ([]string, error) {
	if re == nil {
		return nil, fmt.Errorf("placeholder regex must not be nil")
	}

	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("open docx: %w", err)
	}
	defer zr.Close()

	keys := make(map[string]struct{})
	for _, f := range zr.File {
		if ctx != nil {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
		}

		name := strings.ToLower(f.Name)
		if !strings.HasPrefix(name, "word/") || !strings.HasSuffix(name, ".xml") {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("open %s: %w", f.Name, err)
		}
		text, err := extractTextFromWordXML(rc)
		rc.Close()
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", f.Name, err)
		}

		for _, m := range re.FindAllStringSubmatch(text, -1) {
			var key string
			switch {
			case len(m) > 1:
				key = strings.TrimSpace(m[1])
			case len(m) == 1:
				key = strings.TrimSpace(strings.Trim(m[0], "{}"))
			}
			if key == "" {
				continue
			}
			keys[key] = struct{}{}
		}
	}

	out := make([]string, 0, len(keys))
	for k := range keys {
		out = append(out, k)
	}
	sort.Strings(out)
	return out, nil
}

func extractTextFromWordXML(r io.Reader) (string, error) {
	dec := xml.NewDecoder(r)
	var b strings.Builder
	textDepth := 0

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			local := strings.ToLower(t.Name.Local)
			if isBoundaryElement(local) {
				b.WriteByte('\n')
			}
			if isTabElement(local) {
				b.WriteByte('\t')
			}
			if isTextElement(local) {
				textDepth++
			}
		case xml.EndElement:
			local := strings.ToLower(t.Name.Local)
			if isTextElement(local) && textDepth > 0 {
				textDepth--
			}
		case xml.CharData:
			if textDepth > 0 {
				b.Write([]byte(t))
			}
		}
	}

	return b.String(), nil
}

func isTextElement(local string) bool {
	switch local {
	case "t", "deltext", "instrtext":
		return true
	default:
		return false
	}
}

func isBoundaryElement(local string) bool {
	switch local {
	case "p", "tr", "tc", "tbl", "br", "cr":
		return true
	default:
		return false
	}
}

func isTabElement(local string) bool {
	switch local {
	case "tab":
		return true
	default:
		return false
	}
}
