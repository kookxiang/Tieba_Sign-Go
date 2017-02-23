package conf

import (
	"net/http"
	"fmt"
	"net/url"
	"net/http/cookiejar"
	"os"
	"io/ioutil"
	"bytes"
	"strings"
	"github.com/Evi1/Tieba_Sign-Go/TiebaSign"
)

func getCookies(cookieFileName string) (cookieJar *cookiejar.Jar, hasError bool) {
	hasError = true
	cookieJar, _ = cookiejar.New(nil)
	cookies := make([]*http.Cookie, 0)
	if _, err := os.Stat(cookieFileName); err == nil {
		rawCookie, _ := ioutil.ReadFile(cookieFileName)
		rawCookie = bytes.Trim(rawCookie, "\xef\xbb\xbf")
		rawCookieList := strings.Split(strings.Replace(string(rawCookie), "\r\n", "\n", -1), "\n")
		for _, rawCookieLine := range rawCookieList {
			rawCookieInfo := strings.SplitN(rawCookieLine, "=", 2)
			if len(rawCookieInfo) < 2 {
				continue
			}
			cookies = append(cookies, &http.Cookie{
				Name:   rawCookieInfo[0],
				Value:  rawCookieInfo[1],
				Domain: ".baidu.com",
			})
		}
		fmt.Printf("Verifying imported cookies from %s...", cookieFileName)
		URL, _ := url.Parse("http://baidu.com")
		cookieJar.SetCookies(URL, cookies)
		if TiebaSign.GetLoginStatus(cookieJar) {
			hasError = false
			fmt.Println("OK")
		} else {
			fmt.Println("Failed")
		}
	}
	if hasError {
		return nil, true
	}
	hasError = false
	return
}

func StartCookiesWork(cookieList map[string]*cookiejar.Jar, errorList map[string]bool) {
	fmt.Println("Loading and verifying Cookies from ./cookies/")
	cookieFiles, _ := ioutil.ReadDir("cookies")
	//cookieList = make(map[string]*cookiejar.Jar)
	//errorList = make(map[string]bool)
	for k := range cookieList {
		delete(cookieList, k)
	}
	for k := range errorList {
		delete(errorList, k)
	}
	for _, file := range cookieFiles {
		profileName := strings.Replace(file.Name(), ".txt", "", 1)
		cookie, hasError := getCookies("cookies/" + file.Name())
		if hasError {
			fmt.Errorf("Failed to load profile %s, invalid cookie!\n", profileName)
			errorList[profileName] = true
		} else {
			cookieList[profileName] = cookie
			errorList[profileName] = false
		}
	}
}
