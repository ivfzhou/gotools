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
	"io"
	"math/rand"
	"sync"
	"testing"

	"gitee.com/ivfzhou/gotools"
)

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
