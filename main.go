// A REPL tool for golang.
// Created by simplejia [8/2015]
package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/scanner"
	"go/token"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var (
	home = filepath.Join(os.Getenv("HOME"), ".gop")
)

func cmd(name string, args ...string) (outBuf, errBuf *bytes.Buffer, err error) {
	outBuf = new(bytes.Buffer)
	errBuf = new(bytes.Buffer)

	defer func() {
		if err != nil {
			msg := ""
			if outBuf.Len() > 0 {
				msg += "[stdout]\n" + outBuf.String()
			}
			if errBuf.Len() > 0 {
				msg += "[stderr]\n" + errBuf.String()
			}
			err = fmt.Errorf("ret: %v, msg: %v", err, msg)
			return
		}
	}()

	cmd := exec.Command(name, args...)
	cmdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	cmderr, err := cmd.StderrPipe()
	if err != nil {
		return
	}
	err = cmd.Start()
	if err != nil {
		return
	}
	io.Copy(outBuf, cmdout)
	io.Copy(errBuf, cmderr)
	err = cmd.Wait()
	if err != nil {
		return
	}

	return
}

// Workspace is the main struct for gop
type Workspace struct {
	pkgs           []interface{}
	pkgs_notimport []interface{}
	defs           []interface{}
	codes          []interface{}
	files          *token.FileSet
	args           string
}

func (w *Workspace) source(print_dpc, print_linenums, print_notimport bool) string {
	source := ""
	if print_dpc {
		source += "\t"
	}
	source += "package main\n\n"

	pkgs_num := 0
	for _, v := range w.pkgs {
		str := new(bytes.Buffer)
		printer.Fprint(str, w.files, v)

		if print_dpc {
			source += "p" + strconv.Itoa(pkgs_num) + ":\t"
		}
		source += str.String() + "\n"
		pkgs_num++
	}

	if print_notimport {
		for _, v := range w.pkgs_notimport {
			str := new(bytes.Buffer)
			printer.Fprint(str, w.files, v)

			if print_dpc {
				source += "p" + strconv.Itoa(pkgs_num) + ":\t"
			}
			source += str.String() + " // imported and not used\n"
			pkgs_num++
		}
	}

	source += "\n"

	for pos, v := range w.defs {
		str := new(bytes.Buffer)
		printer.Fprint(str, w.files, v)

		if print_dpc {
			source += "d" + strconv.Itoa(pos) + ":\t"
			source += strings.Join(strings.Split(str.String(), "\n"), "\n\t")
		} else {
			source += str.String()
		}
		source += "\n\n"
	}

	if print_dpc {
		source += "\t"
	}
	source += "func main() {\n"

	for pos, v := range w.codes {
		str := new(bytes.Buffer)
		printer.Fprint(str, w.files, v)

		if print_dpc {
			source += "c" + strconv.Itoa(pos) + ":\t"
			source += "\t" + strings.Join(strings.Split(str.String(), "\n"), "\n\t\t")
		} else {
			source += "\t" + strings.Join(strings.Split(str.String(), "\n"), "\n\t")
		}
		source += "\n"
	}

	if print_dpc {
		source += "\t"
	}
	source += "}\n"

	if print_linenums {
		newsource := ""
		for line, item := range strings.Split(source, "\n") {
			newsource += strconv.Itoa(line+1) + "\t" + item + "\n"
		}
		source = newsource
	}

	return source
}

func compile(w *Workspace) (err error) {
	filePrefix := filepath.Join(home, "gop")
	ioutil.WriteFile(filePrefix+".go", []byte(w.source(false, false, false)), 0644)

	args := []string{}
	args = append(args, "build")
	args = append(args, "-o", filePrefix, filePrefix+".go")
	_, _, err = cmd("go", args...)
	if err != nil {
		return
	}
	return
}

func run(w *Workspace) (outBuf, errBuf *bytes.Buffer, err error) {
	file := filepath.Join(home, "gop")
	matchs := regexp.MustCompile(`-?\w+|".*?[^\\"]"`).FindAllString(w.args, -1)
	for n, match := range matchs {
		matchs[n] = strings.Replace(strings.Trim(match, "\""), `\"`, `"`, -1)
	}
	outBuf, errBuf, err = cmd(file, matchs...)
	if err != nil {
		return
	}
	return
}

func parseDeclList(fset *token.FileSet, filename string, src string) ([]ast.Decl, error) {
	pkg := ""
	if strings.Index(src, "package ") == -1 {
		pkg = "package p;"
	}
	f, err := parser.ParseFile(fset, filename, pkg+src, 0)
	if err != nil {
		return nil, err
	}
	return f.Decls, nil
}

