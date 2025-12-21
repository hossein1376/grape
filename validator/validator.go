package validator

import (
	"cmp"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Case is a test-case, consisting of two parts. Cond will call a function that
// returns a boolean. If it returns false, Msg will be added to the
// [Validator.Errors].
type Case struct {
	Cond bool
	Msg  string
}

// ValidationError holds all validation error messages. It also implements the
// error interface.
type ValidationError map[string][]string

// Error returns validation errors in the following format:
//
//	"field_1: validation error, field_2: first error"
func (v ValidationError) Error() string {
	var s []string
	// Collect keys and sort them to make the error string deterministic.
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	for _, key := range keys {
		for _, msg := range v[key] {
			s = append(s, key+": "+msg)
		}
	}
	return strings.Join(s, ", ")
}

// Validator will check for cases by [Check] method and will return a boolean
// with the [Validate] method. If a validation error happens, the Msg will be
// stored inside the [Errors] map.
type Validator struct {
	Errors ValidationError `json:"errors"`
}

// New will return an instance of Validator.
func New() *Validator {
	return &Validator{Errors: make(map[string][]string)}
}

// Check accepts name of the field as the first argument, following by an
// arbitrary number of validation Case.
func (v *Validator) Check(key string, cases ...Case) {
	for _, c := range cases {
		if !c.Cond {
			v.addError(key, c.Msg)
		}
	}
}

// Validate returns a boolean indicating whether validation was successful or not.
func (v Validator) Validate() bool {
	return len(v.Errors) == 0
}

func (v *Validator) addError(key, message string) {
	v.Errors[key] = append(v.Errors[key], message)
}

// Not returns the negation of the given predicate.
func Not(predicate bool) bool {
	return !predicate
}

// Empty checks if the given input is empty (of len zero).
func Empty[T ~string](input T) bool {
	return len(input) == 0
}

// Len checks if the given input is of given length. For strings, use [Length*]
// functions.
func Len[S any, T ~[]S](input T, l int) bool {
	return len(input) == l
}

// LenRange checks if the given input is in the given range, inclusive. For
// strings, use [Length*] functions.
func LenRange[S any, T ~[]S](input T, low, high int) bool {
	return low <= len(input) && len(input) <= high
}

// EndsWith check whether the input ends with one of the given suffixes.
func EndsWith[T ~string](input T, suffixes ...T) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(string(input), string(suffix)) {
			return true
		}
	}
	return false
}

// StartsWith check whether the input starts with one of the given prefixes.
func StartsWith[T ~string](input T, prefixes ...T) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(string(input), string(prefix)) {
			return true
		}
	}
	return false
}

// Contains check if the input contains any of the given values.
func Contains[T ~string](input T, values ...T) bool {
	for _, value := range values {
		if strings.Contains(string(input), string(value)) {
			return true
		}
	}
	return false
}

// IsNumber checks if the given string is all numbers.
func IsNumber[T ~string](input T) bool {
	_, err := strconv.Atoi(string(input))
	return err == nil
}

// In checks if the input is present in the given list of arguments.
func In[T comparable](input T, list ...T) bool {
	if len(list) == 0 {
		return false
	}
	return slices.Contains(list, input)
}

// Matches will match a string with a regular expression.
func Matches(input string, re *regexp.Regexp) bool {
	return re.MatchString(input)
}

// Max checks if the input is equal or lesser than a maximum.
// For length, use LengthMax instead.
func Max[T cmp.Ordered](input T, max T) bool {
	return cmp.Compare(input, max) == -1 || cmp.Compare(input, max) == 0
}

// Min checks if the input is equal or bigger than a minimum.
// For length, use LengthMin instead.
func Min[T cmp.Ordered](input T, min T) bool {
	return cmp.Compare(input, min) == 1 || cmp.Compare(input, min) == 0
}

// Range checks if the input is inside a number range, inclusive.
// For length, use LengthRange instead.
func Range[T cmp.Ordered](input T, min, max T) bool {
	return (cmp.Compare(input, min) == 1 || cmp.Compare(input, min) == 0) &&
		(cmp.Compare(input, max) == -1 || cmp.Compare(input, max) == 0)
}

// LengthMax checks if a string's utf8 length is equal or lesser the given
// maximum.
func LengthMax[T ~string](input T, max int) bool {
	return utf8.RuneCountInString(string(input)) <= max
}

// LengthMin checks if a string's utf8 length is equal or greater the given
// minimum.
func LengthMin[T ~string](input T, min int) bool {
	return min <= utf8.RuneCountInString(string(input))
}

// LengthRange checks if a string's utf8 length is inside the given range,
// inclusive.
func LengthRange[T ~string](input T, min, max int) bool {
	return min <= utf8.RuneCountInString(string(input)) &&
		utf8.RuneCountInString(string(input)) <= max
}

// Unique checks if all elements of a given slice are unique.
func Unique[T cmp.Ordered](values []T) bool {
	if len(values) == 0 || len(values) == 1 {
		return true
	}
	s := slices.Clone(values)
	slices.Sort(s)
	return len(slices.Compact(s)) == len(values)
}
