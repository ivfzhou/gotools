package gotools_test

import (
	"fmt"
	"math/rand"
	"testing"
	"unicode/utf8"

	"gitee.com/ivfzhou/gotools"
)

func TestRandomChars(t *testing.T) {
	m := make(map[byte]struct{}, 42)
	for i := byte('0'); i <= '9'; i++ {
		m[i] = struct{}{}
	}
	for i := byte('a'); i <= 'z'; i++ {
		m[i] = struct{}{}
	}
	for i := byte('A'); i <= 'Z'; i++ {
		m[i] = struct{}{}
	}

	for {
		l := rand.Intn(33)
		chars := gotools.RandomChars(l)
		if len(chars) != l {
			t.Error("chars len not match")
		}
		if !utf8.ValidString(chars) {
			t.Error("char invalid")
		}
		for i := range chars {
			_, ok := m[chars[i]]
			if ok {
				delete(m, chars[i])
			}
			switch c := chars[i]; {
			case c >= '0' && c <= '9':
			case c >= 'a' && c <= 'z':
			case c >= 'A' && c <= 'Z':
			default:
				t.Error("char invalid", chars)
			}
		}
		if len(m) <= 0 {
			break
		}
	}
}

func TestRandomCharsCaseInsensitive(t *testing.T) {
	m := make(map[byte]struct{}, 42)
	for i := byte('0'); i <= '9'; i++ {
		m[i] = struct{}{}
	}
	for i := byte('a'); i <= 'z'; i++ {
		m[i] = struct{}{}
	}

	for {
		l := rand.Intn(33)
		chars := gotools.RandomCharsCaseInsensitive(l)
		if len(chars) != l {
			t.Error("chars len not match")
		}
		if !utf8.ValidString(chars) {
			t.Error("char invalid")
		}
		for i := range chars {
			_, ok := m[chars[i]]
			if ok {
				delete(m, chars[i])
			}
			switch c := chars[i]; {
			case c >= '0' && c <= '9':
			case c >= 'a' && c <= 'z':
			case c >= 'A' && c <= 'Z':
			default:
				t.Error("char invalid", chars)
			}
		}
		if len(m) <= 0 {
			break
		}
	}
}

func TestRandomCharsCaseInsensitive2(t *testing.T) {
	for i := 0; i < 10; i++ {
		fmt.Println(gotools.UUIDLike())
	}
}
