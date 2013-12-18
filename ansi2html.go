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
	re := regexp.MustCompile("\x1b\\[\\d+m")
	ansi = re.ReplaceAllStringFunc(ansi, func(match string) string {
		if match == "\x1b[1m" {
			return `<span style="font-weight: bold;">`
		}
		for code, name := range ansiColours {
			if strings.Contains(match, strconv.Itoa(code)) {
				return `<span style="color: ` + name + `;">`
			}
		}
		return match
	})
	ansi = strings.Replace(ansi, "\x1b[0m", "</span>", -1)
	return ansi
}
