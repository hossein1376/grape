package grape

import (
	"errors"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"
)

// customType demonstrates a type that implements Parse(s string) (T, error)
// so parse(..., nil) should call this method.
type customType struct {
	V string
}

func (c customType) Parse(s string) (customType, error) {
	return customType{V: s}, nil
}

func TestParamMissing(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	// No path parameter set -> should return ErrMissingParam
	_, err := Param[int](req, "id", ParseInt[int]())
	if !errors.Is(err, ErrMissingParam) {
		t.Fatalf("expected ErrMissingParam, got %v", err)
	}
}

func TestQueryParsingSuccessAndMissing(t *testing.T) {
	values := url.Values{}
	values.Set("n", "42")

	v, err := Query[int](values, "n", ParseInt[int]())
	if err != nil {
		t.Fatalf("unexpected error parsing query: %v", err)
	}
	if v != 42 {
		t.Fatalf("unexpected parsed value: got %d want %d", v, 42)
	}

	// missing query -> ErrMissingQuery
	_, err = Query[string](values, "missing", nil)
	if !errors.Is(err, ErrMissingQuery) {
		t.Fatalf("expected ErrMissingQuery, got %v", err)
	}
}

func TestParseWithNilParserUsesTypeParseMethod(t *testing.T) {
	var zero customType
	got, err := parse(zero, "hello", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.V != "hello" {
		t.Fatalf("expected parsed value 'hello', got %q", got.V)
	}
}

func TestParseWithProvidedParserFunction(t *testing.T) {
	var zero int
	res, err := parse(zero, "100", func(s string) (int, error) { return strconv.Atoi(s) })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != 100 {
		t.Fatalf("expected 100 got %d", res)
	}
}

// ParseInt/ParseUint overflow checks rely on checkOverflow detecting that the
// string representation after casting does not match the original string.
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

func TestCheckOverflowMismatch(t *testing.T) {
	var tval int8 = -128
	_, err := checkOverflow("128", tval)
	if err == nil {
		t.Fatalf("expected ErrOverflow from checkOverflow, got nil")
	}
	if !errors.Is(err, ErrOverflow) {
		t.Fatalf("expected ErrOverflow, got %v", err)
	}
}

func TestGoRecoversFromPanic(t *testing.T) {
	// ensure Go doesn't let a panic escape
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
