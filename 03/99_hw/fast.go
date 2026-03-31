package hw3

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/mailru/easyjson"
)

//easyjson:json
type User struct {
	Browsers []string `json:"browsers"`
	Email    string   `json:"email"`
	Name     string   `json:"name"`
}

// FastSearch является оптимальной версией функции SlowSearch.
func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer func() { _ = file.Close() }()

	seenBrowsers := make(map[string]struct{})

	scanner := bufio.NewScanner(file)
	lineNum := 0

	var builder strings.Builder
	builder.WriteString("found users:\n")

	var user User
	for scanner.Scan() {
		err = easyjson.Unmarshal(scanner.Bytes(), &user)
		if err != nil {
			panic(err)
		}

		var isAndroid, isMSIE bool

		browsers := user.Browsers
		for _, b := range browsers {
			if strings.Contains(b, "Android") {
				isAndroid = true
				if _, ok := seenBrowsers[b]; !ok {
					seenBrowsers[b] = struct{}{}
				}
			} else if strings.Contains(b, "MSIE") {
				isMSIE = true
				if _, ok := seenBrowsers[b]; !ok {
					seenBrowsers[b] = struct{}{}
				}
			}
			if isAndroid && isMSIE {
				email := strings.Replace(user.Email, "@", " [at] ", 1)
				lineNumStr := strconv.Itoa(lineNum)
				builder.WriteString("[")
				builder.WriteString(lineNumStr)
				builder.WriteString("] ")
				builder.WriteString(user.Name)
				builder.WriteString(" <")
				builder.WriteString(email)
				builder.WriteString(">\n")
				break
			}
		}

		lineNum++
	}

	if err = scanner.Err(); err != nil {
		panic(err)
	}

	if _, err = fmt.Fprintln(out, builder.String()); err != nil {
		panic(err)
	}

	_, err = fmt.Fprintf(out, "Total unique browsers %d\n", len(seenBrowsers))
	if err != nil {
		panic(err)
	}
}
