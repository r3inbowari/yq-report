package main

import (
	"bytes"
	"encoding/json"
	"github.com/robfig/cron"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

/**
 * 禁止多次提交!!!
 */
func main() {

	log.Println("[INFO] 当前系统时间 ", time.Now())
	config := GetConfig()
	log.Println("[INFO] 同心战\"疫\" 重回美好")
	log.Println("[INFO] 开始疫情上报服务")
	config.SendWeChatMessage("疫情上报功能已启用", "已成功订阅")
	c := cron.New()
	_ = c.AddFunc("0 16 8 * * ?", func() {
		log.Println("[INFO] 开始上报当天数据")
		session := GetJSession()
		session.GDUPTLogin(config)
		session.GDUPTAddForm(config)
	})
	c.Start()
	//session := GetJSession()
	//session.GDUPTLogin(config)
	//session.GDUPTAddForm(config)
	select {}
}

// 方糖密钥 -> 微信报告
func (con *Config) SendWeChatMessage(title, content string) {
	GetRequest("https://sc.ftqq.com/" + con.FT + ".send?desp=" + content + "&text=" + title)
}

func GetRequest(url string) string {
	// log.Println(url)
	_, err := http.Get(url)
	if err != nil {
		log.Println("[INFO] 方糖调用错误")
		return ""
	}
	return ""

	//client := &http.Client{Timeout: 5 * time.Second}
	//resp, err := client.Get(url)
	//if err != nil {
	//	panic(err)
	//}
	//defer resp.Body.Close()
	//var buffer [512]byte
	//result := bytes.NewBuffer(nil)
	//for {
	//	n, err := resp.Body.Read(buffer[0:])
	//	result.Write(buffer[0:n])
	//	if err != nil && err == io.EOF {
	//		break
	//	} else if err != nil {
	//		panic(err)
	//	}
	//}
	//return result.String()
}

// 校方session
type GDUPTSession string

func GetJSession() GDUPTSession {
	resp, err := http.Get("http://yq.gdupt.edu.cn/")

	defer resp.Body.Close()
	if err != nil {
		log.Println("[FAIL] 获取 JSESSIONID 错误")
	}

	homeCookie := resp.Header.Get("Set-Cookie")
	a := strings.Index(homeCookie, ";")
	log.Println("[INFO] 获取 JSESSIONID ", homeCookie[:a])
	return GDUPTSession(homeCookie[:a])
}

func (gs GDUPTSession) GDUPTAddForm(config *Config) {
	url0 := "http://yq.gdupt.edu.cn/syt/zzapply/operation.htm"
	method := "POST"
	pl := `data={"xmqkb":{"id":"ff8080817056f727017057083b010001"},"pdnf":"2020","type":"yqsjsb","c5":"36-37.2°C","c6":"健康","c7":"健康","c8":"否","c9":"","c2":"","c3":"","c10":"","c11":"","c12":"","c1":"否","c4":""}&msgUrl=syt/zzglappro/index.htm?type=yqsjsb&xmid=ff8080817056f727017057083b010001`
	// pl := `data={"xmqkb":{"id":"ff8080817056f727017057083b010001"},"pdnf":"2020","type":"yqsjsb","c5":"36-37.2°C","c6":"健康","c7":"健康","c8":"否","c9":"","c2":"` + config.City + `","c3":"` + config.Town + `","c10":"2020-03-01","c11":"` + config.ToSchool + `","c12":"否"}&msgUrl=syt/zzglappro/index.htm?type=yqsjsb&xmid=ff8080817056f727017057083b010001`
	escapeUrl := url.QueryEscape(pl)
	a := strings.Index(escapeUrl, "%3D")
	escapeUrl = escapeUrl[:a] + "=" + escapeUrl[a+3:]
	a = strings.Index(escapeUrl, "%3D")
	escapeUrl = escapeUrl[:a] + "=" + escapeUrl[a+3:]
	a = strings.Index(escapeUrl, "%26")
	escapeUrl = escapeUrl[:a] + "&" + escapeUrl[a+3:]
	payload := strings.NewReader(escapeUrl)

	client := &http.Client{
	}
	req, err := http.NewRequest(method, url0, payload)

	if err != nil {
		log.Println(err)
	}
	req.Header.Add("Accept", "text/plain, */*; q=0.01")
	req.Header.Add("Accept-Encoding", "gzip, deflate")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Cookie", "MyCssSkin=header-skin-1; menuVisible=0; "+string(gs))
	req.Header.Add("Host", "yq.gdupt.edu.cn")
	req.Header.Add("Origin", "http://yq.gdupt.edu.cn")
	req.Header.Add("Referer", "http://yq.gdupt.edu.cn/syt/zzapply/apply.htm?type=yqsjsb&judge=sq&xmid=ff8080817056f727017057083b010001&_t=562439&_winid=w8373")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.116 Safari/537.36")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")

	res, err := client.Do(req)
	if err != nil {
		log.Println("[FAIL] 提交失败")
		return
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("[FAIL] 提交失败")
		return
	}
	log.Println(string(body))
	if string(body) == "success" && res.Status == "200 OK" {
		log.Println("[INFO] 提交成功")
		config.SendWeChatMessage("已成功提交疫情报告", "请勿多次提交")
	} else if string(body) == "Applied today" {
		log.Println("[INFO] 不允许多次提交")
		config.SendWeChatMessage("多次提交疫情报告", "请勿多次提交")
	} else {
		log.Println("[INFO] 提交成功")
		config.SendWeChatMessage("已成功提交疫情报告", "请勿多次提交")
	}
}

