package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// number is the set of numeric types accepted by [Validator.Number] and
// [Zero].
type number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 |
		~uint32 | ~uint64 | ~float32 | ~float64
}

// chain holds the state shared by all rule chains. A chain stops at its first
// failing rule, so each chain reports at most one error.
type chain struct {
	v          *Validator
	key        string
	failed     bool // a rule has failed; the remaining rules are skipped
	lastFailed bool // the most recent rule ran and failed
}

func (c *chain) rule(ok func() bool, msg func() string) {
	c.lastFailed = false
	if c.failed || ok() {
		return
	}
	c.failed, c.lastFailed = true, true
	c.v.addError(c.key, msg())
}

// setMsg replaces the error message of the most recent rule, if it failed.
func (c *chain) setMsg(msg string) {
	if !c.lastFailed {
		return
	}
	errs := c.v.Errors[c.key]
	errs[len(errs)-1] = msg
}

// oneOf formats values for error messages, e.g. `"a"` or `one of "a", "b"`.
func oneOf[T any](verb string, values []T) string {
	s := make([]string, len(values))
	for i, v := range values {
		s[i] = fmt.Sprintf(verb, v)
	}
	if len(s) == 1 {
		return s[0]
	}
	return "one of " + strings.Join(s, ", ")
}

// plural returns "n word" or "n words", e.g. "1 item" or "2 items".
func plural(n int, word string) string {
	if n == 1 || n == -1 {
		return fmt.Sprintf("%d %s", n, word)
	}
	return fmt.Sprintf("%d %ss", n, word)
}

// StringRules is a chain of rules for a string field, created by
// [Validator.String]. Lengths are counted in runes.
type StringRules[S ~string] struct {
	chain
	value S
}

// String starts a chain of rules for a string field.
//
//	v.String("username", req.Username).Required().MaxLen(10)
func (v *Validator) String[S ~string](key string, value S) *StringRules[S] {
	return &StringRules[S]{chain: chain{v: v, key: key}, value: value}
}

// Msg replaces the default message of the preceding rule, if it failed.
func (r *StringRules[S]) Msg(msg string) *StringRules[S] {
	r.setMsg(msg)
	return r
}

// Required checks the string is not empty.
func (r *StringRules[S]) Required() *StringRules[S] {
	r.rule(
		func() bool { return !Empty(r.value) },
		func() string { return "must not be empty" },
	)
	return r
}

// MinLen checks the string has at least n runes.
func (r *StringRules[S]) MinLen(n int) *StringRules[S] {
	r.rule(
		func() bool { return LengthMin(r.value, n) },
		func() string { return "must be at least " + plural(n, "character") },
	)
	return r
}

// MaxLen checks the string has at most n runes.
func (r *StringRules[S]) MaxLen(n int) *StringRules[S] {
	r.rule(
		func() bool { return LengthMax(r.value, n) },
		func() string { return "must be at most " + plural(n, "character") },
	)
	return r
}

// LenRange checks the string has between minimum and maximum runes,
// inclusive.
func (r *StringRules[S]) LenRange(minimum, maximum int) *StringRules[S] {
	r.rule(
		func() bool { return LengthRange(r.value, minimum, maximum) },
		func() string {
			return fmt.Sprintf(
				"must be between %d and %d characters", minimum, maximum,
			)
		},
	)
	return r
}

// Digits checks the string consists only of ASCII digits. See [IsNumber].
func (r *StringRules[S]) Digits() *StringRules[S] {
	r.rule(
		func() bool { return IsNumber(r.value) },
		func() string { return "must contain only digits" },
	)
	return r
}

// Matches checks the string matches the regular expression.
func (r *StringRules[S]) Matches(re *regexp.Regexp) *StringRules[S] {
	r.rule(
		func() bool { return re.MatchString(string(r.value)) },
		func() string { return "has an invalid format" },
	)
	return r
}

// In checks the string is one of the given values.
func (r *StringRules[S]) In(values ...S) *StringRules[S] {
	r.rule(
		func() bool { return In(r.value, values...) },
		func() string { return "must be " + oneOf("%q", values) },
	)
	return r
}

// StartsWith checks the string starts with one of the given prefixes.
func (r *StringRules[S]) StartsWith(prefixes ...S) *StringRules[S] {
	r.rule(
		func() bool { return StartsWith(r.value, prefixes...) },
		func() string { return "must start with " + oneOf("%q", prefixes) },
	)
	return r
}

// EndsWith checks the string ends with one of the given suffixes.
func (r *StringRules[S]) EndsWith(suffixes ...S) *StringRules[S] {
	r.rule(
		func() bool { return EndsWith(r.value, suffixes...) },
		func() string { return "must end with " + oneOf("%q", suffixes) },
	)
	return r
}

// Contains checks the string contains one of the given values.
func (r *StringRules[S]) Contains(values ...S) *StringRules[S] {
	r.rule(
		func() bool { return Contains(r.value, values...) },
		func() string { return "must contain " + oneOf("%q", values) },
	)
	return r
}

