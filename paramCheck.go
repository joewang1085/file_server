package service

import (
	"fmt"
	"model"
	//	"strconv"
	//	"sort"
	"path/filepath"
	"strings"

	"github.com/gopkg.in/mgo.v2/bson"
	"github.com/ripple"
)

func CheckFiles(ctx *ripple.Context) {
	defer model.RemoveAll(model.FilesPath + ctx.NewParams["userId"].(string) + string(filepath.Separator) + `parameter` + string(filepath.Separator) + `tmp` + string(filepath.Separator))
	defer CloseProcessMap(ctx.NewParams["userId"].(string), "参数核查")
	//	model.ProgressMap[ctx.NewParams["userId"].(string)] = &model.ProgressBar{}
	if _, ok := model.ProgressMap[ctx.NewParams["userId"].(string)]; !ok {
		model.ProgressMap[ctx.NewParams["userId"].(string)] = &model.ProgressBar{}
		model.ProgressMap[ctx.NewParams["userId"].(string)].CompareValue = float32(-1)
	}
	model.ProgressMap[ctx.NewParams["userId"].(string)].CheckProgress = make(chan float32, 0) // map 或 chan 要make初始化,开辟内存
	model.ProgressMap[ctx.NewParams["userId"].(string)].CheckValue = float32(0)
	progress := float32(0)
	response := make(bson.M, 0)
	resultBody := make([]bson.M, 0)
	resultBody = append(resultBody, bson.M{"title": "核查结果", "表头": []string{"网元", "子网", "小区", "字段", "MO Name", "标准值", "实际值"}})
	resultNum := int(0)
	userIdDir := ctx.NewParams["userId"].(string)
	var errCsvs []string

	go model.SetProcess(userIdDir, "参数核查", 0)
	excel, moNames, err := getExcel(ctx)
	if err != nil {
		response["状态码"] = "失败"
		response["错误"] = "模板文件异常"
		ctx.Response.Body = response
		return
	}

	paramPath := model.FilesPath + userIdDir + string(filepath.Separator) + `parameter`
	files, filesPaths, _ := model.ReadDir(paramPath)
	length := len(files)
	progressUnit := float32(1) / float32(length)

	for k, file := range files {
		go model.SetProcess(userIdDir, "参数核查", progress)
		progress += progressUnit
		if model.IsZip(filesPaths[k]) {
			tmpResult, tmpErrCsvs := checkZip(excel, file, filesPaths[k], moNames, userIdDir, progress-progressUnit, progressUnit)
			resultBody = append(resultBody, tmpResult...)
			errCsvs = append(errCsvs, tmpErrCsvs...)
		} else if 1 == 2 { //model.IsExcel(file) {
			/*多sheet xlsx*/
			tmpResult, tmpErrCsvs := checkExcel(excel, file, filesPaths[k], moNames, userIdDir, progress-progressUnit, progressUnit)
			resultBody = append(resultBody, tmpResult...)
			errCsvs = append(errCsvs, tmpErrCsvs...)
		} else {
			/*csv文件*/
			tmpResult, tmpErrCsvs := checkCsv(excel, file, filesPaths[k], moNames)
			resultBody = append(resultBody, tmpResult...)
			errCsvs = append(errCsvs, tmpErrCsvs...)
		}
	}

	response["状态码"] = "成功"
	resultNum++
	response["resultNum"] = resultNum
	result := make(bson.M, 0)
	result["0"] = sortResult(resultBody)
	response["resultBody"] = result
	errFiles := make(bson.M, 0)
	errFiles["异常参数文件"] = errCsvs
	response["异常文件"] = errFiles //暂时不能识别，非模板内的文件直接跳出
	ctx.Response.Body = response
}

