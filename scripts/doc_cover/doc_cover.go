// Package main calculates AST documentation coverage across Go source files and packages.
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

// calculateCoverageCount walks srcDir and counts total documentable AST nodes vs documented AST nodes.
// Parameter srcDir is the root directory to walk, and it returns total items, documented items, and any error.
func calculateCoverageCount(srcDir string) (int, int, error) {
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

		// File / Package documentation
		total++
		if f.Doc != nil && strings.TrimSpace(f.Doc.Text()) != "" {
			docs++
		}

		for _, decl := range f.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				if d.Tok == token.TYPE || d.Tok == token.VAR || d.Tok == token.CONST {
					for _, spec := range d.Specs {
						switch s := spec.(type) {
						case *ast.TypeSpec:
							total++
							if d.Doc != nil || s.Doc != nil {
								docs++
							}
							if st, ok := s.Type.(*ast.StructType); ok && st.Fields != nil {
								for _, field := range st.Fields.List {
									if len(field.Names) == 0 {
										continue
									}
									total++
									if field.Doc != nil || field.Comment != nil {
										docs++
									}
								}
							} else if iface, ok := s.Type.(*ast.InterfaceType); ok && iface.Methods != nil {
								for _, method := range iface.Methods.List {
									if len(method.Names) == 0 {
										continue
									}
									total++
									if method.Doc != nil || method.Comment != nil {
										docs++
									}
								}
							}
						case *ast.ValueSpec:
							if len(s.Names) > 0 {
								total++
								if d.Doc != nil || s.Doc != nil {
									docs++
								}
							}
						}
					}
				}
			case *ast.FuncDecl:
				total++
				if d.Doc != nil {
					docs++
				}
			}
		}
		return nil
	})

	if err != nil {
		return 0, 0, err
	}

	return total, docs, nil
}

// calculateCoverage computes the percentage of documented AST nodes in srcDir.
// Parameter srcDir is the root directory to analyze, and it returns the coverage percentage and any error.
func calculateCoverage(srcDir string) (float64, error) {
	total, docs, err := calculateCoverageCount(srcDir)
	if err != nil {
		return 0.0, err
	}
	if total == 0 {
		return 100.0, nil
	}
	return float64(docs) / float64(total) * 100.0, nil
}

// osExit is an injectable function variable wrapping os.Exit.
var osExit = os.Exit

// runMain calculates and prints documentation coverage for a single directory.
// Parameter srcDir is the target directory path.
func runMain(srcDir string) {
	total, docs, err := calculateCoverageCount(srcDir)
	if err != nil {
		fmt.Println("0.0%")
		osExit(1)
		return
	}
	if total == 0 {
		fmt.Println("100.0%")
		return
	}
	fmt.Println(fmt.Sprintf("%.1f%%", float64(docs)/float64(total)*100.0))
}

// main calculates aggregate documentation coverage across src, cmd, cdd, and scripts.
func main() {
	var total, docs int
	dirs := []string{"src", "cmd", "cdd", "scripts"}
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}
		t, d, err := calculateCoverageCount(dir)
		if err != nil {
			fmt.Println("0.0%")
			osExit(1)
			return
		}
		total += t
		docs += d
	}
	if total == 0 {
		fmt.Println("100.0%")
		return
	}
	fmt.Println(fmt.Sprintf("%.1f%%", float64(docs)/float64(total)*100.0))
	if docs < total {
		osExit(1)
	}
}
