package template

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/olekukonko/tablewriter"
	"github.com/zclconf/go-cty/cty"
	"gopkg.in/yaml.v3"
)

func CreateGoTemplatingFuncMap(removeDangerousFuncs bool) template.FuncMap {
	funcMap := sprig.FuncMap()
	funcMap["prefixSuffixStringArray"] = PrefixSuffixStringArray
	funcMap["prefixStringArray"] = PrefixStringArray
	funcMap["suffixStringArray"] = SuffixStringArray
	funcMap["removeFromStringArray"] = RemoveFromStringArray

	funcMap["jsonToHcl"] = JsonToHcl
	funcMap["toHcl"] = ObjectToHcl
	funcMap["toYaml"] = toYaml
	funcMap["where"] = Where
	funcMap["select"] = Select
	funcMap["toDnsString"] = ToDnsString
	funcMap["toAsciiTable"] = toAsciiTable
	funcMap["toMarkdownTable"] = toMarkdownTable
	funcMap["fromJson"] = FromJson
	funcMap["firstOrDefault"] = FirstOrDefault
	funcMap["newArray"] = NewArray
	if removeDangerousFuncs {
		delete(funcMap, "env")
		delete(funcMap, "expandenv")
	}
	return funcMap
}

func NewArray() []interface{} {
	return []interface{}{}
}

func FirstOrDefault(defaultVal interface{}, list interface{}) interface{} {
	tp := reflect.TypeOf(list).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(list)

		l := l2.Len()
		if l == 0 {
			return defaultVal
		}

		return l2.Index(0).Interface()
	default:
		return defaultVal
	}
}

