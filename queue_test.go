package gotools_test

import (
	"testing"

	"gitee.com/ivfzhou/gotools/v3"
)

func TestQueue(t *testing.T) {
	queue := &gotools.Queue[int]{}
	queue.Push(1)
	queue.Push(2)
	queue.Push(3)
	queue.Push(4)
	queue.Push(5)
	elem := <-queue.GetFromChan()
	if 1 != elem {
		t.Error("queue: queue elem does not match", elem)
	}
	elem = <-queue.GetFromChan()
	if 2 != elem {
		t.Error("queue: queue elem does not match", elem)
	}
	elem = <-queue.GetFromChan()
	if 3 != elem {
		t.Error("queue: queue elem does not match", elem)
	}
	elem = <-queue.GetFromChan()
	if 4 != elem {
		t.Error("queue: queue elem does not match", elem)
	}
	queue.Close()
	elem = <-queue.GetFromChan()
	if 5 != elem {
		t.Error("queue: queue elem does not match", elem)
	}

	queue.Push(6)
	queue.Push(7)
	elem = <-queue.GetFromChan()
	if 0 != elem {
		t.Error("queue: queue elem does not match", elem)
	}
	if 0 != elem {
		t.Error("queue: queue elem does not match", elem)
	}
}
