package main

import (
	"WHMCS-Chinese/src/provider"
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"
)

const maxRoutine = 5
const maxRetry = 5
const batch = 32
const adminFile = `admin.php`
const langFile = `lang.php`

var lang = map[int]string{}
var lineOffset = 1000000
var p provider.Provider
var resultMap = sync.Map{}

func main() {
	flag.StringVar(&provider.TencentID, "id", "", "")
	flag.StringVar(&provider.TencentSecret, "secret", "", "")
	flag.StringVar(&provider.ProviderType, "provider", "", "")
	flag.StringVar(&provider.GoogleAPIKey, "google", "", "")

	flag.Parse()

	srv, err := provider.SetupProvider()
	if err != nil {
		return
	}

	p = srv

	GetDiff()
	translate()
	CopyLatestFile()
	OverwriteLatestFile()
}

func translate() {
	currentIndex := 0
	keys := SortMap(lang)
	indexLock := sync.Mutex{}
	var WG sync.WaitGroup
	for i := 0; i < maxRoutine; i++ {
		WG.Add(1)
		go func() {
			for currentIndex < len(lang)-1 {
				num := batch
				indexLock.Lock()
				keyid := currentIndex
				if final := len(lang) - 1 - currentIndex; final < batch {
					currentIndex += final
					num = final
				} else {
					currentIndex += batch
				}
				indexLock.Unlock()
				if num == 0 {
					break
				}

				var arr []string
				for j := 0; j < num; j++ {
					arr = append(arr, lang[keys[keyid+j]])
				}

				data, err := p.Translate(arr)
				if err != nil {
					log.Println(err.Error())
				}

				for j := range data {
					resultMap.Store(keys[keyid+j], data[j])
				}
				time.Sleep(15 * time.Second)
			}
			WG.Done()
		}()
	}
	WG.Wait()

}

func readFileToStringMap(path string, isAdmin bool) (map[int]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	ret := map[int]string{}

	index := 0
	if isAdmin {
		index = lineOffset
	}

	for scanner.Scan() {
		index++
		raw := scanner.Text()
		if !strings.HasPrefix(raw, `$_`) {
			continue
		}

		ret[index] = raw
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ret, nil
}

func getBeforeLatest(path string) (string, error, bool) {
	list, err := os.ReadDir(path)
	if err != nil {
		return ``, err, false
	}

	var ret string
	for i := range list {
		if !list[i].IsDir() {
			continue
		}

		if list[i].Name() > ret && list[i].Name() != `latest` {
			ret = list[i].Name()
		}
	}
	if ret == `` {
		return ``, nil, false
	}

	return ret, nil, true
}

func CopyLatestFile() error {
	beforeLatest, err, exist := getBeforeLatest("./archives")
	if !exist {
		os.Create("./archives/latest/" + adminFile)
		os.Create("./archives/latest/" + langFile)
		return nil
	}

	err = exec.Command("rm", "-rf", "./archives/latest").Run()
	if err != nil {
		return err
	}

	return exec.Command("cp", "-r", "./archives/"+beforeLatest, "./archives/latest").Run()
}

func OverwriteLatestFile() error {
	lMap := map[int]string{}
	aMap := map[int]string{}
	for index := range lang {
		var data string
		if v, ok := resultMap.Load(index); ok {
			data = v.(string)
		} else {
			continue
		}

		if index < lineOffset {
			lMap[index] = data
			continue
		}

		aMap[index-lineOffset] = data
	}

	overwrite(lMap, "./archives/latest/"+langFile)
	overwrite(aMap, "./archives/latest/"+adminFile)

	return nil
}

func overwrite(data map[int]string, path string) error {
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	keys := SortMap(data)
	for _, index := range keys {

		for len(lines) <= index {
			lines = append(lines, "")
		}

		lines[index] = data[index]
	}

	// Write the modified content back to the file
	file.Seek(0, 0)  // Move the cursor to the beginning of the file
	file.Truncate(0) // Truncate the file to clear its content

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(writer, line)
	}
	writer.Flush()
	return nil
}

func SortMap(data map[int]string) []int {
	var keys []int
	for k := range data {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	return keys
}

func GetDiff() error {
	latestAdmin, err := readFileToStringMap("./eng/latest/"+adminFile, true)
	latestLang, err := readFileToStringMap("./eng/latest/"+langFile, false)

	beforeLatest, err, exist := getBeforeLatest("./eng")
	if !exist {
		for key, value := range latestAdmin {
			latestLang[key] = value
		}

		lang = latestLang
		return nil
	}
	adminBefore, err := readFileToStringMap("./eng/"+beforeLatest+"/"+adminFile, true)
	langBefore, err := readFileToStringMap("./eng/"+beforeLatest+"/"+langFile, false)

	if err != nil {
		return err
	}

	for key, value := range latestAdmin {
		if v, ok := adminBefore[key]; ok && v == value {
			continue
		}

		latestLang[key] = value
	}

	for key, value := range latestLang {
		if key > lineOffset {
			break
		}

		if v, ok := langBefore[key]; ok && v == value {
			delete(latestLang, key)
		}
	}

	lang = latestLang
	return nil
}
