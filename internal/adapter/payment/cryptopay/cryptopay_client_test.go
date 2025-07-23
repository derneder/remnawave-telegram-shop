package cryptopay_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"remnawave-tg-shop-bot/internal/adapter/payment/cryptopay"
)

type invoiceRequest struct {
	Amount string `json:"amount"`
}

func TestClientCreateInvoice(t *testing.T) {
	var called bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/createInvoice" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Crypto-Pay-API-Token") != "tok" {
			t.Fatalf("missing token header")
		}
		var req invoiceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if req.Amount != "10" {
			t.Fatalf("unexpected amount %s", req.Amount)
		}
		called = true
		resp := cryptopay.ResponseWrapper[cryptopay.InvoiceResponse]{
			Ok:     true,
			Result: cryptopay.InvoiceResponse{InvoiceID: func() *int64 { v := int64(1); return &v }()},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := cryptopay.NewCryptoPayClient(srv.URL, "tok")
	inv, err := c.CreateInvoice(&cryptopay.InvoiceRequest{Amount: "10"})
	if err != nil {
		t.Fatalf("CreateInvoice: %v", err)
	}
	if inv == nil || inv.InvoiceID == nil || *inv.InvoiceID != 1 {
		t.Fatalf("unexpected invoice %+v", inv)
	}
	if !called {
		t.Fatal("handler not called")
	}
}

func TestClientGetInvoices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/getInvoices" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("status") != "paid" {
			t.Fatalf("unexpected query %q", r.URL.RawQuery)
		}
		resp := cryptopay.ResponseListWrapper[cryptopay.InvoiceResponse]{
			Ok:     true,
			Result: cryptopay.ResultListWrapper[cryptopay.InvoiceResponse]{Items: []cryptopay.InvoiceResponse{{Status: "paid"}}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := cryptopay.NewCryptoPayClient(srv.URL, "tok")
	invs, err := c.GetInvoices("paid", "", "", "", 0, 0)
	if err != nil {
		t.Fatalf("GetInvoices: %v", err)
	}
	if invs == nil || len(*invs) != 1 || (*invs)[0].Status != "paid" {
		t.Fatalf("unexpected invoices: %#v", invs)
	}
}
