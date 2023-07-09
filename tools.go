package gotools

import (
	"fmt"
	"strings"
)

type Number interface {
	int | int8 | int16 | int32 | int32 |
	uint | uint8 | uint16 | uint32 | uint64 |
	uintptr | float32 | float64
}

func Max[T Number](x, y T) T {
	if y > x {
		return y
	}
	return x
}

func Min[T Number](x, y T) T {
	if y < x {
		return y
	}
	return x
}

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
