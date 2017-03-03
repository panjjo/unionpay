package unionpay

import "fmt"

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

// 订购类API
type Order struct {
	bizType string
	url     string
}

// 初始化一个订购类
func NewOrder(url string) (o Order, err error) {
	fmt.Printf("%+v", certData)
	if certData.CertId == "" || certData.EncryptId == "" {
		err = fmt.Errorf("请先配置证书信息")
		return
	}
	return Order{bizType: "001001", url: url}, nil
}

// 实名认证接口
func (n *Order) RealNameAuth(orderid string, accNo, bindid string, customer *CustomerInfo) (result interface{}, err error) {
	request := sysParams()
	request["txnType"] = "72"
	request["txnSubType"] = "01"
	request["bindId"] = bindid
	request["bizType"] = n.bizType
	request["txnTime"] = getTxnTime()
	request["orderId"] = orderid
	request["accNo"] = getaccNo(accNo)
	request["customerInfo"] = getCustomerInfo(customer)
	request["signature"], _ = Sign(request)
	return POST(n.url, request)

}
