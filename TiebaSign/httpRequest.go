package TiebaSign

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

var cookies []*http.Cookie
var cookieJar, _ = cookiejar.New(nil)

func Fetch(targetUrl string, postData map[string]string) (string, error) {
	var request *http.Request
	httpClient := &http.Client{
		Jar: cookieJar,
	}
	if nil == postData {
		request, _ = http.NewRequest("GET", targetUrl, nil)
	} else {
		postParams := url.Values{}
		for key, value := range postData {
			postParams.Set(key, value)
		}
		postDataStr := postParams.Encode()
		postDataBytes := []byte(postDataStr)
		postBytesReader := bytes.NewReader(postDataBytes)
		request, _ = http.NewRequest("POST", targetUrl, postBytesReader)
		request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	response, fetchError := httpClient.Do(request)
	if fetchError != nil {
		return "", fetchError
	}
	defer response.Body.Close()
	body, readError := ioutil.ReadAll(response.Body)
	if readError != nil {
		return "", readError
	}
	cookies = cookieJar.Cookies(request.URL)
	return string(body), nil
}

func GetCookie() []*http.Cookie {
	return cookies
}
