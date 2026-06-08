package sdk

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestGeneratedClientPathAndRequestRouting(t *testing.T) {
	var gotPath string
	var gotQuery string
	var gotLang string

	sdk := newTestSDKWithTransport("https://api.test/api/v5", func(r *http.Request) (*http.Response, error) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		gotLang = r.Header.Get("X-Lang")
		return testResponse(http.StatusOK, `{"status":{"code":1,"message":"ok"},"data":{"total":0,"list":[]}}`, r), nil
	})
	client := NewEdgeNextClientFromSDK(sdk, WithLanguage("en"))
	resp, err := client.ListDomains(&ListDomainsRequest{Page: 1})
	if err != nil {
		t.Fatalf("ListDomains returned error: %v", err)
	}
	if gotPath != "/api/v5/domains" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotQuery == "" || !hasRawQueryValue(gotQuery, "page=1") {
		t.Fatalf("unexpected query: %s", gotQuery)
	}
	if gotLang != "en" {
		t.Fatalf("unexpected language header: %s", gotLang)
	}
	if resp.Status.Code != 1 {
		t.Fatalf("unexpected status: %+v", resp.Status)
	}
}

func TestGeneratedClientMovesGetBodyParamsToQuery(t *testing.T) {
	var gotQuery string

	sdk := newTestSDKWithTransport("https://api.test/api/v5", func(r *http.Request) (*http.Response, error) {
		gotQuery = r.URL.RawQuery
		return testResponse(http.StatusOK, `{"status":{"code":1,"message":"ok"},"data":{"total":"0","list":{}}}`, r), nil
	})
	req := &RuleListRequest{PackageId: 12}
	_, err := NewEdgeNextClientFromSDK(sdk).RuleList(req)
	if err != nil {
		t.Fatalf("RuleList returned error: %v", err)
	}
	if !hasRawQueryValue(gotQuery, "package_id=12") {
		t.Fatalf("unexpected query: %s", gotQuery)
	}
}

func TestSDKDemo(t *testing.T) {
	client := NewEdgeNextClient(EdgeNextClientConfig{
		AppId:     os.Getenv("EDGENEXT_APP_ID"),
		AppSecret: os.Getenv("EDGENEXT_APP_SECRET"),
		ApiPre:    os.Getenv("EDGENEXT_API_PRE"),
		UserId:    1,
		Timeout:   30,
	}, WithLanguage("en"))

	resp, err := client.ListDomains(&ListDomainsRequest{
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		fmt.Printf("request failed: %v\n", err)
		return
	}

	fmt.Printf("total domains: %v\n", resp.Data.Total)
	for _, domain := range resp.Data.List {
		fmt.Printf("domain: %s\n", domain.Domain)
	}
}

func TestGeneratedClientReturnsAPIErrorForBusinessError(t *testing.T) {
	client := NewEdgeNextClientFromSDK(&Sdk{ApiPre: "https://api.test/api/v5"})
	client.request = func(uri, method string, reqParams ReqParams) (*Response, error) {
		if uri != "domains" {
			t.Fatalf("unexpected api: %s", uri)
		}
		return &Response{
			RespData: map[string]interface{}{
				"status": map[string]interface{}{"code": 91003, "message": "invalid domain"},
				"data":   map[string]interface{}{},
			},
			BizCode: 91003,
			BizMsg:  "invalid domain",
			BizData: map[string]interface{}{},
		}, nil
	}

	resp, err := client.ListDomains(&ListDomainsRequest{Page: 1})
	if err == nil {
		t.Fatal("expected business error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T %v", err, err)
	}
	if apiErr.Code != 91003 || apiErr.Message != "invalid domain" {
		t.Fatalf("unexpected api error: %+v", apiErr)
	}
	if resp == nil || resp.Status.Code != 91003 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func hasRawQueryValue(rawQuery string, value string) bool {
	for _, part := range strings.Split(rawQuery, "&") {
		if part == value {
			return true
		}
	}
	return false
}
