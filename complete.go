package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"go/build"

	"github.com/motemen/gore/gocode"
)

func completeImport(w *Workspace, prefix string) []string {
	result := []string{}
	seen := map[string]bool{}

	dirs := build.Default.SrcDirs()
	for bpos, epos := 0, len(dirs); bpos < epos; bpos++ {
		pkgdir := dirs[bpos] + "/pkg"
		if fi, err := os.Stat(pkgdir); err != nil || !fi.IsDir() {
			continue
		}
		dirs = append(dirs, pkgdir)
	}

	d, n := path.Split(prefix)
	for _, srcDir := range dirs {
		dir := filepath.Join(srcDir, d)

		if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
			if err != nil && !os.IsNotExist(err) {
				fmt.Printf("Stat %s: %s\n", dir, err)
			}
			continue
		}

		entries, err := ioutil.ReadDir(dir)
		if err != nil {
			fmt.Printf("ReadDir %s: %s\n", dir, err)
			continue
		}
		for _, fi := range entries {
			if !fi.IsDir() {
				continue
			}

			name := fi.Name()
			if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
				continue
			}

			if strings.HasPrefix(name, n) {
				r := path.Join(d, name)
				if !seen[r] {
					result = append(result, r)
					seen[r] = true
				}
			}
		}
	}

	return result
}

func completeTmpl(w *Workspace, prefix string) (result []string) {
	entries, err := ioutil.ReadDir(home)
	if err != nil {
		fmt.Printf("ReadDir %s: %s\n", home, err)
		return
	}

	for _, fi := range entries {
		if fi.IsDir() {
			continue
		}

		name := fi.Name()
		if strings.HasPrefix(name, ".") ||
			!strings.HasSuffix(name, ".tmpl") {
			continue
		}

		if strings.HasPrefix(name, prefix) {
			result = append(result, name)
		}
	}
	return
}

func completeCode(source, in string, pos int) (keep int, candidates []string, err error) {
	p := strings.LastIndex(source, "}")
	if p == -1 {
		err = errors.New("Unexpected error!")
		return
	}
	editingSource := source[0:p] + in[:pos] + source[p:]
	result, err := gocode.Query([]byte(editingSource), p+pos)
	if err != nil {
		return
	}

	keep = pos - result.Cursor
	for _, e := range result.Candidates {
		cand := e.Name
		if e.Class == "func" {
			cand = cand + "("
		}
		candidates = append(candidates, cand)
	}

	return
}

func (w *Workspace) completeWord(line string, pos int) (string, []string, string) {
	if strings.HasPrefix(line, "import ") {
		i := strings.Index(line, "\"")
		if i != -1 {
			if pos > i {
				return line[:i+1], completeImport(w, line[i+1:pos]), line[pos:]
			}
		}
	}

	for _, cmdPrefix := range []string{"<", ">"} {
		if strings.HasPrefix(line, cmdPrefix) {
			i := strings.Index(line, cmdPrefix)
			if i != -1 {
				if pos > i {
					return line[:i+1], completeTmpl(w, line[i+1:pos]), line[pos:]
				}
			}
		}
	}

	if gocode.Available() == false {
		return "", nil, ""
	}

	oldPos := pos
	pos, cands, err := completeCode(w.source(false, false, true), line, pos)
	if err != nil {
		return "", nil, ""
	}

	return line[0:pos], cands, line[oldPos:]
}
