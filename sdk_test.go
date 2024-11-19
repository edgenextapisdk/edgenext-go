package sdk

import (
	"fmt"
	"os"
	"testing"
)

func TestGet(t *testing.T) {
	var sdk = Sdk{
		// 认证用的ak和sk硬编码到代码中或者明文存储都有很大的安全风险，建议在配置文件或者环境变量中密文存放，使用时解密，确保安全；
		// 本示例以ak和sk保存在环境变量中为例，运行本示例前请先在本地环境中设置环境变量X-App-Id和X-App-Secret。
		AppId:     os.Getenv("X-App-Id"),
		AppSecret: os.Getenv("X-App-Secret"),
		ApiPre:    "http://127.0.0.1:60041/api/V4/",
		UserId:    1,
		Timeout:   30,
	}
	var err error
	var reqParams ReqParams
	var resp *Response
	api := "Web.Domain.Info"
	reqParams = ReqParams{
		Query: map[string]interface{}{
			"domain": 101153,
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
