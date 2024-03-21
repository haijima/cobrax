package cobrax

import (
	"errors"
	"io"
	"os"

	"github.com/spf13/afero"
	"golang.org/x/term"
)

var ErrNoFileSpecified = errors.New("no file specified")

type Option struct {
	stdin             io.Reader
	enableManualInput bool
}

type OptionFunc func(*Option)

func WithStdin(stdin io.Reader) OptionFunc { return func(o *Option) { o.stdin = stdin } }

func WithManualInputEnabled(o *Option) { o.enableManualInput = true }

func OpenOrStdIn(filename string, fs afero.Fs, opts ...OptionFunc) (io.ReadCloser, error) {
	o := &Option{stdin: os.Stdin}
	for _, f := range opts {
		f(o)
	}
	if filename != "" {
		f, err := fs.Open(filename)
		if err != nil {
			return nil, err
		}
		return f, nil
	} else if f, ok := o.stdin.(*os.File); ok && !o.enableManualInput && term.IsTerminal(int(f.Fd())) {
		return nil, ErrNoFileSpecified
	} else {
		return io.NopCloser(o.stdin), nil
	}
}
