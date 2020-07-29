package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

type _escLocalFS struct{}

var _escLocal _escLocalFS

type _escStaticFS struct{}

var _escStatic _escStaticFS

type _escDirectory struct {
	fs   http.FileSystem
	name string
}

type _escFile struct {
	compressed string
	size       int64
	modtime    int64
	local      string
	isDir      bool

	once sync.Once
	data []byte
	name string
}

func (_escLocalFS) Open(name string) (http.File, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	return os.Open(f.local)
}

func (_escStaticFS) prepare(name string) (*_escFile, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	var err error
	f.once.Do(func() {
		f.name = path.Base(name)
		if f.size == 0 {
			return
		}
		var gr *gzip.Reader
		b64 := base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(f.compressed))
		gr, err = gzip.NewReader(b64)
		if err != nil {
			return
		}
		f.data, err = ioutil.ReadAll(gr)
	})
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (fs _escStaticFS) Open(name string) (http.File, error) {
	f, err := fs.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.File()
}

func (dir _escDirectory) Open(name string) (http.File, error) {
	return dir.fs.Open(dir.name + name)
}

func (f *_escFile) File() (http.File, error) {
	type httpFile struct {
		*bytes.Reader
		*_escFile
	}
	return &httpFile{
		Reader:   bytes.NewReader(f.data),
		_escFile: f,
	}, nil
}

func (f *_escFile) Close() error {
	return nil
}

func (f *_escFile) Readdir(count int) ([]os.FileInfo, error) {
	if !f.isDir {
		return nil, fmt.Errorf(" escFile.Readdir: '%s' is not directory", f.name)
	}

	fis, ok := _escDirs[f.local]
	if !ok {
		return nil, fmt.Errorf(" escFile.Readdir: '%s' is directory, but we have no info about content of this dir, local=%s", f.name, f.local)
	}
	limit := count
	if count <= 0 || limit > len(fis) {
		limit = len(fis)
	}

	if len(fis) == 0 && count > 0 {
		return nil, io.EOF
	}

	return fis[0:limit], nil
}

func (f *_escFile) Stat() (os.FileInfo, error) {
	return f, nil
}

func (f *_escFile) Name() string {
	return f.name
}

func (f *_escFile) Size() int64 {
	return f.size
}

func (f *_escFile) Mode() os.FileMode {
	return 0
}

func (f *_escFile) ModTime() time.Time {
	return time.Unix(f.modtime, 0)
}

func (f *_escFile) IsDir() bool {
	return f.isDir
}

func (f *_escFile) Sys() interface{} {
	return f
}

// FS returns a http.Filesystem for the embedded assets. If useLocal is true,
// the filesystem's contents are instead used.
func FS(useLocal bool) http.FileSystem {
	if useLocal {
		return _escLocal
	}
	return _escStatic
}

// Dir returns a http.Filesystem for the embedded assets on a given prefix dir.
// If useLocal is true, the filesystem's contents are instead used.
func Dir(useLocal bool, name string) http.FileSystem {
	if useLocal {
		return _escDirectory{fs: _escLocal, name: name}
	}
	return _escDirectory{fs: _escStatic, name: name}
}

// FSByte returns the named file from the embedded assets. If useLocal is
// true, the filesystem's contents are instead used.
func FSByte(useLocal bool, name string) ([]byte, error) {
	if useLocal {
		f, err := _escLocal.Open(name)
		if err != nil {
			return nil, err
		}
		b, err := ioutil.ReadAll(f)
		_ = f.Close()
		return b, err
	}
	f, err := _escStatic.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.data, nil
}

// FSMustByte is the same as FSByte, but panics if name is not present.
func FSMustByte(useLocal bool, name string) []byte {
	b, err := FSByte(useLocal, name)
	if err != nil {
		panic(err)
	}
	return b
}

// FSString is the string version of FSByte.
func FSString(useLocal bool, name string) (string, error) {
	b, err := FSByte(useLocal, name)
	return string(b), err
}

// FSMustString is the string version of FSMustByte.
func FSMustString(useLocal bool, name string) string {
	return string(FSMustByte(useLocal, name))
}

