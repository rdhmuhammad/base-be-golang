package localize

import (
	"encoding/json"
	"fmt"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Language struct {
	bundle        *i18n.Bundle
	localizer     map[string]*i18n.Localizer
	defaultErrMsg *i18n.Message
}

func (l *Language) GetLocalized(lang string, messageId string) (string, error) {

	localizeConfig := i18n.LocalizeConfig{
		MessageID:      messageId,
		DefaultMessage: l.defaultErrMsg,
	}
	localize, err := l.localizer[lang].Localize(&localizeConfig)
	if err != nil {
		return "", err
	}

	return localize, nil
}

func getFileResourceList(basePath string) []string {
	files, err := os.ReadDir(basePath)
	if err != nil {
		panic(fmt.Sprintf("Failed to read directory: %s", err))
	}

	var fileList = make([]string, 0)

	for _, f := range files {
		match, err := filepath.Match("*.json", f.Name())
		if err != nil {
			panic(fmt.Sprintf("Failed to match file: %s", err))
		}
		if match {
			fileList = append(fileList, path.Join(basePath, f.Name()))
		}
	}

	return fileList

}

func NewLanguage(basePath string) Language {
	defaultBundle := i18n.NewBundle(language.English)

	defaultBundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	fileResourceList := getFileResourceList(basePath)
	for _, file := range fileResourceList {
		_, err := defaultBundle.LoadMessageFile(file)
		if err != nil {
			panic(fmt.Sprintf("Error loading English message: %s", err))
		}
	}

	var localizer map[string]*i18n.Localizer
	for _, file := range fileResourceList {
		localCode := strings.Split(filepath.Base(file), ".")
		if len(localCode) < 2 {
			panic(fmt.Sprintf("Invalid localize code: %s", file))
		}
		localizer[localCode[0]] = i18n.NewLocalizer(defaultBundle, localCode[0], language.English.String())
	}

	defaultmessageEn := i18n.Message{
		ID:    "error",
		Other: "Internal Server Error",
	}

	return Language{
		bundle:        defaultBundle,
		localizer:     localizer,
		defaultErrMsg: &defaultmessageEn,
	}
}
