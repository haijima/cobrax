package cobrax

import (
	"io"

	"github.com/spf13/afero"
)

func OpenOrStdIn(filename string, fs afero.Fs, stdin io.Reader) (io.ReadCloser, error) {
	if filename != "" {
		f, err := fs.Open(filename)
		if err != nil {
			return nil, err
		}
		return f, nil
	} else {
		return io.NopCloser(stdin), nil
	}
}
