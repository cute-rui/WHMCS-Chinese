package provider

import (
	"errors"
	"strings"
)

var ProviderType = `gemini`

type Provider interface {
	Translate(str []string) ([]string, error)
}

func SetupProvider() (Provider, error) {
	switch ProviderType {
	case `tencent`:
		return SetupTencent()
	case `gemini`:
		return SetupGemini()
	}

	return nil, errors.New(`provider not found`)
}

func PreProcess(str []string) (string, map[int][]string) {
	ret := map[int][]string{}
	keywords := []string{}
	for i := range str {
		arr := strings.Split(str[i], `"`)
		keywords = append(keywords, arr[1])
		ret[i] = arr
	}

	return strings.Join(keywords, "\n"), ret
}

func PostProcess(raw map[int][]string, data string) []string {
	dataArr := strings.Split(data, "\n")
	ret := []string{}
	for i := range dataArr {
		ret = append(ret, raw[i][0]+`"`+dataArr[i]+`"`+raw[i][2])
	}

	return ret
}
