package sdk

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestGetBuildsSignedRequestAndParsesResponse(t *testing.T) {
	var gotPath string
	var gotQuery url.Values
	var gotHeaders http.Header

	client := newTestSDKWithTransport("https://api.test/api/v4/", func(r *http.Request) (*http.Response, error) {
		gotPath = r.URL.Path
		gotQuery = r.URL.Query()
		gotHeaders = r.Header.Clone()

		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.Body != nil {
			defer r.Body.Close()
		}
		return testResponse(http.StatusOK, `{"status":{"code":1,"message":"ok"},"data":{"name":"edge"}}`, r), nil
	})

	resp, err := client.Get("/domains", ReqParams{
		Query: map[string]interface{}{
			"domain": "example.com",
			"page":   2,
		},
		Headers: map[string]string{"X-Trace-Id": "trace-1"},
	})
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if gotPath != "/api/v4/domains" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotQuery.Get("domain") != "example.com" || gotQuery.Get("page") != "2" {
		t.Fatalf("unexpected query: %s", gotQuery.Encode())
	}
	if gotQuery.Get("user_id") != "99" {
		t.Fatalf("unexpected user_id: %s", gotQuery.Get("user_id"))
	}
	if gotQuery.Get("algorithm") != "HMAC-SHA256" || gotQuery.Get("issued_at") == "" {
		t.Fatalf("missing auth query fields: %s", gotQuery.Encode())
	}
	if gotHeaders.Get("X-Auth-App-Id") != "app-id" {
		t.Fatalf("unexpected app id header: %s", gotHeaders.Get("X-Auth-App-Id"))
	}
	if gotHeaders.Get("X-Auth-Sign") == "" {
		t.Fatal("missing signature header")
	}
	if gotHeaders.Get("X-Trace-Id") != "trace-1" {
		t.Fatalf("unexpected custom header: %s", gotHeaders.Get("X-Trace-Id"))
	}
	if resp.HttpCode != http.StatusOK || resp.BizCode != 1 || resp.BizMsg != "ok" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	data, ok := resp.BizData.(map[string]interface{})
	if !ok || data["name"] != "edge" {
		t.Fatalf("unexpected biz data: %#v", resp.BizData)
	}
	if resp.RespData["status"] == nil {
		t.Fatalf("response data was not retained: %#v", resp.RespData)
	}
	if !strings.Contains(resp.Query, "domain=example.com") {
		t.Fatalf("unexpected response query: %s", resp.Query)
	}
}

func TestPostSendsJSONBodyAndParsesNumericStatus(t *testing.T) {
	var gotBody map[string]interface{}

	client := newTestSDKWithTransport("https://api.test", func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if r.URL.RawQuery != "" {
			t.Fatalf("unexpected query for POST: %s", r.URL.RawQuery)
		}
		return testResponse(http.StatusOK, `{"status":{"code":"2001","message":"created"},"data":[{"id":7}]}`, r), nil
	})

	resp, err := client.Post("domains", ReqParams{
		Data: map[string]interface{}{
			"domain": "example.com",
			"group":  "default",
		},
	})
	if err != nil {
		t.Fatalf("Post returned error: %v", err)
	}

	if gotBody["domain"] != "example.com" || gotBody["group"] != "default" {
		t.Fatalf("unexpected body: %#v", gotBody)
	}
	if gotBody["user_id"] != "99" || gotBody["client_userAgent"] == "" {
		t.Fatalf("missing payload fields: %#v", gotBody)
	}
	if resp.BizCode != 2001 || resp.BizMsg != "created" {
		t.Fatalf("unexpected business status: %d %s", resp.BizCode, resp.BizMsg)
	}
}

func TestRequestHandlesHTTPAndMalformedJSONErrors(t *testing.T) {
	t.Run("http status error", func(t *testing.T) {
		client := newTestSDKWithTransport("https://api.test", func(r *http.Request) (*http.Response, error) {
			return testResponse(http.StatusTooManyRequests, `{"status":{"code":429,"message":"too many"}}`, r), nil
		})

		resp, err := client.Get("limited", ReqParams{})
		if err == nil {
			t.Fatal("expected error")
		}
		if resp == nil || resp.HttpCode != http.StatusTooManyRequests {
			t.Fatalf("unexpected response: %#v", resp)
		}
		if resp.BizMsg != "response code is 429" {
			t.Fatalf("unexpected message: %s", resp.BizMsg)
		}
	})

	t.Run("invalid json error", func(t *testing.T) {
		client := newTestSDKWithTransport("https://api.test", func(r *http.Request) (*http.Response, error) {
			return testResponse(http.StatusOK, `not-json`, r), nil
		})

		resp, err := client.Get("invalid", ReqParams{})
		if err == nil {
			t.Fatal("expected error")
		}
		if resp == nil || !strings.Contains(resp.BizMsg, "json parse response body error") {
			t.Fatalf("unexpected response: %#v", resp)
		}
	})

	t.Run("invalid status format error", func(t *testing.T) {
		client := newTestSDKWithTransport("https://api.test", func(r *http.Request) (*http.Response, error) {
			return testResponse(http.StatusOK, `{"status":{"code":1},"data":{}}`, r), nil
		})

		resp, err := client.Get("invalid-status", ReqParams{})
		if err == nil {
			t.Fatal("expected error")
		}
		if resp == nil || resp.BizMsg != "the json format of response body status is invalid" {
			t.Fatalf("unexpected response: %#v", resp)
		}
	})
}

func TestGeneratedClientUsesLocalServer(t *testing.T) {
	var gotPath string
	var gotQuery url.Values

	sdk := newTestSDKWithTransport("https://api.test/api/v5", func(r *http.Request) (*http.Response, error) {
		gotPath = r.URL.Path
		gotQuery = r.URL.Query()
		return testResponse(http.StatusOK, `{"status":{"code":1,"message":"ok"},"data":{"total":0,"list":[]}}`, r), nil
	})

	client := NewEdgeNextClientFromSDK(sdk, WithLanguage("en"))
	resp, err := client.ListDomains(&ListDomainsRequest{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("ListDomains returned error: %v", err)
	}
	if gotPath != "/api/v5/domains" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotQuery.Get("page") != "1" || gotQuery.Get("page_size") != "20" {
		t.Fatalf("unexpected query: %s", gotQuery.Encode())
	}
	if resp.Status.Code != 1 {
		t.Fatalf("unexpected status: %+v", resp.Status)
	}
}

func TestDefaultTimeoutIsApplied(t *testing.T) {
	sdk := newTestSDKWithTransport("https://api.test", func(r *http.Request) (*http.Response, error) {
		return testResponse(http.StatusOK, `{"status":{"code":1,"message":"ok"},"data":{}}`, r), nil
	})
	sdk.Timeout = 0

	_, _ = sdk.Get("timeout-check", ReqParams{})

	if sdk.Timeout != defaultTimeout {
		t.Fatalf("unexpected default timeout: %d", sdk.Timeout)
	}
}

func newTestSDK(apiPre string) *Sdk {
	return &Sdk{
		AppId:     "app-id",
		AppSecret: "app-secret",
		ApiPre:    apiPre,
		UserId:    99,
		Timeout:   5,
	}
}

func newTestSDKWithTransport(apiPre string, fn func(*http.Request) (*http.Response, error)) *Sdk {
	sdk := newTestSDK(apiPre)
	sdk.httpClient = &http.Client{Transport: roundTripFunc(fn)}
	return sdk
}

func testResponse(statusCode int, body string, req *http.Request) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}
