package gotools_test

import (
	"os"
	"testing"

	"gitee.com/ivfzhou/gotools"
)

func TestZip(t *testing.T) {
	files := []string{
		`testdata/for_zip_test/b.txt`, `testdata/for_zip_test/a.txt`, `testdata/for_zip_test/dir/c.txt`,
	}
	bs, err := gotools.ZipFilesToBytes(files...)
	if err != nil {
		t.Error("unexpected error", err)
	}
	if len(bs) <= 0 {
		t.Error("unexpected bytes")
	}
	bs, _ = os.ReadFile(`testdata/for_zip_test/for_zip_test.zip`)
	filePaths, err := gotools.UnzipFromBytes(bs, `testdata/for_zip_test`)
	if err != nil {
		t.Error("unexpected error", err)
	}
	if len(filePaths) != 3 {
		t.Error("unexpected filePaths", filePaths)
	}
	for _, v := range filePaths {
		has := false
		for i := range files {
			if files[i] == v {
				has = true
				break
			}
		}
		if !has {
			t.Error("unexpected file", v)
		}
	}
}
