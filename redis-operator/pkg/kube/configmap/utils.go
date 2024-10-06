package configmap

import (
	"bufio"
	"strings"

	v1 "k8s.io/api/core/v1"
)

func UpdateConfigMapKey(configMap *v1.ConfigMap, key string, substr string, newSubstr string) bool {

	if foundValue, ok := configMap.Data[key]; ok {
		lines := strings.Split(foundValue, "\n")
		found := false
		for i, line := range lines {
			if strings.Contains(line, substr) {
				lines[i] = newSubstr
				found = true
			}
		}

		if !found {
			lines = append(lines, newSubstr)
		}

		configMap.Data[key] = strings.Join(lines, "\n")
		return true
	}
	return false
}

func AddConfigMapKey(configMap *v1.ConfigMap, key string, newSubStr string) bool {
	foundValue, ok := configMap.Data[key]
	if !ok {
		return false
	}
	configMap.Data[key] = foundValue + newSubStr
	return true
}

func GetConfigMapSubstrings(configMap map[string]string, key string, substr string) []string {
	foundValue, ok := configMap[key]
	if !ok {
		return nil
	}

	scanner := bufio.NewScanner(strings.NewReader(foundValue))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, substr) {
			return strings.Fields(line[len(substr):])
		}
	}
	return nil
}

func GetConfigMapValues(configMap map[string]string, key string, substr string) []string {

	substrFields := strings.Fields(substr)
	if foundValue, ok := configMap[key]; ok {

		scanner := bufio.NewScanner(strings.NewReader(foundValue))
		for scanner.Scan() {

			line := scanner.Text()
			fields := strings.Fields(line)

			if len(fields) < len(substrFields) {
				continue
			}

			found := true
			for index, sub := range substrFields {
				if sub != fields[index] {
					found = false
					break
				}
			}

			if found {
				// remove the substr fields from the line
				for _, sub := range substrFields {
					line = strings.Replace(line, sub, "", 1)
				}
				return strings.Fields(line)
			}
		}
	}
	return nil
}

func GetConfigMapValue(configMap map[string]string, key string, substr string) string {

	substrFields := strings.Fields(substr)
	if foundValue, ok := configMap[key]; ok {

		scanner := bufio.NewScanner(strings.NewReader(foundValue))
		for scanner.Scan() {

			line := scanner.Text()
			fields := strings.Fields(line)

			if len(fields) < len(substrFields) {
				continue
			}

			found := true
			for index, sub := range substrFields {
				if sub != fields[index] {
					found = false
					break
				}
			}

			if found {
				// remove the substr fields from the line
				for _, sub := range substrFields {
					line = strings.Replace(line, sub, "", 1)
				}
				fields = strings.Fields(line)
				if len(fields) > 0 {
					return fields[0]
				}
				return ""
			}
		}
	}
	return ""
}
