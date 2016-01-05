package main

import (
	"bytes"
	"regexp"

	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Converter struct {
	inputLength int

	cursor int

	inInlineMath bool

	in  []rune
	out *bytes.Buffer
}

/* Methods that operate on the input */

// Checks if the cursor has reached the end of the input
func (c *Converter) atEof() bool {
	return c.cursor >= c.inputLength
}

// Returns the character at the given cursor
func (c *Converter) at(cursor int) string {
	if(cursor < 0 || cursor >= c.inputLength) {
		return "\u0003"
	}
	return string(c.in[cursor])
}

// Returns the character at the cursor
func (c *Converter) current() string {
	return c.at(c.cursor)
}

// Returns the next character after the cursor
func (c *Converter) next() string {
	return c.at(c.cursor+1)
}

// Returns the next character after the cursor
func (c *Converter) prev() string {
	return c.at(c.cursor-1)
}

// Returns the next |n| characters after the cursor (i.e. excluding "current()")
func (c *Converter) lookahead(n int) string {
	return string(c.in[c.cursor+1 : c.cursor+1+n])
}

// Same as "lookahead" with a given cursor
func (c *Converter) lookaheadAt(n int, cursor int) string {
	return string(c.in[cursor+1 : cursor+1+n])
}

// Returns the previous |n| characters before the cursor (i.e. excluding "current()")
func (c *Converter) lookback(n int) string {
	return string(c.in[c.cursor-n : c.cursor])
}

/* Methods that operate on the output */

// Writes a string to the output buffer
func (c *Converter) emit(s string) {
	c.out.WriteString(s)
}

/* Parsing \o/ */

// Everything inside an HTML comment is considered to be Latex and thus emitted 1:1
func (c *Converter) handleComments() bool {
	if c.current() != "<" || c.lookahead(3) != "!--" {
		return false
	}

	for !c.atEof() && (c.current() != "-" || c.lookahead(2) != "->") {
		c.emit(c.current())
		c.cursor += 1
	}
	c.emit("-->")
	c.cursor += 3

	return true
}

// CDATA blocks are comments which are completely dropped from the output
func (c *Converter) handleCDATA() bool {
	if c.current() != "<" || c.lookahead(8) != "![CDATA[" {
		return false
	}

	for !c.atEof() && (c.current() != "]" || c.lookahead(2) != "]>") {
		c.cursor += 1
	}
	c.cursor += 3 // For ]]>

	return true
}

func (c *Converter) handleLatex() bool {
	if !c.inInlineMath && c.current() == "\\" && c.next() != "\\" {
		if c.lookahead(5) == "begin" {
			c.handleLatexBlock()
		} else {
			c.handleLatexCommand(true)
		}
		return true
	}
	return false
}

func (c *Converter) handleLatexCommand(emitCommentBlock bool) {
	spaceRegexp := regexp.MustCompile("\\s")

	if emitCommentBlock {
		c.emit("<!--")
	}

	// The command name
	for !c.atEof() &&
		c.current() != "{" &&
		c.current() != "[" &&
		!spaceRegexp.MatchString(c.current()) {

		c.emit(c.current())
		c.cursor += 1
	}

	nesting := 0
	for !c.atEof() {
		// All parameters are closed and there is no next parameter,
		// i.e. \foo{bar}{baz} test 123
		//                    ^
		if nesting == 0 && c.current() != "{" && c.current() != "[" {
			break
		}

		// This will break if there's an unbalanced number of different
		// brace types, i.e. "[[]}" will result in nesting = 0. Don't care
		// to fix that right now.
		if c.current() == "{" || c.current() == "[" {
			nesting += 1
		}

		if c.current() == "}" || c.current() == "]" {
			nesting -= 1
		}

		c.emit(c.current())
		c.cursor += 1
	}

	if emitCommentBlock {
		c.emit("-->")
	}
}

// Handles (nested) \begin{} ... \end{} blocks. Does not care wether you're
// starting/ending the right environment, i.e. this will work:
//
//      \begin{figure} ... \end{math}
//
func (c *Converter) handleLatexBlock() {
	c.emit("<!--")
	nesting := 0

	for !c.atEof() {
		if c.current() == "\\" && c.lookahead(5) == "begin" {
			nesting += 1
		} else if c.current() == "\\" && c.lookahead(3) == "end" {
			nesting -= 1
		}

		// If we're at the last \end, we can just parse it as a command, e.g.:
		//
		//      \end{figure*}
		//      ^
		//
		// At that point, handleLatexCommand will consume everything including
		// "}" and then return.
		if nesting == 0 {
			c.handleLatexCommand(false)
			c.emit("-->")
			break
		}

		c.emit(c.current())
		c.cursor += 1
	}
}

func (c *Converter) handleInlineMath() bool {
	if c.current() == "\\" && c.next() == "$" {
		c.emit("\\$")
		c.cursor += 2
		return true
	}

	if c.current() != "$" {
		return false
	}

	// From http://fletcher.github.io/MultiMarkdown-4/math.html:
	// In order to be correctly parsed as math, there must not be any space
	// between the $ and the actual math on the inside of the delimiter, and
	// there must be space on the outside.
	if c.cursor > 0 && string(c.in[c.cursor-1]) == " " && !c.inInlineMath {
		c.inInlineMath = true
	} else if c.inInlineMath {
		c.inInlineMath = false
	}

	return false
}

func (c *Converter) handleMerkdwernInlineMath() bool {
	if c.current() != "•" {
		return false
	}

	if !c.inInlineMath {
		c.inInlineMath = true
		c.emit("<!--$")
	} else if c.inInlineMath {
		c.inInlineMath = false
		c.emit("$-->")
	}

	c.cursor += 1

	return true
}

func (c *Converter) handleNonBreakingSpace() bool {
	// The second clause is needed because sometimes, e.g. directly after a
	// latex command, the case "c.next() == ~" is not hit.
	if (c.current() != "\\" && c.next() == "~") ||
	   (c.current() == "~" && c.prev() != "\\") {

		if c.current() != "~" {
			c.emit(c.current())
			c.cursor += 2
		} else {
			c.cursor += 1
		}

		c.emit("<!--~-->")

		return true
	}

	return false
}

// Conversion loop iterating over all characters. Not very efficient, but does its job.
func (c *Converter) Convert() []byte {
	for !c.atEof() {
		if c.handleComments() {
			continue
		}

		if c.handleCDATA() {
			continue
		}

		if c.handleInlineMath() {
			continue
		}

		if c.handleMerkdwernInlineMath() {
			continue
		}

		if c.handleNonBreakingSpace() {
			continue
		}

		if c.handleLatex() {
			continue
		}

		c.emit(c.current())
		c.cursor += 1
	}

	return c.out.Bytes()
}

/* Utility */

func ByteArrayToConverter(in []byte) Converter {
	runes := []rune(string(in))
	return Converter{
		inputLength: len(runes),
		cursor:      0,
		in:          runes,
		out:         new(bytes.Buffer),
	}
}

func SXMD(in []byte) []byte {
	c := ByteArrayToConverter(in)
	return c.Convert()
}

func main() {
	flag.Parse()
	if len(flag.Args()) != 1 {
		fmt.Printf("Usage: %s <file-to-convert>\n\tSpecify '-' for standard input\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	inputFilePath := flag.Arg(0)

	var content []byte
	var err error
	if(inputFilePath == "-") {
		content, err = ioutil.ReadAll(os.Stdin)
	} else {
		content, err = ioutil.ReadFile(inputFilePath)
	}

	if err != nil {
		fmt.Printf("Could not read input file %s", inputFilePath)
		os.Exit(1)
	}

	content = SXMD(content)
	os.Stdout.Write(content)
}
