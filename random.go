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
	"math/rand"
	"os"
	"sync"
	"time"
)

var (
	random      *rand.Rand
	randomCache sync.Map
)

func init() {
	name, _ := os.Hostname()
	x := 0
	for i := range name {
		x += int(name[i])
	}
	seed := time.Now().UnixMilli() + int64(x)
	random = rand.New(rand.NewSource(seed))
}

// RandomChars 生成随机字符串（数字加字母组合）。
func RandomChars(length int) string {
	count := 3
	s := make([]byte, length)
FLAG:
	for i := 0; i < length; i++ {
		switch random.Intn(3) % 3 {
		case 0:
			s[i] = '0' + byte(random.Intn('9'-'0'+1))
		case 1:
			s[i] = 'a' + byte(random.Intn('z'-'a'+1))
		case 2:
			s[i] = 'A' + byte(random.Intn('Z'-'A'+1))
		}
	}
	res := string(s)
	actual, loaded := randomCache.LoadOrStore(res, time.Now())
	if loaded {
		t, _ := actual.(time.Time)
		if time.Since(t) < time.Hour*24*30 {
			if count > 0 {
				count--
				goto FLAG
			}
		}
	}
	randomCache.Store(res, time.Now())
	return res
}

// RandomCharsCaseInsensitive 生成随机字符串（数字加字母组合）。
func RandomCharsCaseInsensitive(length int) string {
	count := 3
	s := make([]byte, length)
FLAG:
	for i := 0; i < length; i++ {
		switch random.Intn(2) % 2 {
		case 0:
			s[i] = '0' + byte(random.Intn('9'-'0'+1))
		case 1:
			s[i] = 'a' + byte(random.Intn('z'-'a'+1))
		}
	}
	res := string(s)
	actual, loaded := randomCache.LoadOrStore(res, time.Now())
	if loaded {
		t, _ := actual.(time.Time)
		if time.Since(t) < time.Hour*24*30 {
			if count > 0 {
				count--
				goto FLAG
			}
		}
	}
	randomCache.Store(res, time.Now())
	return res
}

// UUIDLike 生成随机字符串（数字加字母组合）。
func UUIDLike() string {
	s := []byte(RandomCharsCaseInsensitive(32))
	res := make([]byte, 0, len(s)+4)
	res = append(append(res, s[:8]...), '-')
	res = append(append(res, s[8:12]...), '-')
	res = append(append(res, s[12:16]...), '-')
	res = append(append(res, s[16:20]...), '-')
	res = append(res, s[20:]...)
	return string(res)
}
