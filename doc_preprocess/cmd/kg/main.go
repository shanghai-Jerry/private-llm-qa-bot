package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var dpPath, officeJsonPath, format, textMindJson, pdfPath, pdfDir, jsonDir, formatJson, formatJsonDir string
var filePath string
var outPage int
var pp_text_2_json bool

func init() {
	flag.StringVar(&format, "f", "doc", "input file format")
	flag.StringVar(&dpPath, "dp_path", "", "input file path")
	flag.StringVar(&filePath, "path", "", "input file path")
	flag.StringVar(&officeJsonPath, "json", "", `json file path, like: ./office/23-page-6.json`)
	flag.StringVar(&formatJson, "format_json", "", ``)
	flag.StringVar(&formatJsonDir, "format_json_dir", "", ``)
	flag.StringVar(&jsonDir, "json_dir", "", `json file path, like: ./office/`)
	flag.StringVar(&textMindJson, "tjson", "", `textming json file path, like: ./textmind/23.textmind.json`)
	flag.StringVar(&pdfPath, "pdf_path", "", "input pdf file path， like：./b_data/pdf/23.pdf")
	flag.StringVar(&pdfDir, "pdf_dir", "", "input pdf dir like：./b_data/pdf, 将其中的所有文件处理")
	flag.IntVar(&outPage, "out_page", 1, "json out page num")
	flag.BoolVar(&pp_text_2_json, "pp_text2Json", false, "pp text2Json")
	flag.Parse()
}

func ppstructure() {
	if pp_text_2_json {
		pp_text2Json(filePath)
	}
}

func main() {

	start := time.Now()

	ppstructure()

	office_json_handler()

	fmt.Printf("total costs:%v(s) \n", time.Since(start).Seconds())
	fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++")
}

