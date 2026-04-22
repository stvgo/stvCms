package clients

import (
	"context"
	"log/slog"

	openrouter "github.com/revrost/go-openrouter"
)

type IOpenRouterClient interface {
	GenText(text string) (string, error)
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

func (ro *openRouterClient) GenText(text string) (string, error) {
	resp, err := ro.client.CreateChatCompletion(
		context.Background(),
		openrouter.ChatCompletionRequest{
			Model: "openrouter/elephant-alpha",
			Messages: []openrouter.ChatCompletionMessage{
				openrouter.UserMessage(text),
			},
		},
	)
	if err != nil {
		slog.Error("ChatCompletion error", "error", err)
		return "", err
	}

	return resp.Choices[0].Message.Content.Text, nil
}
