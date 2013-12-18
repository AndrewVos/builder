package main

import (
	"strconv"
	"testing"
)

func testConvertsColour(t *testing.T, code int, name string) {
	ansi := "\x1b[" + strconv.Itoa(code) + `m` + `heloo ansi` + "\x1b[0m"
	expectedHtml := `<span style="color: ` + name + `;">heloo ansi</span>`
	if AnsiToHtml(ansi) != expectedHtml {
		t.Errorf("\nExpected:\n%v\nActual:\n%v\n", expectedHtml, AnsiToHtml(ansi))
	}
}

func TestConvertsANSIToHtml(t *testing.T) {
	testConvertsColour(t, 30, "black")
	testConvertsColour(t, 31, "red")
	testConvertsColour(t, 32, "green")
	testConvertsColour(t, 33, "yellow")
	testConvertsColour(t, 34, "blue")
	testConvertsColour(t, 35, "magenta")
	testConvertsColour(t, 36, "cyan")
	testConvertsColour(t, 37, "white")
	testConvertsColour(t, 90, "grey")
}

func TestBoldensText(t *testing.T) {
	ansi := `some text, some ` + "\x1b[1m" + `bold text` + "\x1b[0m"
	expected := `some text, some <span style="font-weight: bold;">bold text</span>`
	if AnsiToHtml(ansi) != expected {
		t.Errorf("\nExpected:\n%v\nGot:\n%v\n", expected, AnsiToHtml(ansi))
	}
}

func TestReplacesInvalidHtmlCharacters(t *testing.T) {
	ansi := "greater than >, less than <, and ampersand &"
	expected := "greater than &gt;, less than &lt;, and ampersand &amp;"
	if AnsiToHtml(ansi) != expected {
		t.Errorf("\nExpected:\n%v\nGot:\n%v\n", expected, AnsiToHtml(ansi))
	}
}
