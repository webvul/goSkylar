package lib

import (
	"strings"
	"strconv"
	"time"
	"gopkg.in/mgo.v2/bson"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"bytes"

	"net/http"
	"crypto/tls"
	"github.com/m3ng9i/go-utils/encoding"
	"goSkylar/server/conf"
	"log"
	"os/exec"
	"os"
	"path/filepath"
	"net"
	"errors"
)

func TimeToStr(intTime int64) string {
	timeLayout := "2006-01-02 15:04:05"                     //转化所需模板
	dataTimeStr := time.Unix(intTime, 0).Format(timeLayout) //设置时间戳 使用模板格式化为日期字符串
	return dataTimeStr
}

func TimeToData(intTime int64) string {
	timeLayout := "2006-01-02"                              //转化所需模板
	dataTimeStr := time.Unix(intTime, 0).Format(timeLayout) //设置时间戳 使用模板格式化为日期字符串
	return dataTimeStr
}

func CurrentTime() string {
	return TimeToStr(time.Now().Unix())
}

func DateToStr(intTime int64) string {
	timeLayout := "2006-01-02"
	dataTimeStr := time.Unix(intTime, 0).Format(timeLayout)
	return dataTimeStr
}

// 获取当天的日期
func CurrentDate() string {
	return DateToStr(time.Now().Unix())
}

//ip转换，如果是ip则转换成ip段
func Iptransfer(ip string) string {
	if strings.Contains(ip, "/") {
		return ip
	}
	return ip + "/32"
}

func InterfaceToStr(inter interface{}) (s string) {
	tempStr := ""
	switch inter.(type) {
	case nil:
		tempStr = ""
		break
	case string:
		tempStr = inter.(string)
		break
	case float64:
		tempStr = strconv.FormatFloat(inter.(float64), 'f', -1, 64)
		break
	case int64:
		tempStr = strconv.FormatInt(inter.(int64), 10)
		break
	case int:
		tempStr = strconv.Itoa(inter.(int))
		break
	case bool:
		tempStr = strconv.FormatBool(inter.(bool))
	case bson.ObjectId:
		tempStr = inter.(bson.ObjectId).Hex()
	case []interface{}:
		tempStr, _ = JsonToString(inter)
	case []int:
		tempStr, _ = JsonToString(inter)
	case []int64:
		tempStr, _ = JsonToString(inter)
	case []float32:
		tempStr, _ = JsonToString(inter)
	case []float64:
		tempStr, _ = JsonToString(inter)
	case map[string]interface{}:
		tempStr, _ = JsonToString(inter)
	case map[string]string:
		tempStr, _ = JsonToString(inter)
	case time.Time:
		tempStr = inter.(time.Time).String()
	default:
		tempStr = "Error! Not Found Type!"
	}
	return tempStr
}

func JsonToString(inter interface{}) (string, error) {
	by, err := json.Marshal(inter)
	if err != nil {
		return "", err
	} else {
		return string(by), nil
	}
}

//md5加密
func Md5Str(str string) string {
	strMd5 := md5.New()
	strMd5.Write([]byte(str))
	return hex.EncodeToString(strMd5.Sum(nil))
}

func MobileMessage(mobile_number string, message_content string) error {
	by, err := encoding.Utf8ToGbk([]byte(message_content))
	if err != nil {
		return err
	}
	message_content_main := string(by)

	uri := conf.MESSAGE_URI

	if err != nil {
		return err
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 5, //超时时间
	}
	body := `<?xml version="1.0" encoding="UTF-8"?>` +
		`<S:Envelope xmlns:S="http://schemas.xmlsoap.org/soap/envelope/">` +
		`<S:Header><AuthenticationHeader xmlns="http://mms.360buy.com/services/MoblMsgSender">` +
		`<Token>666a8EN3oIijHY+KjS+2mg==</Token>` +
		`</AuthenticationHeader>` +
		`</S:Header><S:Body>` +
		`<ns2:jdMmSender xmlns:ns2="http://moblMsgSender.server.ws.sender.mobilePhoneMsg.jd.com/">` +
		`<arg0 xmlns="">` + mobile_number + `</arg0><arg1 xmlns="">` + message_content_main + `</arg1>` +
		`<arg2 xmlns="">yunwei.alarm</arg2><arg3 xmlns=""></arg3></ns2:jdMmSender>` +
		`</S:Body>` +
		`</S:Envelope>`
	req, err := http.NewRequest("POST", uri, bytes.NewBuffer([]byte(body)))
	if err != nil {
		return err
	}
	req.Host = "mms.360buy.com"
	_, err = client.Do(req)
	if err != nil {
		return err
	}
	return err
}

// 发送告警短信，主机停止心跳
func SendAlarmMessage(agentIp string) error {
	phones := strings.Split(conf.MESSAGE_NUMBER, ",")
	for _, v := range phones {
		err := MobileMessage(v, "主机："+agentIp+"停止心跳，请核实")
		if err != nil {
			log.Println("号码：", v, "发送短信失败")
			continue
		}
	}
	return nil
}

