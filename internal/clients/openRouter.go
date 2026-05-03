package clients

import (
	"context"
	"log/slog"

	openrouter "github.com/revrost/go-openrouter"
)

type IOpenRouterClient interface {
	GenAI(systemPrompt, userPrompt string) (string, error)
}

type openRouterClient struct {
	client *openrouter.Client
}

func NewOpenRouter(apiKey string) IOpenRouterClient {
	client := openrouter.NewClient(
		apiKey,
		openrouter.WithXTitle("stv CMS"),
	)
	return &openRouterClient{client: client}
}

func (ro *openRouterClient) GenAI(systemPrompt, userPrompt string) (string, error) {
	messages := []openrouter.ChatCompletionMessage{}

	if systemPrompt != "" {
		messages = append(messages, openrouter.SystemMessage(systemPrompt))
	}

	messages = append(messages, openrouter.UserMessage(userPrompt))

	resp, err := ro.client.CreateChatCompletion(
		context.Background(),
		openrouter.ChatCompletionRequest{
			Model:    "nvidia/nemotron-3-super-120b-a12b:free",
			Messages: messages,
		},
	)
	if err != nil {
		slog.Error("ChatCompletion error", "error", err)
		return "", err
	}

	return resp.Choices[0].Message.Content.Text, nil
}
