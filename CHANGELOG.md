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