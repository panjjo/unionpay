package unionpay

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strings"
	"time"
)

/**
  获取当前时间戳
*/
func getNowSec() int64 {
	return time.Now().Unix()
}

/**
  获取当前时间戳
*/
func str2Sec(layout, str string) int64 {
	tm2, _ := time.ParseInLocation(layout, str, time.Local)
	return tm2.Unix()
}

/**
  获取当前时间
*/
func sec2Str(layout string, sec int64) string {
	t := time.Unix(sec, 0)
	nt := t.Format(layout)
	return nt
}

/**
  struct 转化成map
*/
func obj2Map(obj interface{}) map[string]interface{} {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	var data = make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		if v.Field(i).CanInterface() && t.Field(i).Tag.Get("json") != "-" {
			data[t.Field(i).Tag.Get("json")] = v.Field(i).Interface()
		}
	}
	return data
}

// base64 加密
func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// base64 解密
func Base64Decode(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}

type MapSorter []SortItem

type SortItem struct {
	Key string      `json:"key"`
	Val interface{} `json:"val"`
}

func (ms MapSorter) Len() int {
	return len(ms)
}
func (ms MapSorter) Less(i, j int) bool {
	return ms[i].Key < ms[j].Key // 按键排序
}
func (ms MapSorter) Swap(i, j int) {
	ms[i], ms[j] = ms[j], ms[i]
}
func MapSortByKey(m map[string]string, step1, step2 string) string {
	ms := make(MapSorter, 0, len(m))

	for k, v := range m {
		ms = append(ms, SortItem{k, v})
	}
	sort.Sort(ms)
	s := []string{}
	for _, p := range ms {
		s = append(s, p.Key+step1+p.Val.(string))
	}
	return strings.Join(s, step2)
}
func TimeoutClient() *http.Client {
	connectTimeout := time.Duration(20 * time.Second)
	readWriteTimeout := time.Duration(30 * time.Second)
	return &http.Client{
		Transport: &http.Transport{
			Dial:                timeoutDialer(connectTimeout, readWriteTimeout),
			MaxIdleConnsPerHost: 200,
			DisableKeepAlives:   true,
		},
	}
}
func timeoutDialer(cTimeout time.Duration,
	rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(rwTimeout))
		return conn, nil
	}
}

// 发送post请求
func POST(requrl string, request map[string]string) (interface{}, error) {
	c := TimeoutClient()
	resp, err := c.Post(requrl, "application/x-www-form-urlencoded", strings.NewReader(Http_build_query(request)))
	if err != nil {
		return resp, err
	}
	if resp.StatusCode != 200 {
		return resp, fmt.Errorf("http request response StatusCode:%v", resp.StatusCode)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return data, err
	}
	var fields []string
	fields = strings.Split(string(data), "&")

	vals := url.Values{}
	for _, field := range fields {
		f := strings.SplitN(field, "=", 2)
		if len(f) >= 2 {
			key, val := f[0], f[1]
			vals.Set(key, val)
		}
	}
	return Verify(vals)
}

// urlencode
func Http_build_query(params map[string]string) string {
	qs := url.Values{}
	for k, v := range params {
		qs.Add(k, v)
	}
	return qs.Encode()
}
