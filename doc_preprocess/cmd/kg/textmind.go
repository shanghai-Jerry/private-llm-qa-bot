package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// table_ret

type Response struct {
	Result Result `json:"result"`
}
type Result struct {
	RetCode    int          `json:"ret_code"`
	RetMsg     string       `json:"ret_msg"`
	CostTime   int          `json:"cost_time"`
	ResultList []ResultList `json:"result_list"`
}

type ResultList struct {
	ParaNodes        []ParaNode        `json:"para_nodes"`
	FileContentItems []FileContentItem `json:"file_content"`
}

type Position struct {
	PageNo      int   `json:"pageno"`
	LayoutIndex int   `json:"layout_index"`
	Offset      int   `json:"offset"`
	Length      int   `json:"length"`
	Box         []int `json:"box"`
}

type ParaNode struct {
	NodeID   int        `json:"node_id"`
	Text     string     `json:"text"`
	NodeType string     `json:"node_type"`
	Parent   int        `json:"parent"`
	Children []struct{} `json:"children"`
	ParaType string     `json:"para_type"`
	Position []Position `json:"position"`
}

type Rect struct {
	Height int `json:"height"`
	Left   int `json:"left"`
	Top    int `json:"top"`
	Width  int `json:"width"`
}

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type TableCell struct {
	Column      int     `json:"column"`
	Probability int     `json:"probability"`
	Row         int     `json:"row"`
	Vertexes    []Point `json:"vertexes_location"`
	Words       string  `json:"words"`
}

type TableForm struct {
	Body []TableCell `json:"body"`
}

type TableRets struct {
	TableRet []TableRet `json:"table_ret"`
}

type TableRetContent struct {
	PolyLocation []Point `json:"poly_location"`
	Word         string  `json:"word"`
}

type TableRet struct {
	CellLocation TableCell         `json:"cell_location"`
	CellText     string            `json:"cell_text"`
	TLCol        int               `json:"tl_col"`
	TLRow        int               `json:"tl_row"`
	BRCol        int               `json:"br_col"`
	BRRow        int               `json:"br_row"`
	Contents     []TableRetContent `json:"contents"`
}

type TextMindLayout struct {
	Box             []int            `json:"box"`
	Type            string           `json:"type"`
	Text            string           `json:"text"`
	GlobalLayoutIdx int              `json:"global_layout_index"`
	StartIndex      int              `json:"start_index"`
	EndIndex        int              `json:"end_index"`
	Children        []TextMindLayout `json:"children"`
	Matrix          [][]int          `json:"matrix"`
	// CharIndexList   []string `json:"char_index_list"`
	MergeTable string `json:"merge_table"`
	DataIndex  int    `json:"data_index"`
	Source     string `json:"source"`
	NodeID     int    `json:"node_id"`
}

type Char struct {
	Box               []int  `json:"box"`
	Char              string `json:"char"`
	GlobalCharIndex   int    `json:"global_char_index"`
	CharIndexInLayout int    `json:"char_index_in_layout"`
	GlobalLayoutIndex int    `json:"global_layout_index"`
	Color             string `json:"color"`
	Font              string `json:"font"`
	Size              int    `json:"size"`
	Attribute         string `json:"attribute"`
	FullWidth         bool   `json:"full_width"`
}

type PageContent struct {
	Chars  []Char           `json:"chars"`
	Layout []TextMindLayout `json:"layout"`
}

type FileContentItem struct {
	DisplayData      string     `json:"display_data"`
	DisplaySizeRatio float64    `json:"display_size_ratio"`
	ImageRotated     bool       `json:"image_rotated"`
	IsScan           bool       `json:"is_scan"`
	NeedRotate       bool       `json:"need_rotate"`
	OCRContent       OCRContent `json:"ocr_content"`
	Sections         []struct {
		Attribute    string `json:"attribute"`
		PolyLocation Point  `json:"poly_location"`
		SecIdx       struct {
			ColIdx  []int `json:"col_idx"`
			Idx     []int `json:"idx"`
			ParaIdx []int `json:"para_idx"`
			RowIdx  []int `json:"row_idx"`
		} `json:"sec_idx"`
	} `json:"sections"`
	PageContent PageContent `json:"page_content"`
	PageNum     int         `json:"page_num"`
}

