// Package uncompress permits seamless opening of compressed input files.
package uncompress

import (
	"compress/bzip2"
	"compress/gzip"
	"io"
	"os"
	"strings"

	"github.com/ulikunitz/xz"
)

// An Opener receives a ReadCloser with compressed data and returns a ReadCloser with uncompressed data.
type Opener = func(io.ReadCloser) (io.ReadCloser, error)

// Exts maps a file suffix to the name of the program to use in OpenFile for uncompressing it.
// Callers may modify Exts as needed.
var Exts = map[string]Opener{
	"Z":   gzipOpener,
	"z":   gzipOpener,
	"gz":  gzipOpener,
	"bz2": bzip2Opener,
	"xz":  xzOpener,
}

// Open opens the named file for reading.
// If the file name has one of several special suffixes,
// the resulting ReadCloser contains the result of uncompressing the file contents,
// using the uncompressing method determined by Exts.
// If the file does not have a suffix from Exts,
// OpenFile falls back to using os.Open.
func Open(name string) (io.ReadCloser, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	for ext, prog := range Exts {
		if strings.HasSuffix(name, "."+ext) {
			r, err := prog(f)
			if err != nil {
				f.Close()
				return nil, err
			}
			return r, nil
		}
	}
	return f, nil
}

func gzipOpener(underlying io.ReadCloser) (io.ReadCloser, error) {
	r, err := gzip.NewReader(underlying)
	return &wrapper{r: r, underlying: underlying}, err
}

func bzip2Opener(underlying io.ReadCloser) (io.ReadCloser, error) {
	return &wrapper{r: io.NopCloser(bzip2.NewReader(underlying)), underlying: underlying}, nil
}

func xzOpener(underlying io.ReadCloser) (io.ReadCloser, error) {
	r, err := xz.NewReader(underlying)
	return &wrapper{r: io.NopCloser(r), underlying: underlying}, err
}

type wrapper struct {
	r          io.ReadCloser
	underlying io.Closer
}

func (w *wrapper) Read(p []byte) (int, error) {
	return w.r.Read(p)
}

func (w *wrapper) Close() error {
	if w.r == nil {
		return nil
	}
	err := w.r.Close()
	w.underlying.Close()
	w.r = nil
	w.underlying.Close()
	return err
}
