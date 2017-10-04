package TiebaSign

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	. "github.com/bitly/go-simplejson"
	"io/ioutil"
	"net/http/cookiejar"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

//get baidu username
func GetBaiduIdName(ptrCookieJar *cookiejar.Jar) string {

	url := "http://wapp.baidu.com/"
	body, fetchErr := Fetch(url, nil, ptrCookieJar)
	if fetchErr != nil {
		fmt.Println(fetchErr)
		return ""
	}
	exp := regexp.MustCompile("<a href=.+?>(.*?)的i贴吧</a>")
	Exp := exp.FindStringSubmatch(string(body))[1]
	return Exp
}

//get baidu uid
func getBaiduUID(ptrCookieJar *cookiejar.Jar, username string) string {
	url := "http://tieba.baidu.com/home/get/panel?ie=utf-8&un=" + username
	body, fetchErr := Fetch(url, nil, ptrCookieJar)
	if fetchErr != nil {
		fmt.Println(fetchErr)
		return ""
	}
	data := map[string]interface{}{}
	json.Unmarshal([]byte(string(body)), &data)
	uid := strconv.FormatFloat(data["data"].(map[string]interface{})["id"].(float64), 'f', -1, 64)
	return uid
}

func GetBaiduID(ptrCookieJar *cookiejar.Jar) string {
	baiduID := getBaiduID(ptrCookieJar)
	if baiduID == "" {
		Fetch("https://passport.baidu.com/v2/", nil, ptrCookieJar)
		baiduID = getBaiduID(ptrCookieJar)
	}
	return baiduID
}

func getBaiduID(ptrCookieJar *cookiejar.Jar) string {
	return GetCookie(ptrCookieJar, "BAIDUID")
}

func GetLoginToken(ptrCookieJar *cookiejar.Jar) (string, error) {
	GetBaiduID(ptrCookieJar)
	body, fetchErr := Fetch("https://passport.baidu.com/v2/api/?getapi&tpl=tb&apiver=v3&tt="+GetTimestampStr()+"520&class=login&logintype=dialogLogin", nil, ptrCookieJar)
	if fetchErr != nil {
		return "", fetchErr
	}
	body = strings.Replace(body, "'", "\"", -1)
	json, parseErr := NewJson([]byte(body))
	if parseErr != nil {
		return "", parseErr
	}
	token, accessError := json.Get("data").Get("token").String()
	if accessError != nil {
		return "", accessError
	}
	return token, nil
}

func BaiduLogin(username, password string, ptrCookieJar *cookiejar.Jar) (result int, err error) {
	loginToken, tokenError := GetLoginToken(ptrCookieJar)
	if tokenError != nil {
		return 0, tokenError
	}
	return BaiduLoginWithCaptcha(username, password, ptrCookieJar, "", "", loginToken)
}

