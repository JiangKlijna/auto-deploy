package lib

import (
	"archive/zip"
	"crypto/md5"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// HashCalculation calculat hash
func HashCalculation(h hash.Hash, val string) string {
	h.Write([]byte(val))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// ReverseString reverse string
func ReverseString(s string) string {
	runes := []rune(s)
	for from, to := 0, len(runes)-1; from < to; from, to = from+1, to-1 {
		runes[from], runes[to] = runes[to], runes[from]
	}
	return string(runes)
}

// Zip Compress
func Zip(srcDir, zipFile string, callback func(string)) error {
	os.MkdirAll(path.Dir(zipFile), 0666)
	zipfile, err := os.Create(zipFile)
	if err != nil {
		callback("open file error " + zipFile)
		return err
	}
	defer zipfile.Close()

	// open zip
	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	// foreach dir
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, _ error) error {
		if path == srcDir {
			return nil
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = strings.TrimPrefix(path, srcDir+`\`)
		if info.IsDir() {
			header.Name += `/`
		} else {
			header.Method = zip.Deflate
		}
		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// Unzip Decompress
func Unzip(zipReader *zip.ReadCloser, destDir string) error {
	//zipReader, err := zip.OpenReader(zipFile)
	//if err != nil {
	//	return err
	//}
	//defer zipReader.Close()
	for _, f := range zipReader.File {
		fpath := filepath.Join(destDir, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
		} else {
			if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return err
			}
			inFile, err := f.Open()
			if err != nil {
				return err
			}
			defer inFile.Close()
			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer outFile.Close()
			_, err = io.Copy(outFile, inFile)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func OutputFile(r io.Reader, file string) error {
	os.MkdirAll(path.Dir(file), 0666)
	dest, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	_, err = io.Copy(dest, r)
	return err
}

// SaveFileByRequest save file by request
func SaveFileByRequest(request *http.Request, target string) error {
	os.MkdirAll(path.Dir(target), 0666)
	request.ParseMultipartForm(65535)
	if files, ok := request.MultipartForm.File["file"]; ok {
		if len(files) == 0 {
			return errors.New("file is empty")
		}
		file := files[0]
		src, err := file.Open()
		if err != nil {
			return err
		}
		defer src.Close()
		dest, err := os.OpenFile(target, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			return err
		}
		defer dest.Close()
		_, err = io.Copy(dest, src)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteDirElseSelf delete dir/*
func DeleteDirElseSelf(dirName string) error {
	dir, err := ioutil.ReadDir(dirName)
	if err != nil {
		return err
	}
	for _, d := range dir {
		err = os.RemoveAll(path.Join(dirName, d.Name()))
		if err != nil {
			return err
		}
	}
	return nil
}

// HashCheck hash(name+secret+name) == value
func HashCheck(name, secret, value string) bool {
	return HashCalculation(md5.New(), name+secret+name) == value
}