func parseStmtList(fset *token.FileSet, filename string, src string) ([]ast.Stmt, error) {
	pkg := ""
	if strings.Index(src, "package ") == -1 {
		pkg = "package p;"
	}
	f, err := parser.ParseFile(fset, filename, pkg+"func _(){"+src+"}", 0)
	if err != nil {
		return nil, err
	}
	return f.Decls[0].(*ast.FuncDecl).Body.List, nil
}

func sourceDefaultDPC(w *Workspace) {
	for _, value := range []string{
		"fmt",
		"strconv",
		"strings",
		"time",
		"encoding/json",
		"bytes",
	} {
		if func() bool {
			for _, pkg := range w.pkgs {
				v_j := pkg.(*ast.GenDecl).Specs[0].(*ast.ImportSpec)
				if v_j.Path.Value == "\""+value+"\"" &&
					v_j.Name.String() == "<nil>" {
					return true
				}
			}
			return false
		}() {
			continue
		}
		tree, _ := parseDeclList(w.files, "gop", "import \""+value+"\"")
		w.pkgs_notimport = append(w.pkgs_notimport, tree[0])
	}
}

func execAlias(w *Workspace, line string) string {
	if line == "help" {
		return "?"
	}

	sps := []string{}
	for _, sp := range strings.Split(line, "\n") {
		if p := "echo "; strings.HasPrefix(sp, p) {
			sps = append(sps, "println("+sp[len(p):]+")")
		} else {
			sps = append(sps, sp)
		}
	}
	return strings.Join(sps, "\n")
}

func execSpecial(w *Workspace, line string) bool {
	if strings.HasPrefix(line, ">") {
		file := strings.TrimSpace(line[1:])
		if file != "" {
			file = filepath.Join(home, file)
			if !strings.HasSuffix(file, ".tmpl") {
				file += ".tmpl"
			}
			ioutil.WriteFile(file, []byte(w.source(false, false, false)), 0644)
		}
		return true
	}
	if strings.HasPrefix(line, "<") && !strings.HasPrefix(line, "<-") {
		file := strings.TrimSpace(line[1:])
		if file == "" {
			fmt.Println("No file specified for include.")
			return true
		}
		if !strings.HasSuffix(file, ".tmpl") {
			file += ".tmpl"
		}
		bs, err := ioutil.ReadFile(file)
		if err != nil {
			if os.IsNotExist(err) {
				bs, err = ioutil.ReadFile(filepath.Join(home, file))
			}
			if err != nil {
				fmt.Println("ReadFile error:", err)
				return true
			}
		}

		sepBegin, sepEnd := "func main() {", "}"
		if pos := strings.Index(string(bs), sepBegin); pos != -1 {
			bs = append(bs[:pos], bs[pos+len(sepBegin):]...)
			if pos := strings.LastIndex(string(bs), sepEnd); pos != -1 {
				bs = append(bs[:pos], bs[pos+len(sepEnd):]...)
			}
		}

		w.pkgs = nil
		w.pkgs_notimport = nil
		w.codes = nil
		w.defs = nil
		tmpline := ""
		for _, line := range strings.Split(string(bs), "\n") {
			tmpline += line + "\n"
			notComplete, err := parseGo4import(w, tmpline)
			if err != nil {
				fmt.Println("ParseGo error:", err)
				break
			}
			if notComplete {
				continue
			}
			tmpline = ""
		}
		sourceDefaultDPC(w)
		return true
	}
	if line == "reset" {
		w.pkgs = nil
		w.pkgs_notimport = nil
		w.defs = nil
		w.codes = nil
		sourceDefaultDPC(w)
		return true
	}
	if line == "list" {
		entries, err := ioutil.ReadDir(home)
		if err != nil {
			fmt.Printf("ReadDir %s: %s\n", home, err)
			return true
		}

		tmpls := []string{}
		for _, fi := range entries {
			if fi.IsDir() {
				continue
			}

			name := fi.Name()
			if strings.HasPrefix(name, ".") ||
				!strings.HasSuffix(name, ".tmpl") {
				continue
			}

			tmpls = append(tmpls, name)
		}
		for pos, tmpl := range tmpls {
			fmt.Printf("%d\t%s\n", pos, tmpl)
		}
		return true
	}
	if line == "arg" {
		fmt.Printf("%s\n", w.args)
		return true
	}
	if p := "arg "; strings.HasPrefix(line, p) &&
		!strings.HasPrefix(line, p+"=") &&
		!strings.HasPrefix(line, p+":=") {
		w.args = strings.TrimSpace(line[len(p):])
		return true
	}
	return false
}

