// Copyright 2018 The CUE Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"

	"cuelang.org/go/cue/ast"
	"cuelang.org/go/internal/astinternal"
	"cuelang.org/go/internal/cmd/cue-ast-print/helper"
	"github.com/davecgh/go-spew/spew"
	"github.com/rudifa/goutil/stacktrace"
)

// ----------------------------------------------------------------------------
// export debug functions for use in CLI applications

func DebugAstTree(n ast.Node) string {
	return helper.DebugAstTree(n)
}

// DebugStr returns the one-line debug string of the ast.Node
func DebugStr(n ast.Node) string {
	return astinternal.DebugStr(n)
}

// ----------------------------------------------------------------------------
// parser debugging tools by @rudifa

// CuedoLogStackOneline prints the stacktrace at the point of call
func CuedoLogStackOneline() {
	if os.Getenv("CUEDO_PARSER_STACKTRACE") != "" {
		st := stacktrace.CapturedStacktrace()
		st.Trim(8, 3) // trim off the uninteresting, repeating stacktrace lines
		callpoints := st.StacklineCallpoints(true /*short*/)
		length := st.Len()
		fmt.Printf("stacktrace : %s [%d]\n", callpoints, length)
	}
}

// CuedoPrintToken prints the current token
func (p *parser) CuedoPrintToken() {
	if os.Getenv("CUEDO_PARSER_TOKEN") != "" {
		// p.pos, p.tok, p.lit = p.scanner.Scan()
		info := fmt.Sprintf("       token:  p.pos: %v  p.tok: `%s`  p.lit: `%s`", p.pos, p.tok, replaceTrailingNewline(p.lit))
		p.printTrace(indentOther + info + sep + callpoint(2, ""))
	}
}

// CuedoPrintGroup prints the position and texts of a CommentGroup
func (p *parser) CuedoPrintGroup(cg ast.CommentGroup) {
	if os.Getenv("CUEDO_PARSER_COMMENTS_POS") != "" {
		prefix := "CommentGroup:" + sep
		suffix := sep + callpoint(2)
		for _, str := range p.sprintfCommentGroup(cg) {
			p.printTrace(indentOther + prefix + str + suffix)
			prefix = strings.Repeat(" ", len(prefix))
			suffix = strings.Repeat(" ", len(suffix))
		}
	}
}

// CuedoPrintCommentState prints the position and texts (if any) of *p.comments,
// the top of the parser's comments stack
func (p *parser) CuedoPrintCommentState(remarks ...string) {

	if os.Getenv("CUEDO_PARSER_COMMENTS_STACK") != "" {
		p.printCommentsStack(remarks...)
	}
	if os.Getenv("CUEDO_PARSER_COMMENTS_POS") != "" {

		p.printCommentState(remarks...)
	}
}

// CuedoPrintCommentsStack prints the linked list of comments of the parser
// func (p *parser) CuedoPrintCommentsStack(remarks ...string) {
// 	if os.Getenv("CUEDO_PARSER_COMMENTS_STACK") != "" {
// 		p.printCommentsStack(remarks)
// 	}
// }

// CuedoPrintNode prints the type and comments (if any) of a node, and the debugStr
func (p *parser) CuedoPrintNode(node ast.Node, c *commentState, remarks ...string) {
	if os.Getenv("CUEDO_AST_NODE_TYPE_AND_COMMENTS") != "" {

		prefix := "        node:"

		address := fmtAddress(node, true /*short*/)
		if node == nil {
			// print node line and return
			p.printTrace(indentOther + prefix + sep + "nil" + sep + callpoint(2, remarks...) + sep + address)
			return
		}

		pos := node.Pos()
		end := node.End()
		filepath := pos.Filename()

		posix := pos.Offset()
		endix := end.CuedoSafeOffset()

		groups := node.Comments()

		// prepare and print the node line
		commentsFrom := ""
		if len(groups) > 0 {
			commentsFrom = sep + CuedoSetComments + " " + fmtAddress(c, true /*short*/)
		}
		p.printTrace(indentOther + prefix + sep + address + commentsFrom + sep + callpoint(2))
		prefix = strings.Repeat(" ", len(prefix))
		p.printTrace(indentOther + prefix + sep + "type: " + typeof(node))

		// print comments if any
		for _, str := range p.sprintfCommentGroups(groups) {
			p.printTrace(indentOther + sep + prefix + str)
		}

		// print debugstr if any
		debugstr := astinternal.DebugStr(node)
		if debugstr != "" {
			p.printTrace(indentOther + sep + prefix + "debugStr: " + debugstr)
		}

		// print the input text
		closedintv, decoInputLines := decoratedInputText(filepath, posix, endix)
		prefix2 := prefix + closedintv + ":"
		for _, decoInputText := range decoInputLines {
			p.printTrace(indentOther + prefix2 + decoInputText)
			prefix2 = strings.Repeat(" ", len(prefix2))
		}
	}
}

