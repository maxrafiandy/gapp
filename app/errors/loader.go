package errors

import (
	"fmt"
	"log"
	"os"
	"scm/api/app/locale"
	"strconv"

	"gopkg.in/yaml.v3"
)

type rawYAMLErrors struct {
	HttpStatus int                              `yaml:"http_status"`
	Errors     map[string]map[locale.Tag]string `yaml:"errors"`
}

func loadYamlFile(filename string) error {
	log.Print("load built-in error file: ", filename)

	data, err := os.ReadFile(fmt.Sprintf("./app/errors/yaml_files/%s", filename))

	if err != nil {
		log.Panic(err)
	}

	return collectBuildinErrors(data)
}

func collectBuildinErrors(data []byte) error {
	var raw rawYAMLErrors
	if err := yaml.Unmarshal(data, &raw); err != nil {
		log.Panic(err)
	}

	log.Print("collecting error...")

	for key, val := range raw.Errors {
		code, _ := strconv.Atoi(val["code"])
		log.Printf("registering: %s", key)
		errorLangPack[key] = ErrAttr{
			HttpStatus: raw.HttpStatus,
			Code:       ErrCode(code),
			Messages: []locale.LangPackage{
				{Tag: locale.English, Message: val[locale.English]},
				{Tag: locale.Bahasa, Message: val[locale.Bahasa]},
			},
		}
	}

	log.Print("built-in error load completed.")

	return nil
}
