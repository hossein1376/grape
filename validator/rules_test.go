package validator

import (
	"errors"
	"reflect"
	"regexp"
	"testing"
)

type role string

type item struct {
	ID   int
	Tags []string // makes item non-comparable
}

func TestRuleChains(t *testing.T) {
	digits := regexp.MustCompile(`^\d+$`)
	tests := []struct {
		name  string
		rules func(v *Validator)
		want  string // the field's error, or "" if it must pass
	}{
		// strings
		{"required fails", func(v *Validator) { v.String("f", "").Required() }, "must not be empty"},
		{"required passes", func(v *Validator) { v.String("f", "a").Required() }, ""},
		{"min len counts runes", func(v *Validator) { v.String("f", "çé").MinLen(2) }, ""},
		{"min len fails", func(v *Validator) { v.String("f", "a").MinLen(2) }, "must be at least 2 characters"},
		{"max len fails", func(v *Validator) { v.String("f", "abc").MaxLen(2) }, "must be at most 2 characters"},
		{"len range fails", func(v *Validator) { v.String("f", "a").LenRange(2, 4) }, "must be between 2 and 4 characters"},
		{"digits fails", func(v *Validator) { v.String("f", "12a").Digits() }, "must contain only digits"},
		{"matches fails", func(v *Validator) { v.String("f", "x").Matches(digits) }, "has an invalid format"},
		{"matches passes", func(v *Validator) { v.String("f", "42").Matches(digits) }, ""},
		{"in with named type", func(v *Validator) { v.String("f", role("admin")).In("admin", "user") }, ""},
		{"in fails", func(v *Validator) { v.String("f", role("root")).In("admin", "user") }, `must be one of "admin", "user"`},
		{"in single value", func(v *Validator) { v.String("f", "b").In("a") }, `must be "a"`},
		{"starts with fails", func(v *Validator) { v.String("f", "x").StartsWith("http://", "https://") }, `must start with one of "http://", "https://"`},
		{"ends with fails", func(v *Validator) { v.String("f", "x").EndsWith(".com") }, `must end with ".com"`},
		{"contains fails", func(v *Validator) { v.String("f", "x").Contains("@") }, `must contain "@"`},
		// numbers
		{"not zero fails", func(v *Validator) { v.Number("f", 0.0).NotZero() }, "must not be zero"},
		{"min fails", func(v *Validator) { v.Number("f", -1).Min(0) }, "must be at least 0"},
		{"max fails", func(v *Validator) { v.Number("f", uint8(200)).Max(100) }, "must be at most 100"},
		{"range passes inclusive", func(v *Validator) { v.Number("f", 99).Range(0, 99) }, ""},
		{"range fails", func(v *Validator) { v.Number("f", 1.5).Range(2, 3) }, "must be between 2 and 3"},
		{"number in fails", func(v *Validator) { v.Number("f", 3).In(1, 2) }, "must be one of 1, 2"},
		// slices
		{"slice required fails", func(v *Validator) { v.Slice("f", []int(nil)).Required() }, "must not be empty"},
		{"slice len fails", func(v *Validator) { v.Slice("f", []int{1}).Len(2) }, "must have exactly 2 items"},
		{"slice min len fails", func(v *Validator) { v.Slice("f", []int{1}).MinLen(2) }, "must have at least 2 items"},
		{"slice max len fails", func(v *Validator) { v.Slice("f", []int{1, 2}).MaxLen(1) }, "must have at most 1 item"},
		{"slice len range fails", func(v *Validator) { v.Slice("f", []int{1}).LenRange(2, 3) }, "must have between 2 and 3 items"},
		{"unique fails", func(v *Validator) { v.Slice("f", []string{"a", "b", "a"}).Unique() }, "must not contain duplicates"},
		{"unique passes", func(v *Validator) { v.Slice("f", []string{"a", "b"}).Unique() }, ""},
		{"unique by non-comparable items", func(v *Validator) {
			v.Slice("f", []item{{ID: 1}, {ID: 1}}).UniqueBy(func(i item) int { return i.ID })
		}, "must not contain duplicates"},
		// chaining
		{"stops at first failure", func(v *Validator) { v.String("f", "").Required().MinLen(2).MaxLen(0) }, "must not be empty"},
		{"later rule fails", func(v *Validator) { v.String("f", "a").Required().MinLen(2) }, "must be at least 2 characters"},
		{"msg overrides failed rule", func(v *Validator) { v.String("f", "").Required().Msg("name please") }, "name please"},
		{"msg ignored for passed rule", func(v *Validator) { v.String("f", "a").Required().Msg("x").MinLen(2) }, "must be at least 2 characters"},
		{"msg ignored for skipped rule", func(v *Validator) { v.String("f", "").Required().MinLen(2).Msg("x") }, "must not be empty"},
		{"number msg", func(v *Validator) { v.Number("f", -1).Min(0).Msg("no negatives") }, "no negatives"},
		{"slice msg", func(v *Validator) { v.Slice("f", []int{}).Required().Msg("add one") }, "add one"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v Validator
			tt.rules(&v)
			if tt.want == "" {
				if len(v.Errors) != 0 {
					t.Fatalf("expected no errors, got %v", v.Errors)
				}
				return
			}
			if want := (ValidationError{"f": {tt.want}}); !reflect.DeepEqual(v.Errors, want) {
				t.Fatalf("got errors %v, want %v", v.Errors, want)
			}
		})
	}
}

func TestUnique_PanicsForNonComparableItems(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected a panic")
		}
	}()
	New().Slice("f", []item{{}}).Unique()
}

func TestValidator_Err(t *testing.T) {
	v := New()
	if err := v.Err(); err != nil {
		t.Fatalf("expected nil error for a valid input, got %#v", err)
	}

	v.Number("age", -1).Min(0)
	v.Check("name", Case{Cond: false, Msg: "a"}, Case{Cond: false, Msg: "b"})
	var ve ValidationError
	if err := v.Err(); !errors.As(err, &ve) {
		t.Fatalf("expected a ValidationError, got %#v", err)
	}
	want := ValidationError{"age": {"must be at least 0"}, "name": {"a", "b"}}
	if !reflect.DeepEqual(ve, want) {
		t.Fatalf("got %v, want %v", ve, want)
	}
}
