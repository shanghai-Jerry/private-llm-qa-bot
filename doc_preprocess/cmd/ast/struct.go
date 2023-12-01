package main

import "encoding/json"

type LogData struct {
	RequestID   string   `json:"requestid"`
	IP          string   `json:"ip"`
	Timestamp   int64    `json:"timestamp"`
	AppID       int64    `json:"appid"`
	AccountID   int64    `json:"account_id"`
	AccountType string   `json:"account_type"`
	URL         string   `json:"url"`
	Request     Request  `json:"request"`
	Response    Response `json:"response"`
}

func (data *LogData) Unmarshal(bytes []byte) error {
	return json.Unmarshal(bytes, data)
}

func (data *LogData) GetAccountID() int64 {
	return data.AccountID
}
func (data *LogData) GetAPIURL() string {
	return data.URL
}

type Request struct {
	// Header interface{} `json:"header"`
	Body Body `json:"body"`
}

type Body struct {
	EnableCitation bool      `json:"enable_citation"`
	EnableTrace    bool      `json:"enable_trace"`
	EnableTTS      bool      `json:"enable_tts"`
	EnableVilg     bool      `json:"enable_vilg"`
	Messages       []Message `json:"messages"`
	Source         int       `json:"source"`
	Stream         bool      `json:"stream"`
	TopP           float64   `json:"top_p"`
	UserID         string    `json:"user_id"`
}

type Message struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

type Response struct {
	Status int          `json:"status"`
	Header interface{}  `json:"header"`
	Body   BodyResponse `json:"body"`
}

type BodyResponse struct {
	ID               string `json:"id"`
	Object           string `json:"object"`
	Created          int    `json:"created"`
	SentenceID       int    `json:"sentence_id"`
	Result           string `json:"result"`
	NeedClearHistory bool   `json:"need_clear_history"`
	Usage            Usage  `json:"usage"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
