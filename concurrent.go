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
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"
)

// RunConcurrently 并发运行fn，但有error发生时终止运行。
func RunConcurrently(ctx context.Context, fn ...func(context.Context) error) (wait func() error) {
	select {
	case <-ctx.Done():
		return func() error {
			err := ctx.Err()
			if err == nil {
				err = context.Canceled
			}
			return err
		}
	default:
	}
	ctx, cancel := context.WithCancel(ctx)
	var (
		returnErr error
		cancelErr error
	)
	wg := sync.WaitGroup{}
	errOnce := &sync.Once{}
	wg.Add(len(fn))
	for _, f := range fn {
		go func(f func(context.Context) error) {
			var err error
			defer func() {
				wg.Done()
				if p := recover(); p != nil {
					err = fmt.Errorf("panic: %v [recovered]\n%s\n", p, StackTrace())
				}
				if err != nil {
					cancel()
					if !errors.Is(err, context.Canceled) {
						errOnce.Do(func() { returnErr = err })
					}
				}
			}()
			select {
			case <-ctx.Done():
				cancelErr = ctx.Err()
				if cancelErr == nil {
					cancelErr = context.Canceled
				}
			default:
				err = f(ctx)
			}
		}(f)
	}
	return func() error {
		wg.Wait()
		cancel()
		if returnErr != nil {
			return returnErr
		}
		return cancelErr
	}
}

// RunSequentially 依次运行fn，当有error发生时停止后续fn运行。
func RunSequentially(ctx context.Context, fn ...func(context.Context) error) (wait func() error) {
	select {
	case <-ctx.Done():
		return func() error {
			err := ctx.Err()
			if err == nil {
				err = context.Canceled
			}
			return err
		}
	default:
	}
	var returnErr error
	notify := make(chan struct{})
	go func() {
		for _, f := range fn {
			select {
			case <-ctx.Done():
				returnErr = ctx.Err()
				if returnErr == nil {
					returnErr = context.Canceled
				}
				close(notify)
			default:
			}
			err := func() (err error) {
				defer func() {
					if p := recover(); p != nil {
						err = fmt.Errorf("panic: %v [recovered]\n%s\n", p, StackTrace())
					}
				}()
				err = f(ctx)
				return
			}()
			if err != nil {
				returnErr = err
				close(notify)
				return
			}
		}
		close(notify)
	}()
	return func() error {
		<-notify
		return returnErr
	}
}

// NewRunner 该函数提供同时运行max个协程fn，一旦fn发生error便终止fn运行。
//
// max小于等于0表示不限制协程数。
//
// 朝返回的run函数中添加fn，若block为true表示正在运行的任务数已达到max则会阻塞。
//
// add函数返回error为任务fn返回的第一个error，与wait函数返回的error为同一个。
//
// 注意请在add完所有任务后调用wait。
func NewRunner[T any](ctx context.Context, max int, fn func(context.Context, T) error) (
	run func(t T, block bool) error, wait func() error) {

	var returnErr error
	select {
	case <-ctx.Done():
		returnErr = ctx.Err()
		if returnErr == nil {
			returnErr = context.Canceled
		}
		return func(t T, block bool) error { return returnErr }, func() error { return returnErr }
	default:
	}

	var limiter chan struct{}
	if max > 0 {
		limiter = make(chan struct{}, max)
	}
	ctx, cancel := context.WithCancel(ctx)
	errOnce := &sync.Once{}
	wg := sync.WaitGroup{}
	var cancelErr error
	fnWrapper := func(t T) {
		var err error
		defer func() {
			if p := recover(); p != nil {
				err = fmt.Errorf("panic: %v [recovered]\n%s\n", p, StackTrace())
			}
			if err != nil {
				if !errors.Is(err, context.Canceled) {
					errOnce.Do(func() { returnErr = err })
				}
				cancel()
			}
			if limiter != nil {
				<-limiter
			}
			wg.Done()
		}()
		select {
		case <-ctx.Done():
			ctxErr := ctx.Err()
			if ctxErr != nil && cancelErr == nil {
				cancelErr = ctxErr
			}
		default:
			err = fn(ctx, t)
		}
	}
	run = func(t T, block bool) error {
		wg.Add(1)
		if block && limiter != nil {
			select {
			case <-ctx.Done():
				defer wg.Done()
				if returnErr != nil {
					return returnErr
				}
				if cancelErr != nil {
					return cancelErr
				}
				ctxErr := ctx.Err()
				if ctxErr != nil {
					cancelErr = ctxErr
				}
				return cancelErr
			case limiter <- struct{}{}:
			}
			go fnWrapper(t)
			return returnErr
		}
		go func() {
			if limiter != nil {
				select {
				case <-ctx.Done():
					ctxErr := ctx.Err()
					if ctxErr != nil && cancelErr == nil {
						cancelErr = ctxErr
					}
					wg.Done()
					return
				case limiter <- struct{}{}:
				}
			}
			fnWrapper(t)
		}()

		if returnErr != nil {
			return returnErr
		}
		return cancelErr
	}
	wait = func() error {
		wg.Wait()
		if returnErr != nil {
			return returnErr
		}
		return cancelErr
	}
	return
}

