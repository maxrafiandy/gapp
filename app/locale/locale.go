package locale

import "golang.org/x/text/language"

type Tag string

var (
	Bahasa  Tag = Tag(language.Indonesian.String())
	English Tag = Tag(language.English.String())

	DefaultLocale = English
)

type LangPackage struct {
	Tag     Tag
	Message string
}

func SupportedTags() []Tag {
	return []Tag{
		Bahasa,
		English,
	}
}

func IsSupported(tag Tag) bool {
	for _, t := range SupportedTags() {
		if t == tag {
			return true
		}
	}
	return false
}

var langPackages = make(map[string]LangPackage)

func RegisterLangErrorPackage(key string, tag Tag, message string) {
	langPackages[key] = LangPackage{
		Tag:     tag,
		Message: message,
	}
}
