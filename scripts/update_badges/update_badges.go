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

func parseCoverage(out string, regex string) float64 {
	re := regexp.MustCompile(regex)
	match := re.FindStringSubmatch(out)
	if len(match) > 1 {
		val, _ := strconv.ParseFloat(match[1], 64)
		return val
	}
	return 0.0
}

func formatCoverage(cov float64) string {
	str := fmt.Sprintf("%.1f", cov)
	if float64(int(cov)) == cov {
		str = fmt.Sprintf("%d", int(cov))
	}
	return str
}

func getTestCov() float64 {
	cmd := exec.Command("go", "test", "-coverprofile=coverage.out", "./cmd/...", "./src/...", "./cdd/...")
	cmd.Run() // Ignore error, tests might fail

	cmd = exec.Command("go", "tool", "cover", "-func=coverage.out")
	out, err := cmd.CombinedOutput()
	if err == nil {
		return parseCoverage(string(out), `total:\s+\(statements\)\s+([0-9.]+)%`)
	}
	return 0.0
}

func getDocCov() float64 {
	docCmd := exec.Command("go", "run", filepath.Join("scripts", "doc_cover", "doc_cover.go"))
	docOut, err := docCmd.CombinedOutput()
	if err == nil {
		return parseCoverage(string(docOut), `([0-9.]+)%`)
	}
	return 0.0
}

func replaceReadme(content []byte, testCovStr, testColor, docCovStr, docColor string) []byte {
	strContent := string(content)

	testRe := regexp.MustCompile(`\[\!\[Test Coverage\]\(https://img\.shields\.io/badge/test_coverage-[0-9.]+%25-[a-z]+\.svg\)\]\(#\)`)
	strContent = testRe.ReplaceAllString(strContent, fmt.Sprintf("[![Test Coverage](https://img.shields.io/badge/test_coverage-%s%%25-%s.svg)](#)", testCovStr, testColor))

	docRe := regexp.MustCompile(`\[\!\[Doc Coverage\]\(https://img\.shields\.io/badge/doc_coverage-[0-9.]+%25-[a-z]+\.svg\)\]\(#\)`)
	strContent = docRe.ReplaceAllString(strContent, fmt.Sprintf("[![Doc Coverage](https://img.shields.io/badge/doc_coverage-%s%%25-%s.svg)](#)", docCovStr, docColor))

	return []byte(strContent)
}

func main() {
	_, err := os.Stat("README.md")
	if os.IsNotExist(err) {
		return
	}

	testCov := getTestCov()
	docCov := getDocCov()

	testCovStr := formatCoverage(testCov)
	docCovStr := formatCoverage(docCov)

	testColor := getColor(testCov)
	docColor := getColor(docCov)

	content, err := ioutil.ReadFile("README.md")
	if err != nil {
		return
	}

	newContent := replaceReadme(content, testCovStr, testColor, docCovStr, docColor)

	ioutil.WriteFile("README.md", newContent, 0644)
}