func removeByIndex(w *Workspace, cmd_args string) {
	if len(cmd_args) == 0 {
		fmt.Println("Error: no item specified for remove")
		return
	}

	item_type := cmd_args[0]
	item_list_len := map[byte]int{
		'd': len(w.defs) + 1,
		'p': len(w.pkgs) + len(w.pkgs_notimport) + 1,
		'c': len(w.codes) + 1,
	}[item_type] - 1

	if item_list_len == -1 {
		fmt.Printf("Error: invalid item type '%c'\n", item_type)
		return
	}
	if item_list_len == 0 {
		fmt.Printf("Error: no more '%c' to remove\n", item_type)
		return
	}
	items_to_remove := getIndices(item_list_len, cmd_args[1:])

	switch item_type {
	case 'd':
		removeSlice(&w.defs, items_to_remove)
	case 'p':
		items4import, items4notimport := []bool{}, []bool{}
		for pos, v := range items_to_remove {
			if pos < len(w.pkgs) {
				items4import = append(items4import, v)
			} else {
				items4notimport = append(items4notimport, v)
			}
		}
		removeSlice(&w.pkgs, items4import)
		removeSlice(&w.pkgs_notimport, items4notimport)
	case 'c':
		removeSlice(&w.codes, items_to_remove)
	}
}

func getIndices(item_list_len int, cmd_args string) []bool {
	items_to_remove := make([]bool, item_list_len)

	cmd_args = strings.TrimSpace(cmd_args)
	if len(cmd_args) == 0 {
		items_to_remove[item_list_len-1] = true
		return items_to_remove
	}

	item_indices := []string{}
	for _, vi := range strings.Split(cmd_args, ",") {
		if vj := strings.Split(vi, "-"); len(vj) == 2 {
			i, err := strconv.Atoi(vj[0])
			if err != nil {
				fmt.Printf("Error: %s not integer\n", vj[0])
				continue
			}
			j, err := strconv.Atoi(vj[1])
			if err != nil {
				fmt.Printf("Error: %s not integer\n", vj[1])
				continue
			}
			for k := i; k <= j; k++ {
				item_indices = append(item_indices, strconv.Itoa(k))
			}
		} else {
			item_indices = append(item_indices, vi)
		}
	}

	for _, item_index_str := range item_indices {
		if item_index_str == "" {
			continue
		}
		item_index, err := strconv.Atoi(item_index_str)
		if err != nil {
			fmt.Printf("Error: %s not integer\n", item_index_str)
			continue
		}
		if item_index < 0 || item_index >= item_list_len {
			fmt.Printf("Error: %d out of range\n", item_index)
			continue
		}
		items_to_remove[item_index] = true
	}

	return items_to_remove
}

func removeSlice(ps interface{}, removes []bool) {
	rps := reflect.Indirect(reflect.ValueOf(ps))
	num := rps.Len()
	if num == 0 {
		return
	}

	rpsNew := reflect.MakeSlice(rps.Type(), 0, 0)
	for i := 0; i < num; i++ {
		if i < len(removes) && removes[i] {
			continue
		}
		rpsNew = reflect.Append(rpsNew, rps.Index(i))
	}
	rps.Set(rpsNew)
}

