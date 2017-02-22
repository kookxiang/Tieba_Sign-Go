package TiebaSign

import (
	"net/http/cookiejar"
	"fmt"
	"time"
	"container/list"
	"sync"
)

type SignTask struct {
	cookie         *cookiejar.Jar
	tieba          LikedTieba
	failedAttempts int
}

func StartSign(cookieList map[string]*cookiejar.Jar, runList map[string]map[string]string, maxRetryTimes int) {
	threadList := sync.WaitGroup{}
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
			_, ok := runList[profileName]
			if !ok {
				runList[profileName] = make(map[string]string)
			}
			for _, tieba := range likedTiebaList {
				_, ok = runList[profileName][ToUtf8(tieba.Name)]
				if !ok {
					runList[profileName][ToUtf8(tieba.Name)] = "none"
				}
				taskList.PushBack(SignTask{
					tieba:          tieba,
					cookie:         cookie,
					failedAttempts: 0,
				})
			}
			fmt.Println(runList)
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
					runList[profileName][ToUtf8(task.tieba.Name)] = s
				} else if status == 1 {
					fmt.Printf(s+" [%s] Failed1:  %s\n", profileName, ToUtf8(task.tieba.Name))
					task.failedAttempts++
					if task.failedAttempts <= maxRetryTimes {
						taskList.PushBack(task) // push failed task back to list
					}
					time.Sleep(2e9)
					if runList[profileName][ToUtf8(task.tieba.Name)] == "none" {
						runList[profileName][ToUtf8(task.tieba.Name)] = "Failed"
					}
				} else {
					if runList[profileName][ToUtf8(task.tieba.Name)] == "none" {
						runList[profileName][ToUtf8(task.tieba.Name)] = "Failed"
					}
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
