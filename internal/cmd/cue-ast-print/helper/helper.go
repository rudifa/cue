// Copyright 2024 The CUE Authors
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

// cue-ast-print parses a CUE file and prints its syntax tree, for example:
//
//	cue-ast-print file.cue
//
// package main
package helper

import (
	"bytes"

	"fmt"
	gotoken "go/token"
	"io"
	"reflect"
	"strings"

	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/token"
)

// DebugAstTree returns a string representation of the AST tree.
func DebugAstTree(node ast.Node, indentopt ...string) string {
	// provide indent
	indent := "\t" // default
	if len(indentopt) > 0 {
		indent = indentopt[0]
	}
	// set up the printer
	var buf bytes.Buffer
	w := &buf
	d := &debugPrinter{w: w, indent: indent}

	// print the tree recursively
	d.value(reflect.ValueOf(node))
	d.newline()

	// return the result
	return buf.String()
}

type debugPrinter struct {
	w      io.Writer
	level  int
	indent string
}

func (d *debugPrinter) printf(format string, args ...any) {
	fmt.Fprintf(d.w, format, args...)
}

func (d *debugPrinter) newline() {
	fmt.Fprintf(d.w, "\n%s", strings.Repeat(d.indent, d.level))
}

var (
	typeTokenPos   = reflect.TypeOf((*token.Pos)(nil)).Elem()
	typeTokenToken = reflect.TypeOf((*token.Token)(nil)).Elem()
)

func (d *debugPrinter) value(v reflect.Value) {
	// Skip over interface types.
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	// Check before use, to avoid panicking on zero values.
	if !v.IsValid() {
		// Indirecting a nil interface/pointer gives a zero value.
		d.printf("nil")
		return
	}
	// We print the original pointer type if there was one.
	origType := v.Type()
	v = reflect.Indirect(v)

	t := v.Type()
	switch t {
	// Simple types which can stringify themselves.
	case typeTokenPos, typeTokenToken:
		d.printf("%s(%q)", t, v.Interface())
		return
	}

	switch t.Kind() {
	default:
		// We assume all other kinds are basic in practice, like string or bool.
		if t.PkgPath() != "" {
			// Mention defined and non-predeclared types, for clarity.
			d.printf("%s(%#v)", t, v.Interface())
		} else {
			d.printf("%#v", v.Interface())
		}
	case reflect.Slice:
		d.printf("%s{", origType)
		d.level++
		for i := 0; i < v.Len(); i++ {
			d.newline()
			ev := v.Index(i)
			d.value(ev)
		}
		d.level--
		d.newline()
		d.printf("}")
	case reflect.Struct:
		d.printf("%s{", origType)
		d.level++
		for i := 0; i < v.NumField(); i++ {
			f := t.Field(i)
			if !gotoken.IsExported(f.Name) {
				continue
			}
			switch f.Name {
			// These fields are cyclic, and they don't represent the syntax anyway.
			case "Scope", "Node", "Unresolved":
				continue
			}
			d.newline()
			d.printf("%s: ", f.Name)
			d.value(v.Field(i))
		}
		val := v.Addr().Interface()
		if val, ok := val.(ast.Node); ok {
			// Comments attached to a node aren't a regular field, but are still useful.
			// The majority of nodes won't have comments, so skip them when empty.
			if comments := ast.Comments(val); len(comments) > 0 {
				d.newline()
				d.printf("Comments: ")
				d.value(reflect.ValueOf(comments))
			}
		}
		d.level--
		d.newline()
		d.printf("}")
	}
}