type OCRContent struct {
	ErrMsg    string `json:"err_msg"`
	ErrNo     int    `json:"err_no"`
	ImageInfo struct {
		ImageDir int `json:"image_dir"`
		SecCols  int `json:"sec_cols"`
		SecRows  int `json:"sec_rows"`
	} `json:"image_info"`
	LogID     string `json:"logid"`
	QuerySign string `json:"querysign"`
	Ret       []struct {
		Attribute string `json:"attribute"`
		Charset   []struct {
			Rect Rect   `json:"rect"`
			Word string `json:"word"`
		} `json:"charset"`
		Iou          int    `json:"iou"`
		OriginIdx    int    `json:"origin_idx"`
		PolyLocation Point  `json:"poly_location"`
		Rect         Rect   `json:"rect"`
		Word         string `json:"word"`
	} `json:"ret"`
	TablesRet []TableRets `json:"tables_ret"`
}

func getTextMindCellContent(contents []TableRetContent) string {
	var words []string
	for _, c := range contents {
		words = append(words, c.Word)
	}
	return strings.Join(words, " ")
}

func buildTextMindTable(data []TableRet) [][]string {
	// 构建表格
	var numRows, numCols int
	for _, cell := range data {
		if cell.BRRow > numRows {
			numRows = cell.BRRow
		}
		if cell.BRCol > numCols {
			numCols = cell.BRCol
		}
	}
	table := make([][]string, numRows)
	for i := range table {
		table[i] = make([]string, numCols)
	}
	// 合并的单元格，分开来填充
	for _, cell := range data {
		for i := cell.TLRow; i < cell.BRRow; i++ {
			for j := cell.TLCol; j < cell.BRCol; j++ {
				var v string
				if len(cell.Contents) == 0 {
					v = cell.CellText
				} else {
					v = getTextMindCellContent(cell.Contents)
				}
				table[i][j] = v
			}
		}
	}
	return table
}

func buildTextMindTableWithMatrix(data TextMindLayout) [][]string {

	matrix := data.Matrix
	childrens := data.Children
	rows := len(matrix)
	cols := len(matrix[0])
	// 构建表格
	table := make([][]string, rows)
	for i := range table {
		table[i] = make([]string, cols)
	}
	for i, rows := range matrix {
		cm := make(map[int]struct{})
		for j, c := range rows {
			if _, ok := cm[c]; !ok {
				// table[i][j] = childrens[c].Text
				cm[c] = struct{}{}
			} else {
				// table[i][j] = "-"
			}
			table[i][j] = childrens[c].Text
		}
	}
	return table
}

func parseTextMindTable(fileName string) {
	if len(fileName) == 0 {
		return
	}
	var resp Response
	rbytes, _ := os.ReadFile(fileName)
	json.Unmarshal(rbytes, &resp)
	resultList := resp.Result.ResultList[0]
	pageTableIndex := make(map[int]int)
	for _, paraNode := range resultList.ParaNodes {
		nodeType := paraNode.NodeType
		nodeIndex := paraNode.NodeID
		switch nodeType {
		case layout_title:
			fmt.Printf("textmind:%v, %v \n", layout_title, paraNode.Text)
		case layout_table:
			positions := paraNode.Position
			for _, position := range positions {
				layOutIndex := position.LayoutIndex
				pageNo := position.PageNo
				fileContent := resultList.FileContentItems[pageNo]
				pageContent := fileContent.PageContent
				// ocrContent := fileContent.OCRContent
				tableIndex := pageTableIndex[pageNo]
				// newTables := buildTextMindTable(ocrContent.TablesRet[tableIndex].TableRet)
				newTables := buildTextMindTableWithMatrix(pageContent.Layout[layOutIndex])
				pageTableIndex[pageNo]++
				f, _ := os.Create(fmt.Sprintf("%v-n%v-p%v-table-ret-i%v", fileName, nodeIndex, pageNo, tableIndex))
				writeTable(f, newTables)
				layout := pageContent.Layout[layOutIndex]
				matrix := layout.Matrix
				table := make([][]string, len(matrix))
				childrens := layout.Children
				for i, row := range matrix {
					for _, col := range row {
						table[i] = append(table[i], childrens[col].Text)
					}
				}
				// f1, _ := os.Create(fmt.Sprintf("%v-n%v-p%v-layout-i%v", fileName, nodeIndex, pageNo, index))
				// writeTable(f1, table)
			}
		}
	}
}
