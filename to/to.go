package to

import (
	"strconv"
)

// num to string

type number interface {
	int | int8 | int64 | uint | float64
}

// S i: int8, int, uint, int64, float64(2位小数)
func S[T number](val T) string {
	switch v := any(val).(type) {
	case int8:
		return strconv.Itoa(int(v))
	case int:
		return strconv.Itoa(v)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return Sf(v, 2)
	default:
		return ""
	}
}

func Sf(v float64, prec int) string {
	return strconv.FormatFloat(v, 'f', prec, 64)
}

// string to num

func I(s string) (int, bool) {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return i, true
}

func U(s string) (uint, bool) {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	if i < 0 {
		return 0, false
	}
	return uint(i), true
}

func I64(s string) (int64, bool) {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, false
	}
	return i, true
}

func F(s string) (float64, bool) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

// string to bool

func B(s string) bool {
	switch s {
	case "1", "t", "T", "true", "TRUE", "True":
		return true
	case "0", "f", "F", "false", "FALSE", "False":
		return false
	}
	return false
}
