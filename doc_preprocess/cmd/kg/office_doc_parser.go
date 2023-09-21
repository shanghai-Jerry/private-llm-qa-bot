package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/hashicorp/go-uuid"
)

const (
	text     = "text"
	table    = "table"
	title    = "title"
	contents = "contents"
)

type OutputJson struct {
	Tokens  int    `json:"tokens"`
	Content string `json:"content"`
	// table, text, title
	Type string `json:"type"`
	// 第几页
	Pages int `json:"page"`
}

type OfficeJSONData struct {
	// 将多页内容合并到一个json文件中
	// 合并内容: results, layouts, tables_result
	// 需要额外新增考虑：
	// 1. layouts中的idx在result中的偏移位置
	// 3. tables之间按顺序检索即可
	LayoutOffset int       // 相对于上一个page的offset
	Layouts      []*Layout `json:"layouts"`
	// 版面分析结果数，表示layout的元素个数,
	// layout_analysis=true时返回
	LayoutsNum  int        `json:"layouts_num"`
	LogID       string     `json:"log_id"`
	PDFFileSize int        `json:"pdf_file_size"`
	Results     []*Results `json:"results"`
	// 识别结果数，表示results的元素个数
	ResultsNum   int            `json:"results_num"`
	SealRecogNum int            `json:"seal_recog_num"`
	SecCols      int            `json:"sec_cols"`
	SecRows      int            `json:"sec_rows"`
	Sections     []Sections     `json:"sections"`
	TableNum     int            `json:"table_num"`
	TablesResult []TablesResult `json:"tables_result"`
}

func addExtraMetaToOfficeJson(json *OfficeJSONData, pageNo int, startOffset, offset int) {
	for _, result := range json.Results {
		result.PageNo = pageNo
	}
	for _, layout := range json.Layouts {
		layout.PageNo = pageNo
		// 更新offset
		layout.LayoutOffset += (startOffset + offset)
	}
}

func MergeOfficeRetJson(retJson *OfficeJSONData, pageRetJsons []*OfficeJSONData) {

	addExtraMetaToOfficeJson(retJson, 1, 0, 0)

	fmt.Printf("result size:%v, layout size:%v, table size:%v \n", len(retJson.Results), len(retJson.Layouts), len(retJson.TablesResult))
	for index, pageJson := range pageRetJsons {
		pageNo := index + 2
		for _, result := range pageJson.Results {
			result.PageNo = pageNo
			retJson.Results = append(retJson.Results, result)
		}
		var offset int
		// 上一个page的results数量
		if index == 0 {
			offset = retJson.ResultsNum
		} else {
			offset = pageRetJsons[index-1].ResultsNum
		}
		retJson.ResultsNum += len(pageJson.Results)
		for _, layout := range pageJson.Layouts {
			// 更新offset
			layout.LayoutOffset += (retJson.LayoutOffset + offset)
			layout.PageNo = index + 1
			retJson.Layouts = append(retJson.Layouts, layout)
		}
		retJson.LayoutOffset += offset
		retJson.LayoutsNum += len(pageJson.Layouts)
		for _, tr := range pageJson.TablesResult {
			retJson.TablesResult = append(retJson.TablesResult, tr)
		}
		retJson.TableNum += len(pageJson.TablesResult)
		fmt.Printf("Merging result size:%v, layout size:%v, table size:%v \n", retJson.ResultsNum, retJson.LayoutsNum, retJson.TableNum)
	}
}

type Layout struct {
	// 版面分析的标签结果。
	// 表格:table， 图:figure，文本:text，段落标题:title ，
	// 目录:contents，印章:seal，表标题: table_title，
	// 图标题: figure_title，文档标题：doc_title
	LayoutOffset int
	PageNo       int
	Layout       string `json:"layout"`
	// 对应results中下标位置
	LayoutIdx      []int `json:"layout_idx"`
	LayoutLocation []struct {
		X int `json:"x"`
		Y int `json:"y"`
	} `json:"layout_location"`
	LayoutProb float64 `json:"layout_prob"`
}

