package validator

import (
	"cmp"
	"regexp"
	"testing"
)

func TestEmpty(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{
			name:  "Empty input",
			value: "",
			want:  true,
		},
		{
			name:  "Not empty input",
			value: "mew?",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Empty(tt.value); got != tt.want {
				t.Errorf("EndsWith() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEndsWith(t *testing.T) {
	tests := []struct {
		name   string
		value  string
		suffix string
		want   bool
	}{
		{
			name:   "Ends with the given suffix",
			value:  "if only cats had thumbs",
			suffix: "thumbs",
			want:   true,
		},
		{
			name:   "Doesn't end with the given suffix",
			value:  "if only cats had thumbs",
			suffix: "toes?",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EndsWith(tt.value, tt.suffix); got != tt.want {
				t.Errorf("EndsWith() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsNumber(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{
			name:  "Valid number",
			value: "1234567890",
			want:  true,
		},
		{
			name:  "Invalid number",
			value: "12#de5$9!",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNumber(tt.value); got != tt.want {
				t.Errorf("IsNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIn(t *testing.T) {
	type testCase[T comparable] struct {
		name  string
		value T
		list  []T
		want  bool
	}

	numberTests := []testCase[int]{
		{
			name:  "Number exists in the given list",
			value: 5,
			list:  []int{3, 4, 5, 6},
			want:  true,
		},
		{
			name:  "Number doesn't exist in the given list",
			value: -3,
			list:  []int{0, 2, 0, 4},
			want:  false,
		},
	}
	stringTests := []testCase[string]{
		{
			name:  "String exists in the given list",
			value: "cat",
			list:  []string{"fish", "cat", "dog"},
			want:  true,
		},
		{
			name:  "String with space inside exists in the given list",
			value: "two cats",
			list:  []string{"one cat", "two", "cats", "two cats"},
			want:  true,
		},
		{
			name:  "String doesn't exist in the given list",
			value: "",
			list:  []string{"sleep", "chill", "play"},
			want:  false,
		},
	}

	for _, tt := range numberTests {
		t.Run(tt.name, func(t *testing.T) {
			if got := In(tt.value, tt.list...); got != tt.want {
				t.Errorf("In() = %v, want %v", got, tt.want)
			}
		})
	}

	for _, tt := range stringTests {
		t.Run(tt.name, func(t *testing.T) {
			if got := In(tt.value, tt.list...); got != tt.want {
				t.Errorf("In() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatches(t *testing.T) {
	tests := []struct {
		name  string
		value string
		rx    *regexp.Regexp
		want  bool
	}{
		{
			name:  "Numbers regex",
			value: "3546",
			rx:    regexp.MustCompile(`[0-9]+`),
			want:  true,
		},
		{
			name:  "Uppercase letters regex",
			value: "CUTIE PIE ",
			rx:    regexp.MustCompile(`^\P{L}*\p{Lu}\P{Ll}*$`),
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Matches(tt.value, tt.rx); got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMax(t *testing.T) {
	type testCase[T cmp.Ordered] struct {
		name  string
		value T
		max   T
		want  bool
	}
	numberTests := []testCase[int]{
		{
			name:  "Smaller than max",
			value: -1,
			max:   2,
			want:  true,
		},
		{
			name:  "Equal than max",
			value: 6,
			max:   6,
			want:  true,
		},
		{
			name:  "Bigger than max",
			value: 4,
			max:   3,
			want:  false,
		},
	}

	for _, tt := range numberTests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Max(tt.value, tt.max); got != tt.want {
				t.Errorf("Max() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMaxLength(t *testing.T) {
	tests := []struct {
		name  string
		value string
		max   int
		want  bool
	}{
		{
			name:  "Lesser characters than max",
			value: "abc",
			max:   5,
			want:  true,
		},
		{
			name:  "Same number of characters as max",
			value: "num ",
			max:   4,
			want:  true,
		},
		{
			name:  "More characters than max",
			value: "this a very long sentence",
			max:   6,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaxLength(tt.value, tt.max); got != tt.want {
				t.Errorf("MaxLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMin(t *testing.T) {
	type testCase[T cmp.Ordered] struct {
		name  string
		value T
		min   T
		want  bool
	}
	numberTests := []testCase[int]{
		{
			name:  "Bigger than min",
			value: 18,
			min:   12,
			want:  true,
		},
		{
			name:  "Equal the min",
			value: 9,
			min:   9,
			want:  true,
		},
		{
			name:  "smaller than min",
			value: -2,
			min:   2,
			want:  false,
		},
	}

	for _, tt := range numberTests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Min(tt.value, tt.min); got != tt.want {
				t.Errorf("Min() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMinLength(t *testing.T) {
	tests := []struct {
		name  string
		value string
		min   int
		want  bool
	}{
		{
			name:  "More characters than min",
			value: "rather long sentence",
			min:   4,
			want:  true,
		},
		{
			name:  "Equal number of characters to min",
			value: "  ",
			min:   2,
			want:  true,
		},
		{
			name:  "Less characters than min",
			value: "",
			min:   1,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MinLength(tt.value, tt.min); got != tt.want {
				t.Errorf("MinLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNotEmpty(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{
			name:  "Empty input",
			value: "",
			want:  false,
		},
		{
			name:  "Not empty input",
			value: "text",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NotEmpty(tt.value); got != tt.want {
				t.Errorf("NotEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRange(t *testing.T) {
	type testCase[T cmp.Ordered] struct {
		name  string
		value T
		min   T
		max   T
		want  bool
	}
	tests := []testCase[int]{
		{
			name:  "Input between min and max",
			value: 7,
			min:   3,
			max:   8,
			want:  true,
		},
		{
			name:  "Input outside range",
			value: 2,
			min:   12,
			max:   20,
			want:  false,
		},
		{
			name:  "Input equals min",
			value: 11,
			min:   11,
			max:   19,
			want:  true,
		},
		{
			name:  "Input equals max",
			value: -4,
			min:   -9,
			max:   -4,
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Range(tt.value, tt.min, tt.max); got != tt.want {
				t.Errorf("Range() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRangeLength(t *testing.T) {
	tests := []struct {
		name  string
		value string
		min   int
		max   int
		want  bool
	}{
		{
			name:  "Character count between min and max",
			value: "423",
			min:   1,
			max:   4,
			want:  true,
		},
		{
			name:  "Character count outside range",
			value: "oops!",
			min:   5,
			max:   9,
			want:  true,
		},
		{
			name:  "Character equals min",
			value: "text",
			min:   4,
			max:   6,
			want:  true,
		},
		{
			name:  "Character equals min",
			value: "a text",
			min:   4,
			max:   6,
			want:  true,
		},
		{
			name:  "Non-English text in range",
			value: "تهران",
			min:   4,
			max:   6,
			want:  true,
		},
		{
			name:  "Non-English text outside range",
			value: "شهریار",
			min:   1,
			max:   4,
			want:  false,
		},
		{
			name:  "Non-English text equal min",
			value: "بابا",
			min:   4,
			max:   10,
			want:  true,
		},
		{
			name:  "Non-English text equal max",
			value: "مامان",
			min:   2,
			max:   5,
			want:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RangeLength(tt.value, tt.min, tt.max); got != tt.want {
				t.Errorf("RangeLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStartsWith(t *testing.T) {
	tests := []struct {
		name   string
		value  string
		prefix string
		want   bool
	}{
		{
			name:   "Starts with prefix",
			value:  "life is short",
			prefix: "life",
			want:   true,
		},
		{
			name:   "Doesn't start with prefix",
			value:  "life is short",
			prefix: "Life",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StartsWith(tt.value, tt.prefix); got != tt.want {
				t.Errorf("StartsWith() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnique(t *testing.T) {
	type testCase[T cmp.Ordered] struct {
		name   string
		values []T
		want   bool
	}

	numberTests := []testCase[int]{
		{
			name:   "Unique numbers",
			values: []int{6, 99, 0, -7},
			want:   true,
		},
		{
			name:   "Non-unique numbers",
			values: []int{2, 77, 11, 7, 11},
			want:   false,
		},
	}
	stringTests := []testCase[string]{
		{
			name:   "Unique strings",
			values: []string{"one", "two", "three"},
			want:   true,
		},
		{
			name:   "Non-unique strings",
			values: []string{"one", "two", "three", "two"},
			want:   false,
		},
	}

	for _, tt := range numberTests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Unique(tt.values); got != tt.want {
				t.Errorf("Unique() = %v, want %v", got, tt.want)
			}
		})
	}
	for _, tt := range stringTests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Unique(tt.values); got != tt.want {
				t.Errorf("Unique() = %v, want %v", got, tt.want)
			}
		})
	}
}
