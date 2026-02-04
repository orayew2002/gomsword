# go_word

Word template key extraction and placeholder replacement using `{placeholder}` syntax.

## Overview
- Extract placeholder keys from Word templates.
- Fill values and generate a new `.docx` without modifying the original template.
- Optional `.doc` support via LibreOffice conversion.

## Features
- Extracts placeholder keys from `.docx` files.
- Generates a new `.docx` with placeholder replacements.
- Supports legacy `.doc` by converting to `.docx` with LibreOffice (`soffice`).
- Scans body, tables, headers, footers, and other `word/*.xml` parts.
- Deduplicates keys and returns them sorted.

## Usage (Library)

### Extract keys
```go
keys, err := wordtmpl.ExtractKeys(ctx, "/path/to/template.docx")
if err != nil {
	// handle error
}
// keys is sorted and de-duplicated: ["date", "name", ...]
```

### Extract keys from bytes
```go
keys, err := wordtmpl.ExtractKeysFromBytes(ctx, data)
if err != nil {
	// handle error
}
// keys is sorted and de-duplicated: ["date", "name", ...]
```

### Fill and save a new document
```go
doc, err := wordtmpl.Open(ctx, "/path/to/template.docx")
if err != nil {
	// handle error
}

doc.Val("first_name", "Jordan")
doc.Val("last_name", "Lee")

docxOut := "/path/to/output.docx"
if err := doc.Save(ctx, docxOut); err != nil {
	// handle error
}
```

### Work with binary content (MinIO, S3, etc.)
```go
doc, err := wordtmpl.OpenBytes(ctx, templateBytes)
if err != nil {
	// handle error
}

doc.Val("first_name", "Jordan")
doc.Val("last_name", "Lee")

outBytes, err := doc.SaveBytes(ctx)
if err != nil {
	// handle error
}
// you decide how to store outBytes, including filename and permissions
```

## Options

### Customize placeholder matching
```go
re := regexp.MustCompile(`\{([A-Za-z0-9_.-]+)\}`)
keys, err := wordtmpl.ExtractKeys(ctx, path, wordtmpl.WithPlaceholderRegex(re))
```

### Custom `.doc` converter
```go
converter := myConverter{}
keys, err := wordtmpl.ExtractKeys(ctx, path, wordtmpl.WithDocConverter(converter))
```

## CLI
```bash
go run ./cmd/wordkeys -- /path/to/template.docx
```

## `.doc` Support
Legacy `.doc` files are converted to `.docx` using LibreOffice headless:
```bash
soffice --headless --convert-to docx --outdir /tmp /path/to/file.doc
```
Make sure `soffice` is in `PATH` on the machine running the extractor.

## Limitations
- Placeholders split across multiple Word XML runs may not be replaced.
- Replacements only occur inside `word/*.xml` text nodes.

## Tests
```bash
GOCACHE=/tmp/go-build go test ./...
```
