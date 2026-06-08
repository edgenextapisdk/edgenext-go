# EdgeNext Go SDK

Official Go SDK for EdgeNext SCDN APIs.

## Installation

```bash
go get github.com/edgenextapisdk/edgenext-go
```

## Requirements

- Go 1.18 or later
- EdgeNext `AppId`, `AppSecret`, and API endpoint prefix

## Quick Start

Use `NewEdgeNextClient` to initialize the SDK and call generated API wrappers.

```go
package main

import (
    "fmt"
    "os"

    edgenext "github.com/edgenextapisdk/edgenext-go"
)

func main() {
    client := edgenext.NewEdgeNextClient(edgenext.EdgeNextClientConfig{
        AppId:     os.Getenv("EDGENEXT_APP_ID"),
        AppSecret: os.Getenv("EDGENEXT_APP_SECRET"),
        ApiPre:    os.Getenv("EDGENEXT_API_PRE"),
        UserId:    1,
        Timeout:   30,
    }, edgenext.WithLanguage("en"))

    resp, err := client.ListDomains(&edgenext.ListDomainsRequest{
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
```

## Configuration

`EdgeNextClientConfig` is the recommended way to create a client.

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `AppId` | `string` | Yes | Application ID provided by EdgeNext. |
| `AppSecret` | `string` | Yes | Application secret used for request signing. |
| `ApiPre` | `string` | Yes | API endpoint prefix, for example `https://example.com/api/v5`. |
| `UserId` | `int` | No | User ID passed in SDK payload fields. |
| `Timeout` | `int` | No | Request timeout in seconds. Defaults to `30` when unset. |
| `Debug` | `bool` | No | Reserved debug flag. |

If you already have an `Sdk` instance, initialize the wrapper client with `NewEdgeNextClientFromSDK`:

```go
sdkObj := &edgenext.Sdk{
    AppId:     os.Getenv("EDGENEXT_APP_ID"),
    AppSecret: os.Getenv("EDGENEXT_APP_SECRET"),
    ApiPre:    os.Getenv("EDGENEXT_API_PRE"),
    UserId:    1,
    Timeout:   30,
}

client := edgenext.NewEdgeNextClientFromSDK(sdkObj)
```

## Generated API Wrappers

Generated wrapper methods provide typed request structs for common APIs. Prefer these methods when a wrapper exists.

### Query Domains

```go
resp, err := client.ListDomains(&edgenext.ListDomainsRequest{
    Page:     1,
    PageSize: 20,
    Domain:   "example.com",
})
```

### Add Domain

```go
resp, err := client.AddDomains(&edgenext.AddDomainsRequest{
    Domain:  "www.example.com",
    GroupId: 1,
    Origins: []map[string]interface{}{
        {"addr": "192.0.2.10", "weight": 1},
    },
})
```

### Update Domain

```go
resp, err := client.UpdateDomains(&edgenext.UpdateDomainsRequest{
    DomainId: 123,
    Remark:   "updated by sdk",
})
```

### Delete Domains

```go
resp, err := client.DeleteDomains(&edgenext.DeleteDomainsRequest{
    Ids: []int{123, 456},
})
```

## Low-Level Requests

`EdgeNextClient` also exposes the original SDK request methods:

- `Request(api, method, params)`
- `Get(api, params)`
- `Post(api, params)`
- `Put(api, params)`
- `Delete(api, params)`

Use these methods when you need to call an API that does not have a generated wrapper yet.

```go
resp, err := client.Post("test.sdk.post", edgenext.ReqParams{
    Data: map[string]interface{}{
        "name": "example",
    },
    Headers: map[string]string{
        "X-Request-Id": "req-001",
    },
})
```

For GET requests, put query parameters in `ReqParams.Query`:

```go
resp, err := client.Get("test.sdk.get", edgenext.ReqParams{
    Query: map[string]interface{}{
        "page":      1,
        "page_size": 20,
    },
})
```

## Request Parameters

`ReqParams` supports:

| Field | Type | Description |
| --- | --- | --- |
| `Query` | `map[string]interface{}` | URL query parameters. |
| `Data` | `map[string]interface{}` | JSON request body for non-GET requests. |
| `Headers` | `map[string]string` | Per-request custom headers. |

Default headers can be configured on the client:

```go
client := edgenext.NewEdgeNextClient(config,
    edgenext.WithLanguage("en"),
    edgenext.WithDefaultHeaders(map[string]string{
        "X-Client": "my-service",
    }),
)
```

Per-request headers override default headers with the same key.

## Response Handling

Generated wrapper methods return endpoint-specific response structs, for example `(*ListDomainsResponse, error)`. Dynamic low-level calls return `(*APIResponse, error)`.

| Field | Description |
| --- | --- |
| `Status.Code` | Business status code from `status.code`. |
| `Status.Message` | Business message from `status.message`. |
| `Data` | Typed business payload generated from apidoc `@apiSuccess data.*` fields for endpoint wrappers. |

Always check `error`. HTTP errors, malformed responses, and non-success business status codes are returned as `error`. On success, generated wrappers expose typed apidoc response fields:

```go
resp, err := client.ListDomains(&edgenext.ListDomainsRequest{Page: 1})
if err != nil {
    fmt.Printf("request failed: %v\n", err)
    return
}

fmt.Println(resp.Data.Total)
```

## Signing

The SDK signs every request automatically. You only need to provide `AppId` and `AppSecret`.

The signer adds SDK authentication headers, request metadata, and an HMAC-SHA256 signature before sending the HTTP request.

## Notes

- Do not hard-code `AppId` or `AppSecret` in source code. Use environment variables, secret managers, or encrypted configuration.
- `ApiPre` should be the endpoint prefix assigned to your account or environment.
- Do not put query strings directly in the API path. Use `ReqParams.Query` instead.
- Request bodies are sent as JSON.
- Responses are expected to be JSON with a `status` object.

## Support

For API endpoint configuration, credentials, and API behavior questions, contact the EdgeNext operations or support team.