func BaiduLoginWithCaptcha(username, password string, ptrCookieJar *cookiejar.Jar, codeString, verifyCode, loginToken string) (result int, err error) {
	postData := make(map[string]string)
	postData["apiver"] = "v3"
	postData["charset"] = "UTF-8"
	postData["codestring"] = codeString
	postData["isPhone"] = "false"
	postData["logintype"] = "bascilogin"
	postData["mem_pass"] = "on"
	postData["password"] = password
	postData["ppui_logintime"] = "8888"
	postData["quick_user"] = "0"
	postData["safeflg"] = "0"
	postData["splogin"] = "rate"
	postData["staticpage"] = "http://tieba.baidu.com/tb/static-common/html/pass/v3Jump.html"
	postData["token"] = loginToken
	postData["tpl"] = "tb"
	postData["tt"] = GetTimestampStr() + "520"
	postData["u"] = "http://tieba.baidu.com/"
	postData["username"] = username
	postData["verifycode"] = verifyCode

	body, fetchErr := Fetch("https://passport.baidu.com/v2/api/?login", postData, ptrCookieJar)
	if fetchErr != nil {
		return 0, fetchErr
	}

	errNo := regexp.MustCompile("err_no=(\\d+)").FindString(body)
	if errNo == "err_no=400031" {
		fmt.Println("Login-protect was on, please turn it off as passport.baidu.com")
		return -1, nil // 登陆保护
	}
	if errNo == "err_no=4" {
		fmt.Println("Wrong username or password")
		return -2, nil // 用户名 / 密码有误
	}
	if matched, _ := regexp.Match("captchaservice", []byte(body)); matched {
		reg, _ := regexp.Compile("(captchaservice\\w{200,})")
		fmt.Println("Server denied logging request and sent a captcha.")
		codeString = reg.FindString(body)
		fmt.Println("Please open captcha image manually: captcha.jpg")
		verifyImage, _ := Fetch("https://passport.baidu.com/cgi-bin/genimage?"+codeString, nil, ptrCookieJar)
		ioutil.WriteFile("captcha.jpg", []byte(verifyImage), 0644)
		fmt.Print("Now enter the captcha: ")
		fmt.Scan(&verifyCode)
		return BaiduLoginWithCaptcha(username, password, ptrCookieJar, codeString, verifyCode, loginToken)
	}
	if errNo != "" && errNo != "err_no=0" {
		fmt.Println("Unknown error. Error number:", errNo)
		return -3, nil
	}

	return 1, nil
}

func GetLikedTiebaListNew(ptrCookieJar *cookiejar.Jar) ([]LikedTieba, error) {
	pn := 0
	likedTiebaList := make([]LikedTieba, 0)
	username := GetBaiduIdName(ptrCookieJar)
	for {
		pn++
		pnstr := fmt.Sprintf("%d", pn)
		t := time.Now()
		postData := make(map[string]string)
		postData["page_no"] = pnstr
		postData["_client_id"] = "wappc_" + string(t.Unix()) + "233"
		postData["_client_type"] = "2"
		postData["_client_version"] = "6.5.8"
		postData["_phone_imei"] = "357143042411618"
		postData["from"] = "baidu_appstore"
		postData["is_guest"] = "1"
		postData["model"] = "H60-L01"
		postData["page_size"] = "200"
		postData["timestamp"] = string(t.Unix()) + "903"
		postData["uid"] = getBaiduUID(ptrCookieJar, username)
		var keys []string
		for key, _ := range postData {
			keys = append(keys, key)
		}
		sort.Sort(sort.StringSlice(keys))

		sign_str := ""
		for _, key := range keys {
			sign_str += fmt.Sprintf("%s=%s", key, postData[key])
		}
		sign_str += "tiebaclient!!!"

		MD5 := md5.New()
		MD5.Write([]byte(sign_str))
		MD5Result := MD5.Sum(nil)
		signValue := make([]byte, 32)
		hex.Encode(signValue, MD5Result)
		postData["sign"] = strings.ToUpper(string(signValue))

		var url = "http://c.tieba.baidu.com/c/f/forum/like"

		body, fetchErr := Fetch(url, postData, ptrCookieJar)
		if fetchErr != nil {
			return nil, fetchErr
		}
		data := map[string]interface{}{}
		json.Unmarshal([]byte(string(body)), &data)

		typeit := fmt.Sprintf("%T", data["forum_list"])
		if typeit == "[]interface {}" || data["forum_list"] == nil {
			break
		}

		list, err := ParseLikedTiebaNew(data["forum_list"].(map[string]interface{}))
		likedTiebaList = append(likedTiebaList, list...)
		if err != nil {
			fmt.Println(err)
		}

	}
	return likedTiebaList, nil
}

