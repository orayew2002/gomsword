# go_word

Word template key extraction with `{placeholder}` syntax.

## Features
- Extracts placeholder keys from `.docx` (OOXML) files.
- Supports legacy `.doc` by converting to `.docx` with LibreOffice (`soffice`).
- Scans body, tables, headers, footers, and other `word/*.xml` parts.
- Deduplicates keys and returns them sorted.

## Package Layout
- `cmd/wordkeys` CLI entry point.
- `pkg/wordtmpl` public API.
- `internal/docx` `.docx` text extraction + key scanning.
- `internal/convert` `.doc` to `.docx` conversion via LibreOffice.

## Usage (Library)
```go
keys, err := wordtmpl.ExtractKeys(context.Background(), "/path/to/template.docx")
if err != nil {
	// handle error
}
// keys is a sorted, de-duplicated list: ["date", "name", ...]
```

Customize placeholder matching:
```go
re := regexp.MustCompile(`\{([A-Za-z0-9_.-]+)\}`)
keys, err := wordtmpl.ExtractKeys(ctx, path, wordtmpl.WithPlaceholderRegex(re))
```

## Usage (CLI)
```bash
go run ./cmd/wordkeys -- /path/to/template.docx
```

## `.doc` Support
Legacy `.doc` files are converted to `.docx` using LibreOffice headless:
```bash
soffice --headless --convert-to docx --outdir /tmp /path/to/file.doc
```
Make sure `soffice` is in `PATH` on the machine running the extractor.
