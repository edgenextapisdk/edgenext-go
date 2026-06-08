package core

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestCanonicalRequestSortsAndEscapesQuery(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://api.example.com/api/v5/domains?b=two+words&a=1&b=alpha&space=a%20b", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set(HeaderXContentSha256, "payload-hash")

	canonical, err := CanonicalRequest(req, SignedHeaders(req))
	if err != nil {
		t.Fatalf("CanonicalRequest returned error: %v", err)
	}

	expected := "GET\n/api/v5/domains/\na=1&b=alpha&b=two%20words&space=a%20b\npayload-hash"
	if canonical != expected {
		t.Fatalf("unexpected canonical request:\n%s", canonical)
	}
	if req.URL.RawQuery != "a=1&b=alpha&b=two%20words&space=a%20b" {
		t.Fatalf("unexpected raw query: %s", req.URL.RawQuery)
	}
}

func TestRequestPayloadRestoresBody(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "https://api.example.com/api/v5/domains", strings.NewReader(`{"name":"edge"}`))
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	payload, err := RequestPayload(req)
	if err != nil {
		t.Fatalf("RequestPayload returned error: %v", err)
	}
	if string(payload) != `{"name":"edge"}` {
		t.Fatalf("unexpected payload: %s", payload)
	}

	restored, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("read restored body: %v", err)
	}
	if string(restored) != `{"name":"edge"}` {
		t.Fatalf("body was not restored: %s", restored)
	}
}

func TestSignerSignAndVerify(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "https://api.example.com/api/v5/domains?domain=example.com", strings.NewReader(`{"group":"default"}`))
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set(HeaderXDateTime, "20240521T120000Z")
	req.Header.Set("Content-Type", "application/json")

	signer := &Signer{AppId: "app-id", AppSecret: "app-secret"}
	if err := signer.Sign(req); err != nil {
		t.Fatalf("Sign returned error: %v", err)
	}

	signature := req.Header.Get("X-Auth-Sign")
	if !strings.HasPrefix(signature, "Bearer ") {
		t.Fatalf("unexpected signature header: %s", signature)
	}
	if req.Header.Get(HeaderXDateTime) != "20240521T120000Z" {
		t.Fatalf("signer changed explicit date: %s", req.Header.Get(HeaderXDateTime))
	}
	if err := signer.Verify(req); err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}

	req.URL.RawQuery = "domain=changed.example.com"
	if err := signer.Verify(req); err == nil {
		t.Fatal("expected verify failure after query mutation")
	}
}

func TestSignerAddsDateWhenMissing(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://api.example.com/api/v5/domains", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	if err := (&Signer{AppSecret: "app-secret"}).Sign(req); err != nil {
		t.Fatalf("Sign returned error: %v", err)
	}
	if req.Header.Get(HeaderXDateTime) == "" {
		t.Fatal("missing signer date header")
	}
}

func TestHexEncodeSHA256Hash(t *testing.T) {
	got, err := HexEncodeSHA256Hash([]byte("hello"))
	if err != nil {
		t.Fatalf("HexEncodeSHA256Hash returned error: %v", err)
	}
	if got != "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" {
		t.Fatalf("unexpected hash: %s", got)
	}
}
