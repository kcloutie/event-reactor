package template

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/kcloutie/event-reactor/pkg/logger"
	"github.com/kcloutie/event-reactor/pkg/params/settings"
	"go.uber.org/zap"
)

const (
	LeftDelimPropertyName  = "leftDelim"
	RightDelimPropertyName = "rightDelim"
	IgnoreTemplateErrors   = "ignoreTemplateErrors"
)

type RenderTemplateOptions struct {
	LeftDelim            string
	RightDelim           string
	IgnoreTemplateErrors bool
	RemoveDangerousFuncs bool
}

func NewRenderTemplateOptions() RenderTemplateOptions {
	return RenderTemplateOptions{
		LeftDelim:  settings.GoTemplateDefaultDelimLeft,
		RightDelim: settings.GoTemplateDefaultDelimRight,
	}
}

func RenderTemplateValues(ctx context.Context, templateContent string, path string, dataSource any, replacements []string, opts RenderTemplateOptions) ([]byte, error) {
	toBeTempRemoved := map[string]string{}
	replacementsMap := map[string]string{}
	wildCardReplacementsMap := map[string]string{}

	varRegex := regexp.MustCompile(fmt.Sprintf(`%s\s*(.*?)\s*%s`, regexp.QuoteMeta(opts.LeftDelim), regexp.QuoteMeta(opts.RightDelim)))

	log := logger.FromCtx(ctx).
		With(zap.String("templatePath", path), zap.Strings("replacements", replacements), zap.String("templateContent", templateContent), zap.String("leftDelim", opts.LeftDelim), zap.String("rightDelim", opts.RightDelim)).
		With(zap.Any("dataSource", dataSource))
	for _, val := range replacements {
		lastChar := val[len(val)-1:]
		if lastChar == "*" {
			wildCardReplacementsMap[val] = val
		}
		replacementsMap[val] = val
	}

	findAll := varRegex.FindAllString(templateContent, -1)

	for _, foundGoVariable := range findAll {
		toTest := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(foundGoVariable, opts.LeftDelim, ""), opts.RightDelim, ""))
		_, exists := replacementsMap[toTest]

		if exists {
			replaceWith := fmt.Sprintf("(~(%s)~)", toTest)
			toBeTempRemoved[replaceWith] = foundGoVariable
			templateContent = strings.ReplaceAll(templateContent, foundGoVariable, replaceWith)
		} else {
			cleaned := strings.TrimLeft(toTest, ".")
			if strings.Contains(cleaned, ".") {
				key := "." + strings.Split(cleaned, ".")[0] + "*"
				_, exists := wildCardReplacementsMap[key]
				if exists {
					replaceWith := fmt.Sprintf("(~(%s)~)", toTest)
					toBeTempRemoved[replaceWith] = foundGoVariable
					templateContent = strings.ReplaceAll(templateContent, foundGoVariable, replaceWith)
				}
			}
		}
	}

	funcMap := CreateGoTemplatingFuncMap(opts.RemoveDangerousFuncs)
	templateOptions := "missingkey=error"
	if opts.IgnoreTemplateErrors {
		templateOptions = "missingkey=default"
	}
	contentTpl, err := template.New("content").Option(templateOptions).Funcs(funcMap).Delims(opts.LeftDelim, opts.RightDelim).Parse(templateContent)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to parse the template for '%s'. Error: %v", path, err)
	}
	var tpl bytes.Buffer
	err = contentTpl.Execute(&tpl, dataSource)

	if err != nil {
		newErr := fmt.Errorf("failed to execute the template for '%s'. Error: %v", path, err)
		log.Error(newErr.Error())
		return []byte{}, newErr
	}
	renderedTemplate := tpl.String()
	for find, replaceWith := range toBeTempRemoved {
		renderedTemplate = strings.ReplaceAll(renderedTemplate, find, replaceWith)
	}
	if strings.Contains(renderedTemplate, "<no value>") {
		if !opts.IgnoreTemplateErrors {
			newErr := fmt.Errorf("rendered template '%s' had values that were not found", path)
			log.Error(newErr.Error())
			return []byte{}, newErr
		}
		log.Warn("rendered template had values that were not found")
	}
	return []byte(renderedTemplate), nil
}
