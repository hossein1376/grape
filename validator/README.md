# Validator

## How to use

```go
package main

import (
	"fmt"
	"regexp"

	"github.com/hossein1376/grape/validator"
)

type data struct {
	Name  string
	Age   int
	Email string
	Phone string
}

func main() {
	simpleValidation()
	multiValidation()
}

func simpleValidation() {
	d := data{Name: "", Age: -2, Email: "asdf"}
	emailRX := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

	v := validator.New()
	v.Check("name", validator.Case{Cond: validator.NotEmpty(d.Name), Msg: "must not be empty"})
	v.Check("age", validator.Case{Cond: validator.Min(d.Age, 0), Msg: "must not be positive"})
	v.Check("email", validator.Case{Cond: validator.Matches(d.Email, emailRX), Msg: "must be valid email"})

	if ok := v.Valid(); !ok {
		fmt.Println("simpleValidation:", v.Errors)
		// name: must not be empty, age: must not be positive, email: must be valid email
		return
	}
	fmt.Println("Valid input!")
}

func multiValidation() {
	d := data{Name: "", Phone: "123a"}

	v := validator.New()
	v.Check("name",
		validator.Case{Cond: validator.NotEmpty(d.Name), Msg: "must not be empty"},
		validator.Case{Cond: validator.MinLength(d.Name, 2), Msg: "must not be less than 2 characters"},
	)
	v.Check("phone",
		validator.Case{Cond: validator.IsNumber(d.Phone), Msg: "must be only numbers"},
		validator.Case{Cond: validator.RangeLength(d.Phone, 4, 10), Msg: "must be between 4 to 10 digits"},
	)

	if ok := v.Valid(); !ok {
		fmt.Println("multiValidation:", v.Errors)
		// phone: must be only numbers, name: must not be empty, name: must not be less than 2 characters
		return
	}
	fmt.Println("Valid input!")
}

```