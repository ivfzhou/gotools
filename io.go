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
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"runtime"
	"sync"
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
	writeErr     chan error
	writeErrOnce sync.Once
}

type writeClose struct {
	*writeAtReader
}

type readClose struct {
	*writeAtReader
}

type multiReader struct {
	err      error
	ctx      context.Context
	rcChan   chan io.ReadCloser
	done     chan struct{}
	doneOnce sync.Once
	curRc    io.ReadCloser
	cancel   func()
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
	defer CloseIO(srcFile)
	destFile, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer CloseIO(destFile)
	_, err = io.Copy(destFile, srcFile)
	return err
}

// CloseIO 调用c.Close()。
func CloseIO(c io.Closer) {
	if c != nil {
		err := c.Close()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
		}
	}
}

// MultiReadCloser 依次从rc中读出数据直到io.EOF则close rc。从r获取rc中读出的数据。
// add添加rc，返回error表明读取rc发生错误，可以安全的添加nil。调用endAdd表明不会再有rc添加，当所有数据读完了时，r将返回EOF。
// 如果ctx被cancel，将停止读取并返回error。
// 所有添加进去的io.ReadCloser都会被close。
func MultiReadCloser(ctx context.Context, rc ...io.ReadCloser) (r io.Reader, add func(rc io.ReadCloser) error, endAdd func()) {
	rcChan := make(chan io.ReadCloser, int(math.Max(float64(len(rc)), 256)))
	ctx, cancel := context.WithCancel(ctx)
	cancelWrapper := func() {
		cancel()
		go func() {
			for closer := range rcChan {
				CloseIO(closer)
			}
		}()
	}
	mr := &multiReader{
		ctx:    ctx,
		rcChan: rcChan,
		done:   make(chan struct{}),
		cancel: cancelWrapper,
	}
	for i := range rc {
		if rc[i] == nil {
			continue
		}
		mr.rcChan <- rc[i]
	}
	return mr,
		func(rc io.ReadCloser) error {
			if rc == nil {
				return nil
			}
			select {
			case <-mr.done:
				CloseIO(rc)
				return mr.err
			case <-mr.ctx.Done():
				CloseIO(rc)
				if mr.err != nil {
					return mr.err
				}
				return mr.ctx.Err()
			default:
				mr.rcChan <- rc
				return mr.err
			}
		}, func() { mr.doneOnce.Do(func() { close(mr.done) }) }
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

func (r *multiReader) Read(p []byte) (int, error) {
	if r.err != nil {
		return 0, r.err
	}

	rc := r.curRc
	var ok bool
	for rc == nil {
		select {
		case <-r.ctx.Done():
			if r.err == nil {
				r.err = r.ctx.Err()
			}
			return 0, r.err
		case rc, ok = <-r.rcChan:
			if !ok {
				if r.err == nil {
					r.err = io.EOF
				}
				r.cancel()
				return 0, r.err
			}
		case <-r.done:
			close(r.rcChan)
			r.done = nil
		}
	}
	if r.err != nil {
		CloseIO(rc)
		return 0, r.err
	}

	l, err := rc.Read(p)
	if errors.Is(err, io.EOF) {
		CloseIO(rc)
		r.curRc = nil
		if l <= 0 {
			return r.Read(p)
		}
		return l, nil
	}
	if err == nil {
		r.curRc = rc
		return l, nil
	}

	r.err = err
	r.cancel()
	return 0, err
}
