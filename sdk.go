package sdk

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	v2 "github.com/edgenextapisdk/edgenext-go/core"
	"github.com/gogf/gf/v2/util/gconv"
)

var SDK_VERSION = "2.0.0"
var defaultTimeout = 30

// Sdk holds request configuration.
type Sdk struct {
	AppId        string
	AppSecret    string
	ApiPre       string
	UserId       int
	clientIp     string
	userAgent    string
	Timeout      int
	Debug        bool
	isSetDefault bool
	httpClient   *http.Client
}

type Response struct {
	Url        string
	Api        string
	Method     string
	Query      string
	Data       string
	ReqHeaders map[string]string
	Response   *http.Response
	RespBody   string
	HttpCode   int
	RespData   map[string]interface{}
	BizCode    int
	BizMsg     string
	BizData    interface{}
}

type ReqParams struct {
	Query   map[string]interface{}
	Data    map[string]interface{}
	Headers map[string]string
}

func (sdk *Sdk) payload(method string, reqParams *ReqParams) {
	issuedAt := int(time.Now().Unix())
	if method == "GET" {
		reqParams.Query["user_id"] = strconv.Itoa(sdk.UserId)
		reqParams.Query["client_ip"] = "" // Used by proxies to forward the external client IP; empty by default.
		reqParams.Query["client_userAgent"] = sdk.userAgent
		reqParams.Query["algorithm"] = "HMAC-SHA256"
		reqParams.Query["issued_at"] = issuedAt
	} else {
		reqParams.Data["user_id"] = strconv.Itoa(sdk.UserId)
		reqParams.Data["client_ip"] = "" // Used by proxies to forward the external client IP; empty by default.
		reqParams.Data["client_userAgent"] = sdk.userAgent
		reqParams.Data["algorithm"] = "HMAC-SHA256"
		reqParams.Data["issued_at"] = issuedAt
	}
	reqParams.Headers["X-Auth-App-Id"] = sdk.AppId
	reqParams.Headers["X-Auth-Sdk-Version"] = SDK_VERSION
	reqParams.Headers["Content-Type"] = "application/json; charset=utf-8"
	reqParams.Headers["User-Agent"] = sdk.userAgent
	return
}

func (sdk *Sdk) initDefault() bool {
	sdk.clientIp = ""
	sdk.userAgent = "Sdk " + SDK_VERSION + "; " + runtime.Version() + "; arch/" + runtime.GOARCH + "; os/" + runtime.GOOS
	if sdk.Timeout <= 0 {
		sdk.Timeout = defaultTimeout
	}
	return true
}

func (sdk *Sdk) urlEncode(reqParams ReqParams) string {
	queryMap := gconv.MapStrStr(reqParams.Query)
	var keys []string
	for k := range queryMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var querys []string
	for _, k := range keys {
		querys = append(querys, fmt.Sprintf("%s=%s", url.QueryEscape(k), url.QueryEscape(queryMap[k])))
	}
	return strings.Join(querys, "&")
}

// Request sends an API request.
func (sdk *Sdk) Request(uri, method string, reqParams ReqParams) (*Response, error) {
	// Initialize default values.
	sdk.initDefault()
	response := Response{
		Api: uri,
	}
	if reqParams.Query == nil {
		reqParams.Query = map[string]interface{}{}
	}
	if reqParams.Data == nil {
		reqParams.Data = map[string]interface{}{}
	}
	if reqParams.Headers == nil {
		reqParams.Headers = map[string]string{}
	}
	method = strings.ToUpper(method)
	response.Method = method
	reqUrl := strings.TrimRight(sdk.ApiPre, "/") + "/" + strings.TrimLeft(uri, "/")
	response.Url = reqUrl
	var err error
	sdk.payload(method, &reqParams)
	query := sdk.urlEncode(reqParams)
	response.Query = query
	if query != "" {
		reqUrl = reqUrl + "?" + query
	}
	jsonByte, err := json.Marshal(reqParams.Data)
	if err != nil {
		return &response, err
	}
	response.Data = string(jsonByte)
	body := strings.NewReader(string(jsonByte))
	req, err := http.NewRequest(method, reqUrl, body)
	if err != nil {
		return &response, err
	}
	// Sign the request with the v2 signer.
	singer := v2.Signer{
		AppId:     sdk.AppId,
		AppSecret: sdk.AppSecret,
	}
	response.ReqHeaders = reqParams.Headers
	for k, v := range reqParams.Headers {
		req.Header.Set(k, v)
	}
	err = singer.Sign(req)
	if err != nil {
		return nil, err
	}
	client := sdk.httpClient
	if client == nil {
		client = &http.Client{}
	}
	client.Timeout = time.Duration(sdk.Timeout) * time.Second
	resp, err := client.Do(req)
	if err != nil {
		return &response, err
	}
	response.Response = resp
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	response.HttpCode = resp.StatusCode
	if resp.StatusCode != 200 {
		response.BizCode = 0
		response.BizMsg = "response code is " + strconv.Itoa(resp.StatusCode)
		response.BizData = map[string]interface{}{}
		err = fmt.Errorf("response code is %s", strconv.Itoa(resp.StatusCode))
		return &response, err
	}
	rawByte, err := io.ReadAll(resp.Body)
	if err != nil {
		response.BizCode = 0
		response.BizMsg = "response body read error: " + err.Error()
		response.BizData = map[string]interface{}{}
		err = fmt.Errorf("response body read error: %s", err.Error())
		return &response, err
	}

	response.RespBody = string(rawByte)
	respData := map[string]interface{}{}
	err = json.Unmarshal(rawByte, &respData)
	if err != nil {
		response.BizCode = 0
		response.BizMsg = "json parse response body error: " + err.Error()
		response.BizData = map[string]interface{}{}
		err = fmt.Errorf("json parse response body error: %s", err.Error())
		return &response, err
	}
	response.RespData = respData
	if bizStatus, ok := respData["status"].(map[string]interface{}); ok {
		code, hasCode := bizStatus["code"]
		message, messageOK := bizStatus["message"].(string)
		if !hasCode || !messageOK {
			response.BizCode = 0
			response.BizMsg = "the json format of response body status is invalid"
			response.BizData = map[string]interface{}{}
			err = fmt.Errorf("the json format of response body status is invalid")
			return &response, err
		}
		response.BizCode = gconv.Int(code)
		response.BizMsg = message
		response.BizData = respData["data"]
	} else {
		response.BizCode = 0
		response.BizMsg = "the json format of response body has not status"
		response.BizData = map[string]interface{}{}
		err = fmt.Errorf("the json format of response body has not status")
	}
	return &response, err
}

// Get sends a GET request.
func (sdk *Sdk) Get(api string, reqParams ReqParams) (*Response, error) {
	return sdk.Request(api, "GET", reqParams)
}

// Post sends a POST request.
func (sdk *Sdk) Post(api string, reqParams ReqParams) (*Response, error) {
	return sdk.Request(api, "POST", reqParams)
}

// Put sends a PUT request.
func (sdk *Sdk) Put(api string, reqParams ReqParams) (*Response, error) {
	return sdk.Request(api, "PUT", reqParams)
}

// Delete sends a DELETE request.
func (sdk *Sdk) Delete(api string, reqParams ReqParams) (*Response, error) {
	return sdk.Request(api, "DELETE", reqParams)
}