// CuedoSpew prints the node in depth
func (p *parser) CuedoSpew(node ast.Node) {
	if os.Getenv("CUEDO_AST_NODE_SPEW") != "" {

		p.printTrace(indentOther + "    spewNode:" + sep + callpoint(2))
		spewed := spew.Sdump(node)
		spewed = spewed[:len(spewed)-1] // remove trailing newline
		lines := strings.Split(spewed, "\n")
		for _, line := range lines {
			p.printTrace(indent2 + line)
		}
	}
}

// CuedoAssertEqual asserts that two instances are equal and prints a warning if not
func CuedoAssertEqual(v1 interface{}, v2 interface{}) {
	if reflect.DeepEqual(v1, v2) {
		// fmt.Println("v1 and v2 are deeply equal")
	} else {
		fmt.Printf("*** v1 != v2: %v != %v\n", v1, v2)
	}
}

// CuedoExtractTestCases extracts the test cases from the test file and writes them to a json file
// compatible with the test cases in cue/parser/parser_test.go
func CuedoExtractTestCases(testCases []struct{ desc, in, out string }, outfilename string) {

	if outfilename == "" {
		outfilename = "parser_test.json"
	}

	type ExportedStruct struct {
		Desc, In, Out string
	}

	convert := func(s struct{ desc, in, out string }) ExportedStruct {
		return ExportedStruct{s.desc, s.in, s.out}
	}

	exported := make([]ExportedStruct, len(testCases))
	for i, s := range testCases {
		exported[i] = convert(s)
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "\t")
	if err := enc.Encode(exported); err != nil {
		log.Fatalf("%s\n", err)
	}

	// write to file
	file, err := os.Create(outfilename)
	if err != nil {
		log.Fatalf("Failed to create file: %s\n", err)
	}
	defer file.Close()

	_, err = file.WriteString(buf.String())
	if err != nil {
		log.Fatalf("Failed to write to file: %s\n", err)
	}
}

// ---------------------------------------------------------------------------
// printTrace helpers

// printCommentsStack prints the linked list of comments of the parser
func (p *parser) printCommentsStack(remarks ...string) {
	addresses := []*commentState{}

	// walk the linked list of comments back to the root and collect addresses
	current := p.comments
	for current != nil {
		addresses = append(addresses, current)
		current = current.parent
	}

	// generate a list of formatted addresses
	formattedAddresses := make([]string, len(addresses))
	for i, address := range addresses {
		formattedAddresses[i] = fmtAddress(address, true /*short*/)
	}

	reverse(formattedAddresses) // root comes first

	// compose the final string
	lenstr := fmt.Sprintf("(%d)", len(formattedAddresses))
	commentsstackstr := strings.Join(formattedAddresses, " ⇽ ") + sep + lenstr + sep + callpoint(3, remarks...)

	// print the final string
	const leftadjusted = true
	if leftadjusted {
		fmt.Println("p.comments : " + commentsstackstr) // leftadjusted for vertical alignment
	} else {
		p.printTrace(indentOther + "  p.comments:" + sep + commentsstackstr) // follows the indentation
	}
}

// printCommentState prints the position and texts of groups of a commentState
func (p *parser) printCommentState(remarks ...string) {

	setCommentsStr := ""
	indent := indentOther
	if len(remarks) > 0 {
		if remarks[0] == CuedoSetComments {
			setCommentsStr = " " + remarks[0]
			remarks = remarks[1:]
		}
		if len(remarks) > 0 {
			if remarks[0] == CuedoEnter {
				indent = indentEnter
			} else if remarks[0] == CuedoExit {
				indent = indentExit
			}
		}
	}

	state := p.comments

	if state == nil {
		p.printTrace(indent + "commentState: nil" + sep + callpoint(2, remarks...))
		return
	}
	addrstr := fmtAddress(state, true /*short*/) // for brevity
	callpointStr := callpoint(3, remarks...)     //+ sep + remark

	// print the commentState line
	prefix := "commentState:" + sep
	p.printTrace(indent + prefix + fmt.Sprintf("p.tok: %s", p.tok) + sep + addrstr + setCommentsStr + sep + callpointStr)

	prefix = strings.Repeat(" ", len(prefix))
	p.printTrace(indent + prefix + fmt.Sprintf("pos: %d", state.pos))

	// print commentState properties
	lastChildAddr := fmtAddress(state.lastChild, true /*short*/)
	p.printTrace(indent + prefix + fmt.Sprintf("isList: %d", state.isList))
	p.printTrace(indent + prefix + "lastChild: " + lastChildAddr)
	p.printTrace(indent + prefix + fmt.Sprintf("lastPos: %d", state.lastPos))

	// print parser lead comment if any
	if p.leadComment != nil {
		prefix := " leadComment:" + sep
		for _, str := range p.sprintfCommentGroup(*p.leadComment) {
			p.printTrace(indent + prefix + str)
		}
	}

	// print comment groups if any
	groups := state.groups
	if len(groups) > 0 {
		prefixgroups := prefix + "groups: "
		for _, str := range p.sprintfCommentGroups(groups) {
			p.printTrace(indent + prefixgroups + str)
			prefixgroups = strings.Repeat(" ", len(prefixgroups))
		}
	}
}