func checkZip(excel *model.XlsxFile, fileName, zipPath string, moNames []string, userIdDir string, progress, progressUnit float32) (resultBody []bson.M, errCsvs []string) {
	defer model.RemoveAll(strings.TrimSuffix(zipPath, fileName) + `tmp`)
	model.Unzip(zipPath, strings.TrimSuffix(zipPath, fileName)+"tmp") // 加 err 异常处理
	files, filesPaths, dirs := model.ReadDir(strings.TrimSuffix(zipPath, fileName) + "tmp")
	progressUnit = progressUnit * float32(1) / float32(len(files)+len(dirs))
	for k, file := range files {
		go model.SetProcess(userIdDir, "参数核查", progress)
		progress += progressUnit
		if model.IsZip(filesPaths[k]) {
			/*嵌套zip不处理*/
			tmpResult, tmpErrCsvs := checkZip(excel, file, filesPaths[k], moNames, userIdDir, progress-progressUnit, progressUnit)
			resultBody = append(resultBody, tmpResult...)
			errCsvs = append(errCsvs, tmpErrCsvs...)
		} else if 1 == 2 {
			/*多sheet xlsx*/
		} else {
			/*csv文件*/
			tmpResult, tmpErrCsvs := checkCsv(excel, file, filesPaths[k], moNames)
			resultBody = append(resultBody, tmpResult...)
			errCsvs = append(errCsvs, tmpErrCsvs...)
		}
	}

	for _, dir := range dirs {
		go model.SetProcess(userIdDir, "参数核查", progress)
		progress += progressUnit
		dirPath := strings.TrimSuffix(zipPath, fileName) + `tmp` + string(filepath.Separator) + dir
		tmpResult, tmpErrCsvs := checkDir(excel, dirPath, moNames, userIdDir, progress-progressUnit, progressUnit)
		resultBody = append(resultBody, tmpResult...)
		errCsvs = append(errCsvs, tmpErrCsvs...)
	}
	return
}

func checkDir(excel *model.XlsxFile, dirPath string, moNames []string, userIdDir string, progress, progressUnit float32) (resultBody []bson.M, errCsvs []string) {
	files, filesPaths, dirs := model.ReadDir(dirPath)
	progressUnit = progressUnit * float32(1) / float32(len(files)+len(dirs))
	for k, file := range files {
		go model.SetProcess(userIdDir, "参数核查", progress)
		progress += progressUnit
		if model.IsZip(filesPaths[k]) {
			/*嵌套zip不处理？？？*/
			tmpResult, tmpErrCsvs := checkZip(excel, file, filesPaths[k], moNames, userIdDir, progress-progressUnit, progressUnit)
			resultBody = append(resultBody, tmpResult...)
			errCsvs = append(errCsvs, tmpErrCsvs...)
		} else if 1 == 2 {
			/*多sheet xlsx*/
		} else {
			/*csv文件*/
			tmpResult, tmpErrCsvs := checkCsv(excel, file, filesPaths[k], moNames)
			resultBody = append(resultBody, tmpResult...)
			errCsvs = append(errCsvs, tmpErrCsvs...)
		}
	}

	for _, dir := range dirs {
		go model.SetProcess(userIdDir, "参数核查", progress)
		progress += progressUnit
		tmpResult, tmpErrCsvs := checkDir(excel, dir, moNames, userIdDir, progress-progressUnit, progressUnit)
		resultBody = append(resultBody, tmpResult...)
		errCsvs = append(errCsvs, tmpErrCsvs...)
	}

	return
}

/**/
func checkExcel(excel *model.XlsxFile, file, path string, moNames []string, userIdDir string, progress, progressUnit float32) (resultBody []bson.M, errCsvs []string) {

	var paramExcel = new(model.XlsxFile)
	var err error
	paramExcel.FileName = file
	paramExcel.FilePath = path
	err = excel.ReadToSlice()
	if nil != err {
		errCsvs = append(errCsvs, file)
		fmt.Println(errCsvs)
		return
	}
	fmt.Println(paramExcel.FileName)
	fmt.Println(paramExcel.FilePath)
	fmt.Println(len(paramExcel.FileSlice))
	sheets := paramExcel.SheetsNames
	for _, sheet := range sheets {
		fmt.Println(sheet)
	}

	return
}

