package reactor

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/kcloutie/event-reactor/pkg/config"
	"github.com/kcloutie/event-reactor/pkg/template"
	"go.uber.org/zap"
)

// HasRequiredProperties checks if the given properties map contains all the required properties.
// It takes a map of properties and a slice of required property names as input.
// It returns a boolean indicating whether any required properties are missing, and an error if applicable.
func HasRequiredProperties(properties map[string]config.PropertyAndValue, requiredProperties []string) (bool, error) {
	missingProperties := []string{}
	hasMissing := false
	for _, propName := range requiredProperties {
		_, exists := properties[propName]
		if !exists {
			missingProperties = append(missingProperties, propName)
			hasMissing = true
		}
	}

	if hasMissing {
		return !hasMissing, fmt.Errorf("missing the following required properties: %s", strings.Join(missingProperties, ", "))
	}
	return !hasMissing, nil
}

// SetGoTemplateOptionValues sets the values of the Go template options based on the provided properties.
// It takes a context.Context, a *zap.Logger, a *template.RenderTemplateOptions, and a map[string]config.PropertyAndValue as input parameters.
// The function retrieves the values of the properties for the left delimiter, right delimiter, and ignore template errors from the properties map.
// If the left delimiter property exists, it retrieves its value and assigns it to the config.LeftDelim field.
// If the right delimiter property exists, it retrieves its value and assigns it to the config.RightDelim field.
// If the ignore template errors property exists, it retrieves its value, converts it to a boolean, and assigns it to the config.IgnoreTemplateErrors field.
// If any error occurs during the retrieval or conversion process, an error message is logged using the provided logger.
func SetGoTemplateOptionValues(ctx context.Context, log *zap.Logger, config *template.RenderTemplateOptions, properties map[string]config.PropertyAndValue) {
	propVal, exists := properties[template.LeftDelimPropertyName]
	if exists {
		propVal, err := propVal.GetStringValue(ctx, log, nil)
		if err != nil {
			log.Error(fmt.Sprintf("error getting value for property %s: %s", template.LeftDelimPropertyName, err.Error()))
		} else {
			config.LeftDelim = propVal
		}
	}

	propVal, exists = properties[template.RightDelimPropertyName]
	if exists {
		propVal, err := propVal.GetStringValue(ctx, log, nil)
		if err != nil {
			log.Error(fmt.Sprintf("error getting value for property %s: %s", template.LeftDelimPropertyName, err.Error()))
		} else {
			config.RightDelim = propVal
		}

	}

	propVal, exists = properties[template.IgnoreTemplateErrors]
	if exists {

		propVal, err := propVal.GetStringValue(ctx, log, nil)
		if err != nil {
			log.Error(fmt.Sprintf("error getting value for property %s: %s", template.LeftDelimPropertyName, err.Error()))
		} else {
			boolValue, err := strconv.ParseBool(propVal)
			if err != nil {
				log.Error(fmt.Sprintf("invalid value of '%s' specified for the %s property. Expected true or false", propVal, template.IgnoreTemplateErrors))
				return
			}
			config.IgnoreTemplateErrors = boolValue
		}
	}
}

func GetRequiredPropertyNames(p ReactorInterface) []string {
	results := []string{}
	for _, p := range p.GetProperties() {
		if *p.Required {
			results = append(results, p.Name)
		}
	}
	return results
}

func GetReactorHelp(p ReactorInterface) string {
	var help strings.Builder

	help.WriteString("NAME\n")
	help.WriteString(fmt.Sprintf("  %s\n\n", p.GetName()))
	help.WriteString("DESCRIPTION\n")
	help.WriteString(fmt.Sprintf("  %s\n\n", p.GetDescription()))
	help.WriteString("PROPERTIES\n\n")
	for _, prop := range p.GetProperties() {
		help.WriteString(fmt.Sprintf("  %s\n", prop.Name))
		help.WriteString(fmt.Sprintf("    %s\n\n", prop.Description))
		help.WriteString(fmt.Sprintf("    Required:                 %t\n", *prop.Required))
		help.WriteString(fmt.Sprintf("    Type:                     %s\n", prop.Type))
		if prop.Validation != nil {
			help.WriteString(fmt.Sprintf("    Validation Regex:         %s\n", prop.Validation.ValidationRegex))
			help.WriteString(fmt.Sprintf("    Validation Regex Message: %s\n", prop.Validation.ValidationRegexMessage))
			if prop.Validation.AllowNullOrEmpty != nil {
				help.WriteString(fmt.Sprintf("    Allow Null/Empty:         %t\n", *prop.Validation.AllowNullOrEmpty))
			}
			if prop.Validation.MinLength != nil {
				help.WriteString(fmt.Sprintf("    Min Length:               %d\n", *prop.Validation.MinLength))
			}
			if prop.Validation.MaxLength != nil {
				help.WriteString(fmt.Sprintf("    Max Length:               %d\n", *prop.Validation.MaxLength))
			}
		}
		help.WriteString("\n\n")
	}
	help.WriteString("EXAMPLE CONFIG\n\n")
	help.WriteString(fmt.Sprintf("  %s\n\n", p.GetConfigExample()))
	return help.String()

}
