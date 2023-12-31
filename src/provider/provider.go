package provider

import (
	"errors"
	"log"
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

		arr := splitBySingleQuote(str[i])

		if len(arr) > 3 {
			str[i] = tryMatchSingleQuote(str[i])
			arr = splitBySingleQuote(str[i])
		}

		if len(arr) != 3 {
			log.Println(`invalid data`)
			continue
		}

		arr[1] = strings.ReplaceAll(arr[1], `REPLACEHOLDERFORTEMPORARYMARK`, `\"`)
		keywords = append(keywords, arr[1])
		ret[i] = arr
	}

	return strings.Join(keywords, "\n"), ret
}

func PostProcess(raw map[int][]string, data string) []string {
	dataArr := strings.Split(data, "\n")
	ret := []string{}
	for i := range dataArr {
		if len(raw[i]) != 3 {
			log.Println(`invalid data:`, i, raw[i], raw, data)
			continue
		}
		ret = append(ret, raw[i][0]+`"`+dataArr[i]+`"`+raw[i][2])
	}

	return ret
}

func PHPVarJoiner(string) {

}

func splitBySingleQuote(toProc string) []string {
	toProc = strings.ReplaceAll(toProc, `\"`, `REPLACEHOLDERFORTEMPORARYMARK`)

	return strings.Split(toProc, `"`)
}

func tryMatchSingleQuote(toProc string) string {
	splited := strings.Split(toProc, `=`)
	if len(splited) != 2 {
		return toProc
	}

	splited[1] = strings.ReplaceAll(splited[1], `'`, `"`)

	return strings.Join(splited, `=`)
}
