package gotools_test

import (
	"testing"

	"gitee.com/ivfzhou/gotools/v2"
)

func TestGCD(t *testing.T) {
	if gotools.GCD(9, 3) != 3 {
		t.Error("gcd mismatch", 9, 3)
	}
	if gotools.GCD(12, 9) != 3 {
		t.Error("gcd mismatch", 12, 9)
	}
	if gotools.GCD(15, 10) != 5 {
		t.Error("gcd mismatch", 15, 10)
	}
	if gotools.GCD(-15, -10) != 5 {
		t.Error("gcd mismatch", -15, -10)
	}
	if gotools.GCD(-15, 10) != 5 {
		t.Error("gcd mismatch", -15, 10)
	}
	if gotools.GCD(15, -10) != 5 {
		t.Error("gcd mismatch", 15, -10)
	}
}