func FromJson(data string) (interface{}, error) {
	var result interface{}
	err := json.Unmarshal([]byte(data), &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

// toJson encodes an item into a JSON string
func toYaml(v interface{}) string {
	output, _ := yaml.Marshal(v)
	return string(output)
}

func toAsciiTable(headers []string, v interface{}) string {
	return PrintObjectTable(headers, v, 60)
}

func toMarkdownTable(headers []string, v interface{}) string {
	return PrintMarkdownTable(headers, v, 60)
}

func PrintObjectTable(headers []string, data any, maxWidth int) string {
	all, isArray := ConvertInterfaceToStringArray(data, maxWidth)
	// if isArray && headers[0] == "Name" {
	// 	headers[0] = "INDEX"
	// }

	if isArray {
		data := []string{}
		for i := 0; i < len(all[0]); i++ {
			data = append(data, all[0][i][1])
		}

		return strings.Join(data, " ")
	}

	// If we have a single value with no name, just print the value
	if len(all) == 1 && len(all[0]) == 1 && len(all[0][0]) == 2 && all[0][0][0] == "" {
		return all[0][0][1]
	}
	var sb strings.Builder

	for _, item := range all {
		var b bytes.Buffer
		foo := bufio.NewWriter(&b)
		// https://github.com/olekukonko/tablewriter
		table := tablewriter.NewWriter(foo)
		table.SetHeader(headers)
		table.SetAutoWrapText(false)
		table.SetAutoFormatHeaders(true)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetRowSeparator("-")
		table.SetHeaderLine(true)
		table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
		// table.EnableBorder(false)
		table.SetTablePadding("\t") // pad with tabs
		table.SetNoWhiteSpace(true)
		table.AppendBulk(item) // Add Bulk Data
		table.Render()
		foo.Flush()
		sb.Write(b.Bytes())
	}
	return sb.String()
}

func PrintMarkdownTable(headers []string, data any, maxWidth int) string {
	all, isArray := ConvertInterfaceToStringArray(data, maxWidth)
	if isArray && headers[0] == "Name" {
		headers[0] = "INDEX"
	}
	var sb strings.Builder

	for _, item := range all {
		var b bytes.Buffer
		foo := bufio.NewWriter(&b)
		// https://github.com/olekukonko/tablewriter
		table := tablewriter.NewWriter(foo)
		table.SetHeader(headers)
		table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
		table.SetCenterSeparator("|")
		table.AppendBulk(item) // Add Bulk Data
		table.Render()
		foo.Flush()
		sb.Write(b.Bytes())
	}
	return sb.String()
}

func ConvertInterfaceToStringArray(data any, maxWidth int) ([][][]string, bool) {
	allResults := [][][]string{}
	isArrayOutput := false
	switch d := data.(type) {
	case []map[string]interface{}:

		for _, item := range d {
			results := GetStringArrayFromMap(item, maxWidth)
			allResults = append(allResults, results)
		}
	case map[string]interface{}:
		results := GetStringArrayFromMap(d, maxWidth)
		allResults = append(allResults, results)
	case []interface{}:
		for _, item := range d {
			switch i := item.(type) {
			case []map[string]interface{}:
				fmt.Println(i)
			case map[string]interface{}:
				results := GetStringArrayFromMap(i, maxWidth)
				allResults = append(allResults, results)
			case []interface{}:
				arrayRes := [][]string{}
				for aii, ai := range i {
					switch iiv := ai.(type) {
					case map[string]interface{}:
						results := GetStringArrayFromMap(iiv, maxWidth)
						allResults = append(allResults, results)
						continue
					default:
						arrayRes = append(arrayRes, []string{fmt.Sprintf("%v", aii), cleanUpValue(ai, maxWidth)})
					}
				}
				if len(arrayRes) > 0 {
					isArrayOutput = true
					allResults = append(allResults, arrayRes)
				}
			case interface{}:
				allResults = append(allResults, [][]string{{"", cleanUpValue(i, maxWidth)}})
			}
		}
	default:
		fmt.Printf("Unknown type: %T", data)
	}

	return allResults, isArrayOutput
}

func GetStringArrayFromMap(item map[string]interface{}, maxWidth int) [][]string {
	results := [][]string{}

	keys := []string{}
	for k := range item {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		results = append(results, []string{k, cleanUpValue(item[k], maxWidth)})
	}
	return results
}

func cleanUpValue(value any, maxWidth int) string {
	data := fmt.Sprintf("%v", value)
	data = strings.ReplaceAll(data, "\n", "")
	if len(data) > maxWidth {
		return data[0:(maxWidth-3)] + "..."
	}
	return data
}

func PrefixSuffixStringArray(prefix string, suffix string, users []interface{}) []string {
	tempUsers := []string{}
	for _, user := range users {
		tempUsers = append(tempUsers, fmt.Sprintf("%s%s%s", prefix, user, suffix))
	}

	return tempUsers
}

func ToDnsString(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Replace spaces and underscores with hyphens
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")

	// Remove or replace non-Kubernetes-compatible characters
	// This regex will match any character that's not a lowercase letter, number, hyphen, or dot
	reg := regexp.MustCompile(`[^a-z0-9-.]`)
	s = reg.ReplaceAllString(s, "")

	// Ensure the string starts and ends with an alphanumeric character
	s = strings.Trim(s, "-.")

	// Ensure the string doesn't exceed 253 characters
	if len(s) > 253 {
		s = s[:253]
	}

	return s
}

func PrefixStringArray(prefix string, users []interface{}) []string {
	tempUsers := []string{}
	for _, user := range users {
		tempUsers = append(tempUsers, fmt.Sprintf("%s%s", prefix, user))
	}

	return tempUsers
}

func SuffixStringArray(suffix string, users []interface{}) []string {
	tempUsers := []string{}
	for _, user := range users {
		tempUsers = append(tempUsers, fmt.Sprintf("%s%s", user, suffix))
	}

	return tempUsers
}

func Where(propertyName string, propertyValue string, data interface{}) (map[string]interface{}, error) {
	var results map[string]interface{}
	switch a := data.(type) {
	case []map[string]interface{}:
		for _, item := range a {
			val, exists := item[propertyName]
			if exists {
				if val == propertyValue {
					return item, nil
				}
			}
		}
	case []interface{}:
		for _, item := range a {
			switch ia := item.(type) {
			case map[string]interface{}:
				val, exists := ia[propertyName]
				if exists {
					if val == propertyValue {
						return ia, nil
					}
				}
			}
		}
	default:
		return results, fmt.Errorf("where cannot run against type '%T'. Must be a []map[string]interface{}", data)
	}
	return results, nil
}

func Select(propertyName string, data interface{}) (interface{}, error) {
	switch a := data.(type) {
	case map[string]interface{}:
		res, exists := a[propertyName]
		if exists {
			return res, nil
		}
	default:
		return nil, fmt.Errorf("select cannot run against type '%T'. Must be a map[string]interface{}", data)
	}
	return nil, nil
}

func RemoveFromStringArray(toRemove string, users []interface{}) []string {
	tempUsers := []string{}
	for _, user := range users {
		switch str := user.(type) {
		case string:
			tempUsers = append(tempUsers, strings.ReplaceAll(str, toRemove, ""))
		default:
			tempUsers = append(tempUsers, strings.ReplaceAll(fmt.Sprintf("%v", str), toRemove, ""))
		}
	}

	return tempUsers
}

func JsonToHcl(jsonString string) (*string, error) {
	var obj interface{}
	empty := ""
	err := json.Unmarshal([]byte(jsonString), &obj)
	if err != nil {
		return &empty, err
	}
	return ObjectToHcl(obj), nil
}

func ObjectToHcl(obj interface{}) *string {
	ctyVal := convertToCtyValue(obj)
	val := CtyValueToHCLString(ctyVal)
	return &val
}

func convertToCtyValue(value interface{}) cty.Value {
	switch v := value.(type) {

	case map[string]interface{}:
		m := map[string]cty.Value{}
		for k, val := range v {
			m[k] = convertToCtyValue(val)
		}
		return cty.ObjectVal(m)
	case []interface{}:
		var vals []cty.Value
		for _, val := range v {
			vals = append(vals, convertToCtyValue(val))
		}
		return cty.TupleVal(vals)
	case []string:
		var vals []cty.Value
		for _, val := range v {
			vals = append(vals, convertToCtyValue(val))
		}
		return cty.TupleVal(vals)
	case string:
		return cty.StringVal(v)
	case float64:
		return cty.NumberFloatVal(v)
	case int:
		return cty.NumberIntVal(int64(v))

	case int16:
		return cty.NumberIntVal(int64(v))
	case int32:
		return cty.NumberIntVal(int64(v))
	case bool:
		return cty.BoolVal(v)
	default:
		fmt.Printf("Unknown Type: %T", v)
		return cty.NilVal
	}
}

func CtyValueToHCLString(ctyVal cty.Value) string {
	if ctyVal.Type().IsObjectType() {
		strs := []string{}
		for k, v := range ctyVal.AsValueMap() {
			strs = append(strs, fmt.Sprintf("%s = %s", k, CtyValueToHCLString(v)))
		}
		sort.Strings(strs)
		return "{ " + join(strs, ", ") + " }"
	} else if ctyVal.Type().IsTupleType() {
		strs := []string{}
		it := ctyVal.ElementIterator()
		for it.Next() {
			_, v := it.Element()
			strs = append(strs, CtyValueToHCLString(v))
		}
		return "[" + join(strs, ", ") + "]"
	} else if ctyVal.Type() == cty.String {
		stringValue := ctyVal.AsString()
		if strings.ContainsAny(stringValue, "\n\r") {
			return fmt.Sprintf("<<EOT\n%s\nEOT", stringValue)
		}
		return fmt.Sprintf("\"%s\"", stringValue)
	} else if ctyVal.Type() == cty.Number {
		bf := ctyVal.AsBigFloat()
		if i, _ := bf.Int(nil); bf.IsInt() {
			return fmt.Sprintf("%v", i)
		}
		return fmt.Sprintf("%v", bf)
	} else if ctyVal.Type() == cty.Bool {
		return fmt.Sprintf("%v", ctyVal.True())
	} else if ctyVal.Type() == cty.DynamicPseudoType {
		return "null"
	} else {
		return ctyVal.GoString()
	}
}

func join(strs []string, sep string) string {
	var result string
	for i, str := range strs {
		result += str
		if i < len(strs)-1 {
			result += sep
		}
	}
	return result
}
