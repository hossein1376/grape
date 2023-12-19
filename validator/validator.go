package validator

import (
	"cmp"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Case is a test-case, consisting of two parts. Cond will call a function that returns a boolean.
// if it returns false, Msg will be added to the Validator.Errors.
type Case struct {
	Cond bool
	Msg  string
}

// ValidationError holds all validation error messages. It implements error interface.
type ValidationError map[string][]string

// Error returns validation errors in the following format:
//
//	"field_1: validation error, field_2: first error"
func (v ValidationError) Error() string {
	var s []string
	for key, value := range v {
		for _, msg := range value {
			s = append(s, key+": "+msg)
		}
	}
	return strings.Join(s, ", ")
}

// Validator will check for cases by Check method and will return a boolean with Valid method.
// If a validation error happens, the Msg will be stored inside the Errors map.
type Validator struct {
	Errors ValidationError `json:"errors"`
}

// New will return an instance of Validator.
func New() *Validator {
	return &Validator{Errors: make(map[string][]string)}
}

// Valid returns a boolean indicating whether validation was successful or not.
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// Check accepts name of the field as the first argument, following by an arbitrary number of validation Case.
func (v *Validator) Check(key string, cases ...Case) {
	for _, c := range cases {
		if !c.Cond {
			v.addError(key, c.Msg)
		}
	}
}

func (v *Validator) addError(key, message string) {
	v.Errors[key] = append(v.Errors[key], message)
}

// Empty checks if a string is empty.
func Empty(value string) bool {
	return len(value) == 0
}

// EndsWith check whether a string ends with a particular suffix.
func EndsWith(value, suffix string) bool {
	return strings.HasSuffix(value, suffix)
}

// IsNumber checks if the given string is all numbers.
func IsNumber(value string) bool {
	if _, err := strconv.Atoi(value); err != nil {
		return false
	}
	return true
}

// In checks if the value is present in a given list of arguments.
func In[T comparable](value T, list ...T) bool {
	return slices.Contains(list, value)
}

// Matches will match a string with a regular expression.
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// Max checks if a value is equal or lesser than a maximum.
// For length, use MaxLength instead.
func Max[T cmp.Ordered](value T, max T) bool {
	return cmp.Compare(value, max) == -1 || cmp.Compare(value, max) == 0
}

// MaxLength checks if a string's utf8 length is equal or lesser the given maximum.
func MaxLength(value string, max int) bool {
	return utf8.RuneCountInString(value) <= max
}

// Min checks if a value is equal or bigger than a minimum.
// For length, use MinLength instead.
func Min[T cmp.Ordered](value T, min T) bool {
	return cmp.Compare(value, min) == 1 || cmp.Compare(value, min) == 0
}

// MinLength checks if a string's utf8 length is equal or greater the given minimum.
func MinLength(value string, min int) bool {
	return min <= utf8.RuneCountInString(value)
}

// NotEmpty checks if the given value is not empty.
func NotEmpty(value string) bool {
	return len(value) != 0
}

// Range checks if a value is inside a number range, inclusive.
// For length, use RangeLength instead.
func Range[T cmp.Ordered](value T, min, max T) bool {
	return (cmp.Compare(value, min) == 1 || cmp.Compare(value, min) == 0) && (cmp.Compare(value, max) == -1 || cmp.Compare(value, max) == 0)
}

// RangeLength checks if a string's utf8 length is inside the given range, inclusive.
func RangeLength(value string, min, max int) bool {
	return min <= utf8.RuneCountInString(value) && utf8.RuneCountInString(value) <= max
}

// StartsWith check whether a string starts with a particular prefix.
func StartsWith(value, prefix string) bool {
	return strings.HasPrefix(value, prefix)
}

// Unique checks if all elements of a given slice are unique.
func Unique[T cmp.Ordered](values []T) bool {
	s := slices.Clone(values)
	slices.Sort(s)
	if len(slices.Compact(s)) != len(values) {
		return false
	}
	return true
}
