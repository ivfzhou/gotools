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
	"io/ioutil"
	"os"
	"path/filepath"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// UnzipFromBytes 将压缩数据解压写入硬盘。
func UnzipFromBytes(zippedBytes []byte, toDir string) (filePaths []string, err error) {
	toDir = filepath.Clean(toDir)
	if err = os.MkdirAll(toDir, 0755); err != nil {
		return nil, err
	}

	reader, err := zip.NewReader(bytes.NewReader(zippedBytes), int64(len(zippedBytes)))
	if err != nil {
		return nil, err
	}
	filePaths = make([]string, 0, len(reader.File))
	var (
		r    io.ReadCloser
		file *os.File
	)
	for _, v := range reader.File {
		if v.FileInfo().IsDir() {
			continue
		}

		name := filepath.Join(toDir, string(gbk2Utf8([]byte(v.FileHeader.Name))))
		if err = os.MkdirAll(filepath.Dir(name), 0755); err != nil {
			return nil, err
		}

		if r, err = v.Open(); err != nil {
			return nil, err
		}

		file, err = os.OpenFile(name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			CloseIO(r)
			return nil, err
		}

		_, err = io.Copy(file, r)
		if err != nil {
			CloseIO(r)
			CloseIO(file)
			return nil, err
		}

		filePaths = append(filePaths, name)
	}

	return filePaths, nil
}

// UnzipFromFiles 解压并写入硬盘。
func UnzipFromFiles(zippedFilePath string, toDir string) (filePaths []string, err error) {
	toDir = filepath.Clean(toDir)

	// 创建文件夹
	if err = os.MkdirAll(toDir, 0755); err != nil {
		return nil, err
	}

	// 打开流
	reader, err := zip.OpenReader(zippedFilePath)
	if err != nil {
		return nil, err
	}
	filePaths = make([]string, 0, len(reader.File))

	var (
		r io.ReadCloser
		w *os.File
	)
	for _, v := range reader.File {
		if v.FileInfo().IsDir() {
			continue
		}

		fileName := string(gbk2Utf8([]byte(v.FileHeader.Name)))
		fileName = filepath.Clean(filepath.Join(toDir, fileName))
		if err = os.MkdirAll(filepath.Dir(fileName), 0755); err != nil {
			return nil, err
		}

		// 打开zip流
		r, err = v.Open()
		if err != nil {
			return nil, err
		}

		// 打开磁盘文件流
		w, err = os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}

		// 写入磁盘
		if _, err = io.Copy(w, r); err != nil {
			CloseIO(r)
			CloseIO(w)
			return nil, err
		}

		CloseIO(r)
		CloseIO(w)

		filePaths = append(filePaths, fileName)
	}

	return filePaths, nil
}

// ZipFilesToBytes 将文件打成压缩包。
func ZipFilesToBytes(filePaths ...string) ([]byte, error) {
	buf := bytes.Buffer{}
	writer := zip.NewWriter(&buf)
	var (
		err  error
		w    io.Writer
		file *os.File
	)
	for _, v := range filePaths {
		if w, err = writer.Create(filepath.Base(v)); err != nil {
			return nil, err
		}

		file, err = os.Open(v)
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(w, file)
		if err != nil {
			CloseIO(file)
			return nil, err
		}
		CloseIO(file)
	}
	if err = writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ZipFiles 加压并写入硬盘。
func ZipFiles(toZippedFilePath string, fromDir string) error {
	// 找出所有文件
	files := make(map[string]string, 10)
	var seekFile func(p string) error
	seekFile = func(p string) error {
		fileInfo, err := os.Stat(p)
		if err != nil {
			return err
		}
		if !fileInfo.IsDir() {
			files[p], _ = filepath.Rel(fromDir, p)
			return nil
		}
		dir, err := ioutil.ReadDir(p)
		if err != nil {
			return err
		}
		for i := range dir {
			if err = seekFile(filepath.Join(p, dir[i].Name())); err != nil {
				return err
			}
		}
		return nil
	}
	err := seekFile(fromDir)
	if err != nil {
		return err
	}

	// 打开磁盘文件流
	file, err := os.OpenFile(toZippedFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer CloseIO(file)

	// 加压写入
	writer := zip.NewWriter(file)
	defer CloseIO(writer)
	var (
		w io.Writer
		r *os.File
	)
	for fullPath, relativePath := range files {
		w, err = writer.Create(relativePath)
		if err != nil {
			return err
		}

		r, err = os.Open(fullPath)
		if err != nil {
			return err
		}

		if _, err = io.Copy(w, r); err != nil {
			CloseIO(r)
			return err
		}
		CloseIO(r)
	}

	return nil
}

func gbk2Utf8(bs []byte) []byte {
	reader := transform.NewReader(bytes.NewReader(bs), simplifiedchinese.GBK.NewDecoder())
	res, err := io.ReadAll(reader)
	if err != nil {
		return bs
	}
	return res
}
