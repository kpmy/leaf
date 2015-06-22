package z

import (
	"archive/zip"
	"bytes"
	"io"
	"leaf/ir"
	"leaf/target"
	"leaf/target/yt"
)

const CODE = "code"

func load(sc io.Reader) (ret *ir.Module) {
	buf := bytes.NewBuffer(nil)
	len, _ := io.Copy(buf, sc)
	if r, err := zip.NewReader(bytes.NewReader(buf.Bytes()), len); err == nil {
		for _, f := range r.File {
			if f.Name == CODE {
				rd, _ := f.Open()
				return yt.Load(rd)
				//data, _ = ioutil.ReadAll(r)
			}
		}
	}
	panic(0)
}

func store(mod *ir.Module, tg io.Writer) {
	buf := bytes.NewBuffer(nil)
	yt.Store(mod, buf)
	zw := zip.NewWriter(tg)
	if wr, err := zw.Create(CODE); err == nil {
		io.Copy(wr, buf)
		zw.Close()
	}
}

func init() {
	target.Ext = store
	target.Int = load
}
