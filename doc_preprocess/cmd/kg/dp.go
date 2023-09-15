package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/imroc/req/v3"
)

type DPRequest struct {
	Title              string `json:"title"`
	Analyze_type       int    `json:"analyze_type"`
	NeedLayoutAnalysis bool   `json:"need_layout_analysis"`
	Format             string `json:"format"`
	Content            string `json:"content"`
}

var dpURL = "http://10.216.187.19:8441/dp"

func dp(fileName, format string) ([]byte, error) {
	client := req.C().SetTimeout(100 * time.Minute)
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	dataBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	request := &DPRequest{
		Title:              "123",
		Analyze_type:       1,
		NeedLayoutAnalysis: true,
		Format:             format,
	}
	request.Content = base64.StdEncoding.EncodeToString(dataBytes)
	resp, err := client.R().
		SetBody(request).
		Post(dpURL)
	if err != nil {
		log.Fatal(err)
	}

	if !resp.IsSuccessState() {
		fmt.Println("bad response status:", resp.Status)
		return nil, errors.New("bad response status")
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	writeContent(fileName+"-dp.json", respBytes)
	return parseDPResponse(respBytes)
}

func parseDPResponse(respBytes []byte) ([]byte, error) {
	var resp map[string]interface{}
	err := json.Unmarshal([]byte(respBytes), &resp)
	if err != nil {
		fmt.Println("JSON 解析错误:", err)
		return nil, errors.New("json.Unmarshal error")
	}
	resp["img_data"] = nil
	fmt.Println("response:", string(respBytes))
	result := resp["result"].(map[string]interface{})
	resultBytes, err := json.Marshal(result)
	ioutil.WriteFile("./dp.json", []byte(resultBytes), 0666)
	if err != nil {
		return nil, err
	}
	return resultBytes, nil
}
