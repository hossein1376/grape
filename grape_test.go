package grape

import (
	"errors"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestParamMissing(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	// No path parameter set -> should return ErrMissingParam
	_, err := Param(req, "id", ParseInt[int]())
	if !errors.Is(err, ErrMissingParam) {
		t.Fatalf("expected ErrMissingParam, got %v", err)
	}
}

func TestParamNilParser(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	_, err := Param[string](req, "id", nil)
	if !errors.Is(err, ErrNilParser) {
		t.Fatalf("expected ErrNilParser, got %v", err)
	}
}

func TestParamOrDefault(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	v := ParamOrDefault(req, "id", ParseInt[int](), 11)
	if v != 11 {
		t.Fatalf("expected 11, got %d", v)
	}
}

func TestQueryParsingSuccessAndMissing(t *testing.T) {
	values := url.Values{}
	values.Set("n", "42")

	v, err := Query(values, "n", ParseInt[int]())
	if err != nil {
		t.Fatalf("unexpected error parsing query: %v", err)
	}
	if v != 42 {
		t.Fatalf("unexpected parsed value: got %d want %d", v, 42)
	}

	// missing query -> ErrMissingQuery
	_, err = Query(values, "missing", ParseInt[int]())
	if !errors.Is(err, ErrMissingQuery) {
		t.Fatalf("expected ErrMissingQuery, got %v", err)
	}
}

func TestQueryNilParser(t *testing.T) {
	values := url.Values{}
	values.Set("n", "42")
	_, err := Query[string](values, "n", nil)
	if !errors.Is(err, ErrNilParser) {
		t.Fatalf("expected ErrNilParser, got %v", err)
	}
}

func TestQueryOrDefault(t *testing.T) {
	values := url.Values{}
	values.Set("n", "42")
	v := QueryOrDefault(values, "m", ParseInt[int](), 22)
	if v != 22 {
		t.Fatalf("expected 22, got %d", v)
	}
}

func TestParseIntOverflow(t *testing.T) {
	p := ParseInt[int8]()
	_, err := p("128") // 128 doesn't fit in int8
	if err == nil {
		t.Fatalf("expected overflow error for int8, got nil")
	}
	if !errors.Is(err, ErrOverflow) {
		t.Fatalf("expected ErrOverflow, got %v", err)
	}
}

func TestParseUintOverflow(t *testing.T) {
	p := ParseUint[uint8]()
	_, err := p("256") // 256 doesn't fit in uint8
	if err == nil {
		t.Fatalf("expected overflow error for uint8, got nil")
	}
	if !errors.Is(err, ErrOverflow) {
		t.Fatalf("expected ErrOverflow, got %v", err)
	}
}

func TestParseFloatSuccess(t *testing.T) {
	p := ParseFloat[float32]()
	v, err := p("3.14")
	if err != nil {
		t.Fatalf("unexpected error parsing float: %v", err)
	}
	if v < 3.139 || v > 3.141 {
		t.Fatalf("unexpected float value: %v", v)
	}
}

// Valid inputs whose canonical form differs from the input must parse.
func TestParseNonCanonicalInputs(t *testing.T) {
	ints := map[string]int{"+5": 5, "007": 7, "-0": 0, "-128": -128}
	for s, want := range ints {
		if got, err := ParseInt[int]()(s); err != nil || got != want {
			t.Errorf("ParseInt(%q) = %v, %v; want %v", s, got, err, want)
		}
	}
	if got, err := ParseInt[int8]()("-128"); err != nil || got != -128 {
		t.Errorf("ParseInt[int8](-128) = %v, %v; want -128", got, err)
	}
	if got, err := ParseUint[uint8]()("0255"); err != nil || got != 255 {
		t.Errorf("ParseUint[uint8](0255) = %v, %v; want 255", got, err)
	}
	floats := map[string]float64{"1.0": 1, "1.50": 1.5, "0.10": 0.1, "1e3": 1000}
	for s, want := range floats {
		if got, err := ParseFloat[float64]()(s); err != nil || got != want {
			t.Errorf("ParseFloat(%q) = %v, %v; want %v", s, got, err, want)
		}
	}
	// Precision loss is rounding, not overflow.
	if _, err := ParseFloat[float32]()("3.14159265358979"); err != nil {
		t.Errorf("ParseFloat[float32] precision loss: unexpected error %v", err)
	}
}

func TestParseFloatOverflow(t *testing.T) {
	if _, err := ParseFloat[float32]()("1e40"); !errors.Is(err, ErrOverflow) {
		t.Fatalf("expected ErrOverflow for float32 1e40, got %v", err)
	}
	if _, err := ParseInt[int]()("abc"); err == nil || errors.Is(err, ErrOverflow) {
		t.Fatalf("expected a syntax error, got %v", err)
	}
}

func TestGoRecoversFromPanic(t *testing.T) {
	// ensure the function doesn't let a panic escape
	done := make(chan struct{})
	Go(func() {
		defer close(done)
		panic("boom")
	})
	select {
	case <-done:
		// OK
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("goroutine did not complete in time")
	}
}
