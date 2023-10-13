package gotools

import (
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
