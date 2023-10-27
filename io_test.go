/*
 * Copyright (c) 2023 ivfzhou
 * gotools is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

package gotools_test

import (
	"bytes"
	"context"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"gitee.com/ivfzhou/gotools"
)

type bytesReader struct {
	*bytes.Reader
	closed int
}

func (r *bytesReader) Close() error { r.closed++; return nil }

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

func TestWriteAtReader(t *testing.T) {
	writeAtCloser, readCloser := gotools.WriteAtReader()
	wg := &sync.WaitGroup{}

	file := make([]byte, 1024*1024*20+10)
	for i := range file {
		file[i] = byte(rand.Intn(257))
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

func TestNewMultiReader(t *testing.T) {
	readerNum := 10000
	length := 10
	m := make(map[int]struct{}, readerNum)
	lock := &sync.Mutex{}
	writer := func(order int, p []byte) {
		if len(p) != 10 {
			t.Error("p len not match", len(p))
		}
		s := string(p)
		if s != "0123456789" {
			t.Error("bytes unexpected", s)
		}
		lock.Lock()
		_, ok := m[order]
		if ok {
			t.Error("unexpected order", order)
		}
		m[order] = struct{}{}
		lock.Unlock()
	}
	ctx := context.Background()
	rs := make([]*bytesReader, readerNum)
	send, wait := gotools.NewMultiReader(ctx, writer)
	for i := 0; i < readerNum; i++ {
		reader := &bytesReader{Reader: bytes.NewReader([]byte("0123456789"))}
		rs[i] = reader
		send(length, i, reader)
	}
	err := wait()
	if err != nil {
		t.Error("don't want error", err)
	}
	for i := range rs {
		if rs[i].closed != 1 {
			t.Error("reader doesn't close", rs[i].closed)
		}
	}
	count := len(m)
	if count != readerNum {
		t.Error("reader numbers not match", count)
	}
}

func TestCopyFile(t *testing.T) {
	fileSize := int64(13)
	dest := `testdata/copyfile_test`
	err := gotools.CopyFile(`testdata/copyfile`, dest)
	if err != nil {
		t.Error("unexpected error", err)
	}
	info, err := os.Stat(dest)
	if err != nil {
		t.Error("unexpected error", err)
	}
	if info.IsDir() {
		t.Error("not a file")
	}
	if info.Name() != filepath.Base(dest) {
		t.Error("file name mistake", info.Name())
	}
	if info.Size() != fileSize {
		t.Error("file size mistake", info.Size())
	}
	_ = os.Remove(dest)
}

func TestMultiReadCloser(t *testing.T) {
	ctx := context.Background()
	length := 80*1024*1024 + rand.Intn(1025)
	data := make([]byte, length)
	for i := 0; i < length; i++ {
		data[i] = byte(rand.Intn(255))
	}
	rcs := make([]io.ReadCloser, rand.Intn(5))
	l := 1 + rand.Intn(length/2)
	residue := length - l
	begin := 0
	for i := range rcs {
		rcs[i] = &bytesReader{Reader: bytes.NewReader(data[begin : begin+l])}
		begin += l
		l = rand.Intn(residue / 2)
		residue -= l
	}
	r, add, endAdd := gotools.MultiReadCloser(ctx, rcs...)
	go func() {
		next := true
		for next {
			rc := &bytesReader{Reader: bytes.NewReader(data[begin : begin+l])}
			begin += l
			if residue == 0 {
				next = false
			} else {
				l = 1 + rand.Intn(residue)
				residue -= l
			}
			rcs = append(rcs, rc)
			if rand.Intn(3) <= 0 {
				time.Sleep(time.Millisecond * time.Duration(100+rand.Intn(5000)))
			}
			if err := add(rc); err != nil {
				t.Error("unexpected error", err)
			}
		}
		endAdd()
	}()
	bs, err := io.ReadAll(r)
	if err != nil {
		t.Error("unexpected error", err)
	}
	if len(data) != len(bs) {
		t.Error("bytes length not match", len(data), len(bs))
	} else if bytes.Compare(data, bs) != 0 {
		t.Error("bytes does not compared")
	}
	for i := range rcs {
		if rcs[i].(*bytesReader).closed <= 0 {
			t.Error("reader not closed", i)
		}
	}
}
