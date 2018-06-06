//web 下载服务器文件
package main

import (
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

//type User struct {
//	UserId   string `json:"userId"`
//	FileName string `json:"fileName"`
//	Dir      string `json:"dir"`
//}

var (
	FilesPath string = ""
)

func loadFilesPath() {
	if strings.Contains(runtime.GOOS, "linux") {
		//linux 发布模式
		FilesPath += `/home/parameterCheckww/files/`
		return
	}

	//windows 调试模式
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal("file globalValue.go ,func model.LoadFilesPath , err :", err, " ,os.Getwd failed")
		return
	}
	dirs := strings.Split(pwd, string(filepath.Separator))

	for _, dir := range dirs {
		if dir == "downloadFileServer" {
			break
		}
		FilesPath += (dir + string(filepath.Separator))
	}
	FilesPath += ("files" + string(filepath.Separator))

}

func handler(w http.ResponseWriter, r *http.Request) {

	/*GET url传参*/
	//	if r.Method != "GET" {
	//		return
	//	}
	//	userId, fileName, dir := r.URL.Query().Get("userId"), r.URL.Query().Get("fileName"), r.URL.Query().Get("dir")

	/*POST/PUT     body传参*/
	//	bodyData, err := ioutil.ReadAll(body)
	//	if nil != err {
	//		return err
	//	}
	//	var user = new(User)
	//	err = json.Unmarshal(bodyData, user)
	//	if nil != err {
	//		log.Error("model: ParseJsonFromUrlBody, json.Unmarshal failed, body:", string(bodyData))
	//		log.Error("model: ParseJsonFromUrlBody, json.Unmarshal failed, err:", err)
	//		return err
	//	}
	//	var userId, fileName, dir string
	//	userId, fileName, dir = user.UserId, user.FileName, user.Dir

	/*POST  form传参*/
	if r.Method != "POST" {
		return
	}
	userId, fileName, dir := r.FormValue("userId"), r.FormValue("fileName"), r.FormValue("dir")
	if userId == "" || fileName == "" || dir == "" {
		w.Write([]byte("参数不合法！"))
		return
	}

	fileFullPath := FilesPath + userId + string(filepath.Separator) + `templet` + string(filepath.Separator) + fileName
	if dir == "公共" {
		fileFullPath = FilesPath + `00000000` + string(filepath.Separator) + `templet` + string(filepath.Separator) + fileName
	}

	file, err := os.OpenFile(fileFullPath, os.O_RDONLY, os.ModeType)
	if err != nil {
		w.Write([]byte("文件异常！"))
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("content-disposition", "attachment; filename=\""+fileName+"\"")
	//	buf := make([]byte, 20*1024*1024)
	fileInfo, _ := file.Stat()
	buf := make([]byte, fileInfo.Size())
	_, err = file.Read(buf)
	if nil != err {
		return
	}
	bufBase64 := base64.StdEncoding.EncodeToString(buf)
	w.Write([]byte(bufBase64))
}

func main() {
	loadFilesPath()
	http.HandleFunc("/download", handler)
	http.ListenAndServe(":8802", nil)
}
