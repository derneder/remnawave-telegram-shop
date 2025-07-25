package tribute

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientListOrders(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Api-Key") != "k" {
			t.Fatalf("missing header")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":1}]`))
	}))
	defer ts.Close()
	c := New("k", WithBaseURL(ts.URL), WithHTTPClient(ts.Client()))
	orders, err := c.ListOrders(context.Background(), 0, 0)
	if err != nil || len(orders) != 1 || orders[0].ID != 1 {
		t.Fatalf("unexpected result %v %v", orders, err)
	}
}

func TestClientNetworkError(t *testing.T) {
	rt := roundTripFunc(func(*http.Request) (*http.Response, error) { return nil, errors.New("boom") })
	hc := &http.Client{Transport: rt}
	c := New("k", WithHTTPClient(hc), WithBaseURL("http://example"))
	if _, err := c.ListOrders(context.Background(), 0, 0); err == nil {
		t.Fatal("expected error")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
