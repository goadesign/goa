// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// copied from golang.org/x/tools/imports

package codegen

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"go/token"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

func FormatFile(fset *token.FileSet, f *ast.File) ([]byte, error) {
	sortImports(fset, f)
	imps := astutil.Imports(fset, f)

	var spacesBefore []string // import paths we need spaces before
	for _, impSection := range imps {
		// Within each block of contiguous imports, see if any
		// import lines are in different group numbers. If so,
		// we'll need to put a space between them so it's
		// compatible with gofmt.
		lastGroup := -1
		for _, importSpec := range impSection {
			importPath, _ := strconv.Unquote(importSpec.Path.Value)
			groupNum := importGroup(importPath)
			if groupNum != lastGroup && lastGroup != -1 {
				spacesBefore = append(spacesBefore, importPath)
			}
			lastGroup = groupNum
		}

	}

	var buf bytes.Buffer
	printConfig := &printer.Config{Mode: printer.UseSpaces, Tabwidth: 8}
	err := printConfig.Fprint(&buf, fset, f)
	if err != nil {
		return nil, err
	}
	out := buf.Bytes()
	if len(spacesBefore) > 0 {
		out = addImportSpaces(bytes.NewReader(out), spacesBefore)
	}

	out, err = format.Source(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// importToGroup is a list of functions which map from an import path to
// a group number.
var importToGroup = []func(importPath string) (num int, ok bool){
	func(importPath string) (num int, ok bool) {
		if strings.HasPrefix(importPath, "appengine") {
			return 2, true
		}
		return
	},
	func(importPath string) (num int, ok bool) {
		if strings.Contains(importPath, ".") {
			return 1, true
		}
		return
	},
}

func importGroup(importPath string) int {
	for _, fn := range importToGroup {
		if n, ok := fn(importPath); ok {
			return n
		}
	}
	return 0
}

var impLine = regexp.MustCompile(`^\s+(?:[\w\.]+\s+)?"(.+)"`)

func addImportSpaces(r io.Reader, breaks []string) []byte {
	var out bytes.Buffer
	sc := bufio.NewScanner(r)
	inImports := false
	done := false
	for sc.Scan() {
		s := sc.Text()

		if !inImports && !done && strings.HasPrefix(s, "import") {
			inImports = true
		}
		if inImports && (strings.HasPrefix(s, "var") ||
			strings.HasPrefix(s, "func") ||
			strings.HasPrefix(s, "const") ||
			strings.HasPrefix(s, "type")) {
			done = true
			inImports = false
		}
		if inImports && len(breaks) > 0 {
			if m := impLine.FindStringSubmatch(s); m != nil {
				if m[1] == string(breaks[0]) {
					out.WriteByte('\n')
					breaks = breaks[1:]
				}
			}
		}

		fmt.Fprintln(&out, s)
	}
	return out.Bytes()
}

// sortImports sorts runs of consecutive import lines in import blocks in f.
// It also removes duplicate imports when it is possible to do so without data loss.
func sortImports(fset *token.FileSet, f *ast.File) {
	for i, d := range f.Decls {
		d, ok := d.(*ast.GenDecl)
		if !ok || d.Tok != token.IMPORT {
			// Not an import declaration, so we're done.
			// Imports are always first.
			break
		}

		if len(d.Specs) == 0 {
			// Empty import block, remove it.
			f.Decls = append(f.Decls[:i], f.Decls[i+1:]...)
		}

		if !d.Lparen.IsValid() {
			// Not a block: sorted by default.
			continue
		}

		// Identify and sort runs of specs on successive lines.
		i := 0
		specs := d.Specs[:0]
		for j, s := range d.Specs {
			if j > i && fset.Position(s.Pos()).Line > 1+fset.Position(d.Specs[j-1].End()).Line {
				// j begins a new run.  End this one.
				specs = append(specs, sortSpecs(fset, f, d.Specs[i:j])...)
				i = j
			}
		}
		specs = append(specs, sortSpecs(fset, f, d.Specs[i:])...)
		d.Specs = specs

		// Deduping can leave a blank line before the rparen; clean that up.
		if len(d.Specs) > 0 {
			lastSpec := d.Specs[len(d.Specs)-1]
			lastLine := fset.Position(lastSpec.Pos()).Line
			if rParenLine := fset.Position(d.Rparen).Line; rParenLine > lastLine+1 {
				fset.File(d.Rparen).MergeLine(rParenLine - 1)
			}
		}
	}
}

func importPath(s ast.Spec) string {
	t, err := strconv.Unquote(s.(*ast.ImportSpec).Path.Value)
	if err == nil {
		return t
	}
	return ""
}

func importName(s ast.Spec) string {
	n := s.(*ast.ImportSpec).Name
	if n == nil {
		return ""
	}
	return n.Name
}

func importComment(s ast.Spec) string {
	c := s.(*ast.ImportSpec).Comment
	if c == nil {
		return ""
	}
	return c.Text()
}

// collapse indicates whether prev may be removed, leaving only next.
func collapse(prev, next ast.Spec) bool {
	if importPath(next) != importPath(prev) || importName(next) != importName(prev) {
		return false
	}
	return prev.(*ast.ImportSpec).Comment == nil
}

type posSpan struct {
	Start token.Pos
	End   token.Pos
}

func sortSpecs(fset *token.FileSet, f *ast.File, specs []ast.Spec) []ast.Spec {
	// Can't short-circuit here even if specs are already sorted,
	// since they might yet need deduplication.
	// A lone import, however, may be safely ignored.
	if len(specs) <= 1 {
		return specs
	}

	// Record positions for specs.
	pos := make([]posSpan, len(specs))
	for i, s := range specs {
		pos[i] = posSpan{s.Pos(), s.End()}
	}

	// Identify comments in this range.
	// Any comment from pos[0].Start to the final line counts.
	lastLine := fset.Position(pos[len(pos)-1].End).Line
	cstart := len(f.Comments)
	cend := len(f.Comments)
	for i, g := range f.Comments {
		if g.Pos() < pos[0].Start {
			continue
		}
		if i < cstart {
			cstart = i
		}
		if fset.Position(g.End()).Line > lastLine {
			cend = i
			break
		}
	}
	comments := f.Comments[cstart:cend]

	// Assign each comment to the import spec preceding it.
	importComment := map[*ast.ImportSpec][]*ast.CommentGroup{}
	specIndex := 0
	for _, g := range comments {
		for specIndex+1 < len(specs) && pos[specIndex+1].Start <= g.Pos() {
			specIndex++
		}
		s := specs[specIndex].(*ast.ImportSpec)
		importComment[s] = append(importComment[s], g)
	}

	// Sort the import specs by import path.
	// Remove duplicates, when possible without data loss.
	// Reassign the import paths to have the same position sequence.
	// Reassign each comment to abut the end of its spec.
	// Sort the comments by new position.
	sort.Sort(byImportSpec(specs))

	// Dedup. Thanks to our sorting, we can just consider
	// adjacent pairs of imports.
	deduped := specs[:0]
	for i, s := range specs {
		if i == len(specs)-1 || !collapse(s, specs[i+1]) {
			deduped = append(deduped, s)
		} else {
			p := s.Pos()
			fset.File(p).MergeLine(fset.Position(p).Line)
		}
	}
	specs = deduped

	// Fix up comment positions
	for i, s := range specs {
		s := s.(*ast.ImportSpec)
		if s.Name != nil {
			s.Name.NamePos = pos[i].Start
		}
		s.Path.ValuePos = pos[i].Start
		s.EndPos = pos[i].End
		for _, g := range importComment[s] {
			for _, c := range g.List {
				c.Slash = pos[i].End
			}
		}
	}

	sort.Sort(byCommentPos(comments))

	return specs
}

type byImportSpec []ast.Spec // slice of *ast.ImportSpec

func (x byImportSpec) Len() int      { return len(x) }
func (x byImportSpec) Swap(i, j int) { x[i], x[j] = x[j], x[i] }
func (x byImportSpec) Less(i, j int) bool {
	ipath := importPath(x[i])
	jpath := importPath(x[j])

	igroup := importGroup(ipath)
	jgroup := importGroup(jpath)
	if igroup != jgroup {
		return igroup < jgroup
	}

	if ipath != jpath {
		return ipath < jpath
	}
	iname := importName(x[i])
	jname := importName(x[j])

	if iname != jname {
		return iname < jname
	}
	return importComment(x[i]) < importComment(x[j])
}

type byCommentPos []*ast.CommentGroup

func (x byCommentPos) Len() int           { return len(x) }
func (x byCommentPos) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x byCommentPos) Less(i, j int) bool { return x[i].Pos() < x[j].Pos() }
