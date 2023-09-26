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
	"reflect"
	"sync"
)

// RunConcurrently 并发运行 fn。
func RunConcurrently(fn ...func() error) (wait func() error) {
	ctx, cancel := context.WithCancelCause(context.Background())
	wg := sync.WaitGroup{}
	wg.Add(len(fn))
	for _, f := range fn {
		go func(f func() error) {
			var err error
			defer func() {
				wg.Done()
				if v := recover(); v != nil {
					var ok bool
					if err, ok = v.(error); !ok {
						err = fmt.Errorf("%v", v)
					}
					cancel(err)
				}
			}()
			select {
			case <-ctx.Done():
			default:
				if err = f(); err != nil {
					cancel(err)
				}
			}
		}(f)
	}
	return func() error {
		wg.Wait()
		return context.Cause(ctx)
	}
}

// RunSequently  依次运行 fn，当有err时停止后续 fn 运行。
func RunSequently(fn ...func() error) (wait func() error) {
	errCh := make(chan error)
	go func() {
		for _, f := range fn {
			err := func() (err error) {
				defer func() {
					if v := recover(); v != nil {
						var ok bool
						if err, ok = v.(error); !ok {
							err = fmt.Errorf("%v", v)
						}
					}
				}()
				return f()
			}()
			if err != nil {
				errCh <- err
				close(errCh)
				return
			}
		}
		errCh <- nil
		close(errCh)
	}()
	return func() error {
		return <-errCh
	}
}

// RunParallel 该函数提供同时运行 max 个协程 fn，一旦 fn 有err返回则停止接下来的fn运行。
// 朝返回的 add 函数中添加任务，若正在运行的任务数已达到max则会阻塞当前程序。
// add 函数返回err为任务 fn 返回的第一个err。与 wait 函数返回的err为同一个。
// 注意请在 add 完所有任务后调用 wait。
func RunParallel[T any](max int, fn func(T) error) (add func(T) error, wait func() error) {
	limiter := make(chan struct{}, max)
	ctx, cancel := context.WithCancelCause(context.Background())
	wg := sync.WaitGroup{}
	fnWrap := func(data T) {
		var err error
		defer func() {
			if p := recover(); p != nil {
				var ok bool
				if err, ok = p.(error); !ok {
					err = fmt.Errorf("%v", p)
				}
			}
			if err != nil {
				cancel(err)
			}
			<-limiter
			wg.Done()
		}()
		err = fn(data)
	}
	add = func(data T) error {
		wg.Add(1)
		select {
		case <-ctx.Done():
			wg.Done()
			return context.Cause(ctx)
		case limiter <- struct{}{}:
		}
		select {
		case <-ctx.Done():
			wg.Done()
			return context.Cause(ctx)
		default:
		}
		go fnWrap(data)
		return context.Cause(ctx)
	}
	wait = func() error {
		wg.Wait()
		close(limiter)
		return context.Cause(ctx)
	}
	return
}

// RunParallelNoBlock 该函数提供同 RunParallel 一样，但是 add 函数不会阻塞。注意请在 add 完所有任务后调用 wait。
func RunParallelNoBlock[T any](max int, fn func(T) error) (add func(T) error, wait func() error) {
	limiter := make(chan struct{}, max)
	ctx, cancel := context.WithCancelCause(context.Background())
	wg := sync.WaitGroup{}
	fnWrap := func(data T) {
		var err error
		defer func() {
			if p := recover(); p != nil {
				var ok bool
				if err, ok = p.(error); !ok {
					err = fmt.Errorf("%v", p)
				}
			}
			if err != nil {
				cancel(err)
			}
			<-limiter
		}()
		err = fn(data)
	}
	add = func(data T) error {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			case limiter <- struct{}{}:
			}
			select {
			case <-ctx.Done():
				return
			default:
				fnWrap(data)
			}
		}()
		return context.Cause(ctx)
	}
	wait = func() error {
		wg.Wait()
		close(limiter)
		return context.Cause(ctx)
	}
	return
}

