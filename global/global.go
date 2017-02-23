package global

import "net/http/cookiejar"

var CookieList map[string]*cookiejar.Jar
var ErrorList map[string]bool

var RunList map[string]map[string]string

func init() {
	CookieList = make(map[string]*cookiejar.Jar)
	ErrorList = make(map[string]bool)
	RunList = make(map[string]map[string]string)
}