type Words struct {
	Word          string `json:"word"`
	WordsLocation struct {
		Height int `json:"height"`
		Left   int `json:"left"`
		Top    int `json:"top"`
		Width  int `json:"width"`
	} `json:"words_location"`
	WordsType string `json:"words_type"`
}

type Results struct {
	PageNo    int
	Words     Words  `json:"words"`
	WordsType string `json:"words_type"`
}

type Sections struct {
	AttriLocation struct {
		Points []struct {
			X int `json:"x"`
			Y int `json:"y"`
		} `json:"points"`
	} `json:"attri_location"`
	// 版面分析的属性标签结果，
	// 栏:section, 页眉:header, 页脚:footer, 页码:number，脚注:footnote
	Attribute string `json:"attribute"`
	SecIdx    struct {
		ColIdx []int `json:"col_idx"`
		// 对应results中下标位置
		Idx []int `json:"idx"`
		// 对应layouts中下标位置
		ParaIdx []int `json:"para_idx"`
		RowIdx  []int `json:"row_idx"`
	} `json:"sec_idx"`
	SectionsProb float64 `json:"sections_prob"`
}

type TContent struct {
	PolyLocation []struct {
		X int `json:"x"`
		Y int `json:"y"`
	} `json:"poly_location"`
	Word string `json:"word"`
}

type TableBody struct {
	CellLocation []struct {
		X int `json:"x"`
		Y int `json:"y"`
	} `json:"cell_location"`
	ColEnd   int        `json:"col_end"`
	ColStart int        `json:"col_start"`
	Contents []TContent `json:"contents"`
	RowEnd   int        `json:"row_end"`
	RowStart int        `json:"row_start"`
	Words    string     `json:"words"`
}

type TablesResult struct {
	Body []TableBody `json:"body"`
}

const (
	layout_table        = "table"
	layout_text         = "text"
	layout_figure       = "figure"
	layout_title        = "title"
	layout_content      = "contents"
	layout_table_title  = "table_title"
	layout_seal         = "seal"
	layout_figure_title = "figure_title"
	layout_doc_title    = "doc_title"
	layout_text_title   = "text_title"
	// textmind's layout types with different define

	// equation
	layout_equation       = "equation"
	layout_header         = "header"
	layout_figure_caption = "figure_caption"
)

func buildTable(data []TableBody) [][]string {
	// 构建表格
	var numRows, numCols int
	for _, cell := range data {
		if cell.RowEnd > numRows {
			numRows = cell.RowEnd
		}
		if cell.ColEnd > numCols {
			numCols = cell.ColEnd
		}
	}
	table := make([][]string, numRows)
	for i := range table {
		table[i] = make([]string, numCols)
	}
	// 合并的单元格，分开来填充
	for _, cell := range data {
		for i := cell.RowStart; i < cell.RowEnd; i++ {
			for j := cell.ColStart; j < cell.ColEnd; j++ {
				var v string
				if len(cell.Contents) == 0 {
					v = cell.Words
				} else {
					v = getCellContent(cell.Contents)
				}

				table[i][j] = v
			}
		}
	}
	return table
}

func writeTable(f *os.File, table [][]string) {
	w := tabwriter.NewWriter(f, 0, 0, 1, ' ', tabwriter.Debug)
	for _, row := range table {
		for _, cell := range row {
			fmt.Fprint(w, cell, "|")
		}
		fmt.Fprintln(w)
	}
	w.Flush()
}
func writeTableStringBuffer(table [][]string) string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', tabwriter.Debug)
	for _, row := range table {
		for _, cell := range row {
			fmt.Fprint(w, cell, "|")
		}
		fmt.Fprintln(w)
	}
	w.Flush()
	return buf.String()
}

func writeFileBytes(f *os.File, data []byte) {
	writer := bufio.NewWriter(f)
	writer.Write(data)
	writer.Flush()
}

func writeFile(f *os.File, data []string) {
	writer := bufio.NewWriter(f)
	for _, d := range data {
		// 过滤掉长度为1的短字符串
		if len(d) == 1 {
			continue
		}
		writer.WriteString(d + "\n")
	}
	writer.Flush()
}
func writeFileLine(f *os.File, d string) {
	writer := bufio.NewWriter(f)
	writer.WriteString(d + "\n")
	writer.Flush()
}

