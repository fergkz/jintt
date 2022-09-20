package DomainTool

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type pretty struct{}

var Pretty pretty

func (Pretty pretty) Prepare(data interface{}) string {
	prd, _ := json.MarshalIndent(data, "", "  ")
	return string(prd)
}

func (Pretty pretty) Println(data ...interface{}) {
	var str []interface{}
	if len(data) > 0 {
		for _, d := range data {
			str = append(str, Pretty.Prepare(d))
		}
	}
	fmt.Println(str...)
}

func (Pretty pretty) Save(data interface{}, filename string) {
	file, _ := json.MarshalIndent(data, "", " ")
	_ = ioutil.WriteFile(filename, file, 0644)
}

func (Pretty pretty) Fatalln(data ...interface{}) {
	Pretty.Println(data...)
	os.Exit(1)
}

func GenerateUUIDFromInt(number int) string {
	b := []byte(fmt.Sprintf("%016d", number))
	return fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func GenerateUUIDFromString(text string) string {
	if len(text) > 16 {
		log.Fatalf("text len is over 16 characters '%s'", text)
	}

	b := []byte(fmt.Sprintf("%016s", text))
	return fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func CoalesceString(strs ...string) string {
	nText := ""
	for _, str := range strs {
		if nText != "" {
			break
		}
		nText = str
	}
	return nText
}

func (Pretty pretty) GetCache(dirname string, value interface{}) bool {

	dirname = strings.TrimRight(dirname, "/") + "/"

	fss, err := ioutil.ReadDir(dirname)
	if err != nil {
		return false
	}

	if len(fss) == 0 {
		return false
	}

	for _, fs := range fss {
		timemilit := strings.Split(fs.Name(), ".")

		i, err := strconv.ParseInt(timemilit[0], 10, 64)
		if err != nil {
			log.Panic(err)
		}
		timeFile := time.Unix(i, 0)

		if timeFile.Before(time.Now()) {
			return false
		}

		b, err := ioutil.ReadFile(dirname + fs.Name())
		if err != nil {
			log.Panic(err)
		}

		json.Unmarshal(b, value)
		return true
	}

	return false
}

func (Pretty pretty) SetCache(dirname string, value interface{}, expireSecs int) {

	var nvalue interface{} = value

	dirname = strings.TrimRight(dirname, "/") + "/"

	fss, err := ioutil.ReadDir(dirname)
	if err != nil {
		if err := os.MkdirAll(dirname, os.ModePerm); err != nil {
			log.Panic(err)
		}
	}

	for _, fs := range fss {
		filePath := dirname + fs.Name()
		os.Remove(filePath)
	}

	inSecs := int(time.Now().Unix()) + expireSecs

	cacheFilename := dirname + "/" + strconv.Itoa(inSecs) + ".json"

	file, _ := json.Marshal(nvalue)
	_ = ioutil.WriteFile(cacheFilename, file, 0644)
}
