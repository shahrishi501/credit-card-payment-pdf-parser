package gemini

import (
	"context"
	"fmt"
	"log"
	"os"

	"google.golang.org/genai"
)

func AnalyzePDFWithGemini(text, prompt string) string {
    ctx := context.Background()
    client, err := genai.NewClient(ctx, &genai.ClientConfig{
        APIKey:  os.Getenv("GEMINI_API_KEY"),
        Backend: genai.BackendGeminiAPI,
    })
    if err != nil {
        log.Fatalf("Error creating Gemini client: %v", err)
    }

    parts := []*genai.Part{
        genai.NewPartFromText(prompt + "\n\n" + text),
    }

    contents := []*genai.Content{
        genai.NewContentFromParts(parts, genai.RoleUser),
    }

    result, err := client.Models.GenerateContent(ctx, "gemini-2.5-flash", contents, nil)
    if err != nil {
        log.Fatalf("Error from Gemini: %v", err)
    }

    fmt.Println("\nGemini Output:\n", result.Text())
    return result.Text()
}
