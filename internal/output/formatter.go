package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
	FormatYAML Format = "yaml"
)

type Formatter struct {
	format Format
	writer io.Writer
}

func NewFormatter(format string) *Formatter {
	f := Format(format)
	switch f {
	case FormatJSON, FormatYAML, FormatText:
	default:
		f = FormatText
	}
	return &Formatter{format: f, writer: os.Stdout}
}

func (f *Formatter) SetWriter(w io.Writer) {
	f.writer = w
}

// Print outputs data in the configured format.
// For text format, data should implement fmt.Stringer or be a string.
// For json/yaml, data is marshaled directly.
func (f *Formatter) Print(data any) error {
	switch f.format {
	case FormatJSON:
		enc := json.NewEncoder(f.writer)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	case FormatYAML:
		enc := yaml.NewEncoder(f.writer)
		enc.SetIndent(2)
		return enc.Encode(data)
	default:
		if s, ok := data.(fmt.Stringer); ok {
			_, err := fmt.Fprintln(f.writer, s.String())
			return err
		}
		_, err := fmt.Fprintln(f.writer, data)
		return err
	}
}
