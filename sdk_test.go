package sdk

import (
	"fmt"
	"os"
	"testing"
)

func TestMapStrval(t *testing.T) {
	var vInt8 int8 = 3
	var vuInt8 uint8 = 3
	var vInt16 int16 = 3
	var vuInt16 uint16 = 3
	var vInt32 int32 = 3
	var vuInt32 uint32 = 3
	var vInt64 int64 = 3
	var vuInt64 uint64 = 3
	data := map[string]interface{}{
		"age":          10,
		"name":         "name",
		"int8":         vInt8,
		"uint8":        vuInt8,
		"int16":        vInt16,
		"uint16":       vuInt16,
		"int32":        vInt32,
		"uint32":       vuInt32,
		"int64":        vInt64,
		"uint64":       vuInt64,
		"bytes":        []byte("hello"),
		"ids_ints":     []int{1, 2, 3},
		"ids_int64s":   []int64{1, 2, 3},
		"fen_float32s": []float32{1.0, 2.5, 3.9},
		"fen_float64s": []float64{1.0, 2.5, 3.9},
		"bools":        []bool{true, false, true},
		"fen_strs":     []float64{1, 2, 3},
		"childs": []interface{}{
			map[string]interface{}{
				"age":  1,
				"name": "name_childs_name",
			},
		},
	}
	fmt.Println("MapStrval data: ", data)
	query, err := MapStrval(data)
	fmt.Println("MapStrval data strval: ", query)
	fmt.Println("MapStrval err: ", err)
	fmt.Println("")
}

func TestGet(t *testing.T) {
	var sdk = Sdk{
		AppId:     os.Getenv("SDK_APP_ID"),
		AppSecret: os.Getenv("SDK_APP_SECERT"),
		ApiPre:    os.Getenv("SDK_API_PRE"),
		UserId:    1,
		Timeout:   30,
	}
	var err error
	var reqParams ReqParams
	var resp *Response
	api := "test.sdk.get"
	reqParams = ReqParams{
		Query: map[string]interface{}{
			"page":     1,
			"pagesize": 10,
			"data": map[string]interface{}{
				"name":   "name名称",
				"domain": "baidu.com",
			},
		},
	}
	resp, err = sdk.Get(api, reqParams)
	fmt.Println(api, " http_code: ", resp.HttpCode)
	fmt.Println(api, " biz_code: ", resp.BizCode)
	fmt.Println(api, " biz_msg: ", resp.BizMsg)
	fmt.Println(api, " biz_data: ", resp.BizData)
	fmt.Println(api, " err: ", err)
}

func TestPost(t *testing.T) {
	var sdk = Sdk{
		AppId:     os.Getenv("SDK_APP_ID"),
		AppSecret: os.Getenv("SDK_APP_SECERT"),
		ApiPre:    os.Getenv("SDK_API_PRE"),
		UserId:    1,
		Timeout:   30,
	}
	var err error
	var reqParams ReqParams
	var resp *Response

	api := "test.sdk.post"
	reqParams = ReqParams{
		Data: map[string]interface{}{
			"name": 1,
			"age":  10,
			"data": map[string]interface{}{
				"name":   "name名称",
				"domain": "baidu.com",
			},
		},
	}
	resp, err = sdk.Post(api, reqParams)
	fmt.Println(api, " http_code: ", resp.HttpCode)
	fmt.Println(api, " biz_code: ", resp.BizCode)
	fmt.Println(api, " biz_msg: ", resp.BizMsg)
	fmt.Println(api, " biz_data: ", resp.BizData)
	fmt.Println(api, " err: ", err)
}

func TestPut(t *testing.T) {
	var sdk = Sdk{
		AppId:     os.Getenv("SDK_APP_ID"),
		AppSecret: os.Getenv("SDK_APP_SECERT"),
		ApiPre:    os.Getenv("SDK_API_PRE"),
		UserId:    1,
		Timeout:   30,
	}
	var err error
	var reqParams ReqParams
	var resp *Response

	api := "test.sdk.put"
	reqParams = ReqParams{
		Data: map[string]interface{}{
			"name": 1,
			"age":  10,
			"data": map[string]interface{}{
				"name":   "name名称",
				"domain": "baidu.com",
			},
		},
	}
	resp, err = sdk.Put(api, reqParams)
	fmt.Println(api, " http_code: ", resp.HttpCode)
	fmt.Println(api, " biz_code: ", resp.BizCode)
	fmt.Println(api, " biz_msg: ", resp.BizMsg)
	fmt.Println(api, " biz_data: ", resp.BizData)
	fmt.Println(api, " err: ", err)
}

func TestDelete(t *testing.T) {
	var sdk = Sdk{
		AppId:     os.Getenv("SDK_APP_ID"),
		AppSecret: os.Getenv("SDK_APP_SECERT"),
		ApiPre:    os.Getenv("SDK_API_PRE"),
		UserId:    1,
		Timeout:   30,
	}
	var err error
	var reqParams ReqParams
	var resp *Response

	api := "test.sdk.delete"
	reqParams = ReqParams{
		Data: map[string]interface{}{
			"id": 10,
		},
	}
	resp, err = sdk.Delete(api, reqParams)
	fmt.Println(api, " http_code: ", resp.HttpCode)
	fmt.Println(api, " biz_code: ", resp.BizCode)
	fmt.Println(api, " biz_msg: ", resp.BizMsg)
	fmt.Println(api, " biz_data: ", resp.BizData)
	fmt.Println(api, " err: ", err)
}
