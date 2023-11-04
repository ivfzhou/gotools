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
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"gitee.com/ivfzhou/gotools/v2"
)

const jobCount = 256

func TestRunRunner(t *testing.T) {
	ctx := context.Background()
	chLen := 4
	ch := make(chan int, chLen)
	add, wait := gotools.NewRunner[int](ctx, 4, func(ctx context.Context, i int) error {
		ch <- i
		return nil
	})
	count := 0
	for i := 0; i < jobCount; i++ {
		count += i
		go func(i int) {
			if err := add(i, true); err != nil {
				t.Error(err)
			}
		}(i)
	}
	time.Sleep(time.Millisecond * 500)
	concurrentChLen := len(ch)
	if concurrentChLen != chLen {
		t.Errorf("concurrentChLen is %d not %d", concurrentChLen, chLen)
		return
	}
	go func() {
		times := 0
		for i := range ch {
			count -= i
			times++
		}
		if times != jobCount {
			t.Error("count is not", jobCount, count)
		}
		if count != 0 {
			t.Error("count is not zero", count)
		}
	}()
	err := wait()
	if err != nil {
		t.Error("err not nil", err)
		return
	}
	close(ch)
}

func TestRunRunnerErr(t *testing.T) {
	ctx := context.Background()
	count := int32(0)
	add, wait := gotools.NewRunner[int](ctx, 4, func(ctx context.Context, i int) error {
		if atomic.AddInt32(&count, 1) == 5 {
			return errors.New("is err")
		}
		return nil
	})
	var err1 error
	for i := 0; i < jobCount; i++ {
		go func(i int) {
			if err := add(i, true); err != nil {
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

func TestRunRunnerWait(t *testing.T) {
	ctx := context.Background()
	count := int32(0)
	add, wait := gotools.NewRunner[int](ctx, 4, func(ctx context.Context, i int) error {
		atomic.AddInt32(&count, 1)
		return nil
	})
	for i := 0; i < jobCount; i++ {
		go func(i int) { add(i, true) }(i)
	}
	time.Sleep(time.Millisecond * 100)
	wait()
	if count != jobCount {
		t.Error("count is not", jobCount, count)
		return
	}
}

func TestRunRunnerPanic(t *testing.T) {
	ctx := context.Background()
	count := int32(0)
	perr := errors.New("")
	add, wait := gotools.NewRunner[int](ctx, 4, func(ctx context.Context, i int) error {
		if atomic.AddInt32(&count, 1) == 6 {
			panic(perr)
		}
		return nil
	})
	for i := 0; i < jobCount; i++ {
		go func(i int) { add(i, true) }(i)
	}
	time.Sleep(time.Millisecond * 100)
	err := wait()
	if err != perr {
		// t.Error("err is not equal", err)
		return
	}
}

func ExampleNewRunner() {
	type product struct {
		// some stuff
	}
	ctx := context.Background()
	op := func(ctx context.Context, data *product) error {
		// do something
		return nil
	}
	add, wait := gotools.NewRunner[*product](ctx, 12, op)

	// many products
	var projects []*product
	for _, v := range projects {
		// blocked since number of ops running simultaneously reaches 12
		if err := add(v, true); err != nil {
			// means having a op return err
		}
	}

	// wait all op done and check err
	if err := wait(); err != nil {
		// op occur err
	}
}

func TestRunRunnerNoBlock(t *testing.T) {
	ctx := context.Background()
	chLen := 4
	ch := make(chan int, chLen)
	add, wait := gotools.NewRunner[int](ctx, 4, func(ctx context.Context, i int) error {
		ch <- i
		return nil
	})
	count := 0
	for i := 0; i < jobCount; i++ {
		count += i
		go func(i int) {
			if err := add(i, false); err != nil {
				t.Error(err)
			}
		}(i)
	}
	time.Sleep(time.Millisecond * 500)
	concurrentChLen := len(ch)
	if concurrentChLen != chLen {
		t.Errorf("concurrentChLen is %d not %d", concurrentChLen, chLen)
		return
	}
	go func() {
		times := 0
		for i := range ch {
			count -= i
			times++
		}
		if times != jobCount {
			t.Error("count is not", jobCount, count)
		}
		if count != 0 {
			t.Error("count is not zero", count)
		}
	}()
	err := wait()
	if err != nil {
		t.Error("err not nil", err)
		return
	}
	close(ch)
}

func TestRunRunnerNoBlockErr(t *testing.T) {
	ctx := context.Background()
	count := int32(0)
	add, wait := gotools.NewRunner[int](ctx, 4, func(ctx context.Context, i int) error {
		if atomic.AddInt32(&count, 1) == 3 {
			return errors.New("is err")
		}
		return nil
	})
	var err1 error
	for i := 0; i < jobCount; i++ {
		err := add(i, false)
		if err != nil {
			err1 = err
		}
	}
	time.Sleep(time.Millisecond * 100)
	err2 := wait()
	if err2 != err2 {
		t.Error("err not equal", err1, err2)
		return
	}
}

func TestRunRunnerNoBlockWait(t *testing.T) {
	ctx := context.Background()
	count := int32(0)
	add, wait := gotools.NewRunner[int](ctx, 4, func(ctx context.Context, i int) error {
		atomic.AddInt32(&count, 1)
		return nil
	})
	for i := 0; i < jobCount; i++ {
		add(i, false)
	}
	wait()
	if count != jobCount {
		t.Error("count is not", jobCount, count)
		return
	}
}

func TestRunRunnerNoBlockPanic(t *testing.T) {
	ctx := context.Background()
	count := int32(0)
	err := errors.New("")
	add, wait := gotools.NewRunner[int](ctx, 4, func(ctx context.Context, i int) error {
		if atomic.AddInt32(&count, 1) == 3 {
			panic(err)
		}
		return nil
	})
	for i := 0; i < jobCount; i++ {
		add(i, false)
	}
	perr := wait()
	if !strings.Contains(perr.Error(), err.Error()) {
		t.Error("err is not equal", perr)
		return
	}
}

func TestRunConcurrently(t *testing.T) {
	ctx := context.Background()
	x := int32(0)
	fns := make([]func(ctx context.Context) error, jobCount)
	for i := 0; i < jobCount; i++ {
		fns[i] = func(ctx context.Context) error {
			atomic.AddInt32(&x, 1)
			return nil
		}
	}
	wait := gotools.RunConcurrently(ctx, fns...)
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
	ctx := context.Background()
	x := int32(0)
	fns := make([]func(ctx context.Context) error, jobCount)
	err := errors.New("")
	for i := 0; i < jobCount; i++ {
		fns[i] = func(ctx context.Context) error {
			if atomic.AddInt32(&x, 1) == 5 {
				return err
			}
			return nil
		}
	}
	wait := gotools.RunConcurrently(ctx, fns...)
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
	ctx := context.Background()
	x := int32(0)
	fns := make([]func(ctx context.Context) error, jobCount)
	err := errors.New("")
	for i := 0; i < jobCount; i++ {
		fns[i] = func(ctx context.Context) error {
			if atomic.AddInt32(&x, 1) == 5 {
				panic(err)
			}
			return nil
		}
	}
	wait := gotools.RunConcurrently(ctx, fns...)
	werr := wait()
	if err == nil {
		t.Error("err is nil", err)
		return
	}
	if !strings.Contains(werr.Error(), err.Error()) {
		t.Error("err not equal", werr, err)
		return
	}
}

func ExampleRunConcurrently() {
	ctx := context.Background()
	var order any
	queryDB1 := func(ctx context.Context) error {
		// op order
		order = nil
		return nil
	}

	var stock any
	queryDB2 := func(ctx context.Context) error {
		// op stock
		stock = nil
		return nil
	}
	err := gotools.RunConcurrently(ctx, queryDB1, queryDB2)()
	// check err
	if err != nil {
		return
	}

	// do your want
	_ = order
	_ = stock
}

func TestRunSequentially(t *testing.T) {
	ctx := context.Background()
	x := 1
	err := gotools.RunSequentially(ctx, func(ctx context.Context) error {
		if x != 1 {
			t.Error("x is not 1", x)
			return nil
		}
		x++
		return nil
	}, func(ctx context.Context) error {
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

func TestRunSequentiallyErr(t *testing.T) {
	ctx := context.Background()
	x := 0
	werr := errors.New("")
	err := gotools.RunSequentially(ctx, func(ctx context.Context) error {
		x++
		return nil
	}, func(ctx context.Context) error {
		x++
		return werr
	}, func(ctx context.Context) error {
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

func TestRunSequentiallyPanic(t *testing.T) {
	ctx := context.Background()
	x := 0
	werr := errors.New("test error")
	err := gotools.RunSequentially(ctx, func(ctx context.Context) error {
		x++
		return nil
	}, func(ctx context.Context) error {
		x++
		panic(werr)
	}, func(ctx context.Context) error {
		x++
		return nil
	})()
	if !strings.Contains(err.Error(), werr.Error()) {
		t.Error("err is not equaled", err, werr)
		return
	}
	if x != 2 {
		t.Error("x is not 2", x)
		return
	}
}

func ExampleRunSequentially() {
	ctx := context.Background()
	first := func(context.Context) error { return nil }
	then := func(context.Context) error { return nil }
	last := func(context.Context) error { return nil }
	err := gotools.RunSequentially(ctx, first, then, last)()
	if err != nil {
		// return err
	}
}

func TestRunPipeline(t *testing.T) {
	type data struct {
		name string
		x    int
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	jobs := []*data{{"job1", 0}, {"job2", 0}}
	work1 := func(ctx context.Context, d *data) error {
		// t.Logf("work1 job %s", d.name)
		if d.x == 0 {
			d.x = 1
			return nil
		} else {
			return errors.New("x != 0")
		}
	}
	work2 := func(ctx context.Context, d *data) error {
		// t.Logf("work2 job %s", d.name)
		if d.x == 1 {
			d.x = 2
			return nil
		} else {
			return errors.New("x != 1")
		}
	}
	err := gotools.RunPipeline(ctx, jobs, work1, work2)
	if err != nil {
		t.Error(err)
	}
	for _, v := range jobs {
		if v.x != 2 {
			t.Errorf("x != 2 %d", v.x)
		}
	}
}

func TestRunPipelinePanic(t *testing.T) {
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
	err := gotools.RunPipeline(ctx, jobs, work1, work2)
	if err == nil {
		t.Error("err is nil")
	}
}

func ExampleRunPipeline() {
	type data struct{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobs := []*data{{}, {}}
	work1 := func(ctx context.Context, d *data) error { return nil }
	work2 := func(ctx context.Context, d *data) error { return nil }
	err := gotools.RunPipeline(ctx, jobs, work1, work2) // 等待运行结束
	if err != nil {
		// return err
	}
}

func TestListenChan(t *testing.T) {
	err1 := make(chan error, 1)
	err2 := make(chan error, 1)
	err3 := make(chan error, 1)
	errChans := []<-chan error{
		err1,
		err2,
		err3,
	}

	err := errors.New("")
	go func() {
		time.Sleep(10 * time.Second)
		err1 <- err
	}()
	if err != <-gotools.ListenChan(errChans...) {
		t.Error("err not equal")
		return
	}

	go func() {
		time.Sleep(3 * time.Second)
		close(err1)
		close(err2)
		close(err3)
	}()
	if nil != <-gotools.ListenChan(errChans...) {
		t.Error("err is not nil")
	}
}

func TestRun(t *testing.T) {
	ctx := context.Background()
	count := int32(0)
	err := gotools.Run(ctx, func(ctx context.Context, t int32) error {
		atomic.AddInt32(&count, t)
		return nil
	}, 1, 2, 3, 4)
	if err != nil {
		t.Error("unexpected error", err)
	}
	if count != 10 {
		t.Error("unexpected count", count)
	}

	expectedErr := errors.New("")
	err = gotools.Run(ctx, func(ctx context.Context, t int32) error {
		if t == 3 {
			return expectedErr
		}
		return nil
	}, 1, 2, 3, 4)
	if err != expectedErr {
		t.Error("unexpected error", err)
	}

	err = gotools.Run(ctx, func(ctx context.Context, t int32) error {
		if t == 3 {
			panic(t)
		}
		return nil
	}, 1, 2, 3, 4)
	if err == nil || !strings.Contains(err.Error(), "3") {
		t.Error("unexpected error", err)
	}
}

func TestRunPeriodically(t *testing.T) {
	add := gotools.RunPeriodically(time.Second)
	var now time.Time
	add(func() {
		time.Sleep(3 * time.Second)
		now = time.Now()
	})
	add(func() {
		if time.Since(now) < time.Second {
			t.Error("unexpected time", time.Now())
		}
		time.Sleep(3 * time.Second)
		now = time.Now()
	})
	add(func() {
		if time.Since(now) < time.Second {
			t.Error("unexpected time", time.Now())
		}
		now = time.Now()
	})
	add(func() {
		if time.Since(now) < time.Second {
			t.Error("unexpected time", time.Now())
		}
	})
}
