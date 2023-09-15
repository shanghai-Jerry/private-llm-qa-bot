package main

import (
	"io/ioutil"
	"log"
	"time"

	"github.com/imroc/req/v3"
)

var d2pURL = "http://10.216.187.19:8725/d2p"

type D2PResponse struct {
	Tables  map[string]Table `json:"tables"`
	DocID   string           `json:"docid"`
	Paras   []Para           `json:"paras"`
	Status  int              `json:"status"`
	LogID   string           `json:"logid"`
	ImgData []ImageData      `json:"img_data"`
}

type Table struct {
	TableContent     [][]string `json:"table_content"`
	TableHeader      [][]string `json:"table_header"`
	TableHTMLContent string     `json:"table_htmlContent"`
	TableIndex       int        `json:"table_index"`
	TableName        string     `json:"table_name"`
	TableID          string     `json:"table_id"`
}

type Para struct {
	Title            string         `json:"title"`
	Code             string         `json:"code"`
	SubtitleStr      []string       `json:"subtitle_str"`
	FatherTitlesText string         `json:"father_titles_text"`
	ParaType         string         `json:"para_type"`
	TableContent     []TableContent `json:"table_content"`
	Para             string         `json:"para"`
	CoreTitle        string         `json:"core_title"`
	ParaID           string         `json:"paraid"`
}

type TableContent struct {
}

type ImageData struct {
	Data string `json:"data"`
}

func d2p(result []byte) []byte {
	resp, err := req.C().SetTimeout(100 * time.Minute).R().SetBodyBytes(result).Post(d2pURL)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return body
}
