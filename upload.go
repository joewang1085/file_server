//web 上传文件
package service

import (
	//	"fmt"
	"io"
	"model"
	"net/http"
	"os"
	//	"bytes"
	//	"io/ioutil"
	//	"mime/multipart"

	"path/filepath"

	"github.com/gopkg.in/mgo.v2/bson"
	"github.com/ripple"
)

/*客户端上传文件*/
func UploadFile(ctx *ripple.Context) {
	var err error
	r := ctx.Request
	switch r.Method {
	//POST takes the uploaded file(s) and saves it to disk.
	case "POST":
		//parse the multipart form in the request
		err = r.ParseMultipartForm(100000)
		if err != nil {
			ctx.Response.Body = bson.M{"状态码": "失败", "错误": err}
			return
		}
		//get a ref to the parsed multipart form
		m := r.MultipartForm
		//get the *fileheaders
		files := m.File["myfile"]
		for i, _ := range files {
			//for each fileheader, get a handle to the actual file
			file, err2 := files[i].Open()
			defer file.Close()
			if err2 != nil {
				ctx.Response.Body = bson.M{"状态码": "失败", "错误": err2}
				return
			}
			//create destination file making sure the path is writeable.
			userId := ctx.NewParams["userId"].([]string)
			dir := ctx.NewParams["dir"].([]string)
			if dir[0] == "模板" {
				dir[0] = "templet"
			}
			if dir[0] == "参数" {
				dir[0] = "parameter"
			}
			if dir[0] == "对比参数" {
				dir[0] = "compareParameter"
			}
			err = model.Mkdir(model.FilesPath + string(filepath.Separator) + userId[0] + string(filepath.Separator))
			if err != nil {
				ctx.Response.Body = bson.M{"状态码": "失败", "错误": err}
				return
			}
			err = model.Mkdir(model.FilesPath + string(filepath.Separator) + userId[0] + string(filepath.Separator) + dir[0] + string(filepath.Separator))
			if err != nil {
				ctx.Response.Body = bson.M{"状态码": "失败", "错误": err}
				return
			}
			//create file
			dst, err1 := os.Create(model.FilesPath + string(filepath.Separator) + userId[0] + string(filepath.Separator) + dir[0] + string(filepath.Separator) + files[i].Filename)
			defer dst.Close()
			if err1 != nil {
				ctx.Response.Body = bson.M{"状态码": "失败", "错误": err1}
				return
			}
			//copy the uploaded file to the destination file
			if _, err = io.Copy(dst, file); err != nil {
				ctx.Response.Body = bson.M{"状态码": "失败", "错误": err}
				return
			}
		}

	default:
		ctx.Response.Body = bson.M{"状态码": "失败", "错误": http.StatusMethodNotAllowed}
	}
	ctx.Response.Body = bson.M{"状态码": "成功"}
}
