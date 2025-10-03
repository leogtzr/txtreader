package text

import "regexp"

func Contains(words *[]string, word string) bool {
	found := false
	for _, v := range *words {
		if v == word {
			found = true
			break
		}
	}

	return found
}

func SanitizeWord(line string) string {
	//line = strings.ReplaceAll(line, ".", "")
	//line = strings.ReplaceAll(line, ",", "")
	//line = strings.ReplaceAll(line, "\"", "")
	//line = strings.ReplaceAll(line, ")", "")
	//line = strings.ReplaceAll(line, "(", "")
	//line = strings.ReplaceAll(line, ":", "")
	//line = strings.ReplaceAll(line, ";", "")
	//return line
	re := regexp.MustCompile(`[.,"()?:;\\]`)
	return re.ReplaceAllString(line, "")
}
