package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
)

func getColor(pct float64) string {
	if pct >= 90 {
		return "brightgreen"
	}
	if pct >= 80 {
		return "green"
	}
	if pct >= 70 {
		return "yellowgreen"
	}
	if pct >= 60 {
		return "yellow"
	}
	if pct >= 50 {
		return "orange"
	}
	return "red"
}

func main() {
	_, err := os.Stat("README.md")
	if os.IsNotExist(err) {
		return
	}

	testCov := 0.0
	cmd := exec.Command("go", "test", "-coverprofile=coverage.out", "./...")
	cmd.Run() // Ignore error, tests might fail

	cmd = exec.Command("go", "tool", "cover", "-func=coverage.out")
	out, err := cmd.CombinedOutput()
	if err == nil {
		re := regexp.MustCompile(`total:\s+\(statements\)\s+([0-9.]+)%`)
		match := re.FindStringSubmatch(string(out))
		if len(match) > 1 {
			testCov, _ = strconv.ParseFloat(match[1], 64)
		}
	} else {
		fmt.Printf("Test coverage calculation failed: %v\n", err)
	}

	docCov := 0.0
	docCmd := exec.Command("go", "run", filepath.Join("scripts", "doc_cover", "doc_cover.go"))
	docOut, err := docCmd.CombinedOutput()
	if err == nil {
		docRe := regexp.MustCompile(`([0-9.]+)%`)
		docMatch := docRe.FindStringSubmatch(string(docOut))
		if len(docMatch) > 1 {
			docCov, _ = strconv.ParseFloat(docMatch[1], 64)
		}
	} else {
		fmt.Printf("Doc coverage calculation failed: %v\n", err)
	}

	testCovStr := fmt.Sprintf("%.1f", testCov)
	if float64(int(testCov)) == testCov {
		testCovStr = fmt.Sprintf("%d", int(testCov))
	}

	docCovStr := fmt.Sprintf("%.1f", docCov)
	if float64(int(docCov)) == docCov {
		docCovStr = fmt.Sprintf("%d", int(docCov))
	}

	testColor := getColor(testCov)
	docColor := getColor(docCov)

	content, err := ioutil.ReadFile("README.md")
	if err != nil {
		fmt.Printf("Failed to read README.md: %v\n", err)
		return
	}

	strContent := string(content)

	testRe := regexp.MustCompile(`\[\!\[Test Coverage\]\(https://img\.shields\.io/badge/test_coverage-[0-9.]+%25-[a-z]+\.svg\)\]\(#\)`)
	strContent = testRe.ReplaceAllString(strContent, fmt.Sprintf("[![Test Coverage](https://img.shields.io/badge/test_coverage-%s%%25-%s.svg)](#)", testCovStr, testColor))

	docRe := regexp.MustCompile(`\[\!\[Doc Coverage\]\(https://img\.shields\.io/badge/doc_coverage-[0-9.]+%25-[a-z]+\.svg\)\]\(#\)`)
	strContent = docRe.ReplaceAllString(strContent, fmt.Sprintf("[![Doc Coverage](https://img.shields.io/badge/doc_coverage-%s%%25-%s.svg)](#)", docCovStr, docColor))

	err = ioutil.WriteFile("README.md", []byte(strContent), 0644)
	if err != nil {
		fmt.Printf("Failed to write README.md: %v\n", err)
	}
}
