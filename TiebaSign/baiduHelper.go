package TiebaSign

import (
	"fmt"
	. "github.com/bitly/go-simplejson"
	"io/ioutil"
	"net/url"
	"regexp"
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
	cookies := GetCookie()
	for _, cookie := range cookies {
		if cookie.Name == "BAIDUID" {
			// Already has BAIDUID cookie
			return cookie.Value
		}
	}
	return ""
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
	postData["password"] = url.QueryEscape(password)
	postData["ppui_logintime"] = "8888"
	postData["quick_user"] = "0"
	postData["safeflg"] = "0"
	postData["splogin"] = "rate"
	postData["staticpage"] = "http://tieba.baidu.com/tb/static-common/html/pass/v3Jump.html"
	postData["token"] = loginToken
	postData["tpl"] = "tb"
	postData["tt"] = GetTimestampStr() + "520"
	postData["u"] = "http://tieba.baidu.com/"
	postData["username"] = url.QueryEscape(username)
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
	if errNo != "" && errNo != "err_no=0" {
		fmt.Println("Unknown error. Error number:", errNo)
		return -3, nil
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

	return 1, nil
}

func GetLikedTiebaList() (map[int]string, error) {
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
		break
	}
	return nil, nil
}