func office_json_handler() {
	// 1 ################################################################
	// load_data()
	if len(formatJson) > 0 {
		outDir := "./office/paras"
		get_paras_from_format_json(formatJson, outDir)
	}
	if len(formatJsonDir) > 0 {
		outDir := fmt.Sprintf("%v/paras", formatJsonDir)
		os.Mkdir(outDir, fs.ModePerm)
		visitFunc := func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf("Error accessing path %q: %v\n", path, err)
				return err
			}
			if info.IsDir() {
				fmt.Printf("%s is a directory\n", info.Name())
			} else if strings.HasSuffix(path, "format.json") {
				// 过滤子目录
				if strings.HasPrefix(path, outDir) {
					return nil
				}
				fmt.Printf(" ###### Processing File: %s\n", path)
				get_paras_from_format_json(path, outDir)
			}
			return nil
		}
		err := filepath.Walk(formatJsonDir, visitFunc)
		if err != nil {
			fmt.Printf("Error walk:%v", err)
		}
	}
	// 2 ################################################################
	parseJsonFunc := func(officeJson string) {
		sep := strings.LastIndex(officeJson, "/")
		dir := officeJson[:sep]
		suffixIndex := strings.LastIndex(officeJson, ".")
		formatFileName := officeJson[sep+1 : suffixIndex]
		fileName := fmt.Sprintf("%v/%v-out.txt", dir, formatFileName)
		f, _ := os.Create(fileName)
		f2, _ := os.Create(fmt.Sprintf("%v.log", officeJson))
		f3, _ := os.Create(fmt.Sprintf("%v/%v.format.json", dir, formatFileName))
		body, _ := os.ReadFile(officeJson)
		officeDataParseLayout(f, f2, body)
		officeDataParseLayoutV2(f3, body)
	}
	if len(officeJsonPath) > 0 {
		parseJsonFunc(officeJsonPath)
	}
	// ################################################################
	if len(jsonDir) > 0 {
		visitFunc := func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf("Error accessing path %q: %v\n", path, err)
				return err
			}
			if info.IsDir() {
				fmt.Printf("%s is a directory\n", info.Name())
			} else if strings.HasSuffix(path, ".json") {
				if strings.HasSuffix(path, ".format.json") {
					return nil
				}
				fmt.Printf(" ###### Processing File: %s\n", path)
				parseJsonFunc(path)
			}
			return nil
		}
		err := filepath.Walk(jsonDir, visitFunc)
		if err != nil {
			fmt.Printf("Error walk:%v", err)
		}

	}
	// 3 ################################################################
	if len(pdfDir) > 0 {
		// 输出目录
		outputDir := fmt.Sprintf("%v/output", pdfDir)
		os.Mkdir(outputDir, fs.ModePerm)
		visitFunc := func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf("Error accessing path %q: %v\n", path, err)
				return err
			}
			if info.IsDir() {
				fmt.Printf("%s is a directory\n", info.Name())
			} else {
				if strings.HasPrefix(path, outputDir) {
					return nil
				}
				fmt.Printf(" ###### Processing File: %s\n", path)
				officePDFParser(path, outputDir)
			}
			return nil
		}
		err := filepath.Walk(pdfDir, visitFunc)
		if err != nil {
			fmt.Printf("Error walk:%v", err)
		}
	}
	// 4 ################################################################
	// printDPContent()
	if len(pdfPath) > 0 {
		fmt.Printf("start parsing... %v\n", pdfPath)
		index := strings.LastIndex(pdfPath, "/")
		pdfFile := GetFileContentAsBase64(pdfPath)
		// 23.pdf
		inputFile := pdfPath[index+1:]
		index = strings.LastIndex(inputFile, ".")
		// 23
		inputFileName := strings.ReplaceAll(inputFile[:index], " ", "_")
		fmt.Printf("fileName:%v\n", inputFileName)
		// 输出目录
		outputDir := "./office/" + inputFileName
		os.Mkdir(outputDir, fs.ModePerm)
		if outPage > 1 {
			outputFileName := fmt.Sprintf("%v-page-%v.json", inputFileName, outPage)
			outputFileNameTxt := fmt.Sprintf("%v-page-%v-out.txt", inputFileName, outPage)
			outJsonF, _ := os.Create(outputDir + "/" + outputFileName)
			outTxtF, _ := os.Create(outputDir + "/" + outputFileNameTxt)
			f1, _ := os.Create(fmt.Sprintf(fmt.Sprintf("%v/%v-page-%v.log", outputDir, inputFileName, outPage)))
			fmt.Printf("--- output json of page %d ---- \n", outPage)
			pageJson, _ := parsePDF(pdfFile, outPage)
			writeFileBytes(outJsonF, pageJson)
			officeDataParseLayout(outTxtF, f1, pageJson)
			return
		}
		firstPage := fmt.Sprintf("%v-page-%v.json", inputFileName, 1)
		outputFileTotal := fmt.Sprintf("%v-total.json", inputFileName)

		firstOutJsonF, _ := os.Create(outputDir + "/" + firstPage)
		outputFileTotalF, _ := os.Create(outputDir + "/" + outputFileTotal)

		fileName := fmt.Sprintf("%v-out.txt", pdfPath)
		fileNameJson := fmt.Sprintf("%v-out.json", pdfPath)

		f, _ := os.Create(fileName)
		f2, _ := os.Create(fileNameJson)
		f3, _ := os.Create(fmt.Sprintf("%v.log", pdfPath))
		var retResps []*OfficeJSONData

		fmt.Printf("========= parsing page %d ========== \n", 1)
		firstPageRetJson, err := parsePDF(pdfFile, 1)
		if err != nil {
			fmt.Printf("parsePDF err:%v", err.Error())
			return
		}
		startParse := time.Now()

		var resp OfficeJSONData
		_ = json.Unmarshal(firstPageRetJson, &resp)
		totalPage := resp.PDFFileSize
		for page := 2; page <= totalPage; page++ {
			fmt.Printf("========= parsing page %d ========== \n", page)
			retJson, err := parsePDF(pdfFile, page)
			if err != nil {
				return
			}
			var pageResp OfficeJSONData
			_ = json.Unmarshal(retJson, &pageResp)
			retResps = append(retResps, &pageResp)
		}
		fmt.Printf("parsing costs:%v(s) \n", time.Since(startParse).Seconds())
		// 默认输出第一夜解析josn
		writeFileBytes(firstOutJsonF, firstPageRetJson)

		MergeOfficeRetJson(&resp, retResps)
		totalRetJosnBytes, _ := json.Marshal(&resp)
		fmt.Printf("--- output total file json --- \n")
		writeFileBytes(outputFileTotalF, totalRetJosnBytes)
		fmt.Printf("--- output total file txt  --- \n")
		officeDataParseLayout(f, f3, totalRetJosnBytes)
		fmt.Printf("--- output total file foramt json --- \n")
		officeDataParseLayoutV2(f2, totalRetJosnBytes)
	}

	// parseTextMindTable(textMindJson)
	// 5 ################################################################
	if len(dpPath) > 0 {
		resultBytes, err := dp(dpPath, format)
		if err != nil {
			panic(err)
		}
		d2pRespBytes := d2p(resultBytes)
		// fmt.Println("d2p result:", string(d2pRespBytes))
		var d2pResp D2PResponse
		writeContent(dpPath+"-d2p.json", d2pRespBytes)
		err = json.Unmarshal(d2pRespBytes, &d2pResp)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("paras:", len(d2pResp.Paras))

	}
}
