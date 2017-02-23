package TiebaSign

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	. "github.com/bitly/go-simplejson"
	"net/http/cookiejar"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

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
	for key := range postData {
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
		return 1, fetchErr.Error(), 0
	}
	json, parseErr := NewJson([]byte(body))
	if parseErr != nil {
		return 1, parseErr.Error(), 0
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
