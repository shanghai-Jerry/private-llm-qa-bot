package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"
)

type DPContent struct {
	Content string `json:"content"`
}

func printDPContent() {
	var Content DPContent
	dataBytes, _ := os.ReadFile("./baizhong/dp.json")
	_ = json.Unmarshal(dataBytes, &Content)
	fmt.Printf("%+v\n", Content.Content)
}

func writeContent(fileName string, dataBytes []byte) {
	os.WriteFile(fileName, dataBytes, 0666)
}

func load_data() {
	dataBytes, _ := os.ReadFile("./1.json")
	var m map[string]interface{}
	_ = json.Unmarshal(dataBytes, &m)
	retBytes, _ := json.Marshal(m)
	os.WriteFile("./1.ret.json", retBytes, 0666)
}

// GetValidTokenStr 返回满足eb的最大token数
func GetVaildTokenStr(content string, tokenLimit int, runeLimit int) (string, int, int) {
	totalNum := utf8.RuneCountInString(content)
	totalRunes := []rune(content)
	if totalNum > runeLimit {
		content = string(totalRunes[:runeLimit])
		totalNum = utf8.RuneCountInString(content)
	}
	tokens := CountTokens(content)
	if tokens > tokenLimit {
		return GetVaildTokenStr(content, tokenLimit, totalNum-(tokens-tokenLimit))
	}
	return content, tokens, runeLimit
}

func CountTokens(content string) int {
	totalNum := utf8.RuneCountInString(content)
	englishList := regexp.MustCompile("[a-z|']+").FindAllString(content, -1)
	englishStrNum := len(strings.Join(englishList, ""))
	tokenNum := totalNum - englishStrNum + len(englishList)
	return tokenNum
}

func endsWithPunctuation(str string) bool {
	// 定义正则表达式模式，匹配标点符号结尾
	pattern := `[[:punct:]]$`
	// 编译正则表达式
	re := regexp.MustCompile(pattern)
	// 使用正则表达式匹配
	return re.MatchString(str)
}
