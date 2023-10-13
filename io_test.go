package gotools_test

import (
	"bytes"
	"io"
	"math/rand"
	"os"
	"sync"
	"testing"

	"gitee.com/ivfzhou/gotools"
)

func TestWriteAtReader(t *testing.T) {
	testFile := `testdata/go1.21.1.linux-amd64.tar.gz`
	writeAtCloser, readCloser := gotools.WriteAtReader()
	wg := &sync.WaitGroup{}

	file, err := os.ReadFile(testFile)
	if err != nil {
		t.Error(err)
		return
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		fileSize := int64(len(file))
		gap := int64(1024 * 1024)
		frags := fileSize / gap
		m := make(map[int64]int64, frags+1)
		if residue := fileSize % gap; residue != 0 {
			m[frags] = residue
		}
		for i := frags - 1; i >= 0; i-- {
			m[i] = gap
		}
		defer writeAtCloser.Close()
		for len(m) > 0 {
			for {
				off := int64(rand.Intn(int(frags + 1)))
				f, ok := m[off]
				if ok {
					delete(m, off)
					go func() {
						of := off * gap
						n, err := writeAtCloser.WriteAt(file[of:of+f], of)
						if err != nil {
							t.Error(err)
						}
						if int64(n) != f {
							t.Error("n!=f", n, f)
						}
					}()
					break
				}
			}
		}
	}()

	wg.Add(1)
	buf := &bytes.Buffer{}
	go func() {
		defer wg.Done()
		defer readCloser.Close()
		_, err := io.Copy(buf, readCloser)
		if err != nil {
			t.Error(err)
		}
	}()

	wg.Wait()

	res := buf.Bytes()
	l1 := len(res)
	l2 := l1 != len(file)
	if l2 {
		t.Error("length mot match", l1, len(file))
	}
	for i, v := range res {
		if file[i] != v {
			t.Error("not match", file[i], v, i)
		}
	}
}

func ExampleWriteAtReader() {
	writeAtCloser, readCloser := gotools.WriteAtReader()

	// 并发写入数据
	go func() {
		writeAtCloser.WriteAt(nil, 0)
	}()
	// 写完close
	writeAtCloser.Close()

	// 同时读出写入的数据
	go func() {
		readCloser.Read(nil)
	}()
	// 读完close
	readCloser.Close()
}
