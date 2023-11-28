package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	regexCache = map[string]*regexp.Regexp{
		"bool": regexp.MustCompile(`^bool_[0-9]+$`),
		"list": regexp.MustCompile(`^list_[0-9]+$`),
		"map":  regexp.MustCompile(`^map_[0-9]+$`),
		"null": regexp.MustCompile(`^null_[0-9]+$`),
		"num":  regexp.MustCompile(`^number_[0-9]+$`),
		"str":  regexp.MustCompile(`^string_[0-9]+$`),
	}
)

func main() {
	if len(os.Args) > 1 {
		bytes := readFile(os.Args[1])

		strData := string(bytes)
		strData = strings.ReplaceAll(strData, " ", "")
		bytes = []byte(strData)

		var input map[string]interface{}
		if err := json.Unmarshal(bytes, &input); err != nil {
			log.Fatal(err)
		}

		result := convertJSON(input)

		resultJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(resultJSON))
	} else {
		fmt.Println("Error: no input file specified.")
	}
}

func readFile(path string) []byte {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	return bytes
}

func convertJSON(input map[string]interface{}) []map[string]interface{} {
	var output []map[string]interface{}

	result := map[string]interface{}{}
	for k, v := range input {
		val, ok := v.(map[string]interface{})
		if ok {
			res := processKeyValue(k, val)
			for k1, v1 := range res {
				result[k1] = v1
			}
		}
	}

	output = append(output, result)
	return output
}

func processKeyValue(k string, sourceMap map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	for regexPattern, regex := range regexCache {
		if regex.MatchString(k) {
			switch regexPattern {
			case "bool":
				handleBoolean(k, sourceMap, result)
			case "list":
				handleList(k, sourceMap, result)
			case "map":
				handleMap(k, sourceMap, result)
			case "null":
				handleNull(k, sourceMap, result)
			case "num":
				handleNumber(k, sourceMap, result)
			case "str":
				handleString(k, sourceMap, result)
			}
			break
		}
	}
	return result
}

func handleBoolean(key string, sourceMap, result map[string]interface{}) {
	if v, ok := sourceMap["BOOL"]; ok {
		if valStr, ok := v.(string); ok {
			switch valStr {
			case "1", "t", "T", "TRUE", "true", "True":
				result[key] = true
			case "0", "f", "F", "FALSE", "false", "False":
				result[key] = false
			}
		}
	}
}

func handleList(key string, sourceMap, result map[string]interface{}) {
	if v, ok := sourceMap["L"]; ok {
		resultList := []interface{}{}
		if list, ok := v.([]interface{}); ok {
			for i := range list {
				if item, ok := list[i].(map[string]interface{}); ok {
					res := map[string]interface{}{}
					if _, ok := item["BOOL"]; ok {
						handleBoolean(key, item, res)
					}
					if _, ok := item["N"]; ok {
						handleNumber(key, item, res)
					}
					if _, ok := item["S"]; ok {
						handleString(key, item, res)
					}
					if len(res) > 0 {
						resultList = append(resultList, res[key])
					}
				}
			}
		}
		if len(resultList) > 0 {
			result[key] = resultList
		}
	}
}

func handleMap(key string, sourceMap, result map[string]interface{}) {
	res := map[string]interface{}{}
	for k, v := range sourceMap {
		if innerMap, ok := v.(map[string]interface{}); ok {
			if k == "M" {
				handleMap(key, innerMap, result)
			} else {
				val := processKeyValue(k, innerMap)
				if len(val) > 0 {
					res[k] = val[k]
				}
			}
		}
	}
	if len(res) > 0 {
		result[key] = res
	}
}

func handleNull(key string, sourceMap, result map[string]interface{}) {
	if v, ok := sourceMap["NULL"]; ok {
		if valStr, ok := v.(string); ok {
			switch valStr {
			case "1", "t", "T", "TRUE", "true", "True":
				result[key] = nil
			}
		}
	}
}

func handleNumber(key string, sourceMap, result map[string]interface{}) {
	if v, ok := sourceMap["N"]; ok {
		if valStr, ok := v.(string); ok {
			if num, err := strconv.ParseFloat(valStr, 64); err == nil {
				result[key] = num
			}
		}
	}
}

func handleString(key string, sourceMap, result map[string]interface{}) {
	if v, ok := sourceMap["S"]; ok {
		if valStr, ok := v.(string); ok {
			if epochTime, err := time.Parse(time.RFC3339, valStr); err == nil {
				result[key] = epochTime.Unix()
			} else {
				if valStr != "" {
					result[key] = valStr
				}
			}
		}
	}
}
