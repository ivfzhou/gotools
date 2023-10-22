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

package gotools

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// UnzipFromBytes 将压缩数据解压写入硬盘。
func UnzipFromBytes(bs []byte, parentPath string) (filePaths []string, err error) {
	parentPath = filepath.Clean(parentPath)
	if err = os.MkdirAll(parentPath, 0700); err != nil {
		return nil, err
	}

	reader, err := zip.NewReader(bytes.NewReader(bs), int64(len(bs)))
	if err != nil {
		return nil, err
	}
	filePaths = make([]string, 0, len(reader.File))
	var r io.ReadCloser
	for _, v := range reader.File {
		name := filepath.Join(parentPath, string(gbk2Utf8([]byte(v.FileHeader.Name))))
		if v.FileInfo().IsDir() {
			continue
		}
		if err = os.MkdirAll(filepath.Dir(name), 0700); err != nil {
			return nil, err
		}
		filePaths = append(filePaths, name)
		if r, err = v.Open(); err != nil {
			return nil, err
		}
		if bs, err = io.ReadAll(r); err != nil {
			_ = r.Close()
			return nil, err
		}
		if err = os.WriteFile(name, bs, 0600); err != nil {
			return nil, err
		}
	}

	return filePaths, nil
}

func ZipFilesToBytes(files ...string) ([]byte, error) {
	buf := bytes.Buffer{}
	writer := zip.NewWriter(&buf)
	var (
		err  error
		w    io.Writer
		data []byte
		l    int
	)
	for _, v := range files {
		if w, err = writer.Create(filepath.Base(v)); err != nil {
			return nil, err
		}
		if data, err = os.ReadFile(v); err != nil {
			return nil, err
		}
		for len(data) > 0 {
			if l, err = w.Write(data); err != nil {
				return nil, err
			}
			data = data[l:]
		}
	}
	if err = writer.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func gbk2Utf8(bs []byte) []byte {
	reader := transform.NewReader(bytes.NewReader(bs), simplifiedchinese.GBK.NewDecoder())
	res, err := io.ReadAll(reader)
	if err != nil {
		return bs
	}
	return res
}
