# Validator

Explicit, type-safe validation without struct tags or reflection.

Start with `validator.New()` (or a zero `Validator`), add rules per field, and
return `v.Err()`, which is `nil` when everything passed.

```go
type signup struct {
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Age      int      `json:"age"`
	Role     string   `json:"role"`
	Tags     []string `json:"tags"`
}

var emailRE = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

func (s signup) Validate() error {
	v := validator.New()
	v.String("username", s.Username).Required().LenRange(3, 20)
	v.String("email", s.Email).Required().Matches(emailRE).Msg("must be a valid email")
	v.Number("age", s.Age).Range(13, 130)
	v.String("role", s.Role).In("admin", "member")
	v.Slice("tags", s.Tags).MaxLen(5).Unique()
	return v.Err()
}
```

For `signup{Username: "", Email: "x", Age: 7, Role: "member"}`, `Validate`
returns a `validator.ValidationError`:

```json
{
  "age": ["must be between 13 and 130"],
  "email": ["must be a valid email"],
  "username": ["must not be empty"]
}
```

## Rule chains

| Chain | Rules |
|---|---|
| `v.String(key, s)` | `Required`, `MinLen`, `MaxLen`, `LenRange`, `Digits`, `Matches`, `In`, `StartsWith`, `EndsWith`, `Contains` |
| `v.Number(key, n)` | `NotZero`, `Min`, `Max`, `Range`, `In` |
| `v.Slice(key, s)` | `Required`, `Len`, `MinLen`, `MaxLen`, `LenRange`, `Unique`, `UniqueBy` |

- The value's type is inferred, including named types such as
  `type Role string`, and `In` takes values of that type.
- String lengths are counted in runes, not bytes.
- A chain stops at its first failing rule, so each field gets one error: an
  empty username reports "must not be empty", not also "must be at least 3
  characters".
- Every rule has a default message. `.Msg("...")` right after a rule replaces
  it, and only takes effect if that rule failed.
- `Unique` needs comparable items. For others, give `UniqueBy` a key:
  `v.Slice("users", s.Users).UniqueBy(func(u User) string { return u.Email })`.

## Custom conditions

For anything the chains don't cover, use `Check` with any boolean condition.
Unlike chains, it reports every failing case. The predicate functions, such as
`validator.Min`, `validator.LengthMax` and `validator.Unique`, can be used here
or on their own.

```go
v.Check("end_date",
	validator.Case{Cond: s.End.After(s.Start), Msg: "must be after start_date"},
)
```

## With grape

`grape.ReadValidateJSON` decodes the body and calls `Validate`.
`grape.ExtractFromErr` responds to a `ValidationError` with 400 and the field
errors:

```json
{"message": "Bad Request", "data": {"age": ["must be between 13 and 130"]}}
```

See [`_examples/validator`](../_examples/validator/main.go) for a complete
handler.
