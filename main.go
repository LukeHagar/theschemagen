package theschemagen

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

func PrettyPrint(v interface{}, format string) {
	var b []byte
	var err error

	switch format {
	case "json":
		b, err = json.MarshalIndent(v, "", "  ")
	case "yaml":
		b, err = yaml.Marshal(v)
	default:
		b, err = yaml.Marshal(v)
	}

	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))
}

type SchemaObject struct {
	Type       string                  `json:"type,omitempty" yaml:"type,omitempty"`
	Format     string                  `json:"format,omitempty" yaml:"format,omitempty"`
	Items      interface{}             `json:"items,omitempty" yaml:"items,omitempty"`
	Examples   []interface{}           `json:"examples,omitempty" yaml:"examples,omitempty"`
	Properties map[string]SchemaObject `json:"properties,omitempty" yaml:"properties,omitempty"`
}

func ConvertNumber(number float64) SchemaObject {
	output := SchemaObject{}
	if IsInteger(number) {
		output.Type = "integer"
		if number < 2147483647 && number > -2147483647 {
			output.Format = "int32"
		} else if IsSafeInteger(number) {
			output.Format = "int64"
		}
	} else {
		output.Type = "number"
	}

	output.Examples = []interface{}{number}

	return output
}

func ConvertArray(array []interface{}) SchemaObject {
	output := SchemaObject{Type: "array", Items: nil}
	var outputItems []SchemaObject

	for _, entry := range array {

		objectMap := ConvertObject(entry)
		isDuplicate := false

		for _, item := range outputItems {
			hasSameTypeAndFormat := item.Type == objectMap.Type && item.Format == objectMap.Format
			hasSameProperties := item.Properties != nil && objectMap.Properties != nil &&
				CompareKeys(item.Properties, objectMap.Properties)
			if hasSameTypeAndFormat || hasSameProperties {
				isDuplicate = true
				break
			}
		}

		if !isDuplicate {
			outputItems = append(outputItems, objectMap)
		}

	}

	if len(outputItems) > 1 {
		output.Items = map[string]interface{}{"oneOf": outputItems}
	} else {
		output.Items = outputItems[0]
	}

	return output
}

func ConvertString(str string) SchemaObject {
	output := SchemaObject{Type: "string"}

	regxDate := regexp.MustCompile(`^(19|20)\d{2}-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])$`)
	regxDateTime := regexp.MustCompile(`^(19|20)\d{2}-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])T([0-1][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9]Z$`)

	if regxDateTime.MatchString(str) {
		output.Format = "date-time"
	} else if regxDate.MatchString(str) {
		output.Format = "date"
	}

	output.Examples = []interface{}{str}

	return output
}

func ConvertObject(input interface{}) SchemaObject {
	switch v := input.(type) {
	case nil:
		return SchemaObject{Type: "null"}
	case float64:
		return ConvertNumber(v)
	case []interface{}:
		return ConvertArray(v)
	case map[string]interface{}:
		output := SchemaObject{Type: "object", Properties: make(map[string]SchemaObject)}
		for key, val := range v {
			output.Properties[key] = ConvertObject(val)
		}
		return output
	case string:
		return ConvertString(v)
	case bool:
		output := SchemaObject{Type: "boolean"}

		output.Examples = []interface{}{v}

		return output
	default:
		panic("Invalid type for conversion")
	}
}

func ConvertJSONToOAS(input string) SchemaObject {
	var obj map[string]interface{}
	err := json.Unmarshal([]byte(input), &obj)
	if err != nil {
		panic(err)
	}
	return ConvertObject(obj)
}

func ConvertObjectToOAS(input map[string]interface{}) SchemaObject {
	return ConvertObject(input)
}

var ignoredWords = []string{"the", "a", "an", "of", "to", "in", "for", "with", "on", "at", "from", "by", "and"}

func ConvertSummaryToOperationId(summary string) string {
	words := strings.Split(summary, " ")
	var filteredWords []string
	for i, word := range words {
		if i == 0 {
			filteredWords = append(filteredWords, strings.ToLower(string(word[0]))+word[1:])
		} else {
			if !Contains(ignoredWords, strings.ToLower(word)) {
				filteredWords = append(filteredWords, cases.Title(language.English, cases.NoLower).String(word))
			}
		}
	}
	return strings.Join(filteredWords, "")
}

func Contains(arr []string, str string) bool {
	for _, v := range arr {
		if v == str {
			return true
		}
	}
	return false
}

func CompareKeys(m1, m2 map[string]SchemaObject) bool {
	if len(m1) != len(m2) {
		return false
	}
	for k := range m1 {
		if _, ok := m2[k]; !ok {
			return false
		}
	}
	return true
}

func IsInteger(n float64) bool {
	return n == float64(int(n))
}

func IsSafeInteger(n float64) bool {
	return n <= float64(int64(^uint(0)>>1)) && n >= float64(int64(^uint(0)>>1)*-1)
}
