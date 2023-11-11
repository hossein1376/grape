# Grape 🍇

Grape is a modern, zero-dependency HTTP library for Go.

It's a thin wrapper around the standard library, providing helper functions to facilitate faster and easier development.
Adding only a single dependency to your projects.

## Why?

Go's standard library is awesome. It's fast, easy to use, and has a great API.  
With the addition of log/slog in go1.21 and improved HTTP router in go1.22, there is no reason to look further.
Instead of breaking compatibility with net/http, Grape aims to add commonly used functions within the arm's reach of the
handlers.
