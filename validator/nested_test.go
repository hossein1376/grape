package validator

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

type address struct{ City, Zip string }

func (a address) Validate() error {
	v := New()
	v.String("city", a.City).Required()
	v.String("zip", a.Zip).Digits()
	return v.Err()
}

type lineItem struct {
	SKU  string
	Qty  int
	Ship *address
}

// Pointer receiver, used through a []*lineItem.
func (l *lineItem) Validate() error {
	v := New()
	v.String("sku", l.SKU).Required()
	v.Number("qty", l.Qty).Min(1)
	if l.Ship != nil {
		v.Nested("ship", l.Ship.Validate())
	}
	return v.Err()
}

type order struct {
	Billing address
	Items   []*lineItem
	Emails  []string
}

func (o order) Validate() error {
	v := New()
	v.Nested("billing", o.Billing.Validate())
	v.Slice("items", o.Items).Required()
	v.Each("items", o.Items, (*lineItem).Validate)
	v.Each("emails", o.Emails, func(email string) error {
		v := New()
		v.String("", email).Contains("@")
		return v.Err()
	})
	return v.Err()
}

func TestNestedAndEach(t *testing.T) {
	o := order{
		Billing: address{City: "", Zip: "12a"},
		Items: []*lineItem{
			{SKU: "a", Qty: 1},
			{SKU: "", Qty: 0, Ship: &address{City: "", Zip: "1"}},
		},
		Emails: []string{"a@x.io", "nope"},
	}
	var got ValidationError
	if !errors.As(o.Validate(), &got) {
		t.Fatalf("expected a ValidationError, got %v", o.Validate())
	}
	want := ValidationError{
		"billing.city":       {"must not be empty"},
		"billing.zip":        {"must contain only digits"},
		"items[1].sku":       {"must not be empty"},
		"items[1].qty":       {"must be at least 1"},
		"items[1].ship.city": {"must not be empty"},
		"emails[1]":          {`must contain "@"`},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v\nwant %v", got, want)
	}

	valid := order{
		Billing: address{City: "x", Zip: "1"},
		Items:   []*lineItem{{SKU: "a", Qty: 1}},
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected a valid order, got %v", err)
	}
}

func TestNested(t *testing.T) {
	fields := ValidationError{"city": {"required"}}
	tests := []struct {
		name string
		key  string
		err  error
		want ValidationError
	}{
		{"nil error", "address", nil, nil},
		{"empty ValidationError", "address", ValidationError{}, nil},
		{"wrapped ValidationError", "address", fmt.Errorf("x: %w", fields), ValidationError{"address.city": {"required"}}},
		{"empty key keeps fields", "", fields, ValidationError{"city": {"required"}}},
		{"plain error", "address", errors.New("unknown city"), ValidationError{"address": {"unknown city"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v Validator
			v.Nested(tt.key, tt.err)
			if len(tt.want) == 0 {
				if len(v.Errors) != 0 {
					t.Fatalf("expected no errors, got %v", v.Errors)
				}
				return
			}
			if !reflect.DeepEqual(v.Errors, tt.want) {
				t.Fatalf("got %v, want %v", v.Errors, tt.want)
			}
		})
	}
}

// Nested adds to errors already recorded for the same key.
func TestNested_MergesWithExistingErrors(t *testing.T) {
	v := New()
	v.Check("address.city", Case{Cond: false, Msg: "first"})
	v.Nested("address", ValidationError{"city": {"second"}})
	want := ValidationError{"address.city": {"first", "second"}}
	if !reflect.DeepEqual(v.Errors, want) {
		t.Fatalf("got %v, want %v", v.Errors, want)
	}
}
