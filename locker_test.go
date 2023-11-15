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
	"sync"
	"testing"

	"gitee.com/ivfzhou/gotools/v4"
)

func TestFairLocker(t *testing.T) {
	fairLocker := &gotools.FairLocker{}
	count := 5
	writer := func() {
		fairLocker.WLock()
		defer fairLocker.WUnlock()
		// t.Log("writing")
		count--
	}
	reader := func() {
		fairLocker.RLock()
		defer fairLocker.RUnlock()
		// t.Log("reading count", count)
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
	// t.Log("count:", count)
}

func TestReadFirstLocker(t *testing.T) {
	fairLocker := &gotools.ReadFirstLocker{}
	count := 5
	writer := func() {
		fairLocker.WLock()
		defer fairLocker.WUnlock()
		// t.Log("writing")
		count--
	}
	reader := func() {
		fairLocker.RLock()
		defer fairLocker.RUnlock()
		// t.Log("reading count", count)
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
	// t.Log("count:", count)
}

func TestWriteFirstLocker(t *testing.T) {
	fairLocker := &gotools.WriteFirstLocker{}
	count := 500
	writer := func() {
		fairLocker.WLock()
		defer fairLocker.WUnlock()
		// t.Log("writing")
		count--
	}
	reader := func() {
		fairLocker.RLock()
		defer fairLocker.RUnlock()
		// t.Log("reading count", count)
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
	// t.Log("count:", count)
}
