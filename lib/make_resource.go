package lib

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
)

const staticDir = "html"

const MduiUrl = "https://cdn.w3cbus.com/mdui.org/mdui-v1.0.2.zip"

//const MduiUrl = "https://github.com/zdhxiong/mdui/releases/download/v1.0.2/mdui-v1.0.2.zip"

func DownloadMdui() {
	res, err := http.Get(MduiUrl)
	if err != nil {
		panic("download mdui error " + err.Error())
	}
	defer res.Body.Close()
	mdui, err := ioutil.TempFile("", "mdui*.zip")
	if err != nil {
		panic("create tmp file error " + err.Error())
	}
	_, err = io.Copy(mdui, res.Body)
	if err != nil {
		panic("create tmp file error " + err.Error())
	}

	zipReader, err := zip.OpenReader(mdui.Name())
	if err != nil {
		panic("open mdui zip error " + err.Error())
	}
	defer zipReader.Close()
	err = Unzip(zipReader, "html/static/mdui")
	if err != nil {
		panic("unzip mdui error " + err.Error())
	}
}

// MakeGen generate staticGenGoFile
func MakeGen() {
	f, err := os.Create("res/r.go")
	if err != nil {
		panic(err)
	}
	io.WriteString(f, `package res
func init() {
	r = &R
}
// R Static file resources
var R = map[string][]byte{`)
	getFiles(staticDir, func(sf *staticFile) {
		if sf == nil {
			return
		}
		if strings.Contains(sf.name, "mdui/fonts") {
			return
		}
		if sf.isDir {
			io.WriteString(f, "\n\t\"/"+sf.path)
			io.WriteString(f, `":nil,`)
		} else {
			var isRemoveN = false
			switch path.Ext(sf.name) {
			case ".map":
				return
			case ".js":
				if !strings.HasSuffix(sf.name, ".min.js") {
					return
				}
			case ".css":
				if !strings.HasSuffix(sf.name, ".min.css") {
					return
				}
			case ".html":
				isRemoveN = true
			}
			bs, err := ioutil.ReadFile(sf.name)
			if err != nil {
				panic(err)
			}

			io.WriteString(f, "\n\t\"/"+sf.path)
			io.WriteString(f, `":{`)
			for _, b := range bs {
				if isRemoveN && (b == '\n' || b == '\r') {
					continue
				}
				io.WriteString(f, strconv.Itoa(int(b)))
				io.WriteString(f, ",")
			}
			io.WriteString(f, "},")
		}
		println(sf.name, "generate successful")
	})
	io.WriteString(f, "\n\t\"/\":nil,")
	io.WriteString(f, "\n}")
	if err := f.Close(); err != nil {
		panic(err)
	}
}

// staticFile file info
type staticFile struct {
	isDir      bool
	name, path string
}

// getFiles func(string, func(*StaticFile))
func getFiles(dir string, callback func(*staticFile)) {
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _, f := range fs {
		name := dir + "/" + f.Name()
		sf := &staticFile{f.IsDir(), name, name[len(staticDir)+1:]}
		callback(sf)
		if f.IsDir() {
			getFiles(name, callback)
		}
	}
}