func GetLikedTiebaList(ptrCookieJar *cookiejar.Jar) ([]LikedTieba, error) {
	pn := 0
	likedTiebaList := make([]LikedTieba, 0)
	for {
		pn++
		url := "http://tieba.baidu.com/f/like/mylike?pn=" + fmt.Sprintf("%d", pn)
		body, fetchErr := Fetch(url, nil, ptrCookieJar)
		if fetchErr != nil {
			return nil, fetchErr
		}
		reg := regexp.MustCompile("<tr><td>.+?</tr>")
		allTr := reg.FindAllString(body, -1)
		for _, line := range allTr {
			likedTieba, err := ParseLikedTieba(line)
			if err != nil {
				continue
			}
			likedTiebaList = append(likedTiebaList, likedTieba)
		}
		if allTr == nil {
			break
		}
	}
	return likedTiebaList, nil
}

func getTbs(ptrCookieJar *cookiejar.Jar) string {
	body, err := Fetch("http://tieba.baidu.com/dc/common/tbs", nil, ptrCookieJar)
	if err != nil {
		return ""
	}
	json, parseErr := NewJson([]byte(body))
	if parseErr != nil {
		return ""
	}
	return json.Get("tbs").MustString()
}

func GetLoginStatus(ptrCookieJar *cookiejar.Jar) bool {
	body, err := Fetch("http://tieba.baidu.com/dc/common/tbs", nil, ptrCookieJar)
	if err != nil {
		return false
	}
	json, parseErr := NewJson([]byte(body))
	if parseErr != nil {
		return false
	}
	return json.Get("is_login").MustInt() == 1
}

func TiebaSign(tieba LikedTieba, ptrCookieJar *cookiejar.Jar) (int, string, int) {
	postData := make(map[string]string)
	postData["BDUSS"] = GetCookie(ptrCookieJar, "BDUSS")
	postData["_client_id"] = "03-00-DA-59-05-00-72-96-06-00-01-00-04-00-4C-43-01-00-34-F4-02-00-BC-25-09-00-4E-36"
	postData["_client_type"] = "4"
	postData["_client_version"] = "1.2.1.17"
	postData["_phone_imei"] = "540b43b59d21b7a4824e1fd31b08e9a6"
	postData["fid"] = fmt.Sprintf("%d", tieba.TiebaId)
	postData["kw"] = tieba.Name
	postData["net_type"] = "3"
	postData["tbs"] = getTbs(ptrCookieJar)

	var keys []string
	for key, _ := range postData {
		keys = append(keys, key)
	}
	sort.Sort(sort.StringSlice(keys))

	sign_str := ""
	for _, key := range keys {
		sign_str += fmt.Sprintf("%s=%s", key, postData[key])
	}
	sign_str += "tiebaclient!!!"

	MD5 := md5.New()
	MD5.Write([]byte(sign_str))
	MD5Result := MD5.Sum(nil)
	signValue := make([]byte, 32)
	hex.Encode(signValue, MD5Result)
	postData["sign"] = strings.ToUpper(string(signValue))

	body, fetchErr := Fetch("http://c.tieba.baidu.com/c/c/forum/sign", postData, ptrCookieJar)
	if fetchErr != nil {
		return -1, fetchErr.Error(), 0
	}
	json, parseErr := NewJson([]byte(body))
	if parseErr != nil {
		return -1, parseErr.Error(), 0
	}
	if _exp, succeed := json.Get("user_info").CheckGet("sign_bonus_point"); succeed {
		exp, _ := strconv.Atoi(_exp.MustString())
		return 2, fmt.Sprintf("签到成功，获得经验值 %d", exp), exp
	}
	switch json.Get("error_code").MustString() {
	case "340010":
		fallthrough
	case "160002":
		fallthrough
	case "3":
		return 2, "你已经签到过了", 0
	case "1":
		fallthrough
	case "340008": // 黑名单
		fallthrough
	case "340006": // 被封啦
		fallthrough
	case "160004":
		return -1, fmt.Sprintf("ERROR-%s: %s", json.Get("error_code").MustString(), json.Get("error_msg").MustString()), 0
	case "160003":
		fallthrough
	case "160008":
		fallthrough
	default:
		return 1, fmt.Sprintf("ERROR-%s: %s", json.Get("error_code").MustString(), json.Get("error_msg").MustString()), 0
	}
	return -255, "", 0
}
