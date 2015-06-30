package z

import (
	"archive/zip"
	"bytes"
	"github.com/kpmy/ypk/assert"
	"gopkg.in/yaml.v2"
	"io"
	"leaf/ir"
	"leaf/ir/target"
	"leaf/ir/target/yt"
	"time"
)

const CODE = "code"
const VERSION = "version"

type Version struct {
	Generator float64
	Code      int64
}

func load(sc io.Reader) (ret *ir.Module) {
	buf := bytes.NewBuffer(nil)
	len, _ := io.Copy(buf, sc)
	if r, err := zip.NewReader(bytes.NewReader(buf.Bytes()), len); err == nil {
		for _, f := range r.File {
			if f.Name == VERSION {
				rd, _ := f.Open()
				ver := &Version{}
				buf := bytes.NewBuffer(nil)
				io.Copy(buf, rd)
				yaml.Unmarshal(buf.Bytes(), ver)
				assert.For(ver.Generator == yt.VERSION, 40, "incompatible code version")
			}
			if f.Name == CODE {
				rd, _ := f.Open()
				ret = yt.Load(rd)
				//data, _ = ioutil.ReadAll(r)
			}
		}
	}
	return
}

func store(mod *ir.Module, tg io.Writer) {
	zw := zip.NewWriter(tg)
	if wr, err := zw.Create(CODE); err == nil {
		buf := bytes.NewBuffer(nil)
		yt.Store(mod, buf)
		io.Copy(wr, buf)
	}
	if wr, err := zw.Create(VERSION); err == nil {
		ver := &Version{}
		ver.Generator = yt.VERSION
		ver.Code = time.Now().Unix()
		data, _ := yaml.Marshal(ver)
		buf := bytes.NewBuffer(data)
		io.Copy(wr, buf)
	}
	zw.Close()
}

func init() {
	target.Ext = store
	target.Int = load
}
