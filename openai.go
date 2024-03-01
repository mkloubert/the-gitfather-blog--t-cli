// MIT License
//
// Copyright (c) 2024 Marcel Joachim Kloubert (https://marcel.coffee)
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// ChatResponse represents a success response from
// v1/chat/completions API encpoint
// s. https://platform.openai.com/docs/api-reference/chat/create
type ChatResponseV1 struct {
	// list of choices
	Choices []ChatResponseV1Choice `json:"choices"`
}

// ChatResponseV1Choice represents an item of `Choices`
// in ChatResponseV1
type ChatResponseV1Choice struct {
	// the message data in this choice
	Message ChatResponseV1ChoiceMessage `json:"message"`
}

// ChatResponseV1ChoiceMessage represents `Message`
// value in ChatResponseV1Choice
type ChatResponseV1ChoiceMessage struct {
	// the content
	Content string `json:"content"`
	// the role
	Role string `json:"role"`
}

const chatCompletionV1Url = "https://api.openai.com/v1/chat/completions"

// translateWithGPT function translates an input text
// to a target language by using ChatGPT
func translateWithGPT(textToTranslate string, targetLanguage string, context string) (string, error) {
	OPENAI_API_KEY := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	if OPENAI_API_KEY == "" {
		// we need an API key

		return "", fmt.Errorf("missing OPENAI_API_KEY environment variable")
	}

	promptSuffix := strings.TrimSpace(context)
	if promptSuffix != "" {
		promptSuffix = fmt.Sprintf(" (%s)", promptSuffix)
	}

	// setup user message
	userMessage := map[string]interface{}{
		"role": "user",
		"content": fmt.Sprintf(
			`Your only job is to translate the following text to %s language by keeping its format without assumptions%s:%s`,
			targetLanguage,
			promptSuffix,
			textToTranslate,
		),
	}

	// collect all messages
	var messages []interface{}
	messages = append(messages, userMessage)

	// setup request body
	// s. https://platform.openai.com/docs/api-reference/chat/create
	requestBody := map[string]interface{}{
		"model":       "gpt-3.5-turbo-0125", // we are using GPT 3.5, because it is enough for this usecase
		"messages":    messages,
		"temperature": 0, // with this we tell ChatGPT that is not as less assumptions as possible
	}

	// create JSON string from object in requestBody
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	// start the POST request
	request, err := http.NewRequest("POST", chatCompletionV1Url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	// tell API the content type
	// and the key (go to https://platform.openai.com/api-keys)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", OPENAI_API_KEY))

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close() // keep sure response.Body is really clsed at the end

	if response.StatusCode != 200 {
		return "", fmt.Errorf("unexpected status code %v", response.StatusCode)
	}

	// read all data from response
	responseBodyData, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	// parse JSON to ChatResponseV1 object
	var chatResponse ChatResponseV1
	err = json.Unmarshal(responseBodyData, &chatResponse)
	if err != nil {
		return "", err
	}

	// we should have enough data now
	return chatResponse.Choices[0].Message.Content, nil
}
