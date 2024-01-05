package provider

import (
	"context"
	"fmt"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	"log"
	"strings"
	"time"
)

var GoogleAPIKey = ``

type Gemini struct {
	model *genai.GenerativeModel

	largeBatch int
}

func SetupGemini(largeBatch int) (*Gemini, error) {
	ctx := context.Background()
	// Access your API key as an environment variable (see "Set up your API key" above)
	client, err := genai.NewClient(ctx, option.WithAPIKey(GoogleAPIKey))
	if err != nil {
		log.Fatal(err)
	}

	model := client.GenerativeModel("gemini-pro")

	return &Gemini{model: model, largeBatch: largeBatch}, nil
}

func (g *Gemini) Translate(str []string, isLarge bool) ([]string, error) {
	//return g.SetPrompt(str, false, []genai.Part{})
	if isLarge {
		var data []string
		for i := 0; i < len(str); i += g.largeBatch {
			var arr []string
			var err error

			if i+g.largeBatch >= len(str) {
				arr, err = g.DirectPipe(str[i:], false, []genai.Part{})
			} else {
				arr, err = g.DirectPipe(str[i:i+g.largeBatch], false, []genai.Part{})
			}

			if err != nil {
				return nil, err
			}

			data = append(data, arr...)
			time.Sleep(15 * time.Second)
			continue
		}

		return data, nil
	}
	return g.DirectPipe(str, false, []genai.Part{})
}

func (g *Gemini) DirectPipe(str []string, retry bool, parts []genai.Part) ([]string, error) {
	content := []genai.Part{}
	if !retry {
		content = append(content, GetRawPHPPrompt(strings.Join(str, "\n")))
	} else {
		content = append(parts, PHPVarCheckRetryPrompt())
	}

	resp, err := g.model.GenerateContent(context.Background(), content...)
	if err != nil {
		return nil, err
	}

	data := fmt.Sprintf("%s", resp.Candidates[0].Content.Parts[len(resp.Candidates[0].Content.Parts)-1])
	dataArr := strings.Split(data, "\n")

	if !PHPVarCheck(str, dataArr) {
		if !retry {
			return g.DirectPipe(str, true, content)
		}
		return nil, fmt.Errorf(`php var check failed`)
	}

	return dataArr, nil
}

func (g *Gemini) PreprocessPipe(str []string, retry bool, parts []genai.Part) ([]string, error) {
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
			return g.PreprocessPipe(str, true, prompt)
		}
		return nil, fmt.Errorf(`get result failed`)
	}

	return PostProcess(strArr, fmt.Sprintf("%s", resp.Candidates[0].Content.Parts[len(resp.Candidates[0].Content.Parts)-1])), nil
}

func GetRawPHPPrompt(str string) genai.Text {
	return genai.Text(`请将下列PHP的字符串翻译为中文，这些代码来自于WHMCS销售系统，**仅需要翻译字符串内的英文，禁止更改变量，需要与原格式相同，以文本输出，禁止用markdown代码格式包裹**，并且请使用“您”代替“你”，行数请保持一致，不要有多余的换行, “client”应翻译为客户：` + "\n" + str)
}

func GetPrompt(str string) genai.Text {
	return genai.Text(`请将下列英文词组翻译为中文，这些词组来自于WHMCS销售系统，不需要其他的回应，直接每行输出一个对应的中文翻译即可，并且请使用“您”代替“你”，行数请保持一致，不要有多余的换行，“/n”无需翻译：` + "\n" + str)
}

func PHPVarCheckRetryPrompt() genai.Text {
	return genai.Text(`PHP变量检查失败，请重新翻译，行数务必对应`)
}

func RetryPrompt() genai.Text {
	return genai.Text(`行数错误，请重新翻译，行数务必对应`)
}
