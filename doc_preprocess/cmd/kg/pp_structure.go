package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
)

type OResultJson struct {
	Results []OResult `json:"results"`
}

type OResult struct {
	Type string `json:"type"`
	BBox []int  `json:"bbox"`
	Res  []Res  `json:"res"`
}

type OTResult struct {
	Type    string `json:"type"`
	BBox    []int  `json:"bbox"`
	Res     TRes   `json:"res"`
	ImagIDX int    `json:"img_idx"`
}

type TRes struct {
	CellBbox [][]float32 `json:"cell_bbox"`
	HTML     string      `json:"html"`
}

type Res struct {
	Text       string      `json:"text"`
	Confidence float32     `json:"confidence"`
	TextRegion [][]float32 `json:"text_region"`
}

func compareBBox(bbox1, bbox2 []int) bool {
	// 假设 bbox1 和 bbox2 的格式为 [x1, y1, x2, y2]
	_, y1_1, _, _ := bbox1[0], bbox1[1], bbox1[2], bbox1[3]
	_, y1_2, _, _ := bbox2[0], bbox2[1], bbox2[2], bbox2[3]

	if y1_1 < y1_2 {
		return true
	}
	return false
}

func getMinRightX(results []OResult) int {
	maxX := math.MaxInt
	for _, r := range results {
		ty := r.Type
		if ty == layout_header || ty == layout_figure_caption || ty == layout_title || ty == layout_figure || ty == layout_equation {
			continue
		}
		if len(r.Res) == 0 {
			continue
		}
		if r.BBox[2] < maxX {
			maxX = r.BBox[2]
		}
	}
	return maxX
}

func order_adjuest(results []OResult) []OResult {
	var leftResults []OResult
	var rightResults []OResult
	minX := getMinRightX(results)
	fmt.Printf("minRightX:%v \n", minX)
	for _, result := range results {
		if result.BBox[0] < minX {
			leftResults = append(leftResults, result)
		} else {
			rightResults = append(rightResults, result)
		}
	}
	fmt.Printf("total:%v, left:%v, right:%v\n", len(results), len(leftResults), len(rightResults))
	sort.Slice(leftResults, func(i, j int) bool {
		return compareBBox(leftResults[i].BBox, leftResults[j].BBox)
	})

	sort.Slice(rightResults, func(i, j int) bool {
		return compareBBox(rightResults[i].BBox, rightResults[j].BBox)
	})

	mergeResults := append(leftResults, rightResults...)
	return mergeResults
}

func pp_text2Json(f3 *os.File, filePath string) {
	fmt.Printf("pp_text2Json ..., file: %v\n", filePath)
	lastIndexOfDot := strings.LastIndex(filePath, ".")
	lastIndexOfDash := strings.LastIndex(filePath, "_")
	pageIndx := filePath[lastIndexOfDash:lastIndexOfDot]
	f, _ := os.Open(filePath)
	// 创建一个 Scanner 对象
	scanner := bufio.NewScanner(f)

	jsonResult := OResultJson{}
	var typesM = make(map[string]struct{})
	// 循环读取每一行
	for scanner.Scan() {
		line := scanner.Text()
		var result OResult
		_ = json.Unmarshal([]byte(line), &result)
		if _, ok := typesM[result.Type]; !ok {
			typesM[result.Type] = struct{}{}
		}
		if result.Type == layout_table {
			var tmpR OTResult
			_ = json.Unmarshal([]byte(line), &tmpR)
			result.Res = append(result.Res, Res{Text: tmpR.Res.HTML})
		}
		jsonResult.Results = append(jsonResult.Results, result)
	}
	var types []string
	for k := range typesM {
		types = append(types, k)
	}
	fmt.Printf("types:%v\n", types)
	// 检查是否发生错误
	if err := scanner.Err(); err != nil {
		fmt.Println("读取文件时发生错误:", err)
		return
	}
	// 进行顺序调整
	results := jsonResult.Results

	var headers []OResult
	var others []OResult
	// 先过滤出所有header和其他的
	for _, result := range results {
		if result.Type == layout_header {
			headers = append(headers, result)
		} else {
			others = append(others, result)
		}
	}
	//
	headers = order_adjuest(headers)
	others = order_adjuest(others)

	jsonResult.Results = append(headers, others...)

	dataBytes, _ := json.Marshal(jsonResult)
	outPath := fmt.Sprintf("%v.json", filePath)
	f2, _ := os.Create(outPath)
	var outString []string
	for _, r := range jsonResult.Results {
		if r.Type == layout_footer {
			continue
		}
		if r.Type == layout_figure || r.Type == layout_equation {
			var tmp []string
			for _, res := range r.Res {
				tmp = append(tmp, res.Text)
			}
			outString = append(outString, fmt.Sprintf("[%v] start == ", r.Type))
			outString = append(outString, strings.Join(tmp, " "))
			outString = append(outString, fmt.Sprintf("[%v] end == ", r.Type))

			continue
		}
		if r.Type == layout_table {
			tableFileName := fmt.Sprintf("[%v, %v, %v, %v]_%v.xlsx", r.BBox[0], r.BBox[1], r.BBox[2], r.BBox[3], pageIndx)
			outString = append(outString, fmt.Sprintf("[%v] %v", r.Type, tableFileName))
			continue
		}
		var tmp []string
		if r.Type == layout_title {
			tmp = append(tmp, "[title]")
			for _, res := range r.Res {
				tmp = append(tmp, res.Text)
			}
			outString = append(outString, strings.Join(tmp, " ")+"\n")
			continue
		}
		for _, res := range r.Res {
			tmp = append(tmp, res.Text)
		}
		outString = append(outString, strings.Join(tmp, "\n"))
	}
	writeFileBytes(f2, dataBytes)
	writeFile(f3, outString)
}
