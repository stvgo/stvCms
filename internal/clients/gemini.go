package clients

import (
	"context"
	"log/slog"

	"google.golang.org/genai"
)

type geminiGenAI struct {
}

func NewGeminiGenAI() geminiGenAI {
	return geminiGenAI{}
}

func (g geminiGenAI) GenTextAI(text string) (string, error) {
	ctx := context.Background()
	// The client gets the API key from the environment variable `GEMINI_API_KEY`.
	// TODO: generate Gemini api key
	client, err := genai.NewClient(ctx, nil)

	if err != nil {
		slog.Error(err.Error())
		return "", err
	}

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-3-flash-preview",
		genai.Text(text),
		nil)

	return result.Text(), nil
}
