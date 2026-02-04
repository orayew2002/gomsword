package convert

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ConvertDocToDocx(ctx context.Context, path string) (string, func(), error) {
	if _, err := exec.LookPath("soffice"); err != nil {
		return "", func() {}, fmt.Errorf("libreoffice (soffice) not found in PATH: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "wordtmpl-docx-")
	if err != nil {
		return "", func() {}, fmt.Errorf("create temp dir: %w", err)
	}
	cleanup := func() {
		_ = os.RemoveAll(tmpDir)
	}

	cmd := exec.CommandContext(ctx, "soffice", "--headless", "--convert-to", "docx", "--outdir", tmpDir, path)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		cleanup()
		return "", func() {}, fmt.Errorf("libreoffice conversion failed: %w: %s", err, strings.TrimSpace(out.String()))
	}

	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	docxPath := filepath.Join(tmpDir, base+".docx")
	if _, err := os.Stat(docxPath); err != nil {
		cleanup()
		return "", func() {}, fmt.Errorf("converted file not found at %s: %w", docxPath, err)
	}

	return docxPath, cleanup, nil
}
