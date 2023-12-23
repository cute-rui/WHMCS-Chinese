package provider

import (
	"context"
	"fmt"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	"log"
)

var GoogleAPIKey = ``

type Gemini struct {
	model *genai.GenerativeModel
}

func SetupGemini() (*Gemini, error) {
	ctx := context.Background()
	// Access your API key as an environment variable (see "Set up your API key" above)
	client, err := genai.NewClient(ctx, option.WithAPIKey(GoogleAPIKey))
	if err != nil {
		log.Fatal(err)
	}

	model := client.GenerativeModel("gemini-pro")

	return &Gemini{model: model}, nil
}

func (g *Gemini) Translate(str []string) ([]string, error) {
	keyword, strArr := PreProcess(str)

	resp, err := g.model.GenerateContent(context.Background(), GetPrompt(keyword))
	if err != nil {
		return nil, err
	}

	return PostProcess(strArr, fmt.Sprintf("%s", resp.Candidates[0].Content.Parts[len(resp.Candidates[0].Content.Parts)-1])), nil
}

func GetPrompt(str string) genai.Text {
	return genai.Text(`请将下列英文词组翻译为中文，这些词组来自于WHMCS销售系统，不需要其他的回应，直接每行输出一个对应的中文翻译即可，并且请使用“您”代替“你”：` + "\n" + str)
}