// Run 并发将jobs传递给fn函数运行，一旦发生error便立即返回该error，并结束其它协程。
func Run[T any](ctx context.Context, fn func(context.Context, T) error, jobs ...T) error {
	if len(jobs) <= 0 {
		return nil
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	wg := &sync.WaitGroup{}
	var (
		returnErr error
		cancelErr error
	)
	errOnce := &sync.Once{}
	for i := range jobs {
		wg.Add(1)
		go func(t *T) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				cancelErr = ctx.Err()
				if cancelErr == nil {
					cancelErr = context.Canceled
				}
				return
			default:
			}
			var err error
			defer func() {
				if p := recover(); p != nil {
					err = fmt.Errorf("panic: %v [recovered]\n%s\n", p, StackTrace())
				}
				if err != nil {
					cancel()
					if !errors.Is(err, context.Canceled) {
						errOnce.Do(func() { returnErr = err })
					}
				}
			}()
			err = fn(ctx, *t)
		}(&jobs[i])
	}
	wg.Wait()

	if returnErr != nil {
		return returnErr
	}
	return cancelErr
}

// RunPipeline 将每个jobs依次递给steps函数处理。一旦某个step发生error或者panic，StartProcess立即返回该error，
// 并及时结束其他StartProcess开启的goroutine，也不开启新的goroutine运行step。
//
// 一个job最多在一个step中运行一次，且一个job一定是依次序递给steps，前一个step处理完毕才会给下一个step处理。
//
// 每个step并发运行jobs。
//
// 等待所有goroutine运行结束才返回，或者ctx被cancel时也将及时结束开启的goroutine后返回。
//
// 因被ctx cancel而结束时函数返回nil。
//
// 若steps中含有nil将会panic。
func RunPipeline[T any](ctx context.Context, jobs []T, steps ...func(context.Context, T) error) error {
	if len(steps) <= 0 || len(jobs) <= 0 {
		return nil
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	ch := make(chan *T, len(jobs))
	go func() {
		for i := range jobs {
			ch <- &jobs[i]
		}
		close(ch)
	}()
	errChans := make([]<-chan error, len(steps))
	var nextCh <-chan *T = ch
	for i, step := range steps {
		nextCh, errChans[i] = startStep(ctx, step, cancel, nextCh)
	}

	return <-ListenChan(errChans...)
}

// ListenChan 监听chans，一旦有一个chan激活便立即将T发送给ch，并close ch。
//
// 若所有chans都未曾激活（chan是nil也认为未激活）且都close了，则ch被close。
//
// 若同时多个chans被激活，则随机将一个激活值发送给ch。
func ListenChan[T any](chans ...<-chan T) (ch <-chan T) {
	tch := make(chan T, 1)
	ch = tch
	if len(chans) <= 0 {
		close(tch)
		return
	}
	go func() {
		scs := make([]reflect.SelectCase, 0, len(chans))
		for i := range chans {
			if chans[i] == nil {
				continue
			}
			scs = append(scs, reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(chans[i]),
			})
		}
		for {
			if len(scs) <= 0 {
				close(tch)
				return
			}
			chosen, recv, ok := reflect.Select(scs)
			if !ok {
				scs = append(scs[:chosen], scs[chosen+1:]...)
			} else {
				reflect.ValueOf(tch).Send(recv)
				close(tch)
				return
			}
		}
	}()

	return
}

// RunPeriodically 依次运行fn，每个fn之间至少间隔period时间。
func RunPeriodically(period time.Duration) (run func(fn func())) {
	fnChan := make(chan func())
	lastAccess := time.Time{}
	errCh := make(chan any, 1)
	fnWrapper := func(f func()) (p any) {
		defer func() { p = recover() }()
		f()
		return nil
	}
	go func() {
		for f := range fnChan {
			time.Sleep(period - time.Since(lastAccess))
			errCh <- fnWrapper(f)
			lastAccess = time.Now()
		}
	}()
	return func(fn func()) {
		fnChan <- fn
		if p := <-errCh; p != nil {
			panic(p)
		}
	}
}

func StackTrace() string {
	sb := &strings.Builder{}
	var pc [4096]uintptr
	l := runtime.Callers(2, pc[:])
	frames := runtime.CallersFrames(pc[:l])
	for {
		frame, more := frames.Next()
		_, _ = fmt.Fprintf(sb, "%s\n", frame.Function)
		_, _ = fmt.Fprintf(sb, "    %s:%v\n", frame.File, frame.Line)
		if !more {
			break
		}
	}

	return sb.String()
}

func startStep[T any](ctx context.Context, f func(context.Context, T) error, notify func(), dataChan <-chan *T) (
	<-chan *T, <-chan error) {

	ch := make(chan *T, cap(dataChan))
	errCh := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithCancel(ctx)
		errOnce := &sync.Once{}
		wg := &sync.WaitGroup{}
		defer func() {
			wg.Wait()
			cancel()
			close(ch)
			close(errCh)
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case t, ok := <-dataChan:
				if !ok {
					return
				}
				wg.Add(1)
				go func(t *T) {
					defer wg.Done()
					select {
					case <-ctx.Done():
						return
					default:
					}
					var err error
					defer func() {
						if p := recover(); p != nil {
							err = fmt.Errorf("panic: %v [recovered]\n%s\n", p, StackTrace())
						}
						if err != nil {
							cancel()
							notify()
							if !errors.Is(err, context.Canceled) {
								errOnce.Do(func() { errCh <- err })
							}
						} else {
							ch <- t
						}
					}()
					err = f(ctx, *t)
				}(t)
			}
		}
	}()

	return ch, errCh
}
