package main

import (
	. "./TiebaSign"
	"bufio"
	"container/list"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

func getCookie(cookieFileName string, silence bool) (cookieJar *cookiejar.Jar, hasError bool) {
	needLogin := true
	cookieJar, _ = cookiejar.New(nil)
	cookies := make([]*http.Cookie, 0)
	if _, err := os.Stat(cookieFileName); err == nil {
		rawCookie, _ := ioutil.ReadFile(cookieFileName)
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
		fmt.Printf("Verifying imported cookies from %s...", cookieFileName)
		url, _ := url.Parse("http://baidu.com")
		cookieJar.SetCookies(url, cookies)
		if GetLoginStatus(cookieJar) {
			needLogin = false
			fmt.Println("OK")
		} else {
			fmt.Println("Failed")
		}
	}
	if needLogin && !silence {
		var username, password string
		bufferedReader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter your Baidu ID: ")
		usernameByte, _, _ := bufferedReader.ReadLine()
		username = string(usernameByte)
		if username == "" {
			return nil, true
		}
		fmt.Print("Enter your Baidu Password: ")
		fmt.Scan(&password)
		if password == "" {
			return nil, true
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
			fmt.Println("Error while login:")
			fmt.Println(loginErr)
			time.Sleep(5e9)
			return nil, true
		}
	} else if silence {
		return nil, true
	}
	return cookieJar, false
}

type SignTask struct {
	cookie         *cookiejar.Jar
	tieba          LikedTieba
	failedAttempts int
}

func main() {
	var batchMode = flag.Bool("batch", false, "Batch Sign mode")
	var maxRetryTimes = flag.Int("retry", 7, "Max retry times for a single tieba")
	var cookieFileName = flag.String("cookie", "cookie.txt", "Try to load cookie from specified file")
	flag.Parse()

	fmt.Println("Tieba Sign (Go Version) beta")
	fmt.Println("Author: kookxiang <r18@ikk.me>")
	fmt.Println()
	if *batchMode == false {
		cookie, hasError := getCookie(*cookieFileName, *cookieFileName != "cookie.txt")
		if hasError {
			return
		}
		fmt.Print("Fetching tieba list...")
		likedTiebaList, err := GetLikedTiebaList(cookie)
		if err != nil {
			fmt.Println("Error!")
			fmt.Println(err)
			return
		} else {
			fmt.Println("OK")
		}
		taskList := list.New()
		for _, tieba := range likedTiebaList {
			taskList.PushBack(SignTask{
				tieba:          tieba,
				cookie:         cookie,
				failedAttempts: 0,
			})
		}

		for {
			taskNode := taskList.Front()
			if taskNode == nil {
				break
			}
			taskList.Remove(taskNode)
			task := taskNode.Value.(SignTask)
			fmt.Printf("Now signing %s...", ToUtf8(task.tieba.Name))
			status, _, exp := TiebaSign(task.tieba, task.cookie)
			if status == 2 {
				if exp > 0 {
					fmt.Printf("Succeed! Exp +%d\n", exp)
				} else {
					fmt.Println("Succeed")
				}
			} else if status == 1 {
				task.failedAttempts++
				if task.failedAttempts <= *maxRetryTimes {
					taskList.PushBack(task) // push failed task back to list
					fmt.Println("Failed, retry later")
				} else {
					fmt.Println("Failed")
				}
			} else {
				fmt.Println("Failed")
			}
			if exp > 0 || status == 1 {
				time.Sleep(2e9)
			}
		}
	} else {
		fmt.Println("Loading and verifying Cookies from ./cookies/")
		cookieFiles, _ := ioutil.ReadDir("cookies")
		threadList := sync.WaitGroup{}
		cookieList := map[string]*cookiejar.Jar{}
		for _, file := range cookieFiles {
			profileName := strings.Replace(file.Name(), ".txt", "", 1)
			cookie, hasError := getCookie("cookies/"+file.Name(), true)
			if hasError {
				fmt.Errorf("Failed to load profile %s, invalid cookie!\n", profileName)
			} else {
				cookieList[profileName] = cookie
			}
		}
		for profileName, cookie := range cookieList {
			threadList.Add(1)
			go func(profileName string, cookie *cookiejar.Jar) {
				fmt.Printf("[%s] Go routine started.\n", profileName)
				likedTiebaList, err := GetLikedTiebaList(cookie)
				if err != nil {
					fmt.Printf("[%s] Error while fetching tieba list\n", profileName)
					threadList.Done()
					return
				} else {
					fmt.Printf("[%s] Loaded tieba list.\n", profileName)
				}
				taskList := list.New()
				for _, tieba := range likedTiebaList {
					taskList.PushBack(SignTask{
						tieba:          tieba,
						cookie:         cookie,
						failedAttempts: 0,
					})
				}
				for {
					taskNode := taskList.Front()
					if taskNode == nil {
						break
					}
					taskList.Remove(taskNode)
					task := taskNode.Value.(SignTask)
					status, _, exp := TiebaSign(task.tieba, task.cookie)
					if status == 2 {
						if exp > 0 {
							fmt.Printf("[%s] Succeed: %s, Exp +%d\n", profileName, ToUtf8(task.tieba.Name), exp)
						} else {
							fmt.Printf("[%s] Succeed: %s\n", profileName, ToUtf8(task.tieba.Name))
						}
					} else if status == 1 {
						fmt.Printf("[%s] Failed:  %s\n", profileName, ToUtf8(task.tieba.Name))
						task.failedAttempts++
						if task.failedAttempts <= *maxRetryTimes {
							taskList.PushBack(task) // push failed task back to list
						}
					} else {
						fmt.Printf("[%s] Failed:  %s\n", profileName, ToUtf8(task.tieba.Name))
					}
					if exp > 0 || status == 1 {
						time.Sleep(2e9)
					}
				}
				threadList.Done()
			}(profileName, cookie)
		}
		threadList.Wait()
		fmt.Println("All Task Finished! Congratulation!")
	}
}
