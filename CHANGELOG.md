# Changelog

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
