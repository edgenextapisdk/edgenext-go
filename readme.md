# EdgeNext Go SDK

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.14-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

Official Go SDK for EdgeNext SCDN API. EdgeNext is a global CDN and edge computing platform. Visit [edgenext.com](https://www.edgenext.com/) for more information.

## Features

- ✅ Full RESTful API support (GET, POST, PUT, PATCH, DELETE)
- ✅ Automatic request signing with SHA256 algorithm
- ✅ Type-safe request/response handling
- ✅ Built-in error handling and response parsing
- ✅ Support for query parameters, request body, and custom headers
- ✅ Configurable timeout settings

## Installation

```bash
go get github.com/edgenextapisdk/edgenext-go
```

## Requirements

- Go >= 1.18

## Quick Start

```go
package main

import (
    "fmt"
    "os"
    sdk "github.com/edgenextapisdk/edgenext-go"
)

func main() {
    // Initialize SDK
    sdkObj := sdk.Sdk{
        AppId:     os.Getenv("SDK_APP_ID"),
        AppSecret: os.Getenv("SDK_APP_SECRET"),
        ApiPre:    os.Getenv("SDK_API_PRE"),
        Timeout:   30,
    }

    // Make a GET request
    reqParams := sdk.ReqParams{
        Query: map[string]interface{}{
            "page":     1,
            "pagesize": 10,
        },
    }

    resp, err := sdkObj.Get("test.sdk.get", reqParams)
    if err != nil {
        fmt.Printf("Request error: %v\n", err)
        return
    }

    if resp.BizCode == 1 {
        fmt.Println("Success:", resp.BizData)
    } else {
        fmt.Printf("Business error: %s\n", resp.BizMsg)
    }
}
```

## Configuration

### SDK Parameters

| Parameter | Type | Description | Required |
|-----------|------|-------------|----------|
| `AppId` | `string` | Your application ID provided by EdgeNext | Yes |
| `AppSecret` | `string` | Your application secret for request signing | Yes |
| `ApiPre` | `string` | API endpoint prefix (consult operations team for details) | Yes |
| `Timeout` | `int` | Request timeout in seconds (default: 30) | No |
| `Debug` | `bool` | Enable debug mode | No |

## Usage

### Request Parameters

The `ReqParams` struct supports three optional properties:

- **Query**: Query parameters for GET requests (`map[string]interface{}`)
- **Data**: Request body for non-GET requests (`map[string]interface{}`)
- **Headers**: Custom HTTP headers (`map[string]string`)

### Response Structure

The `Response` struct contains the following fields:

- **HttpCode**: HTTP status code (200 for success)
- **RespBody**: Raw response body as string
- **BizCode**: Business status code (1 = success, others = failure)
- **BizMsg**: Business status message
- **BizData**: Business data (only available when BizCode is 1)

### Examples

#### GET Request

```go
api := "test.sdk.get"
reqParams := sdk.ReqParams{
    Query: map[string]interface{}{
        "page":     1,
        "pagesize": 10,
        "data": map[string]interface{}{
            "name":   "example",
            "domain": "example.com",
        },
    },
}

resp, err := sdkObj.Get(api, reqParams)
if err == nil && resp.BizCode == 1 {
    fmt.Println("Success:", resp.BizData)
}
```

#### POST Request

```go
api := "test.sdk.post"
reqParams := sdk.ReqParams{
    Data: map[string]interface{}{
        "name": 1,
        "age":  10,
        "data": map[string]interface{}{
            "name":   "example",
            "domain": "example.com",
        },
    },
}

resp, err := sdkObj.Post(api, reqParams)
if err == nil && resp.BizCode == 1 {
    fmt.Println("Success:", resp.BizData)
}
```

#### PUT Request

```go
api := "test.sdk.put"
reqParams := sdk.ReqParams{
    Data: map[string]interface{}{
        "id":   10,
        "name": "updated_name",
    },
}

resp, err := sdkObj.Put(api, reqParams)
if err == nil && resp.BizCode == 1 {
    fmt.Println("Success:", resp.BizData)
}
```

#### DELETE Request

```go
api := "test.sdk.delete"
reqParams := sdk.ReqParams{
    Data: map[string]interface{}{
        "id": 10,
    },
}

resp, err := sdkObj.Delete(api, reqParams)
if err == nil && resp.BizCode == 1 {
    fmt.Println("Success:", resp.BizData)
}
```

## Authentication & Signing

The SDK uses SHA256-based request signing to ensure data integrity during transmission.

### Signing Algorithm

1. **Client Side**: 
   - Base64 encode the request parameters
   - Sign the encoded parameters with `app_secret` using SHA256
   - Include the signature in each request

2. **Server Side**:
   - Recalculate the signature using the same algorithm
   - Compare signatures to verify request authenticity

All requests are automatically signed by the SDK - no manual intervention required.

## Important Notes

- **URI and Query Parameters**: For all requests, URI and GET parameters are separated. For example, for `https://apiv4.local.com/V4/version?v=1`, the `v=1` parameter must be passed through `ReqParams.Query`, not in the URI.

- **Request Format**: All requests use JSON format by default.

- **Response Format**: All responses are in JSON format.

## Error Handling

The SDK returns errors in two ways:

1. **Network/HTTP Errors**: Returned as `error` in the function return value
2. **Business Logic Errors**: Indicated by `BizCode != 1` in the `Response` struct

Always check both:

```go
resp, err := sdkObj.Get(api, reqParams)
if err != nil {
    // Handle network/HTTP errors
    fmt.Printf("Request failed: %v\n", err)
    return
}

if resp.BizCode != 1 {
    // Handle business logic errors
    fmt.Printf("Business error: %s (code: %d)\n", resp.BizMsg, resp.BizCode)
    return
}

// Success
fmt.Println("Data:", resp.BizData)
```

## Debug Mode

Enable debug mode to get detailed request/response information:

```go
sdkObj := sdk.Sdk{
    AppId:     "your_app_id",
    AppSecret: "your_app_secret",
    ApiPre:    "your_api_prefix",
    Debug:     true, // Enable debug mode
}
```

## Support

For API endpoint configuration and other inquiries, please contact the EdgeNext operations team.

Visit [edgenext.com](https://www.edgenext.com/) for more information.
