package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func calculateCoverage(srcDir string) (float64, error) {
	fset := token.NewFileSet()
	var total, docs int

	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".go") || strings.HasSuffix(info.Name(), "_test.go") {
			return nil
		}

		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return err
		}

		for _, decl := range f.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				if d.Tok == token.TYPE || d.Tok == token.VAR || d.Tok == token.CONST {
					for _, spec := range d.Specs {
						switch s := spec.(type) {
						case *ast.TypeSpec:
							if s.Name.IsExported() {
								total++
								if d.Doc != nil {
									docs++
								}
							}
						case *ast.ValueSpec:
							for _, name := range s.Names {
								if name.IsExported() {
									total++
									if d.Doc != nil {
										docs++
									}
									break
								}
							}
						}
					}
				}
			case *ast.FuncDecl:
				if d.Name.IsExported() {
					total++
					if d.Doc != nil {
						docs++
					}
				}
			}
		}
		return nil
	})

	if err != nil {
		return 0.0, err
	}

	if total == 0 {
		return 100.0, nil
	}

	return float64(docs) / float64(total) * 100.0, nil
}

var osExit = os.Exit

func runMain(srcDir string) {
	coverage, err := calculateCoverage(srcDir)
	if err != nil {
		fmt.Println("0.0%")
		osExit(1)
		return
	}
	fmt.Printf("%.1f%%\n", coverage)
}

func main() {
	runMain("src")
}
