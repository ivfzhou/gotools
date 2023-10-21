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
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type WriteAtCloser interface {
	io.WriterAt
	io.Closer
}

type writeAtReader struct {
	tmpFile      *os.File
	recordWrite  sync.Map
	unread       int64
	nextOffset   int64
	writeClose   atomic.Int32
	writeErr     chan error
	writeErrOnce sync.Once
}

type writeClose struct {
	*writeAtReader
}

type readClose struct {
	*writeAtReader
}

// WriteAtReader 获取一个WriterAt和Reader对象，其中WriterAt用于并发写入数据，而与此同时Reader对象同时读取出已经写入好的数据。
// WriterAt写入完毕后调用Close，则Reader会全部读取完后结束读取。
// WriterAt发生的error会传递给Reader返回。
// 该接口是特定为一个目的实现————服务器分片下载数据中转给客户端下载，提高中转数据效率。
func WriteAtReader() (WriteAtCloser, io.ReadCloser) {
	temp, err := os.CreateTemp("", "ivfzhou_gotools_WriteAndRead_")
	if err != nil {
		panic(err)
	}
	wr := &writeAtReader{
		tmpFile:  temp,
		writeErr: make(chan error),
	}
	return &writeClose{wr}, &readClose{wr}
}

// CopyFile 复制文件。
func CopyFile(src, dest string) error {
	srcFile, err := os.OpenFile(src, os.O_RDONLY, 0400)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	destFile, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer destFile.Close()
	_, err = io.Copy(destFile, srcFile)
	return err
}

// NewMultiReader 依次从reader读出数据并写入writer中，并close reader。
// 返回send用于添加reader，readSize表示需要从reader读出的字节数，order用于表示记录读取序数并传递给writer，若读取字节数对不上则返回error。
// 返回wait用于等待所有reader读完，若读取发生error，wait返回该error，并结束读取。
// 务必等所有reader都已添加给send后再调用wait。
// 该函数可用于需要非同一时间多个读取流和一个写入流的工作模型。
func NewMultiReader(ctx context.Context, writer func(order int, p []byte)) (
	send func(readSize, order int, reader io.ReadCloser), wait func() error) {

	ctx, cancel := context.WithCancel(ctx)
	errCh := make(chan error, 1)
	once := &sync.Once{}
	waitOnce := &sync.Once{}
	wg := &sync.WaitGroup{}
	var err error
	return func(readSize, order int, reader io.ReadCloser) {
			wg.Add(1)
			go func() {
				defer wg.Done()
				select {
				case <-ctx.Done():
					once.Do(func() { errCh <- fmt.Errorf("context canceled: %v", ctx.Err()); close(errCh) })
					return
				default:
				}
				p := make([]byte, readSize)
				n, err := io.ReadFull(reader, p)
				if err != nil || n != len(p) {
					cancel()
					once.Do(func() {
						errCh <- fmt.Errorf("reading bytes occur error, want read size %d, actual size: %d: %v", len(p), n, err)
						close(errCh)
					})
				}
				_ = reader.Close()
				writer(order, p)
			}()
		},
		func() error {
			waitOnce.Do(func() {
				wg.Wait()
				select {
				case err = <-errCh:
				default:
					once.Do(func() { close(errCh); cancel() })
				}
			})
			return err
		}
}

func (wc *writeClose) Close() error {
	wc.writeErrOnce.Do(func() { close(wc.writeErr) })
	return nil
}

func (rc *readClose) Close() error {
	return os.Remove(rc.tmpFile.Name())
}

func (wr *writeAtReader) WriteAt(p []byte, off int64) (int, error) {
	wr.recordWrite.Store(off, int64(len(p)))
	n, err := wr.tmpFile.WriteAt(p, off)
	if err != nil {
		wr.writeErrOnce.Do(func() { wr.writeErr <- err })
	}
	return n, err
}

func (wr *writeAtReader) Read(p []byte) (int, error) {
	for {
		select {
		case err := <-wr.writeErr:
			if err != nil {
				return 0, err
			}
			return wr.tmpFile.Read(p)
		default:
		}

		pl := int64(len(p))
		if wr.unread >= pl {
			n, err := wr.tmpFile.Read(p)
			wr.unread -= pl
			return n, err
		}
		value, ok := wr.recordWrite.Load(wr.nextOffset)
		if ok {
			wr.recordWrite.Delete(wr.nextOffset)
			l, _ := value.(int64)
			wr.nextOffset += l
			wr.unread += l
			continue
		}

		if wr.unread > 0 {
			n, err := wr.tmpFile.Read(p[:wr.unread])
			wr.unread = 0
			return n, err
		}

		runtime.Gosched()
		time.Sleep(time.Second)
	}
}
