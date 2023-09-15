package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var OCR_API_KEY = ""
var OCR_SECRET_KEY = ""

func parsePDF(pdfFile string, fileNum int) (body []byte, err error) {
	urlPath := "https://aip.baidubce.com/rest/2.0/ocr/v1/doc_analysis_office?access_token=" + GetAccessToken()
	// pdf_file 可以通过 GetFileContentAsBase64("C:\fakepath\23.pdf") 方法获取
	data := url.Values{}
	data.Set("pdf_file", pdfFile)
	data.Set("pdf_file_num", fmt.Sprintf("%v", fileNum))
	data.Set("layout_analysis", "true")
	data.Set("recg_tables", "true")

	client := &http.Client{}
	req, err := http.NewRequest("POST", urlPath, strings.NewReader(data.Encode()))
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
	body, err = io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

/**
 * 获取文件base64编码
 * @param string  path 文件路径
 * @return string base64编码信息，不带文件头
 */
func GetFileContentAsBase64(path string) string {
	srcByte, err := os.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return base64.StdEncoding.EncodeToString(srcByte)
}

/**
 * 使用 AK，SK 生成鉴权签名（Access Token）
 * @return string 鉴权签名信息（Access Token）
 */
func GetAccessToken() string {
	// ocrAPIKey := os.Getenv("OCR_API_KEY")
	// ocrSecretKey := os.Getenv("OCR_SECRET_KEY")
	keyBytes, _ := os.ReadFile("./api_key.txt")
	OCR_API_KEY = string(keyBytes)
	valuBytes, _ := os.ReadFile("./api_secret.txt")
	OCR_SECRET_KEY = string(valuBytes)
	url := fmt.Sprintf("https://aip.baidubce.com/oauth/2.0/token?client_id=%v&client_secret=%v&grant_type=client_credentials", OCR_API_KEY, OCR_SECRET_KEY)
	payload := strings.NewReader(``)
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	accessTokenObj := map[string]string{}
	json.Unmarshal([]byte(body), &accessTokenObj)
	return accessTokenObj["access_token"]
}
