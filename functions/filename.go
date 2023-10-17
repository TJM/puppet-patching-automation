package functions

import (
	"strings"
)

var badCharacters = []string{
	"../",
	"<!--",
	"-->",
	"<",
	">",
	"'",
	"\"",
	"&",
	"$",
	"#",
	":",
	"{", "}", "[", "]", "=",
	";", "?", "%20", "%22",
	"%3c",   // <
	"%253c", // <
	"%3e",   // >
	"%0e",   // > -- fill in with % 0 e - without spaces in between
	"%28",   // (
	"%29",   // )
	"%2528", // (
	"%26",   // &
	"%24",   // $
	"%3f",   // ?
	"%3b",   // ;
	"%3d",   // =
}

func removeBadCharacters(input string, dictionary []string) (output string) {

	temp := input

	for _, badChar := range dictionary {
		temp = strings.Replace(temp, badChar, "", -1)
	}
	return temp
}

// TODO: Think about changing this to an "accept" pattern instead of a "reject" pattern

// SanitizeFilename will clean up a string to be useful as a filename
// if relativePath is TRUE, we preserve the path in the filename
// If FALSE and will cause upper path foldername to merge with filename
// USE WITH CARE!!!
func SanitizeFilename(name string, relativePath bool) string {

	// default settings
	var badDictionary = badCharacters

	if name == "" {
		return name
	}

	if !relativePath {
		// add additional bad characters
		badDictionary = append(badCharacters, "./")
		badDictionary = append(badDictionary, "/")
	}

	// trim(remove)white space
	trimmed := strings.TrimSpace(name)

	// trim(remove) white space in between characters
	trimmed = strings.Replace(trimmed, " ", "", -1)

	// remove bad characters from filename
	trimmed = removeBadCharacters(trimmed, badDictionary)

	stripped := strings.Replace(trimmed, "\\", "", -1)

	return stripped
}
