package sdk

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestWrapperListDomainsDemo(t *testing.T) {
	var gotMethod string
	var gotPath string
	var gotQuery url.Values
	var gotHeaders http.Header

	sdk := newTestSDKWithTransport("https://api.test/api/v5", func(r *http.Request) (*http.Response, error) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.Query()
		gotHeaders = r.Header.Clone()
		return testResponse(http.StatusOK, `{"status":{"code":1,"message":"ok"},"data":{"total":1,"list":[{"domain":"example.com"}]}}`, r), nil
	})

	client := NewEdgeNextClientFromSDK(sdk, WithLanguage("en"))
	client.SetDefaultHeader("X-Client", "demo")

	resp, err := client.ListDomains((&ListDomainsRequest{
		Page:     1,
		PageSize: 20,
		Domain:   "example.com",
	}).SetHeader("X-Request-Id", "req-1"))
	if err != nil {
		t.Fatalf("ListDomains returned error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Fatalf("unexpected method: %s", gotMethod)
	}
	if gotPath != "/api/v5/domains" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotQuery.Get("page") != "1" || gotQuery.Get("page_size") != "20" || gotQuery.Get("domain") != "example.com" {
		t.Fatalf("unexpected query: %s", gotQuery.Encode())
	}
	if gotHeaders.Get("X-Lang") != "en" || gotHeaders.Get("X-Client") != "demo" || gotHeaders.Get("X-Request-Id") != "req-1" {
		t.Fatalf("unexpected headers: %#v", gotHeaders)
	}
	if resp.Status.Code != 1 {
		t.Fatalf("unexpected biz code: %d", resp.Status.Code)
	}
}

func TestWrapperClientConfigDemo(t *testing.T) {
	client := NewEdgeNextClient(EdgeNextClientConfig{
		AppId:     "app-id",
		AppSecret: "app-secret",
		ApiPre:    "https://api.test/api/v5",
		UserId:    99,
		Timeout:   5,
		Debug:     true,
	}, WithLanguage("en"))

	if client.SDK == nil {
		t.Fatal("expected sdk to be initialized")
	}
	if client.SDK.AppId != "app-id" || client.SDK.AppSecret != "app-secret" || client.SDK.ApiPre != "https://api.test/api/v5" {
		t.Fatalf("unexpected sdk config: %#v", client.SDK)
	}
	if client.SDK.UserId != 99 || client.SDK.Timeout != 5 || !client.SDK.Debug {
		t.Fatalf("unexpected sdk options: %#v", client.SDK)
	}
	if client.defaultHeaders["X-Lang"] != "en" {
		t.Fatalf("unexpected default language header: %s", client.defaultHeaders["X-Lang"])
	}
}

func TestWrapperRawMethodsDemo(t *testing.T) {
	tests := []struct {
		name       string
		call       func(*EdgeNextClient) (*APIResponse, error)
		wantMethod string
		wantPath   string
		wantBody   bool
	}{
		{
			name: "get",
			call: func(client *EdgeNextClient) (*APIResponse, error) {
				return client.Get("raw/get", ReqParams{Query: map[string]interface{}{"page": 1}})
			},
			wantMethod: http.MethodGet,
			wantPath:   "/api/v5/raw/get",
		},
		{
			name: "post",
			call: func(client *EdgeNextClient) (*APIResponse, error) {
				return client.Post("raw/post", ReqParams{Data: map[string]interface{}{"name": "post-demo"}})
			},
			wantMethod: http.MethodPost,
			wantPath:   "/api/v5/raw/post",
			wantBody:   true,
		},
		{
			name: "put",
			call: func(client *EdgeNextClient) (*APIResponse, error) {
				return client.Put("raw/put", ReqParams{Data: map[string]interface{}{"name": "put-demo"}})
			},
			wantMethod: http.MethodPut,
			wantPath:   "/api/v5/raw/put",
			wantBody:   true,
		},
		{
			name: "delete",
			call: func(client *EdgeNextClient) (*APIResponse, error) {
				return client.Delete("raw/delete", ReqParams{Data: map[string]interface{}{"id": 123}})
			},
			wantMethod: http.MethodDelete,
			wantPath:   "/api/v5/raw/delete",
			wantBody:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMethod string
			var gotPath string
			var gotHeader string
			var gotBody map[string]interface{}

			sdk := newTestSDKWithTransport("https://api.test/api/v5", func(r *http.Request) (*http.Response, error) {
				gotMethod = r.Method
				gotPath = r.URL.Path
				gotHeader = r.Header.Get("X-Client")
				if tt.wantBody {
					decodeRequestBody(t, r, &gotBody)
				}
				return testResponse(http.StatusOK, `{"status":{"code":1,"message":"ok"},"data":{}}`, r), nil
			})
			client := NewEdgeNextClientFromSDK(sdk, WithDefaultHeaders(map[string]string{"X-Client": "raw-demo"}))

			resp, err := tt.call(client)
			if err != nil {
				t.Fatalf("raw call returned error: %v", err)
			}
			if gotMethod != tt.wantMethod {
				t.Fatalf("unexpected method: %s", gotMethod)
			}
			if gotPath != tt.wantPath {
				t.Fatalf("unexpected path: %s", gotPath)
			}
			if gotHeader != "raw-demo" {
				t.Fatalf("unexpected default header: %s", gotHeader)
			}
			if tt.wantBody && gotBody["algorithm"] != "HMAC-SHA256" {
				t.Fatalf("missing sdk payload fields: %#v", gotBody)
			}
			if resp.Status.Code != 1 {
				t.Fatalf("unexpected biz code: %d", resp.Status.Code)
			}
		})
	}
}

