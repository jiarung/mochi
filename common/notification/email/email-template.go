package email

import (
	"fmt"
	"strings"

	"github.com/jiarung/mochi/database/fixtures"
)

var templateList = map[string]struct{}{
	"GENERIC":       struct{}{},
	"GENERIC_EVENT": struct{}{},
}

func getLocale(id string, locale []string) string {
	if len(locale) > 0 {
		if _, exist := fixtures.EmailTemplates[id]; !exist {
			return "en"
		}
		if _, exist := fixtures.
			EmailTemplates[id].
			TranslationMap[locale[0]]; exist {
			return locale[0]
		}
	}
	return "en"
}

func executeTemplate(template string, paramMap map[string]string) string {
	for k, v := range paramMap {
		template = strings.Replace(template, k, v, -1)
	}
	return template
}

func genTemplate(id, lang, subject string, paramMap map[string]string) (
	template string, substitutions map[string]string) {
	template = fixtures.EmailTemplates[id].SendgridTemplate

	if _, exist := templateList[template]; !exist {
		panic(fmt.Sprintf("sendgrid template id(%v,%v) is not is template list",
			id, template))
	}

	htmltemplate := fixtures.EmailTemplates[id].
		TranslationMap[lang]["template"]
	htmltemplate = executeTemplate(htmltemplate, paramMap)
	substitutions = map[string]string{
		"subject":  subject,
		"template": htmltemplate,
	}

	return
}
