// Package uncompress permits seamless opening of compressed input files.
package uncompress

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

// Exts maps a file suffix to the name of the program to use in OpenFile for uncompressing it.
// Callers may modify Exts as needed.
var Exts = map[string]string{
	"Z":   "zcat",
	"z":   "zcat",
	"gz":  "zcat",
	"bz2": "bzcat",
	"xz":  "xzcat",
}

// OpenFile opens the named file for reading.
// If the file name has one of several special suffixes,
// the resulting ReadCloser contains the output of an uncompressing program,
// determined by Exts.
// If the file does not have a suffix from Exts,
// OpenFile falls back to using os.Open.
func OpenFile(name string) (io.ReadCloser, error) {
	return OpenFileContext(context.Background(), name)
}

// OpenFileContext opens the named file for reading, like OpenFile, but permits specifying a context object.
// Canceling this context will kill the uncompress subprocess (if any) writing to the resulting ReadCloser.
func OpenFileContext(ctx context.Context, name string) (io.ReadCloser, error) {
	for ext, prog := range Exts {
		if strings.HasSuffix(name, "."+ext) {
			ctx, cancel := context.WithCancel(ctx)
			cmd := exec.CommandContext(ctx, prog, name)
			r, err := cmd.StdoutPipe()
			if err != nil {
				cancel()
				return nil, err
			}
			err = cmd.Start()
			if err != nil {
				cancel()
				return nil, err
			}
			return &rwrapper{
				r:      r,
				cmd:    cmd,
				cancel: cancel,
			}, nil
		}
	}
	return os.Open(name)
}

type rwrapper struct {
	r      io.ReadCloser
	cmd    *exec.Cmd
	cancel context.CancelFunc
}

func (r rwrapper) Read(buf []byte) (int, error) {
	return r.r.Read(buf)
}

func (r rwrapper) Close() error {
	// Kill process.
	r.cancel()

	// Discard remaining bytes.
	io.Copy(ioutil.Discard, r.r)

	// Close r.r.
	return r.cmd.Wait()
}
