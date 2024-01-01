package provider

import (
	"context"
	"fmt"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	"log"
	"strings"
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
	//return g.SetPrompt(str, false, []genai.Part{})

	resp, err := g.model.GenerateContent(context.Background(), GetRawPHPPrompt(strings.Join(str, "\n")))
	if err != nil {
		return nil, err
	}

	data := fmt.Sprintf("%s", resp.Candidates[0].Content.Parts[len(resp.Candidates[0].Content.Parts)-1])
	dataArr := strings.Split(data, "\n")

	if !PHPVarCheck(str, dataArr) {
		return nil, fmt.Errorf(`php var check failed`)
	}

	return dataArr, nil
}

func (g *Gemini) SetPrompt(str []string, retry bool, parts []genai.Part) ([]string, error) {
	keyword, strArr := PreProcess(str)

	prompt := []genai.Part{}
	if !retry {
		prompt = append(prompt, GetPrompt(keyword))
	} else {
		prompt = append(parts, RetryPrompt())
	}

	resp, err := g.model.GenerateContent(context.Background(), prompt...)
	if err != nil {
		return nil, err
	}

	if !PreCheckResult(strArr, fmt.Sprintf("%s", resp.Candidates[0].Content.Parts[len(resp.Candidates[0].Content.Parts)-1])) {
		if !retry {
			return g.SetPrompt(str, true, prompt)
		}
		return nil, fmt.Errorf(`get result failed`)
	}

	return PostProcess(strArr, fmt.Sprintf("%s", resp.Candidates[0].Content.Parts[len(resp.Candidates[0].Content.Parts)-1])), nil
}

func GetRawPHPPrompt(str string) genai.Text {
	return genai.Text(`请将下列PHP的字符串翻译为中文，这些代码来自于WHMCS销售系统，**仅需要翻译字符串内的英文，禁止更改变量，需要与原格式相同**，并且请使用“您”代替“你”，行数请保持一致，不要有多余的换行, “client”应翻译为客户：` + "\n" + str)
}

func GetPrompt(str string) genai.Text {
	return genai.Text(`请将下列英文词组翻译为中文，这些词组来自于WHMCS销售系统，不需要其他的回应，直接每行输出一个对应的中文翻译即可，并且请使用“您”代替“你”，行数请保持一致，不要有多余的换行，“/n”无需翻译：` + "\n" + str)
}

func RetryPrompt() genai.Text {
	return genai.Text(`行数错误，请重新翻译，行数务必对应`)
}
