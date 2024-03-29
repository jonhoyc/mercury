package x

import (
	"reflect"
	"strconv"
	"strings"
)

func Join(slice interface{}, sep string) string {
	switch d := slice.(type) {
	case []int64:
		return joinInt64(d, sep)
	case []string:
		return strings.Join(d, sep)
	case []int32:
		return joinInt32(d, sep)
	}
	return ""
}

func joinInt64(slice []int64, sep string) string {
	if len(slice) == 0 {
		return ""
	}
	if len(slice) == 1 {
		return strconv.FormatInt(slice[0], 10)
	}
	a := make([]string, len(slice))
	for i, item := range slice {
		a[i] = strconv.FormatInt(item, 10)
	}
	n := len(sep) * (len(a) - 1)
	for i := 0; i < len(a); i++ {
		n += len(a[i])
	}

	b := make([]byte, n)
	bp := copy(b, a[0])
	for _, s := range a[1:] {
		bp += copy(b[bp:], sep)
		bp += copy(b[bp:], s)
	}
	return string(b)
}

func joinInt32(slice []int32, sep string) string {
	if len(slice) == 0 {
		return ""
	}
	if len(slice) == 1 {
		return strconv.FormatInt(int64(slice[0]), 10)
	}
	a := make([]string, len(slice))
	for i, item := range slice {
		a[i] = strconv.FormatInt(int64(item), 10)
	}
	n := len(sep) * (len(a) - 1)
	for i := 0; i < len(a); i++ {
		n += len(a[i])
	}

	b := make([]byte, n)
	bp := copy(b, a[0])
	for _, s := range a[1:] {
		bp += copy(b[bp:], sep)
		bp += copy(b[bp:], s)
	}
	return string(b)
}

func SplitInt64(s, sep string) []int64 {
	slice := strings.Split(s, sep)

	n := len(slice)
	a := make([]int64, n)
	i := 0
	for i < n {
		a[i], _ = strconv.ParseInt(slice[i], 10, 64)
		i++
	}
	return a
}

func IsInSlice(slice interface{}, item interface{}) bool {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		panic("IsInSlice() given a non-slice type")
	}
	for i := 0; i < s.Len(); i++ {
		if s.Index(i).Interface() == item {
			return true
		}
	}
	return false
}
