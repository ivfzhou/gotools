package gotools_test

import (
	"sync"
	"testing"

	"github.com/ivfzhou/gotools"
)

func TestFailLocker(t *testing.T) {
	fairLocker := &gotools.FairLocker{}
	count := 5
	writer := func() {
		fairLocker.WLock()
		defer fairLocker.WUnlock()
		t.Log("writing")
		count--
	}
	reader := func() {
		fairLocker.RLock()
		defer fairLocker.RUnlock()
		t.Log("reading count", count)
	}
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			reader()
		}()
	}
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			writer()
		}()
	}
	wg.Wait()
	t.Log("count:", count)
}

func TestReadFirstLocker(t *testing.T) {
	fairLocker := &gotools.ReadFirstLocker{}
	count := 5
	writer := func() {
		fairLocker.WLock()
		defer fairLocker.WUnlock()
		t.Log("writing")
		count--
	}
	reader := func() {
		fairLocker.RLock()
		defer fairLocker.RUnlock()
		t.Log("reading count", count)
	}
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			reader()
		}()
	}
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			writer()
		}()
	}
	wg.Wait()
	t.Log("count:", count)
}

func TestWriteFirstLocker(t *testing.T) {
	fairLocker := &gotools.WriteFirstLocker{}
	count := 500
	writer := func() {
		fairLocker.WLock()
		defer fairLocker.WUnlock()
		t.Log("writing")
		count--
	}
	reader := func() {
		fairLocker.RLock()
		defer fairLocker.RUnlock()
		t.Log("reading count", count)
	}
	wg := sync.WaitGroup{}
	for i := 0; i < 500; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			writer()
		}()
	}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			reader()
		}()
	}
	wg.Wait()
	t.Log("count:", count)
}
