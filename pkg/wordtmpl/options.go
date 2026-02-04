package wordtmpl

import (
	"context"
	"regexp"
)

type DocConverter interface {
	ConvertDocToDocx(ctx context.Context, path string) (string, func(), error)
}

type options struct {
	placeholder *regexp.Regexp
	converter   DocConverter
}

var defaultPlaceholder = regexp.MustCompile(`\{([^{}]+)\}`)

type Option func(*options)

func WithPlaceholderRegex(re *regexp.Regexp) Option {
	return func(o *options) {
		if re != nil {
			o.placeholder = re
		}
	}
}

func WithDocConverter(converter DocConverter) Option {
	return func(o *options) {
		if converter != nil {
			o.converter = converter
		}
	}
}

type docConverterFunc func(ctx context.Context, path string) (string, func(), error)

func (f docConverterFunc) ConvertDocToDocx(ctx context.Context, path string) (string, func(), error) {
	return f(ctx, path)
}
