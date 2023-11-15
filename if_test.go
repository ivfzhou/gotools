package gotools_test

import (
	"testing"

	"gitee.com/ivfzhou/gotools/v4"
)

func TestIPv4ToNum(t *testing.T) {
	ipNum := gotools.IPv4ToNum("0.0.0.0")
	if ipNum != 0 {
		t.Error("if: ip num mistake", ipNum)
	}
	ipNum = gotools.IPv4ToNum("127.0.0.1")
	if ipNum != 0b1111111_00000000_00000000_00000001 {
		t.Error("if: ip num mistake", ipNum)
	}
	ipNum = gotools.IPv4ToNum("13.13.13.13")
	if ipNum != 0b1101_00001101_00001101_00001101 {
		t.Error("if: ip num mistake", ipNum)
	}
	ipNum = gotools.IPv4ToNum("13.13.13.256")
	if ipNum != 0 {
		t.Error("if: ip num mistake", ipNum)
	}
	ipNum = gotools.IPv4ToNum("13.13..250")
	if ipNum != 0 {
		t.Error("if: ip num mistake", ipNum)
	}
	ipNum = gotools.IPv4ToNum("13.-13.1.250")
	if ipNum != 0 {
		t.Error("if: ip num mistake", ipNum)
	}
	ipNum = gotools.IPv4ToNum(".13.1.250")
	if ipNum != 0 {
		t.Error("if: ip num mistake", ipNum)
	}
	ipNum = gotools.IPv4ToNum("255.255.255.255")
	if ipNum != 0b11111111_11111111_11111111_11111111 {
		t.Error("if: ip num mistake", ipNum)
	}
}

func TestIPv4ToStr(t *testing.T) {
	ipStr := gotools.IPv4ToStr(0)
	if ipStr != "0.0.0.0" {
		t.Error("if: ip str mistake", ipStr)
	}
	ipStr = gotools.IPv4ToStr(0b1111111_00000000_00000000_00000001)
	if ipStr != "127.0.0.1" {
		t.Error("if: ip str mistake", ipStr)
	}
	ipStr = gotools.IPv4ToStr(0b1101_00001101_00001101_00001101)
	if ipStr != "13.13.13.13" {
		t.Error("if: ip str mistake", ipStr)
	}
	ipStr = gotools.IPv4ToStr(0b11111111_11111111_11111111_11111111)
	if ipStr != "255.255.255.255" {
		t.Error("if: ip str mistake", ipStr)
	}
}

func TestIsIPv4(t *testing.T) {
	if !gotools.IsIPv4("0.0.0.0") {
		t.Error("if: unexpected ip matching")
	}
	if !gotools.IsIPv4("127.0.0.1") {
		t.Error("if: unexpected ip matching")
	}
	if !gotools.IsIPv4("255.255.255.255") {
		t.Error("if: unexpected ip matching")
	}
	if gotools.IsIPv4(".0.0.0") {
		t.Error("if: unexpected ip matching")
	}
	if gotools.IsIPv4("1.-0.0.0") {
		t.Error("if: unexpected ip matching")
	}
	if gotools.IsIPv4("1.0..0") {
		t.Error("if: unexpected ip matching")
	}
	if gotools.IsIPv4("1.0.1.270") {
		t.Error("if: unexpected ip matching")
	}
}

func TestIsIPv6(t *testing.T) {
	if !gotools.IsIPv6("::") {
		t.Error("if: unexpected ip matching")
	}
	if !gotools.IsIPv6("::1") {
		t.Error("if: unexpected ip matching")
	}
	if !gotools.IsIPv6("fda1::1") {
		t.Error("if: unexpected ip matching")
	}
	if !gotools.IsIPv6("fda1:123::1") {
		t.Error("if: unexpected ip matching")
	}
	if !gotools.IsIPv6("fda1:fda1:1:fda1:1:1:3") {
		t.Error("if: unexpected ip matching")
	}
	if gotools.IsIPv6(":fda1:fda1:1:fda1:1:1:3") {
		t.Error("if: unexpected ip matching")
	}
	if gotools.IsIPv6(":::") {
		t.Error("if: unexpected ip matching")
	}
	if gotools.IsIPv6(":1::") {
		t.Error("if: unexpected ip matching")
	}
	if gotools.IsIPv6("x:1") {
		t.Error("if: unexpected ip matching")
	}
	if gotools.IsIPv6("1:1:1:1:1:1:1:1:") {
		t.Error("if: unexpected ip matching")
	}
	if gotools.IsIPv6("1:1:1:1:1:1:1:1:1") {
		t.Error("if: unexpected ip matching")
	}
	if gotools.IsIPv6("1:1::0:1:1:1:1") {
		t.Error("if: unexpected ip matching")
	}
}

func TestIsMAC(t *testing.T) {
	if !gotools.IsMAC("00-00-00-00-00-00") {
		t.Error("if: unexpected mac matching")
	}
	if !gotools.IsMAC("ff-ff-ff-ff-ff-ff") {
		t.Error("if: unexpected mac matching")
	}
	if !gotools.IsMAC("ff-EE-ff-ff-ff-0f") {
		t.Error("if: unexpected mac matching")
	}
	if gotools.IsMAC("") {
		t.Error("if: unexpected mac matching")
	}
	if gotools.IsMAC("ff-0g-ff-ff-ff-0f") {
		t.Error("if: unexpected mac matching")
	}
	if gotools.IsMAC("ff-qE-ff-ff-ff-0f") {
		t.Error("if: unexpected mac matching")
	}
	if gotools.IsMAC("ff-qE-ff-ff-ff-0f-00") {
		t.Error("if: unexpected mac matching")
	}
}

func TestIsIntranet(t *testing.T) {
	if !gotools.IsIntranet("127.0.0.1") {
		t.Error("if: unexpected ip matching")
	}
	if !gotools.IsIntranet("127.127.0.1") {
		t.Error("if: unexpected ip matching")
	}
	if !gotools.IsIntranet("172.16.0.1") {
		t.Error("if: unexpected ip matching")
	}
	if !gotools.IsIntranet("172.16.0.1") {
		t.Error("if: unexpected ip matching")
	}
	if !gotools.IsIntranet("172.20.0.1") {
		t.Error("if: unexpected ip matching")
	}
	if !gotools.IsIntranet("192.168.0.1") {
		t.Error("if: unexpected ip matching")
	}
	if !gotools.IsIntranet("192.168.255.1") {
		t.Error("if: unexpected ip matching")
	}
	if !gotools.IsIntranet("10.48.0.1") {
		t.Error("if: unexpected ip matching")
	}
	if !gotools.IsIntranet("10.0.0.1") {
		t.Error("if: unexpected ip matching")
	}
	if gotools.IsIntranet("0.0.0.1") {
		t.Error("if: unexpected ip matching")
	}
	if gotools.IsIntranet("0.0.0.0") {
		t.Error("if: unexpected ip matching")
	}
	if gotools.IsIntranet("0..0.0") {
		t.Error("if: unexpected ip matching")
	}
}
