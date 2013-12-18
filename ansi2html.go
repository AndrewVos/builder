package main

import (
	"regexp"
	"strconv"
	"strings"
)

var ansiColours = map[int]string{
	30: "black",
	31: "red",
	32: "green",
	33: "yellow",
	34: "blue",
	35: "magenta",
	36: "cyan",
	37: "white",
	90: "grey",
}

func AnsiToHtml(ansi string) string {
	ansi = strings.Replace(ansi, `&`, `&amp;`, -1)
	ansi = strings.Replace(ansi, `>`, `&gt;`, -1)
	ansi = strings.Replace(ansi, `<`, `&lt;`, -1)
	re := regexp.MustCompile(`\\e\[\d\dm`)
	ansi = re.ReplaceAllStringFunc(ansi, func(match string) string {
		for code, name := range ansiColours {
			if strings.Contains(match, strconv.Itoa(code)) {
				return strings.Replace(match, `\e[`+strconv.Itoa(code)+`m`, `<span style="color: `+name+`;">`, 1)
			}
		}
		return match
	})
	ansi = strings.Replace(ansi, `\e[0m`, "</span>", -1)
	return ansi
}
