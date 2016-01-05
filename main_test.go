package main

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func ConvertAndCompareFile(t *testing.T, inputFilePath string) {
	content, err := ioutil.ReadFile(inputFilePath)
	if err != nil {
		t.Fatalf("Could not read input file %s", inputFilePath)
	}

	content = SXMD(content)
	compareFilePath := strings.Replace(inputFilePath, ".xmd", ".md", 1)

	blaname := compareFilePath + "-generated"
	err = ioutil.WriteFile(blaname, content, 0644)
	if err != nil {
		t.Fatalf("Could not write generated file %s", blaname)
	}

	cmd := exec.Command("grc", "wdiff", blaname, compareFilePath)
	stdin, _ := cmd.StdinPipe()
	stdin.Write(content)
	stdin.Close()
	wdiff, err := cmd.Output()

	if err != nil { // Files were not the same
		t.Errorf(`
##############################
Files %s and %s were not the same:
##############################

%s`,
			inputFilePath, compareFilePath, string(wdiff))
	}
}

func TestExampleFiles(t *testing.T) {
	files, _ := filepath.Glob("./example-files/*.xmd")
	for _, file := range files {
		ConvertAndCompareFile(t, file)
	}
}

func TestEofCases(t *testing.T) {
	c := getTestConverter("<!--foobar")
	assert.Equal(t, "<!--foobar-->", string(c.Convert()))

	c = getTestConverter("<![CDATA[foobar")
	assert.Equal(t, "", string(c.Convert()))

	c = getTestConverter("\\foo{")
	assert.Equal(t, "<!--\\foo{-->", string(c.Convert()))

	c = getTestConverter("\\foo[")
	assert.Equal(t, "<!--\\foo[-->", string(c.Convert()))

	c = getTestConverter("\\begin")
	assert.Equal(t, "<!--\\begin", string(c.Convert()))

	c = getTestConverter("\\foobar")
	assert.Equal(t, "<!--\\foobar-->", string(c.Convert()))
}

func getTestConverter(input string) Converter {
	return ByteArrayToConverter([]byte(input))
}

func TestUnicodeLengthIsValid(t *testing.T) {
	c := getTestConverter("Falsches Üben von Xylophonmusik quält jeden größeren Zwerg")
	assert.Equal(t, 58, c.inputLength)
}

func TestGeneralCursorFunctions(t *testing.T) {
	c := getTestConverter("Falsches Üben von Xylophonmusik quält jeden größeren Zwerg")
	assert.Equal(t, "F", c.current())
	assert.Equal(t, "a", c.next())
	assert.Equal(t, "Ü", c.at(9))
	assert.Equal(t, "alsches Üben ", c.lookahead(13))
	assert.Equal(t, "Üben von Xylophonmusik", c.lookaheadAt(22, 8))
}

func TestLookback(t *testing.T) {
	c := getTestConverter("Falsches Üben von Xylophonmusik quält jeden größeren Zwerg")
	c.cursor += 10
	assert.Equal(t, " Ü", c.lookback(2))
	assert.Equal(t, "Ü", c.prev())
	assert.Equal(t, "Falsches Ü", c.lookback(10))
}

func TestInlineMath(t *testing.T) {
	c := getTestConverter("$5")
	s := string(c.Convert())
	assert.Equal(t, "$5", s)

	c = getTestConverter("$\\foo")
	s = string(c.Convert())
	assert.Equal(t, "$<!--\\foo-->", s)

	// Wrongly written $$ command, must have spaces on either side.
	c = getTestConverter("$\\foo$")
	s = string(c.Convert())
	assert.Equal(t, "$<!--\\foo$-->", s)

	c = getTestConverter(" $\\foo$ ")
	s = string(c.Convert())
	assert.Equal(t, " $\\foo$ ", s)
}

func BenchmarkExampleFiles(b *testing.B) {
    files, _ := filepath.Glob("./example-files/*.xmd")
    for n := 0; n < b.N; n++ {
		for _, file := range files {
			content, err := ioutil.ReadFile(file)
			if err == nil {
				SXMD(content)
			}
		}
	}
}