// NumberRules is a chain of rules for a numeric field, created by
// [Validator.Number].
type NumberRules[T number] struct {
	chain
	value T
}

// Number starts a chain of rules for a numeric field.
//
//	v.Number("age", req.Age).Range(0, 130)
func (v *Validator) Number[T number](key string, value T) *NumberRules[T] {
	return &NumberRules[T]{chain: chain{v: v, key: key}, value: value}
}

// Msg replaces the default message of the preceding rule, if it failed.
func (r *NumberRules[T]) Msg(msg string) *NumberRules[T] {
	r.setMsg(msg)
	return r
}

// NotZero checks the number is not 0.
func (r *NumberRules[T]) NotZero() *NumberRules[T] {
	r.rule(
		func() bool { return !Zero(r.value) },
		func() string { return "must not be zero" },
	)
	return r
}

// Min checks the number is at least minimum.
func (r *NumberRules[T]) Min(minimum T) *NumberRules[T] {
	r.rule(
		func() bool { return Min(r.value, minimum) },
		func() string { return fmt.Sprintf("must be at least %v", minimum) },
	)
	return r
}

// Max checks the number is at most maximum.
func (r *NumberRules[T]) Max(maximum T) *NumberRules[T] {
	r.rule(
		func() bool { return Max(r.value, maximum) },
		func() string { return fmt.Sprintf("must be at most %v", maximum) },
	)
	return r
}

// Range checks the number is between minimum and maximum, inclusive.
func (r *NumberRules[T]) Range(minimum, maximum T) *NumberRules[T] {
	r.rule(
		func() bool { return Range(r.value, minimum, maximum) },
		func() string {
			return fmt.Sprintf("must be between %v and %v", minimum, maximum)
		},
	)
	return r
}

// In checks the number is one of the given values.
func (r *NumberRules[T]) In(values ...T) *NumberRules[T] {
	r.rule(
		func() bool { return In(r.value, values...) },
		func() string { return "must be " + oneOf("%v", values) },
	)
	return r
}

// SliceRules is a chain of rules for a slice field, created by
// [Validator.Slice].
type SliceRules[E any] struct {
	chain
	value []E
}

// Slice starts a chain of rules for a slice field.
//
//	v.Slice("tags", req.Tags).Required().MaxLen(5).Unique()
func (v *Validator) Slice[S ~[]E, E any](key string, value S) *SliceRules[E] {
	return &SliceRules[E]{chain: chain{v: v, key: key}, value: value}
}

// Msg replaces the default message of the preceding rule, if it failed.
func (r *SliceRules[E]) Msg(msg string) *SliceRules[E] {
	r.setMsg(msg)
	return r
}

// Required checks the slice is not empty.
func (r *SliceRules[E]) Required() *SliceRules[E] {
	r.rule(
		func() bool { return len(r.value) != 0 },
		func() string { return "must not be empty" },
	)
	return r
}

// Len checks the slice has exactly n items.
func (r *SliceRules[E]) Len(n int) *SliceRules[E] {
	r.rule(
		func() bool { return len(r.value) == n },
		func() string { return "must have exactly " + plural(n, "item") },
	)
	return r
}

// MinLen checks the slice has at least n items.
func (r *SliceRules[E]) MinLen(n int) *SliceRules[E] {
	r.rule(
		func() bool { return len(r.value) >= n },
		func() string { return "must have at least " + plural(n, "item") },
	)
	return r
}

// MaxLen checks the slice has at most n items.
func (r *SliceRules[E]) MaxLen(n int) *SliceRules[E] {
	r.rule(
		func() bool { return len(r.value) <= n },
		func() string { return "must have at most " + plural(n, "item") },
	)
	return r
}

// LenRange checks the slice has between minimum and maximum items, inclusive.
func (r *SliceRules[E]) LenRange(minimum, maximum int) *SliceRules[E] {
	r.rule(
		func() bool { return LenRange(r.value, minimum, maximum) },
		func() string {
			return fmt.Sprintf(
				"must have between %d and %d items", minimum, maximum,
			)
		},
	)
	return r
}

// Unique checks the slice has no duplicate items. It panics if E isn't
// comparable; use [SliceRules.UniqueBy] for such types.
func (r *SliceRules[E]) Unique() *SliceRules[E] {
	if !reflect.TypeFor[E]().Comparable() {
		panic(fmt.Sprintf(
			"validator: Unique needs comparable items, %s isn't; use UniqueBy",
			reflect.TypeFor[E](),
		))
	}
	return r.UniqueBy(func(e E) any { return e })
}

// UniqueBy checks no two items of the slice have the same key.
//
//	v.Slice("users", req.Users).UniqueBy(func(u User) string { return u.Email })
func (r *SliceRules[E]) UniqueBy[K comparable](key func(E) K) *SliceRules[E] {
	r.rule(
		func() bool {
			seen := make(map[K]struct{}, len(r.value))
			for _, e := range r.value {
				k := key(e)
				if _, ok := seen[k]; ok {
					return false
				}
				seen[k] = struct{}{}
			}
			return true
		},
		func() string { return "must not contain duplicates" },
	)
	return r
}
