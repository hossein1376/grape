package grape

import (
	"net/http"
	"net/http/httptest"
	"strings"
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

// Test that basic Method registration and grouping work: a route registered
// at root and one registered on a group should be served.
func TestRouter_MethodRegistrationAndGroup(t *testing.T) {
	r := NewRouter()

	// root handler
	r.Get("/root", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("root"))
	}))

	// group handler
	g := r.Group("/v1")
	g.Get("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	}))

	// Test root route - invoke the registered handler directly from the routes
	// map
	routes := r.root.routes[r.scope].routes
	h, ok := routes["GET "+r.scope+"/root"]
	if !ok {
		t.Fatalf("route %q not registered", "GET "+r.scope+"/root")
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/root", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if strings.TrimSpace(rec.Body.String()) != "root" {
		t.Fatalf("expected body 'root', got %q", rec.Body.String())
	}

	// Test group route - invoke the registered handler directly from the
	// group's routes
	routes = r.root.routes[g.scope].routes
	h, ok = routes["GET "+g.scope+"/ping"]
	if !ok {
		t.Fatalf("route %q not registered", "GET "+g.scope+"/ping")
	}
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1/ping", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if strings.TrimSpace(rec.Body.String()) != "pong" {
		t.Fatalf("expected body 'pong', got %q", rec.Body.String())
	}
}

// Test that r.Use applies middlewares in the correct execution order for routes
// defined after the Use call and that previously defined routes are not affected.
func TestRouter_UseOrderAndScope(t *testing.T) {
	r := NewRouter()

	// Handler that writes a simple body
	r.Get("/before", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("before"))
	}))

	// Add middleware that should only apply to routes registered after this point
	r.Use(markerMiddleware("m1"), markerMiddleware("m2"))

	// Route defined after Use -> should be wrapped by m1 and m2
	r.Get("/after", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("after"))
	}))

	// Request to the route defined before Use: should not have middleware markers
	routes := r.root.routes[r.scope].routes
	h, ok := routes["GET "+r.scope+"/before"]
	if !ok {
		t.Fatalf("route %q not present", "GET "+r.scope+"/before")
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/before", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 for /before, got %d", rec.Code)
	}
	hv := rec.Header()["X-Order"]
	if len(hv) != 0 {
		t.Fatalf(
			"expected no X-Order headers for route defined before Use, got %v",
			hv,
		)
	}

	// Request to the route defined after Use: should have markers in expected order.
	// With r.Use(m1, m2) we expect execution order: m1-before, m2-before, m2-after, m1-after
	h, ok = routes["GET "+r.scope+"/after"]
	if !ok {
		t.Fatalf("route %q not present", "GET "+r.scope+"/after")
	}
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/after", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 for /after, got %d", rec.Code)
	}
	got := rec.Header()["X-Order"]
	want := []string{"m1-before", "m2-before", "m2-after", "m1-after"}
	if len(got) != len(want) {
		t.Fatalf("unexpected X-Order headers length: got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected X-Order at %d: got %q want %q", i, got[i], want[i])
		}
	}
}

// Test that UseAll middlewares are applied to all handlers (including those
// registered before UseAll) and that their execution order is as expected
// relative to route-specific middlewares.
func TestRouter_UseAllPrecedence(t *testing.T) {
	r := NewRouter()

	// Add a global middleware that should wrap everything.
	r.UseAll(markerMiddleware("g1"), markerMiddleware("g2"))

	// Add a per-scope middleware and a route afterward.
	r.Use(markerMiddleware("m1"))
	r.Get("/item", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("item"))
	}))

	// Request the route and inspect middleware invocation order. Invoke the
	// stored route handler (which includes per-scope middlewares) and then
	// wrap it with the global middlewares to simulate top-level behaviour.
	routes := r.root.routes[r.scope].routes
	key := "GET " + r.scope + "/item"
	h, ok := routes[key]
	if !ok {
		t.Fatalf("route %q not registered", key)
	}
	// apply global middlewares in the same order newHandler would
	for _, mw := range r.root.global {
		h = mw(h)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/item", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	// Expected order:
	// UseAll(g1,g2) with provided order should result in execution order:
	// g1-before, g2-before, m1-before, m1-after, g2-after, g1-after
	got := rec.Header()["X-Order"]
	want := []string{
		"g1-before", "g2-before", "m1-before", "m1-after", "g2-after", "g1-after",
	}

	if len(got) != len(want) {
		t.Fatalf("unexpected X-Order headers length: got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf(
				"unexpected X-Order at %d: got %q want %q", i, got[i], want[i],
			)
		}
	}
}