//暂时未调用
func checkCsvs(excel *model.XlsxFile, csvs, csvPaths, moNames []string, userIdDir string, progress, progressUnit float32) (resultBody []bson.M, errCsvs []string) {
	length := float32(len(csvs))
	for k, v := range csvPaths {
		go model.SetProcess(userIdDir, "参数核查", progress)
		progress += float32(k+1) / length * progressUnit
		tmpResult, tmpErrCsvs := checkCsv(excel, csvs[k], v, moNames)
		resultBody = append(resultBody, tmpResult...)
		errCsvs = append(errCsvs, tmpErrCsvs...)
	}

	return
}

func checkCsv(excel *model.XlsxFile, csvFile, filePath string, moNames []string) (resultBody []bson.M, errCsvs []string) {
	if !IscsvInMoNames(csvFile, moNames) {
		return
	}
	var csv = new(model.CsvFile)
	csv.FileName = csvFile
	csv.FormatCsvName()
	csv.FilePath = filePath
	if err := csv.ReadToSlice(); err != nil {
		errCsvs = append(errCsvs, csv.FileName)
		return
	}
	result := fileCheck(csv, excel)
	if len(result) == 0 {
		return
	}
	resultBody = append(resultBody, FormatResult(result, csv.FileName)...)

	return
}

func IscsvInMoNames(csv string, moNames []string) bool {
	for _, v := range moNames {
		if strings.Contains(v, strings.TrimSuffix(csv, `.csv`)) {
			return true
		}
	}
	return false
}

func CloseProcessMap(userId, task string) {
	go model.SetProcess(userId, task, 1)
}

func getExcel(ctx *ripple.Context) (*model.XlsxFile, []string, error) {
	var excel = new(model.XlsxFile)
	var err error
	excel.FileName = ctx.NewParams["templet"].(string)
	excel.FilePath = model.FilesPath + ctx.NewParams["userId"].(string) + string(filepath.Separator) + "templet" + string(filepath.Separator) + excel.FileName
	if ctx.NewParams["isCommonTemplet"].(bool) {
		excel.FilePath = model.FilesPath + "00000000" + string(filepath.Separator) + "templet" + string(filepath.Separator) + excel.FileName
	}
	err = excel.ReadToSlice()
	if nil != err {
		return nil, nil, err
	}

	return excel, getExcelMoNames(excel), nil
}

func getExcelMoNames(excel *model.XlsxFile) (moNames []string) {
	sheetName := excel.SheetsNames[0]
	head := excel.SheetHeader[sheetName]
	index := 2
	if v, ok := head["MO Name"]; ok {
		index = v
	}
	sheet := excel.FileSlice[sheetName]
	for k, row := range sheet {
		if k == 0 {
			continue
		}
		if row[index] == "" || row[index] == "MO Name" {
			continue
		}
		moNames = append(moNames, row[index])
	}
	return
}

func fileCheck(csv *model.CsvFile, excel *model.XlsxFile) (result [][]string) {
	sheetName := excel.SheetsNames[0]
	templetHeader := excel.SheetHeader[sheetName]
	fildNameIndex := templetHeader["Field Name"]
	masterValueIndex := templetHeader["Master Value"]
	templetRows := getTempletRows(csv, excel)
	for _, tempetRow := range templetRows {
		for k, checkRow := range csv.FileSlice {
			if k == 0 {
				continue
			}
			fildName := tempetRow[fildNameIndex]
			csvIndex := csv.Header[fildName]
			if !strings.Contains(tempetRow[masterValueIndex], checkRow[csvIndex]) {
				resultRow := []string{}
				mEID := "-"
				if mEIDIndedx, ok := csv.Header["MEID"]; ok {
					mEID = checkRow[mEIDIndedx]
				}
				subnet := "-"
				if subnetIndedx, ok := csv.Header["SubNetwork"]; ok {
					subnet = checkRow[subnetIndedx]
				}
				cellId := "-"
				if cellIDIndedx, ok := csv.Header["cellLocalId"]; ok {
					cellId = checkRow[cellIDIndedx]
				}
				moName := csv.FileName
				if name, ok := model.MoNameChineseVsEnglishMap[csv.FileName]; ok {
					moName = name + "(" + csv.FileName + ")"
				}
				fieldNameCN := fildName
				if name, ok := model.FieldNameChineseVsEnglishMap[fildName]; ok {
					fieldNameCN = name + "(" + fildName + ")"
				}

				resultRow = append(resultRow, mEID, subnet, cellId, fieldNameCN, moName, tempetRow[masterValueIndex], checkRow[csvIndex])
				result = append(result, resultRow)
			}
		}
	}
	return
}

