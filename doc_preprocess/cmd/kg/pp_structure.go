package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type OResult struct {
	Type string `json:"type"`
	Res  []Res  `json:"res"`
}

type Res struct {
	Text string `json:"text"`
}

func pp_text2Json(filePath string) {
	fmt.Printf("pp_text2Json ..., %v\n", filePath)
	dataBytes, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	outPath := fmt.Sprintf("%v.json", filePath)
	f, _ := os.Create(outPath)
	// json.Unmarshal(dataBytes, &)
	writeFileBytes(f, dataBytes)
}
