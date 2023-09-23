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
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"gitee.com/ivfzhou/gotools"
)

const jobCount = 256

func TestRunParallel(t *testing.T) {
	ch := make(chan int, 4)
	add, wait := gotools.RunParallel[int](4, func(i int) error {
		ch <- i
		return nil
	})
	wg := sync.WaitGroup{}
	for i := 0; i < jobCount; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			err := add(i)
			if err != nil {
				t.Error(err)
			}
		}(i)
	}
	time.Sleep(time.Millisecond * 100)
	l := len(ch)
	if l != 4 {
		t.Error("l is not 4", l)
		return
	}
	go func() { wg.Wait(); close(ch) }()
	i := 0
	for range ch {
		i++
	}
	if i != jobCount {
		t.Error("i is not", jobCount, i)
		return
	}
	err := wait()
	if err != nil {
		t.Error("err not nil", err)
		return
	}
}

func TestRunParallelErr(t *testing.T) {
	count := int32(0)
	add, wait := gotools.RunParallel[int](4, func(i int) error {
		if atomic.AddInt32(&count, 1) == 5 {
			return errors.New("is err")
		}
		return nil
	})
	var err1 error
	for i := 0; i < jobCount; i++ {
		go func(i int) {
			if err := add(i); err != nil {
				err1 = err
			}
		}(i)
	}
	time.Sleep(time.Millisecond * 100)
	err2 := wait()
	if err1 != err2 {
		t.Error("err not equal", err1, err2)
		return
	}
}

func TestRunParallelWait(t *testing.T) {
	count := int32(0)
	add, wait := gotools.RunParallel[int](4, func(i int) error {
		atomic.AddInt32(&count, 1)
		return nil
	})
	for i := 0; i < jobCount; i++ {
		go func(i int) { add(i) }(i)
	}
	time.Sleep(time.Millisecond * 100)
	wait()
	if count != jobCount {
		t.Error("count is not", jobCount, count)
		return
	}
}

func TestRunParallelPanic(t *testing.T) {
	count := int32(0)
	perr := errors.New("")
	add, wait := gotools.RunParallel[int](4, func(i int) error {
		if atomic.AddInt32(&count, 1) == 6 {
			panic(perr)
		}
		return nil
	})
	for i := 0; i < jobCount; i++ {
		go func(i int) { add(i) }(i)
	}
	time.Sleep(time.Millisecond * 100)
	err := wait()
	if err != perr {
		t.Error("err is not equal", err)
		return
	}
}

func ExampleRunParallel() {
	type product struct {
		// some stuff
	}
	op := func(data *product) error {
		// do something
		return nil
	}
	add, wait := gotools.RunParallel[*product](12, op)

	// many products
	var projects []*product
	for _, v := range projects {
		// blocked since number of ops running simultaneously reaches 12
		if err := add(v); err != nil {
			// means having a op return err
		}
	}

	// wait all op done and check err
	if err := wait(); err != nil {
		// op occur err
	}
}

func TestRunParallelNoBlock(t *testing.T) {
	ch := make(chan int, 4)
	wg := sync.WaitGroup{}
	wg.Add(jobCount)
	go func() { wg.Wait(); close(ch) }()
	add, wait := gotools.RunParallelNoBlock[int](4, func(i int) error {
		defer wg.Done()
		ch <- i
		return nil
	})
	for i := 0; i < jobCount; i++ {
		if err := add(i); err != nil {
			t.Error(err)
		}
	}
	time.Sleep(time.Millisecond * 100)
	l := len(ch)
	if l != 4 {
		t.Error("len is not 4", l)
		return
	}
	i := 0
	for range ch {
		i++
	}
	if i != jobCount {
		t.Error("i is not", jobCount, i)
		return
	}
	err := wait()
	if err != nil {
		t.Error("err not nil", err)
		return
	}
}

func TestRunParallelNoBlockErr(t *testing.T) {
	ch := make(chan int, 4)
	count := int32(0)
	add, wait := gotools.RunParallelNoBlock[int](4, func(i int) error {
		ch <- i
		if atomic.AddInt32(&count, 1) == 3 {
			return errors.New("is err")
		}
		return nil
	})
	var err1 error
	for i := 0; i < jobCount; i++ {
		err := add(i)
		if err != nil {
			err1 = err
		}
	}
	time.Sleep(time.Millisecond * 100)
	err2 := wait()
	if err1 != err2 {
		t.Error("err not equal", err1, err2)
		return
	}
}

func TestRunParallelNoBlockWait(t *testing.T) {
	count := int32(0)
	add, wait := gotools.RunParallelNoBlock[int](4, func(i int) error {
		atomic.AddInt32(&count, 1)
		return nil
	})
	for i := 0; i < jobCount; i++ {
		add(i)
	}
	wait()
	if count != jobCount {
		t.Error("count is not", jobCount, count)
		return
	}
}

func TestRunParallelNoBlockPanic(t *testing.T) {
	count := int32(0)
	err := errors.New("")
	add, wait := gotools.RunParallelNoBlock[int](4, func(i int) error {
		if atomic.AddInt32(&count, 1) == 3 {
			panic(err)
		}
		return nil
	})
	for i := 0; i < jobCount; i++ {
		add(i)
	}
	perr := wait()
	if err != perr {
		t.Error("err is not equal", perr)
		return
	}
}

