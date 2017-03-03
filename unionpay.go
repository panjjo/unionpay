package unionpay

import "strings"

type UnionpayData struct {
	version       string `json:"version"`
	encoding      string `json:"encoding"`
	certId        string `json:"certId"`
	encryptCertId string `json:"encryptCertId"`
}

var (
	accessType  string = "0"
	merId       string
	frontUrl    string
	channelType string
	encoding    string = "utf-8"
	signMethod  string = "01"
	version     string = "5.1.0"
)

type Config struct {
	// 商户接入类型0：商户直连接入1：收单机构接入 2：平台商户接入
	AccessType string `json:"accessType"`
	// 商户代码
	MerId string `json:"merId"`
	// 前台通知地址
	FrontUrl string `json:"frontUrl"`
	// 渠道类型 05：语音07：互联网08：移动 16：数字机顶盒
	ChannelType string `json:"channelType"`
	// 版本号 默认5.1.0
	Version string `json:"version"`
}

// 设置用户配置
func SetConfig(config *Config) {
	accessType = config.AccessType
	merId = config.MerId
	frontUrl = config.FrontUrl
	channelType = config.ChannelType
	if config.Version != "" {
		version = config.Version
	}
}

func sysParams() map[string]string {
	return map[string]string{
		"version":       version,
		"encoding":      encoding,
		"certId":        certData.CertId,
		"signMethod":    signMethod,
		"encryptCertId": certData.EncryptId,
		"accessType":    accessType,
		"channelType":   channelType,
		"merId":         merId,
	}
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
	return Base64Encode([]byte("{" + strings.Join(tmp, "&") + "}"))
}