// ---------------------------------------------------------------------------
// info and formatting helpers

// decoratedInputText returns the formatted closed interval [posix, endix] and
// decorated text from the file for the closed interval [posix, endix]
func decoratedInputText(filepath string, posix int, endix int) (inputInterval string, decoratedInputFrag []string) {

	inputInterval = fmt.Sprintf("%sinput[%d...%d]", sep, posix, endix)

	inputText := inputText(filepath, posix, endix)
	// decostr := "⸨" + inputText + "⸩" // may include newlines
	// decostr := "【" + inputText + "】" // may include newlines
	// decostr := "⁅" + inputText + "⁆" // may include newlines
	decostr := "『" + inputText + "』" // may include newlines

	decoratedInputFrag = strings.Split(decostr, "\n")

	return inputInterval, decoratedInputFrag
}

// inputText returns the text of a file from a given position to another, inclusive
func inputText(filepath string, pos, end int) string {

	if pos > end {
		// log.Printf("inputText: pos: %d end: %d\n", pos, end)
		return ""
	}

	// if possible, should read the file from a buffer or cache

	content, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatal(err)
	}

	text := string(content[pos:end])
	return replaceTrailingNewline(text)
}

// replaceTrailingNewline replaces the trailing newline with ␤
func replaceTrailingNewline(text string) string {
	length := len(text)
	if length > 0 && text[length-1] == '\n' {
		text = text[:length-1] + "␤"
	}
	return text
}

// sprintCommentGroups returns a string representing the position and texts of a slice of CommentGroups
func (p *parser) sprintfCommentGroups(groups []*ast.CommentGroup) []string {
	strs := []string{}
	for _, cg := range groups {
		strs = append(strs, p.sprintfCommentGroup(*cg)...)
	}
	return strs
}

// sprintfCommentGroup returns a string representing the position and texts of a CommentGroup
func (p *parser) sprintfCommentGroup(cg ast.CommentGroup) []string {
	strs := []string{}
	strpos := fmt.Sprintf("Position: %d", cg.Position)
	for i := 0; i < len(cg.List); i++ {
		strs = append(strs, strpos+fmt.Sprintf("  %s", cg.List[i].Text))
		strpos = strings.Repeat(" ", len(strpos))
	}
	return strs
}

// callpoint returns the name of the calling function and its line number, at the specified stack level
// it also prints remarks if any, after the callpèoint
func callpoint(level int, remarks ...string) string {

	// " · ‣ ▸ ◂ ▴ ▾ ▷ ◁ △ ▽ ▹ ◃"

	funcsuffix := "·"
	suffix := ""
	if len(remarks) > 0 {
		remark := remarks[0]
		if remark == "▸" || remark == "◂" || remark == "▷" || remark == "◁" {
			funcsuffix = remark
		} else {
			suffix += " " + remark
		}
		if len(remarks) > 1 {
			suffix += " " + strings.Join(remarks[1:], " ")
		}
	}

	// get the name of the calling function
	// the file and line number of the callpoint, at the specified stack level
	pc, file, line, _ := runtime.Caller(level)
	fullFuncname := runtime.FuncForPC(pc).Name()
	funcname := regexp.MustCompile(`.*\.(.*)$`).ReplaceAllString(fullFuncname, "$1")
	filename := filepath.Base(file)
	return fmt.Sprintf("%s%s @ %s:%d %s", funcname, funcsuffix, filename, line, suffix)
}

func typeof(v interface{}) string {
	return fmt.Sprintf("%T", v)
}

// fmtAddress formats the address of a pointer
// if short is true, the top 4 hex digits are excized for brevity
func fmtAddress(ptr interface{}, short bool) string {
	if ptr == nil {
		return "nil"
	}
	addrstr := fmt.Sprintf("%p", ptr)

	// if ptr is a *commentState, Sprintf the int value ptr.pos
	if _, ok := ptr.(*commentState); ok {
		addrstr += fmt.Sprintf(":%d", ptr.(*commentState).pos)
	}

	if short && len(addrstr) > 6 {
		// assume the top 4 hex digits are the same
		// for all pointers so excize them for brevity
		// e.g. 0xc000930c870 -> 0x930c870
		addrstr = addrstr[:2] + addrstr[6:]
	}
	return addrstr
}

func reverse(addrs []string) {
	for i, j := 0, len(addrs)-1; i < j; i, j = i+1, j-1 {
		addrs[i], addrs[j] = addrs[j], addrs[i]
	}
}

const (
	sep         = "  "
	indentOther = "      · "
	indentEnter = "      ▹ "
	indentExit  = "      ◃ "

	indent2 = "          "

	CuedoSetComments = "⍇ⓒ"
	CuedoEnter       = "▸"
	CuedoExit        = "◂"
)