func ExampleRunParallelNoBlock() {
	type product struct {
		// some stuff
	}
	op := func(data *product) error {
		// do something
		return nil
	}
	add, wait := gotools.RunParallelNoBlock[*product](12, op)

	// many products
	var projects []*product
	for _, v := range projects {
		// unblocked anywhere
		if err := add(v); err != nil {
			// means having a op return err
		}
	}

	// wait all op done and check err
	if err := wait(); err != nil {
		// op occur err
	}
}

func TestRunConcurrently(t *testing.T) {
	x := int32(0)
	fns := make([]func() error, jobCount)
	for i := 0; i < jobCount; i++ {
		fns[i] = func() error {
			atomic.AddInt32(&x, 1)
			return nil
		}
	}
	wait := gotools.RunConcurrently(fns...)
	err := wait()
	if err != nil {
		t.Error("err is not nil", err)
		return
	}
	if x != jobCount {
		t.Error("job count is unexpected", x)
		return
	}
}

func TestRunConcurrentlyErr(t *testing.T) {
	x := int32(0)
	fns := make([]func() error, jobCount)
	err := errors.New("")
	for i := 0; i < jobCount; i++ {
		fns[i] = func() error {
			if atomic.AddInt32(&x, 1) == 5 {
				return err
			}
			return nil
		}
	}
	wait := gotools.RunConcurrently(fns...)
	werr := wait()
	if err == nil {
		t.Error("err is nil", err)
		return
	}
	if werr != err {
		t.Error("err not equal", werr, err)
		return
	}
}

func TestRunConcurrentlyPanic(t *testing.T) {
	x := int32(0)
	fns := make([]func() error, jobCount)
	err := errors.New("")
	for i := 0; i < jobCount; i++ {
		fns[i] = func() error {
			if atomic.AddInt32(&x, 1) == 5 {
				panic(err)
			}
			return nil
		}
	}
	wait := gotools.RunConcurrently(fns...)
	werr := wait()
	if err == nil {
		t.Error("err is nil", err)
		return
	}
	if werr != err {
		t.Error("err not equal", werr, err)
		return
	}
}

func ExampleRunConcurrently() {
	var order any
	queryDB1 := func() error {
		// op order
		order = nil
		return nil
	}

	var stock any
	queryDB2 := func() error {
		// op stock
		stock = nil
		return nil
	}
	err := gotools.RunConcurrently(queryDB1, queryDB2)()
	// check err
	if err != nil {
		return
	}

	// do your want
	_ = order
	_ = stock
}

func TestRunSequently(t *testing.T) {
	x := 1
	err := gotools.RunSequently(func() error {
		if x != 1 {
			t.Error("x is not 1", x)
			return nil
		}
		x++
		return nil
	}, func() error {
		if x != 2 {
			t.Error("x is not 2", x)
			return nil
		}
		x++
		return nil
	})()
	if err != nil {
		t.Error("err is not nil", err)
		return
	}
	if x != 3 {
		t.Error("x is not 3", x)
		return
	}
}

func TestRunSequentlyErr(t *testing.T) {
	x := 0
	werr := errors.New("")
	err := gotools.RunSequently(func() error {
		x++
		return nil
	}, func() error {
		x++
		return werr
	}, func() error {
		x++
		return nil
	})()
	if err == nil {
		t.Error("err is nil", err)
		return
	}
	if x != 2 {
		t.Error("x is not 2", x)
		return
	}
}

func TestRunSequentlyPanic(t *testing.T) {
	x := 0
	werr := errors.New("")
	err := gotools.RunSequently(func() error {
		x++
		return nil
	}, func() error {
		x++
		panic(werr)
	}, func() error {
		x++
		return nil
	})()
	if err != werr {
		t.Error("err is not equaled", err, werr)
		return
	}
	if x != 2 {
		t.Error("x is not 2", x)
		return
	}
}

func ExampleRunSequently() {
	first := func() error { return nil }
	then := func() error { return nil }
	last := func() error { return nil }
	err := gotools.RunSequently(first, then, last)()
	if err != nil {
		// return err
	}
}

func TestStartProcess(t *testing.T) {
	type data struct {
		name string
		x    int
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	jobs := []*data{{"job1", 0}, {"job2", 0}}
	work1 := func(ctx context.Context, d *data) error {
		t.Logf("work1 job %s", d.name)
		if d.x == 0 {
			d.x = 1
			return nil
		} else {
			return errors.New("x != 0")
		}
	}
	work2 := func(ctx context.Context, d *data) error {
		t.Logf("work2 job %s", d.name)
		if d.x == 1 {
			d.x = 2
			return nil
		} else {
			return errors.New("x != 1")
		}
	}
	err := gotools.StartProcess(ctx, jobs, work1, work2)
	if err != nil {
		t.Error(err)
	}
	for _, v := range jobs {
		if v.x != 2 {
			t.Errorf("x != 2 %d", v.x)
		}
	}
}

func TestStartProcessPanic(t *testing.T) {
	type data struct {
		name string
		x    int
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	jobs := []*data{{"job1", 0}, {"job2", 0}}
	work1 := func(ctx context.Context, d *data) error {
		return nil
	}
	work2 := func(ctx context.Context, d *data) error {
		panic("panic work2 " + d.name)
	}
	err := gotools.StartProcess(ctx, jobs, work1, work2)
	if err == nil {
		t.Error("err is nil")
	}
	t.Log(err)
}

func ExampleStartProcess() {
	type data struct{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobs := []*data{{}, {}}
	work1 := func(ctx context.Context, d *data) error { return nil }
	work2 := func(ctx context.Context, d *data) error { return nil }
	err := gotools.StartProcess(ctx, jobs, work1, work2) // 等待运行结束
	if err != nil {
		// return err
	}
}
