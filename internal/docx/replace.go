package docx

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// ReplaceKeys copies a .docx and replaces placeholders inside word/*.xml parts.
func ReplaceKeys(ctx context.Context, path string, re *regexp.Regexp, values map[string]string, outPath string) error {
	if path == "" {
		return fmt.Errorf("path is required")
	}
	if outPath == "" {
		return fmt.Errorf("output path is required")
	}
	if re == nil {
		return fmt.Errorf("placeholder regex must not be nil")
	}

	zr, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("open docx: %w", err)
	}
	defer zr.Close()

	outFile, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	success := false
	defer func() {
		if !success {
			_ = outFile.Close()
			_ = os.Remove(outPath)
		}
	}()

	zw := zip.NewWriter(outFile)

	if err := replaceKeysFromFiles(ctx, zr.File, re, values, zw); err != nil {
		return err
	}

	if err := zw.Close(); err != nil {
		return fmt.Errorf("finalize output: %w", err)
	}
	if err := outFile.Close(); err != nil {
		return fmt.Errorf("close output: %w", err)
	}
	success = true
	return nil
}

// ReplaceKeysToBytes returns a new .docx byte slice with placeholders replaced.
func ReplaceKeysToBytes(ctx context.Context, path string, re *regexp.Regexp, values map[string]string) ([]byte, error) {
	if path == "" {
		return nil, fmt.Errorf("path is required")
	}
	if re == nil {
		return nil, fmt.Errorf("placeholder regex must not be nil")
	}

	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("open docx: %w", err)
	}
	defer zr.Close()

	var out bytes.Buffer
	zw := zip.NewWriter(&out)
	if err := replaceKeysFromFiles(ctx, zr.File, re, values, zw); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("finalize output: %w", err)
	}
	return out.Bytes(), nil
}

// ReplaceKeysFromBytes returns a new .docx byte slice with placeholders replaced.
func ReplaceKeysFromBytes(ctx context.Context, data []byte, re *regexp.Regexp, values map[string]string) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("docx data is empty")
	}
	if re == nil {
		return nil, fmt.Errorf("placeholder regex must not be nil")
	}

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("open docx: %w", err)
	}

	var out bytes.Buffer
	zw := zip.NewWriter(&out)
	if err := replaceKeysFromFiles(ctx, zr.File, re, values, zw); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("finalize output: %w", err)
	}
	return out.Bytes(), nil
}

// replaceKeysFromFiles streams zip entries into a new writer while replacing placeholders.
func replaceKeysFromFiles(ctx context.Context, files []*zip.File, re *regexp.Regexp, values map[string]string, zw *zip.Writer) error {
	for _, f := range files {
		if ctx != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
		}

		header := f.FileHeader
		header.Name = f.Name
		w, err := zw.CreateHeader(&header)
		if err != nil {
			return fmt.Errorf("create %s: %w", f.Name, err)
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("open %s: %w", f.Name, err)
		}

		name := strings.ToLower(f.Name)
		if strings.HasPrefix(name, "word/") && strings.HasSuffix(name, ".xml") {
			err = replaceInWordXML(ctx, rc, w, re, values)
		} else {
			_, err = io.Copy(w, rc)
		}
		rc.Close()
		if err != nil {
			return fmt.Errorf("write %s: %w", f.Name, err)
		}
	}
	return nil
}

// replaceInWordXML streams XML and rewrites text nodes with replacements.
func replaceInWordXML(ctx context.Context, r io.Reader, w io.Writer, re *regexp.Regexp, values map[string]string) error {
	dec := xml.NewDecoder(r)
	enc := xml.NewEncoder(w)
	textDepth := 0

	for {
		if ctx != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
		}

		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			local := strings.ToLower(t.Name.Local)
			if isTextElement(local) {
				textDepth++
			}
			if err := enc.EncodeToken(t); err != nil {
				return err
			}
		case xml.EndElement:
			local := strings.ToLower(t.Name.Local)
			if isTextElement(local) && textDepth > 0 {
				textDepth--
			}
			if err := enc.EncodeToken(t); err != nil {
				return err
			}
		case xml.CharData:
			if textDepth > 0 {
				replaced := replacePlaceholders(string([]byte(t)), re, values)
				if err := enc.EncodeToken(xml.CharData([]byte(replaced))); err != nil {
					return err
				}
			} else {
				if err := enc.EncodeToken(t); err != nil {
					return err
				}
			}
		default:
			if err := enc.EncodeToken(t); err != nil {
				return err
			}
		}
	}

	return enc.Flush()
}

// replacePlaceholders substitutes matched placeholders using the values map.
func replacePlaceholders(input string, re *regexp.Regexp, values map[string]string) string {
	if input == "" || re == nil || len(values) == 0 {
		return input
	}
	return re.ReplaceAllStringFunc(input, func(match string) string {
		key := placeholderKey(re, match)
		if key == "" {
			return match
		}
		if val, ok := values[key]; ok {
			return val
		}
		return match
	})
}

// placeholderKey extracts the key name from a matched placeholder string.
func placeholderKey(re *regexp.Regexp, match string) string {
	sub := re.FindStringSubmatch(match)
	switch {
	case len(sub) > 1:
		return strings.TrimSpace(sub[1])
	case len(sub) == 1:
		return strings.TrimSpace(strings.Trim(sub[0], "{}"))
	default:
		return ""
	}
}
