/*
用websocke 来实现进度条功能
*/
package model

import (
	"fmt"
	//	"log"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

//type Websocket struct {
//	ProgressChan chan float32
//}

//多用户？？？
type ProgressBar struct {
	CheckProgress   chan float32
	CompareProgress chan float32
	CheckValue      float32
	CompareValue    float32
}

var ProgressMap map[string]*ProgressBar = make(map[string]*ProgressBar, 0) //使用时加锁！

func Progress(ws *websocket.Conn) {
	defer ws.Close()
	//	defer delete(ProgressMap, ws.Request().URL.Query().Get("userId"))
	defer deleteProgressMap(ws.Request().URL.Query().Get("userId"))
	var userId, task string
	userId = ws.Request().URL.Query().Get("userId")
	task = ws.Request().URL.Query().Get("task")

	for {
		if _, ok := ProgressMap[userId]; ok {
			for {
				var value float32
				if task == "参数核查" {
					<-ProgressMap[userId].CheckProgress
					value = ProgressMap[userId].CheckValue
				} else if task == "参数对比" {
					<-ProgressMap[userId].CompareProgress
					value = ProgressMap[userId].CompareValue
				}
				err := websocket.Message.Send(ws, fmt.Sprintf("%.2f", value*100))
				if err != nil || 1 == value {
					return
				}
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func deleteProgressMap(userId string) {
	if progress, ok := ProgressMap[userId]; ok {
		if (float32(0) <= progress.CheckValue && progress.CheckValue < float32(1)) || (float32(0) <= progress.CompareValue && progress.CompareValue < float32(1)) {
			return
		}
		delete(ProgressMap, userId)
	}
}

func RunWebsocket() {
	http.Handle("/socket", websocket.Handler(Progress))
	http.ListenAndServe(":9000", nil)
}
