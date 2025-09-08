package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type OpenAIClient struct {
	apiKey string
}

// Ensure OpenAIClient implements LLM
var _ LLM = (*OpenAIClient)(nil)

func NewOpenAIClient() *OpenAIClient {
	apiKey := os.Getenv("OPENAI_API_KEY")
	return &OpenAIClient{apiKey: apiKey}
}

func (o *OpenAIClient) AnalyzeText(input string) (
	string, string, []string, string, []string, float64, error,
) {
	payload := map[string]interface{}{
		"model": "gpt-5-nano",
		"input": fmt.Sprintf(
			"Analyze this text and return JSON with fields: summary, title, topics, sentiment, keywords, confidence.\n\n%s",
			input,
		),
		"store": false,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(context.Background(),
		"POST", "https://api.openai.com/v1/responses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", nil, "", nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		var errMsg map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errMsg)
		return "", "", nil, "", nil, 0,
			fmt.Errorf("error %d: %+v", resp.StatusCode, errMsg)
	}

	var parsed struct {
		Output []struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", "", nil, "", nil, 0, err
	}

	if len(parsed.Output) == 0 || len(parsed.Output[0].Content) == 0 {
		return "", "", nil, "", nil, 0, fmt.Errorf("empty response")
	}

	output := parsed.Output[0].Content[0].Text
	return output, "Generated Title", []string{"ai"}, "neutral", []string{"go"}, 0.9, nil
}
