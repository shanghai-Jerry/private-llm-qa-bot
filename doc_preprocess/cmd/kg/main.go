package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

var dpPath, officeJsonPath, format, textMindJson, pdfPath, pdfDir, jsonDir, formatJson, formatJsonDir, dir string
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
	flag.StringVar(&pdfDir, "pdf_dir", "", "最好是按页拆分之后，每页的图片文件")
	flag.StringVar(&dir, "dir", "", "input  dir")
	flag.IntVar(&outPage, "out_page", 1, "json out page num")
	flag.BoolVar(&pp_text_2_json, "pp_text2Json", true, "pp text2Json")
	flag.Parse()
}

func main() {

	start := time.Now()

	ppstructure()

	office_json_handler()

	if textMindJson != "" {
		parseTextMindTable(textMindJson)
	}

	fmt.Printf("total costs:%v(s) \n", time.Since(start).Seconds())
	fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++")
}

func office_json_handler() {

	// 4 ################################################################
	if len(pdfPath) > 0 {
		index := strings.LastIndex(pdfPath, "/")
		// 23.pdf
		inputFile := pdfPath[index+1:]
		index = strings.LastIndex(inputFile, ".")
		// 23
		inputFileName := strings.ReplaceAll(inputFile[:index], " ", "_")
		outputDir := "../../office/" + inputFileName
		imagesDir := outputDir + "/images"
		os.Mkdir(imagesDir, fs.ModePerm)
		f4, _ := os.Create(getOutFilePathFunc(outputDir, fmt.Sprintf("%v-out-section.txt", inputFileName)))
		f5, _ := os.Create(getOutFilePathFunc(outputDir, fmt.Sprintf("%v-out-section.json", inputFileName)))
		finalOutJsonFormat := &OutputJsonFormat{}
		officePDFParserV2(f4, finalOutJsonFormat, pdfPath, outputDir)
		finalOutJsonFormatBytes, _ := json.Marshal(finalOutJsonFormat)
		writeFileBytes(f5, finalOutJsonFormatBytes)

		fmt.Printf("finished file:%v", inputFileName)
		// pdf2Image(pdfPath)
		// put JPEGs in tmp folder under random prefix
		// jpegPrefix := generateRandomString(50)
		// jpegPath := fmt.Sprintf("%v/%s%%d.jpg", imagesDir, jpegPrefix)
		// smallJPEGPath := fmt.Sprintf("%v/%s%%d-small.jpg", imagesDir, jpegPrefix)
		// largeJPEGPath := fmt.Sprintf("%v/%s%%d-large.jpg", imagesDir, jpegPrefix)

		// numPages, err := convertPDFToJPEGs(pdfPath, jpegPath, smallJPEGPath,
		// 	largeJPEGPath)
		// if err != nil {
		// 	fmt.Println("convertPDFToJPEGs failed,", err)
		// 	panic("convertPDFToJPEGs failed")
		// }
		// fmt.Printf("finished file pdf2image:%v, numPages:%v", inputFileName, numPages)

	}
	// 1 ################################################################
	// load_data()
	if len(formatJson) > 0 {
		outDir := "../../office/paras"
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
		officeDataParseLayoutJsonFromat(f3, body)
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
		sectionTextTotalF, _ := os.Create(getOutFilePathFunc(outputDir, "total-out-section.txt"))
		sectionJsonTotalF, _ := os.Create(getOutFilePathFunc(outputDir, "total-out-section.json"))
		var paths []string
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
				paths = append(paths, path)
			}
			return nil
		}
		err := filepath.Walk(pdfDir, visitFunc)
		if err != nil {
			fmt.Printf("Error walk:%v", err)
		}

		sort.Slice(paths, func(i, j int) bool {
			return getInputFilePageIndex(paths[i]) < getInputFilePageIndex(paths[j])
		})
		finalOutJsonFormat := &OutputJsonFormat{}

		for _, path := range paths {
			officePDFParserV2(sectionTextTotalF, finalOutJsonFormat, path, outputDir)
		}
		finalOutJsonFormatBytes, _ := json.Marshal(finalOutJsonFormat)
		writeFileBytes(sectionJsonTotalF, finalOutJsonFormatBytes)
	}
}

func getOutFilePathFunc(dir, fileName string) string {
	return filepath.Join(dir, fileName)
}

func ppstructure() {
	if pp_text_2_json {
		if len(dir) > 0 {
			var filePaths []string
			outDir := fmt.Sprintf("%v/recover", dir)
			os.Mkdir(outDir, os.ModePerm)
			f3, _ := os.Create(fmt.Sprintf("%v/total_recover.txt", outDir))
			visitFunc := func(path string, info os.FileInfo, err error) error {
				if err != nil {
					fmt.Printf("Error accessing path %q: %v\n", path, err)
					return err
				}
				if info.IsDir() {
					fmt.Printf("%s is a directory\n", info.Name())
				} else if strings.HasSuffix(path, ".txt") && !strings.Contains(path, "recover") {
					if strings.HasSuffix(path, outDir) {
						return nil
					}
					filePaths = append(filePaths, path)
				}
				return nil
			}
			err := filepath.Walk(dir, visitFunc)
			if err != nil {
				fmt.Printf("Error walk:%v", err)
			}
			// 按文件页数排序处理
			sort.Slice(filePaths, func(i, j int) bool {
				return getInputFilePageIndex(filePaths[i]) < getInputFilePageIndex(filePaths[j])
			})
			for _, path := range filePaths {
				fmt.Printf(" ###### Processing File: %s\n", path)
				pp_text2Json(f3, path)
			}
		} else {
			if len(filePath) == 0 {
				return
			}
			f3, _ := os.Create(fmt.Sprintf("%v.recover.txt", filePath))
			pp_text2Json(f3, filePath)
		}
	}
}

func getInputFilePageIndex(filePath string) int {
	lastIndexOfDot := strings.LastIndex(filePath, ".")
	lastIndexOfDash := strings.LastIndex(filePath, "_")
	pageIndx := filePath[lastIndexOfDash+1 : lastIndexOfDot]
	index, _ := strconv.Atoi(pageIndx)
	return index
}

func dp_d2p() {
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
