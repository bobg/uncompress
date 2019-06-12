// Package uncompress permits seamless opening of compressed input files.
package uncompress

import (
	"io"
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
	for ext, prog := range Exts {
		if strings.HasSuffix(name, "."+ext) {
			cmd := exec.Command(prog, name)
			r, err := cmd.StdoutPipe()
			if err != nil {
				return nil, err
			}
			err = cmd.Start()
			if err != nil {
				return nil, err
			}
			return &rwrapper{r, cmd}, nil
		}
	}
	return os.Open(name)
}

type rwrapper struct {
	r   io.ReadCloser
	cmd *exec.Cmd
}

func (r rwrapper) Read(buf []byte) (int, error) {
	return r.r.Read(buf)
}

func (r rwrapper) Close() error {
	// Discard remaining bytes.
	var buf [8192]byte
	for {
		_, err := r.r.Read(buf[:])
		if err != nil {
			break
		}
	}

	// Closes r.r.
	return r.cmd.Wait()
}