func parseGo(w *Workspace, line string) (notComplete bool, err error) {
	pos := -1
	if unicode.IsDigit(rune(line[0])) {
		idx := strings.IndexFunc(line[1:], func(r rune) bool { return !unicode.IsDigit(r) })
		if idx == -1 {
			return
		}
		idx++
		pos, err = strconv.Atoi(line[:idx])
		if err != nil {
			return
		}
		line = strings.TrimSpace(line[idx:])
	}

	var tree interface{}
	tree, err = parseDeclList(w.files, "gop", line[0:])
	if err != nil {
		tree, err = parseStmtList(w.files, "gop", line[0:])
		if err != nil {
			if _, ok := err.(scanner.ErrorList); ok {
				err = nil
				notComplete = true
			}
			return
		}
	}

	bkup_pkgs := append([]interface{}(nil), w.pkgs...)
	bkup_pkgs_notimport := append([]interface{}(nil), w.pkgs_notimport...)
	bkup_codes := append([]interface{}(nil), w.codes...)
	bkup_defs := append([]interface{}(nil), w.defs...)
	bkup_files := w.files

	switch v := tree.(type) {
	case []ast.Stmt:
		if pos > len(w.codes) || pos < 0 {
			pos = len(w.codes)
		}
		for i := len(v) - 1; i >= 0; i-- {
			if v_i, ok := v[i].(*ast.AssignStmt); ok {
				if v_i.Tok == token.DEFINE {
					for _, name_i := range v_i.Lhs {
						str_i := new(bytes.Buffer)
						printer.Fprint(str_i, w.files, name_i)
						if str_i.String() == "_" {
							continue
						}
						tree, _ := parseStmtList(w.files, "gop", "_ = "+str_i.String())
						w.codes = append(w.codes, nil)
						copy(w.codes[pos+1:], w.codes[pos:])
						w.codes[pos] = tree[0]
					}
				}
			}
			w.codes = append(w.codes, nil)
			copy(w.codes[pos+1:], w.codes[pos:])
			w.codes[pos] = v[i]
		}
	case []ast.Decl:
		if pos > len(w.defs) || pos < 0 {
			pos = len(w.defs)
		}
		for i := len(v) - 1; i >= 0; i-- {
			if v_i, ok := v[i].(*ast.GenDecl); ok {
				if v_i.Tok == token.IMPORT {
					for _, spec := range v_i.Specs {
						name := spec.(*ast.ImportSpec).Name.String()
						value := spec.(*ast.ImportSpec).Path.Value
						if func() bool {
							for _, pkg := range w.pkgs {
								v_j := pkg.(*ast.GenDecl).Specs[0].(*ast.ImportSpec)
								if v_j.Path.Value == value &&
									v_j.Name.String() == name {
									return true
								}
							}
							return false
						}() {
							continue
						}
						if func() bool {
							for _, pkg := range w.pkgs_notimport {
								v_j := pkg.(*ast.GenDecl).Specs[0].(*ast.ImportSpec)
								if v_j.Path.Value == value &&
									v_j.Name.String() == name {
									return true
								}
							}
							return false
						}() {
							continue
						}
						var tree []ast.Decl
						if spec.(*ast.ImportSpec).Name == nil {
							tree, _ = parseDeclList(w.files, "gop", "import "+value)
						} else {
							tree, _ = parseDeclList(w.files, "gop", "import "+name+" "+value)
						}
						w.pkgs = append(w.pkgs, tree[0])
					}
					continue
				}
			}

			w.defs = append(w.defs, nil)
			copy(w.defs[pos+1:], w.defs[pos:])
			w.defs[pos] = v[i]
		}
	default:
		err = errors.New("Fatal error: Unknown tree type.")
		return
	}

	var outBuf, errBuf *bytes.Buffer
	err = compile(w)
	if err == nil {
		goto run
	}

	for _, arr := range regexp.MustCompile(`imported and not used: (".+?")( as (.+))?`).FindAllStringSubmatch(err.Error(), -1) {
		name, value := arr[3], arr[1]
		if name == "" {
			name = "<nil>"
		}
		for pos, pkg := range w.pkgs {
			v_j := pkg.(*ast.GenDecl).Specs[0].(*ast.ImportSpec)
			if v_j.Path.Value == value &&
				v_j.Name.String() == name {
				w.pkgs = append(w.pkgs[:pos], w.pkgs[pos+1:]...)
				w.pkgs_notimport = append(w.pkgs_notimport, pkg)
				break
			}
		}
	}

	for _, arr := range regexp.MustCompile(`undefined: (.+)`).FindAllStringSubmatch(err.Error(), -1) {
		name := arr[1]
		for pos, pkg := range w.pkgs_notimport {
			v_j := pkg.(*ast.GenDecl).Specs[0].(*ast.ImportSpec)
			if v_j.Name.String() == name ||
				regexp.MustCompile(`["|/]`+name+`"`).MatchString(v_j.Path.Value) {
				w.pkgs_notimport = append(w.pkgs_notimport[:pos], w.pkgs_notimport[pos+1:]...)
				w.pkgs = append(w.pkgs, pkg)
				break
			}
		}
	}

	err = compile(w)
	if err == nil {
		goto run
	}

	goto restore

run:
	outBuf, errBuf, err = run(w)
	fmt.Print(outBuf)
	fmt.Print(errBuf)
	if err != nil || outBuf.Len() > 0 || errBuf.Len() > 0 {
		goto restore
	}
	return

restore:
	w.pkgs = bkup_pkgs
	w.pkgs_notimport = bkup_pkgs_notimport
	w.codes = bkup_codes
	w.defs = bkup_defs
	w.files = bkup_files
	return
}