func TestWrapperAddDomainsDemo(t *testing.T) {
	var gotMethod string
	var gotPath string
	var gotBody map[string]interface{}

	sdk := newTestSDKWithTransport("https://api.test/api/v5", func(r *http.Request) (*http.Response, error) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		decodeRequestBody(t, r, &gotBody)
		return testResponse(http.StatusOK, `{"status":{"code":1,"message":"created"},"data":{"id":123}}`, r), nil
	})

	resp, err := NewEdgeNextClientFromSDK(sdk).AddDomains(&AddDomainsRequest{
		Domain:  "example.com",
		GroupId: 10,
		Origins: []map[string]interface{}{
			{"addr": "192.0.2.10", "weight": 1},
		},
		Remark: "demo domain",
	})
	if err != nil {
		t.Fatalf("AddDomains returned error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Fatalf("unexpected method: %s", gotMethod)
	}
	if gotPath != "/api/v5/domains" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotBody["domain"] != "example.com" || gotBody["group_id"] != float64(10) || gotBody["remark"] != "demo domain" {
		t.Fatalf("unexpected body: %#v", gotBody)
	}
	if gotBody["algorithm"] != "HMAC-SHA256" || gotBody["user_id"] != "99" {
		t.Fatalf("missing sdk payload fields: %#v", gotBody)
	}
	if resp.Data.Id != float64(123) {
		t.Fatalf("unexpected response data: %#v", resp.Data)
	}
}

func TestWrapperUpdateDomainsDemo(t *testing.T) {
	var gotMethod string
	var gotPath string
	var gotBody map[string]interface{}

	sdk := newTestSDKWithTransport("https://api.test/api/v5", func(r *http.Request) (*http.Response, error) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		decodeRequestBody(t, r, &gotBody)
		return testResponse(http.StatusOK, `{"status":{"code":1,"message":"updated"},"data":{}}`, r), nil
	})

	resp, err := NewEdgeNextClientFromSDK(sdk).UpdateDomains(&UpdateDomainsRequest{
		DomainId: 123,
		Remark:   "updated by wrapper demo",
	})
	if err != nil {
		t.Fatalf("UpdateDomains returned error: %v", err)
	}

	if gotMethod != http.MethodPut {
		t.Fatalf("unexpected method: %s", gotMethod)
	}
	if gotPath != "/api/v5/domains" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotBody["domain_id"] != float64(123) || gotBody["remark"] != "updated by wrapper demo" {
		t.Fatalf("unexpected body: %#v", gotBody)
	}
	if resp.Status.Message != "updated" {
		t.Fatalf("unexpected biz message: %s", resp.Status.Message)
	}
}

func TestWrapperDeleteDomainsDemo(t *testing.T) {
	var gotMethod string
	var gotPath string
	var gotBody map[string]interface{}

	sdk := newTestSDKWithTransport("https://api.test/api/v5", func(r *http.Request) (*http.Response, error) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		decodeRequestBody(t, r, &gotBody)
		return testResponse(http.StatusOK, `{"status":{"code":1,"message":"deleted"},"data":{}}`, r), nil
	})

	resp, err := NewEdgeNextClientFromSDK(sdk).DeleteDomains(&DeleteDomainsRequest{
		Ids: []int{123, 456},
	})
	if err != nil {
		t.Fatalf("DeleteDomains returned error: %v", err)
	}

	if gotMethod != http.MethodDelete {
		t.Fatalf("unexpected method: %s", gotMethod)
	}
	if gotPath != "/api/v5/domains" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	ids, ok := gotBody["ids"].([]interface{})
	if !ok || len(ids) != 2 || ids[0] != float64(123) || ids[1] != float64(456) {
		t.Fatalf("unexpected ids body: %#v", gotBody["ids"])
	}
	if resp.Status.Message != "deleted" {
		t.Fatalf("unexpected biz message: %s", resp.Status.Message)
	}
}

