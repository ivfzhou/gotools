package gotools_test

import (
	"math/rand"
	"regexp"
	"testing"
	"unicode/utf8"

	"gitee.com/ivfzhou/gotools/v3"
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
		randomChars := gotools.RandomChars(l)
		cl := len(randomChars)
		if cl != l {
			t.Errorf("random: the randomChars length mismatch: %d %d", cl, l)
		}
		if !utf8.ValidString(randomChars) {
			t.Error("random: randomChars invalid utf8")
		}
		for i := range randomChars {
			_, ok := m[randomChars[i]]
			if ok {
				delete(m, randomChars[i])
			}
			switch c := randomChars[i]; {
			case c >= '0' && c <= '9':
			case c >= 'a' && c <= 'z':
			case c >= 'A' && c <= 'Z':
			default:
				t.Error("random: randomChars invalid", randomChars)
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
		randomChars := gotools.RandomCharsCaseInsensitive(l)
		cl := len(randomChars)
		if cl != l {
			t.Errorf("random: randomChars length mismatch: %d %d", cl, l)
		}
		if !utf8.ValidString(randomChars) {
			t.Error("random: randomChars invalid utf8")
		}
		for i := range randomChars {
			_, ok := m[randomChars[i]]
			if ok {
				delete(m, randomChars[i])
			}
			switch c := randomChars[i]; {
			case c >= '0' && c <= '9':
			case c >= 'a' && c <= 'z':
			case c >= 'A' && c <= 'Z':
			default:
				t.Error("random: randomChars invalid", randomChars)
			}
		}
		if len(m) <= 0 {
			break
		}
	}
}

func TestUUIDLike(t *testing.T) {
	uuidMatcher := regexp.MustCompile(`^[a-z0-9]{8}-([a-z0-9]{4}-){3}[a-z0-9]{12}$`)
	for i := 0; i < 10; i++ {
		uuid := gotools.UUIDLike()
		if !uuidMatcher.MatchString(uuid) {
			t.Errorf("random: uuid mismatch %s", uuid)
		}
	}
}
