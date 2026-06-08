package sdk

// Code generated from apidoc metadata. DO NOT EDIT.

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

var apiDefinitionsByAPIName = map[string]APIDefinition{}
var apiDefinitionsByMethodName = map[string]APIDefinition{}

func init() {
	for _, definition := range APIDefinitions {
		apiDefinitionsByAPIName[definition.APIName] = definition
		apiDefinitionsByMethodName[definition.MethodName] = definition
	}
}

// ClientOption customizes an EdgeNextClient during construction.
type ClientOption func(*EdgeNextClient)

// WithDefaultHeaders configures headers that are sent with every request.
func WithDefaultHeaders(headers map[string]string) ClientOption {
	return func(c *EdgeNextClient) {
		for key, value := range headers {
			c.defaultHeaders[key] = value
		}
	}
}

// WithLanguage sets the X-Lang header for every request.
func WithLanguage(lang string) ClientOption {
	return WithDefaultHeaders(map[string]string{"X-Lang": lang})
}

type RequestParts interface {
	APIName() string
	Method() string
	RequestParts() ReqParams
}

type EdgeNextClient struct {
	SDK            *Sdk
	defaultHeaders map[string]string
	request        func(uri, method string, reqParams ReqParams) (*Response, error)
}

type APIResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type APIResponseStatus struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type APIError struct {
	Code    int
	Message string
}

func (err *APIError) Error() string {
	if err.Message == "" {
		return fmt.Sprintf("api error: code %d", err.Code)
	}
	return fmt.Sprintf("api error: %s (code: %d)", err.Message, err.Code)
}

// EdgeNextClientConfig contains the credentials and endpoint used to initialize the SDK client.
type EdgeNextClientConfig struct {
	AppId     string // Application ID provided by EdgeNext.
	AppSecret string // Application secret used for request signing.
	ApiPre    string // API endpoint prefix, for example https://example.com/api/v5.
	UserId    int    // Optional user ID included in SDK payload fields.
	Timeout   int    // Request timeout in seconds. Defaults to 30 when unset.
	Debug     bool   // Optional debug flag reserved for SDK diagnostics.
}

// SDKConfig converts client configuration into the lower-level SDK configuration.
func (config EdgeNextClientConfig) SDKConfig() *Sdk {
	return &Sdk{
		AppId:     config.AppId,
		AppSecret: config.AppSecret,
		ApiPre:    config.ApiPre,
		UserId:    config.UserId,
		Timeout:   config.Timeout,
		Debug:     config.Debug,
	}
}

// NewEdgeNextClient initializes a client from EdgeNext credentials and endpoint configuration.
func NewEdgeNextClient(config EdgeNextClientConfig, opts ...ClientOption) *EdgeNextClient {
	return NewEdgeNextClientFromSDK(config.SDKConfig(), opts...)
}

// NewEdgeNextClientFromSDK initializes a client from an existing lower-level SDK instance.
func NewEdgeNextClientFromSDK(sdk *Sdk, opts ...ClientOption) *EdgeNextClient {
	if sdk == nil {
		sdk = &Sdk{}
	}
	client := &EdgeNextClient{SDK: sdk, defaultHeaders: map[string]string{}}
	client.request = sdk.Request
	for _, opt := range opts {
		opt(client)
	}
	return client
}

// SetDefaultHeader sets or replaces a default header sent with every request.
func (c *EdgeNextClient) SetDefaultHeader(key string, value string) *EdgeNextClient {
	c.defaultHeaders[key] = value
	return c
}

// Request sends a low-level SDK request through the wrapped SDK instance.
func (c *EdgeNextClient) Request(uri, method string, reqParams ReqParams) (*APIResponse, error) {
	return c.callSDK(uri, method, c.withDefaultHeaders(reqParams))
}

// Get sends a low-level GET request through the wrapped SDK instance.
func (c *EdgeNextClient) Get(api string, reqParams ReqParams) (*APIResponse, error) {
	return c.Request(api, "GET", reqParams)
}

// Post sends a low-level POST request through the wrapped SDK instance.
func (c *EdgeNextClient) Post(api string, reqParams ReqParams) (*APIResponse, error) {
	return c.Request(api, "POST", reqParams)
}

// Put sends a low-level PUT request through the wrapped SDK instance.
func (c *EdgeNextClient) Put(api string, reqParams ReqParams) (*APIResponse, error) {
	return c.Request(api, "PUT", reqParams)
}

// Delete sends a low-level DELETE request through the wrapped SDK instance.
func (c *EdgeNextClient) Delete(api string, reqParams ReqParams) (*APIResponse, error) {
	return c.Request(api, "DELETE", reqParams)
}

func (c *EdgeNextClient) withDefaultHeaders(reqParams ReqParams) ReqParams {
	reqParams.Headers = mergeStringMap(c.defaultHeaders, reqParams.Headers)
	return reqParams
}

func (c *EdgeNextClient) GetAPIDefinition(apiName string) (APIDefinition, bool) {
	definition, ok := apiDefinitionsByAPIName[apiName]
	if ok {
		return definition, true
	}
	definition, ok = apiDefinitionsByMethodName[apiName]
	return definition, ok
}

func (c *EdgeNextClient) CallAPI(apiName string, request RequestParts, query map[string]interface{}, data map[string]interface{}, headers map[string]string, method string) (*APIResponse, error) {
	definition, ok := c.GetAPIDefinition(apiName)
	if !ok {
		return nil, fmt.Errorf("unknown EdgeNext API: %s", apiName)
	}
	return c.callDefinition(definition, request, query, data, headers, method)
}

func (c *EdgeNextClient) callDefinition(definition APIDefinition, request RequestParts, query map[string]interface{}, data map[string]interface{}, headers map[string]string, method string) (*APIResponse, error) {
	selected := strings.ToUpper(method)
	if selected == "" && len(definition.Methods) > 0 {
		selected = strings.ToUpper(definition.Methods[0])
	}
	reqParams := ReqParams{
		Query:   cloneInterfaceMap(query),
		Data:    cloneInterfaceMap(data),
		Headers: mergeStringMap(c.defaultHeaders, headers),
	}
	if request != nil {
		if request.APIName() != "" && request.APIName() != definition.APIName {
			return nil, fmt.Errorf("request %s cannot be used for API %s", request.APIName(), definition.APIName)
		}
		parts := request.RequestParts()
		reqParams.Query = mergeInterfaceMap(parts.Query, reqParams.Query)
		reqParams.Data = mergeInterfaceMap(parts.Data, reqParams.Data)
		reqParams.Headers = mergeStringMap(reqParams.Headers, parts.Headers)
		if method == "" && request.Method() != "" {
			selected = strings.ToUpper(request.Method())
		}
	}
	if !definitionSupportsMethod(definition, selected) {
		return nil, fmt.Errorf("%s does not support %s", definition.APIName, selected)
	}
	if selected == "GET" && len(reqParams.Data) > 0 {
		reqParams.Query = mergeInterfaceMap(reqParams.Data, reqParams.Query)
		reqParams.Data = map[string]interface{}{}
	}
	api := apiPathToSDKAPI(definition.Path, c.SDK.ApiPre)
	return c.callSDK(api, selected, reqParams)
}

func (c *EdgeNextClient) callSDK(api string, method string, reqParams ReqParams) (*APIResponse, error) {
	request := c.request
	if request == nil {
		request = c.SDK.Request
	}
	raw, err := request(api, method, reqParams)
	resp := newAPIResponse(raw)
	if err != nil {
		if resp == nil {
			return nil, err
		}
		return resp, err
	}
	if raw != nil && raw.BizCode != 1 {
		return resp, &APIError{Code: raw.BizCode, Message: raw.BizMsg}
	}
	return resp, nil
}

func newAPIResponse(raw *Response) *APIResponse {
	if raw == nil {
		return nil
	}
	if raw.RespData == nil {
		return nil
	}
	return &APIResponse{
		Status: APIResponseStatus{
			Code:    raw.BizCode,
			Message: raw.BizMsg,
		},
		Data: raw.BizData,
	}
}

func (c *EdgeNextClient) callTypedDefinition(definition APIDefinition, request RequestParts, query map[string]interface{}, data map[string]interface{}, headers map[string]string, method string, out interface{}) (bool, error) {
	response, err := c.callDefinition(definition, request, query, data, headers, method)
	if response == nil {
		return false, err
	}
	encoded, marshalErr := json.Marshal(response)
	if marshalErr != nil {
		return true, marshalErr
	}
	if unmarshalErr := json.Unmarshal(encoded, out); unmarshalErr != nil {
		return true, unmarshalErr
	}
	return true, err
}

func definitionSupportsMethod(definition APIDefinition, method string) bool {
	for _, candidate := range definition.Methods {
		if strings.ToUpper(candidate) == strings.ToUpper(method) {
			return true
		}
	}
	return false
}

func apiPathToSDKAPI(path string, apiPre string) string {
	route := strings.TrimLeft(strings.TrimSpace(path), "/")
	parsed, err := url.Parse(apiPre)
	apiPrePath := ""
	if err == nil {
		apiPrePath = parsed.Path
	}
	apiPrePath = strings.Trim(strings.ToLower(apiPrePath), "/")
	parts := strings.Split(apiPrePath, "/")
	if len(parts) >= 2 && parts[len(parts)-2] == "api" && isVersionSegment(parts[len(parts)-1]) {
		prefix := "api/" + parts[len(parts)-1] + "/"
		if strings.HasPrefix(strings.ToLower(route), prefix) {
			return route[len(prefix):]
		}
	}
	if len(parts) >= 1 && parts[len(parts)-1] == "api" && strings.HasPrefix(strings.ToLower(route), "api/") {
		return route[len("api/"):]
	}
	return route
}

func isVersionSegment(segment string) bool {
	if len(segment) < 2 || segment[0] != 'v' {
		return false
	}
	for _, ch := range segment[1:] {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func cloneInterfaceMap(in map[string]interface{}) map[string]interface{} {
	out := map[string]interface{}{}
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneStringMap(in map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range in {
		out[k] = v
	}
	return out
}

func mergeInterfaceMap(base map[string]interface{}, override map[string]interface{}) map[string]interface{} {
	out := cloneInterfaceMap(base)
	for k, v := range override {
		out[k] = v
	}
	return out
}

func mergeStringMap(base map[string]string, override map[string]string) map[string]string {
	out := cloneStringMap(base)
	for k, v := range override {
		out[k] = v
	}
	return out
}

type CdnHighDefenseIpGetArticleIpRequest struct {
	Headers map[string]string
}

func NewCdnHighDefenseIpGetArticleIpRequest() *CdnHighDefenseIpGetArticleIpRequest {
	return &CdnHighDefenseIpGetArticleIpRequest{Headers: map[string]string{}}
}

func (r *CdnHighDefenseIpGetArticleIpRequest) APIName() string {
	return "CdnHighDefenseIP_getArticleIP"
}
func (r *CdnHighDefenseIpGetArticleIpRequest) Method() string { return "GET" }

func (r *CdnHighDefenseIpGetArticleIpRequest) SetHeader(key string, value string) *CdnHighDefenseIpGetArticleIpRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CdnHighDefenseIpGetArticleIpRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CdnHighDefenseIpGetArticleIpRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type DnsDomainGetDomainListRequest struct {
	Headers map[string]string
}

func NewDnsDomainGetDomainListRequest() *DnsDomainGetDomainListRequest {
	return &DnsDomainGetDomainListRequest{Headers: map[string]string{}}
}

func (r *DnsDomainGetDomainListRequest) APIName() string { return "DnsDomain_getDomainList" }
func (r *DnsDomainGetDomainListRequest) Method() string  { return "GET" }

func (r *DnsDomainGetDomainListRequest) SetHeader(key string, value string) *DnsDomainGetDomainListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DnsDomainGetDomainListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DnsDomainGetDomainListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type DnsDomainAddDomainRequest struct {
	Headers map[string]string
	Domain  interface{} `json:"domain,omitempty"`
}

func NewDnsDomainAddDomainRequest() *DnsDomainAddDomainRequest {
	return &DnsDomainAddDomainRequest{Headers: map[string]string{}}
}

func (r *DnsDomainAddDomainRequest) APIName() string { return "DnsDomain_addDomain" }
func (r *DnsDomainAddDomainRequest) Method() string  { return "POST" }

func (r *DnsDomainAddDomainRequest) SetHeader(key string, value string) *DnsDomainAddDomainRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DnsDomainAddDomainRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DnsDomainAddDomainRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Domain != nil {
		parts.Data["domain"] = r.Domain
	}
	return parts
}

type DnsDomainBatchAddDomainsRequest struct {
	Headers     map[string]string
	Domains     interface{} `json:"domains,omitempty"`
	AddRecord   interface{} `json:"add_record,omitempty"`
	RecordValue interface{} `json:"record_value,omitempty"`
	GroupId     interface{} `json:"group_id,omitempty"`
}

func NewDnsDomainBatchAddDomainsRequest() *DnsDomainBatchAddDomainsRequest {
	return &DnsDomainBatchAddDomainsRequest{Headers: map[string]string{}}
}

func (r *DnsDomainBatchAddDomainsRequest) APIName() string { return "DnsDomain_batchAddDomains" }
func (r *DnsDomainBatchAddDomainsRequest) Method() string  { return "POST" }

func (r *DnsDomainBatchAddDomainsRequest) SetHeader(key string, value string) *DnsDomainBatchAddDomainsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DnsDomainBatchAddDomainsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DnsDomainBatchAddDomainsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Domains != nil {
		parts.Data["domains"] = r.Domains
	}
	parts.Data["add_record"] = "0"
	if r.AddRecord != nil {
		parts.Data["add_record"] = r.AddRecord
	}
	if r.RecordValue != nil {
		parts.Data["record_value"] = r.RecordValue
	}
	if r.GroupId != nil {
		parts.Data["group_id"] = r.GroupId
	}
	return parts
}

type DnsDomainBatchDeleteDomainsRequest struct {
	Headers   map[string]string
	DomainIds interface{} `json:"domain_ids,omitempty"`
}

func NewDnsDomainBatchDeleteDomainsRequest() *DnsDomainBatchDeleteDomainsRequest {
	return &DnsDomainBatchDeleteDomainsRequest{Headers: map[string]string{}}
}

func (r *DnsDomainBatchDeleteDomainsRequest) APIName() string { return "DnsDomain_batchDeleteDomains" }
func (r *DnsDomainBatchDeleteDomainsRequest) Method() string  { return "DELETE" }

func (r *DnsDomainBatchDeleteDomainsRequest) SetHeader(key string, value string) *DnsDomainBatchDeleteDomainsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DnsDomainBatchDeleteDomainsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DnsDomainBatchDeleteDomainsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.DomainIds != nil {
		parts.Data["domain_ids"] = r.DomainIds
	}
	return parts
}

type DnsDomainGetDomainStatRequest struct {
	Headers map[string]string
}

func NewDnsDomainGetDomainStatRequest() *DnsDomainGetDomainStatRequest {
	return &DnsDomainGetDomainStatRequest{Headers: map[string]string{}}
}

func (r *DnsDomainGetDomainStatRequest) APIName() string { return "DnsDomain_getDomainStat" }
func (r *DnsDomainGetDomainStatRequest) Method() string  { return "GET" }

func (r *DnsDomainGetDomainStatRequest) SetHeader(key string, value string) *DnsDomainGetDomainStatRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DnsDomainGetDomainStatRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DnsDomainGetDomainStatRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type DnsDomainGetDomainServersRequest struct {
	Headers map[string]string
}

func NewDnsDomainGetDomainServersRequest() *DnsDomainGetDomainServersRequest {
	return &DnsDomainGetDomainServersRequest{Headers: map[string]string{}}
}

func (r *DnsDomainGetDomainServersRequest) APIName() string { return "DnsDomain_getDomainServers" }
func (r *DnsDomainGetDomainServersRequest) Method() string  { return "GET" }

func (r *DnsDomainGetDomainServersRequest) SetHeader(key string, value string) *DnsDomainGetDomainServersRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DnsDomainGetDomainServersRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DnsDomainGetDomainServersRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type DnsDomainGetTasksListRequest struct {
	Headers map[string]string
}

func NewDnsDomainGetTasksListRequest() *DnsDomainGetTasksListRequest {
	return &DnsDomainGetTasksListRequest{Headers: map[string]string{}}
}

func (r *DnsDomainGetTasksListRequest) APIName() string { return "DnsDomain_getTasksList" }
func (r *DnsDomainGetTasksListRequest) Method() string  { return "GET" }

func (r *DnsDomainGetTasksListRequest) SetHeader(key string, value string) *DnsDomainGetTasksListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DnsDomainGetTasksListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DnsDomainGetTasksListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type DnsDomainGetTaskDetailRequest struct {
	Headers map[string]string
}

func NewDnsDomainGetTaskDetailRequest() *DnsDomainGetTaskDetailRequest {
	return &DnsDomainGetTaskDetailRequest{Headers: map[string]string{}}
}

func (r *DnsDomainGetTaskDetailRequest) APIName() string { return "DnsDomain_getTaskDetail" }
func (r *DnsDomainGetTaskDetailRequest) Method() string  { return "GET" }

func (r *DnsDomainGetTaskDetailRequest) SetHeader(key string, value string) *DnsDomainGetTaskDetailRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DnsDomainGetTaskDetailRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DnsDomainGetTaskDetailRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type CloudDnsDomainGroupGetGroupListRequest struct {
	Headers map[string]string
}

func NewCloudDnsDomainGroupGetGroupListRequest() *CloudDnsDomainGroupGetGroupListRequest {
	return &CloudDnsDomainGroupGetGroupListRequest{Headers: map[string]string{}}
}

func (r *CloudDnsDomainGroupGetGroupListRequest) APIName() string {
	return "CloudDns_DomainGroup_getGroupList"
}
func (r *CloudDnsDomainGroupGetGroupListRequest) Method() string { return "GET" }

func (r *CloudDnsDomainGroupGetGroupListRequest) SetHeader(key string, value string) *CloudDnsDomainGroupGetGroupListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CloudDnsDomainGroupGetGroupListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CloudDnsDomainGroupGetGroupListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type CloudDnsDomainGroupAddGroupRequest struct {
	Headers   map[string]string
	GroupName interface{} `json:"group_name,omitempty"`
	Remark    interface{} `json:"remark,omitempty"`
}

func NewCloudDnsDomainGroupAddGroupRequest() *CloudDnsDomainGroupAddGroupRequest {
	return &CloudDnsDomainGroupAddGroupRequest{Headers: map[string]string{}}
}

func (r *CloudDnsDomainGroupAddGroupRequest) APIName() string { return "CloudDns_DomainGroup_addGroup" }
func (r *CloudDnsDomainGroupAddGroupRequest) Method() string  { return "POST" }

func (r *CloudDnsDomainGroupAddGroupRequest) SetHeader(key string, value string) *CloudDnsDomainGroupAddGroupRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CloudDnsDomainGroupAddGroupRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CloudDnsDomainGroupAddGroupRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.GroupName != nil {
		parts.Data["group_name"] = r.GroupName
	}
	if r.Remark != nil {
		parts.Data["remark"] = r.Remark
	}
	return parts
}

type CloudDnsDomainGroupUpdateGroupRequest struct {
	Headers   map[string]string
	GroupId   interface{} `json:"group_id,omitempty"`
	GroupName interface{} `json:"group_name,omitempty"`
	Remark    interface{} `json:"remark,omitempty"`
	DomainIds interface{} `json:"domain_ids,omitempty"`
}

func NewCloudDnsDomainGroupUpdateGroupRequest() *CloudDnsDomainGroupUpdateGroupRequest {
	return &CloudDnsDomainGroupUpdateGroupRequest{Headers: map[string]string{}}
}

func (r *CloudDnsDomainGroupUpdateGroupRequest) APIName() string {
	return "CloudDns_DomainGroup_updateGroup"
}
func (r *CloudDnsDomainGroupUpdateGroupRequest) Method() string { return "PUT" }

func (r *CloudDnsDomainGroupUpdateGroupRequest) SetHeader(key string, value string) *CloudDnsDomainGroupUpdateGroupRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CloudDnsDomainGroupUpdateGroupRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CloudDnsDomainGroupUpdateGroupRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.GroupId != nil {
		parts.Data["group_id"] = r.GroupId
	}
	if r.GroupName != nil {
		parts.Data["group_name"] = r.GroupName
	}
	if r.Remark != nil {
		parts.Data["remark"] = r.Remark
	}
	if r.DomainIds != nil {
		parts.Data["domain_ids"] = r.DomainIds
	}
	return parts
}

type CloudDnsDomainGroupDeleteGroupRequest struct {
	Headers map[string]string
	GroupId interface{} `json:"group_id,omitempty"`
}

func NewCloudDnsDomainGroupDeleteGroupRequest() *CloudDnsDomainGroupDeleteGroupRequest {
	return &CloudDnsDomainGroupDeleteGroupRequest{Headers: map[string]string{}}
}

func (r *CloudDnsDomainGroupDeleteGroupRequest) APIName() string {
	return "CloudDns_DomainGroup_deleteGroup"
}
func (r *CloudDnsDomainGroupDeleteGroupRequest) Method() string { return "DELETE" }

func (r *CloudDnsDomainGroupDeleteGroupRequest) SetHeader(key string, value string) *CloudDnsDomainGroupDeleteGroupRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CloudDnsDomainGroupDeleteGroupRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CloudDnsDomainGroupDeleteGroupRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.GroupId != nil {
		parts.Data["group_id"] = r.GroupId
	}
	return parts
}

type CloudDnsDomainGroupGetGroupRecordListRequest struct {
	Headers map[string]string
}

func NewCloudDnsDomainGroupGetGroupRecordListRequest() *CloudDnsDomainGroupGetGroupRecordListRequest {
	return &CloudDnsDomainGroupGetGroupRecordListRequest{Headers: map[string]string{}}
}

func (r *CloudDnsDomainGroupGetGroupRecordListRequest) APIName() string {
	return "CloudDns_DomainGroup_getGroupRecordList"
}
func (r *CloudDnsDomainGroupGetGroupRecordListRequest) Method() string { return "GET" }

func (r *CloudDnsDomainGroupGetGroupRecordListRequest) SetHeader(key string, value string) *CloudDnsDomainGroupGetGroupRecordListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CloudDnsDomainGroupGetGroupRecordListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CloudDnsDomainGroupGetGroupRecordListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type CloudDnsDomainGroupSaveDomainToGroupRequest struct {
	Headers   map[string]string
	GroupId   interface{} `json:"group_id,omitempty"`
	DomainIds interface{} `json:"domain_ids,omitempty"`
	Action    interface{} `json:"action,omitempty"`
}

func NewCloudDnsDomainGroupSaveDomainToGroupRequest() *CloudDnsDomainGroupSaveDomainToGroupRequest {
	return &CloudDnsDomainGroupSaveDomainToGroupRequest{Headers: map[string]string{}}
}

func (r *CloudDnsDomainGroupSaveDomainToGroupRequest) APIName() string {
	return "CloudDns_DomainGroup_saveDomainToGroup"
}
func (r *CloudDnsDomainGroupSaveDomainToGroupRequest) Method() string { return "POST" }

func (r *CloudDnsDomainGroupSaveDomainToGroupRequest) SetHeader(key string, value string) *CloudDnsDomainGroupSaveDomainToGroupRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CloudDnsDomainGroupSaveDomainToGroupRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CloudDnsDomainGroupSaveDomainToGroupRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.GroupId != nil {
		parts.Data["group_id"] = r.GroupId
	}
	if r.DomainIds != nil {
		parts.Data["domain_ids"] = r.DomainIds
	}
	if r.Action != nil {
		parts.Data["action"] = r.Action
	}
	return parts
}

type CloudDnsDomainGroupGetGroupDomainListRequest struct {
	Headers map[string]string
}

func NewCloudDnsDomainGroupGetGroupDomainListRequest() *CloudDnsDomainGroupGetGroupDomainListRequest {
	return &CloudDnsDomainGroupGetGroupDomainListRequest{Headers: map[string]string{}}
}

func (r *CloudDnsDomainGroupGetGroupDomainListRequest) APIName() string {
	return "CloudDns_DomainGroup_getGroupDomainList"
}
func (r *CloudDnsDomainGroupGetGroupDomainListRequest) Method() string { return "POST" }

func (r *CloudDnsDomainGroupGetGroupDomainListRequest) SetHeader(key string, value string) *CloudDnsDomainGroupGetGroupDomainListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CloudDnsDomainGroupGetGroupDomainListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CloudDnsDomainGroupGetGroupDomainListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type CloudDnsDomainGroupGetGroupUndistributedDomainListRequest struct {
	Headers map[string]string
	GroupId interface{} `json:"group_id,omitempty"`
	Domain  interface{} `json:"domain,omitempty"`
}

func NewCloudDnsDomainGroupGetGroupUndistributedDomainListRequest() *CloudDnsDomainGroupGetGroupUndistributedDomainListRequest {
	return &CloudDnsDomainGroupGetGroupUndistributedDomainListRequest{Headers: map[string]string{}}
}

func (r *CloudDnsDomainGroupGetGroupUndistributedDomainListRequest) APIName() string {
	return "CloudDns_DomainGroup_getGroupUndistributedDomainList"
}
func (r *CloudDnsDomainGroupGetGroupUndistributedDomainListRequest) Method() string { return "POST" }

func (r *CloudDnsDomainGroupGetGroupUndistributedDomainListRequest) SetHeader(key string, value string) *CloudDnsDomainGroupGetGroupUndistributedDomainListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CloudDnsDomainGroupGetGroupUndistributedDomainListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CloudDnsDomainGroupGetGroupUndistributedDomainListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.GroupId != nil {
		parts.Data["group_id"] = r.GroupId
	}
	if r.Domain != nil {
		parts.Data["domain"] = r.Domain
	}
	return parts
}

type DnsDomainRecordsGetRecordTypesRequest struct {
	Headers map[string]string
}

func NewDnsDomainRecordsGetRecordTypesRequest() *DnsDomainRecordsGetRecordTypesRequest {
	return &DnsDomainRecordsGetRecordTypesRequest{Headers: map[string]string{}}
}

func (r *DnsDomainRecordsGetRecordTypesRequest) APIName() string {
	return "DnsDomainRecords_getRecordTypes"
}
func (r *DnsDomainRecordsGetRecordTypesRequest) Method() string { return "GET" }

func (r *DnsDomainRecordsGetRecordTypesRequest) SetHeader(key string, value string) *DnsDomainRecordsGetRecordTypesRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DnsDomainRecordsGetRecordTypesRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DnsDomainRecordsGetRecordTypesRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type DnsDomainRecordsGetRecordListRequest struct {
	Headers map[string]string
}

func NewDnsDomainRecordsGetRecordListRequest() *DnsDomainRecordsGetRecordListRequest {
	return &DnsDomainRecordsGetRecordListRequest{Headers: map[string]string{}}
}

func (r *DnsDomainRecordsGetRecordListRequest) APIName() string {
	return "DnsDomainRecords_getRecordList"
}
func (r *DnsDomainRecordsGetRecordListRequest) Method() string { return "GET" }

func (r *DnsDomainRecordsGetRecordListRequest) SetHeader(key string, value string) *DnsDomainRecordsGetRecordListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DnsDomainRecordsGetRecordListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DnsDomainRecordsGetRecordListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type DnsDomainRecordsAddRecordRequest struct {
	Headers      map[string]string
	DomainId     interface{} `json:"domain_id,omitempty"`
	RecordName   interface{} `json:"record_name,omitempty"`
	RecordType   interface{} `json:"record_type,omitempty"`
	RecordView   interface{} `json:"record_view,omitempty"`
	RecordValue  interface{} `json:"record_value,omitempty"`
	RecordMx     interface{} `json:"record_mx,omitempty"`
	RecordTtl    interface{} `json:"record_ttl,omitempty"`
	RecordRemark interface{} `json:"record_remark,omitempty"`
}

func NewDnsDomainRecordsAddRecordRequest() *DnsDomainRecordsAddRecordRequest {
	return &DnsDomainRecordsAddRecordRequest{Headers: map[string]string{}}
}

func (r *DnsDomainRecordsAddRecordRequest) APIName() string { return "DnsDomainRecords_addRecord" }
func (r *DnsDomainRecordsAddRecordRequest) Method() string  { return "POST" }

func (r *DnsDomainRecordsAddRecordRequest) SetHeader(key string, value string) *DnsDomainRecordsAddRecordRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DnsDomainRecordsAddRecordRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DnsDomainRecordsAddRecordRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.DomainId != nil {
		parts.Data["domain_id"] = r.DomainId
	}
	if r.RecordName != nil {
		parts.Data["record_name"] = r.RecordName
	}
	if r.RecordType != nil {
		parts.Data["record_type"] = r.RecordType
	}
	if r.RecordView != nil {
		parts.Data["record_view"] = r.RecordView
	}
	if r.RecordValue != nil {
		parts.Data["record_value"] = r.RecordValue
	}
	parts.Data["record_mx"] = "0"
	if r.RecordMx != nil {
		parts.Data["record_mx"] = r.RecordMx
	}
	parts.Data["record_ttl"] = "600"
	if r.RecordTtl != nil {
		parts.Data["record_ttl"] = r.RecordTtl
	}
	if r.RecordRemark != nil {
		parts.Data["record_remark"] = r.RecordRemark
	}
	return parts
}

type DnsDomainRecordsBatchAddRecordsRequest struct {
	Headers   map[string]string
	DomainIds interface{} `json:"domain_ids,omitempty"`
	Records   interface{} `json:"records,omitempty"`
}

func NewDnsDomainRecordsBatchAddRecordsRequest() *DnsDomainRecordsBatchAddRecordsRequest {
	return &DnsDomainRecordsBatchAddRecordsRequest{Headers: map[string]string{}}
}

func (r *DnsDomainRecordsBatchAddRecordsRequest) APIName() string {
	return "DnsDomainRecords_batchAddRecords"
}
func (r *DnsDomainRecordsBatchAddRecordsRequest) Method() string { return "POST" }

func (r *DnsDomainRecordsBatchAddRecordsRequest) SetHeader(key string, value string) *DnsDomainRecordsBatchAddRecordsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DnsDomainRecordsBatchAddRecordsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DnsDomainRecordsBatchAddRecordsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.DomainIds != nil {
		parts.Data["domain_ids"] = r.DomainIds
	}
	if r.Records != nil {
		parts.Data["records"] = r.Records
	}
	return parts
}

type DnsDomainRecordsEditRecordRequest struct {
	Headers      map[string]string
	RecordId     interface{} `json:"record_id,omitempty"`
	DomainId     interface{} `json:"domain_id,omitempty"`
	RecordName   interface{} `json:"record_name,omitempty"`
	RecordType   interface{} `json:"record_type,omitempty"`
	RecordView   interface{} `json:"record_view,omitempty"`
	RecordValue  interface{} `json:"record_value,omitempty"`
	RecordMx     interface{} `json:"record_mx,omitempty"`
	RecordTtl    interface{} `json:"record_ttl,omitempty"`
	RecordRemark interface{} `json:"record_remark,omitempty"`
}

func NewDnsDomainRecordsEditRecordRequest() *DnsDomainRecordsEditRecordRequest {
	return &DnsDomainRecordsEditRecordRequest{Headers: map[string]string{}}
}

func (r *DnsDomainRecordsEditRecordRequest) APIName() string { return "DnsDomainRecords_editRecord" }
func (r *DnsDomainRecordsEditRecordRequest) Method() string  { return "PUT" }

func (r *DnsDomainRecordsEditRecordRequest) SetHeader(key string, value string) *DnsDomainRecordsEditRecordRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DnsDomainRecordsEditRecordRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DnsDomainRecordsEditRecordRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.RecordId != nil {
		parts.Data["record_id"] = r.RecordId
	}
	if r.DomainId != nil {
		parts.Data["domain_id"] = r.DomainId
	}
	if r.RecordName != nil {
		parts.Data["record_name"] = r.RecordName
	}
	if r.RecordType != nil {
		parts.Data["record_type"] = r.RecordType
	}
	if r.RecordView != nil {
		parts.Data["record_view"] = r.RecordView
	}
	if r.RecordValue != nil {
		parts.Data["record_value"] = r.RecordValue
	}
	parts.Data["record_mx"] = "0"
	if r.RecordMx != nil {
		parts.Data["record_mx"] = r.RecordMx
	}
	parts.Data["record_ttl"] = "600"
	if r.RecordTtl != nil {
		parts.Data["record_ttl"] = r.RecordTtl
	}
	if r.RecordRemark != nil {
		parts.Data["record_remark"] = r.RecordRemark
	}
	return parts
}

type DnsDomainRecordsBatchPauseRecordsRequest struct {
	Headers   map[string]string
	DomainId  interface{} `json:"domain_id,omitempty"`
	RecordIds interface{} `json:"record_ids,omitempty"`
}

func NewDnsDomainRecordsBatchPauseRecordsRequest() *DnsDomainRecordsBatchPauseRecordsRequest {
	return &DnsDomainRecordsBatchPauseRecordsRequest{Headers: map[string]string{}}
}

func (r *DnsDomainRecordsBatchPauseRecordsRequest) APIName() string {
	return "DnsDomainRecords_batchPauseRecords"
}
func (r *DnsDomainRecordsBatchPauseRecordsRequest) Method() string { return "POST" }

func (r *DnsDomainRecordsBatchPauseRecordsRequest) SetHeader(key string, value string) *DnsDomainRecordsBatchPauseRecordsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DnsDomainRecordsBatchPauseRecordsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DnsDomainRecordsBatchPauseRecordsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.DomainId != nil {
		parts.Data["domain_id"] = r.DomainId
	}
	if r.RecordIds != nil {
		parts.Data["record_ids"] = r.RecordIds
	}
	return parts
}

type DnsDomainRecordsBatchEnableRecordsRequest struct {
	Headers   map[string]string
	DomainId  interface{} `json:"domain_id,omitempty"`
	RecordIds interface{} `json:"record_ids,omitempty"`
}

func NewDnsDomainRecordsBatchEnableRecordsRequest() *DnsDomainRecordsBatchEnableRecordsRequest {
	return &DnsDomainRecordsBatchEnableRecordsRequest{Headers: map[string]string{}}
}

func (r *DnsDomainRecordsBatchEnableRecordsRequest) APIName() string {
	return "DnsDomainRecords_batchEnableRecords"
}
func (r *DnsDomainRecordsBatchEnableRecordsRequest) Method() string { return "POST" }

func (r *DnsDomainRecordsBatchEnableRecordsRequest) SetHeader(key string, value string) *DnsDomainRecordsBatchEnableRecordsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DnsDomainRecordsBatchEnableRecordsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DnsDomainRecordsBatchEnableRecordsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.DomainId != nil {
		parts.Data["domain_id"] = r.DomainId
	}
	if r.RecordIds != nil {
		parts.Data["record_ids"] = r.RecordIds
	}
	return parts
}

type DnsDomainRecordsDeleteRecordRequest struct {
	Headers  map[string]string
	RecordId interface{} `json:"record_id,omitempty"`
	DomainId interface{} `json:"domain_id,omitempty"`
}

func NewDnsDomainRecordsDeleteRecordRequest() *DnsDomainRecordsDeleteRecordRequest {
	return &DnsDomainRecordsDeleteRecordRequest{Headers: map[string]string{}}
}

func (r *DnsDomainRecordsDeleteRecordRequest) APIName() string {
	return "DnsDomainRecords_deleteRecord"
}
func (r *DnsDomainRecordsDeleteRecordRequest) Method() string { return "DELETE" }

func (r *DnsDomainRecordsDeleteRecordRequest) SetHeader(key string, value string) *DnsDomainRecordsDeleteRecordRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DnsDomainRecordsDeleteRecordRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DnsDomainRecordsDeleteRecordRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.RecordId != nil {
		parts.Data["record_id"] = r.RecordId
	}
	if r.DomainId != nil {
		parts.Data["domain_id"] = r.DomainId
	}
	return parts
}

type DnsDomainRecordsImportRecordsRequest struct {
	Headers map[string]string
	XlsFile interface{} `json:"xls_file,omitempty"`
}

func NewDnsDomainRecordsImportRecordsRequest() *DnsDomainRecordsImportRecordsRequest {
	return &DnsDomainRecordsImportRecordsRequest{Headers: map[string]string{}}
}

func (r *DnsDomainRecordsImportRecordsRequest) APIName() string {
	return "DnsDomainRecords_importRecords"
}
func (r *DnsDomainRecordsImportRecordsRequest) Method() string { return "POST" }

func (r *DnsDomainRecordsImportRecordsRequest) SetHeader(key string, value string) *DnsDomainRecordsImportRecordsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DnsDomainRecordsImportRecordsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DnsDomainRecordsImportRecordsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.XlsFile != nil {
		parts.Data["xls_file"] = r.XlsFile
	}
	return parts
}

type DnsDomainRecordsExportRecordsRequest struct {
	Headers   map[string]string
	DomainIds interface{} `json:"domain_ids,omitempty"`
}

func NewDnsDomainRecordsExportRecordsRequest() *DnsDomainRecordsExportRecordsRequest {
	return &DnsDomainRecordsExportRecordsRequest{Headers: map[string]string{}}
}

func (r *DnsDomainRecordsExportRecordsRequest) APIName() string {
	return "DnsDomainRecords_exportRecords"
}
func (r *DnsDomainRecordsExportRecordsRequest) Method() string { return "POST" }

func (r *DnsDomainRecordsExportRecordsRequest) SetHeader(key string, value string) *DnsDomainRecordsExportRecordsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DnsDomainRecordsExportRecordsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DnsDomainRecordsExportRecordsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.DomainIds != nil {
		parts.Data["domain_ids"] = r.DomainIds
	}
	return parts
}

type DnsDomainRecordsGetLinesRequest struct {
	Headers map[string]string
}

func NewDnsDomainRecordsGetLinesRequest() *DnsDomainRecordsGetLinesRequest {
	return &DnsDomainRecordsGetLinesRequest{Headers: map[string]string{}}
}

func (r *DnsDomainRecordsGetLinesRequest) APIName() string { return "DnsDomainRecords_getLines" }
func (r *DnsDomainRecordsGetLinesRequest) Method() string  { return "GET" }

func (r *DnsDomainRecordsGetLinesRequest) SetHeader(key string, value string) *DnsDomainRecordsGetLinesRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DnsDomainRecordsGetLinesRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DnsDomainRecordsGetLinesRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type DnsDomainRecordsBatchDeleteRecordsRequest struct {
	Headers   map[string]string
	DomainId  interface{} `json:"domain_id,omitempty"`
	RecordIds interface{} `json:"record_ids,omitempty"`
}

func NewDnsDomainRecordsBatchDeleteRecordsRequest() *DnsDomainRecordsBatchDeleteRecordsRequest {
	return &DnsDomainRecordsBatchDeleteRecordsRequest{Headers: map[string]string{}}
}

func (r *DnsDomainRecordsBatchDeleteRecordsRequest) APIName() string {
	return "DnsDomainRecords_batchDeleteRecords"
}
func (r *DnsDomainRecordsBatchDeleteRecordsRequest) Method() string { return "POST" }

func (r *DnsDomainRecordsBatchDeleteRecordsRequest) SetHeader(key string, value string) *DnsDomainRecordsBatchDeleteRecordsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DnsDomainRecordsBatchDeleteRecordsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DnsDomainRecordsBatchDeleteRecordsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.DomainId != nil {
		parts.Data["domain_id"] = r.DomainId
	}
	if r.RecordIds != nil {
		parts.Data["record_ids"] = r.RecordIds
	}
	return parts
}

type DnsDomainRecordsGetRecordGroupsListRequest struct {
	Headers map[string]string
}

func NewDnsDomainRecordsGetRecordGroupsListRequest() *DnsDomainRecordsGetRecordGroupsListRequest {
	return &DnsDomainRecordsGetRecordGroupsListRequest{Headers: map[string]string{}}
}

func (r *DnsDomainRecordsGetRecordGroupsListRequest) APIName() string {
	return "DnsDomainRecords_getRecordGroupsList"
}
func (r *DnsDomainRecordsGetRecordGroupsListRequest) Method() string { return "GET" }

func (r *DnsDomainRecordsGetRecordGroupsListRequest) SetHeader(key string, value string) *DnsDomainRecordsGetRecordGroupsListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DnsDomainRecordsGetRecordGroupsListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DnsDomainRecordsGetRecordGroupsListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type DnsDomainRecordsAddRecordGroupRequest struct {
	Headers   map[string]string
	DomainId  interface{} `json:"domain_id,omitempty"`
	GroupName interface{} `json:"group_name,omitempty"`
}

func NewDnsDomainRecordsAddRecordGroupRequest() *DnsDomainRecordsAddRecordGroupRequest {
	return &DnsDomainRecordsAddRecordGroupRequest{Headers: map[string]string{}}
}

func (r *DnsDomainRecordsAddRecordGroupRequest) APIName() string {
	return "DnsDomainRecords_addRecordGroup"
}
func (r *DnsDomainRecordsAddRecordGroupRequest) Method() string { return "POST" }

func (r *DnsDomainRecordsAddRecordGroupRequest) SetHeader(key string, value string) *DnsDomainRecordsAddRecordGroupRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DnsDomainRecordsAddRecordGroupRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DnsDomainRecordsAddRecordGroupRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.DomainId != nil {
		parts.Data["domain_id"] = r.DomainId
	}
	if r.GroupName != nil {
		parts.Data["group_name"] = r.GroupName
	}
	return parts
}

type DnsDomainRecordsAddRecordGroupRelationsRequest struct {
	Headers   map[string]string
	GroupId   interface{} `json:"group_id,omitempty"`
	RecordIds interface{} `json:"record_ids,omitempty"`
}

func NewDnsDomainRecordsAddRecordGroupRelationsRequest() *DnsDomainRecordsAddRecordGroupRelationsRequest {
	return &DnsDomainRecordsAddRecordGroupRelationsRequest{Headers: map[string]string{}}
}

func (r *DnsDomainRecordsAddRecordGroupRelationsRequest) APIName() string {
	return "DnsDomainRecords_addRecordGroupRelations"
}
func (r *DnsDomainRecordsAddRecordGroupRelationsRequest) Method() string { return "POST" }

func (r *DnsDomainRecordsAddRecordGroupRelationsRequest) SetHeader(key string, value string) *DnsDomainRecordsAddRecordGroupRelationsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DnsDomainRecordsAddRecordGroupRelationsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DnsDomainRecordsAddRecordGroupRelationsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.GroupId != nil {
		parts.Data["group_id"] = r.GroupId
	}
	if r.RecordIds != nil {
		parts.Data["record_ids"] = r.RecordIds
	}
	return parts
}

type DnsDomainRecordsDeleteRecordGroupRequest struct {
	Headers map[string]string
	GroupId interface{} `json:"group_id,omitempty"`
}

func NewDnsDomainRecordsDeleteRecordGroupRequest() *DnsDomainRecordsDeleteRecordGroupRequest {
	return &DnsDomainRecordsDeleteRecordGroupRequest{Headers: map[string]string{}}
}

func (r *DnsDomainRecordsDeleteRecordGroupRequest) APIName() string {
	return "DnsDomainRecords_deleteRecordGroup"
}
func (r *DnsDomainRecordsDeleteRecordGroupRequest) Method() string { return "DELETE" }

func (r *DnsDomainRecordsDeleteRecordGroupRequest) SetHeader(key string, value string) *DnsDomainRecordsDeleteRecordGroupRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DnsDomainRecordsDeleteRecordGroupRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DnsDomainRecordsDeleteRecordGroupRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.GroupId != nil {
		parts.Data["group_id"] = r.GroupId
	}
	return parts
}

type UserIpUserIpListRequest struct {
	Headers map[string]string
}

func NewUserIpUserIpListRequest() *UserIpUserIpListRequest {
	return &UserIpUserIpListRequest{Headers: map[string]string{}}
}

func (r *UserIpUserIpListRequest) APIName() string { return "UserIp_userIpList" }
func (r *UserIpUserIpListRequest) Method() string  { return "GET" }

func (r *UserIpUserIpListRequest) SetHeader(key string, value string) *UserIpUserIpListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *UserIpUserIpListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &UserIpUserIpListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type UserIpUserIpAddRequest struct {
	Headers map[string]string
	Name    interface{} `json:"name,omitempty"`
	Remark  interface{} `json:"remark,omitempty"`
}

func NewUserIpUserIpAddRequest() *UserIpUserIpAddRequest {
	return &UserIpUserIpAddRequest{Headers: map[string]string{}}
}

func (r *UserIpUserIpAddRequest) APIName() string { return "UserIp_userIpAdd" }
func (r *UserIpUserIpAddRequest) Method() string  { return "POST" }

func (r *UserIpUserIpAddRequest) SetHeader(key string, value string) *UserIpUserIpAddRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *UserIpUserIpAddRequest) RequestParts() ReqParams {
	if r == nil {
		r = &UserIpUserIpAddRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Name != nil {
		parts.Data["name"] = r.Name
	}
	if r.Remark != nil {
		parts.Data["remark"] = r.Remark
	}
	return parts
}

type UserIpUserIpSaveRequest struct {
	Headers map[string]string
	Id      interface{} `json:"id,omitempty"`
	Name    interface{} `json:"name,omitempty"`
	Remark  interface{} `json:"remark,omitempty"`
}

func NewUserIpUserIpSaveRequest() *UserIpUserIpSaveRequest {
	return &UserIpUserIpSaveRequest{Headers: map[string]string{}}
}

func (r *UserIpUserIpSaveRequest) APIName() string { return "UserIp_userIpSave" }
func (r *UserIpUserIpSaveRequest) Method() string  { return "PUT" }

func (r *UserIpUserIpSaveRequest) SetHeader(key string, value string) *UserIpUserIpSaveRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *UserIpUserIpSaveRequest) RequestParts() ReqParams {
	if r == nil {
		r = &UserIpUserIpSaveRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Id != nil {
		parts.Data["id"] = r.Id
	}
	if r.Name != nil {
		parts.Data["name"] = r.Name
	}
	if r.Remark != nil {
		parts.Data["remark"] = r.Remark
	}
	return parts
}

type UserIpUserIpDelRequest struct {
	Headers map[string]string
	Ids     interface{} `json:"ids,omitempty"`
}

func NewUserIpUserIpDelRequest() *UserIpUserIpDelRequest {
	return &UserIpUserIpDelRequest{Headers: map[string]string{}}
}

func (r *UserIpUserIpDelRequest) APIName() string { return "UserIp_userIpDel" }
func (r *UserIpUserIpDelRequest) Method() string  { return "DELETE" }

func (r *UserIpUserIpDelRequest) SetHeader(key string, value string) *UserIpUserIpDelRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *UserIpUserIpDelRequest) RequestParts() ReqParams {
	if r == nil {
		r = &UserIpUserIpDelRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Ids != nil {
		parts.Data["ids"] = r.Ids
	}
	return parts
}

type UserIpListUserIpItemRequest struct {
	Headers map[string]string
}

func NewUserIpListUserIpItemRequest() *UserIpListUserIpItemRequest {
	return &UserIpListUserIpItemRequest{Headers: map[string]string{}}
}

func (r *UserIpListUserIpItemRequest) APIName() string { return "UserIp_listUserIpItem" }
func (r *UserIpListUserIpItemRequest) Method() string  { return "GET" }

func (r *UserIpListUserIpItemRequest) SetHeader(key string, value string) *UserIpListUserIpItemRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *UserIpListUserIpItemRequest) RequestParts() ReqParams {
	if r == nil {
		r = &UserIpListUserIpItemRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type UserIpAddUserIpItemRequest struct {
	Headers  map[string]string
	UserIpId interface{} `json:"user_ip_id,omitempty"`
	Ip       interface{} `json:"ip,omitempty"`
	Remark   interface{} `json:"remark,omitempty"`
}

func NewUserIpAddUserIpItemRequest() *UserIpAddUserIpItemRequest {
	return &UserIpAddUserIpItemRequest{Headers: map[string]string{}}
}

func (r *UserIpAddUserIpItemRequest) APIName() string { return "UserIp_AddUserIpItem" }
func (r *UserIpAddUserIpItemRequest) Method() string  { return "POST" }

func (r *UserIpAddUserIpItemRequest) SetHeader(key string, value string) *UserIpAddUserIpItemRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *UserIpAddUserIpItemRequest) RequestParts() ReqParams {
	if r == nil {
		r = &UserIpAddUserIpItemRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.UserIpId != nil {
		parts.Data["user_ip_id"] = r.UserIpId
	}
	if r.Ip != nil {
		parts.Data["ip"] = r.Ip
	}
	if r.Remark != nil {
		parts.Data["remark"] = r.Remark
	}
	return parts
}

type UserIpUpdateUserIpItemRequest struct {
	Headers  map[string]string
	Id       interface{} `json:"_id,omitempty"`
	UserIpId interface{} `json:"user_ip_id,omitempty"`
	Ip       interface{} `json:"ip,omitempty"`
	Remark   interface{} `json:"remark,omitempty"`
}

func NewUserIpUpdateUserIpItemRequest() *UserIpUpdateUserIpItemRequest {
	return &UserIpUpdateUserIpItemRequest{Headers: map[string]string{}}
}

func (r *UserIpUpdateUserIpItemRequest) APIName() string { return "UserIp_UpdateUserIpItem" }
func (r *UserIpUpdateUserIpItemRequest) Method() string  { return "PUT" }

func (r *UserIpUpdateUserIpItemRequest) SetHeader(key string, value string) *UserIpUpdateUserIpItemRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *UserIpUpdateUserIpItemRequest) RequestParts() ReqParams {
	if r == nil {
		r = &UserIpUpdateUserIpItemRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Id != nil {
		parts.Data["_id"] = r.Id
	}
	if r.UserIpId != nil {
		parts.Data["user_ip_id"] = r.UserIpId
	}
	if r.Ip != nil {
		parts.Data["ip"] = r.Ip
	}
	if r.Remark != nil {
		parts.Data["remark"] = r.Remark
	}
	return parts
}

type UserIpBatchDeleteUserIpItemRequest struct {
	Headers  map[string]string
	Ids      interface{} `json:"ids,omitempty"`
	UserIpId interface{} `json:"user_ip_id,omitempty"`
}

func NewUserIpBatchDeleteUserIpItemRequest() *UserIpBatchDeleteUserIpItemRequest {
	return &UserIpBatchDeleteUserIpItemRequest{Headers: map[string]string{}}
}

func (r *UserIpBatchDeleteUserIpItemRequest) APIName() string { return "UserIp_BatchDeleteUserIpItem" }
func (r *UserIpBatchDeleteUserIpItemRequest) Method() string  { return "DELETE" }

func (r *UserIpBatchDeleteUserIpItemRequest) SetHeader(key string, value string) *UserIpBatchDeleteUserIpItemRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *UserIpBatchDeleteUserIpItemRequest) RequestParts() ReqParams {
	if r == nil {
		r = &UserIpBatchDeleteUserIpItemRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Ids != nil {
		parts.Data["ids"] = r.Ids
	}
	if r.UserIpId != nil {
		parts.Data["user_ip_id"] = r.UserIpId
	}
	return parts
}

type UserIpDeleteAllUserIpItemRequest struct {
	Headers  map[string]string
	UserIpId interface{} `json:"user_ip_id,omitempty"`
}

func NewUserIpDeleteAllUserIpItemRequest() *UserIpDeleteAllUserIpItemRequest {
	return &UserIpDeleteAllUserIpItemRequest{Headers: map[string]string{}}
}

func (r *UserIpDeleteAllUserIpItemRequest) APIName() string { return "UserIp_DeleteAllUserIpItem" }
func (r *UserIpDeleteAllUserIpItemRequest) Method() string  { return "POST" }

func (r *UserIpDeleteAllUserIpItemRequest) SetHeader(key string, value string) *UserIpDeleteAllUserIpItemRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *UserIpDeleteAllUserIpItemRequest) RequestParts() ReqParams {
	if r == nil {
		r = &UserIpDeleteAllUserIpItemRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.UserIpId != nil {
		parts.Data["user_ip_id"] = r.UserIpId
	}
	return parts
}

type UserIpCopyUserIpRequest struct {
	Headers  map[string]string
	UserIpId interface{} `json:"user_ip_id,omitempty"`
	Name     interface{} `json:"name,omitempty"`
	Remark   interface{} `json:"remark,omitempty"`
}

func NewUserIpCopyUserIpRequest() *UserIpCopyUserIpRequest {
	return &UserIpCopyUserIpRequest{Headers: map[string]string{}}
}

func (r *UserIpCopyUserIpRequest) APIName() string { return "UserIp_CopyUserIp" }
func (r *UserIpCopyUserIpRequest) Method() string  { return "POST" }

func (r *UserIpCopyUserIpRequest) SetHeader(key string, value string) *UserIpCopyUserIpRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *UserIpCopyUserIpRequest) RequestParts() ReqParams {
	if r == nil {
		r = &UserIpCopyUserIpRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.UserIpId != nil {
		parts.Data["user_ip_id"] = r.UserIpId
	}
	if r.Name != nil {
		parts.Data["name"] = r.Name
	}
	if r.Remark != nil {
		parts.Data["remark"] = r.Remark
	}
	return parts
}

type UserIpFileSaveIpItemRequest struct {
	Headers        map[string]string
	ContentType    interface{} `json:"Content-Type,omitempty"`
	XToken         interface{} `json:"x-token,omitempty"`
	AcceptLanguage interface{} `json:"Accept-Language,omitempty"`
	File           interface{} `json:"file,omitempty"`
	UserIpId       interface{} `json:"user_ip_id,omitempty"`
	Remark         interface{} `json:"remark,omitempty"`
}

func NewUserIpFileSaveIpItemRequest() *UserIpFileSaveIpItemRequest {
	return &UserIpFileSaveIpItemRequest{Headers: map[string]string{}}
}

func (r *UserIpFileSaveIpItemRequest) APIName() string { return "UserIp_FileSaveIpItem" }
func (r *UserIpFileSaveIpItemRequest) Method() string  { return "POST" }

func (r *UserIpFileSaveIpItemRequest) SetHeader(key string, value string) *UserIpFileSaveIpItemRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *UserIpFileSaveIpItemRequest) RequestParts() ReqParams {
	if r == nil {
		r = &UserIpFileSaveIpItemRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.ContentType != nil {
		parts.Headers["Content-Type"] = fmt.Sprint(r.ContentType)
	}
	if r.XToken != nil {
		parts.Headers["x-token"] = fmt.Sprint(r.XToken)
	}
	if r.AcceptLanguage != nil {
		parts.Headers["Accept-Language"] = fmt.Sprint(r.AcceptLanguage)
	}
	if r.File != nil {
		parts.Data["file"] = r.File
	}
	if r.UserIpId != nil {
		parts.Data["user_ip_id"] = r.UserIpId
	}
	if r.Remark != nil {
		parts.Data["remark"] = r.Remark
	}
	return parts
}

type ServiceBatchListTaskRequest struct {
	Headers map[string]string
}

func NewServiceBatchListTaskRequest() *ServiceBatchListTaskRequest {
	return &ServiceBatchListTaskRequest{Headers: map[string]string{}}
}

func (r *ServiceBatchListTaskRequest) APIName() string { return "service_batch_ListTask" }
func (r *ServiceBatchListTaskRequest) Method() string  { return "GET" }

func (r *ServiceBatchListTaskRequest) SetHeader(key string, value string) *ServiceBatchListTaskRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *ServiceBatchListTaskRequest) RequestParts() ReqParams {
	if r == nil {
		r = &ServiceBatchListTaskRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type ServiceBatchListSubTaskRequest struct {
	Headers map[string]string
}

func NewServiceBatchListSubTaskRequest() *ServiceBatchListSubTaskRequest {
	return &ServiceBatchListSubTaskRequest{Headers: map[string]string{}}
}

func (r *ServiceBatchListSubTaskRequest) APIName() string { return "service_batch_ListSubTask" }
func (r *ServiceBatchListSubTaskRequest) Method() string  { return "GET" }

func (r *ServiceBatchListSubTaskRequest) SetHeader(key string, value string) *ServiceBatchListSubTaskRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *ServiceBatchListSubTaskRequest) RequestParts() ReqParams {
	if r == nil {
		r = &ServiceBatchListSubTaskRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type WebCdnCleanCacheGetCacheListRequest struct {
	Headers map[string]string
}

func NewWebCdnCleanCacheGetCacheListRequest() *WebCdnCleanCacheGetCacheListRequest {
	return &WebCdnCleanCacheGetCacheListRequest{Headers: map[string]string{}}
}

func (r *WebCdnCleanCacheGetCacheListRequest) APIName() string {
	return "WebCdnCleanCache_getCacheList"
}
func (r *WebCdnCleanCacheGetCacheListRequest) Method() string { return "GET" }

func (r *WebCdnCleanCacheGetCacheListRequest) SetHeader(key string, value string) *WebCdnCleanCacheGetCacheListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *WebCdnCleanCacheGetCacheListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &WebCdnCleanCacheGetCacheListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type WebCdnCleanCacheSaveCacheRequest struct {
	Headers    map[string]string
	GroupId    interface{} `json:"group_id,omitempty"`
	Protocol   interface{} `json:"protocol,omitempty"`
	Port       interface{} `json:"port,omitempty"`
	Wholesite  interface{} `json:"wholesite,omitempty"`
	Specialurl interface{} `json:"specialurl,omitempty"`
	Specialdir interface{} `json:"specialdir,omitempty"`
}

func NewWebCdnCleanCacheSaveCacheRequest() *WebCdnCleanCacheSaveCacheRequest {
	return &WebCdnCleanCacheSaveCacheRequest{Headers: map[string]string{}}
}

func (r *WebCdnCleanCacheSaveCacheRequest) APIName() string { return "WebCdnCleanCache_saveCache" }
func (r *WebCdnCleanCacheSaveCacheRequest) Method() string  { return "PUT" }

func (r *WebCdnCleanCacheSaveCacheRequest) SetHeader(key string, value string) *WebCdnCleanCacheSaveCacheRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *WebCdnCleanCacheSaveCacheRequest) RequestParts() ReqParams {
	if r == nil {
		r = &WebCdnCleanCacheSaveCacheRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.GroupId != nil {
		parts.Data["group_id"] = r.GroupId
	}
	if r.Protocol != nil {
		parts.Data["protocol"] = r.Protocol
	}
	if r.Port != nil {
		parts.Data["port"] = r.Port
	}
	if r.Wholesite != nil {
		parts.Data["wholesite"] = r.Wholesite
	}
	if r.Specialurl != nil {
		parts.Data["specialurl"] = r.Specialurl
	}
	if r.Specialdir != nil {
		parts.Data["specialdir"] = r.Specialdir
	}
	return parts
}

type WebCdnCleanCacheGetTaskListRequest struct {
	Headers map[string]string
}

func NewWebCdnCleanCacheGetTaskListRequest() *WebCdnCleanCacheGetTaskListRequest {
	return &WebCdnCleanCacheGetTaskListRequest{Headers: map[string]string{}}
}

func (r *WebCdnCleanCacheGetTaskListRequest) APIName() string { return "WebCdnCleanCache_getTaskList" }
func (r *WebCdnCleanCacheGetTaskListRequest) Method() string  { return "GET" }

func (r *WebCdnCleanCacheGetTaskListRequest) SetHeader(key string, value string) *WebCdnCleanCacheGetTaskListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *WebCdnCleanCacheGetTaskListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &WebCdnCleanCacheGetTaskListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type WebCdnCleanCacheGetTaskDetailRequest struct {
	Headers map[string]string
}

func NewWebCdnCleanCacheGetTaskDetailRequest() *WebCdnCleanCacheGetTaskDetailRequest {
	return &WebCdnCleanCacheGetTaskDetailRequest{Headers: map[string]string{}}
}

func (r *WebCdnCleanCacheGetTaskDetailRequest) APIName() string {
	return "WebCdnCleanCache_getTaskDetail"
}
func (r *WebCdnCleanCacheGetTaskDetailRequest) Method() string { return "GET" }

func (r *WebCdnCleanCacheGetTaskDetailRequest) SetHeader(key string, value string) *WebCdnCleanCacheGetTaskDetailRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *WebCdnCleanCacheGetTaskDetailRequest) RequestParts() ReqParams {
	if r == nil {
		r = &WebCdnCleanCacheGetTaskDetailRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type WebCdnPreheatCacheGetPreheatCacheQuotaRequest struct {
	Headers map[string]string
}

func NewWebCdnPreheatCacheGetPreheatCacheQuotaRequest() *WebCdnPreheatCacheGetPreheatCacheQuotaRequest {
	return &WebCdnPreheatCacheGetPreheatCacheQuotaRequest{Headers: map[string]string{}}
}

func (r *WebCdnPreheatCacheGetPreheatCacheQuotaRequest) APIName() string {
	return "WebCdnPreheatCache_getPreheatCacheQuota"
}
func (r *WebCdnPreheatCacheGetPreheatCacheQuotaRequest) Method() string { return "GET" }

func (r *WebCdnPreheatCacheGetPreheatCacheQuotaRequest) SetHeader(key string, value string) *WebCdnPreheatCacheGetPreheatCacheQuotaRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *WebCdnPreheatCacheGetPreheatCacheQuotaRequest) RequestParts() ReqParams {
	if r == nil {
		r = &WebCdnPreheatCacheGetPreheatCacheQuotaRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type WebCdnPreheatCacheGetPreheatCacheListRequest struct {
	Headers map[string]string
}

func NewWebCdnPreheatCacheGetPreheatCacheListRequest() *WebCdnPreheatCacheGetPreheatCacheListRequest {
	return &WebCdnPreheatCacheGetPreheatCacheListRequest{Headers: map[string]string{}}
}

func (r *WebCdnPreheatCacheGetPreheatCacheListRequest) APIName() string {
	return "WebCdnPreheatCache_getPreheatCacheList"
}
func (r *WebCdnPreheatCacheGetPreheatCacheListRequest) Method() string { return "GET" }

func (r *WebCdnPreheatCacheGetPreheatCacheListRequest) SetHeader(key string, value string) *WebCdnPreheatCacheGetPreheatCacheListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *WebCdnPreheatCacheGetPreheatCacheListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &WebCdnPreheatCacheGetPreheatCacheListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type WebCdnPreheatCacheSavePreheatCacheRequest struct {
	Headers    map[string]string
	GroupId    interface{} `json:"group_id,omitempty"`
	Protocol   interface{} `json:"protocol,omitempty"`
	Port       interface{} `json:"port,omitempty"`
	PreheatUrl interface{} `json:"preheat_url,omitempty"`
}

func NewWebCdnPreheatCacheSavePreheatCacheRequest() *WebCdnPreheatCacheSavePreheatCacheRequest {
	return &WebCdnPreheatCacheSavePreheatCacheRequest{Headers: map[string]string{}}
}

func (r *WebCdnPreheatCacheSavePreheatCacheRequest) APIName() string {
	return "WebCdnPreheatCache_savePreheatCache"
}
func (r *WebCdnPreheatCacheSavePreheatCacheRequest) Method() string { return "POST" }

func (r *WebCdnPreheatCacheSavePreheatCacheRequest) SetHeader(key string, value string) *WebCdnPreheatCacheSavePreheatCacheRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *WebCdnPreheatCacheSavePreheatCacheRequest) RequestParts() ReqParams {
	if r == nil {
		r = &WebCdnPreheatCacheSavePreheatCacheRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.GroupId != nil {
		parts.Data["group_id"] = r.GroupId
	}
	if r.Protocol != nil {
		parts.Data["protocol"] = r.Protocol
	}
	if r.Port != nil {
		parts.Data["port"] = r.Port
	}
	if r.PreheatUrl != nil {
		parts.Data["preheat_url"] = r.PreheatUrl
	}
	return parts
}

type OplogInfoRequest struct {
	Headers map[string]string
}

func NewOplogInfoRequest() *OplogInfoRequest {
	return &OplogInfoRequest{Headers: map[string]string{}}
}

func (r *OplogInfoRequest) APIName() string { return "Oplog_info" }
func (r *OplogInfoRequest) Method() string  { return "GET" }

func (r *OplogInfoRequest) SetHeader(key string, value string) *OplogInfoRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *OplogInfoRequest) RequestParts() ReqParams {
	if r == nil {
		r = &OplogInfoRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type OplogMapRequest struct {
	Headers map[string]string
}

func NewOplogMapRequest() *OplogMapRequest {
	return &OplogMapRequest{Headers: map[string]string{}}
}

func (r *OplogMapRequest) APIName() string { return "Oplog_map" }
func (r *OplogMapRequest) Method() string  { return "GET" }

func (r *OplogMapRequest) SetHeader(key string, value string) *OplogMapRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *OplogMapRequest) RequestParts() ReqParams {
	if r == nil {
		r = &OplogMapRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type OplogGetOplogsRequest struct {
	Headers map[string]string
}

func NewOplogGetOplogsRequest() *OplogGetOplogsRequest {
	return &OplogGetOplogsRequest{Headers: map[string]string{}}
}

func (r *OplogGetOplogsRequest) APIName() string { return "Oplog_getOplogs" }
func (r *OplogGetOplogsRequest) Method() string  { return "GET" }

func (r *OplogGetOplogsRequest) SetHeader(key string, value string) *OplogGetOplogsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *OplogGetOplogsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &OplogGetOplogsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type CaCertificateSelfAddCaRequest struct {
	Headers     map[string]string
	CaName      interface{} `json:"ca_name,omitempty"`
	ProductFlag interface{} `json:"product_flag,omitempty"`
	CaCrt       interface{} `json:"ca_crt,omitempty"`
	CaKey       interface{} `json:"ca_key,omitempty"`
}

func NewCaCertificateSelfAddCaRequest() *CaCertificateSelfAddCaRequest {
	return &CaCertificateSelfAddCaRequest{Headers: map[string]string{}}
}

func (r *CaCertificateSelfAddCaRequest) APIName() string { return "CaCertificateSelf_addCa" }
func (r *CaCertificateSelfAddCaRequest) Method() string  { return "POST" }

func (r *CaCertificateSelfAddCaRequest) SetHeader(key string, value string) *CaCertificateSelfAddCaRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CaCertificateSelfAddCaRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CaCertificateSelfAddCaRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.CaName != nil {
		parts.Data["ca_name"] = r.CaName
	}
	if r.ProductFlag != nil {
		parts.Data["product_flag"] = r.ProductFlag
	}
	if r.CaCrt != nil {
		parts.Data["ca_crt"] = r.CaCrt
	}
	if r.CaKey != nil {
		parts.Data["ca_key"] = r.CaKey
	}
	return parts
}

type BatchCaListRequest struct {
	Headers map[string]string
	Domains interface{} `json:"domains,omitempty"`
}

func NewBatchCaListRequest() *BatchCaListRequest {
	return &BatchCaListRequest{Headers: map[string]string{}}
}

func (r *BatchCaListRequest) APIName() string { return "Batch_caList" }
func (r *BatchCaListRequest) Method() string  { return "POST" }

func (r *BatchCaListRequest) SetHeader(key string, value string) *BatchCaListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *BatchCaListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &BatchCaListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Domains != nil {
		parts.Data["domains"] = r.Domains
	}
	return parts
}

type CaCertificateSelfSaveTextCaInfoRequest struct {
	Headers     map[string]string
	Id          interface{} `json:"id,omitempty"`
	CaName      interface{} `json:"ca_name,omitempty"`
	ProductFlag interface{} `json:"product_flag,omitempty"`
	CaCert      interface{} `json:"ca_cert,omitempty"`
	CaKey       interface{} `json:"ca_key,omitempty"`
}

func NewCaCertificateSelfSaveTextCaInfoRequest() *CaCertificateSelfSaveTextCaInfoRequest {
	return &CaCertificateSelfSaveTextCaInfoRequest{Headers: map[string]string{}}
}

func (r *CaCertificateSelfSaveTextCaInfoRequest) APIName() string {
	return "CaCertificateSelf_saveTextCaInfo"
}
func (r *CaCertificateSelfSaveTextCaInfoRequest) Method() string { return "POST" }

func (r *CaCertificateSelfSaveTextCaInfoRequest) SetHeader(key string, value string) *CaCertificateSelfSaveTextCaInfoRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CaCertificateSelfSaveTextCaInfoRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CaCertificateSelfSaveTextCaInfoRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Id != nil {
		parts.Data["id"] = r.Id
	}
	if r.CaName != nil {
		parts.Data["ca_name"] = r.CaName
	}
	if r.ProductFlag != nil {
		parts.Data["product_flag"] = r.ProductFlag
	}
	if r.CaCert != nil {
		parts.Data["ca_cert"] = r.CaCert
	}
	if r.CaKey != nil {
		parts.Data["ca_key"] = r.CaKey
	}
	return parts
}

type CaCertificateSelfEditCaInfoRequest struct {
	Headers     map[string]string
	Id          interface{} `json:"id,omitempty"`
	CaName      interface{} `json:"ca_name,omitempty"`
	ProductFlag interface{} `json:"product_flag,omitempty"`
	CaCert      interface{} `json:"ca_cert,omitempty"`
	CaKey       interface{} `json:"ca_key,omitempty"`
}

func NewCaCertificateSelfEditCaInfoRequest() *CaCertificateSelfEditCaInfoRequest {
	return &CaCertificateSelfEditCaInfoRequest{Headers: map[string]string{}}
}

func (r *CaCertificateSelfEditCaInfoRequest) APIName() string { return "CaCertificateSelf_editCaInfo" }
func (r *CaCertificateSelfEditCaInfoRequest) Method() string  { return "POST" }

func (r *CaCertificateSelfEditCaInfoRequest) SetHeader(key string, value string) *CaCertificateSelfEditCaInfoRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CaCertificateSelfEditCaInfoRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CaCertificateSelfEditCaInfoRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Id != nil {
		parts.Data["id"] = r.Id
	}
	if r.CaName != nil {
		parts.Data["ca_name"] = r.CaName
	}
	if r.ProductFlag != nil {
		parts.Data["product_flag"] = r.ProductFlag
	}
	if r.CaCert != nil {
		parts.Data["ca_cert"] = r.CaCert
	}
	if r.CaKey != nil {
		parts.Data["ca_key"] = r.CaKey
	}
	return parts
}

type CaCertificateSelfListCaRequest struct {
	Headers map[string]string
}

func NewCaCertificateSelfListCaRequest() *CaCertificateSelfListCaRequest {
	return &CaCertificateSelfListCaRequest{Headers: map[string]string{}}
}

func (r *CaCertificateSelfListCaRequest) APIName() string { return "CaCertificateSelf_listCa" }
func (r *CaCertificateSelfListCaRequest) Method() string  { return "GET" }

func (r *CaCertificateSelfListCaRequest) SetHeader(key string, value string) *CaCertificateSelfListCaRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CaCertificateSelfListCaRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CaCertificateSelfListCaRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type CaCertificateSelfCaExportRequest struct {
	Headers map[string]string
}

func NewCaCertificateSelfCaExportRequest() *CaCertificateSelfCaExportRequest {
	return &CaCertificateSelfCaExportRequest{Headers: map[string]string{}}
}

func (r *CaCertificateSelfCaExportRequest) APIName() string { return "CaCertificateSelf_caExport" }
func (r *CaCertificateSelfCaExportRequest) Method() string  { return "GET" }

func (r *CaCertificateSelfCaExportRequest) SetHeader(key string, value string) *CaCertificateSelfCaExportRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CaCertificateSelfCaExportRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CaCertificateSelfCaExportRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type CaCertificateSelfBatchOperatSslRequest struct {
	Headers     map[string]string
	Id          interface{} `json:"id,omitempty"`
	TypeValue   interface{} `json:"type,omitempty"`
	ProductFlag interface{} `json:"product_flag,omitempty"`
	IsConfirm   interface{} `json:"is_confirm,omitempty"`
	DelId       interface{} `json:"del_id,omitempty"`
}

func NewCaCertificateSelfBatchOperatSslRequest() *CaCertificateSelfBatchOperatSslRequest {
	return &CaCertificateSelfBatchOperatSslRequest{Headers: map[string]string{}}
}

func (r *CaCertificateSelfBatchOperatSslRequest) APIName() string {
	return "CaCertificateSelf_batchOperatSsl"
}
func (r *CaCertificateSelfBatchOperatSslRequest) Method() string { return "GET" }

func (r *CaCertificateSelfBatchOperatSslRequest) SetHeader(key string, value string) *CaCertificateSelfBatchOperatSslRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CaCertificateSelfBatchOperatSslRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CaCertificateSelfBatchOperatSslRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Id != nil {
		parts.Query["id"] = r.Id
	}
	if r.TypeValue != nil {
		parts.Query["type"] = r.TypeValue
	}
	if r.ProductFlag != nil {
		parts.Query["product_flag"] = r.ProductFlag
	}
	parts.Query["is_confirm"] = "1"
	if r.IsConfirm != nil {
		parts.Query["is_confirm"] = r.IsConfirm
	}
	if r.DelId != nil {
		parts.Query["del_id"] = r.DelId
	}
	return parts
}

type CaCertificateSelfDelCaRequest struct {
	Headers     map[string]string
	Ids         interface{} `json:"ids,omitempty"`
	ProductFlag interface{} `json:"product_flag,omitempty"`
}

func NewCaCertificateSelfDelCaRequest() *CaCertificateSelfDelCaRequest {
	return &CaCertificateSelfDelCaRequest{Headers: map[string]string{}}
}

func (r *CaCertificateSelfDelCaRequest) APIName() string { return "CaCertificateSelf_delCa" }
func (r *CaCertificateSelfDelCaRequest) Method() string  { return "DELETE" }

func (r *CaCertificateSelfDelCaRequest) SetHeader(key string, value string) *CaCertificateSelfDelCaRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CaCertificateSelfDelCaRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CaCertificateSelfDelCaRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Ids != nil {
		parts.Data["ids"] = r.Ids
	}
	if r.ProductFlag != nil {
		parts.Data["product_flag"] = r.ProductFlag
	}
	return parts
}

type CaCertificateSelfGetCaDetailRequest struct {
	Headers map[string]string
}

func NewCaCertificateSelfGetCaDetailRequest() *CaCertificateSelfGetCaDetailRequest {
	return &CaCertificateSelfGetCaDetailRequest{Headers: map[string]string{}}
}

func (r *CaCertificateSelfGetCaDetailRequest) APIName() string {
	return "CaCertificateSelf_getCaDetail"
}
func (r *CaCertificateSelfGetCaDetailRequest) Method() string { return "GET" }

func (r *CaCertificateSelfGetCaDetailRequest) SetHeader(key string, value string) *CaCertificateSelfGetCaDetailRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CaCertificateSelfGetCaDetailRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CaCertificateSelfGetCaDetailRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type CaCertificateSelfEditCaNameRequest struct {
	Headers     map[string]string
	Id          interface{} `json:"id,omitempty"`
	CaName      interface{} `json:"ca_name,omitempty"`
	ProductFlag interface{} `json:"product_flag,omitempty"`
}

func NewCaCertificateSelfEditCaNameRequest() *CaCertificateSelfEditCaNameRequest {
	return &CaCertificateSelfEditCaNameRequest{Headers: map[string]string{}}
}

func (r *CaCertificateSelfEditCaNameRequest) APIName() string { return "CaCertificateSelf_editCaName" }
func (r *CaCertificateSelfEditCaNameRequest) Method() string  { return "POST" }

func (r *CaCertificateSelfEditCaNameRequest) SetHeader(key string, value string) *CaCertificateSelfEditCaNameRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CaCertificateSelfEditCaNameRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CaCertificateSelfEditCaNameRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Id != nil {
		parts.Data["id"] = r.Id
	}
	if r.CaName != nil {
		parts.Data["ca_name"] = r.CaName
	}
	if r.ProductFlag != nil {
		parts.Data["product_flag"] = r.ProductFlag
	}
	return parts
}

type CaCertificateApplyAddApplyCaRequest struct {
	Headers   map[string]string
	Domain    interface{} `json:"domain,omitempty"`
	TypeValue interface{} `json:"type,omitempty"`
	CaType    interface{} `json:"ca_type,omitempty"`
}

func NewCaCertificateApplyAddApplyCaRequest() *CaCertificateApplyAddApplyCaRequest {
	return &CaCertificateApplyAddApplyCaRequest{Headers: map[string]string{}}
}

func (r *CaCertificateApplyAddApplyCaRequest) APIName() string {
	return "CaCertificateApply_addApplyCa"
}
func (r *CaCertificateApplyAddApplyCaRequest) Method() string { return "POST" }

func (r *CaCertificateApplyAddApplyCaRequest) SetHeader(key string, value string) *CaCertificateApplyAddApplyCaRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CaCertificateApplyAddApplyCaRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CaCertificateApplyAddApplyCaRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Domain != nil {
		parts.Data["domain"] = r.Domain
	}
	parts.Data["type"] = "1"
	if r.TypeValue != nil {
		parts.Data["type"] = r.TypeValue
	}
	parts.Data["ca_type"] = "2"
	if r.CaType != nil {
		parts.Data["ca_type"] = r.CaType
	}
	return parts
}

type CaCertificateApplyGetAddByNsSettingRequest struct {
	Headers map[string]string
}

func NewCaCertificateApplyGetAddByNsSettingRequest() *CaCertificateApplyGetAddByNsSettingRequest {
	return &CaCertificateApplyGetAddByNsSettingRequest{Headers: map[string]string{}}
}

func (r *CaCertificateApplyGetAddByNsSettingRequest) APIName() string {
	return "CaCertificateApply_getAddByNsSetting"
}
func (r *CaCertificateApplyGetAddByNsSettingRequest) Method() string { return "GET" }

func (r *CaCertificateApplyGetAddByNsSettingRequest) SetHeader(key string, value string) *CaCertificateApplyGetAddByNsSettingRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CaCertificateApplyGetAddByNsSettingRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CaCertificateApplyGetAddByNsSettingRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type DomainGroupSaveGroupRequest struct {
	Headers   map[string]string
	GroupId   interface{} `json:"group_id,omitempty"`
	GroupName interface{} `json:"group_name,omitempty"`
	Remark    interface{} `json:"remark,omitempty"`
}

func NewDomainGroupSaveGroupRequest() *DomainGroupSaveGroupRequest {
	return &DomainGroupSaveGroupRequest{Headers: map[string]string{}}
}

func (r *DomainGroupSaveGroupRequest) APIName() string { return "DomainGroup_saveGroup" }
func (r *DomainGroupSaveGroupRequest) Method() string  { return "POST" }

func (r *DomainGroupSaveGroupRequest) SetHeader(key string, value string) *DomainGroupSaveGroupRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DomainGroupSaveGroupRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DomainGroupSaveGroupRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.GroupId != nil {
		parts.Data["group_id"] = r.GroupId
	}
	if r.GroupName != nil {
		parts.Data["group_name"] = r.GroupName
	}
	if r.Remark != nil {
		parts.Data["remark"] = r.Remark
	}
	return parts
}

type DomainGroupGetGroupListRequest struct {
	Headers map[string]string
}

func NewDomainGroupGetGroupListRequest() *DomainGroupGetGroupListRequest {
	return &DomainGroupGetGroupListRequest{Headers: map[string]string{}}
}

func (r *DomainGroupGetGroupListRequest) APIName() string { return "DomainGroup_getGroupList" }
func (r *DomainGroupGetGroupListRequest) Method() string  { return "GET" }

func (r *DomainGroupGetGroupListRequest) SetHeader(key string, value string) *DomainGroupGetGroupListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DomainGroupGetGroupListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DomainGroupGetGroupListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type DomainGroupDelGroupRequest struct {
	Headers map[string]string
	GroupId interface{} `json:"group_id,omitempty"`
}

func NewDomainGroupDelGroupRequest() *DomainGroupDelGroupRequest {
	return &DomainGroupDelGroupRequest{Headers: map[string]string{}}
}

func (r *DomainGroupDelGroupRequest) APIName() string { return "DomainGroup_delGroup" }
func (r *DomainGroupDelGroupRequest) Method() string  { return "POST" }

func (r *DomainGroupDelGroupRequest) SetHeader(key string, value string) *DomainGroupDelGroupRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DomainGroupDelGroupRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DomainGroupDelGroupRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.GroupId != nil {
		parts.Data["group_id"] = r.GroupId
	}
	return parts
}

type DomainGroupGetGroupDomainListRequest struct {
	Headers map[string]string
}

func NewDomainGroupGetGroupDomainListRequest() *DomainGroupGetGroupDomainListRequest {
	return &DomainGroupGetGroupDomainListRequest{Headers: map[string]string{}}
}

func (r *DomainGroupGetGroupDomainListRequest) APIName() string {
	return "DomainGroup_getGroupDomainList"
}
func (r *DomainGroupGetGroupDomainListRequest) Method() string { return "GET" }

func (r *DomainGroupGetGroupDomainListRequest) SetHeader(key string, value string) *DomainGroupGetGroupDomainListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DomainGroupGetGroupDomainListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DomainGroupGetGroupDomainListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type DomainGroupGgtUndistributedDomainListRequest struct {
	Headers map[string]string
}

func NewDomainGroupGgtUndistributedDomainListRequest() *DomainGroupGgtUndistributedDomainListRequest {
	return &DomainGroupGgtUndistributedDomainListRequest{Headers: map[string]string{}}
}

func (r *DomainGroupGgtUndistributedDomainListRequest) APIName() string {
	return "DomainGroup_ggtUndistributedDomainList"
}
func (r *DomainGroupGgtUndistributedDomainListRequest) Method() string { return "GET" }

func (r *DomainGroupGgtUndistributedDomainListRequest) SetHeader(key string, value string) *DomainGroupGgtUndistributedDomainListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DomainGroupGgtUndistributedDomainListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DomainGroupGgtUndistributedDomainListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type DomainGroupAddGroupRequest struct {
	Headers   map[string]string
	GroupName interface{} `json:"group_name,omitempty"`
	Remark    interface{} `json:"remark,omitempty"`
}

func NewDomainGroupAddGroupRequest() *DomainGroupAddGroupRequest {
	return &DomainGroupAddGroupRequest{Headers: map[string]string{}}
}

func (r *DomainGroupAddGroupRequest) APIName() string { return "DomainGroup_addGroup" }
func (r *DomainGroupAddGroupRequest) Method() string  { return "POST" }

func (r *DomainGroupAddGroupRequest) SetHeader(key string, value string) *DomainGroupAddGroupRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DomainGroupAddGroupRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DomainGroupAddGroupRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.GroupName != nil {
		parts.Data["group_name"] = r.GroupName
	}
	if r.Remark != nil {
		parts.Data["remark"] = r.Remark
	}
	return parts
}

type DomainGroupSaveDomainToGroupRequest struct {
	Headers                  map[string]string
	GroupId                  interface{} `json:"group_id,omitempty"`
	DomainIds                interface{} `json:"domain_ids,omitempty"`
	Domains                  interface{} `json:"domains,omitempty"`
	OnlyUnbindTplDomainGroup interface{} `json:"only_unbind_tpl_domain_group,omitempty"`
	Action                   interface{} `json:"action,omitempty"`
}

func NewDomainGroupSaveDomainToGroupRequest() *DomainGroupSaveDomainToGroupRequest {
	return &DomainGroupSaveDomainToGroupRequest{Headers: map[string]string{}}
}

func (r *DomainGroupSaveDomainToGroupRequest) APIName() string {
	return "DomainGroup_saveDomainToGroup"
}
func (r *DomainGroupSaveDomainToGroupRequest) Method() string { return "POST" }

func (r *DomainGroupSaveDomainToGroupRequest) SetHeader(key string, value string) *DomainGroupSaveDomainToGroupRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DomainGroupSaveDomainToGroupRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DomainGroupSaveDomainToGroupRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.GroupId != nil {
		parts.Data["group_id"] = r.GroupId
	}
	if r.DomainIds != nil {
		parts.Data["domain_ids"] = r.DomainIds
	}
	if r.Domains != nil {
		parts.Data["domains"] = r.Domains
	}
	if r.OnlyUnbindTplDomainGroup != nil {
		parts.Data["only_unbind_tpl_domain_group"] = r.OnlyUnbindTplDomainGroup
	}
	if r.Action != nil {
		parts.Data["action"] = r.Action
	}
	return parts
}

type DomainGroupGetGroupInfoRequest struct {
	Headers map[string]string
}

func NewDomainGroupGetGroupInfoRequest() *DomainGroupGetGroupInfoRequest {
	return &DomainGroupGetGroupInfoRequest{Headers: map[string]string{}}
}

func (r *DomainGroupGetGroupInfoRequest) APIName() string { return "DomainGroup_getGroupInfo" }
func (r *DomainGroupGetGroupInfoRequest) Method() string  { return "GET" }

func (r *DomainGroupGetGroupInfoRequest) SetHeader(key string, value string) *DomainGroupGetGroupInfoRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DomainGroupGetGroupInfoRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DomainGroupGetGroupInfoRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type DomainGroupMoveDomainRequest struct {
	Headers     map[string]string
	FromGroupId interface{} `json:"from_group_id,omitempty"`
	ToGroupId   interface{} `json:"to_group_id,omitempty"`
	DomainIds   interface{} `json:"domain_ids,omitempty"`
}

func NewDomainGroupMoveDomainRequest() *DomainGroupMoveDomainRequest {
	return &DomainGroupMoveDomainRequest{Headers: map[string]string{}}
}

func (r *DomainGroupMoveDomainRequest) APIName() string { return "DomainGroup_moveDomain" }
func (r *DomainGroupMoveDomainRequest) Method() string  { return "POST" }

func (r *DomainGroupMoveDomainRequest) SetHeader(key string, value string) *DomainGroupMoveDomainRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DomainGroupMoveDomainRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DomainGroupMoveDomainRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.FromGroupId != nil {
		parts.Data["from_group_id"] = r.FromGroupId
	}
	if r.ToGroupId != nil {
		parts.Data["to_group_id"] = r.ToGroupId
	}
	if r.DomainIds != nil {
		parts.Data["domain_ids"] = r.DomainIds
	}
	return parts
}

type ListDomainsRequest struct {
	Headers             map[string]string
	Page                interface{} `json:"page,omitempty"`
	PageSize            interface{} `json:"page_size,omitempty"`
	AccessProgress      interface{} `json:"access_progress,omitempty"`
	GroupId             interface{} `json:"group_id,omitempty"`
	Domain              interface{} `json:"domain,omitempty"`
	Remark              interface{} `json:"remark,omitempty"`
	OriginIp            interface{} `json:"origin_ip,omitempty"`
	CaStatus            interface{} `json:"ca_status,omitempty"`
	AccessMode          interface{} `json:"access_mode,omitempty"`
	ProtectStatus       interface{} `json:"protect_status,omitempty"`
	ExclusiveResourceId interface{} `json:"exclusive_resource_id,omitempty"`
}

func NewListDomainsRequest() *ListDomainsRequest {
	return &ListDomainsRequest{Headers: map[string]string{}}
}

func (r *ListDomainsRequest) APIName() string { return "ListDomains" }
func (r *ListDomainsRequest) Method() string  { return "GET" }

func (r *ListDomainsRequest) SetHeader(key string, value string) *ListDomainsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *ListDomainsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &ListDomainsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Page != nil {
		parts.Query["page"] = r.Page
	}
	if r.PageSize != nil {
		parts.Query["page_size"] = r.PageSize
	}
	if r.AccessProgress != nil {
		parts.Query["access_progress"] = r.AccessProgress
	}
	if r.GroupId != nil {
		parts.Query["group_id"] = r.GroupId
	}
	if r.Domain != nil {
		parts.Query["domain"] = r.Domain
	}
	if r.Remark != nil {
		parts.Query["remark"] = r.Remark
	}
	if r.OriginIp != nil {
		parts.Query["origin_ip"] = r.OriginIp
	}
	if r.CaStatus != nil {
		parts.Query["ca_status"] = r.CaStatus
	}
	if r.AccessMode != nil {
		parts.Query["access_mode"] = r.AccessMode
	}
	if r.ProtectStatus != nil {
		parts.Query["protect_status"] = r.ProtectStatus
	}
	if r.ExclusiveResourceId != nil {
		parts.Query["exclusive_resource_id"] = r.ExclusiveResourceId
	}
	return parts
}

type AddDomainsRequest struct {
	Headers             map[string]string
	Domain              interface{} `json:"domain,omitempty"`
	GroupId             interface{} `json:"group_id,omitempty"`
	ExclusiveResourceId interface{} `json:"exclusive_resource_id,omitempty"`
	Remark              interface{} `json:"remark,omitempty"`
	TplId               interface{} `json:"tpl_id,omitempty"`
	Origins             interface{} `json:"origins,omitempty"`
	ProtectStatus       interface{} `json:"protect_status,omitempty"`
	TplRecommend        interface{} `json:"tpl_recommend,omitempty"`
}

func NewAddDomainsRequest() *AddDomainsRequest {
	return &AddDomainsRequest{Headers: map[string]string{}}
}

func (r *AddDomainsRequest) APIName() string { return "AddDomains" }
func (r *AddDomainsRequest) Method() string  { return "POST" }

func (r *AddDomainsRequest) SetHeader(key string, value string) *AddDomainsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *AddDomainsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &AddDomainsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Domain != nil {
		parts.Data["domain"] = r.Domain
	}
	if r.GroupId != nil {
		parts.Data["group_id"] = r.GroupId
	}
	if r.ExclusiveResourceId != nil {
		parts.Data["exclusive_resource_id"] = r.ExclusiveResourceId
	}
	if r.Remark != nil {
		parts.Data["remark"] = r.Remark
	}
	if r.TplId != nil {
		parts.Data["tpl_id"] = r.TplId
	}
	if r.Origins != nil {
		parts.Data["origins"] = r.Origins
	}
	if r.ProtectStatus != nil {
		parts.Data["protect_status"] = r.ProtectStatus
	}
	if r.TplRecommend != nil {
		parts.Data["tpl_recommend"] = r.TplRecommend
	}
	return parts
}

type UpdateDomainsRequest struct {
	Headers  map[string]string
	DomainId interface{} `json:"domain_id,omitempty"`
	Remark   interface{} `json:"remark,omitempty"`
}

func NewUpdateDomainsRequest() *UpdateDomainsRequest {
	return &UpdateDomainsRequest{Headers: map[string]string{}}
}

func (r *UpdateDomainsRequest) APIName() string { return "UpdateDomains" }
func (r *UpdateDomainsRequest) Method() string  { return "PUT" }

func (r *UpdateDomainsRequest) SetHeader(key string, value string) *UpdateDomainsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *UpdateDomainsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &UpdateDomainsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.DomainId != nil {
		parts.Data["domain_id"] = r.DomainId
	}
	if r.Remark != nil {
		parts.Data["remark"] = r.Remark
	}
	return parts
}

type BindDomainCertRequest struct {
	Headers  map[string]string
	DomainId interface{} `json:"domain_id,omitempty"`
	CaId     interface{} `json:"ca_id,omitempty"`
}

func NewBindDomainCertRequest() *BindDomainCertRequest {
	return &BindDomainCertRequest{Headers: map[string]string{}}
}

func (r *BindDomainCertRequest) APIName() string { return "BindDomainCert" }
func (r *BindDomainCertRequest) Method() string  { return "POST" }

func (r *BindDomainCertRequest) SetHeader(key string, value string) *BindDomainCertRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *BindDomainCertRequest) RequestParts() ReqParams {
	if r == nil {
		r = &BindDomainCertRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.DomainId != nil {
		parts.Data["domain_id"] = r.DomainId
	}
	if r.CaId != nil {
		parts.Data["ca_id"] = r.CaId
	}
	return parts
}

type UnBindDomainCertRequest struct {
	Headers  map[string]string
	DomainId interface{} `json:"domain_id,omitempty"`
	CaId     interface{} `json:"ca_id,omitempty"`
}

func NewUnBindDomainCertRequest() *UnBindDomainCertRequest {
	return &UnBindDomainCertRequest{Headers: map[string]string{}}
}

func (r *UnBindDomainCertRequest) APIName() string { return "UnBindDomainCert" }
func (r *UnBindDomainCertRequest) Method() string  { return "POST" }

func (r *UnBindDomainCertRequest) SetHeader(key string, value string) *UnBindDomainCertRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *UnBindDomainCertRequest) RequestParts() ReqParams {
	if r == nil {
		r = &UnBindDomainCertRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.DomainId != nil {
		parts.Data["domain_id"] = r.DomainId
	}
	if r.CaId != nil {
		parts.Data["ca_id"] = r.CaId
	}
	return parts
}

type DeleteDomainsRequest struct {
	Headers map[string]string
	Ids     interface{} `json:"ids,omitempty"`
}

func NewDeleteDomainsRequest() *DeleteDomainsRequest {
	return &DeleteDomainsRequest{Headers: map[string]string{}}
}

func (r *DeleteDomainsRequest) APIName() string { return "DeleteDomains" }
func (r *DeleteDomainsRequest) Method() string  { return "DELETE" }

func (r *DeleteDomainsRequest) SetHeader(key string, value string) *DeleteDomainsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DeleteDomainsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DeleteDomainsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Ids != nil {
		parts.Data["ids"] = r.Ids
	}
	return parts
}

type DisableDomainsRequest struct {
	Headers   map[string]string
	DomainIds interface{} `json:"domain_ids,omitempty"`
}

func NewDisableDomainsRequest() *DisableDomainsRequest {
	return &DisableDomainsRequest{Headers: map[string]string{}}
}

func (r *DisableDomainsRequest) APIName() string { return "DisableDomains" }
func (r *DisableDomainsRequest) Method() string  { return "POST" }

func (r *DisableDomainsRequest) SetHeader(key string, value string) *DisableDomainsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DisableDomainsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DisableDomainsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.DomainIds != nil {
		parts.Data["domain_ids"] = r.DomainIds
	}
	return parts
}

type EnableDomainsRequest struct {
	Headers   map[string]string
	DomainIds interface{} `json:"domain_ids,omitempty"`
}

func NewEnableDomainsRequest() *EnableDomainsRequest {
	return &EnableDomainsRequest{Headers: map[string]string{}}
}

func (r *EnableDomainsRequest) APIName() string { return "EnableDomains" }
func (r *EnableDomainsRequest) Method() string  { return "POST" }

func (r *EnableDomainsRequest) SetHeader(key string, value string) *EnableDomainsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *EnableDomainsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &EnableDomainsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.DomainIds != nil {
		parts.Data["domain_ids"] = r.DomainIds
	}
	return parts
}

type RefreshDomainsAccessRequest struct {
	Headers   map[string]string
	DomainIds interface{} `json:"domain_ids,omitempty"`
}

func NewRefreshDomainsAccessRequest() *RefreshDomainsAccessRequest {
	return &RefreshDomainsAccessRequest{Headers: map[string]string{}}
}

func (r *RefreshDomainsAccessRequest) APIName() string { return "RefreshDomainsAccess" }
func (r *RefreshDomainsAccessRequest) Method() string  { return "POST" }

func (r *RefreshDomainsAccessRequest) SetHeader(key string, value string) *RefreshDomainsAccessRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *RefreshDomainsAccessRequest) RequestParts() ReqParams {
	if r == nil {
		r = &RefreshDomainsAccessRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.DomainIds != nil {
		parts.Data["domain_ids"] = r.DomainIds
	}
	return parts
}

type ExportDomainsRequest struct {
	Headers             map[string]string
	DomainIds           interface{} `json:"domain_ids,omitempty"`
	AccessProgress      interface{} `json:"access_progress,omitempty"`
	GroupId             interface{} `json:"group_id,omitempty"`
	Domain              interface{} `json:"domain,omitempty"`
	Remark              interface{} `json:"remark,omitempty"`
	OriginIp            interface{} `json:"origin_ip,omitempty"`
	CaStatus            interface{} `json:"ca_status,omitempty"`
	AccessMode          interface{} `json:"access_mode,omitempty"`
	ProtectStatus       interface{} `json:"protect_status,omitempty"`
	ExclusiveResourceId interface{} `json:"exclusive_resource_id,omitempty"`
}

func NewExportDomainsRequest() *ExportDomainsRequest {
	return &ExportDomainsRequest{Headers: map[string]string{}}
}

func (r *ExportDomainsRequest) APIName() string { return "ExportDomains" }
func (r *ExportDomainsRequest) Method() string  { return "POST" }

func (r *ExportDomainsRequest) SetHeader(key string, value string) *ExportDomainsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *ExportDomainsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &ExportDomainsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.DomainIds != nil {
		parts.Data["domain_ids"] = r.DomainIds
	}
	if r.AccessProgress != nil {
		parts.Data["access_progress"] = r.AccessProgress
	}
	if r.GroupId != nil {
		parts.Data["group_id"] = r.GroupId
	}
	if r.Domain != nil {
		parts.Data["domain"] = r.Domain
	}
	if r.Remark != nil {
		parts.Data["remark"] = r.Remark
	}
	if r.OriginIp != nil {
		parts.Data["origin_ip"] = r.OriginIp
	}
	if r.CaStatus != nil {
		parts.Data["ca_status"] = r.CaStatus
	}
	if r.AccessMode != nil {
		parts.Data["access_mode"] = r.AccessMode
	}
	if r.ProtectStatus != nil {
		parts.Data["protect_status"] = r.ProtectStatus
	}
	if r.ExclusiveResourceId != nil {
		parts.Data["exclusive_resource_id"] = r.ExclusiveResourceId
	}
	return parts
}

type AddOriginsRequest struct {
	Headers  map[string]string
	DomainId interface{} `json:"domain_id,omitempty"`
	Origins  interface{} `json:"origins,omitempty"`
}

func NewAddOriginsRequest() *AddOriginsRequest {
	return &AddOriginsRequest{Headers: map[string]string{}}
}

func (r *AddOriginsRequest) APIName() string { return "AddOrigins" }
func (r *AddOriginsRequest) Method() string  { return "POST" }

func (r *AddOriginsRequest) SetHeader(key string, value string) *AddOriginsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *AddOriginsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &AddOriginsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.DomainId != nil {
		parts.Data["domain_id"] = r.DomainId
	}
	if r.Origins != nil {
		parts.Data["origins"] = r.Origins
	}
	return parts
}

type UpdateOriginsRequest struct {
	Headers  map[string]string
	DomainId interface{} `json:"domain_id,omitempty"`
	Origins  interface{} `json:"origins,omitempty"`
}

func NewUpdateOriginsRequest() *UpdateOriginsRequest {
	return &UpdateOriginsRequest{Headers: map[string]string{}}
}

func (r *UpdateOriginsRequest) APIName() string { return "UpdateOrigins" }
func (r *UpdateOriginsRequest) Method() string  { return "PUT" }

func (r *UpdateOriginsRequest) SetHeader(key string, value string) *UpdateOriginsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *UpdateOriginsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &UpdateOriginsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.DomainId != nil {
		parts.Data["domain_id"] = r.DomainId
	}
	if r.Origins != nil {
		parts.Data["origins"] = r.Origins
	}
	return parts
}

type DeleteOriginsRequest struct {
	Headers  map[string]string
	Ids      interface{} `json:"ids,omitempty"`
	DomainId interface{} `json:"domain_id,omitempty"`
}

func NewDeleteOriginsRequest() *DeleteOriginsRequest {
	return &DeleteOriginsRequest{Headers: map[string]string{}}
}

func (r *DeleteOriginsRequest) APIName() string { return "DeleteOrigins" }
func (r *DeleteOriginsRequest) Method() string  { return "DELETE" }

func (r *DeleteOriginsRequest) SetHeader(key string, value string) *DeleteOriginsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DeleteOriginsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DeleteOriginsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Ids != nil {
		parts.Data["ids"] = r.Ids
	}
	if r.DomainId != nil {
		parts.Data["domain_id"] = r.DomainId
	}
	return parts
}

type ListOriginsRequest struct {
	Headers  map[string]string
	DomainId interface{} `json:"domain_id,omitempty"`
}

func NewListOriginsRequest() *ListOriginsRequest {
	return &ListOriginsRequest{Headers: map[string]string{}}
}

func (r *ListOriginsRequest) APIName() string { return "ListOrigins" }
func (r *ListOriginsRequest) Method() string  { return "GET" }

func (r *ListOriginsRequest) SetHeader(key string, value string) *ListOriginsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *ListOriginsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &ListOriginsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.DomainId != nil {
		parts.Query["domain_id"] = r.DomainId
	}
	return parts
}

type SwitchDomainNodesRequest struct {
	Headers             map[string]string
	DomainId            interface{} `json:"domain_id,omitempty"`
	ProtectStatus       interface{} `json:"protect_status,omitempty"`
	ExclusiveResourceId interface{} `json:"exclusive_resource_id,omitempty"`
}

func NewSwitchDomainNodesRequest() *SwitchDomainNodesRequest {
	return &SwitchDomainNodesRequest{Headers: map[string]string{}}
}

func (r *SwitchDomainNodesRequest) APIName() string { return "SwitchDomainNodes" }
func (r *SwitchDomainNodesRequest) Method() string  { return "POST" }

func (r *SwitchDomainNodesRequest) SetHeader(key string, value string) *SwitchDomainNodesRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *SwitchDomainNodesRequest) RequestParts() ReqParams {
	if r == nil {
		r = &SwitchDomainNodesRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.DomainId != nil {
		parts.Data["domain_id"] = r.DomainId
	}
	if r.ProtectStatus != nil {
		parts.Data["protect_status"] = r.ProtectStatus
	}
	if r.ExclusiveResourceId != nil {
		parts.Data["exclusive_resource_id"] = r.ExclusiveResourceId
	}
	return parts
}

type SwitchDomainAccessModeRequest struct {
	Headers    map[string]string
	DomainId   interface{} `json:"domain_id,omitempty"`
	AccessMode interface{} `json:"access_mode,omitempty"`
}

func NewSwitchDomainAccessModeRequest() *SwitchDomainAccessModeRequest {
	return &SwitchDomainAccessModeRequest{Headers: map[string]string{}}
}

func (r *SwitchDomainAccessModeRequest) APIName() string { return "SwitchDomainAccessMode" }
func (r *SwitchDomainAccessModeRequest) Method() string  { return "POST" }

func (r *SwitchDomainAccessModeRequest) SetHeader(key string, value string) *SwitchDomainAccessModeRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *SwitchDomainAccessModeRequest) RequestParts() ReqParams {
	if r == nil {
		r = &SwitchDomainAccessModeRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.DomainId != nil {
		parts.Data["domain_id"] = r.DomainId
	}
	if r.AccessMode != nil {
		parts.Data["access_mode"] = r.AccessMode
	}
	return parts
}

type UpdateDomainBaseSettingsRequest struct {
	Headers  map[string]string
	DomainId interface{} `json:"domain_id,omitempty"`
	Value    interface{} `json:"value,omitempty"`
}

func NewUpdateDomainBaseSettingsRequest() *UpdateDomainBaseSettingsRequest {
	return &UpdateDomainBaseSettingsRequest{Headers: map[string]string{}}
}

func (r *UpdateDomainBaseSettingsRequest) APIName() string { return "UpdateDomainBaseSettings" }
func (r *UpdateDomainBaseSettingsRequest) Method() string  { return "PUT" }

func (r *UpdateDomainBaseSettingsRequest) SetHeader(key string, value string) *UpdateDomainBaseSettingsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *UpdateDomainBaseSettingsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &UpdateDomainBaseSettingsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.DomainId != nil {
		parts.Data["domain_id"] = r.DomainId
	}
	if r.Value != nil {
		parts.Data["value"] = r.Value
	}
	return parts
}

type GetDomainBaseSettingsRequest struct {
	Headers map[string]string
}

func NewGetDomainBaseSettingsRequest() *GetDomainBaseSettingsRequest {
	return &GetDomainBaseSettingsRequest{Headers: map[string]string{}}
}

func (r *GetDomainBaseSettingsRequest) APIName() string { return "GetDomainBaseSettings" }
func (r *GetDomainBaseSettingsRequest) Method() string  { return "GET" }

func (r *GetDomainBaseSettingsRequest) SetHeader(key string, value string) *GetDomainBaseSettingsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *GetDomainBaseSettingsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &GetDomainBaseSettingsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type ListBriefDomainsRequest struct {
	Headers map[string]string
	Ids     interface{} `json:"ids,omitempty"`
}

func NewListBriefDomainsRequest() *ListBriefDomainsRequest {
	return &ListBriefDomainsRequest{Headers: map[string]string{}}
}

func (r *ListBriefDomainsRequest) APIName() string { return "ListBriefDomains" }
func (r *ListBriefDomainsRequest) Method() string  { return "POST" }

func (r *ListBriefDomainsRequest) SetHeader(key string, value string) *ListBriefDomainsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *ListBriefDomainsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &ListBriefDomainsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Ids != nil {
		parts.Data["ids"] = r.Ids
	}
	return parts
}

type GetDomainTemplatesRequest struct {
	Headers map[string]string
}

func NewGetDomainTemplatesRequest() *GetDomainTemplatesRequest {
	return &GetDomainTemplatesRequest{Headers: map[string]string{}}
}

func (r *GetDomainTemplatesRequest) APIName() string { return "GetDomainTemplates" }
func (r *GetDomainTemplatesRequest) Method() string  { return "GET" }

func (r *GetDomainTemplatesRequest) SetHeader(key string, value string) *GetDomainTemplatesRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *GetDomainTemplatesRequest) RequestParts() ReqParams {
	if r == nil {
		r = &GetDomainTemplatesRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type AccessInfoDownloadRequest struct {
	Headers     map[string]string
	DomainInfos interface{} `json:"domain_infos,omitempty"`
}

func NewAccessInfoDownloadRequest() *AccessInfoDownloadRequest {
	return &AccessInfoDownloadRequest{Headers: map[string]string{}}
}

func (r *AccessInfoDownloadRequest) APIName() string { return "AccessInfoDownload" }
func (r *AccessInfoDownloadRequest) Method() string  { return "POST" }

func (r *AccessInfoDownloadRequest) SetHeader(key string, value string) *AccessInfoDownloadRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *AccessInfoDownloadRequest) RequestParts() ReqParams {
	if r == nil {
		r = &AccessInfoDownloadRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.DomainInfos != nil {
		parts.Data["domain_infos"] = r.DomainInfos
	}
	return parts
}

type OriginGroupGetOriginGroupListRequest struct {
	Headers map[string]string
}

func NewOriginGroupGetOriginGroupListRequest() *OriginGroupGetOriginGroupListRequest {
	return &OriginGroupGetOriginGroupListRequest{Headers: map[string]string{}}
}

func (r *OriginGroupGetOriginGroupListRequest) APIName() string {
	return "OriginGroup_getOriginGroupList"
}
func (r *OriginGroupGetOriginGroupListRequest) Method() string { return "GET" }

func (r *OriginGroupGetOriginGroupListRequest) SetHeader(key string, value string) *OriginGroupGetOriginGroupListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *OriginGroupGetOriginGroupListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &OriginGroupGetOriginGroupListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type OriginGroupGetOriginGroupInfoRequest struct {
	Headers map[string]string
}

func NewOriginGroupGetOriginGroupInfoRequest() *OriginGroupGetOriginGroupInfoRequest {
	return &OriginGroupGetOriginGroupInfoRequest{Headers: map[string]string{}}
}

func (r *OriginGroupGetOriginGroupInfoRequest) APIName() string {
	return "OriginGroup_getOriginGroupInfo"
}
func (r *OriginGroupGetOriginGroupInfoRequest) Method() string { return "GET" }

func (r *OriginGroupGetOriginGroupInfoRequest) SetHeader(key string, value string) *OriginGroupGetOriginGroupInfoRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *OriginGroupGetOriginGroupInfoRequest) RequestParts() ReqParams {
	if r == nil {
		r = &OriginGroupGetOriginGroupInfoRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type OriginGroupAddOriginGroupRequest struct {
	Headers map[string]string
	Name    interface{} `json:"name,omitempty"`
	Remark  interface{} `json:"remark,omitempty"`
	Origins interface{} `json:"origins,omitempty"`
}

func NewOriginGroupAddOriginGroupRequest() *OriginGroupAddOriginGroupRequest {
	return &OriginGroupAddOriginGroupRequest{Headers: map[string]string{}}
}

func (r *OriginGroupAddOriginGroupRequest) APIName() string { return "OriginGroup_addOriginGroup" }
func (r *OriginGroupAddOriginGroupRequest) Method() string  { return "POST" }

func (r *OriginGroupAddOriginGroupRequest) SetHeader(key string, value string) *OriginGroupAddOriginGroupRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *OriginGroupAddOriginGroupRequest) RequestParts() ReqParams {
	if r == nil {
		r = &OriginGroupAddOriginGroupRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Name != nil {
		parts.Data["name"] = r.Name
	}
	if r.Remark != nil {
		parts.Data["remark"] = r.Remark
	}
	if r.Origins != nil {
		parts.Data["origins"] = r.Origins
	}
	return parts
}

type OriginGroupUpdateOriginGroupRequest struct {
	Headers map[string]string
	Id      interface{} `json:"id,omitempty"`
	Name    interface{} `json:"name,omitempty"`
	Remark  interface{} `json:"remark,omitempty"`
	Origins interface{} `json:"origins,omitempty"`
}

func NewOriginGroupUpdateOriginGroupRequest() *OriginGroupUpdateOriginGroupRequest {
	return &OriginGroupUpdateOriginGroupRequest{Headers: map[string]string{}}
}

func (r *OriginGroupUpdateOriginGroupRequest) APIName() string {
	return "OriginGroup_updateOriginGroup"
}
func (r *OriginGroupUpdateOriginGroupRequest) Method() string { return "PUT" }

func (r *OriginGroupUpdateOriginGroupRequest) SetHeader(key string, value string) *OriginGroupUpdateOriginGroupRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *OriginGroupUpdateOriginGroupRequest) RequestParts() ReqParams {
	if r == nil {
		r = &OriginGroupUpdateOriginGroupRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Id != nil {
		parts.Data["id"] = r.Id
	}
	if r.Name != nil {
		parts.Data["name"] = r.Name
	}
	if r.Remark != nil {
		parts.Data["remark"] = r.Remark
	}
	if r.Origins != nil {
		parts.Data["origins"] = r.Origins
	}
	return parts
}

type OriginGroupDelOriginGroupRequest struct {
	Headers map[string]string
	Ids     interface{} `json:"ids,omitempty"`
}

func NewOriginGroupDelOriginGroupRequest() *OriginGroupDelOriginGroupRequest {
	return &OriginGroupDelOriginGroupRequest{Headers: map[string]string{}}
}

func (r *OriginGroupDelOriginGroupRequest) APIName() string { return "OriginGroup_delOriginGroup" }
func (r *OriginGroupDelOriginGroupRequest) Method() string  { return "DELETE" }

func (r *OriginGroupDelOriginGroupRequest) SetHeader(key string, value string) *OriginGroupDelOriginGroupRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *OriginGroupDelOriginGroupRequest) RequestParts() ReqParams {
	if r == nil {
		r = &OriginGroupDelOriginGroupRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Ids != nil {
		parts.Data["ids"] = r.Ids
	}
	return parts
}

type OriginGroupBindOriginGroupToDomainsRequest struct {
	Headers        map[string]string
	OriginGroupId  interface{} `json:"origin_group_id,omitempty"`
	DomainIds      interface{} `json:"domain_ids,omitempty"`
	DomainGroupIds interface{} `json:"domain_group_ids,omitempty"`
	Domains        interface{} `json:"domains,omitempty"`
}

func NewOriginGroupBindOriginGroupToDomainsRequest() *OriginGroupBindOriginGroupToDomainsRequest {
	return &OriginGroupBindOriginGroupToDomainsRequest{Headers: map[string]string{}}
}

func (r *OriginGroupBindOriginGroupToDomainsRequest) APIName() string {
	return "OriginGroup_bindOriginGroupToDomains"
}
func (r *OriginGroupBindOriginGroupToDomainsRequest) Method() string { return "POST" }

func (r *OriginGroupBindOriginGroupToDomainsRequest) SetHeader(key string, value string) *OriginGroupBindOriginGroupToDomainsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *OriginGroupBindOriginGroupToDomainsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &OriginGroupBindOriginGroupToDomainsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.OriginGroupId != nil {
		parts.Data["origin_group_id"] = r.OriginGroupId
	}
	if r.DomainIds != nil {
		parts.Data["domain_ids"] = r.DomainIds
	}
	if r.DomainGroupIds != nil {
		parts.Data["domain_group_ids"] = r.DomainGroupIds
	}
	if r.Domains != nil {
		parts.Data["domains"] = r.Domains
	}
	return parts
}

type OriginGroupGetAllOriginGroupsRequest struct {
	Headers map[string]string
}

func NewOriginGroupGetAllOriginGroupsRequest() *OriginGroupGetAllOriginGroupsRequest {
	return &OriginGroupGetAllOriginGroupsRequest{Headers: map[string]string{}}
}

func (r *OriginGroupGetAllOriginGroupsRequest) APIName() string {
	return "OriginGroup_getAllOriginGroups"
}
func (r *OriginGroupGetAllOriginGroupsRequest) Method() string { return "GET" }

func (r *OriginGroupGetAllOriginGroupsRequest) SetHeader(key string, value string) *OriginGroupGetAllOriginGroupsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *OriginGroupGetAllOriginGroupsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &OriginGroupGetAllOriginGroupsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type OriginGroupCopyOriginGroupRequest struct {
	Headers       map[string]string
	OriginGroupId interface{} `json:"origin_group_id,omitempty"`
	DomainId      interface{} `json:"domain_id,omitempty"`
}

func NewOriginGroupCopyOriginGroupRequest() *OriginGroupCopyOriginGroupRequest {
	return &OriginGroupCopyOriginGroupRequest{Headers: map[string]string{}}
}

func (r *OriginGroupCopyOriginGroupRequest) APIName() string { return "OriginGroup_copyOriginGroup" }
func (r *OriginGroupCopyOriginGroupRequest) Method() string  { return "POST" }

func (r *OriginGroupCopyOriginGroupRequest) SetHeader(key string, value string) *OriginGroupCopyOriginGroupRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *OriginGroupCopyOriginGroupRequest) RequestParts() ReqParams {
	if r == nil {
		r = &OriginGroupCopyOriginGroupRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.OriginGroupId != nil {
		parts.Data["origin_group_id"] = r.OriginGroupId
	}
	if r.DomainId != nil {
		parts.Data["domain_id"] = r.DomainId
	}
	return parts
}

type FireWallReportGetBlockListRequest struct {
	Headers map[string]string
}

func NewFireWallReportGetBlockListRequest() *FireWallReportGetBlockListRequest {
	return &FireWallReportGetBlockListRequest{Headers: map[string]string{}}
}

func (r *FireWallReportGetBlockListRequest) APIName() string { return "FireWallReport_getBlockList" }
func (r *FireWallReportGetBlockListRequest) Method() string  { return "GET" }

func (r *FireWallReportGetBlockListRequest) SetHeader(key string, value string) *FireWallReportGetBlockListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *FireWallReportGetBlockListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &FireWallReportGetBlockListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type FireWallReportGetBlockDetailsRequest struct {
	Headers map[string]string
}

func NewFireWallReportGetBlockDetailsRequest() *FireWallReportGetBlockDetailsRequest {
	return &FireWallReportGetBlockDetailsRequest{Headers: map[string]string{}}
}

func (r *FireWallReportGetBlockDetailsRequest) APIName() string {
	return "FireWallReport_getBlockDetails"
}
func (r *FireWallReportGetBlockDetailsRequest) Method() string { return "GET" }

func (r *FireWallReportGetBlockDetailsRequest) SetHeader(key string, value string) *FireWallReportGetBlockDetailsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *FireWallReportGetBlockDetailsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &FireWallReportGetBlockDetailsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type FireWallReportGetPackageBlockListRequest struct {
	Headers map[string]string
}

func NewFireWallReportGetPackageBlockListRequest() *FireWallReportGetPackageBlockListRequest {
	return &FireWallReportGetPackageBlockListRequest{Headers: map[string]string{}}
}

func (r *FireWallReportGetPackageBlockListRequest) APIName() string {
	return "FireWallReport_getPackageBlockList"
}
func (r *FireWallReportGetPackageBlockListRequest) Method() string { return "GET" }

func (r *FireWallReportGetPackageBlockListRequest) SetHeader(key string, value string) *FireWallReportGetPackageBlockListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *FireWallReportGetPackageBlockListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &FireWallReportGetPackageBlockListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type FireWallReportGetPackageBlockDetailsRequest struct {
	Headers map[string]string
}

func NewFireWallReportGetPackageBlockDetailsRequest() *FireWallReportGetPackageBlockDetailsRequest {
	return &FireWallReportGetPackageBlockDetailsRequest{Headers: map[string]string{}}
}

func (r *FireWallReportGetPackageBlockDetailsRequest) APIName() string {
	return "FireWallReport_getPackageBlockDetails"
}
func (r *FireWallReportGetPackageBlockDetailsRequest) Method() string { return "GET" }

func (r *FireWallReportGetPackageBlockDetailsRequest) SetHeader(key string, value string) *FireWallReportGetPackageBlockDetailsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *FireWallReportGetPackageBlockDetailsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &FireWallReportGetPackageBlockDetailsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type CcQpsMaxRequest struct {
	Headers    map[string]string
	AcctId     interface{} `json:"acct_id,omitempty"`
	SubDomains interface{} `json:"sub_domains,omitempty"`
	StartTime  interface{} `json:"start_time,omitempty"`
	EndTime    interface{} `json:"end_time,omitempty"`
}

func NewCcQpsMaxRequest() *CcQpsMaxRequest {
	return &CcQpsMaxRequest{Headers: map[string]string{}}
}

func (r *CcQpsMaxRequest) APIName() string { return "cc_qps_max" }
func (r *CcQpsMaxRequest) Method() string  { return "POST" }

func (r *CcQpsMaxRequest) SetHeader(key string, value string) *CcQpsMaxRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CcQpsMaxRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CcQpsMaxRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.AcctId != nil {
		parts.Data["acct_id"] = r.AcctId
	}
	if r.SubDomains != nil {
		parts.Data["sub_domains"] = r.SubDomains
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	return parts
}

type CcAttackTimesRequest struct {
	Headers    map[string]string
	AcctId     interface{} `json:"acct_id,omitempty"`
	SubDomains interface{} `json:"sub_domains,omitempty"`
	StartTime  interface{} `json:"start_time,omitempty"`
	EndTime    interface{} `json:"end_time,omitempty"`
}

func NewCcAttackTimesRequest() *CcAttackTimesRequest {
	return &CcAttackTimesRequest{Headers: map[string]string{}}
}

func (r *CcAttackTimesRequest) APIName() string { return "cc_attack_times" }
func (r *CcAttackTimesRequest) Method() string  { return "POST" }

func (r *CcAttackTimesRequest) SetHeader(key string, value string) *CcAttackTimesRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CcAttackTimesRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CcAttackTimesRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.AcctId != nil {
		parts.Data["acct_id"] = r.AcctId
	}
	if r.SubDomains != nil {
		parts.Data["sub_domains"] = r.SubDomains
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	return parts
}

type CcTimesLineRequest struct {
	Headers    map[string]string
	AcctId     interface{} `json:"acct_id,omitempty"`
	SubDomains interface{} `json:"sub_domains,omitempty"`
	StartTime  interface{} `json:"start_time,omitempty"`
	EndTime    interface{} `json:"end_time,omitempty"`
	Interval   interface{} `json:"interval,omitempty"`
}

func NewCcTimesLineRequest() *CcTimesLineRequest {
	return &CcTimesLineRequest{Headers: map[string]string{}}
}

func (r *CcTimesLineRequest) APIName() string { return "cc_times_line" }
func (r *CcTimesLineRequest) Method() string  { return "POST" }

func (r *CcTimesLineRequest) SetHeader(key string, value string) *CcTimesLineRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CcTimesLineRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CcTimesLineRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.AcctId != nil {
		parts.Data["acct_id"] = r.AcctId
	}
	if r.SubDomains != nil {
		parts.Data["sub_domains"] = r.SubDomains
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	if r.Interval != nil {
		parts.Data["interval"] = r.Interval
	}
	return parts
}

type CcReportStatsRequest struct {
	Headers    map[string]string
	AcctId     interface{} `json:"acct_id,omitempty"`
	SubDomains interface{} `json:"sub_domains,omitempty"`
	StartTime  interface{} `json:"start_time,omitempty"`
	EndTime    interface{} `json:"end_time,omitempty"`
}

func NewCcReportStatsRequest() *CcReportStatsRequest {
	return &CcReportStatsRequest{Headers: map[string]string{}}
}

func (r *CcReportStatsRequest) APIName() string { return "cc_report_stats" }
func (r *CcReportStatsRequest) Method() string  { return "POST" }

func (r *CcReportStatsRequest) SetHeader(key string, value string) *CcReportStatsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CcReportStatsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CcReportStatsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.AcctId != nil {
		parts.Data["acct_id"] = r.AcctId
	}
	if r.SubDomains != nil {
		parts.Data["sub_domains"] = r.SubDomains
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	return parts
}

type CdnDomainUaispDistributeRequest struct {
	Headers    map[string]string
	AcctId     interface{} `json:"acct_id,omitempty"`
	SubDomains interface{} `json:"sub_domains,omitempty"`
	StartTime  interface{} `json:"start_time,omitempty"`
	EndTime    interface{} `json:"end_time,omitempty"`
}

func NewCdnDomainUaispDistributeRequest() *CdnDomainUaispDistributeRequest {
	return &CdnDomainUaispDistributeRequest{Headers: map[string]string{}}
}

func (r *CdnDomainUaispDistributeRequest) APIName() string { return "cdn_domain_uaisp_distribute" }
func (r *CdnDomainUaispDistributeRequest) Method() string  { return "POST" }

func (r *CdnDomainUaispDistributeRequest) SetHeader(key string, value string) *CdnDomainUaispDistributeRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CdnDomainUaispDistributeRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CdnDomainUaispDistributeRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.AcctId != nil {
		parts.Data["acct_id"] = r.AcctId
	}
	if r.SubDomains != nil {
		parts.Data["sub_domains"] = r.SubDomains
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	return parts
}

type CdnDomainCountryDistributeRequest struct {
	Headers    map[string]string
	AcctId     interface{} `json:"acct_id,omitempty"`
	SubDomains interface{} `json:"sub_domains,omitempty"`
	StartTime  interface{} `json:"start_time,omitempty"`
	EndTime    interface{} `json:"end_time,omitempty"`
}

func NewCdnDomainCountryDistributeRequest() *CdnDomainCountryDistributeRequest {
	return &CdnDomainCountryDistributeRequest{Headers: map[string]string{}}
}

func (r *CdnDomainCountryDistributeRequest) APIName() string { return "cdn_domain_country_distribute" }
func (r *CdnDomainCountryDistributeRequest) Method() string  { return "POST" }

func (r *CdnDomainCountryDistributeRequest) SetHeader(key string, value string) *CdnDomainCountryDistributeRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CdnDomainCountryDistributeRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CdnDomainCountryDistributeRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.AcctId != nil {
		parts.Data["acct_id"] = r.AcctId
	}
	if r.SubDomains != nil {
		parts.Data["sub_domains"] = r.SubDomains
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	return parts
}

type CdnDomainProvinceDistributeRequest struct {
	Headers    map[string]string
	AcctId     interface{} `json:"acct_id,omitempty"`
	SubDomains interface{} `json:"sub_domains,omitempty"`
	StartTime  interface{} `json:"start_time,omitempty"`
	EndTime    interface{} `json:"end_time,omitempty"`
}

func NewCdnDomainProvinceDistributeRequest() *CdnDomainProvinceDistributeRequest {
	return &CdnDomainProvinceDistributeRequest{Headers: map[string]string{}}
}

func (r *CdnDomainProvinceDistributeRequest) APIName() string {
	return "cdn_domain_province_distribute"
}
func (r *CdnDomainProvinceDistributeRequest) Method() string { return "POST" }

func (r *CdnDomainProvinceDistributeRequest) SetHeader(key string, value string) *CdnDomainProvinceDistributeRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CdnDomainProvinceDistributeRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CdnDomainProvinceDistributeRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.AcctId != nil {
		parts.Data["acct_id"] = r.AcctId
	}
	if r.SubDomains != nil {
		parts.Data["sub_domains"] = r.SubDomains
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	return parts
}

type CdnDomainStatusDistributeRequest struct {
	Headers    map[string]string
	AcctId     interface{} `json:"acct_id,omitempty"`
	SubDomains interface{} `json:"sub_domains,omitempty"`
	StartTime  interface{} `json:"start_time,omitempty"`
	EndTime    interface{} `json:"end_time,omitempty"`
}

func NewCdnDomainStatusDistributeRequest() *CdnDomainStatusDistributeRequest {
	return &CdnDomainStatusDistributeRequest{Headers: map[string]string{}}
}

func (r *CdnDomainStatusDistributeRequest) APIName() string { return "cdn_domain_status_distribute" }
func (r *CdnDomainStatusDistributeRequest) Method() string  { return "POST" }

func (r *CdnDomainStatusDistributeRequest) SetHeader(key string, value string) *CdnDomainStatusDistributeRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CdnDomainStatusDistributeRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CdnDomainStatusDistributeRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.AcctId != nil {
		parts.Data["acct_id"] = r.AcctId
	}
	if r.SubDomains != nil {
		parts.Data["sub_domains"] = r.SubDomains
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	return parts
}

type CdnDomainNodeFlowBandwidthRequest struct {
	Headers    map[string]string
	AcctId     interface{} `json:"acct_id,omitempty"`
	SubDomains interface{} `json:"sub_domains,omitempty"`
	StartTime  interface{} `json:"start_time,omitempty"`
	EndTime    interface{} `json:"end_time,omitempty"`
	Interval   interface{} `json:"interval,omitempty"`
}

func NewCdnDomainNodeFlowBandwidthRequest() *CdnDomainNodeFlowBandwidthRequest {
	return &CdnDomainNodeFlowBandwidthRequest{Headers: map[string]string{}}
}

func (r *CdnDomainNodeFlowBandwidthRequest) APIName() string { return "cdn_domain_node_flow_bandwidth" }
func (r *CdnDomainNodeFlowBandwidthRequest) Method() string  { return "POST" }

func (r *CdnDomainNodeFlowBandwidthRequest) SetHeader(key string, value string) *CdnDomainNodeFlowBandwidthRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CdnDomainNodeFlowBandwidthRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CdnDomainNodeFlowBandwidthRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.AcctId != nil {
		parts.Data["acct_id"] = r.AcctId
	}
	if r.SubDomains != nil {
		parts.Data["sub_domains"] = r.SubDomains
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	if r.Interval != nil {
		parts.Data["interval"] = r.Interval
	}
	return parts
}

type CdnDomainNodeFlowBandwidthCn2Request struct {
	Headers    map[string]string
	AcctId     interface{} `json:"acct_id,omitempty"`
	SubDomains interface{} `json:"sub_domains,omitempty"`
	StartTime  interface{} `json:"start_time,omitempty"`
	EndTime    interface{} `json:"end_time,omitempty"`
	Interval   interface{} `json:"interval,omitempty"`
}

func NewCdnDomainNodeFlowBandwidthCn2Request() *CdnDomainNodeFlowBandwidthCn2Request {
	return &CdnDomainNodeFlowBandwidthCn2Request{Headers: map[string]string{}}
}

func (r *CdnDomainNodeFlowBandwidthCn2Request) APIName() string {
	return "cdn_domain_node_flow_bandwidth_cn2"
}
func (r *CdnDomainNodeFlowBandwidthCn2Request) Method() string { return "POST" }

func (r *CdnDomainNodeFlowBandwidthCn2Request) SetHeader(key string, value string) *CdnDomainNodeFlowBandwidthCn2Request {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CdnDomainNodeFlowBandwidthCn2Request) RequestParts() ReqParams {
	if r == nil {
		r = &CdnDomainNodeFlowBandwidthCn2Request{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.AcctId != nil {
		parts.Data["acct_id"] = r.AcctId
	}
	if r.SubDomains != nil {
		parts.Data["sub_domains"] = r.SubDomains
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	if r.Interval != nil {
		parts.Data["interval"] = r.Interval
	}
	return parts
}

type CdnDomainNodeFlowBandwidthNodeRequest struct {
	Headers    map[string]string
	AcctId     interface{} `json:"acct_id,omitempty"`
	SubDomains interface{} `json:"sub_domains,omitempty"`
	StartTime  interface{} `json:"start_time,omitempty"`
	EndTime    interface{} `json:"end_time,omitempty"`
	Interval   interface{} `json:"interval,omitempty"`
}

func NewCdnDomainNodeFlowBandwidthNodeRequest() *CdnDomainNodeFlowBandwidthNodeRequest {
	return &CdnDomainNodeFlowBandwidthNodeRequest{Headers: map[string]string{}}
}

func (r *CdnDomainNodeFlowBandwidthNodeRequest) APIName() string {
	return "cdn_domain_node_flow_bandwidth_node"
}
func (r *CdnDomainNodeFlowBandwidthNodeRequest) Method() string { return "POST" }

func (r *CdnDomainNodeFlowBandwidthNodeRequest) SetHeader(key string, value string) *CdnDomainNodeFlowBandwidthNodeRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CdnDomainNodeFlowBandwidthNodeRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CdnDomainNodeFlowBandwidthNodeRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.AcctId != nil {
		parts.Data["acct_id"] = r.AcctId
	}
	if r.SubDomains != nil {
		parts.Data["sub_domains"] = r.SubDomains
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	if r.Interval != nil {
		parts.Data["interval"] = r.Interval
	}
	return parts
}

type DomainTimesRequest struct {
	Headers map[string]string
}

func NewDomainTimesRequest() *DomainTimesRequest {
	return &DomainTimesRequest{Headers: map[string]string{}}
}

func (r *DomainTimesRequest) APIName() string { return "domainTimes" }
func (r *DomainTimesRequest) Method() string  { return "POST" }

func (r *DomainTimesRequest) SetHeader(key string, value string) *DomainTimesRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DomainTimesRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DomainTimesRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type DomainQpsRequest struct {
	Headers    map[string]string
	AcctId     interface{} `json:"acct_id,omitempty"`
	SubDomains interface{} `json:"sub_domains,omitempty"`
	StartTime  interface{} `json:"start_time,omitempty"`
	EndTime    interface{} `json:"end_time,omitempty"`
	Interval   interface{} `json:"interval,omitempty"`
}

func NewDomainQpsRequest() *DomainQpsRequest {
	return &DomainQpsRequest{Headers: map[string]string{}}
}

func (r *DomainQpsRequest) APIName() string { return "domainQps" }
func (r *DomainQpsRequest) Method() string  { return "POST" }

func (r *DomainQpsRequest) SetHeader(key string, value string) *DomainQpsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DomainQpsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DomainQpsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.AcctId != nil {
		parts.Data["acct_id"] = r.AcctId
	}
	if r.SubDomains != nil {
		parts.Data["sub_domains"] = r.SubDomains
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	if r.Interval != nil {
		parts.Data["interval"] = r.Interval
	}
	return parts
}

type CdnDomainFlowLineRequest struct {
	Headers    map[string]string
	AcctId     interface{} `json:"acct_id,omitempty"`
	SubDomains interface{} `json:"sub_domains,omitempty"`
	StartTime  interface{} `json:"start_time,omitempty"`
	EndTime    interface{} `json:"end_time,omitempty"`
	Interval   interface{} `json:"interval,omitempty"`
	NodeType   interface{} `json:"node_type,omitempty"`
}

func NewCdnDomainFlowLineRequest() *CdnDomainFlowLineRequest {
	return &CdnDomainFlowLineRequest{Headers: map[string]string{}}
}

func (r *CdnDomainFlowLineRequest) APIName() string { return "cdn_domain_flow_line" }
func (r *CdnDomainFlowLineRequest) Method() string  { return "POST" }

func (r *CdnDomainFlowLineRequest) SetHeader(key string, value string) *CdnDomainFlowLineRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CdnDomainFlowLineRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CdnDomainFlowLineRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.AcctId != nil {
		parts.Data["acct_id"] = r.AcctId
	}
	if r.SubDomains != nil {
		parts.Data["sub_domains"] = r.SubDomains
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	if r.Interval != nil {
		parts.Data["interval"] = r.Interval
	}
	if r.NodeType != nil {
		parts.Data["node_type"] = r.NodeType
	}
	return parts
}

type CdnDomainBandwidthLineRequest struct {
	Headers    map[string]string
	AcctId     interface{} `json:"acct_id,omitempty"`
	SubDomains interface{} `json:"sub_domains,omitempty"`
	StartTime  interface{} `json:"start_time,omitempty"`
	EndTime    interface{} `json:"end_time,omitempty"`
	Interval   interface{} `json:"interval,omitempty"`
	NodeType   interface{} `json:"node_type,omitempty"`
}

func NewCdnDomainBandwidthLineRequest() *CdnDomainBandwidthLineRequest {
	return &CdnDomainBandwidthLineRequest{Headers: map[string]string{}}
}

func (r *CdnDomainBandwidthLineRequest) APIName() string { return "cdn_domain_bandwidth_line" }
func (r *CdnDomainBandwidthLineRequest) Method() string  { return "POST" }

func (r *CdnDomainBandwidthLineRequest) SetHeader(key string, value string) *CdnDomainBandwidthLineRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CdnDomainBandwidthLineRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CdnDomainBandwidthLineRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.AcctId != nil {
		parts.Data["acct_id"] = r.AcctId
	}
	if r.SubDomains != nil {
		parts.Data["sub_domains"] = r.SubDomains
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	if r.Interval != nil {
		parts.Data["interval"] = r.Interval
	}
	if r.NodeType != nil {
		parts.Data["node_type"] = r.NodeType
	}
	return parts
}

type CdnDomainBandwidth95Request struct {
	Headers    map[string]string
	AcctId     interface{} `json:"acct_id,omitempty"`
	SubDomains interface{} `json:"sub_domains,omitempty"`
	StartTime  interface{} `json:"start_time,omitempty"`
	EndTime    interface{} `json:"end_time,omitempty"`
	NodeType   interface{} `json:"node_type,omitempty"`
}

func NewCdnDomainBandwidth95Request() *CdnDomainBandwidth95Request {
	return &CdnDomainBandwidth95Request{Headers: map[string]string{}}
}

func (r *CdnDomainBandwidth95Request) APIName() string { return "cdn_domain_bandwidth_95" }
func (r *CdnDomainBandwidth95Request) Method() string  { return "POST" }

func (r *CdnDomainBandwidth95Request) SetHeader(key string, value string) *CdnDomainBandwidth95Request {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CdnDomainBandwidth95Request) RequestParts() ReqParams {
	if r == nil {
		r = &CdnDomainBandwidth95Request{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.AcctId != nil {
		parts.Data["acct_id"] = r.AcctId
	}
	if r.SubDomains != nil {
		parts.Data["sub_domains"] = r.SubDomains
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	if r.NodeType != nil {
		parts.Data["node_type"] = r.NodeType
	}
	return parts
}

type CdnDomainPvtimesRequest struct {
	Headers    map[string]string
	AcctId     interface{} `json:"acct_id,omitempty"`
	SubDomains interface{} `json:"sub_domains,omitempty"`
	StartTime  interface{} `json:"start_time,omitempty"`
	EndTime    interface{} `json:"end_time,omitempty"`
}

func NewCdnDomainPvtimesRequest() *CdnDomainPvtimesRequest {
	return &CdnDomainPvtimesRequest{Headers: map[string]string{}}
}

func (r *CdnDomainPvtimesRequest) APIName() string { return "cdn_domain_pvtimes" }
func (r *CdnDomainPvtimesRequest) Method() string  { return "POST" }

func (r *CdnDomainPvtimesRequest) SetHeader(key string, value string) *CdnDomainPvtimesRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CdnDomainPvtimesRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CdnDomainPvtimesRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.AcctId != nil {
		parts.Data["acct_id"] = r.AcctId
	}
	if r.SubDomains != nil {
		parts.Data["sub_domains"] = r.SubDomains
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	return parts
}

type CdnDomainFlowTopRequest struct {
	Headers    map[string]string
	AcctId     interface{} `json:"acct_id,omitempty"`
	TopSize    interface{} `json:"top_size,omitempty"`
	SubDomains interface{} `json:"sub_domains,omitempty"`
	StartTime  interface{} `json:"start_time,omitempty"`
	EndTime    interface{} `json:"end_time,omitempty"`
}

func NewCdnDomainFlowTopRequest() *CdnDomainFlowTopRequest {
	return &CdnDomainFlowTopRequest{Headers: map[string]string{}}
}

func (r *CdnDomainFlowTopRequest) APIName() string { return "cdn_domain_flow_top" }
func (r *CdnDomainFlowTopRequest) Method() string  { return "POST" }

func (r *CdnDomainFlowTopRequest) SetHeader(key string, value string) *CdnDomainFlowTopRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CdnDomainFlowTopRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CdnDomainFlowTopRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.AcctId != nil {
		parts.Data["acct_id"] = r.AcctId
	}
	if r.TopSize != nil {
		parts.Data["top_size"] = r.TopSize
	}
	if r.SubDomains != nil {
		parts.Data["sub_domains"] = r.SubDomains
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	return parts
}

type CdnDomainBandwidthTopRequest struct {
	Headers    map[string]string
	AcctId     interface{} `json:"acct_id,omitempty"`
	TopSize    interface{} `json:"top_size,omitempty"`
	SubDomains interface{} `json:"sub_domains,omitempty"`
	StartTime  interface{} `json:"start_time,omitempty"`
	EndTime    interface{} `json:"end_time,omitempty"`
}

func NewCdnDomainBandwidthTopRequest() *CdnDomainBandwidthTopRequest {
	return &CdnDomainBandwidthTopRequest{Headers: map[string]string{}}
}

func (r *CdnDomainBandwidthTopRequest) APIName() string { return "cdn_domain_bandwidth_top" }
func (r *CdnDomainBandwidthTopRequest) Method() string  { return "POST" }

func (r *CdnDomainBandwidthTopRequest) SetHeader(key string, value string) *CdnDomainBandwidthTopRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CdnDomainBandwidthTopRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CdnDomainBandwidthTopRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.AcctId != nil {
		parts.Data["acct_id"] = r.AcctId
	}
	if r.TopSize != nil {
		parts.Data["top_size"] = r.TopSize
	}
	if r.SubDomains != nil {
		parts.Data["sub_domains"] = r.SubDomains
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	return parts
}

type CdnDomainTimesTopRequest struct {
	Headers    map[string]string
	AcctId     interface{} `json:"acct_id,omitempty"`
	TopSize    interface{} `json:"top_size,omitempty"`
	SubDomains interface{} `json:"sub_domains,omitempty"`
	StartTime  interface{} `json:"start_time,omitempty"`
	EndTime    interface{} `json:"end_time,omitempty"`
}

func NewCdnDomainTimesTopRequest() *CdnDomainTimesTopRequest {
	return &CdnDomainTimesTopRequest{Headers: map[string]string{}}
}

func (r *CdnDomainTimesTopRequest) APIName() string { return "cdn_domain_times_top" }
func (r *CdnDomainTimesTopRequest) Method() string  { return "POST" }

func (r *CdnDomainTimesTopRequest) SetHeader(key string, value string) *CdnDomainTimesTopRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CdnDomainTimesTopRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CdnDomainTimesTopRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.AcctId != nil {
		parts.Data["acct_id"] = r.AcctId
	}
	if r.TopSize != nil {
		parts.Data["top_size"] = r.TopSize
	}
	if r.SubDomains != nil {
		parts.Data["sub_domains"] = r.SubDomains
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	return parts
}

type CdnDomainTimesTopEsRequest struct {
	Headers     map[string]string
	AcctId      interface{} `json:"acct_id,omitempty"`
	TopSize     interface{} `json:"top_size,omitempty"`
	SubDomains  interface{} `json:"sub_domains,omitempty"`
	StartTime   interface{} `json:"start_time,omitempty"`
	EndTime     interface{} `json:"end_time,omitempty"`
	HttpReferer interface{} `json:"http_referer,omitempty"`
}

func NewCdnDomainTimesTopEsRequest() *CdnDomainTimesTopEsRequest {
	return &CdnDomainTimesTopEsRequest{Headers: map[string]string{}}
}

func (r *CdnDomainTimesTopEsRequest) APIName() string { return "cdn_domain_times_top_es" }
func (r *CdnDomainTimesTopEsRequest) Method() string  { return "POST" }

func (r *CdnDomainTimesTopEsRequest) SetHeader(key string, value string) *CdnDomainTimesTopEsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CdnDomainTimesTopEsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CdnDomainTimesTopEsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.AcctId != nil {
		parts.Data["acct_id"] = r.AcctId
	}
	if r.TopSize != nil {
		parts.Data["top_size"] = r.TopSize
	}
	if r.SubDomains != nil {
		parts.Data["sub_domains"] = r.SubDomains
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	if r.HttpReferer != nil {
		parts.Data["http_referer"] = r.HttpReferer
	}
	return parts
}

type CdnDomainUrlTopRequest struct {
	Headers     map[string]string
	AcctId      interface{} `json:"acct_id,omitempty"`
	TopSize     interface{} `json:"top_size,omitempty"`
	HttpReferer interface{} `json:"http_referer,omitempty"`
	SubDomains  interface{} `json:"sub_domains,omitempty"`
	StartTime   interface{} `json:"start_time,omitempty"`
	EndTime     interface{} `json:"end_time,omitempty"`
}

func NewCdnDomainUrlTopRequest() *CdnDomainUrlTopRequest {
	return &CdnDomainUrlTopRequest{Headers: map[string]string{}}
}

func (r *CdnDomainUrlTopRequest) APIName() string { return "cdn_domain_url_top" }
func (r *CdnDomainUrlTopRequest) Method() string  { return "POST" }

func (r *CdnDomainUrlTopRequest) SetHeader(key string, value string) *CdnDomainUrlTopRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CdnDomainUrlTopRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CdnDomainUrlTopRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.AcctId != nil {
		parts.Data["acct_id"] = r.AcctId
	}
	if r.TopSize != nil {
		parts.Data["top_size"] = r.TopSize
	}
	if r.HttpReferer != nil {
		parts.Data["http_referer"] = r.HttpReferer
	}
	if r.SubDomains != nil {
		parts.Data["sub_domains"] = r.SubDomains
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	return parts
}

type CdnDomainRefererTopRequest struct {
	Headers    map[string]string
	AcctId     interface{} `json:"acct_id,omitempty"`
	TopSize    interface{} `json:"top_size,omitempty"`
	SubDomains interface{} `json:"sub_domains,omitempty"`
	StartTime  interface{} `json:"start_time,omitempty"`
	EndTime    interface{} `json:"end_time,omitempty"`
}

func NewCdnDomainRefererTopRequest() *CdnDomainRefererTopRequest {
	return &CdnDomainRefererTopRequest{Headers: map[string]string{}}
}

func (r *CdnDomainRefererTopRequest) APIName() string { return "cdn_domain_referer_top" }
func (r *CdnDomainRefererTopRequest) Method() string  { return "POST" }

func (r *CdnDomainRefererTopRequest) SetHeader(key string, value string) *CdnDomainRefererTopRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CdnDomainRefererTopRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CdnDomainRefererTopRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.AcctId != nil {
		parts.Data["acct_id"] = r.AcctId
	}
	if r.TopSize != nil {
		parts.Data["top_size"] = r.TopSize
	}
	if r.SubDomains != nil {
		parts.Data["sub_domains"] = r.SubDomains
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	return parts
}

type CdnDomainStatusTopDownloadRequest struct {
	Headers     map[string]string
	SubDomains  interface{} `json:"sub_domains,omitempty"`
	GroupId     interface{} `json:"group_id,omitempty"`
	ResourceIds interface{} `json:"resource_ids,omitempty"`
	StartTime   interface{} `json:"start_time,omitempty"`
	EndTime     interface{} `json:"end_time,omitempty"`
	TimeZone    interface{} `json:"time_zone,omitempty"`
}

func NewCdnDomainStatusTopDownloadRequest() *CdnDomainStatusTopDownloadRequest {
	return &CdnDomainStatusTopDownloadRequest{Headers: map[string]string{}}
}

func (r *CdnDomainStatusTopDownloadRequest) APIName() string { return "cdn_domain_status_top_download" }
func (r *CdnDomainStatusTopDownloadRequest) Method() string  { return "POST" }

func (r *CdnDomainStatusTopDownloadRequest) SetHeader(key string, value string) *CdnDomainStatusTopDownloadRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CdnDomainStatusTopDownloadRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CdnDomainStatusTopDownloadRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.SubDomains != nil {
		parts.Data["sub_domains"] = r.SubDomains
	}
	if r.GroupId != nil {
		parts.Data["group_id"] = r.GroupId
	}
	if r.ResourceIds != nil {
		parts.Data["resource_ids"] = r.ResourceIds
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	if r.TimeZone != nil {
		parts.Data["time_zone"] = r.TimeZone
	}
	return parts
}

type CdnDomainBandwidthDownloadRequest struct {
	Headers     map[string]string
	SubDomains  interface{} `json:"sub_domains,omitempty"`
	GroupId     interface{} `json:"group_id,omitempty"`
	ResourceIds interface{} `json:"resource_ids,omitempty"`
	StartTime   interface{} `json:"start_time,omitempty"`
	EndTime     interface{} `json:"end_time,omitempty"`
	TimeZone    interface{} `json:"time_zone,omitempty"`
}

func NewCdnDomainBandwidthDownloadRequest() *CdnDomainBandwidthDownloadRequest {
	return &CdnDomainBandwidthDownloadRequest{Headers: map[string]string{}}
}

func (r *CdnDomainBandwidthDownloadRequest) APIName() string { return "cdn_domain_bandwidth_download" }
func (r *CdnDomainBandwidthDownloadRequest) Method() string  { return "POST" }

func (r *CdnDomainBandwidthDownloadRequest) SetHeader(key string, value string) *CdnDomainBandwidthDownloadRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CdnDomainBandwidthDownloadRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CdnDomainBandwidthDownloadRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.SubDomains != nil {
		parts.Data["sub_domains"] = r.SubDomains
	}
	if r.GroupId != nil {
		parts.Data["group_id"] = r.GroupId
	}
	if r.ResourceIds != nil {
		parts.Data["resource_ids"] = r.ResourceIds
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	if r.TimeZone != nil {
		parts.Data["time_zone"] = r.TimeZone
	}
	return parts
}

type CdnDomainFlowDownloadRequest struct {
	Headers     map[string]string
	SubDomains  interface{} `json:"sub_domains,omitempty"`
	GroupId     interface{} `json:"group_id,omitempty"`
	ResourceIds interface{} `json:"resource_ids,omitempty"`
	StartTime   interface{} `json:"start_time,omitempty"`
	EndTime     interface{} `json:"end_time,omitempty"`
	TimeZone    interface{} `json:"time_zone,omitempty"`
}

func NewCdnDomainFlowDownloadRequest() *CdnDomainFlowDownloadRequest {
	return &CdnDomainFlowDownloadRequest{Headers: map[string]string{}}
}

func (r *CdnDomainFlowDownloadRequest) APIName() string { return "cdn_domain_flow_download" }
func (r *CdnDomainFlowDownloadRequest) Method() string  { return "POST" }

func (r *CdnDomainFlowDownloadRequest) SetHeader(key string, value string) *CdnDomainFlowDownloadRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CdnDomainFlowDownloadRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CdnDomainFlowDownloadRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.SubDomains != nil {
		parts.Data["sub_domains"] = r.SubDomains
	}
	if r.GroupId != nil {
		parts.Data["group_id"] = r.GroupId
	}
	if r.ResourceIds != nil {
		parts.Data["resource_ids"] = r.ResourceIds
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	if r.TimeZone != nil {
		parts.Data["time_zone"] = r.TimeZone
	}
	return parts
}

type TcpBandwidthRequest struct {
	Headers    map[string]string
	PackageIds interface{} `json:"package_ids,omitempty"`
	StartTime  interface{} `json:"start_time,omitempty"`
	EndTime    interface{} `json:"end_time,omitempty"`
}

func NewTcpBandwidthRequest() *TcpBandwidthRequest {
	return &TcpBandwidthRequest{Headers: map[string]string{}}
}

func (r *TcpBandwidthRequest) APIName() string { return "tcp_bandwidth" }
func (r *TcpBandwidthRequest) Method() string  { return "POST" }

func (r *TcpBandwidthRequest) SetHeader(key string, value string) *TcpBandwidthRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *TcpBandwidthRequest) RequestParts() ReqParams {
	if r == nil {
		r = &TcpBandwidthRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.PackageIds != nil {
		parts.Data["package_ids"] = r.PackageIds
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	return parts
}

type TcpCcFlawRequest struct {
	Headers   map[string]string
	PackageId interface{} `json:"package_id,omitempty"`
	Ip        interface{} `json:"ip,omitempty"`
	Port      interface{} `json:"port,omitempty"`
	StartTime interface{} `json:"start_time,omitempty"`
	EndTime   interface{} `json:"end_time,omitempty"`
	Interval  interface{} `json:"interval,omitempty"`
}

func NewTcpCcFlawRequest() *TcpCcFlawRequest {
	return &TcpCcFlawRequest{Headers: map[string]string{}}
}

func (r *TcpCcFlawRequest) APIName() string { return "tcp_cc_flaw" }
func (r *TcpCcFlawRequest) Method() string  { return "POST" }

func (r *TcpCcFlawRequest) SetHeader(key string, value string) *TcpCcFlawRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *TcpCcFlawRequest) RequestParts() ReqParams {
	if r == nil {
		r = &TcpCcFlawRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.PackageId != nil {
		parts.Data["package_id"] = r.PackageId
	}
	if r.Ip != nil {
		parts.Data["ip"] = r.Ip
	}
	if r.Port != nil {
		parts.Data["port"] = r.Port
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	if r.Interval != nil {
		parts.Data["interval"] = r.Interval
	}
	return parts
}

type WafAttackTimesRequest struct {
	Headers    map[string]string
	AcctId     interface{} `json:"acct_id,omitempty"`
	SubDomains interface{} `json:"sub_domains,omitempty"`
	StartTime  interface{} `json:"start_time,omitempty"`
	EndTime    interface{} `json:"end_time,omitempty"`
}

func NewWafAttackTimesRequest() *WafAttackTimesRequest {
	return &WafAttackTimesRequest{Headers: map[string]string{}}
}

func (r *WafAttackTimesRequest) APIName() string { return "waf_attack_times" }
func (r *WafAttackTimesRequest) Method() string  { return "POST" }

func (r *WafAttackTimesRequest) SetHeader(key string, value string) *WafAttackTimesRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *WafAttackTimesRequest) RequestParts() ReqParams {
	if r == nil {
		r = &WafAttackTimesRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.AcctId != nil {
		parts.Data["acct_id"] = r.AcctId
	}
	if r.SubDomains != nil {
		parts.Data["sub_domains"] = r.SubDomains
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	return parts
}

type WafReportStatsRequest struct {
	Headers    map[string]string
	AcctId     interface{} `json:"acct_id,omitempty"`
	SubDomains interface{} `json:"sub_domains,omitempty"`
	StartTime  interface{} `json:"start_time,omitempty"`
	EndTime    interface{} `json:"end_time,omitempty"`
}

func NewWafReportStatsRequest() *WafReportStatsRequest {
	return &WafReportStatsRequest{Headers: map[string]string{}}
}

func (r *WafReportStatsRequest) APIName() string { return "waf_report_stats" }
func (r *WafReportStatsRequest) Method() string  { return "POST" }

func (r *WafReportStatsRequest) SetHeader(key string, value string) *WafReportStatsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *WafReportStatsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &WafReportStatsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.AcctId != nil {
		parts.Data["acct_id"] = r.AcctId
	}
	if r.SubDomains != nil {
		parts.Data["sub_domains"] = r.SubDomains
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	return parts
}

type WafWebshellEventListRequest struct {
	Headers    map[string]string
	AcctId     interface{} `json:"acct_id,omitempty"`
	SubDomains interface{} `json:"sub_domains,omitempty"`
	StartTime  interface{} `json:"start_time,omitempty"`
	EndTime    interface{} `json:"end_time,omitempty"`
	Page       interface{} `json:"page,omitempty"`
	PerPage    interface{} `json:"per_page,omitempty"`
}

func NewWafWebshellEventListRequest() *WafWebshellEventListRequest {
	return &WafWebshellEventListRequest{Headers: map[string]string{}}
}

func (r *WafWebshellEventListRequest) APIName() string { return "waf_webshell_event_list" }
func (r *WafWebshellEventListRequest) Method() string  { return "POST" }

func (r *WafWebshellEventListRequest) SetHeader(key string, value string) *WafWebshellEventListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *WafWebshellEventListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &WafWebshellEventListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.AcctId != nil {
		parts.Data["acct_id"] = r.AcctId
	}
	if r.SubDomains != nil {
		parts.Data["sub_domains"] = r.SubDomains
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	if r.Page != nil {
		parts.Data["page"] = r.Page
	}
	if r.PerPage != nil {
		parts.Data["per_page"] = r.PerPage
	}
	return parts
}

type WafWebshellEventDetailRequest struct {
	Headers    map[string]string
	RemoteAddr interface{} `json:"remote_addr,omitempty"`
	RequestUrl interface{} `json:"request_url,omitempty"`
	StartTime  interface{} `json:"start_time,omitempty"`
	EndTime    interface{} `json:"end_time,omitempty"`
	Page       interface{} `json:"page,omitempty"`
	PerPage    interface{} `json:"per_page,omitempty"`
}

func NewWafWebshellEventDetailRequest() *WafWebshellEventDetailRequest {
	return &WafWebshellEventDetailRequest{Headers: map[string]string{}}
}

func (r *WafWebshellEventDetailRequest) APIName() string { return "waf_webshell_event_detail" }
func (r *WafWebshellEventDetailRequest) Method() string  { return "POST" }

func (r *WafWebshellEventDetailRequest) SetHeader(key string, value string) *WafWebshellEventDetailRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *WafWebshellEventDetailRequest) RequestParts() ReqParams {
	if r == nil {
		r = &WafWebshellEventDetailRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.RemoteAddr != nil {
		parts.Data["remote_addr"] = r.RemoteAddr
	}
	if r.RequestUrl != nil {
		parts.Data["request_url"] = r.RequestUrl
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	if r.Page != nil {
		parts.Data["page"] = r.Page
	}
	if r.PerPage != nil {
		parts.Data["per_page"] = r.PerPage
	}
	return parts
}

type WafAttackEventListRequest struct {
	Headers    map[string]string
	AcctId     interface{} `json:"acct_id,omitempty"`
	SubDomains interface{} `json:"sub_domains,omitempty"`
	StartTime  interface{} `json:"start_time,omitempty"`
	EndTime    interface{} `json:"end_time,omitempty"`
	Page       interface{} `json:"page,omitempty"`
	PerPage    interface{} `json:"per_page,omitempty"`
}

func NewWafAttackEventListRequest() *WafAttackEventListRequest {
	return &WafAttackEventListRequest{Headers: map[string]string{}}
}

func (r *WafAttackEventListRequest) APIName() string { return "waf_attack_event_list" }
func (r *WafAttackEventListRequest) Method() string  { return "POST" }

func (r *WafAttackEventListRequest) SetHeader(key string, value string) *WafAttackEventListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *WafAttackEventListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &WafAttackEventListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.AcctId != nil {
		parts.Data["acct_id"] = r.AcctId
	}
	if r.SubDomains != nil {
		parts.Data["sub_domains"] = r.SubDomains
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	if r.Page != nil {
		parts.Data["page"] = r.Page
	}
	if r.PerPage != nil {
		parts.Data["per_page"] = r.PerPage
	}
	return parts
}

type WafAttackEventDetailRequest struct {
	Headers    map[string]string
	RemoteAddr interface{} `json:"remote_addr,omitempty"`
	HttpHost   interface{} `json:"http_host,omitempty"`
	StartTime  interface{} `json:"start_time,omitempty"`
	EndTime    interface{} `json:"end_time,omitempty"`
	Page       interface{} `json:"page,omitempty"`
	PerPage    interface{} `json:"per_page,omitempty"`
}

func NewWafAttackEventDetailRequest() *WafAttackEventDetailRequest {
	return &WafAttackEventDetailRequest{Headers: map[string]string{}}
}

func (r *WafAttackEventDetailRequest) APIName() string { return "waf_attack_event_detail" }
func (r *WafAttackEventDetailRequest) Method() string  { return "POST" }

func (r *WafAttackEventDetailRequest) SetHeader(key string, value string) *WafAttackEventDetailRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *WafAttackEventDetailRequest) RequestParts() ReqParams {
	if r == nil {
		r = &WafAttackEventDetailRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.RemoteAddr != nil {
		parts.Data["remote_addr"] = r.RemoteAddr
	}
	if r.HttpHost != nil {
		parts.Data["http_host"] = r.HttpHost
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	if r.Page != nil {
		parts.Data["page"] = r.Page
	}
	if r.PerPage != nil {
		parts.Data["per_page"] = r.PerPage
	}
	return parts
}

type WafScanEventListRequest struct {
	Headers    map[string]string
	AcctId     interface{} `json:"acct_id,omitempty"`
	SubDomains interface{} `json:"sub_domains,omitempty"`
	StartTime  interface{} `json:"start_time,omitempty"`
	EndTime    interface{} `json:"end_time,omitempty"`
	Page       interface{} `json:"page,omitempty"`
	PerPage    interface{} `json:"per_page,omitempty"`
}

func NewWafScanEventListRequest() *WafScanEventListRequest {
	return &WafScanEventListRequest{Headers: map[string]string{}}
}

func (r *WafScanEventListRequest) APIName() string { return "waf_scan_event_list" }
func (r *WafScanEventListRequest) Method() string  { return "POST" }

func (r *WafScanEventListRequest) SetHeader(key string, value string) *WafScanEventListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *WafScanEventListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &WafScanEventListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.AcctId != nil {
		parts.Data["acct_id"] = r.AcctId
	}
	if r.SubDomains != nil {
		parts.Data["sub_domains"] = r.SubDomains
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	if r.Page != nil {
		parts.Data["page"] = r.Page
	}
	if r.PerPage != nil {
		parts.Data["per_page"] = r.PerPage
	}
	return parts
}

type WafScanEventDetailRequest struct {
	Headers    map[string]string
	RemoteAddr interface{} `json:"remote_addr,omitempty"`
	HttpHost   interface{} `json:"http_host,omitempty"`
	StartTime  interface{} `json:"start_time,omitempty"`
	EndTime    interface{} `json:"end_time,omitempty"`
	Page       interface{} `json:"page,omitempty"`
	PerPage    interface{} `json:"per_page,omitempty"`
}

func NewWafScanEventDetailRequest() *WafScanEventDetailRequest {
	return &WafScanEventDetailRequest{Headers: map[string]string{}}
}

func (r *WafScanEventDetailRequest) APIName() string { return "waf_scan_event_detail" }
func (r *WafScanEventDetailRequest) Method() string  { return "POST" }

func (r *WafScanEventDetailRequest) SetHeader(key string, value string) *WafScanEventDetailRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *WafScanEventDetailRequest) RequestParts() ReqParams {
	if r == nil {
		r = &WafScanEventDetailRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.RemoteAddr != nil {
		parts.Data["remote_addr"] = r.RemoteAddr
	}
	if r.HttpHost != nil {
		parts.Data["http_host"] = r.HttpHost
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	if r.Page != nil {
		parts.Data["page"] = r.Page
	}
	if r.PerPage != nil {
		parts.Data["per_page"] = r.PerPage
	}
	return parts
}

type WafTypeLineRequest struct {
	Headers    map[string]string
	AcctId     interface{} `json:"acct_id,omitempty"`
	SubDomains interface{} `json:"sub_domains,omitempty"`
	StartTime  interface{} `json:"start_time,omitempty"`
	EndTime    interface{} `json:"end_time,omitempty"`
	Interval   interface{} `json:"interval,omitempty"`
}

func NewWafTypeLineRequest() *WafTypeLineRequest {
	return &WafTypeLineRequest{Headers: map[string]string{}}
}

func (r *WafTypeLineRequest) APIName() string { return "waf_type_line" }
func (r *WafTypeLineRequest) Method() string  { return "POST" }

func (r *WafTypeLineRequest) SetHeader(key string, value string) *WafTypeLineRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *WafTypeLineRequest) RequestParts() ReqParams {
	if r == nil {
		r = &WafTypeLineRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.AcctId != nil {
		parts.Data["acct_id"] = r.AcctId
	}
	if r.SubDomains != nil {
		parts.Data["sub_domains"] = r.SubDomains
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	if r.Interval != nil {
		parts.Data["interval"] = r.Interval
	}
	return parts
}

type LogDownloadTaskTaskListRequest struct {
	Headers map[string]string
}

func NewLogDownloadTaskTaskListRequest() *LogDownloadTaskTaskListRequest {
	return &LogDownloadTaskTaskListRequest{Headers: map[string]string{}}
}

func (r *LogDownloadTaskTaskListRequest) APIName() string { return "LogDownloadTask_taskList" }
func (r *LogDownloadTaskTaskListRequest) Method() string  { return "POST" }

func (r *LogDownloadTaskTaskListRequest) SetHeader(key string, value string) *LogDownloadTaskTaskListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *LogDownloadTaskTaskListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &LogDownloadTaskTaskListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type LogDownloadTaskAddTaskRequest struct {
	Headers        map[string]string
	TaskName       interface{} `json:"task_name,omitempty"`
	IsUseTemplate  interface{} `json:"is_use_template,omitempty"`
	TemplateId     interface{} `json:"template_id,omitempty"`
	DataSource     interface{} `json:"data_source,omitempty"`
	DownloadFields interface{} `json:"download_fields,omitempty"`
	SearchTerms    interface{} `json:"search_terms,omitempty"`
	FileType       interface{} `json:"file_type,omitempty"`
	StartTime      interface{} `json:"start_time,omitempty"`
	EndTime        interface{} `json:"end_time,omitempty"`
	Lang           interface{} `json:"lang,omitempty"`
}

func NewLogDownloadTaskAddTaskRequest() *LogDownloadTaskAddTaskRequest {
	return &LogDownloadTaskAddTaskRequest{Headers: map[string]string{}}
}

func (r *LogDownloadTaskAddTaskRequest) APIName() string { return "LogDownloadTask_addTask" }
func (r *LogDownloadTaskAddTaskRequest) Method() string  { return "POST" }

func (r *LogDownloadTaskAddTaskRequest) SetHeader(key string, value string) *LogDownloadTaskAddTaskRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *LogDownloadTaskAddTaskRequest) RequestParts() ReqParams {
	if r == nil {
		r = &LogDownloadTaskAddTaskRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.TaskName != nil {
		parts.Data["task_name"] = r.TaskName
	}
	if r.IsUseTemplate != nil {
		parts.Data["is_use_template"] = r.IsUseTemplate
	}
	if r.TemplateId != nil {
		parts.Data["template_id"] = r.TemplateId
	}
	if r.DataSource != nil {
		parts.Data["data_source"] = r.DataSource
	}
	if r.DownloadFields != nil {
		parts.Data["download_fields"] = r.DownloadFields
	}
	if r.SearchTerms != nil {
		parts.Data["search_terms"] = r.SearchTerms
	}
	if r.FileType != nil {
		parts.Data["file_type"] = r.FileType
	}
	if r.StartTime != nil {
		parts.Data["start_time"] = r.StartTime
	}
	if r.EndTime != nil {
		parts.Data["end_time"] = r.EndTime
	}
	parts.Data["lang"] = "zh_CN"
	if r.Lang != nil {
		parts.Data["lang"] = r.Lang
	}
	return parts
}

type LogDownloadTaskCancelTaskRequest struct {
	Headers map[string]string
	TaskId  interface{} `json:"task_id,omitempty"`
}

func NewLogDownloadTaskCancelTaskRequest() *LogDownloadTaskCancelTaskRequest {
	return &LogDownloadTaskCancelTaskRequest{Headers: map[string]string{}}
}

func (r *LogDownloadTaskCancelTaskRequest) APIName() string { return "LogDownloadTask_cancelTask" }
func (r *LogDownloadTaskCancelTaskRequest) Method() string  { return "POST" }

func (r *LogDownloadTaskCancelTaskRequest) SetHeader(key string, value string) *LogDownloadTaskCancelTaskRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *LogDownloadTaskCancelTaskRequest) RequestParts() ReqParams {
	if r == nil {
		r = &LogDownloadTaskCancelTaskRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.TaskId != nil {
		parts.Data["task_id"] = r.TaskId
	}
	return parts
}

type LogDownloadTaskBatchCancelTaskRequest struct {
	Headers map[string]string
	TaskIds interface{} `json:"task_ids,omitempty"`
}

func NewLogDownloadTaskBatchCancelTaskRequest() *LogDownloadTaskBatchCancelTaskRequest {
	return &LogDownloadTaskBatchCancelTaskRequest{Headers: map[string]string{}}
}

func (r *LogDownloadTaskBatchCancelTaskRequest) APIName() string {
	return "LogDownloadTask_batchCancelTask"
}
func (r *LogDownloadTaskBatchCancelTaskRequest) Method() string { return "DELETE" }

func (r *LogDownloadTaskBatchCancelTaskRequest) SetHeader(key string, value string) *LogDownloadTaskBatchCancelTaskRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *LogDownloadTaskBatchCancelTaskRequest) RequestParts() ReqParams {
	if r == nil {
		r = &LogDownloadTaskBatchCancelTaskRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.TaskIds != nil {
		parts.Data["task_ids"] = r.TaskIds
	}
	return parts
}

type LogDownloadTaskDeleteTaskRequest struct {
	Headers map[string]string
	TaskId  interface{} `json:"task_id,omitempty"`
}

func NewLogDownloadTaskDeleteTaskRequest() *LogDownloadTaskDeleteTaskRequest {
	return &LogDownloadTaskDeleteTaskRequest{Headers: map[string]string{}}
}

func (r *LogDownloadTaskDeleteTaskRequest) APIName() string { return "LogDownloadTask_deleteTask" }
func (r *LogDownloadTaskDeleteTaskRequest) Method() string  { return "DELETE" }

func (r *LogDownloadTaskDeleteTaskRequest) SetHeader(key string, value string) *LogDownloadTaskDeleteTaskRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *LogDownloadTaskDeleteTaskRequest) RequestParts() ReqParams {
	if r == nil {
		r = &LogDownloadTaskDeleteTaskRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.TaskId != nil {
		parts.Data["task_id"] = r.TaskId
	}
	return parts
}

type LogDownloadTaskBatchDeleteTaskRequest struct {
	Headers map[string]string
	TaskIds interface{} `json:"task_ids,omitempty"`
}

func NewLogDownloadTaskBatchDeleteTaskRequest() *LogDownloadTaskBatchDeleteTaskRequest {
	return &LogDownloadTaskBatchDeleteTaskRequest{Headers: map[string]string{}}
}

func (r *LogDownloadTaskBatchDeleteTaskRequest) APIName() string {
	return "LogDownloadTask_batchDeleteTask"
}
func (r *LogDownloadTaskBatchDeleteTaskRequest) Method() string { return "DELETE" }

func (r *LogDownloadTaskBatchDeleteTaskRequest) SetHeader(key string, value string) *LogDownloadTaskBatchDeleteTaskRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *LogDownloadTaskBatchDeleteTaskRequest) RequestParts() ReqParams {
	if r == nil {
		r = &LogDownloadTaskBatchDeleteTaskRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.TaskIds != nil {
		parts.Data["task_ids"] = r.TaskIds
	}
	return parts
}

type LogDownloadTaskRegenerateTaskRequest struct {
	Headers map[string]string
	TaskId  interface{} `json:"task_id,omitempty"`
}

func NewLogDownloadTaskRegenerateTaskRequest() *LogDownloadTaskRegenerateTaskRequest {
	return &LogDownloadTaskRegenerateTaskRequest{Headers: map[string]string{}}
}

func (r *LogDownloadTaskRegenerateTaskRequest) APIName() string {
	return "LogDownloadTask_regenerateTask"
}
func (r *LogDownloadTaskRegenerateTaskRequest) Method() string { return "POST" }

func (r *LogDownloadTaskRegenerateTaskRequest) SetHeader(key string, value string) *LogDownloadTaskRegenerateTaskRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *LogDownloadTaskRegenerateTaskRequest) RequestParts() ReqParams {
	if r == nil {
		r = &LogDownloadTaskRegenerateTaskRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.TaskId != nil {
		parts.Data["task_id"] = r.TaskId
	}
	return parts
}

type LogDownloadFieldConfDownloadFieldsRequest struct {
	Headers map[string]string
}

func NewLogDownloadFieldConfDownloadFieldsRequest() *LogDownloadFieldConfDownloadFieldsRequest {
	return &LogDownloadFieldConfDownloadFieldsRequest{Headers: map[string]string{}}
}

func (r *LogDownloadFieldConfDownloadFieldsRequest) APIName() string {
	return "LogDownloadFieldConf_downloadFields"
}
func (r *LogDownloadFieldConfDownloadFieldsRequest) Method() string { return "GET" }

func (r *LogDownloadFieldConfDownloadFieldsRequest) SetHeader(key string, value string) *LogDownloadFieldConfDownloadFieldsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *LogDownloadFieldConfDownloadFieldsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &LogDownloadFieldConfDownloadFieldsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type LogDownloadTemplateTemplateListRequest struct {
	Headers map[string]string
}

func NewLogDownloadTemplateTemplateListRequest() *LogDownloadTemplateTemplateListRequest {
	return &LogDownloadTemplateTemplateListRequest{Headers: map[string]string{}}
}

func (r *LogDownloadTemplateTemplateListRequest) APIName() string {
	return "LogDownloadTemplate_templateList"
}
func (r *LogDownloadTemplateTemplateListRequest) Method() string { return "POST" }

func (r *LogDownloadTemplateTemplateListRequest) SetHeader(key string, value string) *LogDownloadTemplateTemplateListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *LogDownloadTemplateTemplateListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &LogDownloadTemplateTemplateListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type LogDownloadTemplateGetTemplateDomainListRequest struct {
	Headers map[string]string
}

func NewLogDownloadTemplateGetTemplateDomainListRequest() *LogDownloadTemplateGetTemplateDomainListRequest {
	return &LogDownloadTemplateGetTemplateDomainListRequest{Headers: map[string]string{}}
}

func (r *LogDownloadTemplateGetTemplateDomainListRequest) APIName() string {
	return "LogDownloadTemplate_getTemplateDomainList"
}
func (r *LogDownloadTemplateGetTemplateDomainListRequest) Method() string { return "GET" }

func (r *LogDownloadTemplateGetTemplateDomainListRequest) SetHeader(key string, value string) *LogDownloadTemplateGetTemplateDomainListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *LogDownloadTemplateGetTemplateDomainListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &LogDownloadTemplateGetTemplateDomainListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type LogDownloadTemplateAddTemplateRequest struct {
	Headers          map[string]string
	TemplateName     interface{} `json:"template_name,omitempty"`
	GroupName        interface{} `json:"group_name,omitempty"`
	GroupId          interface{} `json:"group_id,omitempty"`
	DataSource       interface{} `json:"data_source,omitempty"`
	Status           interface{} `json:"status,omitempty"`
	DownloadFields   interface{} `json:"download_fields,omitempty"`
	SearchTerms      interface{} `json:"search_terms,omitempty"`
	DomainSelectType interface{} `json:"domain_select_type,omitempty"`
}

func NewLogDownloadTemplateAddTemplateRequest() *LogDownloadTemplateAddTemplateRequest {
	return &LogDownloadTemplateAddTemplateRequest{Headers: map[string]string{}}
}

func (r *LogDownloadTemplateAddTemplateRequest) APIName() string {
	return "LogDownloadTemplate_addTemplate"
}
func (r *LogDownloadTemplateAddTemplateRequest) Method() string { return "POST" }

func (r *LogDownloadTemplateAddTemplateRequest) SetHeader(key string, value string) *LogDownloadTemplateAddTemplateRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *LogDownloadTemplateAddTemplateRequest) RequestParts() ReqParams {
	if r == nil {
		r = &LogDownloadTemplateAddTemplateRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.TemplateName != nil {
		parts.Data["template_name"] = r.TemplateName
	}
	if r.GroupName != nil {
		parts.Data["group_name"] = r.GroupName
	}
	if r.GroupId != nil {
		parts.Data["group_id"] = r.GroupId
	}
	if r.DataSource != nil {
		parts.Data["data_source"] = r.DataSource
	}
	if r.Status != nil {
		parts.Data["status"] = r.Status
	}
	if r.DownloadFields != nil {
		parts.Data["download_fields"] = r.DownloadFields
	}
	if r.SearchTerms != nil {
		parts.Data["search_terms"] = r.SearchTerms
	}
	if r.DomainSelectType != nil {
		parts.Data["domain_select_type"] = r.DomainSelectType
	}
	return parts
}

type LogDownloadTemplateSaveTemplateRequest struct {
	Headers        map[string]string
	TemplateId     interface{} `json:"template_id,omitempty"`
	TemplateName   interface{} `json:"template_name,omitempty"`
	GroupName      interface{} `json:"group_name,omitempty"`
	GroupId        interface{} `json:"group_id,omitempty"`
	DataSource     interface{} `json:"data_source,omitempty"`
	Status         interface{} `json:"status,omitempty"`
	DownloadFields interface{} `json:"download_fields,omitempty"`
	SearchTerms    interface{} `json:"search_terms,omitempty"`
}

func NewLogDownloadTemplateSaveTemplateRequest() *LogDownloadTemplateSaveTemplateRequest {
	return &LogDownloadTemplateSaveTemplateRequest{Headers: map[string]string{}}
}

func (r *LogDownloadTemplateSaveTemplateRequest) APIName() string {
	return "LogDownloadTemplate_saveTemplate"
}
func (r *LogDownloadTemplateSaveTemplateRequest) Method() string { return "POST" }

func (r *LogDownloadTemplateSaveTemplateRequest) SetHeader(key string, value string) *LogDownloadTemplateSaveTemplateRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *LogDownloadTemplateSaveTemplateRequest) RequestParts() ReqParams {
	if r == nil {
		r = &LogDownloadTemplateSaveTemplateRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.TemplateId != nil {
		parts.Data["template_id"] = r.TemplateId
	}
	if r.TemplateName != nil {
		parts.Data["template_name"] = r.TemplateName
	}
	if r.GroupName != nil {
		parts.Data["group_name"] = r.GroupName
	}
	if r.GroupId != nil {
		parts.Data["group_id"] = r.GroupId
	}
	if r.DataSource != nil {
		parts.Data["data_source"] = r.DataSource
	}
	parts.Data["status"] = "1"
	if r.Status != nil {
		parts.Data["status"] = r.Status
	}
	if r.DownloadFields != nil {
		parts.Data["download_fields"] = r.DownloadFields
	}
	if r.SearchTerms != nil {
		parts.Data["search_terms"] = r.SearchTerms
	}
	return parts
}

type LogDownloadTemplateDelTemplateRequest struct {
	Headers    map[string]string
	TemplateId interface{} `json:"template_id,omitempty"`
}

func NewLogDownloadTemplateDelTemplateRequest() *LogDownloadTemplateDelTemplateRequest {
	return &LogDownloadTemplateDelTemplateRequest{Headers: map[string]string{}}
}

func (r *LogDownloadTemplateDelTemplateRequest) APIName() string {
	return "LogDownloadTemplate_delTemplate"
}
func (r *LogDownloadTemplateDelTemplateRequest) Method() string { return "DELETE" }

func (r *LogDownloadTemplateDelTemplateRequest) SetHeader(key string, value string) *LogDownloadTemplateDelTemplateRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *LogDownloadTemplateDelTemplateRequest) RequestParts() ReqParams {
	if r == nil {
		r = &LogDownloadTemplateDelTemplateRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.TemplateId != nil {
		parts.Data["template_id"] = r.TemplateId
	}
	return parts
}

type LogDownloadTemplateBatchDelTemplateRequest struct {
	Headers     map[string]string
	TemplateIds interface{} `json:"template_ids,omitempty"`
}

func NewLogDownloadTemplateBatchDelTemplateRequest() *LogDownloadTemplateBatchDelTemplateRequest {
	return &LogDownloadTemplateBatchDelTemplateRequest{Headers: map[string]string{}}
}

func (r *LogDownloadTemplateBatchDelTemplateRequest) APIName() string {
	return "LogDownloadTemplate_batchDelTemplate"
}
func (r *LogDownloadTemplateBatchDelTemplateRequest) Method() string { return "DELETE" }

func (r *LogDownloadTemplateBatchDelTemplateRequest) SetHeader(key string, value string) *LogDownloadTemplateBatchDelTemplateRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *LogDownloadTemplateBatchDelTemplateRequest) RequestParts() ReqParams {
	if r == nil {
		r = &LogDownloadTemplateBatchDelTemplateRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.TemplateIds != nil {
		parts.Data["template_ids"] = r.TemplateIds
	}
	return parts
}

type LogDownloadTemplateChangeStatusRequest struct {
	Headers    map[string]string
	TemplateId interface{} `json:"template_id,omitempty"`
	Status     interface{} `json:"status,omitempty"`
}

func NewLogDownloadTemplateChangeStatusRequest() *LogDownloadTemplateChangeStatusRequest {
	return &LogDownloadTemplateChangeStatusRequest{Headers: map[string]string{}}
}

func (r *LogDownloadTemplateChangeStatusRequest) APIName() string {
	return "LogDownloadTemplate_changeStatus"
}
func (r *LogDownloadTemplateChangeStatusRequest) Method() string { return "POST" }

func (r *LogDownloadTemplateChangeStatusRequest) SetHeader(key string, value string) *LogDownloadTemplateChangeStatusRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *LogDownloadTemplateChangeStatusRequest) RequestParts() ReqParams {
	if r == nil {
		r = &LogDownloadTemplateChangeStatusRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.TemplateId != nil {
		parts.Data["template_id"] = r.TemplateId
	}
	if r.Status != nil {
		parts.Data["status"] = r.Status
	}
	return parts
}

type LogDownloadTemplateBatchChangeStatusRequest struct {
	Headers     map[string]string
	TemplateIds interface{} `json:"template_ids,omitempty"`
	Status      interface{} `json:"status,omitempty"`
}

func NewLogDownloadTemplateBatchChangeStatusRequest() *LogDownloadTemplateBatchChangeStatusRequest {
	return &LogDownloadTemplateBatchChangeStatusRequest{Headers: map[string]string{}}
}

func (r *LogDownloadTemplateBatchChangeStatusRequest) APIName() string {
	return "LogDownloadTemplate_batchChangeStatus"
}
func (r *LogDownloadTemplateBatchChangeStatusRequest) Method() string { return "POST" }

func (r *LogDownloadTemplateBatchChangeStatusRequest) SetHeader(key string, value string) *LogDownloadTemplateBatchChangeStatusRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *LogDownloadTemplateBatchChangeStatusRequest) RequestParts() ReqParams {
	if r == nil {
		r = &LogDownloadTemplateBatchChangeStatusRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.TemplateIds != nil {
		parts.Data["template_ids"] = r.TemplateIds
	}
	if r.Status != nil {
		parts.Data["status"] = r.Status
	}
	return parts
}

type LogDownloadTemplateAllTemplateRequest struct {
	Headers map[string]string
}

func NewLogDownloadTemplateAllTemplateRequest() *LogDownloadTemplateAllTemplateRequest {
	return &LogDownloadTemplateAllTemplateRequest{Headers: map[string]string{}}
}

func (r *LogDownloadTemplateAllTemplateRequest) APIName() string {
	return "LogDownloadTemplate_allTemplate"
}
func (r *LogDownloadTemplateAllTemplateRequest) Method() string { return "POST" }

func (r *LogDownloadTemplateAllTemplateRequest) SetHeader(key string, value string) *LogDownloadTemplateAllTemplateRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *LogDownloadTemplateAllTemplateRequest) RequestParts() ReqParams {
	if r == nil {
		r = &LogDownloadTemplateAllTemplateRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type LogDownloadTemplateAllTemplateGroupRequest struct {
	Headers map[string]string
}

func NewLogDownloadTemplateAllTemplateGroupRequest() *LogDownloadTemplateAllTemplateGroupRequest {
	return &LogDownloadTemplateAllTemplateGroupRequest{Headers: map[string]string{}}
}

func (r *LogDownloadTemplateAllTemplateGroupRequest) APIName() string {
	return "LogDownloadTemplate_allTemplateGroup"
}
func (r *LogDownloadTemplateAllTemplateGroupRequest) Method() string { return "POST" }

func (r *LogDownloadTemplateAllTemplateGroupRequest) SetHeader(key string, value string) *LogDownloadTemplateAllTemplateGroupRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *LogDownloadTemplateAllTemplateGroupRequest) RequestParts() ReqParams {
	if r == nil {
		r = &LogDownloadTemplateAllTemplateGroupRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type TjkdPlusPackageGetMemberPackageListRequest struct {
	Headers map[string]string
}

func NewTjkdPlusPackageGetMemberPackageListRequest() *TjkdPlusPackageGetMemberPackageListRequest {
	return &TjkdPlusPackageGetMemberPackageListRequest{Headers: map[string]string{}}
}

func (r *TjkdPlusPackageGetMemberPackageListRequest) APIName() string {
	return "TjkdPlusPackage_getMemberPackageList"
}
func (r *TjkdPlusPackageGetMemberPackageListRequest) Method() string { return "GET" }

func (r *TjkdPlusPackageGetMemberPackageListRequest) SetHeader(key string, value string) *TjkdPlusPackageGetMemberPackageListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *TjkdPlusPackageGetMemberPackageListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &TjkdPlusPackageGetMemberPackageListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type TjkdPlusPackageGetAllPackageRequest struct {
	Headers map[string]string
}

func NewTjkdPlusPackageGetAllPackageRequest() *TjkdPlusPackageGetAllPackageRequest {
	return &TjkdPlusPackageGetAllPackageRequest{Headers: map[string]string{}}
}

func (r *TjkdPlusPackageGetAllPackageRequest) APIName() string {
	return "TjkdPlusPackage_getAllPackage"
}
func (r *TjkdPlusPackageGetAllPackageRequest) Method() string { return "GET" }

func (r *TjkdPlusPackageGetAllPackageRequest) SetHeader(key string, value string) *TjkdPlusPackageGetAllPackageRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *TjkdPlusPackageGetAllPackageRequest) RequestParts() ReqParams {
	if r == nil {
		r = &TjkdPlusPackageGetAllPackageRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type TjkdPlusPackageGetPackageInfoRequest struct {
	Headers map[string]string
}

func NewTjkdPlusPackageGetPackageInfoRequest() *TjkdPlusPackageGetPackageInfoRequest {
	return &TjkdPlusPackageGetPackageInfoRequest{Headers: map[string]string{}}
}

func (r *TjkdPlusPackageGetPackageInfoRequest) APIName() string {
	return "TjkdPlusPackage_getPackageInfo"
}
func (r *TjkdPlusPackageGetPackageInfoRequest) Method() string { return "GET" }

func (r *TjkdPlusPackageGetPackageInfoRequest) SetHeader(key string, value string) *TjkdPlusPackageGetPackageInfoRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *TjkdPlusPackageGetPackageInfoRequest) RequestParts() ReqParams {
	if r == nil {
		r = &TjkdPlusPackageGetPackageInfoRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type TjkdPlusPackageGetPackageIpListRequest struct {
	Headers map[string]string
}

func NewTjkdPlusPackageGetPackageIpListRequest() *TjkdPlusPackageGetPackageIpListRequest {
	return &TjkdPlusPackageGetPackageIpListRequest{Headers: map[string]string{}}
}

func (r *TjkdPlusPackageGetPackageIpListRequest) APIName() string {
	return "TjkdPlusPackage_getPackageIpList"
}
func (r *TjkdPlusPackageGetPackageIpListRequest) Method() string { return "GET" }

func (r *TjkdPlusPackageGetPackageIpListRequest) SetHeader(key string, value string) *TjkdPlusPackageGetPackageIpListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *TjkdPlusPackageGetPackageIpListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &TjkdPlusPackageGetPackageIpListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type TjkdPlusPackageGetPackageOverviewRequest struct {
	Headers map[string]string
}

func NewTjkdPlusPackageGetPackageOverviewRequest() *TjkdPlusPackageGetPackageOverviewRequest {
	return &TjkdPlusPackageGetPackageOverviewRequest{Headers: map[string]string{}}
}

func (r *TjkdPlusPackageGetPackageOverviewRequest) APIName() string {
	return "TjkdPlusPackage_getPackageOverview"
}
func (r *TjkdPlusPackageGetPackageOverviewRequest) Method() string { return "GET" }

func (r *TjkdPlusPackageGetPackageOverviewRequest) SetHeader(key string, value string) *TjkdPlusPackageGetPackageOverviewRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *TjkdPlusPackageGetPackageOverviewRequest) RequestParts() ReqParams {
	if r == nil {
		r = &TjkdPlusPackageGetPackageOverviewRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type TjkdPlusPackageGetPackagePortListRequest struct {
	Headers map[string]string
}

func NewTjkdPlusPackageGetPackagePortListRequest() *TjkdPlusPackageGetPackagePortListRequest {
	return &TjkdPlusPackageGetPackagePortListRequest{Headers: map[string]string{}}
}

func (r *TjkdPlusPackageGetPackagePortListRequest) APIName() string {
	return "TjkdPlusPackage_getPackagePortList"
}
func (r *TjkdPlusPackageGetPackagePortListRequest) Method() string { return "GET" }

func (r *TjkdPlusPackageGetPackagePortListRequest) SetHeader(key string, value string) *TjkdPlusPackageGetPackagePortListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *TjkdPlusPackageGetPackagePortListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &TjkdPlusPackageGetPackagePortListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type TjkdPlusPackageSavePackageRequest struct {
	Headers map[string]string
}

func NewTjkdPlusPackageSavePackageRequest() *TjkdPlusPackageSavePackageRequest {
	return &TjkdPlusPackageSavePackageRequest{Headers: map[string]string{}}
}

func (r *TjkdPlusPackageSavePackageRequest) APIName() string { return "TjkdPlusPackage_savePackage" }
func (r *TjkdPlusPackageSavePackageRequest) Method() string  { return "POST" }

func (r *TjkdPlusPackageSavePackageRequest) SetHeader(key string, value string) *TjkdPlusPackageSavePackageRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *TjkdPlusPackageSavePackageRequest) RequestParts() ReqParams {
	if r == nil {
		r = &TjkdPlusPackageSavePackageRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type TjkdPlusPackageSavePackageHealthyConfRequest struct {
	Headers        map[string]string
	PackageId      interface{} `json:"package_id,omitempty"`
	FailsTimeout   interface{} `json:"fails_timeout,omitempty"`
	MaxFails       interface{} `json:"max_fails,omitempty"`
	KeepNewSrcTime interface{} `json:"keep_new_src_time,omitempty"`
}

func NewTjkdPlusPackageSavePackageHealthyConfRequest() *TjkdPlusPackageSavePackageHealthyConfRequest {
	return &TjkdPlusPackageSavePackageHealthyConfRequest{Headers: map[string]string{}}
}

func (r *TjkdPlusPackageSavePackageHealthyConfRequest) APIName() string {
	return "TjkdPlusPackage_savePackageHealthyConf"
}
func (r *TjkdPlusPackageSavePackageHealthyConfRequest) Method() string { return "POST" }

func (r *TjkdPlusPackageSavePackageHealthyConfRequest) SetHeader(key string, value string) *TjkdPlusPackageSavePackageHealthyConfRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *TjkdPlusPackageSavePackageHealthyConfRequest) RequestParts() ReqParams {
	if r == nil {
		r = &TjkdPlusPackageSavePackageHealthyConfRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.PackageId != nil {
		parts.Data["package_id"] = r.PackageId
	}
	if r.FailsTimeout != nil {
		parts.Data["fails_timeout"] = r.FailsTimeout
	}
	if r.MaxFails != nil {
		parts.Data["max_fails"] = r.MaxFails
	}
	if r.KeepNewSrcTime != nil {
		parts.Data["keep_new_src_time"] = r.KeepNewSrcTime
	}
	return parts
}

type TjkdPlusForwardRuleSavePlusForwardRuleRequest struct {
	Headers         map[string]string
	PackageId       interface{} `json:"package_id,omitempty"`
	Protocol        interface{} `json:"protocol,omitempty"`
	ProtocolPort    interface{} `json:"protocol_port,omitempty"`
	Loading         interface{} `json:"loading,omitempty"`
	SourceIp        interface{} `json:"source_ip,omitempty"`
	SourcePort      interface{} `json:"source_port,omitempty"`
	Backup          interface{} `json:"backup,omitempty"`
	SourceType      interface{} `json:"source_type,omitempty"`
	Actions         interface{} `json:"actions,omitempty"`
	Remark          interface{} `json:"remark,omitempty"`
	ProtocolPortOld interface{} `json:"protocol_port_old,omitempty"`
}

func NewTjkdPlusForwardRuleSavePlusForwardRuleRequest() *TjkdPlusForwardRuleSavePlusForwardRuleRequest {
	return &TjkdPlusForwardRuleSavePlusForwardRuleRequest{Headers: map[string]string{}}
}

func (r *TjkdPlusForwardRuleSavePlusForwardRuleRequest) APIName() string {
	return "TjkdPlusForwardRule_savePlusForwardRule"
}
func (r *TjkdPlusForwardRuleSavePlusForwardRuleRequest) Method() string { return "POST" }

func (r *TjkdPlusForwardRuleSavePlusForwardRuleRequest) SetHeader(key string, value string) *TjkdPlusForwardRuleSavePlusForwardRuleRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *TjkdPlusForwardRuleSavePlusForwardRuleRequest) RequestParts() ReqParams {
	if r == nil {
		r = &TjkdPlusForwardRuleSavePlusForwardRuleRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.PackageId != nil {
		parts.Data["package_id"] = r.PackageId
	}
	if r.Protocol != nil {
		parts.Data["protocol"] = r.Protocol
	}
	if r.ProtocolPort != nil {
		parts.Data["protocol_port"] = r.ProtocolPort
	}
	if r.Loading != nil {
		parts.Data["loading"] = r.Loading
	}
	if r.SourceIp != nil {
		parts.Data["source_ip"] = r.SourceIp
	}
	if r.SourcePort != nil {
		parts.Data["source_port"] = r.SourcePort
	}
	if r.Backup != nil {
		parts.Data["backup"] = r.Backup
	}
	if r.SourceType != nil {
		parts.Data["source_type"] = r.SourceType
	}
	if r.Actions != nil {
		parts.Data["actions"] = r.Actions
	}
	if r.Remark != nil {
		parts.Data["remark"] = r.Remark
	}
	if r.ProtocolPortOld != nil {
		parts.Data["protocol_port_old"] = r.ProtocolPortOld
	}
	return parts
}

type TjkdPlusForwardRuleBatchAddPlusForwardRuleRequest struct {
	Headers      map[string]string
	PackageId    interface{} `json:"package_id,omitempty"`
	Protocol     interface{} `json:"protocol,omitempty"`
	SourceType   interface{} `json:"source_type,omitempty"`
	ProtocolPort interface{} `json:"protocol_port,omitempty"`
	Loading      interface{} `json:"loading,omitempty"`
	SourceIp     interface{} `json:"source_ip,omitempty"`
	Backup       interface{} `json:"backup,omitempty"`
	Remark       interface{} `json:"remark,omitempty"`
}

func NewTjkdPlusForwardRuleBatchAddPlusForwardRuleRequest() *TjkdPlusForwardRuleBatchAddPlusForwardRuleRequest {
	return &TjkdPlusForwardRuleBatchAddPlusForwardRuleRequest{Headers: map[string]string{}}
}

func (r *TjkdPlusForwardRuleBatchAddPlusForwardRuleRequest) APIName() string {
	return "TjkdPlusForwardRule_batchAddPlusForwardRule"
}
func (r *TjkdPlusForwardRuleBatchAddPlusForwardRuleRequest) Method() string { return "POST" }

func (r *TjkdPlusForwardRuleBatchAddPlusForwardRuleRequest) SetHeader(key string, value string) *TjkdPlusForwardRuleBatchAddPlusForwardRuleRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *TjkdPlusForwardRuleBatchAddPlusForwardRuleRequest) RequestParts() ReqParams {
	if r == nil {
		r = &TjkdPlusForwardRuleBatchAddPlusForwardRuleRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.PackageId != nil {
		parts.Data["package_id"] = r.PackageId
	}
	if r.Protocol != nil {
		parts.Data["protocol"] = r.Protocol
	}
	if r.SourceType != nil {
		parts.Data["source_type"] = r.SourceType
	}
	if r.ProtocolPort != nil {
		parts.Data["protocol_port"] = r.ProtocolPort
	}
	if r.Loading != nil {
		parts.Data["loading"] = r.Loading
	}
	if r.SourceIp != nil {
		parts.Data["source_ip"] = r.SourceIp
	}
	if r.Backup != nil {
		parts.Data["backup"] = r.Backup
	}
	if r.Remark != nil {
		parts.Data["remark"] = r.Remark
	}
	return parts
}

type TjkdPlusForwardRuleBatchSavePlusForwardRuleRequest struct {
	Headers         map[string]string
	PackageId       interface{} `json:"package_id,omitempty"`
	Protocol        interface{} `json:"protocol,omitempty"`
	SourceType      interface{} `json:"source_type,omitempty"`
	ProtocolPort    interface{} `json:"protocol_port,omitempty"`
	Loading         interface{} `json:"loading,omitempty"`
	SourceIp        interface{} `json:"source_ip,omitempty"`
	Backup          interface{} `json:"backup,omitempty"`
	ProtocolPortOld interface{} `json:"protocol_port_old,omitempty"`
	Remark          interface{} `json:"remark,omitempty"`
}

func NewTjkdPlusForwardRuleBatchSavePlusForwardRuleRequest() *TjkdPlusForwardRuleBatchSavePlusForwardRuleRequest {
	return &TjkdPlusForwardRuleBatchSavePlusForwardRuleRequest{Headers: map[string]string{}}
}

func (r *TjkdPlusForwardRuleBatchSavePlusForwardRuleRequest) APIName() string {
	return "TjkdPlusForwardRule_batchSavePlusForwardRule"
}
func (r *TjkdPlusForwardRuleBatchSavePlusForwardRuleRequest) Method() string { return "POST" }

func (r *TjkdPlusForwardRuleBatchSavePlusForwardRuleRequest) SetHeader(key string, value string) *TjkdPlusForwardRuleBatchSavePlusForwardRuleRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *TjkdPlusForwardRuleBatchSavePlusForwardRuleRequest) RequestParts() ReqParams {
	if r == nil {
		r = &TjkdPlusForwardRuleBatchSavePlusForwardRuleRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.PackageId != nil {
		parts.Data["package_id"] = r.PackageId
	}
	if r.Protocol != nil {
		parts.Data["protocol"] = r.Protocol
	}
	if r.SourceType != nil {
		parts.Data["source_type"] = r.SourceType
	}
	if r.ProtocolPort != nil {
		parts.Data["protocol_port"] = r.ProtocolPort
	}
	if r.Loading != nil {
		parts.Data["loading"] = r.Loading
	}
	if r.SourceIp != nil {
		parts.Data["source_ip"] = r.SourceIp
	}
	if r.Backup != nil {
		parts.Data["backup"] = r.Backup
	}
	if r.ProtocolPortOld != nil {
		parts.Data["protocol_port_old"] = r.ProtocolPortOld
	}
	if r.Remark != nil {
		parts.Data["remark"] = r.Remark
	}
	return parts
}

type TjkdPlusForwardRuleDelPlusForwardRuleRequest struct {
	Headers map[string]string
	Ids     interface{} `json:"ids,omitempty"`
}

func NewTjkdPlusForwardRuleDelPlusForwardRuleRequest() *TjkdPlusForwardRuleDelPlusForwardRuleRequest {
	return &TjkdPlusForwardRuleDelPlusForwardRuleRequest{Headers: map[string]string{}}
}

func (r *TjkdPlusForwardRuleDelPlusForwardRuleRequest) APIName() string {
	return "TjkdPlusForwardRule_delPlusForwardRule"
}
func (r *TjkdPlusForwardRuleDelPlusForwardRuleRequest) Method() string { return "DELETE" }

func (r *TjkdPlusForwardRuleDelPlusForwardRuleRequest) SetHeader(key string, value string) *TjkdPlusForwardRuleDelPlusForwardRuleRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *TjkdPlusForwardRuleDelPlusForwardRuleRequest) RequestParts() ReqParams {
	if r == nil {
		r = &TjkdPlusForwardRuleDelPlusForwardRuleRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Ids != nil {
		parts.Data["ids"] = r.Ids
	}
	return parts
}

type TjkdPlusForwardRuleGetPlusForwardRuleListRequest struct {
	Headers map[string]string
}

func NewTjkdPlusForwardRuleGetPlusForwardRuleListRequest() *TjkdPlusForwardRuleGetPlusForwardRuleListRequest {
	return &TjkdPlusForwardRuleGetPlusForwardRuleListRequest{Headers: map[string]string{}}
}

func (r *TjkdPlusForwardRuleGetPlusForwardRuleListRequest) APIName() string {
	return "TjkdPlusForwardRule_getPlusForwardRuleList"
}
func (r *TjkdPlusForwardRuleGetPlusForwardRuleListRequest) Method() string { return "GET" }

func (r *TjkdPlusForwardRuleGetPlusForwardRuleListRequest) SetHeader(key string, value string) *TjkdPlusForwardRuleGetPlusForwardRuleListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *TjkdPlusForwardRuleGetPlusForwardRuleListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &TjkdPlusForwardRuleGetPlusForwardRuleListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type TjkdPlusForwardRuleGetBatchPlusForwardRuleInfoRequest struct {
	Headers map[string]string
	Ids     interface{} `json:"ids,omitempty"`
}

func NewTjkdPlusForwardRuleGetBatchPlusForwardRuleInfoRequest() *TjkdPlusForwardRuleGetBatchPlusForwardRuleInfoRequest {
	return &TjkdPlusForwardRuleGetBatchPlusForwardRuleInfoRequest{Headers: map[string]string{}}
}

func (r *TjkdPlusForwardRuleGetBatchPlusForwardRuleInfoRequest) APIName() string {
	return "TjkdPlusForwardRule_getBatchPlusForwardRuleInfo"
}
func (r *TjkdPlusForwardRuleGetBatchPlusForwardRuleInfoRequest) Method() string { return "POST" }

func (r *TjkdPlusForwardRuleGetBatchPlusForwardRuleInfoRequest) SetHeader(key string, value string) *TjkdPlusForwardRuleGetBatchPlusForwardRuleInfoRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *TjkdPlusForwardRuleGetBatchPlusForwardRuleInfoRequest) RequestParts() ReqParams {
	if r == nil {
		r = &TjkdPlusForwardRuleGetBatchPlusForwardRuleInfoRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Ids != nil {
		parts.Data["ids"] = r.Ids
	}
	return parts
}

type TjkdPlusPackageGetPackageDomainListRequest struct {
	Headers map[string]string
}

func NewTjkdPlusPackageGetPackageDomainListRequest() *TjkdPlusPackageGetPackageDomainListRequest {
	return &TjkdPlusPackageGetPackageDomainListRequest{Headers: map[string]string{}}
}

func (r *TjkdPlusPackageGetPackageDomainListRequest) APIName() string {
	return "TjkdPlusPackage_getPackageDomainList"
}
func (r *TjkdPlusPackageGetPackageDomainListRequest) Method() string { return "GET" }

func (r *TjkdPlusPackageGetPackageDomainListRequest) SetHeader(key string, value string) *TjkdPlusPackageGetPackageDomainListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *TjkdPlusPackageGetPackageDomainListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &TjkdPlusPackageGetPackageDomainListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type TjkdPlusDomainGetTjkdPlusDomainListRequest struct {
	Headers map[string]string
}

func NewTjkdPlusDomainGetTjkdPlusDomainListRequest() *TjkdPlusDomainGetTjkdPlusDomainListRequest {
	return &TjkdPlusDomainGetTjkdPlusDomainListRequest{Headers: map[string]string{}}
}

func (r *TjkdPlusDomainGetTjkdPlusDomainListRequest) APIName() string {
	return "TjkdPlusDomain_getTjkdPlusDomainList"
}
func (r *TjkdPlusDomainGetTjkdPlusDomainListRequest) Method() string { return "GET" }

func (r *TjkdPlusDomainGetTjkdPlusDomainListRequest) SetHeader(key string, value string) *TjkdPlusDomainGetTjkdPlusDomainListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *TjkdPlusDomainGetTjkdPlusDomainListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &TjkdPlusDomainGetTjkdPlusDomainListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type TjkdPlusDomainAddTjkdPlusDomainRequest struct {
	Headers   map[string]string
	PackageId interface{} `json:"package_id,omitempty"`
	DomainId  interface{} `json:"domain_id,omitempty"`
}

func NewTjkdPlusDomainAddTjkdPlusDomainRequest() *TjkdPlusDomainAddTjkdPlusDomainRequest {
	return &TjkdPlusDomainAddTjkdPlusDomainRequest{Headers: map[string]string{}}
}

func (r *TjkdPlusDomainAddTjkdPlusDomainRequest) APIName() string {
	return "TjkdPlusDomain_addTjkdPlusDomain"
}
func (r *TjkdPlusDomainAddTjkdPlusDomainRequest) Method() string { return "POST" }

func (r *TjkdPlusDomainAddTjkdPlusDomainRequest) SetHeader(key string, value string) *TjkdPlusDomainAddTjkdPlusDomainRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *TjkdPlusDomainAddTjkdPlusDomainRequest) RequestParts() ReqParams {
	if r == nil {
		r = &TjkdPlusDomainAddTjkdPlusDomainRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.PackageId != nil {
		parts.Data["package_id"] = r.PackageId
	}
	if r.DomainId != nil {
		parts.Data["domain_id"] = r.DomainId
	}
	return parts
}

type TjkdPlusDomainDelTjkdPlusDomainRequest struct {
	Headers               map[string]string
	PackageDomainIds      interface{} `json:"package_domain_ids,omitempty"`
	PackageDomains        interface{} `json:"package_domains,omitempty"`
	IgnoreNotExistsDomain interface{} `json:"ignore_not_exists_domain,omitempty"`
}

func NewTjkdPlusDomainDelTjkdPlusDomainRequest() *TjkdPlusDomainDelTjkdPlusDomainRequest {
	return &TjkdPlusDomainDelTjkdPlusDomainRequest{Headers: map[string]string{}}
}

func (r *TjkdPlusDomainDelTjkdPlusDomainRequest) APIName() string {
	return "TjkdPlusDomain_delTjkdPlusDomain"
}
func (r *TjkdPlusDomainDelTjkdPlusDomainRequest) Method() string { return "DELETE" }

func (r *TjkdPlusDomainDelTjkdPlusDomainRequest) SetHeader(key string, value string) *TjkdPlusDomainDelTjkdPlusDomainRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *TjkdPlusDomainDelTjkdPlusDomainRequest) RequestParts() ReqParams {
	if r == nil {
		r = &TjkdPlusDomainDelTjkdPlusDomainRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.PackageDomainIds != nil {
		parts.Data["package_domain_ids"] = r.PackageDomainIds
	}
	if r.PackageDomains != nil {
		parts.Data["package_domains"] = r.PackageDomains
	}
	if r.IgnoreNotExistsDomain != nil {
		parts.Data["ignore_not_exists_domain"] = r.IgnoreNotExistsDomain
	}
	return parts
}

type NetworkSpeedGetCacheRuleListRequest struct {
	Headers map[string]string
}

func NewNetworkSpeedGetCacheRuleListRequest() *NetworkSpeedGetCacheRuleListRequest {
	return &NetworkSpeedGetCacheRuleListRequest{Headers: map[string]string{}}
}

func (r *NetworkSpeedGetCacheRuleListRequest) APIName() string { return "NetworkSpeedGetCacheRuleList" }
func (r *NetworkSpeedGetCacheRuleListRequest) Method() string  { return "GET" }

func (r *NetworkSpeedGetCacheRuleListRequest) SetHeader(key string, value string) *NetworkSpeedGetCacheRuleListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *NetworkSpeedGetCacheRuleListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &NetworkSpeedGetCacheRuleListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type NetworkSpeedCreateCacheRuleRequest struct {
	Headers      map[string]string
	BusinessId   interface{} `json:"business_id,omitempty"`
	BusinessType interface{} `json:"business_type,omitempty"`
	Name         interface{} `json:"name,omitempty"`
	Expr         interface{} `json:"expr,omitempty"`
	Remark       interface{} `json:"remark,omitempty"`
	Conf         interface{} `json:"conf,omitempty"`
}

func NewNetworkSpeedCreateCacheRuleRequest() *NetworkSpeedCreateCacheRuleRequest {
	return &NetworkSpeedCreateCacheRuleRequest{Headers: map[string]string{}}
}

func (r *NetworkSpeedCreateCacheRuleRequest) APIName() string { return "NetworkSpeedCreateCacheRule" }
func (r *NetworkSpeedCreateCacheRuleRequest) Method() string  { return "POST" }

func (r *NetworkSpeedCreateCacheRuleRequest) SetHeader(key string, value string) *NetworkSpeedCreateCacheRuleRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *NetworkSpeedCreateCacheRuleRequest) RequestParts() ReqParams {
	if r == nil {
		r = &NetworkSpeedCreateCacheRuleRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.BusinessId != nil {
		parts.Data["business_id"] = r.BusinessId
	}
	if r.BusinessType != nil {
		parts.Data["business_type"] = r.BusinessType
	}
	if r.Name != nil {
		parts.Data["name"] = r.Name
	}
	if r.Expr != nil {
		parts.Data["expr"] = r.Expr
	}
	if r.Remark != nil {
		parts.Data["remark"] = r.Remark
	}
	if r.Conf != nil {
		parts.Data["conf"] = r.Conf
	}
	return parts
}

type NetworkSpeedUpdateCacheRuleRequest struct {
	Headers map[string]string
	Id      interface{} `json:"id,omitempty"`
	Name    interface{} `json:"name,omitempty"`
	Remark  interface{} `json:"remark,omitempty"`
}

func NewNetworkSpeedUpdateCacheRuleRequest() *NetworkSpeedUpdateCacheRuleRequest {
	return &NetworkSpeedUpdateCacheRuleRequest{Headers: map[string]string{}}
}

func (r *NetworkSpeedUpdateCacheRuleRequest) APIName() string { return "NetworkSpeedUpdateCacheRule" }
func (r *NetworkSpeedUpdateCacheRuleRequest) Method() string  { return "PUT" }

func (r *NetworkSpeedUpdateCacheRuleRequest) SetHeader(key string, value string) *NetworkSpeedUpdateCacheRuleRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *NetworkSpeedUpdateCacheRuleRequest) RequestParts() ReqParams {
	if r == nil {
		r = &NetworkSpeedUpdateCacheRuleRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Id != nil {
		parts.Data["id"] = r.Id
	}
	if r.Name != nil {
		parts.Data["name"] = r.Name
	}
	if r.Remark != nil {
		parts.Data["remark"] = r.Remark
	}
	return parts
}

type NetworkSpeedUpdateCacheRuleConfigRequest struct {
	Headers map[string]string
	Id      interface{} `json:"id,omitempty"`
	Name    interface{} `json:"name,omitempty"`
	Remark  interface{} `json:"remark,omitempty"`
	Expr    interface{} `json:"expr,omitempty"`
	Conf    interface{} `json:"conf,omitempty"`
}

func NewNetworkSpeedUpdateCacheRuleConfigRequest() *NetworkSpeedUpdateCacheRuleConfigRequest {
	return &NetworkSpeedUpdateCacheRuleConfigRequest{Headers: map[string]string{}}
}

func (r *NetworkSpeedUpdateCacheRuleConfigRequest) APIName() string {
	return "NetworkSpeedUpdateCacheRuleConfig"
}
func (r *NetworkSpeedUpdateCacheRuleConfigRequest) Method() string { return "PUT" }

func (r *NetworkSpeedUpdateCacheRuleConfigRequest) SetHeader(key string, value string) *NetworkSpeedUpdateCacheRuleConfigRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *NetworkSpeedUpdateCacheRuleConfigRequest) RequestParts() ReqParams {
	if r == nil {
		r = &NetworkSpeedUpdateCacheRuleConfigRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Id != nil {
		parts.Data["id"] = r.Id
	}
	if r.Name != nil {
		parts.Data["name"] = r.Name
	}
	if r.Remark != nil {
		parts.Data["remark"] = r.Remark
	}
	if r.Expr != nil {
		parts.Data["expr"] = r.Expr
	}
	if r.Conf != nil {
		parts.Data["conf"] = r.Conf
	}
	return parts
}

type NetworkSpeedUpdateCacheRuleStatusRequest struct {
	Headers      map[string]string
	BusinessId   interface{} `json:"business_id,omitempty"`
	BusinessType interface{} `json:"business_type,omitempty"`
	Ids          interface{} `json:"ids,omitempty"`
	Status       interface{} `json:"status,omitempty"`
}

func NewNetworkSpeedUpdateCacheRuleStatusRequest() *NetworkSpeedUpdateCacheRuleStatusRequest {
	return &NetworkSpeedUpdateCacheRuleStatusRequest{Headers: map[string]string{}}
}

func (r *NetworkSpeedUpdateCacheRuleStatusRequest) APIName() string {
	return "NetworkSpeedUpdateCacheRuleStatus"
}
func (r *NetworkSpeedUpdateCacheRuleStatusRequest) Method() string { return "PUT" }

func (r *NetworkSpeedUpdateCacheRuleStatusRequest) SetHeader(key string, value string) *NetworkSpeedUpdateCacheRuleStatusRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *NetworkSpeedUpdateCacheRuleStatusRequest) RequestParts() ReqParams {
	if r == nil {
		r = &NetworkSpeedUpdateCacheRuleStatusRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.BusinessId != nil {
		parts.Data["business_id"] = r.BusinessId
	}
	if r.BusinessType != nil {
		parts.Data["business_type"] = r.BusinessType
	}
	if r.Ids != nil {
		parts.Data["ids"] = r.Ids
	}
	if r.Status != nil {
		parts.Data["status"] = r.Status
	}
	return parts
}

type NetworkSpeedSortCacheRulesRequest struct {
	Headers      map[string]string
	BusinessId   interface{} `json:"business_id,omitempty"`
	BusinessType interface{} `json:"business_type,omitempty"`
	Ids          interface{} `json:"ids,omitempty"`
}

func NewNetworkSpeedSortCacheRulesRequest() *NetworkSpeedSortCacheRulesRequest {
	return &NetworkSpeedSortCacheRulesRequest{Headers: map[string]string{}}
}

func (r *NetworkSpeedSortCacheRulesRequest) APIName() string { return "NetworkSpeedSortCacheRules" }
func (r *NetworkSpeedSortCacheRulesRequest) Method() string  { return "PUT" }

func (r *NetworkSpeedSortCacheRulesRequest) SetHeader(key string, value string) *NetworkSpeedSortCacheRulesRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *NetworkSpeedSortCacheRulesRequest) RequestParts() ReqParams {
	if r == nil {
		r = &NetworkSpeedSortCacheRulesRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.BusinessId != nil {
		parts.Data["business_id"] = r.BusinessId
	}
	if r.BusinessType != nil {
		parts.Data["business_type"] = r.BusinessType
	}
	if r.Ids != nil {
		parts.Data["ids"] = r.Ids
	}
	return parts
}

type NetworkSpeedGetGlobalCacheConfigRequest struct {
	Headers map[string]string
}

func NewNetworkSpeedGetGlobalCacheConfigRequest() *NetworkSpeedGetGlobalCacheConfigRequest {
	return &NetworkSpeedGetGlobalCacheConfigRequest{Headers: map[string]string{}}
}

func (r *NetworkSpeedGetGlobalCacheConfigRequest) APIName() string {
	return "NetworkSpeedGetGlobalCacheConfig"
}
func (r *NetworkSpeedGetGlobalCacheConfigRequest) Method() string { return "GET" }

func (r *NetworkSpeedGetGlobalCacheConfigRequest) SetHeader(key string, value string) *NetworkSpeedGetGlobalCacheConfigRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *NetworkSpeedGetGlobalCacheConfigRequest) RequestParts() ReqParams {
	if r == nil {
		r = &NetworkSpeedGetGlobalCacheConfigRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type NetworkSpeedDeleteCacheRuleRequest struct {
	Headers      map[string]string
	BusinessId   interface{} `json:"business_id,omitempty"`
	BusinessType interface{} `json:"business_type,omitempty"`
	Ids          interface{} `json:"ids,omitempty"`
}

func NewNetworkSpeedDeleteCacheRuleRequest() *NetworkSpeedDeleteCacheRuleRequest {
	return &NetworkSpeedDeleteCacheRuleRequest{Headers: map[string]string{}}
}

func (r *NetworkSpeedDeleteCacheRuleRequest) APIName() string { return "NetworkSpeedDeleteCacheRule" }
func (r *NetworkSpeedDeleteCacheRuleRequest) Method() string  { return "DELETE" }

func (r *NetworkSpeedDeleteCacheRuleRequest) SetHeader(key string, value string) *NetworkSpeedDeleteCacheRuleRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *NetworkSpeedDeleteCacheRuleRequest) RequestParts() ReqParams {
	if r == nil {
		r = &NetworkSpeedDeleteCacheRuleRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.BusinessId != nil {
		parts.Data["business_id"] = r.BusinessId
	}
	if r.BusinessType != nil {
		parts.Data["business_type"] = r.BusinessType
	}
	if r.Ids != nil {
		parts.Data["ids"] = r.Ids
	}
	return parts
}

type NetworkSpeedGetTemplateConfigRequest struct {
	Headers       map[string]string
	BusinessId    interface{} `json:"business_id,omitempty"`
	BusinessType  interface{} `json:"business_type,omitempty"`
	ConfigGroups  interface{} `json:"config_groups,omitempty"`
	UpstreamCheck interface{} `json:"upstream_check,omitempty"`
}

func NewNetworkSpeedGetTemplateConfigRequest() *NetworkSpeedGetTemplateConfigRequest {
	return &NetworkSpeedGetTemplateConfigRequest{Headers: map[string]string{}}
}

func (r *NetworkSpeedGetTemplateConfigRequest) APIName() string {
	return "NetworkSpeedGetTemplateConfig"
}
func (r *NetworkSpeedGetTemplateConfigRequest) Method() string { return "POST" }

func (r *NetworkSpeedGetTemplateConfigRequest) SetHeader(key string, value string) *NetworkSpeedGetTemplateConfigRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *NetworkSpeedGetTemplateConfigRequest) RequestParts() ReqParams {
	if r == nil {
		r = &NetworkSpeedGetTemplateConfigRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.BusinessId != nil {
		parts.Data["business_id"] = r.BusinessId
	}
	if r.BusinessType != nil {
		parts.Data["business_type"] = r.BusinessType
	}
	if r.ConfigGroups != nil {
		parts.Data["config_groups"] = r.ConfigGroups
	}
	if r.UpstreamCheck != nil {
		parts.Data["upstream_check"] = r.UpstreamCheck
	}
	return parts
}

type NetworkSpeedUpdateTemplateConfigRequest struct {
	Headers              map[string]string
	BusinessId           interface{} `json:"business_id,omitempty"`
	BusinessType         interface{} `json:"business_type,omitempty"`
	DomainProxyConf      interface{} `json:"domain_proxy_conf,omitempty"`
	UpstreamRedirect     interface{} `json:"upstream_redirect,omitempty"`
	CustomizedReqHeaders interface{} `json:"customized_req_headers,omitempty"`
	SourceSiteProtect    interface{} `json:"source_site_protect,omitempty"`
	Slice                interface{} `json:"slice,omitempty"`
	Https                interface{} `json:"https,omitempty"`
	PageGzip             interface{} `json:"page_gzip,omitempty"`
	Webp                 interface{} `json:"webp,omitempty"`
	UploadFile           interface{} `json:"upload_file,omitempty"`
	Websocket            interface{} `json:"websocket,omitempty"`
	MobileJump           interface{} `json:"mobile_jump,omitempty"`
	CustomPage           interface{} `json:"custom_page,omitempty"`
	UpstreamUriChange    interface{} `json:"upstream_uri_change,omitempty"`
	RespHeaders          interface{} `json:"resp_headers,omitempty"`
	UpstreamCheck        interface{} `json:"upstream_check,omitempty"`
}

func NewNetworkSpeedUpdateTemplateConfigRequest() *NetworkSpeedUpdateTemplateConfigRequest {
	return &NetworkSpeedUpdateTemplateConfigRequest{Headers: map[string]string{}}
}

func (r *NetworkSpeedUpdateTemplateConfigRequest) APIName() string {
	return "NetworkSpeedUpdateTemplateConfig"
}
func (r *NetworkSpeedUpdateTemplateConfigRequest) Method() string { return "PUT" }

func (r *NetworkSpeedUpdateTemplateConfigRequest) SetHeader(key string, value string) *NetworkSpeedUpdateTemplateConfigRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *NetworkSpeedUpdateTemplateConfigRequest) RequestParts() ReqParams {
	if r == nil {
		r = &NetworkSpeedUpdateTemplateConfigRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.BusinessId != nil {
		parts.Data["business_id"] = r.BusinessId
	}
	if r.BusinessType != nil {
		parts.Data["business_type"] = r.BusinessType
	}
	if r.DomainProxyConf != nil {
		parts.Data["domain_proxy_conf"] = r.DomainProxyConf
	}
	if r.UpstreamRedirect != nil {
		parts.Data["upstream_redirect"] = r.UpstreamRedirect
	}
	if r.CustomizedReqHeaders != nil {
		parts.Data["customized_req_headers"] = r.CustomizedReqHeaders
	}
	if r.SourceSiteProtect != nil {
		parts.Data["source_site_protect"] = r.SourceSiteProtect
	}
	if r.Slice != nil {
		parts.Data["slice"] = r.Slice
	}
	if r.Https != nil {
		parts.Data["https"] = r.Https
	}
	if r.PageGzip != nil {
		parts.Data["page_gzip"] = r.PageGzip
	}
	if r.Webp != nil {
		parts.Data["webp"] = r.Webp
	}
	if r.UploadFile != nil {
		parts.Data["upload_file"] = r.UploadFile
	}
	if r.Websocket != nil {
		parts.Data["websocket"] = r.Websocket
	}
	if r.MobileJump != nil {
		parts.Data["mobile_jump"] = r.MobileJump
	}
	if r.CustomPage != nil {
		parts.Data["custom_page"] = r.CustomPage
	}
	if r.UpstreamUriChange != nil {
		parts.Data["upstream_uri_change"] = r.UpstreamUriChange
	}
	if r.RespHeaders != nil {
		parts.Data["resp_headers"] = r.RespHeaders
	}
	if r.UpstreamCheck != nil {
		parts.Data["upstream_check"] = r.UpstreamCheck
	}
	return parts
}

type NetworkSpeedGetRulesRequest struct {
	Headers map[string]string
}

func NewNetworkSpeedGetRulesRequest() *NetworkSpeedGetRulesRequest {
	return &NetworkSpeedGetRulesRequest{Headers: map[string]string{}}
}

func (r *NetworkSpeedGetRulesRequest) APIName() string { return "NetworkSpeedGetRules" }
func (r *NetworkSpeedGetRulesRequest) Method() string  { return "GET" }

func (r *NetworkSpeedGetRulesRequest) SetHeader(key string, value string) *NetworkSpeedGetRulesRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *NetworkSpeedGetRulesRequest) RequestParts() ReqParams {
	if r == nil {
		r = &NetworkSpeedGetRulesRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type NetworkSpeedCreateRuleRequest struct {
	Headers                  map[string]string
	BusinessId               interface{} `json:"business_id,omitempty"`
	BusinessType             interface{} `json:"business_type,omitempty"`
	ConfigGroup              interface{} `json:"config_group,omitempty"`
	CustomPage               interface{} `json:"custom_page,omitempty"`
	UpstreamUriChangeRule    interface{} `json:"upstream_uri_change_rule,omitempty"`
	RespHeadersRule          interface{} `json:"resp_headers_rule,omitempty"`
	CustomizedReqHeadersRule interface{} `json:"customized_req_headers_rule,omitempty"`
}

func NewNetworkSpeedCreateRuleRequest() *NetworkSpeedCreateRuleRequest {
	return &NetworkSpeedCreateRuleRequest{Headers: map[string]string{}}
}

func (r *NetworkSpeedCreateRuleRequest) APIName() string { return "NetworkSpeedCreateRule" }
func (r *NetworkSpeedCreateRuleRequest) Method() string  { return "POST" }

func (r *NetworkSpeedCreateRuleRequest) SetHeader(key string, value string) *NetworkSpeedCreateRuleRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *NetworkSpeedCreateRuleRequest) RequestParts() ReqParams {
	if r == nil {
		r = &NetworkSpeedCreateRuleRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.BusinessId != nil {
		parts.Data["business_id"] = r.BusinessId
	}
	if r.BusinessType != nil {
		parts.Data["business_type"] = r.BusinessType
	}
	if r.ConfigGroup != nil {
		parts.Data["config_group"] = r.ConfigGroup
	}
	if r.CustomPage != nil {
		parts.Data["custom_page"] = r.CustomPage
	}
	if r.UpstreamUriChangeRule != nil {
		parts.Data["upstream_uri_change_rule"] = r.UpstreamUriChangeRule
	}
	if r.RespHeadersRule != nil {
		parts.Data["resp_headers_rule"] = r.RespHeadersRule
	}
	if r.CustomizedReqHeadersRule != nil {
		parts.Data["customized_req_headers_rule"] = r.CustomizedReqHeadersRule
	}
	return parts
}

type NetworkSpeedDeleteRuleRequest struct {
	Headers      map[string]string
	BusinessId   interface{} `json:"business_id,omitempty"`
	BusinessType interface{} `json:"business_type,omitempty"`
	ConfigGroup  interface{} `json:"config_group,omitempty"`
	Ids          interface{} `json:"ids,omitempty"`
}

func NewNetworkSpeedDeleteRuleRequest() *NetworkSpeedDeleteRuleRequest {
	return &NetworkSpeedDeleteRuleRequest{Headers: map[string]string{}}
}

func (r *NetworkSpeedDeleteRuleRequest) APIName() string { return "NetworkSpeedDeleteRule" }
func (r *NetworkSpeedDeleteRuleRequest) Method() string  { return "DELETE" }

func (r *NetworkSpeedDeleteRuleRequest) SetHeader(key string, value string) *NetworkSpeedDeleteRuleRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *NetworkSpeedDeleteRuleRequest) RequestParts() ReqParams {
	if r == nil {
		r = &NetworkSpeedDeleteRuleRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.BusinessId != nil {
		parts.Data["business_id"] = r.BusinessId
	}
	if r.BusinessType != nil {
		parts.Data["business_type"] = r.BusinessType
	}
	if r.ConfigGroup != nil {
		parts.Data["config_group"] = r.ConfigGroup
	}
	if r.Ids != nil {
		parts.Data["ids"] = r.Ids
	}
	return parts
}

type NetworkSpeedSortRulesRequest struct {
	Headers      map[string]string
	BusinessId   interface{} `json:"business_id,omitempty"`
	BusinessType interface{} `json:"business_type,omitempty"`
	ConfigGroup  interface{} `json:"config_group,omitempty"`
	Ids          interface{} `json:"ids,omitempty"`
}

func NewNetworkSpeedSortRulesRequest() *NetworkSpeedSortRulesRequest {
	return &NetworkSpeedSortRulesRequest{Headers: map[string]string{}}
}

func (r *NetworkSpeedSortRulesRequest) APIName() string { return "NetworkSpeedSortRules" }
func (r *NetworkSpeedSortRulesRequest) Method() string  { return "PUT" }

func (r *NetworkSpeedSortRulesRequest) SetHeader(key string, value string) *NetworkSpeedSortRulesRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *NetworkSpeedSortRulesRequest) RequestParts() ReqParams {
	if r == nil {
		r = &NetworkSpeedSortRulesRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.BusinessId != nil {
		parts.Data["business_id"] = r.BusinessId
	}
	if r.BusinessType != nil {
		parts.Data["business_type"] = r.BusinessType
	}
	if r.ConfigGroup != nil {
		parts.Data["config_group"] = r.ConfigGroup
	}
	if r.Ids != nil {
		parts.Data["ids"] = r.Ids
	}
	return parts
}

type NetworkSpeedUpdateRuleRequest struct {
	Headers                  map[string]string
	Id                       interface{} `json:"id,omitempty"`
	ConfigGroup              interface{} `json:"config_group,omitempty"`
	CustomPage               interface{} `json:"custom_page,omitempty"`
	UpstreamUriChangeRule    interface{} `json:"upstream_uri_change_rule,omitempty"`
	RespHeadersRule          interface{} `json:"resp_headers_rule,omitempty"`
	CustomizedReqHeadersRule interface{} `json:"customized_req_headers_rule,omitempty"`
}

func NewNetworkSpeedUpdateRuleRequest() *NetworkSpeedUpdateRuleRequest {
	return &NetworkSpeedUpdateRuleRequest{Headers: map[string]string{}}
}

func (r *NetworkSpeedUpdateRuleRequest) APIName() string { return "NetworkSpeedUpdateRule" }
func (r *NetworkSpeedUpdateRuleRequest) Method() string  { return "PUT" }

func (r *NetworkSpeedUpdateRuleRequest) SetHeader(key string, value string) *NetworkSpeedUpdateRuleRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *NetworkSpeedUpdateRuleRequest) RequestParts() ReqParams {
	if r == nil {
		r = &NetworkSpeedUpdateRuleRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Id != nil {
		parts.Data["id"] = r.Id
	}
	if r.ConfigGroup != nil {
		parts.Data["config_group"] = r.ConfigGroup
	}
	if r.CustomPage != nil {
		parts.Data["custom_page"] = r.CustomPage
	}
	if r.UpstreamUriChangeRule != nil {
		parts.Data["upstream_uri_change_rule"] = r.UpstreamUriChangeRule
	}
	if r.RespHeadersRule != nil {
		parts.Data["resp_headers_rule"] = r.RespHeadersRule
	}
	if r.CustomizedReqHeadersRule != nil {
		parts.Data["customized_req_headers_rule"] = r.CustomizedReqHeadersRule
	}
	return parts
}

type UpdateRuleTemplateRequest struct {
	Headers     map[string]string
	Id          interface{} `json:"id,omitempty"`
	Name        interface{} `json:"name,omitempty"`
	Description interface{} `json:"description,omitempty"`
}

func NewUpdateRuleTemplateRequest() *UpdateRuleTemplateRequest {
	return &UpdateRuleTemplateRequest{Headers: map[string]string{}}
}

func (r *UpdateRuleTemplateRequest) APIName() string { return "UpdateRuleTemplate" }
func (r *UpdateRuleTemplateRequest) Method() string  { return "PUT" }

func (r *UpdateRuleTemplateRequest) SetHeader(key string, value string) *UpdateRuleTemplateRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *UpdateRuleTemplateRequest) RequestParts() ReqParams {
	if r == nil {
		r = &UpdateRuleTemplateRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Id != nil {
		parts.Data["id"] = r.Id
	}
	if r.Name != nil {
		parts.Data["name"] = r.Name
	}
	if r.Description != nil {
		parts.Data["description"] = r.Description
	}
	return parts
}

type DeleteRuleTemplateRequest struct {
	Headers map[string]string
	Id      interface{} `json:"id,omitempty"`
}

func NewDeleteRuleTemplateRequest() *DeleteRuleTemplateRequest {
	return &DeleteRuleTemplateRequest{Headers: map[string]string{}}
}

func (r *DeleteRuleTemplateRequest) APIName() string { return "DeleteRuleTemplate" }
func (r *DeleteRuleTemplateRequest) Method() string  { return "DELETE" }

func (r *DeleteRuleTemplateRequest) SetHeader(key string, value string) *DeleteRuleTemplateRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DeleteRuleTemplateRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DeleteRuleTemplateRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Id != nil {
		parts.Data["id"] = r.Id
	}
	return parts
}

type GetRuleTemplateListRequest struct {
	Headers map[string]string
}

func NewGetRuleTemplateListRequest() *GetRuleTemplateListRequest {
	return &GetRuleTemplateListRequest{Headers: map[string]string{}}
}

func (r *GetRuleTemplateListRequest) APIName() string { return "GetRuleTemplateList" }
func (r *GetRuleTemplateListRequest) Method() string  { return "GET" }

func (r *GetRuleTemplateListRequest) SetHeader(key string, value string) *GetRuleTemplateListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *GetRuleTemplateListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &GetRuleTemplateListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type UnbindRuleTemplateRequest struct {
	Headers   map[string]string
	Id        interface{} `json:"id,omitempty"`
	DomainIds interface{} `json:"domain_ids,omitempty"`
}

func NewUnbindRuleTemplateRequest() *UnbindRuleTemplateRequest {
	return &UnbindRuleTemplateRequest{Headers: map[string]string{}}
}

func (r *UnbindRuleTemplateRequest) APIName() string { return "UnbindRuleTemplate" }
func (r *UnbindRuleTemplateRequest) Method() string  { return "PUT" }

func (r *UnbindRuleTemplateRequest) SetHeader(key string, value string) *UnbindRuleTemplateRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *UnbindRuleTemplateRequest) RequestParts() ReqParams {
	if r == nil {
		r = &UnbindRuleTemplateRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Id != nil {
		parts.Data["id"] = r.Id
	}
	if r.DomainIds != nil {
		parts.Data["domain_ids"] = r.DomainIds
	}
	return parts
}

type BindRuleTemplateRequest struct {
	Headers   map[string]string
	Id        interface{} `json:"id,omitempty"`
	DomainIds interface{} `json:"domain_ids,omitempty"`
}

func NewBindRuleTemplateRequest() *BindRuleTemplateRequest {
	return &BindRuleTemplateRequest{Headers: map[string]string{}}
}

func (r *BindRuleTemplateRequest) APIName() string { return "BindRuleTemplate" }
func (r *BindRuleTemplateRequest) Method() string  { return "PUT" }

func (r *BindRuleTemplateRequest) SetHeader(key string, value string) *BindRuleTemplateRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *BindRuleTemplateRequest) RequestParts() ReqParams {
	if r == nil {
		r = &BindRuleTemplateRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Id != nil {
		parts.Data["id"] = r.Id
	}
	if r.DomainIds != nil {
		parts.Data["domain_ids"] = r.DomainIds
	}
	return parts
}

type ListRuleTpsDomainsRequest struct {
	Headers map[string]string
}

func NewListRuleTpsDomainsRequest() *ListRuleTpsDomainsRequest {
	return &ListRuleTpsDomainsRequest{Headers: map[string]string{}}
}

func (r *ListRuleTpsDomainsRequest) APIName() string { return "ListRuleTpsDomains" }
func (r *ListRuleTpsDomainsRequest) Method() string  { return "GET" }

func (r *ListRuleTpsDomainsRequest) SetHeader(key string, value string) *ListRuleTpsDomainsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *ListRuleTpsDomainsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &ListRuleTpsDomainsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type CreateRuleTemplateRequest struct {
	Headers     map[string]string
	Name        interface{} `json:"name,omitempty"`
	Description interface{} `json:"description,omitempty"`
	AppType     interface{} `json:"app_type,omitempty"`
	TplType     interface{} `json:"tpl_type,omitempty"`
	DomainId    interface{} `json:"domain_id,omitempty"`
	FromTplId   interface{} `json:"from_tpl_id,omitempty"`
	FromTplType interface{} `json:"from_tpl_type,omitempty"`
	BindDomain  interface{} `json:"bind_domain,omitempty"`
}

func NewCreateRuleTemplateRequest() *CreateRuleTemplateRequest {
	return &CreateRuleTemplateRequest{Headers: map[string]string{}}
}

func (r *CreateRuleTemplateRequest) APIName() string { return "CreateRuleTemplate" }
func (r *CreateRuleTemplateRequest) Method() string  { return "POST" }

func (r *CreateRuleTemplateRequest) SetHeader(key string, value string) *CreateRuleTemplateRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CreateRuleTemplateRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CreateRuleTemplateRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Name != nil {
		parts.Data["name"] = r.Name
	}
	if r.Description != nil {
		parts.Data["description"] = r.Description
	}
	parts.Data["app_type"] = "network_speed"
	if r.AppType != nil {
		parts.Data["app_type"] = r.AppType
	}
	if r.TplType != nil {
		parts.Data["tpl_type"] = r.TplType
	}
	if r.DomainId != nil {
		parts.Data["domain_id"] = r.DomainId
	}
	if r.FromTplId != nil {
		parts.Data["from_tpl_id"] = r.FromTplId
	}
	if r.FromTplType != nil {
		parts.Data["from_tpl_type"] = r.FromTplType
	}
	if r.BindDomain != nil {
		parts.Data["bind_domain"] = r.BindDomain
	}
	return parts
}

type SwitchDomainTemplateRequest struct {
	Headers    map[string]string
	AppType    interface{} `json:"app_type,omitempty"`
	DomainIds  interface{} `json:"domain_ids,omitempty"`
	NewTplId   interface{} `json:"new_tpl_id,omitempty"`
	NewTplType interface{} `json:"new_tpl_type,omitempty"`
}

func NewSwitchDomainTemplateRequest() *SwitchDomainTemplateRequest {
	return &SwitchDomainTemplateRequest{Headers: map[string]string{}}
}

func (r *SwitchDomainTemplateRequest) APIName() string { return "SwitchDomainTemplate" }
func (r *SwitchDomainTemplateRequest) Method() string  { return "PUT" }

func (r *SwitchDomainTemplateRequest) SetHeader(key string, value string) *SwitchDomainTemplateRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *SwitchDomainTemplateRequest) RequestParts() ReqParams {
	if r == nil {
		r = &SwitchDomainTemplateRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	parts.Data["app_type"] = "network_speed"
	if r.AppType != nil {
		parts.Data["app_type"] = r.AppType
	}
	if r.DomainIds != nil {
		parts.Data["domain_ids"] = r.DomainIds
	}
	if r.NewTplId != nil {
		parts.Data["new_tpl_id"] = r.NewTplId
	}
	if r.NewTplType != nil {
		parts.Data["new_tpl_type"] = r.NewTplType
	}
	return parts
}

type FirewallPageCfgRequest struct {
	Headers map[string]string
}

func NewFirewallPageCfgRequest() *FirewallPageCfgRequest {
	return &FirewallPageCfgRequest{Headers: map[string]string{}}
}

func (r *FirewallPageCfgRequest) APIName() string { return "Firewall_pageCfg" }
func (r *FirewallPageCfgRequest) Method() string  { return "GET" }

func (r *FirewallPageCfgRequest) SetHeader(key string, value string) *FirewallPageCfgRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *FirewallPageCfgRequest) RequestParts() ReqParams {
	if r == nil {
		r = &FirewallPageCfgRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type FirewallPageCfgHwwsRequest struct {
	Headers map[string]string
}

func NewFirewallPageCfgHwwsRequest() *FirewallPageCfgHwwsRequest {
	return &FirewallPageCfgHwwsRequest{Headers: map[string]string{}}
}

func (r *FirewallPageCfgHwwsRequest) APIName() string { return "Firewall_pageCfgHwws" }
func (r *FirewallPageCfgHwwsRequest) Method() string  { return "GET" }

func (r *FirewallPageCfgHwwsRequest) SetHeader(key string, value string) *FirewallPageCfgHwwsRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *FirewallPageCfgHwwsRequest) RequestParts() ReqParams {
	if r == nil {
		r = &FirewallPageCfgHwwsRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type FirewallSavePolicyRequest struct {
	Headers     map[string]string
	Id          interface{} `json:"id,omitempty"`
	BusinessId  interface{} `json:"business_id,omitempty"`
	PackageId   interface{} `json:"package_id,omitempty"`
	ProductFlag interface{} `json:"product_flag,omitempty"`
	GroupId     interface{} `json:"group_id,omitempty"`
	TjkdAppId   interface{} `json:"tjkd_app_id,omitempty"`
	From        interface{} `json:"from,omitempty"`
	Remark      interface{} `json:"remark,omitempty"`
	TypeValue   interface{} `json:"type,omitempty"`
	UseType     interface{} `json:"use_type,omitempty"`
	Action      interface{} `json:"action,omitempty"`
	ActionData  interface{} `json:"action_data,omitempty"`
	Rules       interface{} `json:"rules,omitempty"`
}

func NewFirewallSavePolicyRequest() *FirewallSavePolicyRequest {
	return &FirewallSavePolicyRequest{Headers: map[string]string{}}
}

func (r *FirewallSavePolicyRequest) APIName() string { return "Firewall_savePolicy" }
func (r *FirewallSavePolicyRequest) Method() string  { return "POST" }

func (r *FirewallSavePolicyRequest) SetHeader(key string, value string) *FirewallSavePolicyRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *FirewallSavePolicyRequest) RequestParts() ReqParams {
	if r == nil {
		r = &FirewallSavePolicyRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Id != nil {
		parts.Data["id"] = r.Id
	}
	if r.BusinessId != nil {
		parts.Data["business_id"] = r.BusinessId
	}
	if r.PackageId != nil {
		parts.Data["package_id"] = r.PackageId
	}
	parts.Data["product_flag"] = "plus"
	if r.ProductFlag != nil {
		parts.Data["product_flag"] = r.ProductFlag
	}
	if r.GroupId != nil {
		parts.Data["group_id"] = r.GroupId
	}
	if r.TjkdAppId != nil {
		parts.Data["tjkd_app_id"] = r.TjkdAppId
	}
	if r.From != nil {
		parts.Data["from"] = r.From
	}
	if r.Remark != nil {
		parts.Data["remark"] = r.Remark
	}
	if r.TypeValue != nil {
		parts.Data["type"] = r.TypeValue
	}
	if r.UseType != nil {
		parts.Data["use_type"] = r.UseType
	}
	if r.Action != nil {
		parts.Data["action"] = r.Action
	}
	if r.ActionData != nil {
		parts.Data["action_data"] = r.ActionData
	}
	if r.Rules != nil {
		parts.Data["rules"] = r.Rules
	}
	return parts
}

type FirewallGetPolicyRequest struct {
	Headers map[string]string
}

func NewFirewallGetPolicyRequest() *FirewallGetPolicyRequest {
	return &FirewallGetPolicyRequest{Headers: map[string]string{}}
}

func (r *FirewallGetPolicyRequest) APIName() string { return "Firewall_getPolicy" }
func (r *FirewallGetPolicyRequest) Method() string  { return "GET" }

func (r *FirewallGetPolicyRequest) SetHeader(key string, value string) *FirewallGetPolicyRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *FirewallGetPolicyRequest) RequestParts() ReqParams {
	if r == nil {
		r = &FirewallGetPolicyRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type FirewallGetPolicyByCodeRequest struct {
	Headers map[string]string
}

func NewFirewallGetPolicyByCodeRequest() *FirewallGetPolicyByCodeRequest {
	return &FirewallGetPolicyByCodeRequest{Headers: map[string]string{}}
}

func (r *FirewallGetPolicyByCodeRequest) APIName() string { return "Firewall_getPolicyByCode" }
func (r *FirewallGetPolicyByCodeRequest) Method() string  { return "GET" }

func (r *FirewallGetPolicyByCodeRequest) SetHeader(key string, value string) *FirewallGetPolicyByCodeRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *FirewallGetPolicyByCodeRequest) RequestParts() ReqParams {
	if r == nil {
		r = &FirewallGetPolicyByCodeRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type FirewallStatsPolicyRequest struct {
	Headers map[string]string
}

func NewFirewallStatsPolicyRequest() *FirewallStatsPolicyRequest {
	return &FirewallStatsPolicyRequest{Headers: map[string]string{}}
}

func (r *FirewallStatsPolicyRequest) APIName() string { return "Firewall_statsPolicy" }
func (r *FirewallStatsPolicyRequest) Method() string  { return "GET" }

func (r *FirewallStatsPolicyRequest) SetHeader(key string, value string) *FirewallStatsPolicyRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *FirewallStatsPolicyRequest) RequestParts() ReqParams {
	if r == nil {
		r = &FirewallStatsPolicyRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type FirewallOpenRequest struct {
	Headers map[string]string
}

func NewFirewallOpenRequest() *FirewallOpenRequest {
	return &FirewallOpenRequest{Headers: map[string]string{}}
}

func (r *FirewallOpenRequest) APIName() string { return "Firewall_open" }
func (r *FirewallOpenRequest) Method() string  { return "POST" }

func (r *FirewallOpenRequest) SetHeader(key string, value string) *FirewallOpenRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *FirewallOpenRequest) RequestParts() ReqParams {
	if r == nil {
		r = &FirewallOpenRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type FirewallStopRequest struct {
	Headers map[string]string
	Ids     interface{} `json:"ids,omitempty"`
}

func NewFirewallStopRequest() *FirewallStopRequest {
	return &FirewallStopRequest{Headers: map[string]string{}}
}

func (r *FirewallStopRequest) APIName() string { return "Firewall_stop" }
func (r *FirewallStopRequest) Method() string  { return "POST" }

func (r *FirewallStopRequest) SetHeader(key string, value string) *FirewallStopRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *FirewallStopRequest) RequestParts() ReqParams {
	if r == nil {
		r = &FirewallStopRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Ids != nil {
		parts.Data["ids"] = r.Ids
	}
	return parts
}

type FirewallDeleteRequest struct {
	Headers map[string]string
	Ids     interface{} `json:"ids,omitempty"`
}

func NewFirewallDeleteRequest() *FirewallDeleteRequest {
	return &FirewallDeleteRequest{Headers: map[string]string{}}
}

func (r *FirewallDeleteRequest) APIName() string { return "Firewall_delete" }
func (r *FirewallDeleteRequest) Method() string  { return "POST" }

func (r *FirewallDeleteRequest) SetHeader(key string, value string) *FirewallDeleteRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *FirewallDeleteRequest) RequestParts() ReqParams {
	if r == nil {
		r = &FirewallDeleteRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Ids != nil {
		parts.Data["ids"] = r.Ids
	}
	return parts
}

type FirewallSortRequest struct {
	Headers  map[string]string
	NewSorts interface{} `json:"new_sorts,omitempty"`
}

func NewFirewallSortRequest() *FirewallSortRequest {
	return &FirewallSortRequest{Headers: map[string]string{}}
}

func (r *FirewallSortRequest) APIName() string { return "Firewall_sort" }
func (r *FirewallSortRequest) Method() string  { return "POST" }

func (r *FirewallSortRequest) SetHeader(key string, value string) *FirewallSortRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *FirewallSortRequest) RequestParts() ReqParams {
	if r == nil {
		r = &FirewallSortRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.NewSorts != nil {
		parts.Data["new_sorts"] = r.NewSorts
	}
	return parts
}

type FirewallGetsPolicyByMainidRequest struct {
	Headers map[string]string
}

func NewFirewallGetsPolicyByMainidRequest() *FirewallGetsPolicyByMainidRequest {
	return &FirewallGetsPolicyByMainidRequest{Headers: map[string]string{}}
}

func (r *FirewallGetsPolicyByMainidRequest) APIName() string { return "Firewall_getsPolicyByMainid" }
func (r *FirewallGetsPolicyByMainidRequest) Method() string  { return "GET" }

func (r *FirewallGetsPolicyByMainidRequest) SetHeader(key string, value string) *FirewallGetsPolicyByMainidRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *FirewallGetsPolicyByMainidRequest) RequestParts() ReqParams {
	if r == nil {
		r = &FirewallGetsPolicyByMainidRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type FirewallGetsPolicyByPackageidRequest struct {
	Headers map[string]string
}

func NewFirewallGetsPolicyByPackageidRequest() *FirewallGetsPolicyByPackageidRequest {
	return &FirewallGetsPolicyByPackageidRequest{Headers: map[string]string{}}
}

func (r *FirewallGetsPolicyByPackageidRequest) APIName() string {
	return "Firewall_getsPolicyByPackageid"
}
func (r *FirewallGetsPolicyByPackageidRequest) Method() string { return "GET" }

func (r *FirewallGetsPolicyByPackageidRequest) SetHeader(key string, value string) *FirewallGetsPolicyByPackageidRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *FirewallGetsPolicyByPackageidRequest) RequestParts() ReqParams {
	if r == nil {
		r = &FirewallGetsPolicyByPackageidRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type FirewallSavePolicyGroupRequest struct {
	Headers    map[string]string
	Id         interface{} `json:"id,omitempty"`
	BusinessId interface{} `json:"business_id,omitempty"`
	From       interface{} `json:"from,omitempty"`
	Remark     interface{} `json:"remark,omitempty"`
	Name       interface{} `json:"name,omitempty"`
}

func NewFirewallSavePolicyGroupRequest() *FirewallSavePolicyGroupRequest {
	return &FirewallSavePolicyGroupRequest{Headers: map[string]string{}}
}

func (r *FirewallSavePolicyGroupRequest) APIName() string { return "Firewall_savePolicyGroup" }
func (r *FirewallSavePolicyGroupRequest) Method() string  { return "POST" }

func (r *FirewallSavePolicyGroupRequest) SetHeader(key string, value string) *FirewallSavePolicyGroupRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *FirewallSavePolicyGroupRequest) RequestParts() ReqParams {
	if r == nil {
		r = &FirewallSavePolicyGroupRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Id != nil {
		parts.Data["id"] = r.Id
	}
	if r.BusinessId != nil {
		parts.Data["business_id"] = r.BusinessId
	}
	parts.Data["from"] = "diy"
	if r.From != nil {
		parts.Data["from"] = r.From
	}
	if r.Remark != nil {
		parts.Data["remark"] = r.Remark
	}
	if r.Name != nil {
		parts.Data["name"] = r.Name
	}
	return parts
}

type FirewallGetsPolicyGroupByDomainidRequest struct {
	Headers map[string]string
}

func NewFirewallGetsPolicyGroupByDomainidRequest() *FirewallGetsPolicyGroupByDomainidRequest {
	return &FirewallGetsPolicyGroupByDomainidRequest{Headers: map[string]string{}}
}

func (r *FirewallGetsPolicyGroupByDomainidRequest) APIName() string {
	return "Firewall_getsPolicyGroupByDomainid"
}
func (r *FirewallGetsPolicyGroupByDomainidRequest) Method() string { return "GET" }

func (r *FirewallGetsPolicyGroupByDomainidRequest) SetHeader(key string, value string) *FirewallGetsPolicyGroupByDomainidRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *FirewallGetsPolicyGroupByDomainidRequest) RequestParts() ReqParams {
	if r == nil {
		r = &FirewallGetsPolicyGroupByDomainidRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type FirewallStopGroupRequest struct {
	Headers map[string]string
	Ids     interface{} `json:"ids,omitempty"`
}

func NewFirewallStopGroupRequest() *FirewallStopGroupRequest {
	return &FirewallStopGroupRequest{Headers: map[string]string{}}
}

func (r *FirewallStopGroupRequest) APIName() string { return "Firewall_stopGroup" }
func (r *FirewallStopGroupRequest) Method() string  { return "POST" }

func (r *FirewallStopGroupRequest) SetHeader(key string, value string) *FirewallStopGroupRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *FirewallStopGroupRequest) RequestParts() ReqParams {
	if r == nil {
		r = &FirewallStopGroupRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Ids != nil {
		parts.Data["ids"] = r.Ids
	}
	return parts
}

type FirewallOpenGroupRequest struct {
	Headers map[string]string
	Ids     interface{} `json:"ids,omitempty"`
}

func NewFirewallOpenGroupRequest() *FirewallOpenGroupRequest {
	return &FirewallOpenGroupRequest{Headers: map[string]string{}}
}

func (r *FirewallOpenGroupRequest) APIName() string { return "Firewall_openGroup" }
func (r *FirewallOpenGroupRequest) Method() string  { return "POST" }

func (r *FirewallOpenGroupRequest) SetHeader(key string, value string) *FirewallOpenGroupRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *FirewallOpenGroupRequest) RequestParts() ReqParams {
	if r == nil {
		r = &FirewallOpenGroupRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Ids != nil {
		parts.Data["ids"] = r.Ids
	}
	return parts
}

type FirewallDeleteGroupRequest struct {
	Headers map[string]string
	Ids     interface{} `json:"ids,omitempty"`
}

func NewFirewallDeleteGroupRequest() *FirewallDeleteGroupRequest {
	return &FirewallDeleteGroupRequest{Headers: map[string]string{}}
}

func (r *FirewallDeleteGroupRequest) APIName() string { return "Firewall_deleteGroup" }
func (r *FirewallDeleteGroupRequest) Method() string  { return "POST" }

func (r *FirewallDeleteGroupRequest) SetHeader(key string, value string) *FirewallDeleteGroupRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *FirewallDeleteGroupRequest) RequestParts() ReqParams {
	if r == nil {
		r = &FirewallDeleteGroupRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Ids != nil {
		parts.Data["ids"] = r.Ids
	}
	return parts
}

type FirewallSortGroupRequest struct {
	Headers  map[string]string
	NewSorts interface{} `json:"new_sorts,omitempty"`
}

func NewFirewallSortGroupRequest() *FirewallSortGroupRequest {
	return &FirewallSortGroupRequest{Headers: map[string]string{}}
}

func (r *FirewallSortGroupRequest) APIName() string { return "Firewall_sortGroup" }
func (r *FirewallSortGroupRequest) Method() string  { return "POST" }

func (r *FirewallSortGroupRequest) SetHeader(key string, value string) *FirewallSortGroupRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *FirewallSortGroupRequest) RequestParts() ReqParams {
	if r == nil {
		r = &FirewallSortGroupRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.NewSorts != nil {
		parts.Data["new_sorts"] = r.NewSorts
	}
	return parts
}

type FirewallGetsPolicyByGroupIdRequest struct {
	Headers map[string]string
}

func NewFirewallGetsPolicyByGroupIdRequest() *FirewallGetsPolicyByGroupIdRequest {
	return &FirewallGetsPolicyByGroupIdRequest{Headers: map[string]string{}}
}

func (r *FirewallGetsPolicyByGroupIdRequest) APIName() string { return "Firewall_getsPolicyByGroupId" }
func (r *FirewallGetsPolicyByGroupIdRequest) Method() string  { return "GET" }

func (r *FirewallGetsPolicyByGroupIdRequest) SetHeader(key string, value string) *FirewallGetsPolicyByGroupIdRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *FirewallGetsPolicyByGroupIdRequest) RequestParts() ReqParams {
	if r == nil {
		r = &FirewallGetsPolicyByGroupIdRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type GetPolicyGroupTplRequest struct {
	Headers map[string]string
}

func NewGetPolicyGroupTplRequest() *GetPolicyGroupTplRequest {
	return &GetPolicyGroupTplRequest{Headers: map[string]string{}}
}

func (r *GetPolicyGroupTplRequest) APIName() string { return "getPolicyGroupTPL" }
func (r *GetPolicyGroupTplRequest) Method() string  { return "GET" }

func (r *GetPolicyGroupTplRequest) SetHeader(key string, value string) *GetPolicyGroupTplRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *GetPolicyGroupTplRequest) RequestParts() ReqParams {
	if r == nil {
		r = &GetPolicyGroupTplRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type GetDdosProtectionConfigRequest struct {
	Headers map[string]string
}

func NewGetDdosProtectionConfigRequest() *GetDdosProtectionConfigRequest {
	return &GetDdosProtectionConfigRequest{Headers: map[string]string{}}
}

func (r *GetDdosProtectionConfigRequest) APIName() string { return "GetDdosProtectionConfig" }
func (r *GetDdosProtectionConfigRequest) Method() string  { return "GET" }

func (r *GetDdosProtectionConfigRequest) SetHeader(key string, value string) *GetDdosProtectionConfigRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *GetDdosProtectionConfigRequest) RequestParts() ReqParams {
	if r == nil {
		r = &GetDdosProtectionConfigRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type UpdateDdosProtectionConfigRequest struct {
	Headers                   map[string]string
	BusinessId                interface{} `json:"business_id,omitempty"`
	ApplicationDdosProtection interface{} `json:"application_ddos_protection,omitempty"`
	VisitorAuthentication     interface{} `json:"visitor_authentication,omitempty"`
}

func NewUpdateDdosProtectionConfigRequest() *UpdateDdosProtectionConfigRequest {
	return &UpdateDdosProtectionConfigRequest{Headers: map[string]string{}}
}

func (r *UpdateDdosProtectionConfigRequest) APIName() string { return "UpdateDdosProtectionConfig" }
func (r *UpdateDdosProtectionConfigRequest) Method() string  { return "PUT" }

func (r *UpdateDdosProtectionConfigRequest) SetHeader(key string, value string) *UpdateDdosProtectionConfigRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *UpdateDdosProtectionConfigRequest) RequestParts() ReqParams {
	if r == nil {
		r = &UpdateDdosProtectionConfigRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.BusinessId != nil {
		parts.Data["business_id"] = r.BusinessId
	}
	if r.ApplicationDdosProtection != nil {
		parts.Data["application_ddos_protection"] = r.ApplicationDdosProtection
	}
	if r.VisitorAuthentication != nil {
		parts.Data["visitor_authentication"] = r.VisitorAuthentication
	}
	return parts
}

type GetWafRuleConfigRequest struct {
	Headers map[string]string
}

func NewGetWafRuleConfigRequest() *GetWafRuleConfigRequest {
	return &GetWafRuleConfigRequest{Headers: map[string]string{}}
}

func (r *GetWafRuleConfigRequest) APIName() string { return "GetWafRuleConfig" }
func (r *GetWafRuleConfigRequest) Method() string  { return "GET" }

func (r *GetWafRuleConfigRequest) SetHeader(key string, value string) *GetWafRuleConfigRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *GetWafRuleConfigRequest) RequestParts() ReqParams {
	if r == nil {
		r = &GetWafRuleConfigRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type UpdateWafRuleConfigRequest struct {
	Headers                map[string]string
	BusinessId             interface{} `json:"business_id,omitempty"`
	WafRuleConfig          interface{} `json:"waf_rule_config,omitempty"`
	WafInterceptPage       interface{} `json:"waf_intercept_page,omitempty"`
	ReplayAttackProtection interface{} `json:"replay_attack_protection,omitempty"`
	CsrfProtection         interface{} `json:"csrf_protection,omitempty"`
	WebShellProtection     interface{} `json:"web_shell_protection,omitempty"`
}

func NewUpdateWafRuleConfigRequest() *UpdateWafRuleConfigRequest {
	return &UpdateWafRuleConfigRequest{Headers: map[string]string{}}
}

func (r *UpdateWafRuleConfigRequest) APIName() string { return "UpdateWafRuleConfig" }
func (r *UpdateWafRuleConfigRequest) Method() string  { return "PUT" }

func (r *UpdateWafRuleConfigRequest) SetHeader(key string, value string) *UpdateWafRuleConfigRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *UpdateWafRuleConfigRequest) RequestParts() ReqParams {
	if r == nil {
		r = &UpdateWafRuleConfigRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.BusinessId != nil {
		parts.Data["business_id"] = r.BusinessId
	}
	if r.WafRuleConfig != nil {
		parts.Data["waf_rule_config"] = r.WafRuleConfig
	}
	if r.WafInterceptPage != nil {
		parts.Data["waf_intercept_page"] = r.WafInterceptPage
	}
	if r.ReplayAttackProtection != nil {
		parts.Data["replay_attack_protection"] = r.ReplayAttackProtection
	}
	if r.CsrfProtection != nil {
		parts.Data["csrf_protection"] = r.CsrfProtection
	}
	if r.WebShellProtection != nil {
		parts.Data["web_shell_protection"] = r.WebShellProtection
	}
	return parts
}

type GetMemberGlobalTemplateRequest struct {
	Headers map[string]string
}

func NewGetMemberGlobalTemplateRequest() *GetMemberGlobalTemplateRequest {
	return &GetMemberGlobalTemplateRequest{Headers: map[string]string{}}
}

func (r *GetMemberGlobalTemplateRequest) APIName() string { return "GetMemberGlobalTemplate" }
func (r *GetMemberGlobalTemplateRequest) Method() string  { return "GET" }

func (r *GetMemberGlobalTemplateRequest) SetHeader(key string, value string) *GetMemberGlobalTemplateRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *GetMemberGlobalTemplateRequest) RequestParts() ReqParams {
	if r == nil {
		r = &GetMemberGlobalTemplateRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type CreateTemplateRequest struct {
	Headers          map[string]string
	Name             interface{} `json:"name,omitempty"`
	Remark           interface{} `json:"remark,omitempty"`
	TemplateSourceId interface{} `json:"template_source_id,omitempty"`
	DomainIds        interface{} `json:"domain_ids,omitempty"`
	GroupIds         interface{} `json:"group_ids,omitempty"`
	Domains          interface{} `json:"domains,omitempty"`
	BindAll          interface{} `json:"bind_all,omitempty"`
}

func NewCreateTemplateRequest() *CreateTemplateRequest {
	return &CreateTemplateRequest{Headers: map[string]string{}}
}

func (r *CreateTemplateRequest) APIName() string { return "CreateTemplate" }
func (r *CreateTemplateRequest) Method() string  { return "POST" }

func (r *CreateTemplateRequest) SetHeader(key string, value string) *CreateTemplateRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CreateTemplateRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CreateTemplateRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Name != nil {
		parts.Data["name"] = r.Name
	}
	if r.Remark != nil {
		parts.Data["remark"] = r.Remark
	}
	if r.TemplateSourceId != nil {
		parts.Data["template_source_id"] = r.TemplateSourceId
	}
	if r.DomainIds != nil {
		parts.Data["domain_ids"] = r.DomainIds
	}
	if r.GroupIds != nil {
		parts.Data["group_ids"] = r.GroupIds
	}
	if r.Domains != nil {
		parts.Data["domains"] = r.Domains
	}
	if r.BindAll != nil {
		parts.Data["bind_all"] = r.BindAll
	}
	return parts
}

type CreateDomainTemplateRequest struct {
	Headers          map[string]string
	DomainIds        interface{} `json:"domain_ids,omitempty"`
	TemplateSourceId interface{} `json:"template_source_id,omitempty"`
}

func NewCreateDomainTemplateRequest() *CreateDomainTemplateRequest {
	return &CreateDomainTemplateRequest{Headers: map[string]string{}}
}

func (r *CreateDomainTemplateRequest) APIName() string { return "CreateDomainTemplate" }
func (r *CreateDomainTemplateRequest) Method() string  { return "POST" }

func (r *CreateDomainTemplateRequest) SetHeader(key string, value string) *CreateDomainTemplateRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *CreateDomainTemplateRequest) RequestParts() ReqParams {
	if r == nil {
		r = &CreateDomainTemplateRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.DomainIds != nil {
		parts.Data["domain_ids"] = r.DomainIds
	}
	if r.TemplateSourceId != nil {
		parts.Data["template_source_id"] = r.TemplateSourceId
	}
	return parts
}

type GetTemplateListRequest struct {
	Headers    map[string]string
	TplType    interface{} `json:"tpl_type,omitempty"`
	SearchType interface{} `json:"search_type,omitempty"`
	SearchKey  interface{} `json:"search_key,omitempty"`
	Page       interface{} `json:"page,omitempty"`
	PageSize   interface{} `json:"page_size,omitempty"`
}

func NewGetTemplateListRequest() *GetTemplateListRequest {
	return &GetTemplateListRequest{Headers: map[string]string{}}
}

func (r *GetTemplateListRequest) APIName() string { return "GetTemplateList" }
func (r *GetTemplateListRequest) Method() string  { return "POST" }

func (r *GetTemplateListRequest) SetHeader(key string, value string) *GetTemplateListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *GetTemplateListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &GetTemplateListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.TplType != nil {
		parts.Data["tpl_type"] = r.TplType
	}
	if r.SearchType != nil {
		parts.Data["search_type"] = r.SearchType
	}
	if r.SearchKey != nil {
		parts.Data["search_key"] = r.SearchKey
	}
	if r.Page != nil {
		parts.Data["page"] = r.Page
	}
	if r.PageSize != nil {
		parts.Data["page_size"] = r.PageSize
	}
	return parts
}

type GetTemplateBindDomainListRequest struct {
	Headers    map[string]string
	BusinessId interface{} `json:"business_id,omitempty"`
	Page       interface{} `json:"page,omitempty"`
	PageSize   interface{} `json:"page_size,omitempty"`
	Domain     interface{} `json:"domain,omitempty"`
	TplType    interface{} `json:"tpl_type,omitempty"`
}

func NewGetTemplateBindDomainListRequest() *GetTemplateBindDomainListRequest {
	return &GetTemplateBindDomainListRequest{Headers: map[string]string{}}
}

func (r *GetTemplateBindDomainListRequest) APIName() string { return "GetTemplateBindDomainList" }
func (r *GetTemplateBindDomainListRequest) Method() string  { return "POST" }

func (r *GetTemplateBindDomainListRequest) SetHeader(key string, value string) *GetTemplateBindDomainListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *GetTemplateBindDomainListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &GetTemplateBindDomainListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.BusinessId != nil {
		parts.Data["business_id"] = r.BusinessId
	}
	if r.Page != nil {
		parts.Data["page"] = r.Page
	}
	if r.PageSize != nil {
		parts.Data["page_size"] = r.PageSize
	}
	if r.Domain != nil {
		parts.Data["domain"] = r.Domain
	}
	if r.TplType != nil {
		parts.Data["tpl_type"] = r.TplType
	}
	return parts
}

type BindTemplateDomainRequest struct {
	Headers         map[string]string
	BusinessId      interface{} `json:"business_id,omitempty"`
	DomainIds       interface{} `json:"domain_ids,omitempty"`
	BindBusinessIds interface{} `json:"bind_business_ids,omitempty"`
}

func NewBindTemplateDomainRequest() *BindTemplateDomainRequest {
	return &BindTemplateDomainRequest{Headers: map[string]string{}}
}

func (r *BindTemplateDomainRequest) APIName() string { return "BindTemplateDomain" }
func (r *BindTemplateDomainRequest) Method() string  { return "POST" }

func (r *BindTemplateDomainRequest) SetHeader(key string, value string) *BindTemplateDomainRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *BindTemplateDomainRequest) RequestParts() ReqParams {
	if r == nil {
		r = &BindTemplateDomainRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.BusinessId != nil {
		parts.Data["business_id"] = r.BusinessId
	}
	if r.DomainIds != nil {
		parts.Data["domain_ids"] = r.DomainIds
	}
	if r.BindBusinessIds != nil {
		parts.Data["bind_business_ids"] = r.BindBusinessIds
	}
	return parts
}

type DeleteTemplateRequest struct {
	Headers    map[string]string
	BusinessId interface{} `json:"business_id,omitempty"`
}

func NewDeleteTemplateRequest() *DeleteTemplateRequest {
	return &DeleteTemplateRequest{Headers: map[string]string{}}
}

func (r *DeleteTemplateRequest) APIName() string { return "DeleteTemplate" }
func (r *DeleteTemplateRequest) Method() string  { return "DELETE" }

func (r *DeleteTemplateRequest) SetHeader(key string, value string) *DeleteTemplateRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DeleteTemplateRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DeleteTemplateRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.BusinessId != nil {
		parts.Data["business_id"] = r.BusinessId
	}
	return parts
}

type BatchConfigTemplateRequest struct {
	Headers                    map[string]string
	TemplateIds                interface{} `json:"template_ids,omitempty"`
	DdosConfig                 interface{} `json:"ddos_config,omitempty"`
	PreciseAccessControlConfig interface{} `json:"precise_access_control_config,omitempty"`
	WafRuleConfig              interface{} `json:"waf_rule_config,omitempty"`
	BotManagementConfig        interface{} `json:"bot_management_config,omitempty"`
}

func NewBatchConfigTemplateRequest() *BatchConfigTemplateRequest {
	return &BatchConfigTemplateRequest{Headers: map[string]string{}}
}

func (r *BatchConfigTemplateRequest) APIName() string { return "BatchConfigTemplate" }
func (r *BatchConfigTemplateRequest) Method() string  { return "POST" }

func (r *BatchConfigTemplateRequest) SetHeader(key string, value string) *BatchConfigTemplateRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *BatchConfigTemplateRequest) RequestParts() ReqParams {
	if r == nil {
		r = &BatchConfigTemplateRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.TemplateIds != nil {
		parts.Data["template_ids"] = r.TemplateIds
	}
	if r.DdosConfig != nil {
		parts.Data["ddos_config"] = r.DdosConfig
	}
	if r.PreciseAccessControlConfig != nil {
		parts.Data["precise_access_control_config"] = r.PreciseAccessControlConfig
	}
	if r.WafRuleConfig != nil {
		parts.Data["waf_rule_config"] = r.WafRuleConfig
	}
	if r.BotManagementConfig != nil {
		parts.Data["bot_management_config"] = r.BotManagementConfig
	}
	return parts
}

type IotaRequest struct {
	Headers map[string]string
}

func NewIotaRequest() *IotaRequest {
	return &IotaRequest{Headers: map[string]string{}}
}

func (r *IotaRequest) APIName() string { return "Iota" }
func (r *IotaRequest) Method() string  { return "GET" }

func (r *IotaRequest) SetHeader(key string, value string) *IotaRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *IotaRequest) RequestParts() ReqParams {
	if r == nil {
		r = &IotaRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type GetUnboundTemplateDomainListRequest struct {
	Headers map[string]string
}

func NewGetUnboundTemplateDomainListRequest() *GetUnboundTemplateDomainListRequest {
	return &GetUnboundTemplateDomainListRequest{Headers: map[string]string{}}
}

func (r *GetUnboundTemplateDomainListRequest) APIName() string { return "GetUnboundTemplateDomainList" }
func (r *GetUnboundTemplateDomainListRequest) Method() string  { return "POST" }

func (r *GetUnboundTemplateDomainListRequest) SetHeader(key string, value string) *GetUnboundTemplateDomainListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *GetUnboundTemplateDomainListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &GetUnboundTemplateDomainListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type EditTemplateRequest struct {
	Headers    map[string]string
	BusinessId interface{} `json:"business_id,omitempty"`
	Name       interface{} `json:"name,omitempty"`
	Remark     interface{} `json:"remark,omitempty"`
}

func NewEditTemplateRequest() *EditTemplateRequest {
	return &EditTemplateRequest{Headers: map[string]string{}}
}

func (r *EditTemplateRequest) APIName() string { return "EditTemplate" }
func (r *EditTemplateRequest) Method() string  { return "PUT" }

func (r *EditTemplateRequest) SetHeader(key string, value string) *EditTemplateRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *EditTemplateRequest) RequestParts() ReqParams {
	if r == nil {
		r = &EditTemplateRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.BusinessId != nil {
		parts.Data["business_id"] = r.BusinessId
	}
	if r.Name != nil {
		parts.Data["name"] = r.Name
	}
	if r.Remark != nil {
		parts.Data["remark"] = r.Remark
	}
	return parts
}

type FirewallSavePolicyGroupRegionalShieldingRequest struct {
	Headers    map[string]string
	BusinessId interface{} `json:"business_id,omitempty"`
	From       interface{} `json:"from,omitempty"`
	Name       interface{} `json:"name,omitempty"`
}

func NewFirewallSavePolicyGroupRegionalShieldingRequest() *FirewallSavePolicyGroupRegionalShieldingRequest {
	return &FirewallSavePolicyGroupRegionalShieldingRequest{Headers: map[string]string{}}
}

func (r *FirewallSavePolicyGroupRegionalShieldingRequest) APIName() string {
	return "Firewall_savePolicyGroupRegionalShielding"
}
func (r *FirewallSavePolicyGroupRegionalShieldingRequest) Method() string { return "POST" }

func (r *FirewallSavePolicyGroupRegionalShieldingRequest) SetHeader(key string, value string) *FirewallSavePolicyGroupRegionalShieldingRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *FirewallSavePolicyGroupRegionalShieldingRequest) RequestParts() ReqParams {
	if r == nil {
		r = &FirewallSavePolicyGroupRegionalShieldingRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.BusinessId != nil {
		parts.Data["business_id"] = r.BusinessId
	}
	if r.From != nil {
		parts.Data["from"] = r.From
	}
	if r.Name != nil {
		parts.Data["name"] = r.Name
	}
	return parts
}

type FirewallSavePolicyGroupAntiLeechRequest struct {
	Headers    map[string]string
	BusinessId interface{} `json:"business_id,omitempty"`
	From       interface{} `json:"from,omitempty"`
	Name       interface{} `json:"name,omitempty"`
}

func NewFirewallSavePolicyGroupAntiLeechRequest() *FirewallSavePolicyGroupAntiLeechRequest {
	return &FirewallSavePolicyGroupAntiLeechRequest{Headers: map[string]string{}}
}

func (r *FirewallSavePolicyGroupAntiLeechRequest) APIName() string {
	return "Firewall_savePolicyGroupAntiLeech"
}
func (r *FirewallSavePolicyGroupAntiLeechRequest) Method() string { return "POST" }

func (r *FirewallSavePolicyGroupAntiLeechRequest) SetHeader(key string, value string) *FirewallSavePolicyGroupAntiLeechRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *FirewallSavePolicyGroupAntiLeechRequest) RequestParts() ReqParams {
	if r == nil {
		r = &FirewallSavePolicyGroupAntiLeechRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.BusinessId != nil {
		parts.Data["business_id"] = r.BusinessId
	}
	if r.From != nil {
		parts.Data["from"] = r.From
	}
	if r.Name != nil {
		parts.Data["name"] = r.Name
	}
	return parts
}

type TjkdappsaveFirewallPolicyRequest struct {
	Headers   map[string]string
	Id        interface{} `json:"id,omitempty"`
	TjkdAppId interface{} `json:"tjkd_app_id,omitempty"`
	TypeValue interface{} `json:"type,omitempty"`
	Rules     interface{} `json:"rules,omitempty"`
	Action    interface{} `json:"action,omitempty"`
	Remark    interface{} `json:"remark,omitempty"`
	Status    interface{} `json:"status,omitempty"`
}

func NewTjkdappsaveFirewallPolicyRequest() *TjkdappsaveFirewallPolicyRequest {
	return &TjkdappsaveFirewallPolicyRequest{Headers: map[string]string{}}
}

func (r *TjkdappsaveFirewallPolicyRequest) APIName() string { return "TjkdappsaveFirewallPolicy" }
func (r *TjkdappsaveFirewallPolicyRequest) Method() string  { return "POST" }

func (r *TjkdappsaveFirewallPolicyRequest) SetHeader(key string, value string) *TjkdappsaveFirewallPolicyRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *TjkdappsaveFirewallPolicyRequest) RequestParts() ReqParams {
	if r == nil {
		r = &TjkdappsaveFirewallPolicyRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Id != nil {
		parts.Data["id"] = r.Id
	}
	if r.TjkdAppId != nil {
		parts.Data["tjkd_app_id"] = r.TjkdAppId
	}
	if r.TypeValue != nil {
		parts.Data["type"] = r.TypeValue
	}
	if r.Rules != nil {
		parts.Data["rules"] = r.Rules
	}
	if r.Action != nil {
		parts.Data["action"] = r.Action
	}
	if r.Remark != nil {
		parts.Data["remark"] = r.Remark
	}
	if r.Status != nil {
		parts.Data["status"] = r.Status
	}
	return parts
}

type TjkdappsortFirewallPolicyRequest struct {
	Headers  map[string]string
	NewSorts interface{} `json:"new_sorts,omitempty"`
}

func NewTjkdappsortFirewallPolicyRequest() *TjkdappsortFirewallPolicyRequest {
	return &TjkdappsortFirewallPolicyRequest{Headers: map[string]string{}}
}

func (r *TjkdappsortFirewallPolicyRequest) APIName() string { return "TjkdappsortFirewallPolicy" }
func (r *TjkdappsortFirewallPolicyRequest) Method() string  { return "POST" }

func (r *TjkdappsortFirewallPolicyRequest) SetHeader(key string, value string) *TjkdappsortFirewallPolicyRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *TjkdappsortFirewallPolicyRequest) RequestParts() ReqParams {
	if r == nil {
		r = &TjkdappsortFirewallPolicyRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.NewSorts != nil {
		parts.Data["new_sorts"] = r.NewSorts
	}
	return parts
}

type TjkdappopenFirewallPolicyRequest struct {
	Headers map[string]string
	Ids     interface{} `json:"ids,omitempty"`
}

func NewTjkdappopenFirewallPolicyRequest() *TjkdappopenFirewallPolicyRequest {
	return &TjkdappopenFirewallPolicyRequest{Headers: map[string]string{}}
}

func (r *TjkdappopenFirewallPolicyRequest) APIName() string { return "TjkdappopenFirewallPolicy" }
func (r *TjkdappopenFirewallPolicyRequest) Method() string  { return "POST" }

func (r *TjkdappopenFirewallPolicyRequest) SetHeader(key string, value string) *TjkdappopenFirewallPolicyRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *TjkdappopenFirewallPolicyRequest) RequestParts() ReqParams {
	if r == nil {
		r = &TjkdappopenFirewallPolicyRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Ids != nil {
		parts.Data["ids"] = r.Ids
	}
	return parts
}

type TjkdappstopFirewallPolicyRequest struct {
	Headers map[string]string
	Ids     interface{} `json:"ids,omitempty"`
}

func NewTjkdappstopFirewallPolicyRequest() *TjkdappstopFirewallPolicyRequest {
	return &TjkdappstopFirewallPolicyRequest{Headers: map[string]string{}}
}

func (r *TjkdappstopFirewallPolicyRequest) APIName() string { return "TjkdappstopFirewallPolicy" }
func (r *TjkdappstopFirewallPolicyRequest) Method() string  { return "POST" }

func (r *TjkdappstopFirewallPolicyRequest) SetHeader(key string, value string) *TjkdappstopFirewallPolicyRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *TjkdappstopFirewallPolicyRequest) RequestParts() ReqParams {
	if r == nil {
		r = &TjkdappstopFirewallPolicyRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Ids != nil {
		parts.Data["ids"] = r.Ids
	}
	return parts
}

type TjkdappgetFirewallPolicyRequest struct {
	Headers map[string]string
}

func NewTjkdappgetFirewallPolicyRequest() *TjkdappgetFirewallPolicyRequest {
	return &TjkdappgetFirewallPolicyRequest{Headers: map[string]string{}}
}

func (r *TjkdappgetFirewallPolicyRequest) APIName() string { return "TjkdappgetFirewallPolicy" }
func (r *TjkdappgetFirewallPolicyRequest) Method() string  { return "GET" }

func (r *TjkdappgetFirewallPolicyRequest) SetHeader(key string, value string) *TjkdappgetFirewallPolicyRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *TjkdappgetFirewallPolicyRequest) RequestParts() ReqParams {
	if r == nil {
		r = &TjkdappgetFirewallPolicyRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type TjkdappdeleteFirewallPolicyRequest struct {
	Headers map[string]string
	Ids     interface{} `json:"ids,omitempty"`
}

func NewTjkdappdeleteFirewallPolicyRequest() *TjkdappdeleteFirewallPolicyRequest {
	return &TjkdappdeleteFirewallPolicyRequest{Headers: map[string]string{}}
}

func (r *TjkdappdeleteFirewallPolicyRequest) APIName() string { return "TjkdappdeleteFirewallPolicy" }
func (r *TjkdappdeleteFirewallPolicyRequest) Method() string  { return "POST" }

func (r *TjkdappdeleteFirewallPolicyRequest) SetHeader(key string, value string) *TjkdappdeleteFirewallPolicyRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *TjkdappdeleteFirewallPolicyRequest) RequestParts() ReqParams {
	if r == nil {
		r = &TjkdappdeleteFirewallPolicyRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Ids != nil {
		parts.Data["ids"] = r.Ids
	}
	return parts
}

type AddForwardRuleRequest struct {
	Headers           map[string]string
	PackageId         interface{} `json:"package_id,omitempty"`
	Domain            interface{} `json:"domain,omitempty"`
	Protocol          interface{} `json:"protocol,omitempty"`
	Port              interface{} `json:"port,omitempty"`
	Loading           interface{} `json:"loading,omitempty"`
	Remark            interface{} `json:"remark,omitempty"`
	SourceType        interface{} `json:"source_type,omitempty"`
	SourceList        interface{} `json:"source_list,omitempty"`
	ChannelStatus     interface{} `json:"channel_status,omitempty"`
	ChannelLoading    interface{} `json:"channel_loading,omitempty"`
	ChannelSourceList interface{} `json:"channel_source_list,omitempty"`
}

func NewAddForwardRuleRequest() *AddForwardRuleRequest {
	return &AddForwardRuleRequest{Headers: map[string]string{}}
}

func (r *AddForwardRuleRequest) APIName() string { return "addForwardRule" }
func (r *AddForwardRuleRequest) Method() string  { return "POST" }

func (r *AddForwardRuleRequest) SetHeader(key string, value string) *AddForwardRuleRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *AddForwardRuleRequest) RequestParts() ReqParams {
	if r == nil {
		r = &AddForwardRuleRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.PackageId != nil {
		parts.Data["package_id"] = r.PackageId
	}
	if r.Domain != nil {
		parts.Data["domain"] = r.Domain
	}
	if r.Protocol != nil {
		parts.Data["protocol"] = r.Protocol
	}
	if r.Port != nil {
		parts.Data["port"] = r.Port
	}
	if r.Loading != nil {
		parts.Data["loading"] = r.Loading
	}
	if r.Remark != nil {
		parts.Data["remark"] = r.Remark
	}
	if r.SourceType != nil {
		parts.Data["source_type"] = r.SourceType
	}
	if r.SourceList != nil {
		parts.Data["source_list"] = r.SourceList
	}
	if r.ChannelStatus != nil {
		parts.Data["channel_status"] = r.ChannelStatus
	}
	if r.ChannelLoading != nil {
		parts.Data["channel_loading"] = r.ChannelLoading
	}
	if r.ChannelSourceList != nil {
		parts.Data["channel_source_list"] = r.ChannelSourceList
	}
	return parts
}

type DeleteForwardRuleRequest struct {
	Headers map[string]string
	Ids     interface{} `json:"ids,omitempty"`
}

func NewDeleteForwardRuleRequest() *DeleteForwardRuleRequest {
	return &DeleteForwardRuleRequest{Headers: map[string]string{}}
}

func (r *DeleteForwardRuleRequest) APIName() string { return "deleteForwardRule" }
func (r *DeleteForwardRuleRequest) Method() string  { return "DELETE" }

func (r *DeleteForwardRuleRequest) SetHeader(key string, value string) *DeleteForwardRuleRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *DeleteForwardRuleRequest) RequestParts() ReqParams {
	if r == nil {
		r = &DeleteForwardRuleRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Ids != nil {
		parts.Data["ids"] = r.Ids
	}
	return parts
}

type EditRuleRequest struct {
	Headers           map[string]string
	Id                interface{} `json:"id,omitempty"`
	PackageId         interface{} `json:"package_id,omitempty"`
	Protocol          interface{} `json:"protocol,omitempty"`
	Domain            interface{} `json:"domain,omitempty"`
	Port              interface{} `json:"port,omitempty"`
	Loading           interface{} `json:"loading,omitempty"`
	SourceType        interface{} `json:"source_type,omitempty"`
	SourceList        interface{} `json:"source_list,omitempty"`
	ChannelStatus     interface{} `json:"channel_status,omitempty"`
	ChannelLoading    interface{} `json:"channel_loading,omitempty"`
	ChannelSourceList interface{} `json:"channel_source_list,omitempty"`
}

func NewEditRuleRequest() *EditRuleRequest {
	return &EditRuleRequest{Headers: map[string]string{}}
}

func (r *EditRuleRequest) APIName() string { return "editRule" }
func (r *EditRuleRequest) Method() string  { return "POST" }

func (r *EditRuleRequest) SetHeader(key string, value string) *EditRuleRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *EditRuleRequest) RequestParts() ReqParams {
	if r == nil {
		r = &EditRuleRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Id != nil {
		parts.Data["id"] = r.Id
	}
	if r.PackageId != nil {
		parts.Data["package_id"] = r.PackageId
	}
	if r.Protocol != nil {
		parts.Data["protocol"] = r.Protocol
	}
	if r.Domain != nil {
		parts.Data["domain"] = r.Domain
	}
	if r.Port != nil {
		parts.Data["port"] = r.Port
	}
	if r.Loading != nil {
		parts.Data["loading"] = r.Loading
	}
	if r.SourceType != nil {
		parts.Data["source_type"] = r.SourceType
	}
	if r.SourceList != nil {
		parts.Data["source_list"] = r.SourceList
	}
	if r.ChannelStatus != nil {
		parts.Data["channel_status"] = r.ChannelStatus
	}
	if r.ChannelLoading != nil {
		parts.Data["channel_loading"] = r.ChannelLoading
	}
	if r.ChannelSourceList != nil {
		parts.Data["channel_source_list"] = r.ChannelSourceList
	}
	return parts
}

type RuleListRequest struct {
	Headers   map[string]string
	Page      interface{} `json:"page,omitempty"`
	PrePage   interface{} `json:"pre_page,omitempty"`
	Order     interface{} `json:"order,omitempty"`
	PackageId interface{} `json:"package_id,omitempty"`
}

func NewRuleListRequest() *RuleListRequest {
	return &RuleListRequest{Headers: map[string]string{}}
}

func (r *RuleListRequest) APIName() string { return "ruleList" }
func (r *RuleListRequest) Method() string  { return "GET" }

func (r *RuleListRequest) SetHeader(key string, value string) *RuleListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *RuleListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &RuleListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.Page != nil {
		parts.Query["page"] = r.Page
	}
	if r.PrePage != nil {
		parts.Query["pre_page"] = r.PrePage
	}
	if r.Order != nil {
		parts.Query["order"] = r.Order
	}
	if r.PackageId != nil {
		parts.Query["package_id"] = r.PackageId
	}
	return parts
}

type GetRuleInfoRequest struct {
	Headers map[string]string
}

func NewGetRuleInfoRequest() *GetRuleInfoRequest {
	return &GetRuleInfoRequest{Headers: map[string]string{}}
}

func (r *GetRuleInfoRequest) APIName() string { return "getRuleInfo" }
func (r *GetRuleInfoRequest) Method() string  { return "GET" }

func (r *GetRuleInfoRequest) SetHeader(key string, value string) *GetRuleInfoRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *GetRuleInfoRequest) RequestParts() ReqParams {
	if r == nil {
		r = &GetRuleInfoRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type TijkdappListPackageRequest struct {
	Headers map[string]string
}

func NewTijkdappListPackageRequest() *TijkdappListPackageRequest {
	return &TijkdappListPackageRequest{Headers: map[string]string{}}
}

func (r *TijkdappListPackageRequest) APIName() string { return "TIJKDAPP_ListPackage" }
func (r *TijkdappListPackageRequest) Method() string  { return "GET" }

func (r *TijkdappListPackageRequest) SetHeader(key string, value string) *TijkdappListPackageRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *TijkdappListPackageRequest) RequestParts() ReqParams {
	if r == nil {
		r = &TijkdappListPackageRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type TijkdappSavePackageRequest struct {
	Headers     map[string]string
	PackageId   interface{} `json:"package_id,omitempty"`
	PackageName interface{} `json:"package_name,omitempty"`
}

func NewTijkdappSavePackageRequest() *TijkdappSavePackageRequest {
	return &TijkdappSavePackageRequest{Headers: map[string]string{}}
}

func (r *TijkdappSavePackageRequest) APIName() string { return "TIJKDAPP_SavePackage" }
func (r *TijkdappSavePackageRequest) Method() string  { return "PUT" }

func (r *TijkdappSavePackageRequest) SetHeader(key string, value string) *TijkdappSavePackageRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *TijkdappSavePackageRequest) RequestParts() ReqParams {
	if r == nil {
		r = &TijkdappSavePackageRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	if r.PackageId != nil {
		parts.Data["package_id"] = r.PackageId
	}
	if r.PackageName != nil {
		parts.Data["package_name"] = r.PackageName
	}
	return parts
}

type GetChannelListRequest struct {
	Headers map[string]string
}

func NewGetChannelListRequest() *GetChannelListRequest {
	return &GetChannelListRequest{Headers: map[string]string{}}
}

func (r *GetChannelListRequest) APIName() string { return "getChannelList" }
func (r *GetChannelListRequest) Method() string  { return "GET" }

func (r *GetChannelListRequest) SetHeader(key string, value string) *GetChannelListRequest {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *GetChannelListRequest) RequestParts() ReqParams {
	if r == nil {
		r = &GetChannelListRequest{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	return parts
}

type ApiNameV5Request struct {
	Headers map[string]string
	XLang   interface{} `json:"x-lang,omitempty"`
}

func NewApiNameV5Request() *ApiNameV5Request {
	return &ApiNameV5Request{Headers: map[string]string{}}
}

func (r *ApiNameV5Request) APIName() string { return "api_name_v5" }
func (r *ApiNameV5Request) Method() string  { return "POST" }

func (r *ApiNameV5Request) SetHeader(key string, value string) *ApiNameV5Request {
	if r.Headers == nil {
		r.Headers = map[string]string{}
	}
	r.Headers[key] = value
	return r
}

func (r *ApiNameV5Request) RequestParts() ReqParams {
	if r == nil {
		r = &ApiNameV5Request{}
	}
	parts := ReqParams{Query: map[string]interface{}{}, Data: map[string]interface{}{}, Headers: cloneStringMap(r.Headers)}
	parts.Headers["x-lang"] = "zh"
	if r.XLang != nil {
		parts.Headers["x-lang"] = fmt.Sprint(r.XLang)
	}
	return parts
}

type CdnHighDefenseIpGetArticleIpResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type DnsDomainGetDomainListResponse struct {
	Status APIResponseStatus                  `json:"status"`
	Data   DnsDomainGetDomainListResponseData `json:"data,omitempty"`
}

type DnsDomainGetDomainListResponseData struct {
	Total float64                                  `json:"total,omitempty"`
	List  []DnsDomainGetDomainListResponseDataList `json:"list,omitempty"`
}

type DnsDomainGetDomainListResponseDataList struct {
	Id              float64 `json:"id,omitempty"`
	MemberId        float64 `json:"member_id,omitempty"`
	Domain          string  `json:"domain,omitempty"`
	TrustStatus     string  `json:"trust_status,omitempty"`
	TrustStatusDesc string  `json:"trust_status_desc,omitempty"`
	Status          float64 `json:"status,omitempty"`
}

type DnsDomainAddDomainResponse struct {
	Status APIResponseStatus              `json:"status"`
	Data   DnsDomainAddDomainResponseData `json:"data,omitempty"`
}

type DnsDomainAddDomainResponseData struct {
	Id float64 `json:"id,omitempty"`
}

type DnsDomainBatchAddDomainsResponse struct {
	Status APIResponseStatus      `json:"status"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

type DnsDomainBatchDeleteDomainsResponse struct {
	Status APIResponseStatus      `json:"status"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

type DnsDomainGetDomainStatResponse struct {
	Status APIResponseStatus                  `json:"status"`
	Data   DnsDomainGetDomainStatResponseData `json:"data,omitempty"`
}

type DnsDomainGetDomainStatResponseData struct {
	Unit string                 `json:"unit,omitempty"`
	List map[string]interface{} `json:"list,omitempty"`
}

type DnsDomainGetDomainServersResponse struct {
	Status APIResponseStatus                     `json:"status"`
	Data   DnsDomainGetDomainServersResponseData `json:"data,omitempty"`
}

type DnsDomainGetDomainServersResponseData struct {
	TypeValue      float64  `json:"type,omitempty"`
	Message        string   `json:"message,omitempty"`
	CurrentServers []string `json:"current_servers,omitempty"`
	MyServers      []string `json:"my_servers,omitempty"`
}

type DnsDomainGetTasksListResponse struct {
	Status APIResponseStatus                 `json:"status"`
	Data   DnsDomainGetTasksListResponseData `json:"data,omitempty"`
}

type DnsDomainGetTasksListResponseData struct {
	Total      float64                                 `json:"total,omitempty"`
	List       []DnsDomainGetTasksListResponseDataList `json:"list,omitempty"`
	StatusList map[string]interface{}                  `json:"status_list,omitempty"`
	TypeList   map[string]interface{}                  `json:"type_list,omitempty"`
}

type DnsDomainGetTasksListResponseDataList struct {
	Id               float64 `json:"id,omitempty"`
	Status           float64 `json:"status,omitempty"`
	Total            float64 `json:"total,omitempty"`
	Success          float64 `json:"success,omitempty"`
	Fail             float64 `json:"fail,omitempty"`
	Msg              string  `json:"msg,omitempty"`
	TypeValue        float64 `json:"type,omitempty"`
	MsgCode          float64 `json:"msg_code,omitempty"`
	OperatorUserName string  `json:"operator_user_name,omitempty"`
	Desc             string  `json:"desc,omitempty"`
	CreatedAt        string  `json:"created_at,omitempty"`
}

type DnsDomainGetTaskDetailResponse struct {
	Status APIResponseStatus                  `json:"status"`
	Data   DnsDomainGetTaskDetailResponseData `json:"data,omitempty"`
}

type DnsDomainGetTaskDetailResponseData struct {
	Id        float64                                    `json:"id,omitempty"`
	TypeValue float64                                    `json:"type,omitempty"`
	QwJobId   float64                                    `json:"qw_job_id,omitempty"`
	Uid       float64                                    `json:"uid,omitempty"`
	Status    float64                                    `json:"status,omitempty"`
	Total     float64                                    `json:"total,omitempty"`
	Success   float64                                    `json:"success,omitempty"`
	Fail      float64                                    `json:"fail,omitempty"`
	Msg       string                                     `json:"msg,omitempty"`
	MsgCode   float64                                    `json:"msg_code,omitempty"`
	CreatedAt string                                     `json:"created_at,omitempty"`
	UpdatedAt string                                     `json:"updated_at,omitempty"`
	SubUserId float64                                    `json:"sub_user_id,omitempty"`
	Detail    []DnsDomainGetTaskDetailResponseDataDetail `json:"detail,omitempty"`
}

type DnsDomainGetTaskDetailResponseDataDetail struct {
	Id        float64                                           `json:"id,omitempty"`
	JobId     float64                                           `json:"job_id,omitempty"`
	Domain    string                                            `json:"domain,omitempty"`
	Status    float64                                           `json:"status,omitempty"`
	Msg       string                                            `json:"msg,omitempty"`
	MsgCode   float64                                           `json:"msg_code,omitempty"`
	CreatedAt string                                            `json:"created_at,omitempty"`
	UpdatedAt string                                            `json:"updated_at,omitempty"`
	Records   []DnsDomainGetTaskDetailResponseDataDetailRecords `json:"records,omitempty"`
}

type DnsDomainGetTaskDetailResponseDataDetailRecords struct {
	OpRecordId   float64 `json:"op_record_id,omitempty"`
	OpDomainId   float64 `json:"op_domain_id,omitempty"`
	RecordName   string  `json:"record_name,omitempty"`
	RecordType   string  `json:"record_type,omitempty"`
	RecordValue  string  `json:"record_value,omitempty"`
	RecordView   string  `json:"record_view,omitempty"`
	RecordTtl    float64 `json:"record_ttl,omitempty"`
	RecordMx     float64 `json:"record_mx,omitempty"`
	Msg          string  `json:"msg,omitempty"`
	MsgCode      float64 `json:"msg_code,omitempty"`
	YurlTitle    string  `json:"yurl_title,omitempty"`
	RecordRemark string  `json:"record_remark,omitempty"`
	Status       float64 `json:"status,omitempty"`
	CreatedAt    string  `json:"created_at,omitempty"`
	UpdatedAt    string  `json:"updated_at,omitempty"`
}

type CloudDnsDomainGroupGetGroupListResponse struct {
	Status APIResponseStatus                           `json:"status"`
	Data   CloudDnsDomainGroupGetGroupListResponseData `json:"data,omitempty"`
}

type CloudDnsDomainGroupGetGroupListResponseData struct {
	Total float64                                           `json:"total,omitempty"`
	List  []CloudDnsDomainGroupGetGroupListResponseDataList `json:"list,omitempty"`
}

type CloudDnsDomainGroupGetGroupListResponseDataList struct {
	Id        float64 `json:"id,omitempty"`
	MemberId  float64 `json:"member_id,omitempty"`
	GroupName string  `json:"group_name,omitempty"`
	Remark    string  `json:"remark,omitempty"`
	CreatedAt string  `json:"created_at,omitempty"`
	UpdatedAt string  `json:"updated_at,omitempty"`
}

type CloudDnsDomainGroupAddGroupResponse struct {
	Status APIResponseStatus                       `json:"status"`
	Data   CloudDnsDomainGroupAddGroupResponseData `json:"data,omitempty"`
}

type CloudDnsDomainGroupAddGroupResponseData struct {
	Id float64 `json:"id,omitempty"`
}

type CloudDnsDomainGroupUpdateGroupResponse struct {
	Status APIResponseStatus      `json:"status"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

type CloudDnsDomainGroupDeleteGroupResponse struct {
	Status APIResponseStatus      `json:"status"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

type CloudDnsDomainGroupGetGroupRecordListResponse struct {
	Status APIResponseStatus                                 `json:"status"`
	Data   CloudDnsDomainGroupGetGroupRecordListResponseData `json:"data,omitempty"`
}

type CloudDnsDomainGroupGetGroupRecordListResponseData struct {
	Total float64                                                 `json:"total,omitempty"`
	List  []CloudDnsDomainGroupGetGroupRecordListResponseDataList `json:"list,omitempty"`
}

type CloudDnsDomainGroupGetGroupRecordListResponseDataList struct {
	DomainId float64 `json:"domain_id,omitempty"`
	Domain   string  `json:"domain,omitempty"`
}

type CloudDnsDomainGroupSaveDomainToGroupResponse struct {
	Status APIResponseStatus      `json:"status"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

type CloudDnsDomainGroupGetGroupDomainListResponse struct {
	Status APIResponseStatus                                 `json:"status"`
	Data   CloudDnsDomainGroupGetGroupDomainListResponseData `json:"data,omitempty"`
}

type CloudDnsDomainGroupGetGroupDomainListResponseData struct {
	Total string                                                  `json:"total,omitempty"`
	List  []CloudDnsDomainGroupGetGroupDomainListResponseDataList `json:"list,omitempty"`
}

type CloudDnsDomainGroupGetGroupDomainListResponseDataList struct {
	DomainId string `json:"domain_id,omitempty"`
	Domain   string `json:"domain,omitempty"`
}

type CloudDnsDomainGroupGetGroupUndistributedDomainListResponse struct {
	Status APIResponseStatus                                              `json:"status"`
	Data   CloudDnsDomainGroupGetGroupUndistributedDomainListResponseData `json:"data,omitempty"`
}

type CloudDnsDomainGroupGetGroupUndistributedDomainListResponseData struct {
	Total string                                                               `json:"total,omitempty"`
	List  []CloudDnsDomainGroupGetGroupUndistributedDomainListResponseDataList `json:"list,omitempty"`
}

type CloudDnsDomainGroupGetGroupUndistributedDomainListResponseDataList struct {
	DomainId string `json:"domain_id,omitempty"`
	Domain   string `json:"domain,omitempty"`
}

type DnsDomainRecordsGetRecordTypesResponse struct {
	Status APIResponseStatus      `json:"status"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

type DnsDomainRecordsGetRecordListResponse struct {
	Status APIResponseStatus                         `json:"status"`
	Data   DnsDomainRecordsGetRecordListResponseData `json:"data,omitempty"`
}

type DnsDomainRecordsGetRecordListResponseData struct {
	Total float64                                         `json:"total,omitempty"`
	List  []DnsDomainRecordsGetRecordListResponseDataList `json:"list,omitempty"`
}

type DnsDomainRecordsGetRecordListResponseDataList struct {
	Id        float64 `json:"id,omitempty"`
	DomainId  float64 `json:"domain_id,omitempty"`
	Name      string  `json:"name,omitempty"`
	TypeValue string  `json:"type,omitempty"`
	View      string  `json:"view,omitempty"`
	Value     string  `json:"value,omitempty"`
	Mx        float64 `json:"mx,omitempty"`
	Ttl       float64 `json:"ttl,omitempty"`
	Status    string  `json:"status,omitempty"`
	Sort      float64 `json:"sort,omitempty"`
	IsSyncCdn float64 `json:"is_sync_cdn,omitempty"`
	Remark    string  `json:"remark,omitempty"`
	Locked    bool    `json:"locked,omitempty"`
	CreatedAt string  `json:"created_at,omitempty"`
	UpdatedAt string  `json:"updated_at,omitempty"`
}

type DnsDomainRecordsAddRecordResponse struct {
	Status APIResponseStatus                     `json:"status"`
	Data   DnsDomainRecordsAddRecordResponseData `json:"data,omitempty"`
}

type DnsDomainRecordsAddRecordResponseData struct {
	Id float64 `json:"id,omitempty"`
}

type DnsDomainRecordsBatchAddRecordsResponse struct {
	Status APIResponseStatus                           `json:"status"`
	Data   DnsDomainRecordsBatchAddRecordsResponseData `json:"data,omitempty"`
}

type DnsDomainRecordsBatchAddRecordsResponseData struct {
	JobId float64 `json:"job_id,omitempty"`
}

type DnsDomainRecordsEditRecordResponse struct {
	Status APIResponseStatus      `json:"status"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

type DnsDomainRecordsBatchPauseRecordsResponse struct {
	Status APIResponseStatus      `json:"status"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

type DnsDomainRecordsBatchEnableRecordsResponse struct {
	Status APIResponseStatus      `json:"status"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

type DnsDomainRecordsDeleteRecordResponse struct {
	Status APIResponseStatus      `json:"status"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

type DnsDomainRecordsImportRecordsResponse struct {
	Status APIResponseStatus                         `json:"status"`
	Data   DnsDomainRecordsImportRecordsResponseData `json:"data,omitempty"`
}

type DnsDomainRecordsImportRecordsResponseData struct {
	JobId float64 `json:"job_id,omitempty"`
}

type DnsDomainRecordsExportRecordsResponse struct {
	Status APIResponseStatus                         `json:"status"`
	Data   DnsDomainRecordsExportRecordsResponseData `json:"data,omitempty"`
}

type DnsDomainRecordsExportRecordsResponseData struct {
	RealUrl string `json:"real_url,omitempty"`
}

type DnsDomainRecordsGetLinesResponse struct {
	Status APIResponseStatus                    `json:"status"`
	Data   DnsDomainRecordsGetLinesResponseData `json:"data,omitempty"`
}

type DnsDomainRecordsGetLinesResponseData struct {
	Lines []DnsDomainRecordsGetLinesResponseDataLines `json:"lines,omitempty"`
}

type DnsDomainRecordsGetLinesResponseDataLines struct {
	Id            float64 `json:"id,omitempty"`
	Name          string  `json:"name,omitempty"`
	Desc          string  `json:"desc,omitempty"`
	Key           string  `json:"key,omitempty"`
	Sort          float64 `json:"sort,omitempty"`
	IsVirtual     float64 `json:"is_virtual,omitempty"`
	VirtualConfig string  `json:"virtual_config,omitempty"`
	Status        float64 `json:"status,omitempty"`
	ParentId      float64 `json:"parent_id,omitempty"`
	Depth         float64 `json:"depth,omitempty"`
	CreatedAt     string  `json:"created_at,omitempty"`
	UpdatedAt     string  `json:"updated_at,omitempty"`
}

type DnsDomainRecordsBatchDeleteRecordsResponse struct {
	Status APIResponseStatus      `json:"status"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

type DnsDomainRecordsGetRecordGroupsListResponse struct {
	Status APIResponseStatus                               `json:"status"`
	Data   DnsDomainRecordsGetRecordGroupsListResponseData `json:"data,omitempty"`
}

type DnsDomainRecordsGetRecordGroupsListResponseData struct {
	Total float64                                               `json:"total,omitempty"`
	List  []DnsDomainRecordsGetRecordGroupsListResponseDataList `json:"list,omitempty"`
}

type DnsDomainRecordsGetRecordGroupsListResponseDataList struct {
	Id        float64 `json:"id,omitempty"`
	MemberId  float64 `json:"member_id,omitempty"`
	GroupName string  `json:"group_name,omitempty"`
	DomainId  float64 `json:"domain_id,omitempty"`
	CreatedAt string  `json:"created_at,omitempty"`
	UpdatedAt string  `json:"updated_at,omitempty"`
}

type DnsDomainRecordsAddRecordGroupResponse struct {
	Status APIResponseStatus                          `json:"status"`
	Data   DnsDomainRecordsAddRecordGroupResponseData `json:"data,omitempty"`
}

type DnsDomainRecordsAddRecordGroupResponseData struct {
	Id float64 `json:"id,omitempty"`
}

type DnsDomainRecordsAddRecordGroupRelationsResponse struct {
	Status APIResponseStatus      `json:"status"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

type DnsDomainRecordsDeleteRecordGroupResponse struct {
	Status APIResponseStatus      `json:"status"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

type UserIpUserIpListResponse struct {
	Status APIResponseStatus            `json:"status"`
	Data   UserIpUserIpListResponseData `json:"data,omitempty"`
}

type UserIpUserIpListResponseData struct {
	Total string                             `json:"total,omitempty"`
	List  []UserIpUserIpListResponseDataList `json:"list,omitempty"`
}

type UserIpUserIpListResponseDataList struct {
	Id          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	MemberId    string `json:"member_id,omitempty"`
	SubUserId   string `json:"sub_user_id,omitempty"`
	ItemNum     string `json:"item_num,omitempty"`
	Remark      string `json:"remark,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
	FileKey     string `json:"file_key,omitempty"`
	FileVersion string `json:"file_version,omitempty"`
	WriteMmdb   string `json:"write_mmdb,omitempty"`
	FileError   string `json:"file_error,omitempty"`
	MqTime      string `json:"mq_time,omitempty"`
	OwnerName   string `json:"owner_name,omitempty"`
}

type UserIpUserIpAddResponse struct {
	Status APIResponseStatus           `json:"status"`
	Data   UserIpUserIpAddResponseData `json:"data,omitempty"`
}

type UserIpUserIpAddResponseData struct {
	Id string `json:"id,omitempty"`
}

type UserIpUserIpSaveResponse struct {
	Status APIResponseStatus            `json:"status"`
	Data   UserIpUserIpSaveResponseData `json:"data,omitempty"`
}

type UserIpUserIpSaveResponseData struct {
	Id float64 `json:"id,omitempty"`
}

type UserIpUserIpDelResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type UserIpListUserIpItemResponse struct {
	Status APIResponseStatus                `json:"status"`
	Data   UserIpListUserIpItemResponseData `json:"data,omitempty"`
}

type UserIpListUserIpItemResponseData struct {
	Total float64                                `json:"total,omitempty"`
	List  []UserIpListUserIpItemResponseDataList `json:"list,omitempty"`
}

type UserIpListUserIpItemResponseDataList struct {
	Id              string  `json:"_id,omitempty"`
	Ip              string  `json:"ip,omitempty"`
	Remark          string  `json:"remark,omitempty"`
	UserIpId        float64 `json:"user_ip_id,omitempty"`
	FormatCreatedAt string  `json:"format_created_at,omitempty"`
	FormatUpdatedAt string  `json:"format_updated_at,omitempty"`
}

type UserIpAddUserIpItemResponse struct {
	Status APIResponseStatus               `json:"status"`
	Data   UserIpAddUserIpItemResponseData `json:"data,omitempty"`
}

type UserIpAddUserIpItemResponseData struct {
	Ids []string `json:"ids,omitempty"`
}

type UserIpUpdateUserIpItemResponse struct {
	Status APIResponseStatus                  `json:"status"`
	Data   UserIpUpdateUserIpItemResponseData `json:"data,omitempty"`
}

type UserIpUpdateUserIpItemResponseData struct {
	Id string `json:"id,omitempty"`
}

type UserIpBatchDeleteUserIpItemResponse struct {
	Status APIResponseStatus                       `json:"status"`
	Data   UserIpBatchDeleteUserIpItemResponseData `json:"data,omitempty"`
}

type UserIpBatchDeleteUserIpItemResponseData struct {
	Id float64 `json:"id,omitempty"`
}

type UserIpDeleteAllUserIpItemResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type UserIpCopyUserIpResponse struct {
	Status APIResponseStatus            `json:"status"`
	Data   UserIpCopyUserIpResponseData `json:"data,omitempty"`
}

type UserIpCopyUserIpResponseData struct {
	Id float64 `json:"id,omitempty"`
}

type UserIpFileSaveIpItemResponse struct {
	Status APIResponseStatus                `json:"status"`
	Data   UserIpFileSaveIpItemResponseData `json:"data,omitempty"`
}

type UserIpFileSaveIpItemResponseData struct {
	Ids []string `json:"ids,omitempty"`
}

type ServiceBatchListTaskResponse struct {
	Status APIResponseStatus                `json:"status"`
	Data   ServiceBatchListTaskResponseData `json:"data,omitempty"`
}

type ServiceBatchListTaskResponseData struct {
	Total float64                              `json:"total,omitempty"`
	List  ServiceBatchListTaskResponseDataList `json:"list,omitempty"`
}

type ServiceBatchListTaskResponseDataList struct {
	Id              string      `json:"_id,omitempty"`
	Bank            string      `json:"bank,omitempty"`
	ApiName         string      `json:"api_name,omitempty"`
	UserId          float64     `json:"user_id,omitempty"`
	SubuserId       float64     `json:"subuser_id,omitempty"`
	Params          interface{} `json:"params,omitempty"`
	ExecStatus      float64     `json:"exec_status,omitempty"`
	ExecStart       string      `json:"exec_start,omitempty"`
	ExecStartFormat string      `json:"exec_start_format,omitempty"`
	ExecEnd         string      `json:"exec_end,omitempty"`
	ExecEndFormat   string      `json:"exec_end_format,omitempty"`
	TotalDomain     float64     `json:"total_domain,omitempty"`
	Total           float64     `json:"total,omitempty"`
	TotalOk         float64     `json:"total_ok,omitempty"`
	TotalFail       float64     `json:"total_fail,omitempty"`
	CreateAt        string      `json:"create_at,omitempty"`
	CreateAtFormat  string      `json:"create_at_format,omitempty"`
	UpdateAt        string      `json:"update_at,omitempty"`
	UpdateAtFormat  string      `json:"update_at_format,omitempty"`
}

type ServiceBatchListSubTaskResponse struct {
	Status APIResponseStatus                   `json:"status"`
	Data   ServiceBatchListSubTaskResponseData `json:"data,omitempty"`
}

type ServiceBatchListSubTaskResponseData struct {
	Total float64                                 `json:"total,omitempty"`
	List  ServiceBatchListSubTaskResponseDataList `json:"list,omitempty"`
}

type ServiceBatchListSubTaskResponseDataList struct {
	Id              string                 `json:"_id,omitempty"`
	Pid             string                 `json:"pid,omitempty"`
	Bank            string                 `json:"bank,omitempty"`
	ApiName         string                 `json:"api_name,omitempty"`
	UserId          float64                `json:"user_id,omitempty"`
	SubuserId       float64                `json:"subuser_id,omitempty"`
	Params          map[string]interface{} `json:"params,omitempty"`
	Request         map[string]interface{} `json:"request,omitempty"`
	Response        string                 `json:"response,omitempty"`
	ExecMsg         string                 `json:"exec_msg,omitempty"`
	ExecStatus      float64                `json:"exec_status,omitempty"`
	ExecError       string                 `json:"exec_error,omitempty"`
	ExecStart       string                 `json:"exec_start,omitempty"`
	ExecStartFormat string                 `json:"exec_start_format,omitempty"`
	ExecEnd         string                 `json:"exec_end,omitempty"`
	ExecEndFormat   string                 `json:"exec_end_format,omitempty"`
	CreateAt        string                 `json:"create_at,omitempty"`
	CreateAtFormat  string                 `json:"create_at_format,omitempty"`
	UpdateAt        string                 `json:"update_at,omitempty"`
	UpdateAtFormat  string                 `json:"update_at_format,omitempty"`
}

type WebCdnCleanCacheGetCacheListResponse struct {
	Status APIResponseStatus                        `json:"status"`
	Data   WebCdnCleanCacheGetCacheListResponseData `json:"data,omitempty"`
}

type WebCdnCleanCacheGetCacheListResponseData struct {
	Id               float64  `json:"id,omitempty"`
	Wholesite        []string `json:"wholesite,omitempty"`
	Specialurl       []string `json:"specialurl,omitempty"`
	Specialdir       []string `json:"specialdir,omitempty"`
	AvailableUrl     float64  `json:"available_url,omitempty"`
	MaximumLimitUrl  float64  `json:"maximum_limit_url,omitempty"`
	AvailableSite    float64  `json:"available_site,omitempty"`
	MaximumLimitSite float64  `json:"maximum_limit_site,omitempty"`
	AvailableDir     float64  `json:"available_dir,omitempty"`
	MaximumLimitDir  float64  `json:"maximum_limit_dir,omitempty"`
}

type WebCdnCleanCacheSaveCacheResponse struct {
	Status APIResponseStatus                     `json:"status"`
	Data   WebCdnCleanCacheSaveCacheResponseData `json:"data,omitempty"`
}

type WebCdnCleanCacheSaveCacheResponseData struct {
	Wholesite  []string `json:"wholesite,omitempty"`
	Specialurl []string `json:"specialurl,omitempty"`
	Specialdir []string `json:"specialdir,omitempty"`
	TaskId     float64  `json:"task_id,omitempty"`
}

type WebCdnCleanCacheGetTaskListResponse struct {
	Status APIResponseStatus                       `json:"status"`
	Data   WebCdnCleanCacheGetTaskListResponseData `json:"data,omitempty"`
}

type WebCdnCleanCacheGetTaskListResponseData struct {
	Total float64                                       `json:"total,omitempty"`
	List  []WebCdnCleanCacheGetTaskListResponseDataList `json:"list,omitempty"`
}

type WebCdnCleanCacheGetTaskListResponseDataList struct {
	SubType   string `json:"sub_type,omitempty"`
	Total     string `json:"total,omitempty"`
	Succeed   string `json:"succeed,omitempty"`
	Failed    string `json:"failed,omitempty"`
	Ongoing   string `json:"ongoing,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
	TaskId    string `json:"task_id,omitempty"`
	Status    string `json:"status,omitempty"`
}

type WebCdnCleanCacheGetTaskDetailResponse struct {
	Status APIResponseStatus                         `json:"status"`
	Data   WebCdnCleanCacheGetTaskDetailResponseData `json:"data,omitempty"`
}

type WebCdnCleanCacheGetTaskDetailResponseData struct {
	Total float64                                         `json:"total,omitempty"`
	List  []WebCdnCleanCacheGetTaskDetailResponseDataList `json:"list,omitempty"`
}

type WebCdnCleanCacheGetTaskDetailResponseDataList struct {
	Result    string `json:"result,omitempty"`
	Message   string `json:"message,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
	Directory string `json:"directory,omitempty"`
	Subdomain string `json:"subdomain,omitempty"`
	Url       string `json:"url,omitempty"`
}

type WebCdnPreheatCacheGetPreheatCacheQuotaResponse struct {
	Status APIResponseStatus                                  `json:"status"`
	Data   WebCdnPreheatCacheGetPreheatCacheQuotaResponseData `json:"data,omitempty"`
}

type WebCdnPreheatCacheGetPreheatCacheQuotaResponseData struct {
	Available    float64 `json:"available,omitempty"`
	MaximumLimit float64 `json:"maximum_limit,omitempty"`
}

type WebCdnPreheatCacheGetPreheatCacheListResponse struct {
	Status APIResponseStatus                                 `json:"status"`
	Data   WebCdnPreheatCacheGetPreheatCacheListResponseData `json:"data,omitempty"`
}

type WebCdnPreheatCacheGetPreheatCacheListResponseData struct {
	Total     float64                                                    `json:"total,omitempty"`
	StatusMap WebCdnPreheatCacheGetPreheatCacheListResponseDataStatusMap `json:"status_map,omitempty"`
	List      []WebCdnPreheatCacheGetPreheatCacheListResponseDataList    `json:"list,omitempty"`
}

type WebCdnPreheatCacheGetPreheatCacheListResponseDataStatusMap struct {
	Api1 string `json:"1,omitempty"`
	Api2 string `json:"2,omitempty"`
	Api3 string `json:"3,omitempty"`
	Api4 string `json:"4,omitempty"`
}

type WebCdnPreheatCacheGetPreheatCacheListResponseDataList struct {
	Id               float64 `json:"id,omitempty"`
	UserId           float64 `json:"user_id,omitempty"`
	TimeCreate       string  `json:"time_create,omitempty"`
	TimeUpdate       string  `json:"time_update,omitempty"`
	TaskId           float64 `json:"task_id,omitempty"`
	DomainId         float64 `json:"domain_id,omitempty"`
	Url              string  `json:"url,omitempty"`
	Status           float64 `json:"status,omitempty"`
	Total            float64 `json:"total,omitempty"`
	Weight           float64 `json:"weight,omitempty"`
	SubUserId        float64 `json:"sub_user_id,omitempty"`
	UserName         string  `json:"user_name,omitempty"`
	StrategyId       float64 `json:"strategy_id,omitempty"`
	Strategy         string  `json:"strategy,omitempty"`
	OperatorUserName string  `json:"operator_user_name,omitempty"`
}

type WebCdnPreheatCacheSavePreheatCacheResponse struct {
	Status APIResponseStatus      `json:"status"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

type OplogInfoResponse struct {
	Status APIResponseStatus     `json:"status"`
	Data   OplogInfoResponseData `json:"data,omitempty"`
}

type OplogInfoResponseData struct {
	Id_          string                 `json:"_id,omitempty"`
	Id           string                 `json:"id,omitempty"`
	BizType      string                 `json:"biz_type,omitempty"`
	OpType       string                 `json:"op_type,omitempty"`
	OpAdminEmail string                 `json:"op_admin_email,omitempty"`
	OpAdminId    float64                `json:"op_admin_id,omitempty"`
	OpAdminName  string                 `json:"op_admin_name,omitempty"`
	OpUid        float64                `json:"op_uid,omitempty"`
	OpUserEmail  string                 `json:"op_user_email,omitempty"`
	OpUsername   string                 `json:"op_username,omitempty"`
	Result       string                 `json:"result,omitempty"`
	Uid          float64                `json:"uid,omitempty"`
	UserEmail    string                 `json:"user_email,omitempty"`
	Username     string                 `json:"username,omitempty"`
	SubuserId    float64                `json:"subuser_id,omitempty"`
	SubuserName  string                 `json:"subuser_name,omitempty"`
	Url          string                 `json:"url,omitempty"`
	Ip           string                 `json:"ip,omitempty"`
	ClientPort   string                 `json:"client_port,omitempty"`
	Content      string                 `json:"content,omitempty"`
	OpData       map[string]interface{} `json:"op_data,omitempty"`
	DataKey      map[string]interface{} `json:"data_key,omitempty"`
	Params       map[string]interface{} `json:"params,omitempty"`
	CreateAt     map[string]interface{} `json:"create_at,omitempty"`
	UpdateAt     map[string]interface{} `json:"update_at,omitempty"`
}

type OplogMapResponse struct {
	Status APIResponseStatus    `json:"status"`
	Data   OplogMapResponseData `json:"data,omitempty"`
}

type OplogMapResponseData struct {
	DataKeyMap map[string]interface{} `json:"data_key_map,omitempty"`
	Parts      float64                `json:"parts,omitempty"`
	BizTypes   interface{}            `json:"biz_types,omitempty"`
	OpTypes    interface{}            `json:"op_types,omitempty"`
	Help       interface{}            `json:"help,omitempty"`
}

type OplogGetOplogsResponse struct {
	Status APIResponseStatus          `json:"status"`
	Data   OplogGetOplogsResponseData `json:"data,omitempty"`
}

type OplogGetOplogsResponseData struct {
	DataKeyMap map[string]interface{}         `json:"data_key_map,omitempty"`
	Total      float64                        `json:"total,omitempty"`
	TotalDesc  float64                        `json:"total_desc,omitempty"`
	List       OplogGetOplogsResponseDataList `json:"list,omitempty"`
}

type OplogGetOplogsResponseDataList struct {
	Id_          string                 `json:"_id,omitempty"`
	Id           string                 `json:"id,omitempty"`
	BizType      string                 `json:"biz_type,omitempty"`
	OpType       string                 `json:"op_type,omitempty"`
	OpAdminEmail string                 `json:"op_admin_email,omitempty"`
	OpAdminId    float64                `json:"op_admin_id,omitempty"`
	OpAdminName  string                 `json:"op_admin_name,omitempty"`
	OpUid        float64                `json:"op_uid,omitempty"`
	OpUserEmail  string                 `json:"op_user_email,omitempty"`
	OpUsername   string                 `json:"op_username,omitempty"`
	Result       string                 `json:"result,omitempty"`
	Uid          float64                `json:"uid,omitempty"`
	UserEmail    string                 `json:"user_email,omitempty"`
	Username     string                 `json:"username,omitempty"`
	SubuserId    float64                `json:"subuser_id,omitempty"`
	SubuserName  string                 `json:"subuser_name,omitempty"`
	Url          string                 `json:"url,omitempty"`
	Ip           string                 `json:"ip,omitempty"`
	ClientPort   string                 `json:"client_port,omitempty"`
	Content      string                 `json:"content,omitempty"`
	OpData       map[string]interface{} `json:"op_data,omitempty"`
	DataKey      map[string]interface{} `json:"data_key,omitempty"`
	Params       map[string]interface{} `json:"params,omitempty"`
	CreateAt     map[string]interface{} `json:"create_at,omitempty"`
	UpdateAt     map[string]interface{} `json:"update_at,omitempty"`
}

type CaCertificateSelfAddCaResponse struct {
	Status APIResponseStatus                  `json:"status"`
	Data   CaCertificateSelfAddCaResponseData `json:"data,omitempty"`
}

type CaCertificateSelfAddCaResponseData struct {
	Id   string `json:"id,omitempty"`
	CaSn string `json:"ca_sn,omitempty"`
}

type BatchCaListResponse struct {
	Status APIResponseStatus       `json:"status"`
	Data   BatchCaListResponseData `json:"data,omitempty"`
}

type BatchCaListResponseData struct {
	Id               string `json:"id,omitempty"`
	MemberId         string `json:"member_id,omitempty"`
	CaSn             string `json:"ca_sn,omitempty"`
	CaName           string `json:"ca_name,omitempty"`
	IssuerStartTime  string `json:"issuer_start_time,omitempty"`
	IssuerExpiryTime string `json:"issuer_expiry_time,omitempty"`
}

type CaCertificateSelfSaveTextCaInfoResponse struct {
	Status APIResponseStatus                           `json:"status"`
	Data   CaCertificateSelfSaveTextCaInfoResponseData `json:"data,omitempty"`
}

type CaCertificateSelfSaveTextCaInfoResponseData struct {
	Id   string `json:"id,omitempty"`
	CaSn string `json:"ca_sn,omitempty"`
}

type CaCertificateSelfEditCaInfoResponse struct {
	Status APIResponseStatus                       `json:"status"`
	Data   CaCertificateSelfEditCaInfoResponseData `json:"data,omitempty"`
}

type CaCertificateSelfEditCaInfoResponseData struct {
	Id   string `json:"id,omitempty"`
	CaSn string `json:"ca_sn,omitempty"`
}

type CaCertificateSelfListCaResponse struct {
	Status APIResponseStatus                   `json:"status"`
	Data   CaCertificateSelfListCaResponseData `json:"data,omitempty"`
}

type CaCertificateSelfListCaResponseData struct {
	Total      float64                                   `json:"total,omitempty"`
	IssuerList []string                                  `json:"issuer_list,omitempty"`
	List       []CaCertificateSelfListCaResponseDataList `json:"list,omitempty"`
}

type CaCertificateSelfListCaResponseDataList struct {
	Id                              string   `json:"id,omitempty"`
	MemberId                        string   `json:"member_id,omitempty"`
	CaName                          string   `json:"ca_name,omitempty"`
	Issuer                          string   `json:"issuer,omitempty"`
	IssuerStartTime                 string   `json:"issuer_start_time,omitempty"`
	IssuerExpiryTime                string   `json:"issuer_expiry_time,omitempty"`
	IssuerExpiryTimeDesc            string   `json:"issuer_expiry_time_desc,omitempty"`
	IssuerExpiryTimeAutoRenewStatus string   `json:"issuer_expiry_time_auto_renew_status,omitempty"`
	RenewStatus                     string   `json:"renew_status,omitempty"`
	Binded                          bool     `json:"binded,omitempty"`
	CaDomain                        []string `json:"ca_domain,omitempty"`
	ApplyStatus                     string   `json:"apply_status,omitempty"`
	CaType                          string   `json:"ca_type,omitempty"`
	CaTypeDomain                    string   `json:"ca_type_domain,omitempty"`
	Code                            string   `json:"code,omitempty"`
	Msg                             string   `json:"msg,omitempty"`
}

type CaCertificateSelfCaExportResponse struct {
	Status APIResponseStatus                       `json:"status"`
	Data   []CaCertificateSelfCaExportResponseData `json:"data,omitempty"`
}

type CaCertificateSelfCaExportResponseData struct {
	Hash    string `json:"hash,omitempty"`
	Key     string `json:"key,omitempty"`
	RealUrl string `json:"real_url,omitempty"`
}

type CaCertificateSelfBatchOperatSslResponse struct {
	Status APIResponseStatus                           `json:"status"`
	Data   CaCertificateSelfBatchOperatSslResponseData `json:"data,omitempty"`
}

type CaCertificateSelfBatchOperatSslResponseData struct {
	Total     float64 `json:"total,omitempty"`
	FailTotal float64 `json:"fail_total,omitempty"`
	FailList  float64 `json:"fail_list,omitempty"`
	Remark    string  `json:"remark,omitempty"`
}

type CaCertificateSelfDelCaResponse struct {
	Status APIResponseStatus                  `json:"status"`
	Data   CaCertificateSelfDelCaResponseData `json:"data,omitempty"`
}

type CaCertificateSelfDelCaResponseData struct {
	Info string `json:"info,omitempty"`
}

type CaCertificateSelfGetCaDetailResponse struct {
	Status APIResponseStatus                        `json:"status"`
	Data   CaCertificateSelfGetCaDetailResponseData `json:"data,omitempty"`
}

type CaCertificateSelfGetCaDetailResponseData struct {
	Id                              string   `json:"id,omitempty"`
	CaId                            string   `json:"ca_id,omitempty"`
	MemberId                        string   `json:"member_id,omitempty"`
	CaName                          string   `json:"ca_name,omitempty"`
	Issuer                          string   `json:"issuer,omitempty"`
	IssuerStartTime                 string   `json:"issuer_start_time,omitempty"`
	IssuerExpiryTime                string   `json:"issuer_expiry_time,omitempty"`
	IssuerExpiryTimeDesc            string   `json:"issuer_expiry_time_desc,omitempty"`
	IssuerExpiryTimeAutoRenewStatus string   `json:"issuer_expiry_time_auto_renew_status,omitempty"`
	RenewStatus                     string   `json:"renew_status,omitempty"`
	Binded                          bool     `json:"binded,omitempty"`
	CaDomain                        []string `json:"ca_domain,omitempty"`
	ApplyStatus                     string   `json:"apply_status,omitempty"`
	CaType                          string   `json:"ca_type,omitempty"`
	CaTypeDomain                    string   `json:"ca_type_domain,omitempty"`
	Code                            string   `json:"code,omitempty"`
	Msg                             string   `json:"msg,omitempty"`
	CreatedAt                       string   `json:"created_at,omitempty"`
	UpdatedAt                       string   `json:"updated_at,omitempty"`
	IssuerOrganization              string   `json:"issuer_organization,omitempty"`
	IssuerOrganizationElement       string   `json:"issuer_organization_element,omitempty"`
	SerialNumber                    string   `json:"serial_number,omitempty"`
	IssuerObject                    string   `json:"issuer_object,omitempty"`
	UseOrganization                 string   `json:"use_organization,omitempty"`
	UseOrganizationElement          string   `json:"use_organization_element,omitempty"`
	City                            string   `json:"city,omitempty"`
	Province                        string   `json:"province,omitempty"`
	Country                         string   `json:"country,omitempty"`
	AuthenticationUsableDomain      string   `json:"authentication_usable_domain,omitempty"`
}

type CaCertificateSelfEditCaNameResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type CaCertificateApplyAddApplyCaResponse struct {
	Status APIResponseStatus                        `json:"status"`
	Data   CaCertificateApplyAddApplyCaResponseData `json:"data,omitempty"`
}

type CaCertificateApplyAddApplyCaResponseData struct {
	CaIdDomains map[string]interface{} `json:"ca_id_domains,omitempty"`
	CaIdNames   map[string]interface{} `json:"ca_id_names,omitempty"`
}

type CaCertificateApplyGetAddByNsSettingResponse struct {
	Status APIResponseStatus                               `json:"status"`
	Data   CaCertificateApplyGetAddByNsSettingResponseData `json:"data,omitempty"`
}

type CaCertificateApplyGetAddByNsSettingResponseData struct {
	Name      string   `json:"name,omitempty"`
	TypeValue string   `json:"type,omitempty"`
	Ns        []string `json:"ns,omitempty"`
}

type DomainGroupSaveGroupResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type DomainGroupGetGroupListResponse struct {
	Status APIResponseStatus                   `json:"status"`
	Data   DomainGroupGetGroupListResponseData `json:"data,omitempty"`
}

type DomainGroupGetGroupListResponseData struct {
	Total string                                    `json:"total,omitempty"`
	List  []DomainGroupGetGroupListResponseDataList `json:"list,omitempty"`
}

type DomainGroupGetGroupListResponseDataList struct {
	Id          string `json:"id,omitempty"`
	MemberId    string `json:"member_id,omitempty"`
	ProductFlag string `json:"product_flag,omitempty"`
	GroupName   string `json:"group_name,omitempty"`
	Remark      string `json:"remark,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

type DomainGroupDelGroupResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type DomainGroupGetGroupDomainListResponse struct {
	Status APIResponseStatus                         `json:"status"`
	Data   DomainGroupGetGroupDomainListResponseData `json:"data,omitempty"`
}

type DomainGroupGetGroupDomainListResponseData struct {
	Total string                                          `json:"total,omitempty"`
	Ports []string                                        `json:"ports,omitempty"`
	List  []DomainGroupGetGroupDomainListResponseDataList `json:"list,omitempty"`
}

type DomainGroupGetGroupDomainListResponseDataList struct {
	DomainId string `json:"domain_id,omitempty"`
	Domain   string `json:"domain,omitempty"`
}

type DomainGroupGgtUndistributedDomainListResponse struct {
	Status APIResponseStatus                                 `json:"status"`
	Data   DomainGroupGgtUndistributedDomainListResponseData `json:"data,omitempty"`
}

type DomainGroupGgtUndistributedDomainListResponseData struct {
	Total string                                                  `json:"total,omitempty"`
	List  []DomainGroupGgtUndistributedDomainListResponseDataList `json:"list,omitempty"`
}

type DomainGroupGgtUndistributedDomainListResponseDataList struct {
	DomainId string `json:"domain_id,omitempty"`
	Domain   string `json:"domain,omitempty"`
}

type DomainGroupAddGroupResponse struct {
	Status APIResponseStatus               `json:"status"`
	Data   DomainGroupAddGroupResponseData `json:"data,omitempty"`
}

type DomainGroupAddGroupResponseData struct {
	Id string `json:"id,omitempty"`
}

type DomainGroupSaveDomainToGroupResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type DomainGroupGetGroupInfoResponse struct {
	Status APIResponseStatus                   `json:"status"`
	Data   DomainGroupGetGroupInfoResponseData `json:"data,omitempty"`
}

type DomainGroupGetGroupInfoResponseData struct {
	Id          string `json:"id,omitempty"`
	MemberId    string `json:"member_id,omitempty"`
	ProductFlag string `json:"product_flag,omitempty"`
	GroupName   string `json:"group_name,omitempty"`
	Remark      string `json:"remark,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

type DomainGroupMoveDomainResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type ListDomainsResponse struct {
	Status APIResponseStatus       `json:"status"`
	Data   ListDomainsResponseData `json:"data,omitempty"`
}

type ListDomainsResponseData struct {
	Total float64                       `json:"total,omitempty"`
	List  []ListDomainsResponseDataList `json:"list,omitempty"`
}

type ListDomainsResponseDataList struct {
	Id                  float64                          `json:"id,omitempty"`
	Domain              string                           `json:"domain,omitempty"`
	Remark              string                           `json:"remark,omitempty"`
	AccessProgress      string                           `json:"access_progress,omitempty"`
	AccessMode          string                           `json:"access_mode,omitempty"`
	ProtectStatus       string                           `json:"protect_status,omitempty"`
	EiForwardStatus     string                           `json:"ei_forward_status,omitempty"`
	Cname               ListDomainsResponseDataListCname `json:"cname,omitempty"`
	UseMyCname          float64                          `json:"use_my_cname,omitempty"`
	UseMyDns            float64                          `json:"use_my_dns,omitempty"`
	CaStatus            string                           `json:"ca_status,omitempty"`
	ExclusiveResourceId float64                          `json:"exclusive_resource_id,omitempty"`
	AccessProgressDesc  string                           `json:"access_progress_desc,omitempty"`
	HasOrigin           bool                             `json:"has_origin,omitempty"`
	CaId                float64                          `json:"ca_id,omitempty"`
	CreatedAt           string                           `json:"created_at,omitempty"`
	UpdatedAt           string                           `json:"updated_at,omitempty"`
	PriDomain           string                           `json:"pri_domain,omitempty"`
}

type ListDomainsResponseDataListCname struct {
	Master string   `json:"master,omitempty"`
	Slaves []string `json:"slaves,omitempty"`
}

type AddDomainsResponse struct {
	Status APIResponseStatus      `json:"status"`
	Data   AddDomainsResponseData `json:"data,omitempty"`
}

type AddDomainsResponseData struct {
	Id         float64 `json:"id,omitempty"`
	Record     string  `json:"record,omitempty"`
	Cname      string  `json:"cname,omitempty"`
	RecordType string  `json:"record_type,omitempty"`
	Domain     string  `json:"domain,omitempty"`
	PriDomain  string  `json:"pri_domain,omitempty"`
}

type UpdateDomainsResponse struct {
	Status APIResponseStatus         `json:"status"`
	Data   UpdateDomainsResponseData `json:"data,omitempty"`
}

type UpdateDomainsResponseData struct {
	DomainId float64 `json:"domain_id,omitempty"`
}

type BindDomainCertResponse struct {
	Status APIResponseStatus          `json:"status"`
	Data   BindDomainCertResponseData `json:"data,omitempty"`
}

type BindDomainCertResponseData struct {
	CaId float64 `json:"ca_id,omitempty"`
}

type UnBindDomainCertResponse struct {
	Status APIResponseStatus            `json:"status"`
	Data   UnBindDomainCertResponseData `json:"data,omitempty"`
}

type UnBindDomainCertResponseData struct {
	CaId float64 `json:"ca_id,omitempty"`
}

type DeleteDomainsResponse struct {
	Status APIResponseStatus         `json:"status"`
	Data   DeleteDomainsResponseData `json:"data,omitempty"`
}

type DeleteDomainsResponseData struct {
	Ids []float64 `json:"ids,omitempty"`
}

type DisableDomainsResponse struct {
	Status APIResponseStatus          `json:"status"`
	Data   DisableDomainsResponseData `json:"data,omitempty"`
}

type DisableDomainsResponseData struct {
	DomainIds []float64 `json:"domain_ids,omitempty"`
}

type EnableDomainsResponse struct {
	Status APIResponseStatus         `json:"status"`
	Data   EnableDomainsResponseData `json:"data,omitempty"`
}

type EnableDomainsResponseData struct {
	DomainIds []float64 `json:"domain_ids,omitempty"`
}

type RefreshDomainsAccessResponse struct {
	Status APIResponseStatus                `json:"status"`
	Data   RefreshDomainsAccessResponseData `json:"data,omitempty"`
}

type RefreshDomainsAccessResponseData struct {
	DomainIds []float64 `json:"domain_ids,omitempty"`
}

type ExportDomainsResponse struct {
	Status APIResponseStatus         `json:"status"`
	Data   ExportDomainsResponseData `json:"data,omitempty"`
}

type ExportDomainsResponseData struct {
	Hash    string `json:"hash,omitempty"`
	Key     string `json:"key,omitempty"`
	RealUrl string `json:"real_url,omitempty"`
}

type AddOriginsResponse struct {
	Status APIResponseStatus      `json:"status"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

type UpdateOriginsResponse struct {
	Status APIResponseStatus         `json:"status"`
	Data   UpdateOriginsResponseData `json:"data,omitempty"`
}

type UpdateOriginsResponseData struct {
	Id float64 `json:"id,omitempty"`
}

type DeleteOriginsResponse struct {
	Status APIResponseStatus         `json:"status"`
	Data   DeleteOriginsResponseData `json:"data,omitempty"`
}

type DeleteOriginsResponseData struct {
	Ids []float64 `json:"ids,omitempty"`
}

type ListOriginsResponse struct {
	Status APIResponseStatus       `json:"status"`
	Data   ListOriginsResponseData `json:"data,omitempty"`
}

type ListOriginsResponseData struct {
	Total float64                       `json:"total,omitempty"`
	List  []ListOriginsResponseDataList `json:"list,omitempty"`
}

type ListOriginsResponseDataList struct {
	Id             float64                              `json:"id,omitempty"`
	DomainId       float64                              `json:"domain_id,omitempty"`
	Protocol       float64                              `json:"protocol,omitempty"`
	ListenPort     float64                              `json:"listen_port,omitempty"`
	OriginProtocol float64                              `json:"origin_protocol,omitempty"`
	LoadBalance    float64                              `json:"load_balance,omitempty"`
	OriginType     float64                              `json:"origin_type,omitempty"`
	Records        []ListOriginsResponseDataListRecords `json:"records,omitempty"`
}

type ListOriginsResponseDataListRecords struct {
	View     string  `json:"view,omitempty"`
	Value    string  `json:"value,omitempty"`
	Port     float64 `json:"port,omitempty"`
	Priority float64 `json:"priority,omitempty"`
	Host     string  `json:"host,omitempty"`
}

type SwitchDomainNodesResponse struct {
	Status APIResponseStatus             `json:"status"`
	Data   SwitchDomainNodesResponseData `json:"data,omitempty"`
}

type SwitchDomainNodesResponseData struct {
	DomainId float64 `json:"domain_id,omitempty"`
}

type SwitchDomainAccessModeResponse struct {
	Status APIResponseStatus                  `json:"status"`
	Data   SwitchDomainAccessModeResponseData `json:"data,omitempty"`
}

type SwitchDomainAccessModeResponseData struct {
	DomainId float64 `json:"domain_id,omitempty"`
}

type UpdateDomainBaseSettingsResponse struct {
	Status APIResponseStatus                    `json:"status"`
	Data   UpdateDomainBaseSettingsResponseData `json:"data,omitempty"`
}

type UpdateDomainBaseSettingsResponseData struct {
	DomainId float64 `json:"domain_id,omitempty"`
}

type GetDomainBaseSettingsResponse struct {
	Status APIResponseStatus                 `json:"status"`
	Data   GetDomainBaseSettingsResponseData `json:"data,omitempty"`
}

type GetDomainBaseSettingsResponseData struct {
	DomainId       float64                                         `json:"domain_id,omitempty"`
	ProxyHost      GetDomainBaseSettingsResponseDataProxyHost      `json:"proxy_host,omitempty"`
	ProxySni       GetDomainBaseSettingsResponseDataProxySni       `json:"proxy_sni,omitempty"`
	DomainRedirect GetDomainBaseSettingsResponseDataDomainRedirect `json:"domain_redirect,omitempty"`
}

type GetDomainBaseSettingsResponseDataProxyHost struct {
	ProxyHost     string `json:"proxy_host,omitempty"`
	ProxyHostType string `json:"proxy_host_type,omitempty"`
}

type GetDomainBaseSettingsResponseDataProxySni struct {
	ProxySni string `json:"proxy_sni,omitempty"`
	Status   string `json:"status,omitempty"`
}

type GetDomainBaseSettingsResponseDataDomainRedirect struct {
	Status   string `json:"status,omitempty"`
	JumpTo   string `json:"jump_to,omitempty"`
	JumpType string `json:"jump_type,omitempty"`
}

type ListBriefDomainsResponse struct {
	Status APIResponseStatus            `json:"status"`
	Data   ListBriefDomainsResponseData `json:"data,omitempty"`
}

type ListBriefDomainsResponseData struct {
	Total float64                            `json:"total,omitempty"`
	List  []ListBriefDomainsResponseDataList `json:"list,omitempty"`
}

type ListBriefDomainsResponseDataList struct {
	Id     float64 `json:"id,omitempty"`
	Domain string  `json:"domain,omitempty"`
}

type GetDomainTemplatesResponse struct {
	Status APIResponseStatus              `json:"status"`
	Data   GetDomainTemplatesResponseData `json:"data,omitempty"`
}

type GetDomainTemplatesResponseData struct {
	DomainId        float64                                         `json:"domain_id,omitempty"`
	BindedTemplates []GetDomainTemplatesResponseDataBindedTemplates `json:"binded_templates,omitempty"`
}

type GetDomainTemplatesResponseDataBindedTemplates struct {
	BusinessId   float64 `json:"business_id,omitempty"`
	BusinessType string  `json:"business_type,omitempty"`
	AppType      string  `json:"app_type,omitempty"`
	Name         string  `json:"name,omitempty"`
}

type AccessInfoDownloadResponse struct {
	Status APIResponseStatus              `json:"status"`
	Data   AccessInfoDownloadResponseData `json:"data,omitempty"`
}

type AccessInfoDownloadResponseData struct {
	Hash    string `json:"hash,omitempty"`
	Key     string `json:"key,omitempty"`
	RealUrl string `json:"real_url,omitempty"`
}

type OriginGroupGetOriginGroupListResponse struct {
	Status APIResponseStatus                         `json:"status"`
	Data   OriginGroupGetOriginGroupListResponseData `json:"data,omitempty"`
}

type OriginGroupGetOriginGroupListResponseData struct {
	Total float64                                         `json:"total,omitempty"`
	List  []OriginGroupGetOriginGroupListResponseDataList `json:"list,omitempty"`
}

type OriginGroupGetOriginGroupListResponseDataList struct {
	Id        float64                                                `json:"id,omitempty"`
	Name      string                                                 `json:"name,omitempty"`
	Remark    string                                                 `json:"remark,omitempty"`
	MemberId  float64                                                `json:"member_id,omitempty"`
	Username  string                                                 `json:"username,omitempty"`
	Origins   []OriginGroupGetOriginGroupListResponseDataListOrigins `json:"origins,omitempty"`
	CreatedAt string                                                 `json:"created_at,omitempty"`
	UpdatedAt string                                                 `json:"updated_at,omitempty"`
}

type OriginGroupGetOriginGroupListResponseDataListOrigins struct {
	OriginType     float64                                                             `json:"origin_type,omitempty"`
	Records        []OriginGroupGetOriginGroupListResponseDataListOriginsRecords       `json:"records,omitempty"`
	ProtocolPorts  []OriginGroupGetOriginGroupListResponseDataListOriginsProtocolPorts `json:"protocol_ports,omitempty"`
	OriginProtocol interface{}                                                         `json:"origin_protocol,omitempty"`
	LoadBalance    interface{}                                                         `json:"load_balance,omitempty"`
}

type OriginGroupGetOriginGroupListResponseDataListOriginsRecords struct {
	Value    string      `json:"value,omitempty"`
	Port     float64     `json:"port,omitempty"`
	Priority float64     `json:"priority,omitempty"`
	View     interface{} `json:"view,omitempty"`
	Host     string      `json:"host,omitempty"`
}

type OriginGroupGetOriginGroupListResponseDataListOriginsProtocolPorts struct {
	Protocol    interface{} `json:"protocol,omitempty"`
	ListenPorts []float64   `json:"listen_ports,omitempty"`
}

type OriginGroupGetOriginGroupInfoResponse struct {
	Status APIResponseStatus                         `json:"status"`
	Data   OriginGroupGetOriginGroupInfoResponseData `json:"data,omitempty"`
}

type OriginGroupGetOriginGroupInfoResponseData struct {
	OriginGroup OriginGroupGetOriginGroupInfoResponseDataOriginGroup `json:"origin_group,omitempty"`
}

type OriginGroupGetOriginGroupInfoResponseDataOriginGroup struct {
	Id        float64                                                       `json:"id,omitempty"`
	Name      string                                                        `json:"name,omitempty"`
	Remark    string                                                        `json:"remark,omitempty"`
	MemberId  float64                                                       `json:"member_id,omitempty"`
	Username  string                                                        `json:"username,omitempty"`
	Origins   []OriginGroupGetOriginGroupInfoResponseDataOriginGroupOrigins `json:"origins,omitempty"`
	CreatedAt string                                                        `json:"created_at,omitempty"`
	UpdatedAt string                                                        `json:"updated_at,omitempty"`
}

type OriginGroupGetOriginGroupInfoResponseDataOriginGroupOrigins struct {
	OriginType     float64                                                                    `json:"origin_type,omitempty"`
	Records        []OriginGroupGetOriginGroupInfoResponseDataOriginGroupOriginsRecords       `json:"records,omitempty"`
	ProtocolPorts  []OriginGroupGetOriginGroupInfoResponseDataOriginGroupOriginsProtocolPorts `json:"protocol_ports,omitempty"`
	OriginProtocol interface{}                                                                `json:"origin_protocol,omitempty"`
	LoadBalance    interface{}                                                                `json:"load_balance,omitempty"`
}

type OriginGroupGetOriginGroupInfoResponseDataOriginGroupOriginsRecords struct {
	Value    string      `json:"value,omitempty"`
	Port     float64     `json:"port,omitempty"`
	Priority float64     `json:"priority,omitempty"`
	View     interface{} `json:"view,omitempty"`
	Host     string      `json:"host,omitempty"`
}

type OriginGroupGetOriginGroupInfoResponseDataOriginGroupOriginsProtocolPorts struct {
	Protocol    interface{} `json:"protocol,omitempty"`
	ListenPorts []float64   `json:"listen_ports,omitempty"`
}

type OriginGroupAddOriginGroupResponse struct {
	Status APIResponseStatus                     `json:"status"`
	Data   OriginGroupAddOriginGroupResponseData `json:"data,omitempty"`
}

type OriginGroupAddOriginGroupResponseData struct {
	Id float64 `json:"id,omitempty"`
}

type OriginGroupUpdateOriginGroupResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type OriginGroupDelOriginGroupResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type OriginGroupBindOriginGroupToDomainsResponse struct {
	Status APIResponseStatus                               `json:"status"`
	Data   OriginGroupBindOriginGroupToDomainsResponseData `json:"data,omitempty"`
}

type OriginGroupBindOriginGroupToDomainsResponseData struct {
	JobId string `json:"job_id,omitempty"`
}

type OriginGroupGetAllOriginGroupsResponse struct {
	Status APIResponseStatus                         `json:"status"`
	Data   OriginGroupGetAllOriginGroupsResponseData `json:"data,omitempty"`
}

type OriginGroupGetAllOriginGroupsResponseData struct {
	Total float64                                         `json:"total,omitempty"`
	List  []OriginGroupGetAllOriginGroupsResponseDataList `json:"list,omitempty"`
}

type OriginGroupGetAllOriginGroupsResponseDataList struct {
	Id      float64                                                `json:"id,omitempty"`
	Name    string                                                 `json:"name,omitempty"`
	Origins []OriginGroupGetAllOriginGroupsResponseDataListOrigins `json:"origins,omitempty"`
}

type OriginGroupGetAllOriginGroupsResponseDataListOrigins struct {
	OriginType     float64                                                             `json:"origin_type,omitempty"`
	Records        []OriginGroupGetAllOriginGroupsResponseDataListOriginsRecords       `json:"records,omitempty"`
	ProtocolPorts  []OriginGroupGetAllOriginGroupsResponseDataListOriginsProtocolPorts `json:"protocol_ports,omitempty"`
	OriginProtocol interface{}                                                         `json:"origin_protocol,omitempty"`
	LoadBalance    interface{}                                                         `json:"load_balance,omitempty"`
}

type OriginGroupGetAllOriginGroupsResponseDataListOriginsRecords struct {
	Value    string      `json:"value,omitempty"`
	Port     float64     `json:"port,omitempty"`
	Priority float64     `json:"priority,omitempty"`
	View     interface{} `json:"view,omitempty"`
	Host     string      `json:"host,omitempty"`
}

type OriginGroupGetAllOriginGroupsResponseDataListOriginsProtocolPorts struct {
	Protocol    interface{} `json:"protocol,omitempty"`
	ListenPorts []float64   `json:"listen_ports,omitempty"`
}

type OriginGroupCopyOriginGroupResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type FireWallReportGetBlockListResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type FireWallReportGetBlockDetailsResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type FireWallReportGetPackageBlockListResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type FireWallReportGetPackageBlockDetailsResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type CcQpsMaxResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type CcAttackTimesResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type CcTimesLineResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type CcReportStatsResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type CdnDomainUaispDistributeResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type CdnDomainCountryDistributeResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type CdnDomainProvinceDistributeResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type CdnDomainStatusDistributeResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type CdnDomainNodeFlowBandwidthResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type CdnDomainNodeFlowBandwidthCn2Response struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type CdnDomainNodeFlowBandwidthNodeResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type DomainTimesResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type DomainQpsResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type CdnDomainFlowLineResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type CdnDomainBandwidthLineResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type CdnDomainBandwidth95Response struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type CdnDomainPvtimesResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type CdnDomainFlowTopResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type CdnDomainBandwidthTopResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type CdnDomainTimesTopResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type CdnDomainTimesTopEsResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type CdnDomainUrlTopResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type CdnDomainRefererTopResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type CdnDomainStatusTopDownloadResponse struct {
	Status APIResponseStatus                      `json:"status"`
	Data   CdnDomainStatusTopDownloadResponseData `json:"data,omitempty"`
}

type CdnDomainStatusTopDownloadResponseData struct {
	Hash    string `json:"hash,omitempty"`
	Key     string `json:"key,omitempty"`
	RealUrl string `json:"real_url,omitempty"`
}

type CdnDomainBandwidthDownloadResponse struct {
	Status APIResponseStatus                      `json:"status"`
	Data   CdnDomainBandwidthDownloadResponseData `json:"data,omitempty"`
}

type CdnDomainBandwidthDownloadResponseData struct {
	Hash    string `json:"hash,omitempty"`
	Key     string `json:"key,omitempty"`
	RealUrl string `json:"real_url,omitempty"`
}

type CdnDomainFlowDownloadResponse struct {
	Status APIResponseStatus                 `json:"status"`
	Data   CdnDomainFlowDownloadResponseData `json:"data,omitempty"`
}

type CdnDomainFlowDownloadResponseData struct {
	Hash    string `json:"hash,omitempty"`
	Key     string `json:"key,omitempty"`
	RealUrl string `json:"real_url,omitempty"`
}

type TcpBandwidthResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type TcpCcFlawResponse struct {
	Status APIResponseStatus       `json:"status"`
	Data   []TcpCcFlawResponseData `json:"data,omitempty"`
}

type TcpCcFlawResponseData struct {
	Max   TcpCcFlawResponseDataMax   `json:"max,omitempty"`
	Trend TcpCcFlawResponseDataTrend `json:"trend,omitempty"`
}

type TcpCcFlawResponseDataMax struct {
	Data map[string]interface{} `json:"data,omitempty"`
	Time string                 `json:"time,omitempty"`
}

type TcpCcFlawResponseDataTrend struct {
	XData []string                          `json:"x_data,omitempty"`
	YData []TcpCcFlawResponseDataTrendYData `json:"y_data,omitempty"`
}

type TcpCcFlawResponseDataTrendYData struct {
	Data []float64 `json:"data,omitempty"`
	Unit string    `json:"unit,omitempty"`
}

type WafAttackTimesResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type WafReportStatsResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type WafWebshellEventListResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type WafWebshellEventDetailResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type WafAttackEventListResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type WafAttackEventDetailResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type WafScanEventListResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type WafScanEventDetailResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type WafTypeLineResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type LogDownloadTaskTaskListResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type LogDownloadTaskAddTaskResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type LogDownloadTaskCancelTaskResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type LogDownloadTaskBatchCancelTaskResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type LogDownloadTaskDeleteTaskResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type LogDownloadTaskBatchDeleteTaskResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type LogDownloadTaskRegenerateTaskResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type LogDownloadFieldConfDownloadFieldsResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type LogDownloadTemplateTemplateListResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type LogDownloadTemplateGetTemplateDomainListResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type LogDownloadTemplateAddTemplateResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type LogDownloadTemplateSaveTemplateResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type LogDownloadTemplateDelTemplateResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type LogDownloadTemplateBatchDelTemplateResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type LogDownloadTemplateChangeStatusResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type LogDownloadTemplateBatchChangeStatusResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type LogDownloadTemplateAllTemplateResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type LogDownloadTemplateAllTemplateGroupResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type TjkdPlusPackageGetMemberPackageListResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type TjkdPlusPackageGetAllPackageResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type TjkdPlusPackageGetPackageInfoResponse struct {
	Status APIResponseStatus                         `json:"status"`
	Data   TjkdPlusPackageGetPackageInfoResponseData `json:"data,omitempty"`
}

type TjkdPlusPackageGetPackageInfoResponseData struct {
	Id              string `json:"id,omitempty"`
	PackageName     string `json:"package_name,omitempty"`
	PackageTypeName string `json:"package_type_name,omitempty"`
	PackageRulesNum string `json:"package_rules_num,omitempty"`
	ExpireTime      string `json:"expire_time,omitempty"`
	Remark          string `json:"remark,omitempty"`
}

type TjkdPlusPackageGetPackageIpListResponse struct {
	Status APIResponseStatus                           `json:"status"`
	Data   TjkdPlusPackageGetPackageIpListResponseData `json:"data,omitempty"`
}

type TjkdPlusPackageGetPackageIpListResponseData struct {
	List []TjkdPlusPackageGetPackageIpListResponseDataList `json:"list,omitempty"`
}

type TjkdPlusPackageGetPackageIpListResponseDataList struct {
	Id       string `json:"id,omitempty"`
	IpString string `json:"ip_string,omitempty"`
}

type TjkdPlusPackageGetPackageOverviewResponse struct {
	Status APIResponseStatus                             `json:"status"`
	Data   TjkdPlusPackageGetPackageOverviewResponseData `json:"data,omitempty"`
}

type TjkdPlusPackageGetPackageOverviewResponseData struct {
	ExpireTotal float64                                             `json:"expire_total,omitempty"`
	List        []TjkdPlusPackageGetPackageOverviewResponseDataList `json:"list,omitempty"`
}

type TjkdPlusPackageGetPackageOverviewResponseDataList struct {
	Id               string `json:"id,omitempty"`
	PackageTypeId    string `json:"package_type_id,omitempty"`
	PackageTypeSign  string `json:"package_type_sign,omitempty"`
	PackageName      string `json:"package_name,omitempty"`
	ExpireTime       string `json:"expire_time,omitempty"`
	PackageRulesNum  string `json:"package_rules_num,omitempty"`
	ElasticityStatus string `json:"elasticity_status,omitempty"`
	PackageType      string `json:"package_type,omitempty"`
	DomainRule       string `json:"domain_rule,omitempty"`
	TcpRule          string `json:"tcp_rule,omitempty"`
}

type TjkdPlusPackageGetPackagePortListResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type TjkdPlusPackageSavePackageResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type TjkdPlusPackageSavePackageHealthyConfResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type TjkdPlusForwardRuleSavePlusForwardRuleResponse struct {
	Status APIResponseStatus                                  `json:"status"`
	Data   TjkdPlusForwardRuleSavePlusForwardRuleResponseData `json:"data,omitempty"`
}

type TjkdPlusForwardRuleSavePlusForwardRuleResponseData struct {
	Info string `json:"info,omitempty"`
}

type TjkdPlusForwardRuleBatchAddPlusForwardRuleResponse struct {
	Status APIResponseStatus                                      `json:"status"`
	Data   TjkdPlusForwardRuleBatchAddPlusForwardRuleResponseData `json:"data,omitempty"`
}

type TjkdPlusForwardRuleBatchAddPlusForwardRuleResponseData struct {
	Info string `json:"info,omitempty"`
}

type TjkdPlusForwardRuleBatchSavePlusForwardRuleResponse struct {
	Status APIResponseStatus                                       `json:"status"`
	Data   TjkdPlusForwardRuleBatchSavePlusForwardRuleResponseData `json:"data,omitempty"`
}

type TjkdPlusForwardRuleBatchSavePlusForwardRuleResponseData struct {
	Info string `json:"info,omitempty"`
}

type TjkdPlusForwardRuleDelPlusForwardRuleResponse struct {
	Status APIResponseStatus                                 `json:"status"`
	Data   TjkdPlusForwardRuleDelPlusForwardRuleResponseData `json:"data,omitempty"`
}

type TjkdPlusForwardRuleDelPlusForwardRuleResponseData struct {
	Info string `json:"info,omitempty"`
}

type TjkdPlusForwardRuleGetPlusForwardRuleListResponse struct {
	Status APIResponseStatus                                     `json:"status"`
	Data   TjkdPlusForwardRuleGetPlusForwardRuleListResponseData `json:"data,omitempty"`
}

type TjkdPlusForwardRuleGetPlusForwardRuleListResponseData struct {
	Total float64                                                     `json:"total,omitempty"`
	List  []TjkdPlusForwardRuleGetPlusForwardRuleListResponseDataList `json:"list,omitempty"`
}

type TjkdPlusForwardRuleGetPlusForwardRuleListResponseDataList struct {
	Ids          []string                                                           `json:"ids,omitempty"`
	PackageName  string                                                             `json:"package_name,omitempty"`
	ExpireTime   string                                                             `json:"expire_time,omitempty"`
	PackageId    string                                                             `json:"package_id,omitempty"`
	Protocol     string                                                             `json:"protocol,omitempty"`
	ProtocolName string                                                             `json:"protocol_name,omitempty"`
	ProtocolPort string                                                             `json:"protocol_port,omitempty"`
	Loading      string                                                             `json:"loading,omitempty"`
	LoadingName  string                                                             `json:"loading_name,omitempty"`
	SourceType   string                                                             `json:"source_type,omitempty"`
	Status       string                                                             `json:"status,omitempty"`
	Remark       string                                                             `json:"remark,omitempty"`
	Sourcelength string                                                             `json:"sourceLength,omitempty"`
	InstanceId   string                                                             `json:"instance_id,omitempty"`
	Records      []TjkdPlusForwardRuleGetPlusForwardRuleListResponseDataListRecords `json:"records,omitempty"`
}

type TjkdPlusForwardRuleGetPlusForwardRuleListResponseDataListRecords struct {
	Id    string `json:"id,omitempty"`
	Value string `json:"value,omitempty"`
	Port  string `json:"port,omitempty"`
	View  string `json:"view,omitempty"`
}

type TjkdPlusForwardRuleGetBatchPlusForwardRuleInfoResponse struct {
	Status APIResponseStatus                                          `json:"status"`
	Data   TjkdPlusForwardRuleGetBatchPlusForwardRuleInfoResponseData `json:"data,omitempty"`
}

type TjkdPlusForwardRuleGetBatchPlusForwardRuleInfoResponseData struct {
	Protocol     string                                                             `json:"protocol,omitempty"`
	Remark       string                                                             `json:"remark,omitempty"`
	ProtocolPort []string                                                           `json:"protocol_port,omitempty"`
	Loading      string                                                             `json:"loading,omitempty"`
	SourceType   string                                                             `json:"source_type,omitempty"`
	Length       string                                                             `json:"length,omitempty"`
	Source       []TjkdPlusForwardRuleGetBatchPlusForwardRuleInfoResponseDataSource `json:"source,omitempty"`
}

type TjkdPlusForwardRuleGetBatchPlusForwardRuleInfoResponseDataSource struct {
	Id    string `json:"id,omitempty"`
	Value string `json:"value,omitempty"`
	Port  string `json:"port,omitempty"`
	View  string `json:"view,omitempty"`
}

type TjkdPlusPackageGetPackageDomainListResponse struct {
	Status APIResponseStatus                               `json:"status"`
	Data   TjkdPlusPackageGetPackageDomainListResponseData `json:"data,omitempty"`
}

type TjkdPlusPackageGetPackageDomainListResponseData struct {
	List []TjkdPlusPackageGetPackageDomainListResponseDataList `json:"list,omitempty"`
}

type TjkdPlusPackageGetPackageDomainListResponseDataList struct {
	Id     string `json:"id,omitempty"`
	Domain string `json:"domain,omitempty"`
}

type TjkdPlusDomainGetTjkdPlusDomainListResponse struct {
	Status APIResponseStatus                               `json:"status"`
	Data   TjkdPlusDomainGetTjkdPlusDomainListResponseData `json:"data,omitempty"`
}

type TjkdPlusDomainGetTjkdPlusDomainListResponseData struct {
	Total string                                                `json:"total,omitempty"`
	List  []TjkdPlusDomainGetTjkdPlusDomainListResponseDataList `json:"list,omitempty"`
}

type TjkdPlusDomainGetTjkdPlusDomainListResponseDataList struct {
	Id              string                                                       `json:"id,omitempty"`
	MemberId        string                                                       `json:"member_id,omitempty"`
	Domain          string                                                       `json:"domain,omitempty"`
	DomainId        string                                                       `json:"domain_id,omitempty"`
	PackageId       string                                                       `json:"package_id,omitempty"`
	PackageName     string                                                       `json:"package_name,omitempty"`
	ExpireTime      string                                                       `json:"expire_time,omitempty"`
	ProtectStatus   string                                                       `json:"protect_status,omitempty"`
	Status          string                                                       `json:"status,omitempty"`
	CreatedAt       string                                                       `json:"created_at,omitempty"`
	UpdatedAt       string                                                       `json:"updated_at,omitempty"`
	ProtectedStatus string                                                       `json:"protected_status,omitempty"`
	Records         []TjkdPlusDomainGetTjkdPlusDomainListResponseDataListRecords `json:"records,omitempty"`
}

type TjkdPlusDomainGetTjkdPlusDomainListResponseDataListRecords struct {
	Id           string `json:"id,omitempty"`
	ListenPortId string `json:"listen_port_id,omitempty"`
	TypeValue    string `json:"type,omitempty"`
	View         string `json:"view,omitempty"`
	Value        string `json:"value,omitempty"`
	Priority     string `json:"priority,omitempty"`
	Port         string `json:"port,omitempty"`
	ListenPort   string `json:"listen_port,omitempty"`
}

type TjkdPlusDomainAddTjkdPlusDomainResponse struct {
	Status APIResponseStatus                           `json:"status"`
	Data   TjkdPlusDomainAddTjkdPlusDomainResponseData `json:"data,omitempty"`
}

type TjkdPlusDomainAddTjkdPlusDomainResponseData struct {
	Info string `json:"info,omitempty"`
}

type TjkdPlusDomainDelTjkdPlusDomainResponse struct {
	Status APIResponseStatus                           `json:"status"`
	Data   TjkdPlusDomainDelTjkdPlusDomainResponseData `json:"data,omitempty"`
}

type TjkdPlusDomainDelTjkdPlusDomainResponseData struct {
	Info string `json:"info,omitempty"`
}

type NetworkSpeedGetCacheRuleListResponse struct {
	Status APIResponseStatus                        `json:"status"`
	Data   NetworkSpeedGetCacheRuleListResponseData `json:"data,omitempty"`
}

type NetworkSpeedGetCacheRuleListResponseData struct {
	Page     float64                                        `json:"page,omitempty"`
	PageSize float64                                        `json:"page_size,omitempty"`
	Total    float64                                        `json:"total,omitempty"`
	List     []NetworkSpeedGetCacheRuleListResponseDataList `json:"list,omitempty"`
}

type NetworkSpeedGetCacheRuleListResponseDataList struct {
	Id        float64                                          `json:"id,omitempty"`
	Name      string                                           `json:"name,omitempty"`
	Remark    string                                           `json:"remark,omitempty"`
	Status    interface{}                                      `json:"status,omitempty"`
	Weight    float64                                          `json:"weight,omitempty"`
	Mode      float64                                          `json:"mode,omitempty"`
	Expr      string                                           `json:"expr,omitempty"`
	TypeValue string                                           `json:"type,omitempty"`
	Conf      NetworkSpeedGetCacheRuleListResponseDataListConf `json:"conf,omitempty"`
}

type NetworkSpeedGetCacheRuleListResponseDataListConf struct {
	Nocache          bool                                                             `json:"nocache,omitempty"`
	CacheRule        NetworkSpeedGetCacheRuleListResponseDataListConfCacheRule        `json:"cache_rule,omitempty"`
	BrowserCacheRule NetworkSpeedGetCacheRuleListResponseDataListConfBrowserCacheRule `json:"browser_cache_rule,omitempty"`
	CacheErrstatus   NetworkSpeedGetCacheRuleListResponseDataListConfCacheErrstatus   `json:"cache_errstatus,omitempty"`
	CacheUrlRewrite  NetworkSpeedGetCacheRuleListResponseDataListConfCacheUrlRewrite  `json:"cache_url_rewrite,omitempty"`
	CacheShare       NetworkSpeedGetCacheRuleListResponseDataListConfCacheShare       `json:"cache_share,omitempty"`
}

type NetworkSpeedGetCacheRuleListResponseDataListConfCacheRule struct {
	Cachetime float64     `json:"cachetime,omitempty"`
	Action    interface{} `json:"action,omitempty"`
}

type NetworkSpeedGetCacheRuleListResponseDataListConfBrowserCacheRule struct {
	Cachetime       float64 `json:"cachetime,omitempty"`
	IgnoreCacheTime bool    `json:"ignore_cache_time,omitempty"`
	Nocache         bool    `json:"nocache,omitempty"`
}

type NetworkSpeedGetCacheRuleListResponseDataListConfCacheErrstatus struct {
	Cachetime float64   `json:"cachetime,omitempty"`
	ErrStatus []float64 `json:"err_status,omitempty"`
}

type NetworkSpeedGetCacheRuleListResponseDataListConfCacheUrlRewrite struct {
	SortArgs   bool                                                                   `json:"sort_args,omitempty"`
	IgnoreCase bool                                                                   `json:"ignore_case,omitempty"`
	Queries    NetworkSpeedGetCacheRuleListResponseDataListConfCacheUrlRewriteQueries `json:"queries,omitempty"`
	Cookies    NetworkSpeedGetCacheRuleListResponseDataListConfCacheUrlRewriteCookies `json:"cookies,omitempty"`
}

type NetworkSpeedGetCacheRuleListResponseDataListConfCacheUrlRewriteQueries struct {
	ArgsMethod interface{} `json:"args_method,omitempty"`
	Items      []string    `json:"items,omitempty"`
}

type NetworkSpeedGetCacheRuleListResponseDataListConfCacheUrlRewriteCookies struct {
	ArgsMethod interface{} `json:"args_method,omitempty"`
	Items      []string    `json:"items,omitempty"`
}

type NetworkSpeedGetCacheRuleListResponseDataListConfCacheShare struct {
	Scheme interface{} `json:"scheme,omitempty"`
}

type NetworkSpeedCreateCacheRuleResponse struct {
	Status APIResponseStatus                       `json:"status"`
	Data   NetworkSpeedCreateCacheRuleResponseData `json:"data,omitempty"`
}

type NetworkSpeedCreateCacheRuleResponseData struct {
	Id float64 `json:"id,omitempty"`
}

type NetworkSpeedUpdateCacheRuleResponse struct {
	Status APIResponseStatus                       `json:"status"`
	Data   NetworkSpeedUpdateCacheRuleResponseData `json:"data,omitempty"`
}

type NetworkSpeedUpdateCacheRuleResponseData struct {
	Id float64 `json:"id,omitempty"`
}

type NetworkSpeedUpdateCacheRuleConfigResponse struct {
	Status APIResponseStatus                             `json:"status"`
	Data   NetworkSpeedUpdateCacheRuleConfigResponseData `json:"data,omitempty"`
}

type NetworkSpeedUpdateCacheRuleConfigResponseData struct {
	Id float64 `json:"id,omitempty"`
}

type NetworkSpeedUpdateCacheRuleStatusResponse struct {
	Status APIResponseStatus                             `json:"status"`
	Data   NetworkSpeedUpdateCacheRuleStatusResponseData `json:"data,omitempty"`
}

type NetworkSpeedUpdateCacheRuleStatusResponseData struct {
	Ids []float64 `json:"ids,omitempty"`
}

type NetworkSpeedSortCacheRulesResponse struct {
	Status APIResponseStatus                      `json:"status"`
	Data   NetworkSpeedSortCacheRulesResponseData `json:"data,omitempty"`
}

type NetworkSpeedSortCacheRulesResponseData struct {
	Ids []float64 `json:"ids,omitempty"`
}

type NetworkSpeedGetGlobalCacheConfigResponse struct {
	Status APIResponseStatus                            `json:"status"`
	Data   NetworkSpeedGetGlobalCacheConfigResponseData `json:"data,omitempty"`
}

type NetworkSpeedGetGlobalCacheConfigResponseData struct {
	Id   float64                                          `json:"id,omitempty"`
	Name string                                           `json:"name,omitempty"`
	Conf NetworkSpeedGetGlobalCacheConfigResponseDataConf `json:"conf,omitempty"`
}

type NetworkSpeedGetGlobalCacheConfigResponseDataConf struct {
	Nocache          bool                                                             `json:"nocache,omitempty"`
	CacheRule        NetworkSpeedGetGlobalCacheConfigResponseDataConfCacheRule        `json:"cache_rule,omitempty"`
	BrowserCacheRule NetworkSpeedGetGlobalCacheConfigResponseDataConfBrowserCacheRule `json:"browser_cache_rule,omitempty"`
	CacheUrlRewrite  NetworkSpeedGetGlobalCacheConfigResponseDataConfCacheUrlRewrite  `json:"cache_url_rewrite,omitempty"`
}

type NetworkSpeedGetGlobalCacheConfigResponseDataConfCacheRule struct {
	Cachetime float64     `json:"cachetime,omitempty"`
	Action    interface{} `json:"action,omitempty"`
}

type NetworkSpeedGetGlobalCacheConfigResponseDataConfBrowserCacheRule struct {
	Cachetime       float64 `json:"cachetime,omitempty"`
	IgnoreCacheTime bool    `json:"ignore_cache_time,omitempty"`
	Nocache         bool    `json:"nocache,omitempty"`
}

type NetworkSpeedGetGlobalCacheConfigResponseDataConfCacheUrlRewrite struct {
	SortArgs   bool                                                                   `json:"sort_args,omitempty"`
	IgnoreCase bool                                                                   `json:"ignore_case,omitempty"`
	Queries    NetworkSpeedGetGlobalCacheConfigResponseDataConfCacheUrlRewriteQueries `json:"queries,omitempty"`
	Cookies    NetworkSpeedGetGlobalCacheConfigResponseDataConfCacheUrlRewriteCookies `json:"cookies,omitempty"`
}

type NetworkSpeedGetGlobalCacheConfigResponseDataConfCacheUrlRewriteQueries struct {
	ArgsMethod interface{} `json:"args_method,omitempty"`
	Items      []string    `json:"items,omitempty"`
}

type NetworkSpeedGetGlobalCacheConfigResponseDataConfCacheUrlRewriteCookies struct {
	ArgsMethod interface{} `json:"args_method,omitempty"`
	Items      []string    `json:"items,omitempty"`
}

type NetworkSpeedDeleteCacheRuleResponse struct {
	Status APIResponseStatus                       `json:"status"`
	Data   NetworkSpeedDeleteCacheRuleResponseData `json:"data,omitempty"`
}

type NetworkSpeedDeleteCacheRuleResponseData struct {
	Ids []float64 `json:"ids,omitempty"`
}

type NetworkSpeedGetTemplateConfigResponse struct {
	Status APIResponseStatus                         `json:"status"`
	Data   NetworkSpeedGetTemplateConfigResponseData `json:"data,omitempty"`
}

type NetworkSpeedGetTemplateConfigResponseData struct {
	BusinessType         string                                                        `json:"business_type,omitempty"`
	BusinessId           float64                                                       `json:"business_id,omitempty"`
	DomainProxyConf      NetworkSpeedGetTemplateConfigResponseDataDomainProxyConf      `json:"domain_proxy_conf,omitempty"`
	UpstreamRedirect     NetworkSpeedGetTemplateConfigResponseDataUpstreamRedirect     `json:"upstream_redirect,omitempty"`
	CustomizedReqHeaders NetworkSpeedGetTemplateConfigResponseDataCustomizedReqHeaders `json:"customized_req_headers,omitempty"`
	RespHeaders          NetworkSpeedGetTemplateConfigResponseDataRespHeaders          `json:"resp_headers,omitempty"`
	UpstreamUriChange    NetworkSpeedGetTemplateConfigResponseDataUpstreamUriChange    `json:"upstream_uri_change,omitempty"`
	SourceSiteProtect    NetworkSpeedGetTemplateConfigResponseDataSourceSiteProtect    `json:"source_site_protect,omitempty"`
	Slice                NetworkSpeedGetTemplateConfigResponseDataSlice                `json:"slice,omitempty"`
	Https                NetworkSpeedGetTemplateConfigResponseDataHttps                `json:"https,omitempty"`
	PageGzip             NetworkSpeedGetTemplateConfigResponseDataPageGzip             `json:"page_gzip,omitempty"`
	Webp                 NetworkSpeedGetTemplateConfigResponseDataWebp                 `json:"webp,omitempty"`
	UploadFile           NetworkSpeedGetTemplateConfigResponseDataUploadFile           `json:"upload_file,omitempty"`
	Websocket            NetworkSpeedGetTemplateConfigResponseDataWebsocket            `json:"websocket,omitempty"`
	MobileJump           NetworkSpeedGetTemplateConfigResponseDataMobileJump           `json:"mobile_jump,omitempty"`
	CustomPage           NetworkSpeedGetTemplateConfigResponseDataCustomPage           `json:"custom_page,omitempty"`
}

type NetworkSpeedGetTemplateConfigResponseDataDomainProxyConf struct {
	ProxyConnectTimeout float64 `json:"proxy_connect_timeout,omitempty"`
	FailsTimeout        float64 `json:"fails_timeout,omitempty"`
	KeepNewSrcTime      float64 `json:"keep_new_src_time,omitempty"`
	MaxFails            float64 `json:"max_fails,omitempty"`
	ProxyKeepalive      float64 `json:"proxy_keepalive,omitempty"`
}

type NetworkSpeedGetTemplateConfigResponseDataUpstreamRedirect struct {
	Status string `json:"status,omitempty"`
}

type NetworkSpeedGetTemplateConfigResponseDataCustomizedReqHeaders struct {
	Status string `json:"status,omitempty"`
}

type NetworkSpeedGetTemplateConfigResponseDataRespHeaders struct {
	Status string `json:"status,omitempty"`
}

type NetworkSpeedGetTemplateConfigResponseDataUpstreamUriChange struct {
	Status string `json:"status,omitempty"`
}

type NetworkSpeedGetTemplateConfigResponseDataSourceSiteProtect struct {
	Status string  `json:"status,omitempty"`
	Num    float64 `json:"num,omitempty"`
	Second float64 `json:"second,omitempty"`
}

type NetworkSpeedGetTemplateConfigResponseDataSlice struct {
	Status string `json:"status,omitempty"`
}

type NetworkSpeedGetTemplateConfigResponseDataHttps struct {
	Status                 string      `json:"status,omitempty"`
	Http2https             interface{} `json:"http2https,omitempty"`
	Http2httpsPort         float64     `json:"http2https_port,omitempty"`
	Http2                  string      `json:"http2,omitempty"`
	Hsts                   string      `json:"hsts,omitempty"`
	OcspStapling           string      `json:"ocsp_stapling,omitempty"`
	MinVersion             interface{} `json:"min_version,omitempty"`
	CiphersPreset          interface{} `json:"ciphers_preset,omitempty"`
	CustomEncryptAlgorithm []string    `json:"custom_encrypt_algorithm,omitempty"`
}

type NetworkSpeedGetTemplateConfigResponseDataPageGzip struct {
	Status string `json:"status,omitempty"`
}

type NetworkSpeedGetTemplateConfigResponseDataWebp struct {
	Status string `json:"status,omitempty"`
}

type NetworkSpeedGetTemplateConfigResponseDataUploadFile struct {
	UploadSize     float64 `json:"upload_size,omitempty"`
	UploadSizeUnit string  `json:"upload_size_unit,omitempty"`
}

type NetworkSpeedGetTemplateConfigResponseDataWebsocket struct {
	Status string `json:"status,omitempty"`
}

type NetworkSpeedGetTemplateConfigResponseDataMobileJump struct {
	Status  string `json:"status,omitempty"`
	JumpUrl string `json:"jump_url,omitempty"`
}

type NetworkSpeedGetTemplateConfigResponseDataCustomPage struct {
	Status string `json:"status,omitempty"`
}

type NetworkSpeedUpdateTemplateConfigResponse struct {
	Status APIResponseStatus                            `json:"status"`
	Data   NetworkSpeedUpdateTemplateConfigResponseData `json:"data,omitempty"`
}

type NetworkSpeedUpdateTemplateConfigResponseData struct {
	BusinessType string  `json:"business_type,omitempty"`
	BusinessId   float64 `json:"business_id,omitempty"`
	Updates      float64 `json:"updates,omitempty"`
	Adds         float64 `json:"adds,omitempty"`
}

type NetworkSpeedGetRulesResponse struct {
	Status APIResponseStatus                `json:"status"`
	Data   NetworkSpeedGetRulesResponseData `json:"data,omitempty"`
}

type NetworkSpeedGetRulesResponseData struct {
	Page     float64                                `json:"page,omitempty"`
	PageSize float64                                `json:"page_size,omitempty"`
	Total    float64                                `json:"total,omitempty"`
	List     []NetworkSpeedGetRulesResponseDataList `json:"list,omitempty"`
}

type NetworkSpeedGetRulesResponseDataList struct {
	Id                       float64                                                      `json:"id,omitempty"`
	BusinessType             string                                                       `json:"business_type,omitempty"`
	BusinessId               float64                                                      `json:"business_id,omitempty"`
	ConfigGroup              string                                                       `json:"config_group,omitempty"`
	CustomPage               NetworkSpeedGetRulesResponseDataListCustomPage               `json:"custom_page,omitempty"`
	UpstreamUriChangeRule    NetworkSpeedGetRulesResponseDataListUpstreamUriChangeRule    `json:"upstream_uri_change_rule,omitempty"`
	RespHeadersRule          NetworkSpeedGetRulesResponseDataListRespHeadersRule          `json:"resp_headers_rule,omitempty"`
	CustomizedReqHeadersRule NetworkSpeedGetRulesResponseDataListCustomizedReqHeadersRule `json:"customized_req_headers_rule,omitempty"`
}

type NetworkSpeedGetRulesResponseDataListCustomPage struct {
	StatusCode  float64 `json:"status_code,omitempty"`
	PageType    string  `json:"page_type,omitempty"`
	PageContent string  `json:"page_content,omitempty"`
}

type NetworkSpeedGetRulesResponseDataListUpstreamUriChangeRule struct {
	Typ    string `json:"typ,omitempty"`
	Action string `json:"action,omitempty"`
	Match  string `json:"match,omitempty"`
	Target string `json:"target,omitempty"`
}

type NetworkSpeedGetRulesResponseDataListRespHeadersRule struct {
	TypeValue string `json:"type,omitempty"`
	Content   string `json:"content,omitempty"`
	Remark    string `json:"remark,omitempty"`
}

type NetworkSpeedGetRulesResponseDataListCustomizedReqHeadersRule struct {
	TypeValue string `json:"type,omitempty"`
	Content   string `json:"content,omitempty"`
	Remark    string `json:"remark,omitempty"`
}

type NetworkSpeedCreateRuleResponse struct {
	Status APIResponseStatus                  `json:"status"`
	Data   NetworkSpeedCreateRuleResponseData `json:"data,omitempty"`
}

type NetworkSpeedCreateRuleResponseData struct {
	Id float64 `json:"id,omitempty"`
}

type NetworkSpeedDeleteRuleResponse struct {
	Status APIResponseStatus                  `json:"status"`
	Data   NetworkSpeedDeleteRuleResponseData `json:"data,omitempty"`
}

type NetworkSpeedDeleteRuleResponseData struct {
	Ids []float64 `json:"ids,omitempty"`
}

type NetworkSpeedSortRulesResponse struct {
	Status APIResponseStatus                 `json:"status"`
	Data   NetworkSpeedSortRulesResponseData `json:"data,omitempty"`
}

type NetworkSpeedSortRulesResponseData struct {
	Ids []float64 `json:"ids,omitempty"`
}

type NetworkSpeedUpdateRuleResponse struct {
	Status APIResponseStatus                  `json:"status"`
	Data   NetworkSpeedUpdateRuleResponseData `json:"data,omitempty"`
}

type NetworkSpeedUpdateRuleResponseData struct {
	Id float64 `json:"id,omitempty"`
}

type UpdateRuleTemplateResponse struct {
	Status APIResponseStatus              `json:"status"`
	Data   UpdateRuleTemplateResponseData `json:"data,omitempty"`
}

type UpdateRuleTemplateResponseData struct {
	Id float64 `json:"id,omitempty"`
}

type DeleteRuleTemplateResponse struct {
	Status APIResponseStatus              `json:"status"`
	Data   DeleteRuleTemplateResponseData `json:"data,omitempty"`
}

type DeleteRuleTemplateResponseData struct {
	Id float64 `json:"id,omitempty"`
}

type GetRuleTemplateListResponse struct {
	Status APIResponseStatus               `json:"status"`
	Data   GetRuleTemplateListResponseData `json:"data,omitempty"`
}

type GetRuleTemplateListResponseData struct {
	List  []GetRuleTemplateListResponseDataList `json:"list,omitempty"`
	Total float64                               `json:"total,omitempty"`
}

type GetRuleTemplateListResponseDataList struct {
	Id          float64                                          `json:"id,omitempty"`
	Name        string                                           `json:"name,omitempty"`
	Description string                                           `json:"description,omitempty"`
	AppType     string                                           `json:"app_type,omitempty"`
	BindDomains []GetRuleTemplateListResponseDataListBindDomains `json:"bind_domains,omitempty"`
	CreatedAt   string                                           `json:"created_at,omitempty"`
}

type GetRuleTemplateListResponseDataListBindDomains struct {
	DomainId  float64 `json:"domain_id,omitempty"`
	Domain    string  `json:"domain,omitempty"`
	CreatedAt string  `json:"created_at,omitempty"`
}

type UnbindRuleTemplateResponse struct {
	Status APIResponseStatus              `json:"status"`
	Data   UnbindRuleTemplateResponseData `json:"data,omitempty"`
}

type UnbindRuleTemplateResponseData struct {
	Id float64 `json:"id,omitempty"`
}

type BindRuleTemplateResponse struct {
	Status APIResponseStatus            `json:"status"`
	Data   BindRuleTemplateResponseData `json:"data,omitempty"`
}

type BindRuleTemplateResponseData struct {
	Id float64 `json:"id,omitempty"`
}

type ListRuleTpsDomainsResponse struct {
	Status APIResponseStatus              `json:"status"`
	Data   ListRuleTpsDomainsResponseData `json:"data,omitempty"`
}

type ListRuleTpsDomainsResponseData struct {
	Total float64                              `json:"total,omitempty"`
	List  []ListRuleTpsDomainsResponseDataList `json:"list,omitempty"`
}

type ListRuleTpsDomainsResponseDataList struct {
	Id        float64 `json:"id,omitempty"`
	Domain    string  `json:"domain,omitempty"`
	CreatedAt string  `json:"created_at,omitempty"`
}

type CreateRuleTemplateResponse struct {
	Status APIResponseStatus              `json:"status"`
	Data   CreateRuleTemplateResponseData `json:"data,omitempty"`
}

type CreateRuleTemplateResponseData struct {
	Id float64 `json:"id,omitempty"`
}

type SwitchDomainTemplateResponse struct {
	Status APIResponseStatus                `json:"status"`
	Data   SwitchDomainTemplateResponseData `json:"data,omitempty"`
}

type SwitchDomainTemplateResponseData struct {
	Info string `json:"info,omitempty"`
}

type FirewallPageCfgResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type FirewallPageCfgHwwsResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type FirewallSavePolicyResponse struct {
	Status APIResponseStatus              `json:"status"`
	Data   FirewallSavePolicyResponseData `json:"data,omitempty"`
}

type FirewallSavePolicyResponseData struct {
	Id string `json:"id,omitempty"`
}

type FirewallGetPolicyResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type FirewallGetPolicyByCodeResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type FirewallStatsPolicyResponse struct {
	Status APIResponseStatus               `json:"status"`
	Data   FirewallStatsPolicyResponseData `json:"data,omitempty"`
}

type FirewallStatsPolicyResponseData struct {
	Total  float64                               `json:"total,omitempty"`
	Used   float64                               `json:"used,omitempty"`
	Remain float64                               `json:"remain,omitempty"`
	Status FirewallStatsPolicyResponseDataStatus `json:"status,omitempty"`
}

type FirewallStatsPolicyResponseDataStatus struct {
	Api0  float64 `json:"0,omitempty"`
	Api1  float64 `json:"1,omitempty"`
	Api2  float64 `json:"2,omitempty"`
	Api_1 float64 `json:"-1,omitempty"`
}

type FirewallOpenResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type FirewallStopResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type FirewallDeleteResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type FirewallSortResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type FirewallGetsPolicyByMainidResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type FirewallGetsPolicyByPackageidResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type FirewallSavePolicyGroupResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type FirewallGetsPolicyGroupByDomainidResponse struct {
	Status APIResponseStatus                             `json:"status"`
	Data   FirewallGetsPolicyGroupByDomainidResponseData `json:"data,omitempty"`
}

type FirewallGetsPolicyGroupByDomainidResponseData struct {
	Total   float64                                             `json:"total,omitempty"`
	MapFrom map[string]interface{}                              `json:"map_from,omitempty"`
	List    []FirewallGetsPolicyGroupByDomainidResponseDataList `json:"list,omitempty"`
}

type FirewallGetsPolicyGroupByDomainidResponseDataList struct {
	Id         string `json:"id,omitempty"`
	Code       string `json:"code,omitempty"`
	BusinessId string `json:"business_id,omitempty"`
	MemberId   string `json:"member_id,omitempty"`
	From       string `json:"from,omitempty"`
	Name       string `json:"name,omitempty"`
	Remark     string `json:"remark,omitempty"`
	Sort       string `json:"sort,omitempty"`
	Status     string `json:"status,omitempty"`
	CreateAt   string `json:"create_at,omitempty"`
	UpdateAt   string `json:"update_at,omitempty"`
	PolicyNum  string `json:"policy_num,omitempty"`
}

type FirewallStopGroupResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type FirewallOpenGroupResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type FirewallDeleteGroupResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type FirewallSortGroupResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type FirewallGetsPolicyByGroupIdResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type GetPolicyGroupTplResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type GetDdosProtectionConfigResponse struct {
	Status APIResponseStatus                   `json:"status"`
	Data   GetDdosProtectionConfigResponseData `json:"data,omitempty"`
}

type GetDdosProtectionConfigResponseData struct {
	ApplicationDdosProtection GetDdosProtectionConfigResponseDataApplicationDdosProtection `json:"application_ddos_protection,omitempty"`
	VisitorAuthentication     GetDdosProtectionConfigResponseDataVisitorAuthentication     `json:"visitor_authentication,omitempty"`
}

type GetDdosProtectionConfigResponseDataApplicationDdosProtection struct {
	Status              interface{} `json:"status,omitempty"`
	AiCcStatus          interface{} `json:"ai_cc_status,omitempty"`
	TypeValue           interface{} `json:"type,omitempty"`
	NeedAttackDetection float64     `json:"need_attack_detection,omitempty"`
	AiStatus            interface{} `json:"ai_status,omitempty"`
}

type GetDdosProtectionConfigResponseDataVisitorAuthentication struct {
	Status         interface{} `json:"status,omitempty"`
	AuthToken      string      `json:"auth_token,omitempty"`
	PassStillCheck float64     `json:"pass_still_check,omitempty"`
}

type UpdateDdosProtectionConfigResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type GetWafRuleConfigResponse struct {
	Status APIResponseStatus            `json:"status"`
	Data   GetWafRuleConfigResponseData `json:"data,omitempty"`
}

type GetWafRuleConfigResponseData struct {
	WafRuleConfig          GetWafRuleConfigResponseDataWafRuleConfig    `json:"waf_rule_config,omitempty"`
	WafInterceptPage       GetWafRuleConfigResponseDataWafInterceptPage `json:"waf_intercept_page,omitempty"`
	ReplayAttackProtection map[string]interface{}                       `json:"replay_attack_protection,omitempty"`
	CsrfProtection         map[string]interface{}                       `json:"csrf_protection,omitempty"`
	WebShellProtection     map[string]interface{}                       `json:"web_shell_protection,omitempty"`
}

type GetWafRuleConfigResponseDataWafRuleConfig struct {
	Status   interface{} `json:"status,omitempty"`
	AiStatus interface{} `json:"ai_status,omitempty"`
	WafLevel interface{} `json:"waf_level,omitempty"`
	WafMode  interface{} `json:"waf_mode,omitempty"`
}

type GetWafRuleConfigResponseDataWafInterceptPage struct {
	Status    interface{} `json:"status,omitempty"`
	TypeValue interface{} `json:"type,omitempty"`
	Content   string      `json:"content,omitempty"`
}

type UpdateWafRuleConfigResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type GetMemberGlobalTemplateResponse struct {
	Status APIResponseStatus                   `json:"status"`
	Data   GetMemberGlobalTemplateResponseData `json:"data,omitempty"`
}

type GetMemberGlobalTemplateResponseData struct {
	Template        GetMemberGlobalTemplateResponseDataTemplate `json:"template,omitempty"`
	BindDomainCount float64                                     `json:"bind_domain_count,omitempty"`
}

type GetMemberGlobalTemplateResponseDataTemplate struct {
	Id        float64     `json:"id,omitempty"`
	Name      string      `json:"name,omitempty"`
	TypeValue interface{} `json:"type,omitempty"`
	CreatedAt string      `json:"created_at,omitempty"`
	Remark    string      `json:"remark,omitempty"`
}

type CreateTemplateResponse struct {
	Status APIResponseStatus          `json:"status"`
	Data   CreateTemplateResponseData `json:"data,omitempty"`
}

type CreateTemplateResponseData struct {
	BusinessId  float64                `json:"business_id,omitempty"`
	FailDomains map[string]interface{} `json:"fail_domains,omitempty"`
}

type CreateDomainTemplateResponse struct {
	Status APIResponseStatus                `json:"status"`
	Data   CreateDomainTemplateResponseData `json:"data,omitempty"`
}

type CreateDomainTemplateResponseData struct {
	FailDomains map[string]interface{} `json:"fail_domains,omitempty"`
	BusinessIds []float64              `json:"business_ids,omitempty"`
}

type GetTemplateListResponse struct {
	Status APIResponseStatus           `json:"status"`
	Data   GetTemplateListResponseData `json:"data,omitempty"`
}

type GetTemplateListResponseData struct {
	Templates []map[string]interface{} `json:"templates,omitempty"`
	Total     float64                  `json:"total,omitempty"`
}

type GetTemplateBindDomainListResponse struct {
	Status APIResponseStatus                     `json:"status"`
	Data   GetTemplateBindDomainListResponseData `json:"data,omitempty"`
}

type GetTemplateBindDomainListResponseData struct {
	Domains []map[string]interface{} `json:"domains,omitempty"`
	Total   float64                  `json:"total,omitempty"`
}

type BindTemplateDomainResponse struct {
	Status APIResponseStatus              `json:"status"`
	Data   BindTemplateDomainResponseData `json:"data,omitempty"`
}

type BindTemplateDomainResponseData struct {
	FailDomains map[string]interface{} `json:"fail_domains,omitempty"`
}

type DeleteTemplateResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type BatchConfigTemplateResponse struct {
	Status APIResponseStatus               `json:"status"`
	Data   BatchConfigTemplateResponseData `json:"data,omitempty"`
}

type BatchConfigTemplateResponseData struct {
	FailTemplates map[string]interface{} `json:"fail_templates,omitempty"`
}

type IotaResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   IotaResponseData  `json:"data,omitempty"`
}

type IotaResponseData struct {
	Iota map[string]interface{} `json:"iota,omitempty"`
}

type GetUnboundTemplateDomainListResponse struct {
	Status APIResponseStatus                        `json:"status"`
	Data   GetUnboundTemplateDomainListResponseData `json:"data,omitempty"`
}

type GetUnboundTemplateDomainListResponseData struct {
	Domains []map[string]interface{} `json:"domains,omitempty"`
	Total   float64                  `json:"total,omitempty"`
}

type EditTemplateResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type FirewallSavePolicyGroupRegionalShieldingResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type FirewallSavePolicyGroupAntiLeechResponse struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

type TjkdappsaveFirewallPolicyResponse struct {
	Status APIResponseStatus                     `json:"status"`
	Data   TjkdappsaveFirewallPolicyResponseData `json:"data,omitempty"`
}

type TjkdappsaveFirewallPolicyResponseData struct {
	Info string `json:"info,omitempty"`
}

type TjkdappsortFirewallPolicyResponse struct {
	Status APIResponseStatus                     `json:"status"`
	Data   TjkdappsortFirewallPolicyResponseData `json:"data,omitempty"`
}

type TjkdappsortFirewallPolicyResponseData struct {
	Info string `json:"info,omitempty"`
}

type TjkdappopenFirewallPolicyResponse struct {
	Status APIResponseStatus                     `json:"status"`
	Data   TjkdappopenFirewallPolicyResponseData `json:"data,omitempty"`
}

type TjkdappopenFirewallPolicyResponseData struct {
	Info string `json:"info,omitempty"`
}

type TjkdappstopFirewallPolicyResponse struct {
	Status APIResponseStatus                     `json:"status"`
	Data   TjkdappstopFirewallPolicyResponseData `json:"data,omitempty"`
}

type TjkdappstopFirewallPolicyResponseData struct {
	Info string `json:"info,omitempty"`
}

type TjkdappgetFirewallPolicyResponse struct {
	Status APIResponseStatus                    `json:"status"`
	Data   TjkdappgetFirewallPolicyResponseData `json:"data,omitempty"`
}

type TjkdappgetFirewallPolicyResponseData struct {
	List  TjkdappgetFirewallPolicyResponseDataList `json:"list,omitempty"`
	Total float64                                  `json:"total,omitempty"`
}

type TjkdappgetFirewallPolicyResponseDataList struct {
	Id         string                                        `json:"id,omitempty"`
	Status     string                                        `json:"status,omitempty"`
	Action     string                                        `json:"action,omitempty"`
	ActionData string                                        `json:"action_data,omitempty"`
	Rules      TjkdappgetFirewallPolicyResponseDataListRules `json:"rules,omitempty"`
	CreateAt   string                                        `json:"create_at,omitempty"`
}

type TjkdappgetFirewallPolicyResponseDataListRules struct {
	RuleType string      `json:"rule_type,omitempty"`
	Logic    string      `json:"logic,omitempty"`
	Data     interface{} `json:"data,omitempty"`
	DataType string      `json:"data_type,omitempty"`
}

type TjkdappdeleteFirewallPolicyResponse struct {
	Status APIResponseStatus                       `json:"status"`
	Data   TjkdappdeleteFirewallPolicyResponseData `json:"data,omitempty"`
}

type TjkdappdeleteFirewallPolicyResponseData struct {
	Info string `json:"info,omitempty"`
}

type AddForwardRuleResponse struct {
	Status APIResponseStatus          `json:"status"`
	Data   AddForwardRuleResponseData `json:"data,omitempty"`
}

type AddForwardRuleResponseData struct {
	Info string `json:"info,omitempty"`
}

type DeleteForwardRuleResponse struct {
	Status APIResponseStatus             `json:"status"`
	Data   DeleteForwardRuleResponseData `json:"data,omitempty"`
}

type DeleteForwardRuleResponseData struct {
	Info string `json:"info,omitempty"`
}

type EditRuleResponse struct {
	Status APIResponseStatus    `json:"status"`
	Data   EditRuleResponseData `json:"data,omitempty"`
}

type EditRuleResponseData struct {
	Info string `json:"info,omitempty"`
}

type RuleListResponse struct {
	Status APIResponseStatus    `json:"status"`
	Data   RuleListResponseData `json:"data,omitempty"`
}

type RuleListResponseData struct {
	List  RuleListResponseDataList `json:"list,omitempty"`
	Total string                   `json:"total,omitempty"`
}

type RuleListResponseDataList struct {
	Id         float64                            `json:"id,omitempty"`
	PackageId  float64                            `json:"package_id,omitempty"`
	MemberId   float64                            `json:"member_id,omitempty"`
	Domain     string                             `json:"domain,omitempty"`
	Port       float64                            `json:"port,omitempty"`
	Loading    float64                            `json:"loading,omitempty"`
	SourceList RuleListResponseDataListSourceList `json:"source_list,omitempty"`
	Status     float64                            `json:"status,omitempty"`
	Remark     string                             `json:"remark,omitempty"`
	CreatedAt  string                             `json:"created_at,omitempty"`
	UpdatedAt  string                             `json:"updated_at,omitempty"`
}

type RuleListResponseDataListSourceList struct {
	Ip     string  `json:"ip,omitempty"`
	Port   float64 `json:"port,omitempty"`
	Backup float64 `json:"backup,omitempty"`
}

type GetRuleInfoResponse struct {
	Status APIResponseStatus       `json:"status"`
	Data   GetRuleInfoResponseData `json:"data,omitempty"`
}

type GetRuleInfoResponseData struct {
	Id                string                                   `json:"id,omitempty"`
	PackageId         string                                   `json:"package_id,omitempty"`
	Domain            string                                   `json:"domain,omitempty"`
	Port              string                                   `json:"port,omitempty"`
	Loading           string                                   `json:"loading,omitempty"`
	SourceList        GetRuleInfoResponseDataSourceList        `json:"source_list,omitempty"`
	ChannelLoading    float64                                  `json:"channel_loading,omitempty"`
	ChannelSourceList GetRuleInfoResponseDataChannelSourceList `json:"channel_source_list,omitempty"`
	Status            string                                   `json:"status,omitempty"`
	Remark            string                                   `json:"remark,omitempty"`
	CreatedAt         string                                   `json:"created_at,omitempty"`
	UpdatedAt         string                                   `json:"updated_at,omitempty"`
}

type GetRuleInfoResponseDataSourceList struct {
	Ip     string  `json:"ip,omitempty"`
	Port   string  `json:"port,omitempty"`
	Backup float64 `json:"backup,omitempty"`
}

type GetRuleInfoResponseDataChannelSourceList struct {
	Id     float64 `json:"id,omitempty"`
	Backup float64 `json:"backup,omitempty"`
}

type TijkdappListPackageResponse struct {
	Status APIResponseStatus               `json:"status"`
	Data   TijkdappListPackageResponseData `json:"data,omitempty"`
}

type TijkdappListPackageResponseData struct {
	List  TijkdappListPackageResponseDataList `json:"list,omitempty"`
	Total float64                             `json:"total,omitempty"`
}

type TijkdappListPackageResponseDataList struct {
	Id            float64 `json:"id,omitempty"`
	MemberEmail   string  `json:"member_email,omitempty"`
	MemberId      float64 `json:"member_id,omitempty"`
	PackageName   string  `json:"package_name,omitempty"`
	ExpireTime    string  `json:"expire_time,omitempty"`
	AccessKey     string  `json:"access_key,omitempty"`
	Status        float64 `json:"status,omitempty"`
	CreatedAt     string  `json:"created_at,omitempty"`
	UpdatedAt     string  `json:"updated_at,omitempty"`
	AdminUserId   float64 `json:"admin_user_id,omitempty"`
	AdminInfo     string  `json:"admin_info,omitempty"`
	IpGroupChoose float64 `json:"ip_group_choose,omitempty"`
	ExpireStatus  float64 `json:"expire_status,omitempty"`
}

type TijkdappSavePackageResponse struct {
	Status APIResponseStatus               `json:"status"`
	Data   TijkdappSavePackageResponseData `json:"data,omitempty"`
}

type TijkdappSavePackageResponseData struct {
	Info string `json:"info,omitempty"`
}

type GetChannelListResponse struct {
	Status APIResponseStatus          `json:"status"`
	Data   GetChannelListResponseData `json:"data,omitempty"`
}

type GetChannelListResponseData struct {
	List GetChannelListResponseDataList `json:"list,omitempty"`
}

type GetChannelListResponseDataList struct {
	Id          string `json:"id,omitempty"`
	ChannelName string `json:"channel_name,omitempty"`
}

type ApiNameV5Response struct {
	Status APIResponseStatus `json:"status"`
	Data   interface{}       `json:"data,omitempty"`
}

func (c *EdgeNextClient) CdnHighDefenseIpGetArticleIp(request *CdnHighDefenseIpGetArticleIpRequest) (*CdnHighDefenseIpGetArticleIpResponse, error) {
	response := &CdnHighDefenseIpGetArticleIpResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["CdnHighDefenseIP_getArticleIP"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DnsDomainGetDomainList(request *DnsDomainGetDomainListRequest) (*DnsDomainGetDomainListResponse, error) {
	response := &DnsDomainGetDomainListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DnsDomain_getDomainList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DnsDomainAddDomain(request *DnsDomainAddDomainRequest) (*DnsDomainAddDomainResponse, error) {
	response := &DnsDomainAddDomainResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DnsDomain_addDomain"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DnsDomainBatchAddDomains(request *DnsDomainBatchAddDomainsRequest) (*DnsDomainBatchAddDomainsResponse, error) {
	response := &DnsDomainBatchAddDomainsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DnsDomain_batchAddDomains"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DnsDomainBatchDeleteDomains(request *DnsDomainBatchDeleteDomainsRequest) (*DnsDomainBatchDeleteDomainsResponse, error) {
	response := &DnsDomainBatchDeleteDomainsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DnsDomain_batchDeleteDomains"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DnsDomainGetDomainStat(request *DnsDomainGetDomainStatRequest) (*DnsDomainGetDomainStatResponse, error) {
	response := &DnsDomainGetDomainStatResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DnsDomain_getDomainStat"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DnsDomainGetDomainServers(request *DnsDomainGetDomainServersRequest) (*DnsDomainGetDomainServersResponse, error) {
	response := &DnsDomainGetDomainServersResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DnsDomain_getDomainServers"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DnsDomainGetTasksList(request *DnsDomainGetTasksListRequest) (*DnsDomainGetTasksListResponse, error) {
	response := &DnsDomainGetTasksListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DnsDomain_getTasksList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DnsDomainGetTaskDetail(request *DnsDomainGetTaskDetailRequest) (*DnsDomainGetTaskDetailResponse, error) {
	response := &DnsDomainGetTaskDetailResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DnsDomain_getTaskDetail"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CloudDnsDomainGroupGetGroupList(request *CloudDnsDomainGroupGetGroupListRequest) (*CloudDnsDomainGroupGetGroupListResponse, error) {
	response := &CloudDnsDomainGroupGetGroupListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["CloudDns_DomainGroup_getGroupList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CloudDnsDomainGroupAddGroup(request *CloudDnsDomainGroupAddGroupRequest) (*CloudDnsDomainGroupAddGroupResponse, error) {
	response := &CloudDnsDomainGroupAddGroupResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["CloudDns_DomainGroup_addGroup"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CloudDnsDomainGroupUpdateGroup(request *CloudDnsDomainGroupUpdateGroupRequest) (*CloudDnsDomainGroupUpdateGroupResponse, error) {
	response := &CloudDnsDomainGroupUpdateGroupResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["CloudDns_DomainGroup_updateGroup"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CloudDnsDomainGroupDeleteGroup(request *CloudDnsDomainGroupDeleteGroupRequest) (*CloudDnsDomainGroupDeleteGroupResponse, error) {
	response := &CloudDnsDomainGroupDeleteGroupResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["CloudDns_DomainGroup_deleteGroup"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CloudDnsDomainGroupGetGroupRecordList(request *CloudDnsDomainGroupGetGroupRecordListRequest) (*CloudDnsDomainGroupGetGroupRecordListResponse, error) {
	response := &CloudDnsDomainGroupGetGroupRecordListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["CloudDns_DomainGroup_getGroupRecordList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CloudDnsDomainGroupSaveDomainToGroup(request *CloudDnsDomainGroupSaveDomainToGroupRequest) (*CloudDnsDomainGroupSaveDomainToGroupResponse, error) {
	response := &CloudDnsDomainGroupSaveDomainToGroupResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["CloudDns_DomainGroup_saveDomainToGroup"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CloudDnsDomainGroupGetGroupDomainList(request *CloudDnsDomainGroupGetGroupDomainListRequest) (*CloudDnsDomainGroupGetGroupDomainListResponse, error) {
	response := &CloudDnsDomainGroupGetGroupDomainListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["CloudDns_DomainGroup_getGroupDomainList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CloudDnsDomainGroupGetGroupUndistributedDomainList(request *CloudDnsDomainGroupGetGroupUndistributedDomainListRequest) (*CloudDnsDomainGroupGetGroupUndistributedDomainListResponse, error) {
	response := &CloudDnsDomainGroupGetGroupUndistributedDomainListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["CloudDns_DomainGroup_getGroupUndistributedDomainList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DnsDomainRecordsGetRecordTypes(request *DnsDomainRecordsGetRecordTypesRequest) (*DnsDomainRecordsGetRecordTypesResponse, error) {
	response := &DnsDomainRecordsGetRecordTypesResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DnsDomainRecords_getRecordTypes"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DnsDomainRecordsGetRecordList(request *DnsDomainRecordsGetRecordListRequest) (*DnsDomainRecordsGetRecordListResponse, error) {
	response := &DnsDomainRecordsGetRecordListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DnsDomainRecords_getRecordList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DnsDomainRecordsAddRecord(request *DnsDomainRecordsAddRecordRequest) (*DnsDomainRecordsAddRecordResponse, error) {
	response := &DnsDomainRecordsAddRecordResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DnsDomainRecords_addRecord"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DnsDomainRecordsBatchAddRecords(request *DnsDomainRecordsBatchAddRecordsRequest) (*DnsDomainRecordsBatchAddRecordsResponse, error) {
	response := &DnsDomainRecordsBatchAddRecordsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DnsDomainRecords_batchAddRecords"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DnsDomainRecordsEditRecord(request *DnsDomainRecordsEditRecordRequest) (*DnsDomainRecordsEditRecordResponse, error) {
	response := &DnsDomainRecordsEditRecordResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DnsDomainRecords_editRecord"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DnsDomainRecordsBatchPauseRecords(request *DnsDomainRecordsBatchPauseRecordsRequest) (*DnsDomainRecordsBatchPauseRecordsResponse, error) {
	response := &DnsDomainRecordsBatchPauseRecordsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DnsDomainRecords_batchPauseRecords"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DnsDomainRecordsBatchEnableRecords(request *DnsDomainRecordsBatchEnableRecordsRequest) (*DnsDomainRecordsBatchEnableRecordsResponse, error) {
	response := &DnsDomainRecordsBatchEnableRecordsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DnsDomainRecords_batchEnableRecords"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DnsDomainRecordsDeleteRecord(request *DnsDomainRecordsDeleteRecordRequest) (*DnsDomainRecordsDeleteRecordResponse, error) {
	response := &DnsDomainRecordsDeleteRecordResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DnsDomainRecords_deleteRecord"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DnsDomainRecordsImportRecords(request *DnsDomainRecordsImportRecordsRequest) (*DnsDomainRecordsImportRecordsResponse, error) {
	response := &DnsDomainRecordsImportRecordsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DnsDomainRecords_importRecords"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DnsDomainRecordsExportRecords(request *DnsDomainRecordsExportRecordsRequest) (*DnsDomainRecordsExportRecordsResponse, error) {
	response := &DnsDomainRecordsExportRecordsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DnsDomainRecords_exportRecords"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DnsDomainRecordsGetLines(request *DnsDomainRecordsGetLinesRequest) (*DnsDomainRecordsGetLinesResponse, error) {
	response := &DnsDomainRecordsGetLinesResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DnsDomainRecords_getLines"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DnsDomainRecordsBatchDeleteRecords(request *DnsDomainRecordsBatchDeleteRecordsRequest) (*DnsDomainRecordsBatchDeleteRecordsResponse, error) {
	response := &DnsDomainRecordsBatchDeleteRecordsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DnsDomainRecords_batchDeleteRecords"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DnsDomainRecordsGetRecordGroupsList(request *DnsDomainRecordsGetRecordGroupsListRequest) (*DnsDomainRecordsGetRecordGroupsListResponse, error) {
	response := &DnsDomainRecordsGetRecordGroupsListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DnsDomainRecords_getRecordGroupsList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DnsDomainRecordsAddRecordGroup(request *DnsDomainRecordsAddRecordGroupRequest) (*DnsDomainRecordsAddRecordGroupResponse, error) {
	response := &DnsDomainRecordsAddRecordGroupResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DnsDomainRecords_addRecordGroup"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DnsDomainRecordsAddRecordGroupRelations(request *DnsDomainRecordsAddRecordGroupRelationsRequest) (*DnsDomainRecordsAddRecordGroupRelationsResponse, error) {
	response := &DnsDomainRecordsAddRecordGroupRelationsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DnsDomainRecords_addRecordGroupRelations"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DnsDomainRecordsDeleteRecordGroup(request *DnsDomainRecordsDeleteRecordGroupRequest) (*DnsDomainRecordsDeleteRecordGroupResponse, error) {
	response := &DnsDomainRecordsDeleteRecordGroupResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DnsDomainRecords_deleteRecordGroup"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) UserIpUserIpList(request *UserIpUserIpListRequest) (*UserIpUserIpListResponse, error) {
	response := &UserIpUserIpListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["UserIp_userIpList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) UserIpUserIpAdd(request *UserIpUserIpAddRequest) (*UserIpUserIpAddResponse, error) {
	response := &UserIpUserIpAddResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["UserIp_userIpAdd"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) UserIpUserIpSave(request *UserIpUserIpSaveRequest) (*UserIpUserIpSaveResponse, error) {
	response := &UserIpUserIpSaveResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["UserIp_userIpSave"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) UserIpUserIpDel(request *UserIpUserIpDelRequest) (*UserIpUserIpDelResponse, error) {
	response := &UserIpUserIpDelResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["UserIp_userIpDel"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) UserIpListUserIpItem(request *UserIpListUserIpItemRequest) (*UserIpListUserIpItemResponse, error) {
	response := &UserIpListUserIpItemResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["UserIp_listUserIpItem"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) UserIpAddUserIpItem(request *UserIpAddUserIpItemRequest) (*UserIpAddUserIpItemResponse, error) {
	response := &UserIpAddUserIpItemResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["UserIp_AddUserIpItem"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) UserIpUpdateUserIpItem(request *UserIpUpdateUserIpItemRequest) (*UserIpUpdateUserIpItemResponse, error) {
	response := &UserIpUpdateUserIpItemResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["UserIp_UpdateUserIpItem"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) UserIpBatchDeleteUserIpItem(request *UserIpBatchDeleteUserIpItemRequest) (*UserIpBatchDeleteUserIpItemResponse, error) {
	response := &UserIpBatchDeleteUserIpItemResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["UserIp_BatchDeleteUserIpItem"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) UserIpDeleteAllUserIpItem(request *UserIpDeleteAllUserIpItemRequest) (*UserIpDeleteAllUserIpItemResponse, error) {
	response := &UserIpDeleteAllUserIpItemResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["UserIp_DeleteAllUserIpItem"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) UserIpCopyUserIp(request *UserIpCopyUserIpRequest) (*UserIpCopyUserIpResponse, error) {
	response := &UserIpCopyUserIpResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["UserIp_CopyUserIp"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) UserIpFileSaveIpItem(request *UserIpFileSaveIpItemRequest) (*UserIpFileSaveIpItemResponse, error) {
	response := &UserIpFileSaveIpItemResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["UserIp_FileSaveIpItem"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) ServiceBatchListTask(request *ServiceBatchListTaskRequest) (*ServiceBatchListTaskResponse, error) {
	response := &ServiceBatchListTaskResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["service_batch_ListTask"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) ServiceBatchListSubTask(request *ServiceBatchListSubTaskRequest) (*ServiceBatchListSubTaskResponse, error) {
	response := &ServiceBatchListSubTaskResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["service_batch_ListSubTask"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) WebCdnCleanCacheGetCacheList(request *WebCdnCleanCacheGetCacheListRequest) (*WebCdnCleanCacheGetCacheListResponse, error) {
	response := &WebCdnCleanCacheGetCacheListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["WebCdnCleanCache_getCacheList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) WebCdnCleanCacheSaveCache(request *WebCdnCleanCacheSaveCacheRequest) (*WebCdnCleanCacheSaveCacheResponse, error) {
	response := &WebCdnCleanCacheSaveCacheResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["WebCdnCleanCache_saveCache"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) WebCdnCleanCacheGetTaskList(request *WebCdnCleanCacheGetTaskListRequest) (*WebCdnCleanCacheGetTaskListResponse, error) {
	response := &WebCdnCleanCacheGetTaskListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["WebCdnCleanCache_getTaskList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) WebCdnCleanCacheGetTaskDetail(request *WebCdnCleanCacheGetTaskDetailRequest) (*WebCdnCleanCacheGetTaskDetailResponse, error) {
	response := &WebCdnCleanCacheGetTaskDetailResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["WebCdnCleanCache_getTaskDetail"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) WebCdnPreheatCacheGetPreheatCacheQuota(request *WebCdnPreheatCacheGetPreheatCacheQuotaRequest) (*WebCdnPreheatCacheGetPreheatCacheQuotaResponse, error) {
	response := &WebCdnPreheatCacheGetPreheatCacheQuotaResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["WebCdnPreheatCache_getPreheatCacheQuota"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) WebCdnPreheatCacheGetPreheatCacheList(request *WebCdnPreheatCacheGetPreheatCacheListRequest) (*WebCdnPreheatCacheGetPreheatCacheListResponse, error) {
	response := &WebCdnPreheatCacheGetPreheatCacheListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["WebCdnPreheatCache_getPreheatCacheList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) WebCdnPreheatCacheSavePreheatCache(request *WebCdnPreheatCacheSavePreheatCacheRequest) (*WebCdnPreheatCacheSavePreheatCacheResponse, error) {
	response := &WebCdnPreheatCacheSavePreheatCacheResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["WebCdnPreheatCache_savePreheatCache"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) OplogInfo(request *OplogInfoRequest) (*OplogInfoResponse, error) {
	response := &OplogInfoResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["Oplog_info"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) OplogMap(request *OplogMapRequest) (*OplogMapResponse, error) {
	response := &OplogMapResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["Oplog_map"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) OplogGetOplogs(request *OplogGetOplogsRequest) (*OplogGetOplogsResponse, error) {
	response := &OplogGetOplogsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["Oplog_getOplogs"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CaCertificateSelfAddCa(request *CaCertificateSelfAddCaRequest) (*CaCertificateSelfAddCaResponse, error) {
	response := &CaCertificateSelfAddCaResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["CaCertificateSelf_addCa"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) BatchCaList(request *BatchCaListRequest) (*BatchCaListResponse, error) {
	response := &BatchCaListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["Batch_caList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CaCertificateSelfSaveTextCaInfo(request *CaCertificateSelfSaveTextCaInfoRequest) (*CaCertificateSelfSaveTextCaInfoResponse, error) {
	response := &CaCertificateSelfSaveTextCaInfoResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["CaCertificateSelf_saveTextCaInfo"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CaCertificateSelfEditCaInfo(request *CaCertificateSelfEditCaInfoRequest) (*CaCertificateSelfEditCaInfoResponse, error) {
	response := &CaCertificateSelfEditCaInfoResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["CaCertificateSelf_editCaInfo"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CaCertificateSelfListCa(request *CaCertificateSelfListCaRequest) (*CaCertificateSelfListCaResponse, error) {
	response := &CaCertificateSelfListCaResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["CaCertificateSelf_listCa"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CaCertificateSelfCaExport(request *CaCertificateSelfCaExportRequest) (*CaCertificateSelfCaExportResponse, error) {
	response := &CaCertificateSelfCaExportResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["CaCertificateSelf_caExport"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CaCertificateSelfBatchOperatSsl(request *CaCertificateSelfBatchOperatSslRequest) (*CaCertificateSelfBatchOperatSslResponse, error) {
	response := &CaCertificateSelfBatchOperatSslResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["CaCertificateSelf_batchOperatSsl"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CaCertificateSelfDelCa(request *CaCertificateSelfDelCaRequest) (*CaCertificateSelfDelCaResponse, error) {
	response := &CaCertificateSelfDelCaResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["CaCertificateSelf_delCa"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CaCertificateSelfGetCaDetail(request *CaCertificateSelfGetCaDetailRequest) (*CaCertificateSelfGetCaDetailResponse, error) {
	response := &CaCertificateSelfGetCaDetailResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["CaCertificateSelf_getCaDetail"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CaCertificateSelfEditCaName(request *CaCertificateSelfEditCaNameRequest) (*CaCertificateSelfEditCaNameResponse, error) {
	response := &CaCertificateSelfEditCaNameResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["CaCertificateSelf_editCaName"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CaCertificateApplyAddApplyCa(request *CaCertificateApplyAddApplyCaRequest) (*CaCertificateApplyAddApplyCaResponse, error) {
	response := &CaCertificateApplyAddApplyCaResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["CaCertificateApply_addApplyCa"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CaCertificateApplyGetAddByNsSetting(request *CaCertificateApplyGetAddByNsSettingRequest) (*CaCertificateApplyGetAddByNsSettingResponse, error) {
	response := &CaCertificateApplyGetAddByNsSettingResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["CaCertificateApply_getAddByNsSetting"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DomainGroupSaveGroup(request *DomainGroupSaveGroupRequest) (*DomainGroupSaveGroupResponse, error) {
	response := &DomainGroupSaveGroupResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DomainGroup_saveGroup"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DomainGroupGetGroupList(request *DomainGroupGetGroupListRequest) (*DomainGroupGetGroupListResponse, error) {
	response := &DomainGroupGetGroupListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DomainGroup_getGroupList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DomainGroupDelGroup(request *DomainGroupDelGroupRequest) (*DomainGroupDelGroupResponse, error) {
	response := &DomainGroupDelGroupResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DomainGroup_delGroup"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DomainGroupGetGroupDomainList(request *DomainGroupGetGroupDomainListRequest) (*DomainGroupGetGroupDomainListResponse, error) {
	response := &DomainGroupGetGroupDomainListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DomainGroup_getGroupDomainList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DomainGroupGgtUndistributedDomainList(request *DomainGroupGgtUndistributedDomainListRequest) (*DomainGroupGgtUndistributedDomainListResponse, error) {
	response := &DomainGroupGgtUndistributedDomainListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DomainGroup_ggtUndistributedDomainList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DomainGroupAddGroup(request *DomainGroupAddGroupRequest) (*DomainGroupAddGroupResponse, error) {
	response := &DomainGroupAddGroupResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DomainGroup_addGroup"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DomainGroupSaveDomainToGroup(request *DomainGroupSaveDomainToGroupRequest) (*DomainGroupSaveDomainToGroupResponse, error) {
	response := &DomainGroupSaveDomainToGroupResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DomainGroup_saveDomainToGroup"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DomainGroupGetGroupInfo(request *DomainGroupGetGroupInfoRequest) (*DomainGroupGetGroupInfoResponse, error) {
	response := &DomainGroupGetGroupInfoResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DomainGroup_getGroupInfo"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DomainGroupMoveDomain(request *DomainGroupMoveDomainRequest) (*DomainGroupMoveDomainResponse, error) {
	response := &DomainGroupMoveDomainResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DomainGroup_moveDomain"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) ListDomains(request *ListDomainsRequest) (*ListDomainsResponse, error) {
	response := &ListDomainsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["ListDomains"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) AddDomains(request *AddDomainsRequest) (*AddDomainsResponse, error) {
	response := &AddDomainsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["AddDomains"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) UpdateDomains(request *UpdateDomainsRequest) (*UpdateDomainsResponse, error) {
	response := &UpdateDomainsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["UpdateDomains"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) BindDomainCert(request *BindDomainCertRequest) (*BindDomainCertResponse, error) {
	response := &BindDomainCertResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["BindDomainCert"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) UnBindDomainCert(request *UnBindDomainCertRequest) (*UnBindDomainCertResponse, error) {
	response := &UnBindDomainCertResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["UnBindDomainCert"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DeleteDomains(request *DeleteDomainsRequest) (*DeleteDomainsResponse, error) {
	response := &DeleteDomainsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DeleteDomains"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DisableDomains(request *DisableDomainsRequest) (*DisableDomainsResponse, error) {
	response := &DisableDomainsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DisableDomains"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) EnableDomains(request *EnableDomainsRequest) (*EnableDomainsResponse, error) {
	response := &EnableDomainsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["EnableDomains"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) RefreshDomainsAccess(request *RefreshDomainsAccessRequest) (*RefreshDomainsAccessResponse, error) {
	response := &RefreshDomainsAccessResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["RefreshDomainsAccess"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) ExportDomains(request *ExportDomainsRequest) (*ExportDomainsResponse, error) {
	response := &ExportDomainsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["ExportDomains"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) AddOrigins(request *AddOriginsRequest) (*AddOriginsResponse, error) {
	response := &AddOriginsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["AddOrigins"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) UpdateOrigins(request *UpdateOriginsRequest) (*UpdateOriginsResponse, error) {
	response := &UpdateOriginsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["UpdateOrigins"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DeleteOrigins(request *DeleteOriginsRequest) (*DeleteOriginsResponse, error) {
	response := &DeleteOriginsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DeleteOrigins"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) ListOrigins(request *ListOriginsRequest) (*ListOriginsResponse, error) {
	response := &ListOriginsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["ListOrigins"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) SwitchDomainNodes(request *SwitchDomainNodesRequest) (*SwitchDomainNodesResponse, error) {
	response := &SwitchDomainNodesResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["SwitchDomainNodes"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) SwitchDomainAccessMode(request *SwitchDomainAccessModeRequest) (*SwitchDomainAccessModeResponse, error) {
	response := &SwitchDomainAccessModeResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["SwitchDomainAccessMode"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) UpdateDomainBaseSettings(request *UpdateDomainBaseSettingsRequest) (*UpdateDomainBaseSettingsResponse, error) {
	response := &UpdateDomainBaseSettingsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["UpdateDomainBaseSettings"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) GetDomainBaseSettings(request *GetDomainBaseSettingsRequest) (*GetDomainBaseSettingsResponse, error) {
	response := &GetDomainBaseSettingsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["GetDomainBaseSettings"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) ListBriefDomains(request *ListBriefDomainsRequest) (*ListBriefDomainsResponse, error) {
	response := &ListBriefDomainsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["ListBriefDomains"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) GetDomainTemplates(request *GetDomainTemplatesRequest) (*GetDomainTemplatesResponse, error) {
	response := &GetDomainTemplatesResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["GetDomainTemplates"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) AccessInfoDownload(request *AccessInfoDownloadRequest) (*AccessInfoDownloadResponse, error) {
	response := &AccessInfoDownloadResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["AccessInfoDownload"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) OriginGroupGetOriginGroupList(request *OriginGroupGetOriginGroupListRequest) (*OriginGroupGetOriginGroupListResponse, error) {
	response := &OriginGroupGetOriginGroupListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["OriginGroup_getOriginGroupList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) OriginGroupGetOriginGroupInfo(request *OriginGroupGetOriginGroupInfoRequest) (*OriginGroupGetOriginGroupInfoResponse, error) {
	response := &OriginGroupGetOriginGroupInfoResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["OriginGroup_getOriginGroupInfo"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) OriginGroupAddOriginGroup(request *OriginGroupAddOriginGroupRequest) (*OriginGroupAddOriginGroupResponse, error) {
	response := &OriginGroupAddOriginGroupResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["OriginGroup_addOriginGroup"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) OriginGroupUpdateOriginGroup(request *OriginGroupUpdateOriginGroupRequest) (*OriginGroupUpdateOriginGroupResponse, error) {
	response := &OriginGroupUpdateOriginGroupResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["OriginGroup_updateOriginGroup"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) OriginGroupDelOriginGroup(request *OriginGroupDelOriginGroupRequest) (*OriginGroupDelOriginGroupResponse, error) {
	response := &OriginGroupDelOriginGroupResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["OriginGroup_delOriginGroup"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) OriginGroupBindOriginGroupToDomains(request *OriginGroupBindOriginGroupToDomainsRequest) (*OriginGroupBindOriginGroupToDomainsResponse, error) {
	response := &OriginGroupBindOriginGroupToDomainsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["OriginGroup_bindOriginGroupToDomains"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) OriginGroupGetAllOriginGroups(request *OriginGroupGetAllOriginGroupsRequest) (*OriginGroupGetAllOriginGroupsResponse, error) {
	response := &OriginGroupGetAllOriginGroupsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["OriginGroup_getAllOriginGroups"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) OriginGroupCopyOriginGroup(request *OriginGroupCopyOriginGroupRequest) (*OriginGroupCopyOriginGroupResponse, error) {
	response := &OriginGroupCopyOriginGroupResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["OriginGroup_copyOriginGroup"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) FireWallReportGetBlockList(request *FireWallReportGetBlockListRequest) (*FireWallReportGetBlockListResponse, error) {
	response := &FireWallReportGetBlockListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["FireWallReport_getBlockList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) FireWallReportGetBlockDetails(request *FireWallReportGetBlockDetailsRequest) (*FireWallReportGetBlockDetailsResponse, error) {
	response := &FireWallReportGetBlockDetailsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["FireWallReport_getBlockDetails"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) FireWallReportGetPackageBlockList(request *FireWallReportGetPackageBlockListRequest) (*FireWallReportGetPackageBlockListResponse, error) {
	response := &FireWallReportGetPackageBlockListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["FireWallReport_getPackageBlockList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) FireWallReportGetPackageBlockDetails(request *FireWallReportGetPackageBlockDetailsRequest) (*FireWallReportGetPackageBlockDetailsResponse, error) {
	response := &FireWallReportGetPackageBlockDetailsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["FireWallReport_getPackageBlockDetails"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CcQpsMax(request *CcQpsMaxRequest) (*CcQpsMaxResponse, error) {
	response := &CcQpsMaxResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["cc_qps_max"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CcAttackTimes(request *CcAttackTimesRequest) (*CcAttackTimesResponse, error) {
	response := &CcAttackTimesResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["cc_attack_times"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CcTimesLine(request *CcTimesLineRequest) (*CcTimesLineResponse, error) {
	response := &CcTimesLineResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["cc_times_line"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CcReportStats(request *CcReportStatsRequest) (*CcReportStatsResponse, error) {
	response := &CcReportStatsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["cc_report_stats"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CdnDomainUaispDistribute(request *CdnDomainUaispDistributeRequest) (*CdnDomainUaispDistributeResponse, error) {
	response := &CdnDomainUaispDistributeResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["cdn_domain_uaisp_distribute"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CdnDomainCountryDistribute(request *CdnDomainCountryDistributeRequest) (*CdnDomainCountryDistributeResponse, error) {
	response := &CdnDomainCountryDistributeResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["cdn_domain_country_distribute"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CdnDomainProvinceDistribute(request *CdnDomainProvinceDistributeRequest) (*CdnDomainProvinceDistributeResponse, error) {
	response := &CdnDomainProvinceDistributeResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["cdn_domain_province_distribute"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CdnDomainStatusDistribute(request *CdnDomainStatusDistributeRequest) (*CdnDomainStatusDistributeResponse, error) {
	response := &CdnDomainStatusDistributeResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["cdn_domain_status_distribute"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CdnDomainNodeFlowBandwidth(request *CdnDomainNodeFlowBandwidthRequest) (*CdnDomainNodeFlowBandwidthResponse, error) {
	response := &CdnDomainNodeFlowBandwidthResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["cdn_domain_node_flow_bandwidth"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CdnDomainNodeFlowBandwidthCn2(request *CdnDomainNodeFlowBandwidthCn2Request) (*CdnDomainNodeFlowBandwidthCn2Response, error) {
	response := &CdnDomainNodeFlowBandwidthCn2Response{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["cdn_domain_node_flow_bandwidth_cn2"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CdnDomainNodeFlowBandwidthNode(request *CdnDomainNodeFlowBandwidthNodeRequest) (*CdnDomainNodeFlowBandwidthNodeResponse, error) {
	response := &CdnDomainNodeFlowBandwidthNodeResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["cdn_domain_node_flow_bandwidth_node"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DomainTimes(request *DomainTimesRequest) (*DomainTimesResponse, error) {
	response := &DomainTimesResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["domainTimes"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DomainQps(request *DomainQpsRequest) (*DomainQpsResponse, error) {
	response := &DomainQpsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["domainQps"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CdnDomainFlowLine(request *CdnDomainFlowLineRequest) (*CdnDomainFlowLineResponse, error) {
	response := &CdnDomainFlowLineResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["cdn_domain_flow_line"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CdnDomainBandwidthLine(request *CdnDomainBandwidthLineRequest) (*CdnDomainBandwidthLineResponse, error) {
	response := &CdnDomainBandwidthLineResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["cdn_domain_bandwidth_line"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CdnDomainBandwidth95(request *CdnDomainBandwidth95Request) (*CdnDomainBandwidth95Response, error) {
	response := &CdnDomainBandwidth95Response{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["cdn_domain_bandwidth_95"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CdnDomainPvtimes(request *CdnDomainPvtimesRequest) (*CdnDomainPvtimesResponse, error) {
	response := &CdnDomainPvtimesResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["cdn_domain_pvtimes"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CdnDomainFlowTop(request *CdnDomainFlowTopRequest) (*CdnDomainFlowTopResponse, error) {
	response := &CdnDomainFlowTopResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["cdn_domain_flow_top"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CdnDomainBandwidthTop(request *CdnDomainBandwidthTopRequest) (*CdnDomainBandwidthTopResponse, error) {
	response := &CdnDomainBandwidthTopResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["cdn_domain_bandwidth_top"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CdnDomainTimesTop(request *CdnDomainTimesTopRequest) (*CdnDomainTimesTopResponse, error) {
	response := &CdnDomainTimesTopResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["cdn_domain_times_top"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CdnDomainTimesTopEs(request *CdnDomainTimesTopEsRequest) (*CdnDomainTimesTopEsResponse, error) {
	response := &CdnDomainTimesTopEsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["cdn_domain_times_top_es"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CdnDomainUrlTop(request *CdnDomainUrlTopRequest) (*CdnDomainUrlTopResponse, error) {
	response := &CdnDomainUrlTopResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["cdn_domain_url_top"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CdnDomainRefererTop(request *CdnDomainRefererTopRequest) (*CdnDomainRefererTopResponse, error) {
	response := &CdnDomainRefererTopResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["cdn_domain_referer_top"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CdnDomainStatusTopDownload(request *CdnDomainStatusTopDownloadRequest) (*CdnDomainStatusTopDownloadResponse, error) {
	response := &CdnDomainStatusTopDownloadResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["cdn_domain_status_top_download"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CdnDomainBandwidthDownload(request *CdnDomainBandwidthDownloadRequest) (*CdnDomainBandwidthDownloadResponse, error) {
	response := &CdnDomainBandwidthDownloadResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["cdn_domain_bandwidth_download"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CdnDomainFlowDownload(request *CdnDomainFlowDownloadRequest) (*CdnDomainFlowDownloadResponse, error) {
	response := &CdnDomainFlowDownloadResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["cdn_domain_flow_download"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) TcpBandwidth(request *TcpBandwidthRequest) (*TcpBandwidthResponse, error) {
	response := &TcpBandwidthResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["tcp_bandwidth"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) TcpCcFlaw(request *TcpCcFlawRequest) (*TcpCcFlawResponse, error) {
	response := &TcpCcFlawResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["tcp_cc_flaw"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) WafAttackTimes(request *WafAttackTimesRequest) (*WafAttackTimesResponse, error) {
	response := &WafAttackTimesResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["waf_attack_times"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) WafReportStats(request *WafReportStatsRequest) (*WafReportStatsResponse, error) {
	response := &WafReportStatsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["waf_report_stats"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) WafWebshellEventList(request *WafWebshellEventListRequest) (*WafWebshellEventListResponse, error) {
	response := &WafWebshellEventListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["waf_webshell_event_list"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) WafWebshellEventDetail(request *WafWebshellEventDetailRequest) (*WafWebshellEventDetailResponse, error) {
	response := &WafWebshellEventDetailResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["waf_webshell_event_detail"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) WafAttackEventList(request *WafAttackEventListRequest) (*WafAttackEventListResponse, error) {
	response := &WafAttackEventListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["waf_attack_event_list"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) WafAttackEventDetail(request *WafAttackEventDetailRequest) (*WafAttackEventDetailResponse, error) {
	response := &WafAttackEventDetailResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["waf_attack_event_detail"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) WafScanEventList(request *WafScanEventListRequest) (*WafScanEventListResponse, error) {
	response := &WafScanEventListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["waf_scan_event_list"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) WafScanEventDetail(request *WafScanEventDetailRequest) (*WafScanEventDetailResponse, error) {
	response := &WafScanEventDetailResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["waf_scan_event_detail"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) WafTypeLine(request *WafTypeLineRequest) (*WafTypeLineResponse, error) {
	response := &WafTypeLineResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["waf_type_line"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) LogDownloadTaskTaskList(request *LogDownloadTaskTaskListRequest) (*LogDownloadTaskTaskListResponse, error) {
	response := &LogDownloadTaskTaskListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["LogDownloadTask_taskList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) LogDownloadTaskAddTask(request *LogDownloadTaskAddTaskRequest) (*LogDownloadTaskAddTaskResponse, error) {
	response := &LogDownloadTaskAddTaskResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["LogDownloadTask_addTask"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) LogDownloadTaskCancelTask(request *LogDownloadTaskCancelTaskRequest) (*LogDownloadTaskCancelTaskResponse, error) {
	response := &LogDownloadTaskCancelTaskResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["LogDownloadTask_cancelTask"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) LogDownloadTaskBatchCancelTask(request *LogDownloadTaskBatchCancelTaskRequest) (*LogDownloadTaskBatchCancelTaskResponse, error) {
	response := &LogDownloadTaskBatchCancelTaskResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["LogDownloadTask_batchCancelTask"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) LogDownloadTaskDeleteTask(request *LogDownloadTaskDeleteTaskRequest) (*LogDownloadTaskDeleteTaskResponse, error) {
	response := &LogDownloadTaskDeleteTaskResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["LogDownloadTask_deleteTask"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) LogDownloadTaskBatchDeleteTask(request *LogDownloadTaskBatchDeleteTaskRequest) (*LogDownloadTaskBatchDeleteTaskResponse, error) {
	response := &LogDownloadTaskBatchDeleteTaskResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["LogDownloadTask_batchDeleteTask"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) LogDownloadTaskRegenerateTask(request *LogDownloadTaskRegenerateTaskRequest) (*LogDownloadTaskRegenerateTaskResponse, error) {
	response := &LogDownloadTaskRegenerateTaskResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["LogDownloadTask_regenerateTask"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) LogDownloadFieldConfDownloadFields(request *LogDownloadFieldConfDownloadFieldsRequest) (*LogDownloadFieldConfDownloadFieldsResponse, error) {
	response := &LogDownloadFieldConfDownloadFieldsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["LogDownloadFieldConf_downloadFields"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) LogDownloadTemplateTemplateList(request *LogDownloadTemplateTemplateListRequest) (*LogDownloadTemplateTemplateListResponse, error) {
	response := &LogDownloadTemplateTemplateListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["LogDownloadTemplate_templateList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) LogDownloadTemplateGetTemplateDomainList(request *LogDownloadTemplateGetTemplateDomainListRequest) (*LogDownloadTemplateGetTemplateDomainListResponse, error) {
	response := &LogDownloadTemplateGetTemplateDomainListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["LogDownloadTemplate_getTemplateDomainList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) LogDownloadTemplateAddTemplate(request *LogDownloadTemplateAddTemplateRequest) (*LogDownloadTemplateAddTemplateResponse, error) {
	response := &LogDownloadTemplateAddTemplateResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["LogDownloadTemplate_addTemplate"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) LogDownloadTemplateSaveTemplate(request *LogDownloadTemplateSaveTemplateRequest) (*LogDownloadTemplateSaveTemplateResponse, error) {
	response := &LogDownloadTemplateSaveTemplateResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["LogDownloadTemplate_saveTemplate"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) LogDownloadTemplateDelTemplate(request *LogDownloadTemplateDelTemplateRequest) (*LogDownloadTemplateDelTemplateResponse, error) {
	response := &LogDownloadTemplateDelTemplateResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["LogDownloadTemplate_delTemplate"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) LogDownloadTemplateBatchDelTemplate(request *LogDownloadTemplateBatchDelTemplateRequest) (*LogDownloadTemplateBatchDelTemplateResponse, error) {
	response := &LogDownloadTemplateBatchDelTemplateResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["LogDownloadTemplate_batchDelTemplate"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) LogDownloadTemplateChangeStatus(request *LogDownloadTemplateChangeStatusRequest) (*LogDownloadTemplateChangeStatusResponse, error) {
	response := &LogDownloadTemplateChangeStatusResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["LogDownloadTemplate_changeStatus"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) LogDownloadTemplateBatchChangeStatus(request *LogDownloadTemplateBatchChangeStatusRequest) (*LogDownloadTemplateBatchChangeStatusResponse, error) {
	response := &LogDownloadTemplateBatchChangeStatusResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["LogDownloadTemplate_batchChangeStatus"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) LogDownloadTemplateAllTemplate(request *LogDownloadTemplateAllTemplateRequest) (*LogDownloadTemplateAllTemplateResponse, error) {
	response := &LogDownloadTemplateAllTemplateResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["LogDownloadTemplate_allTemplate"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) LogDownloadTemplateAllTemplateGroup(request *LogDownloadTemplateAllTemplateGroupRequest) (*LogDownloadTemplateAllTemplateGroupResponse, error) {
	response := &LogDownloadTemplateAllTemplateGroupResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["LogDownloadTemplate_allTemplateGroup"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) TjkdPlusPackageGetMemberPackageList(request *TjkdPlusPackageGetMemberPackageListRequest) (*TjkdPlusPackageGetMemberPackageListResponse, error) {
	response := &TjkdPlusPackageGetMemberPackageListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["TjkdPlusPackage_getMemberPackageList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) TjkdPlusPackageGetAllPackage(request *TjkdPlusPackageGetAllPackageRequest) (*TjkdPlusPackageGetAllPackageResponse, error) {
	response := &TjkdPlusPackageGetAllPackageResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["TjkdPlusPackage_getAllPackage"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) TjkdPlusPackageGetPackageInfo(request *TjkdPlusPackageGetPackageInfoRequest) (*TjkdPlusPackageGetPackageInfoResponse, error) {
	response := &TjkdPlusPackageGetPackageInfoResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["TjkdPlusPackage_getPackageInfo"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) TjkdPlusPackageGetPackageIpList(request *TjkdPlusPackageGetPackageIpListRequest) (*TjkdPlusPackageGetPackageIpListResponse, error) {
	response := &TjkdPlusPackageGetPackageIpListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["TjkdPlusPackage_getPackageIpList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) TjkdPlusPackageGetPackageOverview(request *TjkdPlusPackageGetPackageOverviewRequest) (*TjkdPlusPackageGetPackageOverviewResponse, error) {
	response := &TjkdPlusPackageGetPackageOverviewResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["TjkdPlusPackage_getPackageOverview"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) TjkdPlusPackageGetPackagePortList(request *TjkdPlusPackageGetPackagePortListRequest) (*TjkdPlusPackageGetPackagePortListResponse, error) {
	response := &TjkdPlusPackageGetPackagePortListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["TjkdPlusPackage_getPackagePortList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) TjkdPlusPackageSavePackage(request *TjkdPlusPackageSavePackageRequest) (*TjkdPlusPackageSavePackageResponse, error) {
	response := &TjkdPlusPackageSavePackageResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["TjkdPlusPackage_savePackage"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) TjkdPlusPackageSavePackageHealthyConf(request *TjkdPlusPackageSavePackageHealthyConfRequest) (*TjkdPlusPackageSavePackageHealthyConfResponse, error) {
	response := &TjkdPlusPackageSavePackageHealthyConfResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["TjkdPlusPackage_savePackageHealthyConf"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) TjkdPlusForwardRuleSavePlusForwardRule(request *TjkdPlusForwardRuleSavePlusForwardRuleRequest) (*TjkdPlusForwardRuleSavePlusForwardRuleResponse, error) {
	response := &TjkdPlusForwardRuleSavePlusForwardRuleResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["TjkdPlusForwardRule_savePlusForwardRule"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) TjkdPlusForwardRuleBatchAddPlusForwardRule(request *TjkdPlusForwardRuleBatchAddPlusForwardRuleRequest) (*TjkdPlusForwardRuleBatchAddPlusForwardRuleResponse, error) {
	response := &TjkdPlusForwardRuleBatchAddPlusForwardRuleResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["TjkdPlusForwardRule_batchAddPlusForwardRule"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) TjkdPlusForwardRuleBatchSavePlusForwardRule(request *TjkdPlusForwardRuleBatchSavePlusForwardRuleRequest) (*TjkdPlusForwardRuleBatchSavePlusForwardRuleResponse, error) {
	response := &TjkdPlusForwardRuleBatchSavePlusForwardRuleResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["TjkdPlusForwardRule_batchSavePlusForwardRule"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) TjkdPlusForwardRuleDelPlusForwardRule(request *TjkdPlusForwardRuleDelPlusForwardRuleRequest) (*TjkdPlusForwardRuleDelPlusForwardRuleResponse, error) {
	response := &TjkdPlusForwardRuleDelPlusForwardRuleResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["TjkdPlusForwardRule_delPlusForwardRule"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) TjkdPlusForwardRuleGetPlusForwardRuleList(request *TjkdPlusForwardRuleGetPlusForwardRuleListRequest) (*TjkdPlusForwardRuleGetPlusForwardRuleListResponse, error) {
	response := &TjkdPlusForwardRuleGetPlusForwardRuleListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["TjkdPlusForwardRule_getPlusForwardRuleList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) TjkdPlusForwardRuleGetBatchPlusForwardRuleInfo(request *TjkdPlusForwardRuleGetBatchPlusForwardRuleInfoRequest) (*TjkdPlusForwardRuleGetBatchPlusForwardRuleInfoResponse, error) {
	response := &TjkdPlusForwardRuleGetBatchPlusForwardRuleInfoResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["TjkdPlusForwardRule_getBatchPlusForwardRuleInfo"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) TjkdPlusPackageGetPackageDomainList(request *TjkdPlusPackageGetPackageDomainListRequest) (*TjkdPlusPackageGetPackageDomainListResponse, error) {
	response := &TjkdPlusPackageGetPackageDomainListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["TjkdPlusPackage_getPackageDomainList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) TjkdPlusDomainGetTjkdPlusDomainList(request *TjkdPlusDomainGetTjkdPlusDomainListRequest) (*TjkdPlusDomainGetTjkdPlusDomainListResponse, error) {
	response := &TjkdPlusDomainGetTjkdPlusDomainListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["TjkdPlusDomain_getTjkdPlusDomainList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) TjkdPlusDomainAddTjkdPlusDomain(request *TjkdPlusDomainAddTjkdPlusDomainRequest) (*TjkdPlusDomainAddTjkdPlusDomainResponse, error) {
	response := &TjkdPlusDomainAddTjkdPlusDomainResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["TjkdPlusDomain_addTjkdPlusDomain"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) TjkdPlusDomainDelTjkdPlusDomain(request *TjkdPlusDomainDelTjkdPlusDomainRequest) (*TjkdPlusDomainDelTjkdPlusDomainResponse, error) {
	response := &TjkdPlusDomainDelTjkdPlusDomainResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["TjkdPlusDomain_delTjkdPlusDomain"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) NetworkSpeedGetCacheRuleList(request *NetworkSpeedGetCacheRuleListRequest) (*NetworkSpeedGetCacheRuleListResponse, error) {
	response := &NetworkSpeedGetCacheRuleListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["NetworkSpeedGetCacheRuleList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) NetworkSpeedCreateCacheRule(request *NetworkSpeedCreateCacheRuleRequest) (*NetworkSpeedCreateCacheRuleResponse, error) {
	response := &NetworkSpeedCreateCacheRuleResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["NetworkSpeedCreateCacheRule"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) NetworkSpeedUpdateCacheRule(request *NetworkSpeedUpdateCacheRuleRequest) (*NetworkSpeedUpdateCacheRuleResponse, error) {
	response := &NetworkSpeedUpdateCacheRuleResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["NetworkSpeedUpdateCacheRule"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) NetworkSpeedUpdateCacheRuleConfig(request *NetworkSpeedUpdateCacheRuleConfigRequest) (*NetworkSpeedUpdateCacheRuleConfigResponse, error) {
	response := &NetworkSpeedUpdateCacheRuleConfigResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["NetworkSpeedUpdateCacheRuleConfig"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) NetworkSpeedUpdateCacheRuleStatus(request *NetworkSpeedUpdateCacheRuleStatusRequest) (*NetworkSpeedUpdateCacheRuleStatusResponse, error) {
	response := &NetworkSpeedUpdateCacheRuleStatusResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["NetworkSpeedUpdateCacheRuleStatus"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) NetworkSpeedSortCacheRules(request *NetworkSpeedSortCacheRulesRequest) (*NetworkSpeedSortCacheRulesResponse, error) {
	response := &NetworkSpeedSortCacheRulesResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["NetworkSpeedSortCacheRules"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) NetworkSpeedGetGlobalCacheConfig(request *NetworkSpeedGetGlobalCacheConfigRequest) (*NetworkSpeedGetGlobalCacheConfigResponse, error) {
	response := &NetworkSpeedGetGlobalCacheConfigResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["NetworkSpeedGetGlobalCacheConfig"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) NetworkSpeedDeleteCacheRule(request *NetworkSpeedDeleteCacheRuleRequest) (*NetworkSpeedDeleteCacheRuleResponse, error) {
	response := &NetworkSpeedDeleteCacheRuleResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["NetworkSpeedDeleteCacheRule"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) NetworkSpeedGetTemplateConfig(request *NetworkSpeedGetTemplateConfigRequest) (*NetworkSpeedGetTemplateConfigResponse, error) {
	response := &NetworkSpeedGetTemplateConfigResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["NetworkSpeedGetTemplateConfig"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) NetworkSpeedUpdateTemplateConfig(request *NetworkSpeedUpdateTemplateConfigRequest) (*NetworkSpeedUpdateTemplateConfigResponse, error) {
	response := &NetworkSpeedUpdateTemplateConfigResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["NetworkSpeedUpdateTemplateConfig"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) NetworkSpeedGetRules(request *NetworkSpeedGetRulesRequest) (*NetworkSpeedGetRulesResponse, error) {
	response := &NetworkSpeedGetRulesResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["NetworkSpeedGetRules"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) NetworkSpeedCreateRule(request *NetworkSpeedCreateRuleRequest) (*NetworkSpeedCreateRuleResponse, error) {
	response := &NetworkSpeedCreateRuleResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["NetworkSpeedCreateRule"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) NetworkSpeedDeleteRule(request *NetworkSpeedDeleteRuleRequest) (*NetworkSpeedDeleteRuleResponse, error) {
	response := &NetworkSpeedDeleteRuleResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["NetworkSpeedDeleteRule"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) NetworkSpeedSortRules(request *NetworkSpeedSortRulesRequest) (*NetworkSpeedSortRulesResponse, error) {
	response := &NetworkSpeedSortRulesResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["NetworkSpeedSortRules"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) NetworkSpeedUpdateRule(request *NetworkSpeedUpdateRuleRequest) (*NetworkSpeedUpdateRuleResponse, error) {
	response := &NetworkSpeedUpdateRuleResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["NetworkSpeedUpdateRule"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) UpdateRuleTemplate(request *UpdateRuleTemplateRequest) (*UpdateRuleTemplateResponse, error) {
	response := &UpdateRuleTemplateResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["UpdateRuleTemplate"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DeleteRuleTemplate(request *DeleteRuleTemplateRequest) (*DeleteRuleTemplateResponse, error) {
	response := &DeleteRuleTemplateResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DeleteRuleTemplate"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) GetRuleTemplateList(request *GetRuleTemplateListRequest) (*GetRuleTemplateListResponse, error) {
	response := &GetRuleTemplateListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["GetRuleTemplateList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) UnbindRuleTemplate(request *UnbindRuleTemplateRequest) (*UnbindRuleTemplateResponse, error) {
	response := &UnbindRuleTemplateResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["UnbindRuleTemplate"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) BindRuleTemplate(request *BindRuleTemplateRequest) (*BindRuleTemplateResponse, error) {
	response := &BindRuleTemplateResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["BindRuleTemplate"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) ListRuleTpsDomains(request *ListRuleTpsDomainsRequest) (*ListRuleTpsDomainsResponse, error) {
	response := &ListRuleTpsDomainsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["ListRuleTpsDomains"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CreateRuleTemplate(request *CreateRuleTemplateRequest) (*CreateRuleTemplateResponse, error) {
	response := &CreateRuleTemplateResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["CreateRuleTemplate"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) SwitchDomainTemplate(request *SwitchDomainTemplateRequest) (*SwitchDomainTemplateResponse, error) {
	response := &SwitchDomainTemplateResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["SwitchDomainTemplate"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) FirewallPageCfg(request *FirewallPageCfgRequest) (*FirewallPageCfgResponse, error) {
	response := &FirewallPageCfgResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["Firewall_pageCfg"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) FirewallPageCfgHwws(request *FirewallPageCfgHwwsRequest) (*FirewallPageCfgHwwsResponse, error) {
	response := &FirewallPageCfgHwwsResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["Firewall_pageCfgHwws"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) FirewallSavePolicy(request *FirewallSavePolicyRequest) (*FirewallSavePolicyResponse, error) {
	response := &FirewallSavePolicyResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["Firewall_savePolicy"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) FirewallGetPolicy(request *FirewallGetPolicyRequest) (*FirewallGetPolicyResponse, error) {
	response := &FirewallGetPolicyResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["Firewall_getPolicy"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) FirewallGetPolicyByCode(request *FirewallGetPolicyByCodeRequest) (*FirewallGetPolicyByCodeResponse, error) {
	response := &FirewallGetPolicyByCodeResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["Firewall_getPolicyByCode"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) FirewallStatsPolicy(request *FirewallStatsPolicyRequest) (*FirewallStatsPolicyResponse, error) {
	response := &FirewallStatsPolicyResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["Firewall_statsPolicy"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) FirewallOpen(request *FirewallOpenRequest) (*FirewallOpenResponse, error) {
	response := &FirewallOpenResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["Firewall_open"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) FirewallStop(request *FirewallStopRequest) (*FirewallStopResponse, error) {
	response := &FirewallStopResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["Firewall_stop"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) FirewallDelete(request *FirewallDeleteRequest) (*FirewallDeleteResponse, error) {
	response := &FirewallDeleteResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["Firewall_delete"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) FirewallSort(request *FirewallSortRequest) (*FirewallSortResponse, error) {
	response := &FirewallSortResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["Firewall_sort"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) FirewallGetsPolicyByMainid(request *FirewallGetsPolicyByMainidRequest) (*FirewallGetsPolicyByMainidResponse, error) {
	response := &FirewallGetsPolicyByMainidResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["Firewall_getsPolicyByMainid"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) FirewallGetsPolicyByPackageid(request *FirewallGetsPolicyByPackageidRequest) (*FirewallGetsPolicyByPackageidResponse, error) {
	response := &FirewallGetsPolicyByPackageidResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["Firewall_getsPolicyByPackageid"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) FirewallSavePolicyGroup(request *FirewallSavePolicyGroupRequest) (*FirewallSavePolicyGroupResponse, error) {
	response := &FirewallSavePolicyGroupResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["Firewall_savePolicyGroup"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) FirewallGetsPolicyGroupByDomainid(request *FirewallGetsPolicyGroupByDomainidRequest) (*FirewallGetsPolicyGroupByDomainidResponse, error) {
	response := &FirewallGetsPolicyGroupByDomainidResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["Firewall_getsPolicyGroupByDomainid"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) FirewallStopGroup(request *FirewallStopGroupRequest) (*FirewallStopGroupResponse, error) {
	response := &FirewallStopGroupResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["Firewall_stopGroup"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) FirewallOpenGroup(request *FirewallOpenGroupRequest) (*FirewallOpenGroupResponse, error) {
	response := &FirewallOpenGroupResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["Firewall_openGroup"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) FirewallDeleteGroup(request *FirewallDeleteGroupRequest) (*FirewallDeleteGroupResponse, error) {
	response := &FirewallDeleteGroupResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["Firewall_deleteGroup"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) FirewallSortGroup(request *FirewallSortGroupRequest) (*FirewallSortGroupResponse, error) {
	response := &FirewallSortGroupResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["Firewall_sortGroup"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) FirewallGetsPolicyByGroupId(request *FirewallGetsPolicyByGroupIdRequest) (*FirewallGetsPolicyByGroupIdResponse, error) {
	response := &FirewallGetsPolicyByGroupIdResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["Firewall_getsPolicyByGroupId"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) GetPolicyGroupTpl(request *GetPolicyGroupTplRequest) (*GetPolicyGroupTplResponse, error) {
	response := &GetPolicyGroupTplResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["getPolicyGroupTPL"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) GetDdosProtectionConfig(request *GetDdosProtectionConfigRequest) (*GetDdosProtectionConfigResponse, error) {
	response := &GetDdosProtectionConfigResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["GetDdosProtectionConfig"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) UpdateDdosProtectionConfig(request *UpdateDdosProtectionConfigRequest) (*UpdateDdosProtectionConfigResponse, error) {
	response := &UpdateDdosProtectionConfigResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["UpdateDdosProtectionConfig"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) GetWafRuleConfig(request *GetWafRuleConfigRequest) (*GetWafRuleConfigResponse, error) {
	response := &GetWafRuleConfigResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["GetWafRuleConfig"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) UpdateWafRuleConfig(request *UpdateWafRuleConfigRequest) (*UpdateWafRuleConfigResponse, error) {
	response := &UpdateWafRuleConfigResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["UpdateWafRuleConfig"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) GetMemberGlobalTemplate(request *GetMemberGlobalTemplateRequest) (*GetMemberGlobalTemplateResponse, error) {
	response := &GetMemberGlobalTemplateResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["GetMemberGlobalTemplate"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CreateTemplate(request *CreateTemplateRequest) (*CreateTemplateResponse, error) {
	response := &CreateTemplateResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["CreateTemplate"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) CreateDomainTemplate(request *CreateDomainTemplateRequest) (*CreateDomainTemplateResponse, error) {
	response := &CreateDomainTemplateResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["CreateDomainTemplate"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) GetTemplateList(request *GetTemplateListRequest) (*GetTemplateListResponse, error) {
	response := &GetTemplateListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["GetTemplateList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) GetTemplateBindDomainList(request *GetTemplateBindDomainListRequest) (*GetTemplateBindDomainListResponse, error) {
	response := &GetTemplateBindDomainListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["GetTemplateBindDomainList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) BindTemplateDomain(request *BindTemplateDomainRequest) (*BindTemplateDomainResponse, error) {
	response := &BindTemplateDomainResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["BindTemplateDomain"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DeleteTemplate(request *DeleteTemplateRequest) (*DeleteTemplateResponse, error) {
	response := &DeleteTemplateResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["DeleteTemplate"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) BatchConfigTemplate(request *BatchConfigTemplateRequest) (*BatchConfigTemplateResponse, error) {
	response := &BatchConfigTemplateResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["BatchConfigTemplate"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) Iota(request *IotaRequest) (*IotaResponse, error) {
	response := &IotaResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["Iota"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) GetUnboundTemplateDomainList(request *GetUnboundTemplateDomainListRequest) (*GetUnboundTemplateDomainListResponse, error) {
	response := &GetUnboundTemplateDomainListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["GetUnboundTemplateDomainList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) EditTemplate(request *EditTemplateRequest) (*EditTemplateResponse, error) {
	response := &EditTemplateResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["EditTemplate"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) FirewallSavePolicyGroupRegionalShielding(request *FirewallSavePolicyGroupRegionalShieldingRequest) (*FirewallSavePolicyGroupRegionalShieldingResponse, error) {
	response := &FirewallSavePolicyGroupRegionalShieldingResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["Firewall_savePolicyGroupRegionalShielding"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) FirewallSavePolicyGroupAntiLeech(request *FirewallSavePolicyGroupAntiLeechRequest) (*FirewallSavePolicyGroupAntiLeechResponse, error) {
	response := &FirewallSavePolicyGroupAntiLeechResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["Firewall_savePolicyGroupAntiLeech"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) TjkdappsaveFirewallPolicy(request *TjkdappsaveFirewallPolicyRequest) (*TjkdappsaveFirewallPolicyResponse, error) {
	response := &TjkdappsaveFirewallPolicyResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["TjkdappsaveFirewallPolicy"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) TjkdappsortFirewallPolicy(request *TjkdappsortFirewallPolicyRequest) (*TjkdappsortFirewallPolicyResponse, error) {
	response := &TjkdappsortFirewallPolicyResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["TjkdappsortFirewallPolicy"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) TjkdappopenFirewallPolicy(request *TjkdappopenFirewallPolicyRequest) (*TjkdappopenFirewallPolicyResponse, error) {
	response := &TjkdappopenFirewallPolicyResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["TjkdappopenFirewallPolicy"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) TjkdappstopFirewallPolicy(request *TjkdappstopFirewallPolicyRequest) (*TjkdappstopFirewallPolicyResponse, error) {
	response := &TjkdappstopFirewallPolicyResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["TjkdappstopFirewallPolicy"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) TjkdappgetFirewallPolicy(request *TjkdappgetFirewallPolicyRequest) (*TjkdappgetFirewallPolicyResponse, error) {
	response := &TjkdappgetFirewallPolicyResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["TjkdappgetFirewallPolicy"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) TjkdappdeleteFirewallPolicy(request *TjkdappdeleteFirewallPolicyRequest) (*TjkdappdeleteFirewallPolicyResponse, error) {
	response := &TjkdappdeleteFirewallPolicyResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["TjkdappdeleteFirewallPolicy"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) AddForwardRule(request *AddForwardRuleRequest) (*AddForwardRuleResponse, error) {
	response := &AddForwardRuleResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["addForwardRule"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) DeleteForwardRule(request *DeleteForwardRuleRequest) (*DeleteForwardRuleResponse, error) {
	response := &DeleteForwardRuleResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["deleteForwardRule"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) EditRule(request *EditRuleRequest) (*EditRuleResponse, error) {
	response := &EditRuleResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["editRule"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) RuleList(request *RuleListRequest) (*RuleListResponse, error) {
	response := &RuleListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["ruleList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) GetRuleInfo(request *GetRuleInfoRequest) (*GetRuleInfoResponse, error) {
	response := &GetRuleInfoResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["getRuleInfo"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) TijkdappListPackage(request *TijkdappListPackageRequest) (*TijkdappListPackageResponse, error) {
	response := &TijkdappListPackageResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["TIJKDAPP_ListPackage"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) TijkdappSavePackage(request *TijkdappSavePackageRequest) (*TijkdappSavePackageResponse, error) {
	response := &TijkdappSavePackageResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["TIJKDAPP_SavePackage"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) GetChannelList(request *GetChannelListRequest) (*GetChannelListResponse, error) {
	response := &GetChannelListResponse{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["getChannelList"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}

func (c *EdgeNextClient) ApiNameV5(request *ApiNameV5Request) (*ApiNameV5Response, error) {
	response := &ApiNameV5Response{}
	ok, err := c.callTypedDefinition(apiDefinitionsByAPIName["api_name_v5"], request, nil, nil, nil, "", response)
	if !ok {
		return nil, err
	}
	return response, err
}
