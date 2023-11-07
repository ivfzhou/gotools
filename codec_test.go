package gotools_test

import (
	"os"
	"path/filepath"
	"testing"

	"gitee.com/ivfzhou/gotools/v3"
)

func TestZip(t *testing.T) {
	files := []string{
		`testdata/for_zip_test/b.txt`, `testdata/for_zip_test/a.txt`, `testdata/for_zip_test/dir/c.txt`,
	}
	bs, err := gotools.ZipFilesToBytes(files...)
	if err != nil {
		t.Error("codec: failed to zip file", err)
	}
	if len(bs) <= 0 {
		t.Error("codec: zip bytes can not empty")
	}
	bs, _ = os.ReadFile(`testdata/for_zip_test/for_zip_test.zip`)
	filePaths, err := gotools.UnzipFromBytes(bs, `testdata/for_zip_test`)
	if err != nil {
		t.Error("codec: failed to unzip file", err)
	}
	if len(filePaths) != 3 {
		t.Error("codec: unexpected filePaths number, expected 3 but give", len(filePaths))
	}
	for _, v := range files {
		has := false
		for i := range filePaths {
			if filepath.Base(filePaths[i]) == filepath.Base(v) {
				has = true
				break
			}
		}
		if !has {
			t.Error("codec: unexpected filePath", v)
		}
	}

	filePaths, err = gotools.UnzipFromFiles(`testdata/for_zip_test/for_zip_test.zip`, `testdata/for_zip_test/test`)
	if err != nil {
		t.Error("codec: failed to unzip file", err)
	}
	for _, v := range files {
		has := false
		for i := range filePaths {
			if filepath.Base(filePaths[i]) == filepath.Base(v) {
				has = true
				break
			}
		}
		if !has {
			t.Error("codec: unexpected filePath", v)
		}
	}
	err = gotools.ZipFiles(`testdata/test.zip`, `testdata/for_zip_test/test`)
	if err != nil {
		t.Error("codec: failed to zip file", err)
	}
	_ = os.Remove(`testdata/test.zip`)
	_ = os.RemoveAll(`testdata/for_zip_test/test`)
}