func TestWrapperRefreshDomainsAccessDemo(t *testing.T) {
	var gotPath string
	var gotBody map[string]interface{}

	sdk := newTestSDKWithTransport("https://api.test/api/v5", func(r *http.Request) (*http.Response, error) {
		gotPath = r.URL.Path
		decodeRequestBody(t, r, &gotBody)
		return testResponse(http.StatusOK, `{"status":{"code":1,"message":"refreshed"},"data":{"domain_ids":[123]}}`, r), nil
	})

	resp, err := NewEdgeNextClientFromSDK(sdk).RefreshDomainsAccess(&RefreshDomainsAccessRequest{
		DomainIds: []int{123},
	})
	if err != nil {
		t.Fatalf("RefreshDomainsAccess returned error: %v", err)
	}

	if gotPath != "/api/v5/domains/access_refresh" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if _, ok := gotBody["domain_ids"].([]interface{}); !ok {
		t.Fatalf("unexpected body: %#v", gotBody)
	}
	if len(resp.Data.DomainIds) != 1 || resp.Data.DomainIds[0] != float64(123) {
		t.Fatalf("unexpected response data: %#v", resp.Data)
	}
}

func TestWrapperWebCdnCleanCacheSaveCacheDemo(t *testing.T) {
	var gotMethod string
	var gotPath string
	var gotBody map[string]interface{}

	sdk := newTestSDKWithTransport("https://api.test/api/v5", func(r *http.Request) (*http.Response, error) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		decodeRequestBody(t, r, &gotBody)
		return testResponse(http.StatusOK, `{"status":{"code":1,"message":"ok"},"data":{"task_id":88}}`, r), nil
	})

	resp, err := NewEdgeNextClientFromSDK(sdk).WebCdnCleanCacheSaveCache(&WebCdnCleanCacheSaveCacheRequest{
		GroupId:    10,
		Protocol:   "https",
		Port:       443,
		Wholesite:  0,
		Specialurl: []string{"https://example.com/a.js"},
	})
	if err != nil {
		t.Fatalf("WebCdnCleanCacheSaveCache returned error: %v", err)
	}

	if gotMethod != http.MethodPut {
		t.Fatalf("unexpected method: %s", gotMethod)
	}
	if gotPath != "/api/v5/Web.Domain.DashBoard.saveCache" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotBody["group_id"] != float64(10) || gotBody["protocol"] != "https" || gotBody["port"] != float64(443) {
		t.Fatalf("unexpected body: %#v", gotBody)
	}
	if resp.Data.TaskId != float64(88) {
		t.Fatalf("unexpected response data: %#v", resp.Data)
	}
}

func TestWrapperCallAPIByNameDemo(t *testing.T) {
	var gotPath string
	var gotQuery url.Values
	var gotHeader string

	sdk := newTestSDKWithTransport("https://api.test/api/v5", func(r *http.Request) (*http.Response, error) {
		gotPath = r.URL.Path
		gotQuery = r.URL.Query()
		gotHeader = r.Header.Get("X-Request-Id")
		return testResponse(http.StatusOK, `{"status":{"code":1,"message":"ok"},"data":{"total":0,"list":[]}}`, r), nil
	})

	_, err := NewEdgeNextClientFromSDK(sdk).CallAPI(
		"ListDomains",
		nil,
		map[string]interface{}{"page": 2, "page_size": 50},
		nil,
		map[string]string{"X-Request-Id": "req-2"},
		http.MethodGet,
	)
	if err != nil {
		t.Fatalf("CallAPI returned error: %v", err)
	}

	if gotPath != "/api/v5/domains" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotQuery.Get("page") != "2" || gotQuery.Get("page_size") != "50" {
		t.Fatalf("unexpected query: %s", gotQuery.Encode())
	}
	if gotHeader != "req-2" {
		t.Fatalf("unexpected request header: %s", gotHeader)
	}
}

func TestWrapperInvalidCallsDemo(t *testing.T) {
	sdk := newTestSDKWithTransport("https://api.test/api/v5", func(r *http.Request) (*http.Response, error) {
		t.Fatal("transport should not be called")
		return nil, nil
	})
	client := NewEdgeNextClientFromSDK(sdk)

	_, err := client.CallAPI("MissingAPI", nil, nil, nil, nil, http.MethodGet)
	if err == nil || !strings.Contains(err.Error(), "unknown EdgeNext API") {
		t.Fatalf("unexpected unknown api error: %v", err)
	}

	_, err = client.CallAPI("ListDomains", &AddDomainsRequest{}, nil, nil, nil, "")
	if err == nil || !strings.Contains(err.Error(), "cannot be used for API") {
		t.Fatalf("unexpected mismatch error: %v", err)
	}

	_, err = client.CallAPI("ListDomains", nil, nil, nil, nil, http.MethodPost)
	if err == nil || !strings.Contains(err.Error(), "does not support POST") {
		t.Fatalf("unexpected method error: %v", err)
	}
}

func decodeRequestBody(t *testing.T, r *http.Request, out interface{}) {
	t.Helper()
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(out); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
}
