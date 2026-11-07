package grape

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
)

// helper middleware generator that appends markers before and after calling next
func markerMiddleware(name string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("X-Order", name+"-before")
			next.ServeHTTP(w, r)
			w.Header().Add("X-Order", name+"-after")
		})
	}
}

func writeBody(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(body))
	}
}

// serve sends a request through the router and returns the recorded response.
func serve(r *Router, method, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	return rec
}

func assertOrder(t *testing.T, rec *httptest.ResponseRecorder, want ...string) {
	t.Helper()
	if got := rec.Header()["X-Order"]; !slices.Equal(got, want) {
		t.Fatalf("unexpected X-Order headers: got %v want %v", got, want)
	}
}

// Test that basic Method registration and grouping work: a route registered
// at root and one registered on a group should be served.
func TestRouter_MethodRegistrationAndGroup(t *testing.T) {
	r := NewRouter()
	r.Get("/root", writeBody("root"))
	r.Group("/v1").Get("/ping", writeBody("pong"))

	for path, want := range map[string]string{"/root": "root", "/v1/ping": "pong"} {
		rec := serve(r, http.MethodGet, path)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: expected status 200, got %d", path, rec.Code)
		}
		if body := strings.TrimSpace(rec.Body.String()); body != want {
			t.Fatalf("%s: expected body %q, got %q", path, want, body)
		}
	}
}

// Test that r.Use applies middlewares in the correct execution order for routes
// defined after the Use call and that previously defined routes are not affected.
func TestRouter_UseOrderAndScope(t *testing.T) {
	r := NewRouter()
	r.Get("/before", writeBody("before"))
	r.Use(markerMiddleware("m1"), markerMiddleware("m2"))
	r.Get("/after", writeBody("after"))

	assertOrder(t, serve(r, http.MethodGet, "/before"))
	assertOrder(t, serve(r, http.MethodGet, "/after"),
		"m1-before", "m2-before", "m2-after", "m1-after")
}

// Test that UseAll middlewares are applied to all handlers (including those
// registered before UseAll) and that their execution order is as expected
// relative to route-specific middlewares.
func TestRouter_UseAllPrecedence(t *testing.T) {
	r := NewRouter()
	r.Get("/early", writeBody("early"))
	r.UseAll(markerMiddleware("g1"), markerMiddleware("g2"))
	r.Use(markerMiddleware("m1"))
	r.Get("/item", writeBody("item"))

	assertOrder(t, serve(r, http.MethodGet, "/early"),
		"g1-before", "g2-before", "g2-after", "g1-after")
	assertOrder(t, serve(r, http.MethodGet, "/item"),
		"g1-before", "g2-before", "m1-before", "m1-after", "g2-after", "g1-after")
	// Default handlers are wrapped too.
	assertOrder(t, serve(r, http.MethodGet, "/missing"),
		"g1-before", "g2-before", "g2-after", "g1-after")
}

// Calling Group again with the same prefix must not hand back a router that
// lacks middlewares added to the parent in between, e.g. an auth check.
func TestRouter_GroupSamePrefixKeepsLaterMiddlewares(t *testing.T) {
	r := NewRouter()
	r.Group("/api").Get("/public", writeBody("public"))
	r.Use(func(http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		})
	})
	r.Group("/api").Get("/secret", writeBody("secret"))

	if code := serve(r, http.MethodGet, "/api/public").Code; code != http.StatusOK {
		t.Fatalf("/api/public: expected 200, got %d", code)
	}
	if code := serve(r, http.MethodGet, "/api/secret").Code; code != http.StatusUnauthorized {
		t.Fatalf("/api/secret: expected 401 from middleware, got %d", code)
	}
}

func TestRouter_RoutesAndUseAllAfterFirstRequest(t *testing.T) {
	r := NewRouter()
	r.Get("/a", writeBody("a"))
	serve(r, http.MethodGet, "/a")

	r.Get("/b", writeBody("b"))
	if code := serve(r, http.MethodGet, "/b").Code; code != http.StatusOK {
		t.Fatalf("route added after first request: expected 200, got %d", code)
	}
	r.UseAll(markerMiddleware("g"))
	assertOrder(t, serve(r, http.MethodGet, "/a"), "g-before", "g-after")
}

// Run with -race: the handler is built on the first request.
func TestRouter_ConcurrentFirstRequests(t *testing.T) {
	r := NewRouter()
	r.UseAll(markerMiddleware("g"))
	r.Get("/a", writeBody("a"))

	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() { serve(r, http.MethodGet, "/a") })
	}
	wg.Wait()
}

func TestRouter_UseDoesNotModifyArgument(t *testing.T) {
	m1, m2 := markerMiddleware("m1"), markerMiddleware("m2")
	mws := []func(http.Handler) http.Handler{m1, m2}
	r := NewRouter()
	r.Use(mws...)
	r.UseAll(mws...)
	r.Get("/", writeBody(""))

	// Had Use reversed mws in place, UseAll would register them as m2, m1.
	assertOrder(t, serve(r, http.MethodGet, "/"),
		"m1-before", "m2-before", "m1-before", "m2-before",
		"m2-after", "m1-after", "m2-after", "m1-after")
}