func writeTableFile(fileName string, table [][]string) {
	file, _ := os.Create(fileName + ".table.txt")
	w := tabwriter.NewWriter(file, 0, 0, 1, ' ', tabwriter.Debug)
	for _, row := range table {
		for _, cell := range row {
			fmt.Fprint(w, cell, "|")
		}
		fmt.Fprintln(w)
	}
	w.Flush()
}

func getCellContent(contents []TContent) string {
	var words []string
	for _, c := range contents {
		words = append(words, c.Word)
	}
	return strings.Join(words, " ")
}

func getResultParas(result []*Results, start, end int) []string {
	var words []string
	for i := start; i < end; i++ {
		words = append(words, result[i].Words.Word)
	}
	return words
}

func isNumeric(s string) bool {
	match, _ := regexp.MatchString("^[0-9]+$", s)
	return match
}

func getContentParas(result []*Results, start, end int) []string {
	var words []string
	i := start
	for i < end {
		var lines []string
		current := i
		w := result[current].Words.Word
		nextw := result[current+1].Words.Word
		if isNumeric(nextw) {
			lines = append(lines, w)
			lines = append(lines, nextw)
			words = append(words, strings.Join(lines, "\t"))
			i += 2
		} else {
			words = append(words, w)
			i += 1
		}
	}
	return words
}

func officeDataParseLayout(f *os.File, logF *os.File, retJson []byte) {
	if len(retJson) == 0 {
		return
	}
	var resp OfficeJSONData
	_ = json.Unmarshal(retJson, &resp)
	writeFileLine(logF, fmt.Sprintf("result size:%v, layout size:%v", len(resp.Results), len(resp.Layouts)))
	tableIndex := 0
	var outString []string
	lastIndex := -1
	results := resp.Results
	for _, lout := range resp.Layouts {
		layout := lout.Layout
		layoutIndexs := lout.LayoutIdx
		if len(layoutIndexs) == 0 {
			writeFileLine(logF, fmt.Sprintf("layoutIndexs is empty，page: %v", lout.PageNo))
			continue
		}
		result := resp.Results[layoutIndexs[0]+lout.LayoutOffset]
		switch layout {
		// 图表(识别出的文本丢掉)
		case layout_figure:
			startIndex := lout.LayoutIdx[0] + lout.LayoutOffset
			endIndex := lout.LayoutIdx[len(lout.LayoutIdx)-1] + lout.LayoutOffset
			outString = append(outString, getResultParas(results, lastIndex+1, startIndex)...)
			imageIdStr, _ := uuid.GenerateUUID()
			outString = append(outString, fmt.Sprintf("[image: %v]", imageIdStr))
			writeFile(f, outString)
			// 清空已写入数据
			outString = []string{}
			lastIndex = endIndex
		// 表格
		case layout_table:
			index := lout.LayoutIdx[0] + lout.LayoutOffset
			outString = append(outString, getResultParas(results, lastIndex+1, index)...)
			lastIndex = lout.LayoutIdx[len(lout.LayoutIdx)-1] + lout.LayoutOffset
			writeFile(f, outString)
			// 清空已写入数据
			outString = []string{}
			table := buildTable(resp.TablesResult[tableIndex].Body)
			writeTable(f, table)
			writeFileLine(logF, fmt.Sprintf("office: %v, i:%v, l:%v\n", layout_table, index, lastIndex))
			tableIndex++
		// 表格标题
		case layout_table_title:
			title := result.Words.Word
			writeFileLine(logF, fmt.Sprintf("office: %v, %v\n", layout_table_title, title))
		// 正常的文本
		// layout_content: 可能需要特殊处理一下
		case layout_content:
			start := lout.LayoutIdx[0] + lout.LayoutOffset
			outString = append(outString, getContentParas(results, lastIndex+1, start)...)
			end := lout.LayoutIdx[len(lout.LayoutIdx)-1] + lout.LayoutOffset
			outString = append(outString, getContentParas(results, start, end)...)
			lastIndex = end
			writeFile(f, outString)
			// 清空已写入数据
			outString = []string{}
		case layout_text:
		// text := result.Words.Word
		// 段落标题: 每个段落都有一个标题
		case layout_text_title:
			title := result.Words.Word
			index := lout.LayoutIdx[0] + lout.LayoutOffset
			writeFileLine(logF, fmt.Sprintf("office: %v, %v, lix:%v\n", layout_text_title, title, index))
		// 文档标题
		case layout_doc_title:
			title := result.Words.Word
			writeFileLine(logF, fmt.Sprintf("%v:%v\n", layout_doc_title, title))
		}
	}
	if lastIndex < len(results)-1 {
		outString = append(outString, getResultParas(results, lastIndex+1, len(results))...)
	}
	if len(outString) > 0 {
		writeFile(f, outString)
	}
}

