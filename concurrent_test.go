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

func ExampleRunConcurrently() {
	ctx := context.Background()
	var order any
	work1 := func(ctx context.Context) error {
		// op order
		order = nil
		return nil
	}

	var stock any
	work2 := func(ctx context.Context) error {
		// op stock
		stock = nil
		return nil
	}
	err := gotools.RunConcurrently(ctx, work1, work2)()
	// check err
	if err != nil {
		return
	}

	// do your want
	_ = order
	_ = stock
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

		// no block
		if err := add(v, false); err != nil {
			// means having a op return err
		}
	}

	// wait all op done and check err
	if err := wait(); err != nil {
		// op occur err
	}
}

func ExampleRunPipeline() {
	type data struct{}
	ctx := context.Background()

	jobs := []*data{{}, {}}
	work1 := func(ctx context.Context, d *data) error { return nil }
	work2 := func(ctx context.Context, d *data) error { return nil }

	err := gotools.RunPipeline(ctx, jobs, work1, work2) // wait done
	if err != nil {
		// return err
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
		t.Error("concurrent: err is not nil", err)
	}
	if x != jobCount {
		t.Error("concurrent: job count is unexpected", x)
	}
}

func TestRunConcurrentlyErr(t *testing.T) {
	ctx := context.Background()
	x := int32(0)
	fns := make([]func(ctx context.Context) error, jobCount)
	err := errors.New("expected error")
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
		t.Error("concurrent: err is nil", err)
	}
	if werr != err {
		t.Error("concurrent: err not equal", werr, err)
	}
}

func TestRunConcurrentlyPanic(t *testing.T) {
	ctx := context.Background()
	x := int32(0)
	fns := make([]func(ctx context.Context) error, jobCount)
	err := errors.New("expected error")
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
	if werr == nil {
		t.Error("concurrent: err is nil", werr)
	}
	if werr != nil && !strings.Contains(werr.Error(), err.Error()) {
		t.Error("concurrent: err not equal", werr, err)
	}
}

func TestRunSequentially(t *testing.T) {
	ctx := context.Background()
	x := 1
	err := gotools.RunSequentially(ctx, func(ctx context.Context) error {
		if x != 1 {
			t.Error("concurrent: x is not 1", x)
			return nil
		}
		x++
		return nil
	}, func(ctx context.Context) error {
		if x != 2 {
			t.Error("concurrent: x is not 2", x)
			return nil
		}
		x++
		return nil
	})()
	if err != nil {
		t.Error("concurrent: err is not nil", err)
	}
	if x != 3 {
		t.Error("concurrent: x is not 3", x)
	}
}

func TestRunSequentiallyErr(t *testing.T) {
	ctx := context.Background()
	x := 0
	werr := errors.New("expected error")
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
		t.Error("concurrent: err is nil", err)
	}
	if x != 2 {
		t.Error("concurrent: x is not 2", x)
	}
	if err != werr {
		t.Error("concurrent: unexpected err", err)
	}
}

func TestRunSequentiallyPanic(t *testing.T) {
	ctx := context.Background()
	x := 0
	werr := errors.New("expected error")
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
	if err == nil {
		t.Error("concurrent: err is nil", err)
	}
	if err != nil && !strings.Contains(err.Error(), werr.Error()) {
		t.Error("concurrent: err is not equaled", err, werr)
	}
	if x != 2 {
		t.Error("x is not 2", x)
	}
}

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
		t.Errorf("concurrent: concurrentChLen is %d not %d", concurrentChLen, chLen)
	}
	go func() {
		times := 0
		for i := range ch {
			count -= i
			times++
		}
		if times != jobCount {
			t.Error("concurrent: count is not", jobCount, count)
		}
		if count != 0 {
			t.Error("concurrent: count is not zero", count)
		}
	}()
	err := wait()
	if err != nil {
		t.Error("concurrent: err not nil", err)
		return
	}
	close(ch)
}

