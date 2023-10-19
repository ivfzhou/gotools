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
	"fmt"
	"reflect"
	"strings"
)

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~uintptr | ~float32 | ~float64
}

// Max 求最大值。
func Max[T Number](x, y T) T {
	if y > x {
		return y
	}
	return x
}

// Min 求最小值。
func Min[T Number](x, y T) T {
	if y < x {
		return y
	}
	return x
}

// Join 拼接元素返回字符串。
func Join[T fmt.Stringer](arr []T, sep string) string {
	sb := strings.Builder{}
	for i := 0; i < len(arr)-1; i++ {
		sb.WriteString(arr[i].String())
		sb.WriteString(sep)
	}
	if len(arr) > 0 {
		sb.WriteString(arr[len(arr)-1].String())
		sb.WriteString(sep)
	}
	return sb.String()
}

func Convert[E, T any](sli []E, fn func(E) T) []T {
	list := make([]T, len(sli))
	for i := range sli {
		list[i] = fn(sli[i])
	}
	return list
}

func Distinct[E comparable](sli []E) []E {
	list := make([]E, 0, len(sli))
	m := make(map[E]struct{}, len(sli))
	for i := range sli {
		_, ok := m[sli[i]]
		if ok {
			continue
		}
		m[sli[i]] = struct{}{}
		list = append(list, sli[i])
	}
	return list
}

func Filter[E any](sli []E, fn func(E) bool) []E {
	list := make([]E, 0, len(sli))
	for i := range sli {
		if fn(sli[i]) {
			list = append(list, sli[i])
		}
	}
	return list
}

func DropZero[E any](sli []E) []E {
	list := make([]E, 0, len(sli))
	for i := range sli {
		val := reflect.ValueOf(sli[i])
		if !val.IsValid() {
			continue
		}
		if val.IsZero() {
			continue
		}
		switch val.Kind() {
		case reflect.Slice:
			fallthrough
		case reflect.Array:
			fallthrough
		case reflect.Map:
			if val.Len() <= 0 {
				continue
			}
		}
		list = append(list, sli[i])
	}
	return list
}

func Foreach[E any](sli []E, fn func(E)) {
	for i := range sli {
		fn(sli[i])
	}
}

func ForeachCanBreak[E any](sli []E, fn func(E) bool) {
	for i := range sli {
		if !fn(sli[i]) {
			break
		}
	}
}

func ForeachWithReturn[E, T any](sli []E, fn func(E) (T, bool)) T {
	for i := range sli {
		if t, ok := fn(sli[i]); ok {
			return t
		}
	}
	var t T
	return t
}

func FilterMap[K comparable, V any](m map[K]V, fn func(K, V) bool) map[K]V {
	nm := make(map[K]V, len(m))
	for k, v := range m {
		if fn(k, v) {
			nm[k] = v
		}
	}
	return nm
}

func PickMapValue[K comparable, V any](m map[K]V) []V {
	list := make([]V, 0, len(m))
	for _, v := range m {
		list = append(list, v)
	}
	return list
}

func PickMapKey[K comparable, V any](m map[K]V) []K {
	list := make([]K, 0, len(m))
	for k := range m {
		list = append(list, k)
	}
	return list
}

func ConvertToSlice[K comparable, V, T any](m map[K]V, fn func(K, V) T) []T {
	list := make([]T, 0, len(m))
	for k, v := range m {
		list = append(list, fn(k, v))
	}
	return list
}

func ConvertToMap[K comparable, V, E any](sli []E, fn func(E) (K, V)) map[K]V {
	m := make(map[K]V, len(sli))
	for i := range sli {
		k, v := fn(sli[i])
		m[k] = v
	}
	return m
}

func Contains[E comparable](arr []E, elem E) bool {
	for i := range arr {
		if arr[i] == elem {
			return true
		}
	}
	return false
}