func parseGo4import(w *Workspace, line string) (notComplete bool, err error) {
	var tree interface{}
	tree, err = parseDeclList(w.files, "gop", line[0:])
	if err != nil {
		tree, err = parseStmtList(w.files, "gop", line[0:])
		if err != nil {
			if _, ok := err.(scanner.ErrorList); ok {
				err = nil
				notComplete = true
			}
			return
		}
	}

	switch v := tree.(type) {
	case []ast.Stmt:
		for _, e := range v {
			w.codes = append(w.codes, e)
		}
	case []ast.Decl:
		for _, e := range v {
			if v_i, ok := e.(*ast.GenDecl); ok {
				if v_i.Tok == token.IMPORT {
					for _, spec := range v_i.Specs {
						name := spec.(*ast.ImportSpec).Name.String()
						value := spec.(*ast.ImportSpec).Path.Value
						var tree []ast.Decl
						if spec.(*ast.ImportSpec).Name == nil {
							tree, _ = parseDeclList(w.files, "gop", "import "+value)
						} else {
							tree, _ = parseDeclList(w.files, "gop", "import "+name+" "+value)
						}
						w.pkgs = append(w.pkgs, tree[0])
					}
					continue
				}
			}
			w.defs = append(w.defs, e)
		}
	default:
		err = errors.New("Fatal error: Unknown tree type.")
		return
	}

	return
}

func dispatch(w *Workspace, line string) (notComplete bool, err error) {
	line = strings.TrimSpace(line)

	line = execAlias(w, line)
	if line == "" {
		return
	}

	if execSpecial(w, line) {
		return
	}

	switch line[0] {
	case '?':
		fmt.Println("Commands:")
		fmt.Println("\t?|help\thelp menu")
		fmt.Println("\t-[dpc][#],[#]-[#],...\tpop last/specific (declaration|package|code)")
		fmt.Println("\t![!]\tinspect source [with linenum]")
		fmt.Println("\t<tmpl\tsource tmpl")
		fmt.Println("\t>tmpl\twrite tmpl")
		fmt.Println("\t[#](...)\tadd def or code")
		fmt.Println("\treset\treset")
		fmt.Println("\tlist\ttmpl list")
		fmt.Println("\targ\tset or get command-line argument")
	case '-':
		cmd_args := strings.TrimSpace(line[1:])
		removeByIndex(w, cmd_args)
	case '!':
		cmd_args := strings.TrimSpace(line[1:])
		if cmd_args == "!" {
			fmt.Println(w.source(true, true, true))
		} else {
			fmt.Println(w.source(true, false, true))
		}
	default:
		return parseGo(w, line)
	}

	return
}

func main() {
	fmt.Println("Welcome to the Go Partner! [version: 1.7, created by simplejia]")
	fmt.Println("Enter '?' for a list of commands.")

	w := &Workspace{
		files: token.NewFileSet(),
	}
	sourceDefaultDPC(w)

	ifTmplExist, tmplFile := true, "gop.tmpl"
	if _, err := os.Stat(filepath.Join(home, tmplFile)); os.IsNotExist(err) {
		if _, err := os.Stat(tmplFile); os.IsNotExist(err) {
			ifTmplExist = false
		}
	}
	if ifTmplExist {
		dispatch(w, "<"+tmplFile)
	}

	rl := newContLiner()
	defer rl.Close()

	if err := os.MkdirAll(home, 0755); err != nil {
		fmt.Println("Mkdir error: ", err)
		os.Exit(1)
	}

	historyFile := filepath.Join(home, "history")
	if f, err := os.Open(historyFile); err != nil {
		if !os.IsNotExist(err) {
			fmt.Printf("OpenFile %s error: %v\n", historyFile, err)
		}
	} else {
		defer f.Close()
		rl.ReadHistory(f)
	}

	defer func() {
		if f, err := os.Create(historyFile); err != nil {
			fmt.Printf("Open %s error: %v\n", historyFile, err)
		} else {
			rl.WriteHistory(f)
		}
	}()

	for {
		rl.SetWordCompleter(w.completeWord)

		PS1 := "GOP$ "
		in, err := rl.Prompt(PS1)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				fmt.Println("Unexpected error:", err)
				continue
			}
		}

		if in == "" {
			continue
		}

		rl.Reindent()

		notComplete, err := dispatch(w, in)
		if err != nil {
			fmt.Println("Error:", err)
		} else if notComplete {
			continue
		}

		rl.Accepted()
	}
}
