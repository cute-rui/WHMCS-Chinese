package provider

import (
	"errors"
	"log"
	"regexp"
	"strings"
)

var ProviderType = `gemini`

type Provider interface {
	Translate(str []string, isLarge bool) ([]string, error)
}

func SetupProvider(largeBatch int) (Provider, error) {
	switch ProviderType {
	case `tencent`:
		return SetupTencent(largeBatch)
	case `gemini`:
		return SetupGemini(largeBatch)
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
			log.Println(`invalid data`, str[i])
			continue
		}

		arr[1] = strings.ReplaceAll(arr[1], `REPLACEHOLDERFORTEMPORARYMARK`, `\"`)
		arr[1] = ReplaceAllReturn(arr[1])
		keywords = append(keywords, arr[1])
		ret[i] = arr
	}

	return strings.Join(keywords, "\n"), ret
}

func PreCheckResult(raw map[int][]string, data string) bool {
	dataArr := strings.Split(data, "\n")
	if len(raw) != len(dataArr) {
		log.Println(`precheck failed`)
		for i := range dataArr {
			log.Println(raw[i][1], dataArr[i])
		}
		PostProcess(raw, data)
		return false
	}

	return true
}

func PostProcess(raw map[int][]string, data string) []string {
	dataArr := strings.Split(data, "\n")
	ret := []string{}
	for i := range dataArr {
		if len(raw[i]) != 3 {
			log.Println(`invalid data:`, i, raw[i], raw, data[i], data)
			log.Println(`on:`, raw[i], `=>`, data[i])
			log.Println(raw, dataArr)
			continue
		}
		dataArr[i] = ReplaceReturnBack(dataArr[i])
		ret = append(ret, raw[i][0]+`"`+dataArr[i]+`"`+raw[i][2])
	}

	return ret
}

func PHPVarJoiner(string) {

}

func PHPVarCheck(str []string, dataArr []string) int {
	if len(dataArr) != len(str) {
		maxLen := len(dataArr)
		if len(str) > len(dataArr) {
			maxLen = len(str)
		}

		for i := 0; i < maxLen; i++ {
			s := ``
			if i < len(dataArr) {
				s += dataArr[i]
			}
			s += `	`
			if i < len(str) {
				s += str[i]
			}
			log.Println(s)
		}

		return LENGTH_NOT_MATCH // len not match
	}

	for i := range str {
		if !PHPSemicolonCheck(str[i]) {
			return SEMICOLON_NOT_MATCH // semicolon not match
		}

		prefix := strings.Split(str[i], `=`)[0]
		if !strings.HasPrefix(dataArr[i], prefix) {
			log.Println(`prefix check failed`, str[i], dataArr[i])
			return PREFIX_NOT_MATCH // prefix not match
		}
	}

	return ALL_MATCH // all match
}

func PHPSemicolonCheck(data string) bool {
	if !(strings.HasSuffix(removeWhitespace(data), `;`)) {
		log.Println(`has no semicolon`, data)
		return false
	}
	return true
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

func ReplaceAllReturn(str string) string {
	return strings.ReplaceAll(str, "\n", "/n")
}

func ReplaceReturnBack(str string) string {
	return strings.ReplaceAll(str, "/n", "\n")
}

func removeWhitespace(str string) string {
	r, _ := regexp.Compile(`\s+`)
	return r.ReplaceAllString(str, ``)
}

func removeComment(str []string) []string {
	ret := []string{}
	for i := range str {
		temp := strings.ReplaceAll(str[i], `://`, `SCHEMA`)
		if strings.Contains(temp, `//`) {
			log.Println(`contains comment`, str[i])
			temp = strings.Split(temp, `//`)[0]
		}
		ret = append(ret, strings.ReplaceAll(temp, `SCHEMA`, `://`))
	}

	return ret
}
