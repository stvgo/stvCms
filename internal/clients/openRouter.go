package clients

import (
	"context"
	"log/slog"
	"os"

	openrouter "github.com/revrost/go-openrouter"
)

type openRouter struct {
}

func NewOpenRouter() *openRouter {
	return &openRouter{}
}

func (ro *openRouter) Init() (string, error) {
	client := openrouter.NewClient(
		os.Getenv("OPEN_ROUTER_API_KEY"),
		openrouter.WithXTitle("stv CMS"))

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openrouter.ChatCompletionRequest{
			Model: "deepseek/deepseek-chat-v3.1:free",
			Messages: []openrouter.ChatCompletionMessage{
				openrouter.UserMessage("Hello!"),
			},
		},
	)

	if err != nil {
		slog.Error("ChatCompletion error: %v\n", err)
		return "", err
	}

	return resp.Choices[0].Message.Content.Text, nil
}
