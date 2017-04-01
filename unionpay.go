package unionpay

import "strings"

type UnionpayData struct {
	version       string `json:"version"`
	encoding      string `json:"encoding"`
	certId        string `json:"certId"`
	encryptCertId string `json:"encryptCertId"`
}

var (
	merId      string
	frontUrl   string
	encoding   string = "utf-8"
	signMethod string = "01"
	version    string = "5.1.0"
	baseUrl    string = "https://gateway.test.95516.com/"
)

//初始使用的配置
type Config struct {
	// 版本号 默认5.1.0
	Version string

	// 请求银联的地址
	Url string

	// 商户代码
	MerId string

	// 前台通知地址
	FrontUrl string

	// pfx 证书路径,和同时传入PrivatePath和CertPath 效果一样
	PfxPath string

	// pfx 证书的密码
	PfxPwd string

	// 验签私钥证书地址，传入pfx此路径可不传
	// openssl pkcs12 -in xxxx.pfx -nodes -out server.pem 生成为原生格式pem 私钥
	// openssl rsa -in server.pem -out server.key  生成为rsa格式私钥文件
	PrivatePath string

	// 验签证书地址,传入pfx此路径可以不传
	// openssl pkcs12 -in xxxx.pfx -clcerts -nokeys -out key.cert
	CertPath string

	// 加密证书地址
	EncryptCertPath string
}

func Init(config *Config) error {
	if err := LoadCert(config); err != nil {
		return err
	}
	SetConfig(config)
	return nil
}

// 设置用户配置
func SetConfig(config *Config) {
	merId = config.MerId
	frontUrl = config.FrontUrl
	if config.Version != "" {
		version = config.Version
	}
	if config.Url != "" {
		baseUrl = config.Url
	}
}

func sysParams(c ApiConfig, data *RequestParams) map[string]string {
	request := map[string]string{
		"version":       version,
		"encoding":      encoding,
		"certId":        certData.CertId,
		"signMethod":    signMethod,
		"encryptCertId": certData.EncryptId,
		"accessType":    c.accessType,
		"channelType":   c.channelType,
		"bizType":       c.bizType,
		"merId":         merId,
	}
	if data.TnxTime == "" {
		request["txnTime"] = getTxnTime()
	} else {
		request["txnTime"] = data.TnxTime
	}
	if data.OrderId == "" {
		data.OrderId = randomString(10)
	}
	request["orderId"] = data.OrderId
	request["accNo"] = getaccNo(data.AccNo)
	request["customerInfo"] = getCustomerInfo(data.Customer)
	if data.Extend != "" {
		request["reqReserved"] = data.Extend
	}
	if data.Reserved != nil {
		list := []string{}
		for k, v := range data.Reserved {
			list = append(list, k+"&"+v)
		}
		if len(list) > 0 {
			request["reserved"] = "{" + strings.Join(list, "&") + "}"
		}
	}

	return request
}

func getTxnTime() string {
	return sec2Str("20060102150405", getNowSec())
}
func getaccNo(no string) string {
	str, _ := EncryptData(no)
	return str
}
func getCustomerInfo(customer *CustomerInfo) string {
	enmap := map[string]string{}
	other := map[string]string{}
	m := obj2Map(*customer)
	for k, v := range m {
		if v.(string) != "" {
			if k == "phoneNo" || k == "cvn2" || k == "expired" {
				enmap[k] = v.(string)
			} else {
				other[k] = v.(string)
			}
		}
	}
	if len(enmap) > 0 {
		tmp := []string{}
		for k, v := range enmap {
			tmp = append(tmp, k+"="+v)
		}
		str := strings.Join(tmp, "&")
		enc, _ := EncryptData(str)
		other["encryptedInfo"] = enc
	}
	tmp := []string{}
	for k, v := range other {
		tmp = append(tmp, k+"="+v)
	}
	return base64Encode([]byte("{" + strings.Join(tmp, "&") + "}"))
}