func (gs GDUPTSession) GDUPTLogin(config *Config) bool {

	url := "http://yq.gdupt.edu.cn//login/Login.htm"

	var jsonStr = []byte("username=" + config.Username + "&password=" + config.Password)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))

	req.Header.Set("Accept", "text/html, */*; q=0.01")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cookie", "username="+config.Username+"; "+string(gs)+" username="+config.Password)
	req.Header.Set("Host", "yq.gdupt.edu.cn")
	req.Header.Set("Referer", "http://yq.gdupt.edu.cn/")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.116 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
		return false
	}
	defer resp.Body.Close()

	// log.Println("response Status:", resp.Status)
	// log.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	// log.Println("response Body:", string(body))
	if string(body) == "" && resp.Status == "200 OK" {
		log.Println("[INFO] 登录成功")
		return true
	} else {
		return false
	}
}

const configFileSizeLimit = 10 << 20

type Config struct {
	Username string `json:"username"`
	Password string `json:"password"`
	City     string `json:"city"`
	Town     string `json:"town"`
	ToSchool string `json:"toSchool"`
	FT       string `json:"fangTang"`
}

func GetConfig() *Config {
	config := LoadConfig("./config.json")
	return config
}

func LoadConfig(path string) *Config {
	var config Config
	config_file, err := os.Open(path)
	if err != nil {
		emit("Failed to open config file '%s': %s\n", path, err)
		return &config
	}

	fi, _ := config_file.Stat()
	if size := fi.Size(); size > (configFileSizeLimit) {
		emit("config file (%q) size exceeds reasonable limit (%d) - aborting", path, size)
		return &config // REVU: shouldn't this return an error, then?
	}

	if fi.Size() == 0 {
		emit("config file (%q) is empty, skipping", path)
		return &config
	}

	buffer := make([]byte, fi.Size())
	_, err = config_file.Read(buffer)

	buffer, err = StripComments(buffer)
	if err != nil {
		emit("Failed to strip comments from json: %s\n", err)
		return &config
	}

	buffer = []byte(os.ExpandEnv(string(buffer)))

	err = json.Unmarshal(buffer, &config)
	if err != nil {
		emit("Failed unmarshalling json: %s\n", err)
		return &config
	}
	return &config
}

func StripComments(data []byte) ([]byte, error) {
	data = bytes.Replace(data, []byte("\r"), []byte(""), 0)
	lines := bytes.Split(data, []byte("\n"))
	filtered := make([][]byte, 0)

	for _, line := range lines {
		match, err := regexp.Match(`^\s*#`, line)
		if err != nil {
			return nil, err
		}
		if !match {
			filtered = append(filtered, line)
		}
	}

	return bytes.Join(filtered, []byte("\n")), nil
}

func emit(msgfmt string, args ...interface{}) {
	log.Printf(msgfmt, args...)
}
func ResultConfig(test []map[string]interface{}) (port_password []map[string]interface{}) {
	return
}
