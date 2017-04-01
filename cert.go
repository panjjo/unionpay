package unionpay

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/url"

	"golang.org/x/crypto/pkcs12"
)

var certData *Cert

// 证书信息结构体
type Cert struct {
	// 私钥 签名使用
	Private *rsa.PrivateKey
	// 证书 与私钥为一套
	Cert *x509.Certificate
	// 签名证书ID
	CertId string
	// 加密证书
	EncryptCert *x509.Certificate
	// 公钥 加密验签使用
	Public *rsa.PublicKey
	// 加密公钥ID
	EncryptId string
}

// 根据配置加载证书信息
func LoadCert(info *Config) (err error) {
	certData = &Cert{}
	certData.EncryptCert, err = ParseCertificateFromFile(info.EncryptCertPath)
	if err != nil {
		err = fmt.Errorf("encryptCert ERR:%v", err)
		return
	}
	certData.EncryptId = fmt.Sprintf("%v", certData.EncryptCert.SerialNumber)
	certData.Public = certData.EncryptCert.PublicKey.(*rsa.PublicKey)
	if info.PfxPwd != "" && info.PfxPath != "" {
		certData.Private, certData.Cert, err = ParserPfxToCert(info.PfxPath, info.PfxPwd)
	} else if info.CertPath != "" && info.PrivatePath != "" {
		certData.Cert, err = ParseCertificateFromFile(info.CertPath)
		if err != nil {
			return
		}
		certData.Private, err = ParsePrivateFromFile(info.PrivatePath)
		if err != nil {
			return
		}
	} else {
		err = fmt.Errorf("请输入正确的证书地址或者密码")
	}
	certData.CertId = fmt.Sprintf("%v", certData.Cert.SerialNumber)
	return
}

// 根据银联获取到的PFX文件和密码来解析出里面包含的私钥(rsa)和证书(x509)
func ParserPfxToCert(path string, password string) (private *rsa.PrivateKey, cert *x509.Certificate, err error) {
	var pfxData []byte
	pfxData, err = ioutil.ReadFile(path)
	if err != nil {
		return
	}
	var priv interface{}
	priv, cert, err = pkcs12.Decode(pfxData, password)
	if err != nil {
		return
	}
	private = priv.(*rsa.PrivateKey)
	return
}

// 根据文件名解析出私钥 ,文件必须是rsa 私钥格式。
// openssl pkcs12 -in xxxx.pfx -nodes -out server.pem 生成为原生格式pem 私钥
// openssl rsa -in server.pem -out server.key  生成为rsa格式私钥文件
func ParsePrivateFromFile(path string) (private *rsa.PrivateKey, err error) {
	// Read the private key
	pemData, err := ioutil.ReadFile(path)
	if err != nil {
		err = fmt.Errorf("read key file: %s", err)
		return
	}

	// Extract the PEM-encoded data block
	block, _ := pem.Decode(pemData)
	if block == nil {
		err = fmt.Errorf("bad key data: %s", "not PEM-encoded")
		return
	}
	if got, want := block.Type, "RSA PRIVATE KEY"; got != want {
		err = fmt.Errorf("unknown key type %q, want %q", got, want)
		return
	}

	// Decode the RSA private key
	private, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		err = fmt.Errorf("bad private key: %s", err)
		return
	}
	return
}

// 根据文件名解析出证书
// openssl pkcs12 -in xxxx.pfx -clcerts -nokeys -out key.cert
func ParseCertificateFromFile(path string) (cert *x509.Certificate, err error) {
	// Read the verify sign certification key
	pemData, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	// Extract the PEM-encoded data block
	block, _ := pem.Decode(pemData)
	if block == nil {
		err = fmt.Errorf("bad key data: %s", "not PEM-encoded")
		return
	}
	if got, want := block.Type, "CERTIFICATE"; got != want {
		err = fmt.Errorf("unknown key type %q, want %q", got, want)
		return
	}

	// Decode the certification
	cert, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		err = fmt.Errorf("bad private key: %s", err)
		return
	}
	return
}

// 利用加密证书公钥对数据加密
func EncryptData(data string) (res string, err error) {
	if certData.EncryptId == "" {
		err = fmt.Errorf("请先配置加密证书信息")
		return
	}
	rng := rand.Reader
	signer, err := rsa.EncryptPKCS1v15(rng, certData.Public, []byte(data))
	res = base64Encode(signer)
	return
}

// sign 做签
func Sign(request map[string]string) (string, error) {
	str := mapSortByKey(request, "=", "&")
	rng := rand.Reader
	hashed := sha256.Sum256([]byte(fmt.Sprintf("%x", sha256.Sum256([]byte(str)))))
	signer, err := rsa.SignPKCS1v15(rng, certData.Private, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}
	return base64Encode(signer), nil
}

// 返回数据验签
func Verify(vals url.Values) (res interface{}, err error) {
	var signature string
	kvs := map[string]string{}
	for k := range vals {
		if k == "signature" {
			signature = vals.Get(k)
			continue
		}
		if vals.Get(k) == "" {
			continue
		}
		kvs[k] = vals.Get(k)
	}
	str := mapSortByKey(kvs, "=", "&")
	hashed := sha256.Sum256([]byte(fmt.Sprintf("%x", sha256.Sum256([]byte(str)))))
	var inSign []byte
	inSign, err = base64Decode(signature)
	if err != nil {
		return nil, fmt.Errorf("解析返回signature失败 %v", err)
	}

	err = rsa.VerifyPKCS1v15(certData.Public, crypto.SHA256, hashed[:], inSign)
	if err != nil {
		return nil, fmt.Errorf("返回数据验签失败 ERR:%v", err)
	}
	return kvs, nil
}
