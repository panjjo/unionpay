package unionpay

import "fmt"

type ApiConfig struct {
	Url         string
	bizType     string
	accessType  string
	channelType string
}

// 用户数据
type CustomerInfo struct {
	// 证件类型 01：身份证 02：军官证 03：护照 04：回乡证 05：台胞证 06：警官证 07：士兵证 99：其它证件
	CertifTp string `json:"certifTp"`

	// 证件ID
	CertifId string `json:"certifId"`

	// 名称
	CustomerNm string `json:"customerNm"`

	// 短信验证码
	SmsCode string `json:"smsCode"`

	//使用敏感信息加密证书对 ANSI X9.8 格式的 PIN 加密，并做 Base64 编码
	Pin string `json:"pin"`

	// 三位长度的cvn 信用卡反面后三位
	Cvn2 string `json:"cvn2"`

	// YYMM四位长度的过期时间
	Expired string `json:"expired"`

	// 开卡时预留的手机号
	PhoneNo string `json:"phoneNo"`
}

// 入参
type RequestParams struct {
	// 订单ID 默认随机长度为10的字符串
	OrderId string
	// 订单日期 format=20060102150405 默认当前时间
	TnxTime string

	// 账户（银行卡号）
	AccNo string

	// 用户验证信息
	Customer *CustomerInfo

	// 扩展参数 交易应答或callback时原样返回 非必填
	Extend string

	// 保留域 非必填
	Reserved map[string]string
}

// 订购类API
type Order struct {
	c ApiConfig
}

// 初始化一个订购类
func NewOrder(c ApiConfig) (o Order, err error) {
	if certData.CertId == "" || certData.EncryptId == "" {
		err = fmt.Errorf("请先配置证书信息")
		return
	}
	if c.Url == "" {
		c.Url = baseUrl
	}
	c.bizType = "001001"
	c.channelType = "07"
	c.accessType = "0"
	return Order{c}, nil
}

// 实名认证接口
func (n *Order) RealNameAuth(bindid string, data *RequestParams) (result interface{}, err error) {
	request := sysParams(n.c, data)
	request["bindId"] = bindid
	request["txnType"] = "72"
	request["txnSubType"] = "01"
	request["signature"], _ = Sign(request)
	return POST(n.c.Url+"/gateway/api/backTransReq.do", request)
}

// 代付类API
type PayForAnother struct {
	c ApiConfig
}

// 初始化一个代付类
func NewPayForAnother(c ApiConfig) (o PayForAnother, err error) {
	if certData.CertId == "" || certData.EncryptId == "" {
		err = fmt.Errorf("请先配置证书信息")
		return
	}
	if c.Url == "" {
		c.Url = baseUrl
	}
	c.bizType = "000401"
	c.channelType = "07"
	c.accessType = "0"
	return PayForAnother{c}, nil
}

// 实名认证接口
func (n *PayForAnother) RealNameAuth(bindid string, data *RequestParams) (result interface{}, err error) {
	request := sysParams(n.c, data)
	request["bindId"] = bindid
	request["txnType"] = "72"
	request["txnSubType"] = "01"
	request["signature"], _ = Sign(request)
	return POST(n.c.Url+"/gateway/api/backTransReq.do", request)
}

// 支付接口
func (n *PayForAnother) Pay(amount int64, data *RequestParams) (result interface{}, err error) {
	request := sysParams(n.c, data)
	request["txnAmt"] = fmt.Sprintf("%d", amount)
	request["txnType"] = "12"
	request["txnSubType"] = "00"
	request["currencyCode"] = "156"
	request["signature"], _ = Sign(request)
	return POST(n.c.Url+"/gateway/api/backTransReq.do", request)
}