func SaveResult(userId, filePath, sheetName string, result [][]string) (err error) {
	var excel = new(model.XlsxFile)
	excel.FileName = `result.xlsx`
	excel.FilePath = filePath
	excel.FileSlice = make(map[string][][]string, 0)
	excel.FileSlice[sheetName] = result
	err = excel.Create(userId, "参数核查")
	return
}

func getTempletRows(csv *model.CsvFile, excel *model.XlsxFile) (rows [][]string) {
	sheetName := excel.SheetsNames[0]
	sheet := excel.FileSlice[sheetName]
	length := len(sheet)
	header := excel.SheetHeader[sheetName]
	var index = header["MO Name"]
	var start int
	for i := 1; i < length; i++ {
		row := sheet[i]
		if strings.Contains(row[index], csv.FileName) {
			start = i
			break
		}
	}
	for i := start + 1; i < length; i++ {
		row := sheet[i]
		if row[index] == "" {
			rows = append(rows, row)
			continue
		}
		break
	}

	return
}

// 暂时不需要 title，待定
func FormatResult(result [][]string, title string) (formatResult []bson.M) {
	head := []string{"网元", "子网", "小区", "字段", "MO Name", "标准值", "实际值"}
	for _, row := range result {
		formatRow := make(bson.M, 0)
		for m, v := range row {
			formatRow[head[m]] = v
		}
		formatResult = append(formatResult, formatRow)
	}
	return
}

func quickSortByKey(arr []bson.M, key string, start, end int) []bson.M {
	if start < end {
		i, j := start, end
		value := arr[(start+end)/2][key].(string)
		for i <= j {
			for arr[i][key].(string) > value {
				i++
			}
			for arr[j][key].(string) < value {
				j--
			}

			if i <= j {
				arr[i], arr[j] = arr[j], arr[i] //666
				i++
				j--
			}
		}

		if start < j {
			quickSortByKey(arr, key, start, j)
		}
		if end > i {
			quickSortByKey(arr, key, i, end)
		}
	}

	return arr
}

func quickSortByKeyAsc(arr []bson.M, key string, start, end int) []bson.M {
	if start < end {
		i, j := start, end
		value := arr[(start+end)/2][key].(string)
		for i <= j {
			for arr[i][key].(string) < value {
				i++
			}
			for arr[j][key].(string) > value {
				j--
			}

			if i <= j {
				arr[i], arr[j] = arr[j], arr[i] //666
				i++
				j--
			}
		}

		if start < j {
			quickSortByKeyAsc(arr, key, start, j)
		}
		if end > i {
			quickSortByKeyAsc(arr, key, i, end)
		}
	}

	return arr
}

func sortResult(arr []bson.M) []bson.M {
	if len(arr) <= 2 {
		return arr
	}
	key := []string{"网元", "小区", "字段"}

	arr = quickSortByKeyAsc(arr, key[0], 1, len(arr)-1) //表头不能排序
	var start, end int
	var mEId, subnet string

	start, end = 1, 1
	mEId = arr[start]["网元"].(string)
	for k, v := range arr {
		if k == 0 {
			continue
		}

		if v[key[0]].(string) == mEId {
			end = k
			continue
		}
		arr = quickSortByKey(arr, key[1], start, end)
		start = k
		mEId = v[key[0]].(string)
	}

	start, end = 1, 1
	mEId, subnet = arr[start]["网元"].(string), arr[start]["子网"].(string)
	for k, v := range arr {
		if k == 0 {
			continue
		}

		if v[key[0]].(string) == mEId && v[key[1]].(string) == subnet {
			end = k
			continue
		}
		arr = quickSortByKeyAsc(arr, key[2], start, end)
		start = k
		mEId, subnet = v[key[0]].(string), v[key[1]].(string)
	}

	return arr
}