var _escData = map[string]*_escFile{

	"/index.html": {
		name:    "index.html",
		local:   "web_assets/index.html",
		size:    3548,
		modtime: 1576748870,
		compressed: `
H4sIAAAAAAAC/9RXX3PbNhJ/16fYQ3NH6mSSkmXHjiyq47NSx+mlzsl2OrlOHkBwRUImAQYA9acef/cb
kJRM2clNH1s92MDuYnd/+1capybPJh2AcY6GAkup0mhCcnVz7Z2eHr/xBuSJK2iOIVlyXBVSGQJMCoPC
hGTFY5OGMS45Q6+6HAAX3HCaeZrRDMOB3z+AnK55XuZtUqlRVXcaZRj2a2Mp0tgeAMaGmwwn76SZXU/B
gxmPUcO1gCnmVMTjoObXspopXhjQioUkNabQoyBgMkZ/8bVEtfGZzIP66A39gT/wcy78hSaTcVA/bfRk
XNyDwiwk2mwy1CmiIZAqnD/pzemaxcKPpDTaKFrYi9W/IwRDf+ifBEzrJ1plkGlNgAuDieJmExKd0uHp
kfevT585v7n6CX8exJf5+9n5/YaV787fzZLh4XV+x1arEymGs89xcvSJ9j7mN7f69+Dn16fLKH67SI9K
AkxJraXiCRchoUKKTS5LTf5PcP4oiMVzDItvQrhlx1f/4VH/8OTrcrO4+TB/t7j+QP99Py9//bT+7/ru
o7h4f36SHeYXv/5yVVy+yS8vpqery1+u2Mfpye2afh/CU4IaMDYvk45fljyGB8ipSrjwjCxGMDgu1mfw
2PFTaZSMvag0Rgp4gILGMRfJCA77VoKVSks1gkJaIOpsX0n/W0pGqVyigoeXb+c8M6hGECmepEag1u7p
8d+7VsUPjYpMJt/x9AfDi++wKrBBg9Z2RrBtjXEk402T2pgvgWVU65DYjqRcoGrSvs+twkUzVKb+63Ex
lza6MV/u5BlaTNur7caB7T+Y+df+1B8H6aDNO5qMMZ+8aEvMJ+MgPWpJttxQckWeOC8hZF4ee0OwB517
r5/J1gVQUPGCaj+NksgIiIyoAFaHKJPsHvbSSb6pIKaGeqzURuaoQjI4HJLJjLIUM0fDT5lUNIMpap4I
PQ6sG8+QtGP5Zwc3fHNIJrdK5nCRSiYzajiqvzyqk+GATN7TggrUaHOlUZm/frKOX5+QyXlOf+cigQs5
nyPCTFJtUP0RcM+vFiePQ2J4QSYXGWf3IAVszVWrHmgklwhGglQxKqDAqPK/p+hpzpEt9gxp/Hy8BO35
smONg3qedXabatLpzEvBDJcC5lLl1ExLRe3VjZtDFx46AEuqIIYQtlQIwB30qw/8Ewb1v9f97lkjWwpu
NITg5Fw4lqjQlErAB2pSX8lSxG7chV4td9Z57HTsK5ZxFObu7moKYVu0PlIRy9ztNvasLfsmo9rM8GuJ
2lTP+medziuXVFuLdH37zcsln2WpYIVRY8HRwOOR3XBKimRCoNc23QNi10HN6jbq9kup6zObTHcbPBeX
pg7Unju93jYequXhvinPWt97tX0zV6jTC6oghFfuK5e0lhzp+oXCAkXsOu1mqp54jCpSbYwp1wU1LLXF
XNeV7/+m8OsInN7Oo57zpVkltkycrs9SnsUKhdv9rf9ll9Fd0YaAS+MbqhI0vm0fjcbfcq203Z6obPYf
mnp0FhQTVF5Ek4Qm6IzA0ag1lyJ0nsfeOdgGq+Lt3OwAPFrtTAotM/QzmbiNpZ2PEc6lQghhSg36Qq5c
mz6AIIA7jfYbhUJh4G52BVRDRDUW1KS28oEu6HprTDfqLPOjwjlfQwgrLmK58jPJqgbwLdP2r7W9J9i6
/C0EEhD4sU0bgeNUTr3yrU23xeqBE8RNyn7czaQqQtvg98D5h5BCY0Xe64uDJtxNUEbbw0FFzdGkMh6B
c/n21qlJumQMtR7BroptNg/A4NrcGGpK3d1l0IaDzk2V/nZwmyncyonVsWNUY2M7MsIXI4Ya6r+9Pd+J
byu+7lqn+UEyjiYWbCU9VXxZh2EcRBOgSvGlrW4uoJLZ2uqBA02lt4uorq+MGhRsU/PcCpZXl44dSU6u
vziNS482Uo/ds85jXUitr8jjoP5R978AAAD//wouYFDcDQAA
`,
	},

	"/": {
		name:  "/",
		local: `web_assets`,
		isDir: true,
	},
}

var _escDirs = map[string][]os.FileInfo{

	"web_assets": {
		_escData["/index.html"],
	},
}
