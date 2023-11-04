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
	"runtime"
	"sync/atomic"
)

// FairLocker 公平读写锁。
//
// 先到达的请求先获取处理器，无论是读请求还是写请求。
type FairLocker struct {
	rMutex, wMutex, mutex int32
	readCount             uint64
}

func (locker *FairLocker) RLock() {
	enterMutex(&locker.mutex)
	enterMutex(&locker.rMutex)
	if locker.readCount == 0 {
		enterMutex(&locker.wMutex)
	}
	exitMutex(&locker.mutex)
	locker.readCount++
	exitMutex(&locker.rMutex)
}

func (locker *FairLocker) RUnlock() {
	enterMutex(&locker.rMutex)
	locker.readCount--
	if locker.readCount == 0 {
		exitMutex(&locker.wMutex)
	}
	exitMutex(&locker.rMutex)
}

func (locker *FairLocker) WLock() {
	enterMutex(&locker.mutex)
	enterMutex(&locker.wMutex)
	exitMutex(&locker.mutex)
}

func (locker *FairLocker) WUnlock() {
	exitMutex(&locker.wMutex)
}

// ReadFirstLocker 读优先锁。
//
// 读优先意味着，连续地读请求获取处理器的概率高于写请求。
type ReadFirstLocker struct {
	rMutex, wMutex int32
	readCount      uint64
}

func (locker *ReadFirstLocker) RLock() {
	enterMutex(&locker.rMutex)
	if locker.readCount == 0 {
		enterMutex(&locker.wMutex)
	}
	locker.readCount++
	exitMutex(&locker.rMutex)
}

func (locker *ReadFirstLocker) RUnlock() {
	enterMutex(&locker.rMutex)
	locker.readCount--
	if locker.readCount == 0 {
		exitMutex(&locker.wMutex)
	}
	exitMutex(&locker.rMutex)
}

func (locker *ReadFirstLocker) WLock() {
	enterMutex(&locker.wMutex)
}

func (locker *ReadFirstLocker) WUnlock() {
	exitMutex(&locker.wMutex)
}

// WriteFirstLocker 写优先锁。
//
// 读优先意味着，连续地写请求获取处理器的概率高于读请求。
type WriteFirstLocker struct {
	rMutex, wMutex, mutex int32
	readCount, writeCount uint64
}

func (locker *WriteFirstLocker) RLock() {
	enterMutex(&locker.rMutex)
	if locker.readCount == 0 {
		enterMutex(&locker.wMutex)
	}
	locker.readCount++
	exitMutex(&locker.rMutex)
}

func (locker *WriteFirstLocker) RUnlock() {
	enterMutex(&locker.rMutex)
	locker.readCount--
	if locker.readCount == 0 {
		exitMutex(&locker.wMutex)
	}
	exitMutex(&locker.rMutex)
}

func (locker *WriteFirstLocker) WLock() {
	enterMutex(&locker.mutex)
	if locker.writeCount == 0 {
		enterMutex(&locker.rMutex)
	}
	locker.writeCount++
	exitMutex(&locker.mutex)
	enterMutex(&locker.wMutex)
}

func (locker *WriteFirstLocker) WUnlock() {
	exitMutex(&locker.wMutex)
	enterMutex(&locker.mutex)
	locker.writeCount--
	if locker.writeCount == 0 {
		exitMutex(&locker.rMutex)
	}
	exitMutex(&locker.mutex)
}

func enterMutex(mutex *int32) {
	for atomic.CompareAndSwapInt32(mutex, 0, 1) {
		runtime.Gosched()
	}
}

func exitMutex(mutex *int32) {
	for atomic.CompareAndSwapInt32(mutex, 1, 0) {
		runtime.Gosched()
	}
}
