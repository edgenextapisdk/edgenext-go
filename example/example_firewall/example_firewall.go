package main

import (
	"fmt"
	"os"

	sdk "github.com/edgenextapisdk/edgenext-go"
)

func main() {
	app_id := os.Getenv("SDK_APP_ID")         //这里是用的环境变量，业务中使用时，请替换成具体的参数
	app_secret := os.Getenv("SDK_APP_SECRET") //这里是用的环境变量，业务中使用时，请替换成具体的参数
	api_pre := os.Getenv("SDK_API_PRE")       //这里是用的环境变量，业务中使用时，请替换成具体的参数

	sdkObj := sdk.Sdk{
		AppId:     app_id,
		AppSecret: app_secret,
		ApiPre:    api_pre,
		Timeout:   30,
	}
	var (
		api       string
		err       error
		reqParams sdk.ReqParams
		resp      *sdk.Response
	)
	//ReqParams 有三个属性，如果用不到，不设置即可：
	//ReqParams.Query 对应的是GET请求的参数，map[string]interface{}
	//ReqParams.Data  对应的是非GET请求的参数，map[string]interface{}
	//ReqParams.Header  对应的是发起请求的header头，map[string]string

	// Response包括响应的全部信息，其中：
	// Response.HttpCode Http请求响应状态码，成功是200
	// Response.RespBody 请求返回的body字符串
	// Response.BizCode 是业务的状态码，1代码请求成功，非1代码请求失败
	// Response.BizMsg  是业务的状态码对应的信息
	// Response.BizData  返回的业务数据，只有BizCode为1时，才会有数据

	// 增加精准访问控制规则集 example
	api = "firewall.policyGroup.save"
	domainID := os.Getenv("SDK_DOMAIN_ID") //这里是用的环境变量，业务中使用时，请替换成具体的参数
	reqParams = sdk.ReqParams{
		Data: map[string]interface{}{
			"name":      "example", // 规则集名称
			"remark":    "example", // 备注
			"from":      "diy",     // 规则集来源，支持diy(用户自定义规则集)/system(系统内置规则集)/quote(引用规则集)
			"domain_id": domainID,  // 用户名ID
		},
	}
	resp, err = sdkObj.Post(api, reqParams)
	if err == nil {
		if resp.BizCode == 1 {
			fmt.Println(api, " 添加精准访问控制规则集成功")
		} else {
			fmt.Println(api, " 添加精准访问控制规则集失败错误: ")
		}
		fmt.Println(api, " http_code: ", resp.HttpCode)
		fmt.Println(api, " body: ", resp.RespBody)
		fmt.Println(api, " biz_code: ", resp.BizCode)
		fmt.Println(api, " biz_msg: ", resp.BizMsg)
		fmt.Println(api, " biz_data: ", resp.BizData)
		if resp.BizCode != 1 {
			return
		}
	} else {
		fmt.Println(api, " 添加精准访问控制规则集失败: ", err)
		return
	}
	fwPolicyGroupID := resp.BizData.(map[string]interface{})["id"]
	fmt.Println("添加精准访问控制规则集成功: ", fwPolicyGroupID)

	// 为规则集 example, 增加 请求类型 和 单IP单URL请求频率 规则
	api = "firewall.policy.save"
	reqParams = sdk.ReqParams{
		Data: map[string]interface{}{
			//"id": 0,                      // 如果指定ID，则是编辑已有的规则
			"group_id":  fwPolicyGroupID, // 规则集ID
			"domain_id": domainID,        // 域名ID
			"remark":    "",              // 备注
			"type":      "plus",          // scdn域名专属的类型plus
			"action":    "block",         // 封禁
			"action_data": map[string]interface{}{ // 封禁10分钟
				"time_unit": "minute", // 支持: second, minute, hour, day(最大7天)
				"interval":  10,
			},
			"rules": []map[string]interface{}{
				{
					"rule_type": "url_type",          // 请求类型
					"logic":     "belongs",           // 属于
					"data":      []string{"dynamic"}, // 动态
				},
				{
					"rule_type": "ip_url_rate_limit", // 单IP单URL请求频率
					"logic":     "greater_than",      // 大于
					"data": map[string]interface{}{ // 10 秒内频率大于 100 次
						"interval": 10,  // 计数时间间隔，单位秒
						"reqs":     100, // 时间间隔内，请求次数
					},
				},
			},
		},
	}
	resp, err = sdkObj.Post(api, reqParams)
	if err == nil {
		if resp.BizCode == 1 {
			fmt.Println(api, "为规则集", fwPolicyGroupID, "添加 请求类型 和 单IP单URL请求频率 规则成功")
		} else {
			fmt.Println(api, "为规则集", fwPolicyGroupID, "添加 请求类型 和 单IP单URL请求频率 规则错误: ")
		}
		fmt.Println(api, " http_code: ", resp.HttpCode)
		fmt.Println(api, " body: ", resp.RespBody)
		fmt.Println(api, " biz_code: ", resp.BizCode)
		fmt.Println(api, " biz_msg: ", resp.BizMsg)
		fmt.Println(api, " biz_data: ", resp.BizData)
		if resp.BizCode != 1 {
			return
		}
	} else {
		fmt.Println(api, "为规则集", fwPolicyGroupID, "添加 请求类型 和 单IP单URL请求频率 规则失败: ", err)
		return
	}
	fwPolicyId := resp.BizData.(map[string]interface{})["id"]
	fmt.Println("为规则集", fwPolicyGroupID, "添加 请求类型 和 单IP单URL请求频率 规则成功: ", fwPolicyId)

	// 为规则集 example, 增加 IP类型 和 IP请求频率 规则
	api = "firewall.policy.save"
	reqParams = sdk.ReqParams{
		Data: map[string]interface{}{
			//"id": 0,                      // 如果指定ID，则是编辑已有的规则
			"group_id":  fwPolicyGroupID, // 规则集ID
			"domain_id": domainID,        // 域名ID
			"remark":    "",              // 备注
			"type":      "plus",          // scdn域名专属的类型plus
			"action":    "verification",  // 人机验证
			"action_data": map[string]interface{}{
				"next_rules": 1,        // 继续执行下一规则集 0/1
				"cc":         1,        // 继续执行CC防护 0/1
				"waf":        0,        // 继续执行漏洞攻击防护 0/1
				"type":       "cookie", // 人机验证方式：cookie(Cookie验证), js(JS验证), captcha(智能验证码)
			},
			"rules": []map[string]interface{}{
				{
					"rule_type": "ip_type", // IP类型
					"logic":     "equals",  // 等于
					"data": []string{ // 支持的IP类型
						"botnet",      // 僵尸网络
						"Tor",         // 洋葱路由
						"Proxy",       // 代理池IP
						"IpBlackList", // IP信誉库黑名单
						"FakeUA",      // 伪造搜索引擎
						//"spider",            // 搜索引擎
						//"partner",           // 合作伙伴
						//"monitor",           // 监控
						//"aggregator",        // 聚合器
						//"Social",            // 社交网络
						//"ad",                // 广告网络
						//"BackLinkWatch",     // 反向链接检测
						//"IDC",               // IDC数据
						//"MaliciousUA",       // 恶意UA
						//"Export",            // 公共出口
					},
				},
				{
					"rule_type": "ip_rate_limit", // IP请求频率
					"logic":     "greater_than",  // 大于
					"data": map[string]interface{}{
						"interval": 20, // 20秒，单位：秒
						"reqs":     3,  // 请求3次
					},
				},
			},
		},
	}
	resp, err = sdkObj.Post(api, reqParams)
	if err == nil {
		if resp.BizCode == 1 {
			fmt.Println(api, "为规则集", fwPolicyGroupID, "添加 IP类型 和 IP请求频率 规则成功")
		} else {
			fmt.Println(api, "为规则集", fwPolicyGroupID, "添加 IP类型 和 IP请求频率 规则错误: ")
		}
		fmt.Println(api, " http_code: ", resp.HttpCode)
		fmt.Println(api, " body: ", resp.RespBody)
		fmt.Println(api, " biz_code: ", resp.BizCode)
		fmt.Println(api, " biz_msg: ", resp.BizMsg)
		fmt.Println(api, " biz_data: ", resp.BizData)
		if resp.BizCode != 1 {
			return
		}
	} else {
		fmt.Println(api, "为规则集", fwPolicyGroupID, "添加 IP类型 和 IP请求频率 规则失败: ", err)
		return
	}
	fwPolicyId = resp.BizData.(map[string]interface{})["id"] // 接收新增的规则ID
	fmt.Println("为规则集", fwPolicyGroupID, "添加 IP类型 和 IP请求频率 规则成功: ", fwPolicyId)

	// 为规则集 example, 增加 URL 和 单IP单URL请求频率 规则
	api = "firewall.policy.save"
	reqParams = sdk.ReqParams{
		Data: map[string]interface{}{
			//"id": 0,                      // 如果指定ID，则是编辑已有的规则
			"group_id":  fwPolicyGroupID, // 规则集ID
			"domain_id": domainID,        // 域名ID
			"remark":    "",              // 备注
			"type":      "plus",          // scdn域名专属的类型plus
			"action":    "block",         // 封禁
			"action_data": map[string]interface{}{ // 封禁10分钟
				"time_unit": "minute", // 支持: second, minute, hour, day(最大7天)
				"interval":  10,
			},
			"rules": []map[string]interface{}{
				{
					"rule_type": "url",                           // URL规则
					"logic":     "contains",                      // 包含
					"data":      []string{"/login", "/register"}, // 规则值，必须以/开头
				},
				{
					"rule_type": "ip_url_rate_limit", // 单IP单URL请求频率
					"logic":     "greater_than",      // 大于
					"data": map[string]interface{}{ // 10秒25次请求
						"interval": 10, // 计数时间间隔，单位秒
						"reqs":     25, // 时间间隔内，请求次数
					},
				},
			},
		},
	}
	resp, err = sdkObj.Post(api, reqParams)
	if err == nil {
		if resp.BizCode == 1 {
			fmt.Println(api, "为规则集", fwPolicyGroupID, "添加 URL 和 单IP单URL请求频率 规则成功")
		} else {
			fmt.Println(api, "为规则集", fwPolicyGroupID, "添加 URL 和 单IP单URL请求频率 规则错误: ")
		}
		fmt.Println(api, " http_code: ", resp.HttpCode)
		fmt.Println(api, " body: ", resp.RespBody)
		fmt.Println(api, " biz_code: ", resp.BizCode)
		fmt.Println(api, " biz_msg: ", resp.BizMsg)
		fmt.Println(api, " biz_data: ", resp.BizData)
		if resp.BizCode != 1 {
			return
		}
	} else {
		fmt.Println(api, "为规则集", fwPolicyGroupID, "添加 URL 和 单IP单URL请求频率 规则失败: ", err)
		return
	}
	fmt.Println("为规则集", fwPolicyGroupID, "添加 URL 和 单IP单URL请求频率 规则成功: ", fwPolicyId)
}
