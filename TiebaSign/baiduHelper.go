package TiebaSign

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	. "github.com/bitly/go-simplejson"
	"io/ioutil"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func GetBaiduID() string {
	baiduID := getBaiduID()
	if baiduID == "" {
		Fetch("https://passport.baidu.com/v2/", nil)
		baiduID = getBaiduID()
	}
	return baiduID
}

func getBaiduID() string {
	return GetCookie("BAIDUID")
}

func GetLoginToken() (string, error) {
	GetBaiduID()
	body, fetchErr := Fetch("https://passport.baidu.com/v2/api/?getapi&tpl=tb&apiver=v3&tt="+GetTimestampStr()+"520&class=login&logintype=dialogLogin", nil)
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

func BaiduLogin(username, password string) (result int, err error) {
	loginToken, tokenError := GetLoginToken()
	if tokenError != nil {
		return 0, tokenError
	}
	return BaiduLoginWithCaptcha(username, password, "", "", loginToken)
}

func BaiduLoginWithCaptcha(username, password, codeString, verifyCode, loginToken string) (result int, err error) {
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

	body, fetchErr := Fetch("https://passport.baidu.com/v2/api/?login", postData)
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
		verifyImage, _ := Fetch("https://passport.baidu.com/cgi-bin/genimage?"+codeString, nil)
		ioutil.WriteFile("captcha.jpg", []byte(verifyImage), 0644)
		fmt.Print("Now enter the captcha: ")
		fmt.Scan(&verifyCode)
		return BaiduLoginWithCaptcha(username, password, codeString, verifyCode, loginToken)
	}
	if errNo != "" && errNo != "err_no=0" {
		fmt.Println("Unknown error. Error number:", errNo)
		return -3, nil
	}

	return 1, nil
}

func GetLikedTiebaList() ([]LikedTieba, error) {
	pn := 0
	likedTiebaList := make([]LikedTieba, 0)
	for {
		pn++
		url := "http://tieba.baidu.com/f/like/mylike?pn=" + fmt.Sprintf("%d", pn)
		body, fetchErr := Fetch(url, nil)
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

func getTbs() string {
	body, err := Fetch("http://tieba.baidu.com/dc/common/tbs", nil)
	if err != nil {
		return ""
	}
	json, parseErr := NewJson([]byte(body))
	if parseErr != nil {
		return ""
	}
	return json.Get("tbs").MustString()
}

func TiebaSign(tieba LikedTieba) (int, string, int) {
	postData := make(map[string]string)
	postData["BDUSS"] = GetCookie("BDUSS")
	postData["_client_id"] = "03-00-DA-59-05-00-72-96-06-00-01-00-04-00-4C-43-01-00-34-F4-02-00-BC-25-09-00-4E-36"
	postData["_client_type"] = "4"
	postData["_client_version"] = "1.2.1.17"
	postData["_phone_imei"] = "540b43b59d21b7a4824e1fd31b08e9a6"
	postData["fid"] = fmt.Sprintf("%d", tieba.TiebaId)
	postData["kw"] = tieba.Name
	postData["net_type"] = "3"
	postData["tbs"] = getTbs()

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

	body, fetchErr := Fetch("http://c.tieba.baidu.com/c/c/forum/sign", postData)
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