func TestRunRunnerErr(t *testing.T) {
	ctx := context.Background()
	count := int32(0)
	add, wait := gotools.NewRunner[int](ctx, 4, func(ctx context.Context, i int) error {
		if atomic.AddInt32(&count, 1) == 5 {
			return errors.New("expected err")
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
		t.Error("concurrent: err not equal", err1, err2)
	}
}

func TestRunRunnerPanic(t *testing.T) {
	ctx := context.Background()
	count := int32(0)
	perr := errors.New("expected error")
	add, wait := gotools.NewRunner[int](ctx, 4, func(ctx context.Context, i int) error {
		if atomic.AddInt32(&count, 1) == 6 {
			panic(perr)
		}
		return nil
	})
	for i := 0; i < jobCount; i++ {
		_ = add(i, false)
	}
	err := wait()
	if err == nil {
		t.Error("concurrent: err is nil", err)
	}
	if err != nil && !strings.Contains(err.Error(), perr.Error()) {
		t.Error("concurrent: err is not equal", err)
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
		t.Error("concurrent: unexpected error", err)
	}
	for _, v := range jobs {
		if v.x != 2 {
			t.Errorf("concurrent: x != 2 %d", v.x)
		}
	}
}

func TestRunPipelineErr(t *testing.T) {
	type data struct {
		name string
		x    int
	}
	ctx := context.Background()
	jobs := []*data{{"job1", 0}, {"job2", 0}}
	perr := errors.New("expected error")
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
		return perr
	}
	err := gotools.RunPipeline(ctx, jobs, work1, work2)
	if err == nil {
		t.Error("concurrent: err is nil")
	}
	if err != perr {
		t.Error("concurrent: err not equal", err, perr)
	}
}

func TestRunPipelinePanic(t *testing.T) {
	type data struct {
		name string
		x    int
	}
	ctx := context.Background()
	jobs := []*data{{"job1", 0}, {"job2", 0}}
	perr := errors.New("expected error")
	work1 := func(ctx context.Context, d *data) error {
		return nil
	}
	work2 := func(ctx context.Context, d *data) error {
		panic(perr)
	}
	err := gotools.RunPipeline(ctx, jobs, work1, work2)
	if err == nil {
		t.Error("concurrent: err is nil")
	}
	if err != nil && !strings.Contains(err.Error(), perr.Error()) {
		t.Error("concurrent: unexpected error", err)
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
		t.Error("concurrent: unexpected error", err)
	}
	if count != 10 {
		t.Error("concurrent: unexpected count", count)
	}

	expectedErr := errors.New("expected error")
	err = gotools.Run(ctx, func(ctx context.Context, t int32) error {
		if t == 3 {
			return expectedErr
		}
		return nil
	}, 1, 2, 3, 4)
	if err != expectedErr {
		t.Error("concurrent: unexpected error", err)
	}

	err = gotools.Run(ctx, func(ctx context.Context, t int32) error {
		if t == 3 {
			panic(t)
		}
		return nil
	}, 1, 2, 3, 4)
	if err == nil || !strings.Contains(err.Error(), "3") {
		t.Error("concurrent: unexpected error", err)
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

	err := errors.New("expected error")
	go func() {
		time.Sleep(time.Second)
		err1 <- err
	}()
	if err != <-gotools.ListenChan(errChans...) {
		t.Error("concurrent: err not equal")
		return
	}

	go func() {
		time.Sleep(time.Second)
		close(err1)
		close(err2)
		close(err3)
	}()
	if nil != <-gotools.ListenChan(errChans...) {
		t.Error("concurrent: err is not nil")
	}
}

func TestRunPeriodically(t *testing.T) {
	run := gotools.RunPeriodically(time.Second)
	var now time.Time
	run(func() {
		time.Sleep(time.Second)
		now = time.Now()
	})
	run(func() {
		if time.Since(now) < time.Second {
			t.Error("concurrent: unexpected time", time.Now())
		}
		time.Sleep(time.Second)
		now = time.Now()
	})
	run(func() {
		if time.Since(now) < time.Second {
			t.Error("concurrent: unexpected time", time.Now())
		}
		now = time.Now()
	})
	run(func() {
		if time.Since(now) < time.Second {
			t.Error("concurrent: unexpected time", time.Now())
		}
	})
}