// Run 并发将jobs传递给proc函数运行，一旦发生error便立即返回该error，并结束其它协程。
// 当ctx被cancel时也将立即返回，此时返回cancel时的error。
// 当proc运行发生panic将立即返回该panic字符串化的error。
// proc为nil时函数将panic。
func Run[T any](ctx context.Context, proc func(context.Context, T) error, jobs ...T) error {
	if proc == nil {
		panic("the proc cannot nil")
	}
	if len(jobs) <= 0 {
		return nil
	}
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	wg := &sync.WaitGroup{}
	for i := range jobs {
		wg.Add(1)
		go func(t *T) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			default:
			}
			defer func() {
				if v := recover(); v != nil {
					cancel(fmt.Errorf("%v", v))
				}
			}()
			if err := proc(ctx, *t); err != nil {
				cancel(err)
			}
		}(&jobs[i])
	}
	wg.Wait()
	return context.Cause(ctx)
}

// StartProcess 将每个jobs依次递给steps函数处理。一旦某个step发生error或者panic，StartProcess立即返回该error，
// 并及时结束其他StartProcess开启的goroutine，也不开启新的goroutine运行step。
// 一个job最多在一个step中运行一次，且一个job一定是依次序递给steps，前一个step处理完毕才会给下一个step处理。
// 每个step并发运行jobs。
// StartProcess等待所有goroutine运行结束才返回，或者ctx被cancel时也将及时结束开启的goroutine后返回。
// StartProcess因被ctx cancel而结束时函数返回nil。若steps中含有nil StartProcess将会panic。
func StartProcess[T any](ctx context.Context, jobs []T, steps ...func(context.Context, T) error) error {
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
		nextCh, errChans[i] = startOneProcess(ctx, step, cancel, nextCh)
	}

	for _, errCh := range errChans {
		if err := <-errCh; err != nil {
			return err
		}
	}
	return nil
}

// Listen 监听chans，一旦有一个chan激活便立即将T发送给ch，并close ch。
// 若所有chan都未曾激活（chan是nil也认为未激活）且都close了，或者ctx被cancel了，则ch被close。
// 若同时chan被激活和ctx被cancel，则随机返回一个激活发送给chan的值。
func Listen[T any](ctx context.Context, chans ...<-chan T) (ch <-chan T) {
	tch := make(chan T, 1)
	ch = tch
	if len(chans) <= 0 {
		close(tch)
		return
	}
	select {
	case <-ctx.Done():
		close(tch)
		return
	default:
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
		scs = append(scs, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ctx.Done()),
		})
		for {
			if len(scs) <= 1 {
				close(tch)
				return
			}
			chosen, recv, ok := reflect.Select(scs)
			if chosen == len(scs)-1 {
				close(tch)
				return
			}
			if !ok {
				scs = append(scs[:chosen], scs[chosen+1:]...)
			} else {
				reflect.ValueOf(tch).Send(recv)
				close(tch)
				return
			}
		}
	}()
	return ch
}

func startOneProcess[T any](ctx context.Context, f func(context.Context, T) error, notify func(),
	dataChan <-chan *T) (<-chan *T, <-chan error) {
	ctx, cancel := context.WithCancel(ctx)
	ch := make(chan *T, cap(dataChan))
	errCh := make(chan error, 1)
	go func() {
		once := &sync.Once{}
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
			case data, ok := <-dataChan:
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
					defer func() {
						if v := recover(); v != nil {
							err := fmt.Errorf("%v", v)
							cancel()
							notify()
							once.Do(func() { errCh <- err })
						}
					}()
					if err := f(ctx, *t); err != nil {
						cancel()
						notify()
						once.Do(func() { errCh <- err })
					} else {
						ch <- t
					}
				}(data)
			}
		}
	}()
	return ch, errCh
}
