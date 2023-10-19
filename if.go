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
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var (
	ipv4Matcher = func() func() *regexp.Regexp {
		once := sync.Once{}
		var re *regexp.Regexp
		return func() *regexp.Regexp {
			once.Do(func() { re = regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`) })
			return re
		}
	}()
	ipv6Matcher = func() func() *regexp.Regexp {
		once := sync.Once{}
		var re *regexp.Regexp
		return func() *regexp.Regexp {
			once.Do(func() {
				re = regexp.MustCompile(`(^::[0-9a-fA-F]{1,4})|([0-9a-fA-F]{1,4}(:{1,2}[0-9a-fA-F]{1,4}){1,7})$`)
			})
			return re
		}
	}()
	macMatcher = func() func() *regexp.Regexp {
		once := sync.Once{}
		var re *regexp.Regexp
		return func() *regexp.Regexp {
			once.Do(func() {
				re = regexp.MustCompile(`^[0-9a-fA-F]{2}(-[0-9a-fA-F]{2}){5}$`)
			})
			return re
		}
	}()
)

// IPv4ToNum ipv4字符串转数字。
func IPv4ToNum(ip string) uint32 {
	res := uint32(0)
	arr := strings.Split(ip, ".")
	if len(arr) == 4 {
		num0, _ := strconv.ParseUint(arr[0], 10, 32)
		num1, _ := strconv.ParseUint(arr[1], 10, 32)
		num2, _ := strconv.ParseUint(arr[2], 10, 32)
		num3, _ := strconv.ParseUint(arr[3], 10, 32)
		res = uint32(num3)
		res |= uint32(num2) << 8
		res |= uint32(num1) << 16
		res |= uint32(num0) << 24
	}

	return res
}

// IPv4ToStr ipv4数字转字符串。
func IPv4ToStr(ip uint32) string {
	res := uint64(ip)
	s1 := strconv.FormatUint(res>>24&0xff, 10)
	s2 := strconv.FormatUint(res>>16&0xff, 10)
	s3 := strconv.FormatUint(res>>8&0xff, 10)
	s4 := strconv.FormatUint(res>>0&0xff, 10)
	return fmt.Sprintf("%s.%s.%s.%s", s1, s2, s3, s4)
}

// IsIPv4 判断是否是ipv4。
func IsIPv4(s string) bool {
	return ipv4Matcher().MatchString(s)
}

// IsIPv6 判断是否是ipv6。
func IsIPv6(s string) bool {
	return ipv6Matcher().MatchString(s)
}

// IsMAC 判断是否是mac地址。
func IsMAC(s string) bool {
	return macMatcher().MatchString(s)
}

// IsIntranet 判断是否是内网IP。
func IsIntranet(ipv4 string) bool {
	ipNum := IPv4ToNum(ipv4)
	if ipNum>>16 == (192<<8 | 168) {
		return true
	}
	if ipNum>>20 == (172<<4 | 16>>4) {
		return true
	}
	if ipNum>>24 == 10 {
		return true
	}
	if ipNum>>24 == 127 {
		return true
	}
	return false
}
