package text

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
