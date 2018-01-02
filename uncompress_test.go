package uncompress

import (
	"bytes"
	"io/ioutil"
	"path"
	"testing"
)

func TestUncompress(t *testing.T) {
	infos, err := ioutil.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}
	for _, info := range infos {
		file := path.Join("testdata", info.Name())
		t.Run(file, func(t *testing.T) {
			f, err := OpenFile(file)
			if err != nil {
				t.Fatal(err)
			}
			buf, err := ioutil.ReadAll(f)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(buf, []byte("Excelsior")) {
				t.Errorf("got %s, want Excelsior", string(buf))
			}
			err = f.Close()
			if err != nil {
				t.Errorf("closing file: %s", err)
			}
		})
	}
}
