package uncompress

import (
	"io"
	"os"
	"os/exec"
	"strings"
)

var exts = map[string]string{
	"Z":   "zcat",
	"z":   "zcat",
	"gz":  "zcat",
	"bz2": "bzcat",
	"xz":  "xzcat",
}

func OpenFile(name string) (io.ReadCloser, error) {
	for ext, prog := range exts {
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
	return r.cmd.Wait() // closes r.r
}
