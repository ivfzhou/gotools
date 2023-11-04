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
	"strconv"
	"strings"
)

// IPv4ToNum ipv4字符串转数字。
//
// 若是非ip地址则返回0。
func IPv4ToNum(ip string) uint32 {
	res := uint32(0)
	arr := strings.Split(ip, ".")
	if len(arr) == 4 {
		num0, err := strconv.ParseUint(arr[0], 10, 32)
		if err != nil || num0 > 255 {
			return 0
		}
		num1, err := strconv.ParseUint(arr[1], 10, 32)
		if err != nil || num1 > 255 {
			return 0
		}
		num2, err := strconv.ParseUint(arr[2], 10, 32)
		if err != nil || num2 > 255 {
			return 0
		}
		num3, err := strconv.ParseUint(arr[3], 10, 32)
		if err != nil || num3 > 255 {
			return 0
		}
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
	arr := strings.Split(s, ".")
	if len(arr) == 4 {
		num, err := strconv.ParseUint(arr[0], 10, 32)
		if err != nil || num > 255 {
			return false
		}
		num, err = strconv.ParseUint(arr[1], 10, 32)
		if err != nil || num > 255 {
			return false
		}
		num, err = strconv.ParseUint(arr[2], 10, 32)
		if err != nil || num > 255 {
			return false
		}
		num, err = strconv.ParseUint(arr[3], 10, 32)
		if err != nil || num > 255 {
			return false
		}
		return true
	}
	return false
}

// IsIPv6 判断是否是ipv6。
func IsIPv6(s string) bool {
	if len(s) <= 0 {
		return false
	}
	compressIndex := strings.Index(s, "::")
	if compressIndex != strings.LastIndex(s, "::") {
		return false
	}
	if compressIndex > -1 && len(s) > compressIndex+2 && s[compressIndex+2] == '0' {
		return false
	}
	if compressIndex == 0 {
		s = s[2:]
	} else if compressIndex > -1 && compressIndex == len(s)-2 {
		s = s[:len(s)-2]
	} else if compressIndex > -1 {
		s = string(append([]byte(s[:compressIndex]), []byte(s[compressIndex+1:])...))
	}
	if len(s) <= 0 {
		return true
	}
	arr := strings.Split(s, ":")
	if len(arr) > 7 || compressIndex > -1 && len(arr) > 7 {
		return false
	}
	for _, v := range arr {
		if vl := len(v); vl > 4 || vl <= 0 {
			return false
		}
		num, err := strconv.ParseUint(v, 16, 16)
		if err != nil || num > 0xffff {
			return false
		}
	}
	return true
}

// IsMAC 判断是否是mac地址。
func IsMAC(s string) bool {
	arr := strings.Split(s, "-")
	if len(arr) != 6 {
		return false
	}
	for _, v := range arr {
		if len(v) != 2 {
			return false
		}
		num, err := strconv.ParseUint(v, 16, 8)
		if err != nil || num > 0xff {
			return false
		}
	}
	return true
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
