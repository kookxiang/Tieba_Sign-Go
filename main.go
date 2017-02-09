package main

import (
	. "github.com/Evi1/Tieba_Sign-Go/TiebaSign"
	"bytes"
	"container/list"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
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
		if GetLoginStatus(cookieJar) {
			needLogin = false
			fmt.Println("OK")
		} else {
			fmt.Println("Failed")
		}
	}
	if needLogin && !silence {
		fmt.Println("Cannot login, since baidu has switched to RSA login, and I have no time to make a new login program. You have to make a cookie.txt, and paste your cookie in the file.")
		fmt.Println("Cookie Format: BDUSS=xxxxxxx")
		return nil, true
	} else if needLogin && silence {
		return nil, true
	}
	hasError=false
	return
}

type SignTask struct {
	cookie         *cookiejar.Jar
	tieba          LikedTieba
	failedAttempts int
}

func main() {
	var batchMode = flag.Bool("batch", true, "Batch Sign mode")
	var maxRetryTimes = flag.Int("retry", 7, "Max retry times for a single tieba")
	var cookieFileName = flag.String("cookie", "cookie.txt", "Try to load cookie from specified file")
	flag.Parse()

	currentDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	os.Chdir(currentDir)

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
					fmt.Printf("[%s] Go routine stopped.\n", profileName)
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
					status, s, exp := TiebaSign(task.tieba, task.cookie)
					if status == 2 {
						if exp > 0 {
							fmt.Printf(s+" [%s] Succeed: %s, Exp +%d\n", profileName, ToUtf8(task.tieba.Name), exp)
						} else {
							fmt.Printf(s+" [%s] Succeed: %s\n", profileName, ToUtf8(task.tieba.Name))
						}
					} else if status == 1 {
						fmt.Printf(s+" [%s] Failed1:  %s\n", profileName, ToUtf8(task.tieba.Name))
						task.failedAttempts++
						if task.failedAttempts <= *maxRetryTimes {
							taskList.PushBack(task) // push failed task back to list
						}
						time.Sleep(2e9)
					} else {
						fmt.Printf(s+" [%s] Failed2:  %s\n", profileName, ToUtf8(task.tieba.Name))
					}
				}
				fmt.Printf("[%s] Finished!\n", profileName)
				fmt.Printf("[%s] Go routine stopped.\n", profileName)
				threadList.Done()
			}(profileName, cookie)
		}
		threadList.Wait()
		fmt.Println("All Task Finished! Congratulation!")
	}
}
