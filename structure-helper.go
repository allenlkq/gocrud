package main

import (
	"fmt"
	"go/format"
	"sort"
	"strings"
	"unicode"
	"os"
	"log"
	"io/ioutil"
	"os/exec"
	"bytes"
	. "github.com/manishrjain/gocrud/entities"
)

func createStructure(s Structure) {
	// create tmp dir if not already exist
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		os.Mkdir(tmpDir, 0777)
	}
	src := fmt.Sprintf("package %s\ntype %s %s}", pkg, s.Name, generateTypes(s.Definition))

	formatted, err := format.Source([]byte(src))
	if err != nil {
		log.Fatalf("ERROR: formatting: %s, was formatting\n%s", err, src)
	}

	gofile := fmt.Sprintf("%s/%s.go", tmpDir, s.Name)
	err = ioutil.WriteFile(gofile, formatted, 0644)
	if err != nil {
		log.Fatalf("ERROR: writing to tmp file: %q - %s", gofile, err)
	}

	// compile to validate
	cmd := exec.Command("go", "tool", "compile", gofile)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		fmt.Printf(string(formatted))
		log.Fatalf("ERROR: failed to compile %q\n%s", cmd.Path + strings.Join(cmd.Args, " "), out.String())
	}

	// copy to src tree
	err = FileCopy(gofile, fmt.Sprintf("%s/%s.go", pkg, formatted))
	if err != nil {
		fmt.Printf("ERROR: Failed to copy file from %q to %q\n", out.String())
		log.Fatal(err)
	}
	fmt.Printf("INFO: DONE - %q\n", out.String())
}

func generateTypes(obj map[string]string) string {
	structure := "struct {"

	keys := make([]string, 0, len(obj))
	for key := range obj {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		valueType := obj[key]

		//check if value is string or existing types

		fieldName := FmtFieldName(key)

		jsonTag := fmt.Sprintf("json:\"%s\"", key)

		structure += fmt.Sprintf("\n%s %s `%s`",
			fieldName,
			valueType,
			jsonTag)
	}
	return structure
}

// FmtFieldName formats a string as a struct key
//
// Example:
// 	FmtFieldName("foo_id")
// Output: FooID
func FmtFieldName(s string) string {
	runes := []rune(s)
	for len(runes) > 0 && !unicode.IsLetter(runes[0]) && !unicode.IsDigit(runes[0]) {
		runes = runes[1:]
	}
	if len(runes) == 0 {
		return "_"
	}

	name := lintFieldName(s)
	runes = []rune(name)
	for i, c := range runes {
		ok := unicode.IsLetter(c) || unicode.IsDigit(c)
		if i == 0 {
			ok = unicode.IsLetter(c)
		}
		if !ok {
			runes[i] = '_'
		}
	}
	s = string(runes)
	s = strings.Trim(s, "_")
	if len(s) == 0 {
		return "_"
	}
	return s
}

// commonInitialisms is a set of common initialisms.
// Only add entries that are highly unlikely to be non-initialisms.
// For instance, "ID" is fine (Freudian code is rare), but "AND" is not.
var commonInitialisms = map[string]bool{
	"API":   true,
	"ASCII": true,
	"CPU":   true,
	"CSS":   true,
	"DNS":   true,
	"EOF":   true,
	"GUID":  true,
	"HTML":  true,
	"HTTP":  true,
	"HTTPS": true,
	"ID":    true,
	"IP":    true,
	"JSON":  true,
	"LHS":   true,
	"QPS":   true,
	"RAM":   true,
	"RHS":   true,
	"RPC":   true,
	"SLA":   true,
	"SMTP":  true,
	"SSH":   true,
	"TLS":   true,
	"TTL":   true,
	"UI":    true,
	"UID":   true,
	"UUID":  true,
	"URI":   true,
	"URL":   true,
	"UTF8":  true,
	"VM":    true,
	"XML":   true,
	"NTP":   true,
	"DB":    true,
}


func lintFieldName(name string) string {
	// Fast path for simple cases: "_" and all lowercase.
	if name == "_" {
		return name
	}

	allLower := true
	for _, r := range name {
		if !unicode.IsLower(r) {
			allLower = false
			break
		}
	}
	if allLower {
		runes := []rune(name)
		if u := strings.ToUpper(name); commonInitialisms[u] {
			copy(runes[0:], []rune(u))
		} else {
			runes[0] = unicode.ToUpper(runes[0])
		}
		return string(runes)
	}

	allUpperWithUnderscore := true
	for _, r := range name {
		if !unicode.IsUpper(r) && r != '_' {
			allUpperWithUnderscore = false
			break
		}
	}
	if allUpperWithUnderscore {
		name = strings.ToLower(name)
	}

	// Split camelCase at any lower->upper transition, and split on underscores.
	// Check each word for common initialisms.
	runes := []rune(name)
	w, i := 0, 0 // index of start of word, scan
	for i+1 <= len(runes) {
		eow := false // whether we hit the end of a word

		if i+1 == len(runes) {
			eow = true
		} else if runes[i+1] == '_' {
			// underscore; shift the remainder forward over any run of underscores
			eow = true
			n := 1
			for i+n+1 < len(runes) && runes[i+n+1] == '_' {
				n++
			}

			// Leave at most one underscore if the underscore is between two digits
			if i+n+1 < len(runes) && unicode.IsDigit(runes[i]) && unicode.IsDigit(runes[i+n+1]) {
				n--
			}

			copy(runes[i+1:], runes[i+n+1:])
			runes = runes[:len(runes)-n]
		} else if unicode.IsLower(runes[i]) && !unicode.IsLower(runes[i+1]) {
			// lower->non-lower
			eow = true
		}
		i++
		if !eow {
			continue
		}

		// [w,i) is a word.
		word := string(runes[w:i])
		if u := strings.ToUpper(word); commonInitialisms[u] {
			// All the common initialisms are ASCII,
			// so we can replace the bytes exactly.
			copy(runes[w:], []rune(u))

		} else if strings.ToLower(word) == word {
			// already all lowercase, and not the first word, so uppercase the first character.
			runes[w] = unicode.ToUpper(runes[w])
		}
		w = i
	}
	return string(runes)
}



