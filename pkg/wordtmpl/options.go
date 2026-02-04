package wordtmpl

import (
	"context"
	"fmt"
	"regexp"

	"github.com/orayew2002/gomsword/internal/convert"
)

// DocConverter converts legacy .doc files into temporary .docx files.
// Provide one when you want a custom conversion pipeline.
type DocConverter interface {
	ConvertDocToDocx(ctx context.Context, path string) (string, func(), error)
}

// options holds resolved settings for placeholder matching and .doc conversion.
type options struct {
	placeholder *regexp.Regexp
	converter   DocConverter
}

// defaultPlaceholder matches `{key}`-style placeholders.
var defaultPlaceholder = regexp.MustCompile(`\{([^{}]+)\}`)

// Option customizes template parsing and conversion behavior.
type Option func(*options)

// WithPlaceholderRegex overrides the placeholder regex used for matching keys.
// Use it when your templates do not follow `{key}` syntax.
func WithPlaceholderRegex(re *regexp.Regexp) Option {
	return func(o *options) {
		if re != nil {
			o.placeholder = re
		}
	}
}

// WithDocConverter overrides the converter used for .doc -> .docx conversion.
// Use it to plug in a different conversion implementation.
func WithDocConverter(converter DocConverter) Option {
	return func(o *options) {
		if converter != nil {
			o.converter = converter
		}
	}
}

type docConverterFunc func(ctx context.Context, path string) (string, func(), error)

// ConvertDocToDocx adapts a function into a DocConverter implementation.
func (f docConverterFunc) ConvertDocToDocx(ctx context.Context, path string) (string, func(), error) {
	return f(ctx, path)
}

// buildOptions applies defaults and user options, returning a validated config.
func buildOptions(opts ...Option) (options, error) {
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
		return o, fmt.Errorf("placeholder regex must not be nil")
	}
	return o, nil
}
