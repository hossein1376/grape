# Changelog

## Version 0.7

### grape

- [BREAKING] Require Go 1.27, for generic methods and the `uuid` package.
- [BREAKING] `ReadJSON`, `ReadValidateJSON`, `Param` and `Query` return
  `errs.Error` values for client mistakes, so `ExtractFromErr` responds with
  the right status instead of 500: 415 for a non-JSON content type, 413 for a
  body over the limit, and 400 for malformed bodies, failed validation, and
  missing or invalid parameters. Their `Error()` text now starts with the
  status, e.g. `[415] content type must be application/json`. Underlying
  errors, such as `*json.SyntaxError`, stay in the chain.
- [BREAKING] `ReadOpts` is no longer generic. `WithMaxBodySize` now works with
  any type; it used to compile only for `ReadJSON[any]`.
- [BREAKING] The router's own 404 and 405 responses are JSON, and 405 keeps its
  `Allow` header.
- [BREAKING] Registering a conflicting route panics, as `http.ServeMux` does,
  instead of silently replacing the previous one.
- [BREAKING] `CORSMiddleware` no longer sends `Access-Control-Allow-Credentials`,
  which browsers reject alongside a `*` origin, and only answers real preflight
  requests. It must be registered with `UseAll` to see them.
- [BREAKING] `LoggerMiddleware` logs the connection's remote address as
  `client_ip`. The client-supplied `X-Forwarded-For` and `X-Real-Ip` headers
  are logged separately, as `forwarded_for` and `real_ip`.
- `ExtractFromErr` responds to a `validator.ValidationError` with 400 and the
  field errors as data, recognizes `*errs.Error`, and treats an `errs.Error`
  without a valid status code as an internal error instead of panicking.
- `RecoverMiddleware` responds with the same JSON 500 body as `ExtractFromErr`,
  including the request ID, and aborts the connection if the response had
  already started. It passes `http.ErrAbortHandler` panics on, and logs with
  the request context.
- `ReadJSON` accepts JSON-based content types such as
  `application/merge-patch+json`, and rejects look-alikes such as
  `application/jsonx`.
- `ReadValidateJSON` rejects a `null` body for pointer types, instead of
  calling `Validate` on a nil pointer.
- Fix `Group` with a previously used prefix returning a router without the
  middlewares added to the parent since, e.g. skipping an auth check.
- Fix a data race on the first requests, and serve routes and `UseAll`
  middlewares added after the server has started.
- Fix `Use` and `UseAll` reversing the caller's slice in place.
- Fix `ParseInt`, `ParseUint` and `ParseFloat` rejecting valid inputs such as
  `1.0`, `1e3`, `+5` and `007` as overflows.
- Fix `LoggerMiddleware` logging status 0 when nothing was written, and the
  last status of repeated `WriteHeader` calls instead of the first.
- Add `Hijack`, `Flush` and `Unwrap` to the logger's response writer, so
  WebSockets, streaming and `http.ResponseController` work through it.

### reqid

- [BREAKING] Request IDs are UUID v7 strings, instead of 26-character base32
  ones. They're still sortable by creation time.

### slogger

- Fix loggers derived with `With` and `WithGroup` dropping context attributes,
  such as the request ID.
- Fix contexts derived from the same parent overwriting each other's attributes
  in `WithAttrs`.

### validator

- Add rule chains with default messages: `v.String`, `v.Number` and `v.Slice`,
  e.g. `v.String("name", req.Name).Required().MaxLen(50)`. A chain stops at its
  first failing rule, and `Msg` replaces the message of a failed rule.
- Add `Nested` and `Each` for validating nested structs and slices, with keys
  such as `billing.city` and `items[1].qty`.
- Add `Err`, which returns nil when there are no errors.
- [BREAKING] `IsNumber` accepts only ASCII digits, of any length. It used to
  accept signs and reject numbers that don't fit in an `int`.
- A zero `Validator` is ready to use; it used to panic on the first error.

### Examples and documentation

- Pass grape's errors to `ExtractFromErr` directly, and document the rule
  chains and nested validation.
- Fix the validator example responding to validation failures with an empty
  400, and `getPermitByID` in the advanced example never sending its error.
- Register `CORSMiddleware` with `UseAll` in the advanced example.

## Version 0.6

### grape

- Improve `Param` and `Query` error messages.
- All `PATCH` method to the CORS middleware.
- Revert minimum required Go version to 1.22.
- Created new `ParamOrDefault` and `QueryOrDefault` functions.
- Fallback to `math/rand/v2` to generate request IDs if `crypto/rand` fails.
- Introduce the `Validator` interface and `ReadValidateJSON` function.
- Remove implicit Parse method in favor of explicit parse functions.
- Update documentations and test cases.

### errs

- Add new `WithErrMsg` optional func.
- Document options.

### slogger

- Add `String` and `Number` attribute functions.

### validator

- Added the `Zero` predicate.

## Version 0.5

- Fix a bug in logging response HTTP status code.
- [BREAKING] Rename `Json` to `JSON` in read and write methods.
- [BREAKING] Rename `RespondFromErr` to `ExtractFromErr`
- Allow `Param` to use type's Parse when parser is nil
- Fix a bug in the router where sometimes middlewares were incorrectly mutated.
- Improve `errs` package.
- Generate lexicographically, time-sortable random request IDs.
- Add tests and improve documentation.
- Bug fixes and improvements to various parts.

## Version 0.4

- Overhaul the whole project, with new API design and implementation. Too many changes to list here.

## Version 0.3

- Router: new ServeTLS method.
- Grape: new Go method for graceful panic recover in goroutines.
- Add linters. [PR#3](https://github.com/hossein1376/grape/pull/3)
- Fix net/http.Server security vulnerability. [PR#3](https://github.com/hossein1376/grape/pull/3)
- Return error as second value in ParamInt and ParamInt64. [PR#4](https://github.com/hossein1376/grape/pull/4)

Thanks to all of our contributors and users!

## Version 0.2

- Add route grouping and scope specific middlewares.
- Accepts interfaces in `grape.Options` struct. Allowing for better customizability.
- Any types implementing `grape.Logger`, including `slog.Logger`, can be used for logging.
- Use `grape.Serializer` to configure reading and writing json.
- Optionally, specify requests' max body size.
- Improvements to the validator package.
- New example, graceful shutdown.
- Improve documentation and update examples.

## Version 0.1

First release, including core features and ideas.
