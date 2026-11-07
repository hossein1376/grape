package validator

import (
	"errors"
	"fmt"
)

// Nested adds the errors of a nested value's validation under key, so a
// struct field can be validated by the field type's own Validate method:
//
//	v.Nested("address", req.Address.Validate())
//
// The field errors of a [ValidationError] in err's tree are added with their
// keys prefixed, e.g. "address.city". Empty parts are left out: with an empty
// key the fields are added as is, and a field with an empty key is added
// under key.
// Any other non-nil error is added under key, with its text as the message. A
// nil error, or an empty ValidationError, adds nothing.
func (v *Validator) Nested(key string, err error) {
	if err == nil {
		return
	}
	var fields ValidationError
	if !errors.As(err, &fields) {
		v.addError(key, err.Error())
		return
	}
	for field, msgs := range fields {
		for _, msg := range msgs {
			v.addError(joinKey(key, field), msg)
		}
	}
}

// joinKey joins key and a nested field with a dot, leaving out empty parts.
func joinKey(key, field string) string {
	switch {
	case key == "":
		return field
	case field == "":
		return key
	default:
		return key + "." + field
	}
}

// Each validates every item of a slice with the given function, adding the
// errors of each item as with [Validator.Nested], under key and the item's
// index, e.g. "items[0].name".
//
//	v.Each("items", req.Items, Item.Validate)
//	v.Each("emails", req.Emails, func(email string) error {
//		v := validator.New()
//		v.String("", email).Matches(emailRE)
//		return v.Err()
//	})
func (v *Validator) Each[S ~[]E, E any](key string, items S, validate func(E) error) {
	for i, item := range items {
		v.Nested(fmt.Sprintf("%s[%d]", key, i), validate(item))
	}
}
