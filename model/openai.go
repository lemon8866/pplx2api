package model

import (
	"encoding/json"
	"fmt"
	"pplx2api/logger"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ChatCompletionRequest struct {
	Model    string                   `json:"model"`
	Messages []map[string]interface{} `json:"messages"`
	Stream   bool                     `json:"stream"`
	Tools    []map[string]interface{} `json:"tools,omitempty"`
}

// OpenAISrteamResponse 定义 OpenAI 的流式响应结构
type OpenAISrteamResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []StreamChoice `json:"choices"`
}

// Choice 结构表示 OpenAI 返回的单个选项
type StreamChoice struct {
	Index        int         `json:"index"`
	Delta        Delta       `json:"delta"`
	Logprobs     interface{} `json:"logprobs"`
	FinishReason interface{} `json:"finish_reason"`
}

type NoStreamChoice struct {
	Index        int         `json:"index"`
	Message      Message     `json:"message"`
	Logprobs     interface{} `json:"logprobs"`
	FinishReason string      `json:"finish_reason"`
}

// Delta 结构用于存储返回的文本内容
type Delta struct {
	Content string `json:"content"`
}
type Message struct {
	Role       string        `json:"role"`
	Content    string        `json:"content"`
	Refusal    interface{}   `json:"refusal"`
	Annotation []interface{} `json:"annotation"`
}

type OpenAIResponse struct {
	ID      string           `json:"id"`
	Object  string           `json:"object"`
	Created int64            `json:"created"`
	Model   string           `json:"model"`
	Choices []NoStreamChoice `json:"choices"`
	Usage   Usage            `json:"usage"`
}
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func ReturnOpenAIResponse(text string, stream bool, gc *gin.Context) error {
	if stream {
		return streamRespose(text, gc)
	} else {
		return noStreamResponse(text, gc)
	}
}

func estimateTokens(text string) int {
	if text == "" {
		return 0
	}

	tokens := 0
	for _, r := range text {
		if r < 128 {
			tokens += 1
		} else {
			tokens += 2
		}
	}
	return tokens
}

func streamRespose(text string, gc *gin.Context) error {
	openAIResp := &OpenAISrteamResponse{
		ID:      uuid.New().String(),
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   "gemini-3-flash",
		Choices: []StreamChoice{
			{
				Index: 0,
				Delta: Delta{
					Content: text,
				},
				Logprobs:     nil,
				FinishReason: nil,
			},
		},
	}

	jsonBytes, err := json.Marshal(openAIResp)
	jsonBytes = append([]byte("data: "), jsonBytes...)
	jsonBytes = append(jsonBytes, []byte("\n\n")...)
	if err != nil {
		logger.Error(fmt.Sprintf("Error marshalling JSON: %v", err))
		return err
	}

	// 发送数据
	gc.Writer.Write(jsonBytes)
	gc.Writer.Flush()
	return nil
}

func noStreamResponse(text string, gc *gin.Context) error {
	completionTokens := estimateTokens(text)
	openAIResp := &OpenAIResponse{
		ID:      uuid.New().String(),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   "gemini-3-flash",
		Choices: []NoStreamChoice{
			{
				Index: 0,
				Message: Message{
					Role:    "assistant",
					Content: text,
				},
				Logprobs:     nil,
				FinishReason: "stop",
			},
		},
		Usage: Usage{
			PromptTokens:     0,
			CompletionTokens: completionTokens,
			TotalTokens:      completionTokens,
		},
	}

	gc.JSON(200, openAIResp)
	return nil
}
