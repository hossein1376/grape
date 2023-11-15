package validator

import (
	"cmp"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Case struct {
	Cond bool
	Msg  string
}

type Validator struct {
	Errors map[string]string `json:"errors"`
}

func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validator) Check(key string, cases ...Case) {
	for _, c := range cases {
		if !c.Cond {
			v.addError(key, c.Msg)
		}
	}
}

func (v *Validator) addError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

func Contains(value string, list ...string) bool {
	for _, item := range list {
		if ok := strings.Contains(value, item); ok {
			return false
		}
	}
	return true
}

func Empty(value string) bool {
	return len(value) == 0
}

func EndsWith(value, suffix string) bool {
	return strings.HasSuffix(value, suffix)
}

func IsNumber(value string) bool {
	if _, err := strconv.Atoi(value); err != nil {
		return false
	}
	return true
}

func In[T comparable](value T, list ...T) bool {
	return slices.Contains(list, value)
}

func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

func Max[T cmp.Ordered](value T, max T) bool {
	return cmp.Compare(value, max) == -1 || cmp.Compare(value, max) == 0
}

func MaxLength(value string, max int) bool {
	return utf8.RuneCountInString(value) <= max
}

func Min[T cmp.Ordered](value T, min T) bool {
	return cmp.Compare(value, min) == 1 || cmp.Compare(value, min) == 0
}

func MinLength(value string, min int) bool {
	return min <= utf8.RuneCountInString(value)
}

func NotEmpty(value string) bool {
	return len(value) != 0
}

func Range[T cmp.Ordered](value T, min, max T) bool {
	return (cmp.Compare(value, min) == 1 || cmp.Compare(value, min) == 0) && (cmp.Compare(value, max) == -1 || cmp.Compare(value, max) == 0)
}

func RangeLength(value string, min, max int) bool {
	return min <= utf8.RuneCountInString(value) && utf8.RuneCountInString(value) <= max
}

func StartsWith(value, prefix string) bool {
	return strings.HasPrefix(value, prefix)
}

func Unique[T cmp.Ordered](values []T) bool {
	s := slices.Clone(values)
	slices.Sort(s)
	if len(slices.Compact(s)) != len(values) {
		return false
	}
	return true
}