func officeDataParseLayoutV2(f *os.File, retJson []byte) {
	if len(retJson) == 0 {
		return
	}
	var resp OfficeJSONData
	_ = json.Unmarshal(retJson, &resp)
	tableIndex := 0
	var outString []string
	lastIndex := -1
	results := resp.Results
	var docTitles []string
	var outPutJsons []*OutputJson
	for _, lout := range resp.Layouts {
		layout := lout.Layout
		pageNo := lout.PageNo
		layoutIndexs := lout.LayoutIdx
		if len(layoutIndexs) == 0 {
			fmt.Printf("layoutIndexs is empty，page: %v \n", lout.PageNo)
			continue
		}
		result := resp.Results[layoutIndexs[0]+lout.LayoutOffset]
		switch layout {
		// 图表(识别出的文本丢掉)
		case layout_figure:
			startIndex := lout.LayoutIdx[0] + lout.LayoutOffset
			endIndex := lout.LayoutIdx[len(lout.LayoutIdx)-1] + lout.LayoutOffset
			outString = append(outString, getResultParas(results, lastIndex+1, startIndex)...)
			imageIdStr, _ := uuid.GenerateUUID()
			outString = append(outString, fmt.Sprintf("[image: %v]", imageIdStr))
			outputJson := &OutputJson{
				Type:    "text",
				Pages:   pageNo,
				Content: strings.Join(outString, "\n"),
			}
			outPutJsons = append(outPutJsons, outputJson)
			// 清空已写入数据
			outString = []string{}
			lastIndex = endIndex
		// 表格
		case layout_table:
			index := lout.LayoutIdx[0] + lout.LayoutOffset
			outString = append(outString, getResultParas(results, lastIndex+1, index)...)
			lastIndex = lout.LayoutIdx[len(lout.LayoutIdx)-1] + lout.LayoutOffset
			if len(outString) > 0 {
				outputJson := &OutputJson{

					Type:    "text",
					Pages:   pageNo,
					Content: strings.Join(outString, "\n"),
				}
				outPutJsons = addOutputJson(outPutJsons, outputJson)
				// 清空已写入数据
				outString = []string{}
			}
			table := buildTable(resp.TablesResult[tableIndex].Body)
			outputJson := &OutputJson{

				Type:    "table",
				Pages:   pageNo,
				Content: writeTableStringBuffer(table),
			}
			outPutJsons = addOutputJson(outPutJsons, outputJson)
			tableIndex++
		// 表格标题
		case layout_table_title:
			index := lout.LayoutIdx[0] + lout.LayoutOffset
			lastIndex = index
			title := result.Words.Word
			outString = append(outString, title)
			outputJson := &OutputJson{
				Type:    "text",
				Pages:   pageNo,
				Content: strings.Join(outString, "\n"),
			}
			outPutJsons = addOutputJson(outPutJsons, outputJson)
			// 清空已写入数据
			outString = []string{}
		case layout_content:
			start := lout.LayoutIdx[0] + lout.LayoutOffset
			end := lout.LayoutIdx[len(lout.LayoutIdx)-1] + lout.LayoutOffset
			outString = append(outString, getResultParas(results, lastIndex+1, start)...)
			if len(outString) != 0 {
				outputJson := &OutputJson{
					Type:    "text",
					Pages:   lout.PageNo,
					Content: strings.Join(outString, "\n"),
				}
				// 清空已写入数据
				outString = []string{}
				outPutJsons = addOutputJson(outPutJsons, outputJson)
			}
			// add contents
			outputJson := &OutputJson{
				Type:    "contents",
				Pages:   pageNo,
				Content: strings.Join(getContentParas(results, start, end), "\n"),
			}
			outPutJsons = addOutputJson(outPutJsons, outputJson)
			lastIndex = end
		// 正常的文本
		case layout_text, layout_figure_title:
			text := result.Words.Word
			outString = append(outString, text)
			index := lout.LayoutIdx[0] + lout.LayoutOffset
			lastIndex = index
		// 段落标题: 每个段落都有一个标题
		case layout_text_title:
			title := result.Words.Word
			index := lout.LayoutIdx[0] + lout.LayoutOffset
			outString = append(outString, getResultParas(results, lastIndex+1, index)...)
			lastIndex = index
			if len(outString) != 0 {
				outputJson := &OutputJson{
					Type:    "text",
					Pages:   pageNo,
					Content: strings.Join(outString, "\n"),
				}
				// 清空已写入数据
				outString = []string{}
				outPutJsons = addOutputJson(outPutJsons, outputJson)
			}
			// add title
			outputJson := &OutputJson{
				Type:    "title",
				Pages:   pageNo,
				Content: title,
			}
			outPutJsons = addOutputJson(outPutJsons, outputJson)
		// 文档标题
		case layout_doc_title:
			title := result.Words.Word
			docTitles = append(docTitles, title)
		}
	}
	if lastIndex < len(results)-1 {
		outputJson := &OutputJson{
			Type:    "text",
			Pages:   results[lastIndex].PageNo,
			Content: strings.Join(outString, "\n"),
		}
		outPutJsons = addOutputJson(outPutJsons, outputJson)
	}
	for _, o := range outPutJsons {
		o.Tokens = CountTokens(o.Content)
	}
	formatOutJson := &OutputJsonFormat{
		DocTitles:  docTitles,
		OutputJson: outPutJsons,
	}
	outputBytes, _ := json.Marshal(formatOutJson)
	writeFileBytes(f, outputBytes)
}