// 发送重启短信，主机已经重连
func SendRebootMessage(agentIp string) error {
	phones := strings.Split(conf.MESSAGE_NUMBER, ",")
	for _, v := range phones {
		err := MobileMessage(v, "主机："+agentIp+"已经重启")
		if err != nil {
			log.Println("号码：", v, "发送短信失败")
			continue
		}
	}
	return nil
}

// 当前一轮任务已经扫描完成，发送短信
func SendSMessage(msg string) error {
	phones := strings.Split(conf.MESSAGE_NUMBER, ",")
	for _, v := range phones {
		err := MobileMessage(v, msg)
		if err != nil {
			log.Println("号码：", v, "发送短信失败")
			continue
		}
	}
	return nil
}

func getCurrentPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	i := strings.LastIndex(path, "/")
	if i < 0 {
		i = strings.LastIndex(path, "\\")
	}
	if i < 0 {
		return "", errors.New(`error: Can't find "/" or "\".`)
	}
	return string(path[0: i+1]), nil
}
func IpStringToInt(ipstring string) int {

	ipSegs := strings.Split(ipstring, ".")
	var ipInt int = 0
	var pos uint = 24
	for _, ipSeg := range ipSegs {
		tempInt, _ := strconv.Atoi(ipSeg)
		tempInt = tempInt << pos
		ipInt = ipInt | tempInt
		pos -= 8
	}
	return ipInt
}

func IpIntToString(ipInt int) string {
	ipSegs := make([]string, 4)
	var length int = len(ipSegs)
	buffer := bytes.NewBufferString("")
	for i := 0; i < length; i++ {
		tempInt := ipInt & 0xFF
		ipSegs[length-i-1] = strconv.Itoa(tempInt)
		ipInt = ipInt >> 8
	}
	for i := 0; i < length; i++ {
		buffer.WriteString(ipSegs[i])
		if i < length-1 {
			buffer.WriteString(".")
		}
	}
	return buffer.String()
}

func IpRangeToIPs(ipaddr string) []string {
	ipRangeList := strings.Split(ipaddr, "/")
	if len(ipRangeList) != 2 {
		return []string{}
	}
	ip := ipRangeList[0]
	mask, err := strconv.Atoi(ipRangeList[1])
	if err != nil {
		return []string{}
	}
	var result []string
	if mask > 32 || mask < 0 {
		log.Println("netmask error")
		return result
	}

	maskhead := 0xFFFFFFFF
	for i := 1; i <= 32-mask; i++ {
		maskhead = maskhead << 1
	}

	masktail := 0xFFFFFFFF
	for i := 1; i <= mask; i++ {
		masktail = masktail >> 1
	}
	ipint := IpStringToInt(ip)
	IPintstart := ipint & maskhead
	IPintend := ipint | masktail

	for i := IPintstart; i <= IPintend; i++ {
		result = append(result, IpIntToString(i))
	}
	return result
}

func IpRangeToIPsSplit(ipaddr string, mask int) []string {
	var result []string
	if mask > 32 || mask < 0 {
		log.Println("netmask error")
		return result
	}

	maskhead := 0xFFFFFFFF
	for i := 1; i <= 32-mask; i++ {
		maskhead = maskhead << 1
	}

	masktail := 0xFFFFFFFF
	for i := 1; i <= mask; i++ {
		masktail = masktail >> 1
	}
	ipint := IpStringToInt(ipaddr)
	IPintstart := ipint & maskhead
	IPintend := ipint | masktail

	for i := IPintstart; i <= IPintend; i++ {
		result = append(result, IpIntToString(i))
	}
	return result
}

func TransferJson(str string) string {
	str = strings.Replace(str, "ip", `"ip"`, -1)
	str = strings.Replace(str, "port", `"port"`, -1)
	str = strings.Replace(str, `"port"s`, `"ports"`, -1)
	str = strings.Replace(str, "proto", `"proto"`, -1)
	str = strings.Replace(str, "status", `"status"`, -1)
	str = strings.Replace(str, "reason", `"reason"`, -1)
	str = strings.Replace(str, "ttl", `"ttl"`, -1)
	return str
}

func TimeToStrForPrint(intTime int64) string {
	timeLayout := "2006-01-02 15:04:05"                     //转化所需模板
	dataTimeStr := time.Unix(intTime, 0).Format(timeLayout) //设置时间戳 使用模板格式化为日期字符串
	return dataTimeStr
}

func CurrentTimeForPrint() string {
	return TimeToStrForPrint(time.Now().Unix())
}

func DealError(err error) error {
	if err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

func IsPublicIP(IP net.IP) bool {
	if IP.IsLoopback() || IP.IsLinkLocalMulticast() || IP.IsLinkLocalUnicast() {
		return false
	}
	if ip4 := IP.To4(); ip4 != nil {
		i := ip4[0]
		u := ip4[1]
		switch true {
		case i == 10:
			return false
		case i == 172 && u >= 16 && u <= 31:
			return false
		case i == 192 && u == 168:
			return false
		case i == 100 && u >= 64 && u <= 127:
			return false
		case i == 169 && u == 254:
			return false
		case i == 11:
			return false
		case i == 19:
			return false
		case i == 20:
			return false
		case i == 21:
			return false
		case i == 22:
			return false
		default:
			return true
		}
	}
	return false
}
