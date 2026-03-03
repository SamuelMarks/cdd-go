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

func main() {
	fset := token.NewFileSet()
	var total, docs int

	err := filepath.Walk("src", func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".go") || strings.HasSuffix(info.Name(), "_test.go") {
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
		fmt.Println("0.0%")
		os.Exit(1)
	}

	if total == 0 {
		fmt.Println("100.0%")
		return
	}

	fmt.Printf("%.1f%%\n", float64(docs)/float64(total)*100.0)
}
