package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"
)

// 文档解析输出格式
type OutputJsonFormat struct {
	DocTitles  []string      `json:"doc_title"`
	OutputJson []*OutputJson `json:"output_json"`
}

type OutputJson struct {
	Tokens  int    `json:"tokens"`
	Content string `json:"content"`
	// table, text, title, content
	Type string `json:"type"`
	// 第几页
	Pages int `json:"page"`
}

// 分段输出 格式
type ParaJson struct {
	DocTitle string   `json:"doc_title"`
	Titles   []string `json:"titles"`
	Para     string   `json:"paragraph"`
	// table, text, title, content
	Type   string `json:"type"`
	Tokens int64  `json:"tokens"`
	// 内容所属的页信息
	Pages []int `json:"pages"`
}

func get_paras_from_format_json(path string, outdir string) {
	fmt.Printf("start format json parsing... %v\n", path)
	index := strings.LastIndex(path, "/")
	// 23.pdf
	inputFile := path[index+1:]
	index = strings.LastIndex(inputFile, ".")
	// 23
	inputFileName := strings.ReplaceAll(inputFile[0:index], " ", "_")
	outFileJsonPathFunc := func(seq int) string {
		return fmt.Sprintf("%s/%v-seq-%v.txt", outdir, inputFileName, seq)
	}
	var outPutJsonFromat OutputJsonFormat
	var paraJsons []*ParaJson
	pathBytes, _ := os.ReadFile(path)
	_ = json.Unmarshal(pathBytes, &outPutJsonFromat)

	docTitles := strings.Join(outPutJsonFromat.DocTitles, "\n")
	docTokens := CountTokens(docTitles)
	currentTokens := 0
	var contentParas []string
	for _, json := range outPutJsonFromat.OutputJson {
		// var titles []string
		ot := json.Type
		tokens := json.Tokens
		content := json.Content
		originContent := content
		adds := currentTokens + tokens
		if adds <= 384 {
			contentParas = append(contentParas, content)
			currentTokens += tokens
		} else {
			// 是否为表格或者content
			if ot == table || ot == contents {
				if len(contentParas) > 0 {
					paraJson := &ParaJson{
						DocTitle: docTitles,
						Para:     strings.Join(contentParas, "\n"),
						Type:     text,
						Tokens:   int64(currentTokens),
					}
					paraJsons = append(paraJsons, paraJson)
					contentParas = []string{}
				}
				// 添加table和contents
				// TODO(chaojiang)
				continue
			}
			end := utf8.RuneCountInString(content)
			partContent, tokens, index := GetVaildTokenStr(content, 384-currentTokens, end)
			contentParas = append(contentParas, partContent)
			paraJson := &ParaJson{
				DocTitle: docTitles,
				// Titles:   titles,
				Para:   strings.Join(contentParas, "\n"),
				Type:   text,
				Tokens: int64(currentTokens + tokens),
			}
			contentParas = []string{}
			paraJsons = append(paraJsons, paraJson)
			// 在原始内容中的偏移位置
			offset := index
			for {
				content = string([]rune(originContent)[index:])
				end = utf8.RuneCountInString(content)
				limitToken := 384 - docTokens
				partContent, rtokens, retIndex := GetVaildTokenStr(content, limitToken, end)
				if rtokens < limitToken {
					currentTokens = rtokens
					contentParas = append(contentParas, partContent)
					break
				} else {
					paraJson := &ParaJson{
						DocTitle: docTitles,
						// Titles:   titles,
						Para:   partContent,
						Type:   text,
						Tokens: int64(rtokens),
					}
					offset += retIndex
					paraJsons = append(paraJsons, paraJson)
					if rtokens == limitToken {
						break
					}
				}
			}
		}
	}
	if len(contentParas) > 0 {
		paraJson := &ParaJson{
			DocTitle: docTitles,
			Para:     strings.Join(contentParas, "\n"),
			Type:     text,
			Tokens:   int64(currentTokens),
		}
		paraJsons = append(paraJsons, paraJson)

	}
	for index, pjson := range paraJsons {
		fmt.Printf("index:%v, token:%v\n", index, pjson.Tokens)
		f, _ := os.Create(outFileJsonPathFunc(index))
		writeFileBytes(f, []byte(pjson.Para))
	}
}
