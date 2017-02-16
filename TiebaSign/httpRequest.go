package TiebaSign

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

func Fetch(targetUrl string, postData map[string]string, ptrCookieJar *cookiejar.Jar) (string, error) {
	var request *http.Request
	httpClient := &http.Client{
		Jar: ptrCookieJar,
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
	return string(body), nil
}

func GetCookie(cookieJar *cookiejar.Jar, name string) string {
	cookieUrl, _ := url.Parse("http://tieba.baidu.com")
	cookies := cookieJar.Cookies(cookieUrl)
	for _, cookie := range cookies {
		if name == cookie.Name {
			return cookie.Value
		}
	}
	return ""
}

