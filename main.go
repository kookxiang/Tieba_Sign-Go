package main

import (
	. "./TiebaSign"
	"bufio"
	"container/list"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"
)

func main() {
	needLogin := true
	cookieJar, _ := cookiejar.New(nil)
	cookies := make([]*http.Cookie, 0)
	if _, err := os.Stat("cookie.txt"); err == nil {
		rawCookie, _ := ioutil.ReadFile("cookie.txt")
		rawCookieList := strings.Split(string(rawCookie), "\n")
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
		url, _ := url.Parse("http://baidu.com")
		cookieJar.SetCookies(url, cookies)
		if GetLoginStatus(cookieJar) {
			needLogin = false
		}
	}
	if needLogin {
		var username, password string
		bufferedReader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter your Baidu ID: ")
		usernameByte, _, _ := bufferedReader.ReadLine()
		username = string(usernameByte)
		if username == "" {
			return
		}
		fmt.Print("Enter your Baidu Password: ")
		fmt.Scan(&password)
		if password == "" {
			return
		}
		result, loginErr := BaiduLogin(username, password, cookieJar)
		if loginErr == nil && result > 0 {
			fmt.Println("Successfully login")
			cookieStr := ""
			for _, cookie := range GetCookies(cookieJar) {
				cookieStr += cookie.Name + "=" + cookie.Value + "\n"
			}
			ioutil.WriteFile("cookie.txt", []byte(cookieStr), 0644)

			fmt.Println("Your cookie has been written into cookie.txt")
		} else {
			time.Sleep(5e9)
			return
		}
	}

	// Start sign
	likedTiebaList, err := GetLikedTiebaList(cookieJar)
	if err != nil {
		fmt.Println(err)
		return
	}
	linkedList := list.New() // Create sign list
	for _, tieba := range likedTiebaList {
		linkedList.PushBack(tieba)
	}
	failedAttempts := make(map[int]int)
	for {
		listItem := linkedList.Front()
		if listItem == nil {
			break
		}
		linkedList.Remove(listItem)
		tieba := listItem.Value.(LikedTieba)
		status, message, exp := TiebaSign(tieba, cookieJar)
		fmt.Printf("%s\t%d: %s\tEXP+%d\n", ToUtf8(tieba.Name), status, message, exp)
		if exp > 0 || status == 1 {
			time.Sleep(2e9)
		}
		if status == 1 {
			failedAttempts[tieba.TiebaId]++
			if failedAttempts[tieba.TiebaId] <= 15 {
				linkedList.PushBack(tieba) // push failed items back to list
			}
		}
	}
	time.Sleep(3e9)
}
