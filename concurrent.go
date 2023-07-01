package gotools

import (
	"context"
	"fmt"
	"sync"
)

// RunConcurrently n sender : 1 receiver
func RunConcurrently(fn ...func() error) <-chan error {
	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error, len(fn)) // 必须使用缓存
	wg := sync.WaitGroup{}
	for _, f := range fn {
		wg.Add(1)
		go func(f func() error) {
			defer func() {
				wg.Done()
				if v := recover(); v != nil {
					cancel()
					errChan <- fmt.Errorf("%v", v)
				}
			}()
			select {
			case <-ctx.Done():
			default:
				err := f()
				if err != nil {
					cancel()
					errChan <- err
				}
			}
		}(f)
	}
	go func() {
		wg.Wait()
		close(errChan)
		cancel()
	}()
	return errChan
}

func RunSequently(fn ...func() error) <-chan error {
	errChan := make(chan error)
	go func() {
		for _, f := range fn {
			err := func() (err error) {
				defer func() {
					if v := recover(); v != nil {
						err = fmt.Errorf("%v", v)
					}
				}()
				return f()
			}()
			if err != nil {
				errChan <- err
				break
			}
		}
		close(errChan)
	}()
	return errChan
}