func addOutputJson(o []*OutputJson, json *OutputJson) []*OutputJson {
	if len(json.Content) > 2 {
		o = append(o, json)
	}
	return o
}

func officePDFParser(pdfPath string, outDir string) {
	fmt.Printf("start parsing... %v\n", pdfPath)
	index := strings.LastIndex(pdfPath, "/")
	pdfFile := GetFileContentAsBase64(pdfPath)
	// 23.pdf
	inputFile := pdfPath[index+1:]
	index = strings.LastIndex(inputFile, ".")
	// 23
	inputFileName := strings.ReplaceAll(inputFile[0:index], " ", "_")
	outFileJsonPath := fmt.Sprintf("%s/%s.json", outDir, inputFileName)
	outFileTxtPath := fmt.Sprintf("%s/%s.txt", outDir, inputFileName)
	outFileFormatJsonPath := fmt.Sprintf("%s/%s.format.json", outDir, inputFileName)
	logPath := fmt.Sprintf("%s/%s.log", outDir, inputFileName)
	f1, _ := os.Create(outFileJsonPath)
	f2, _ := os.Create(outFileTxtPath)
	f3, _ := os.Create(outFileFormatJsonPath)
	f4, _ := os.Create(logPath)
	// 开始处理pdf
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
	MergeOfficeRetJson(&resp, retResps)
	totalRetJosnBytes, _ := json.Marshal(&resp)
	fmt.Printf("========= output total file json  ========== \n")
	writeFileBytes(f1, totalRetJosnBytes)
	fmt.Printf("========= output total file txt  ========== \n")
	officeDataParseLayout(f2, f4, totalRetJosnBytes)
	fmt.Printf("========= output total file foramt json  ========== \n")
	officeDataParseLayoutV2(f3, totalRetJosnBytes)
}
