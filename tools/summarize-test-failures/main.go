// Command summarize-test-failures prints the failing tests in go-junit-report
// XML files as Slack mrkdwn.
//
// Usage: go run ./tools/summarize-test-failures REPORT.xml [REPORT.xml ...]
//
// Missing or malformed report files are ignored: when an earlier CI step
// fails, the later test steps never run and never write their report.
//
// Only the innermost failing tests are listed. A parent test fails whenever
// one of its subtests does, so listing it too would only add noise.
package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

const (
	maxTests  = 10
	maxDetail = 200
)

// detailRe matches the first line of a failure which points at the failing
// assertion or error.
var detailRe = regexp.MustCompile(`_test\.go:\d+:|Error:|panic:`)

type report struct {
	Cases []struct {
		Name    string  `xml:"name,attr"`
		Failure *string `xml:"failure"`
	} `xml:"testsuite>testcase"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: summarize-test-failures REPORT.xml [REPORT.xml ...]")
		os.Exit(2)
	}

	failed := map[string]string{}
	for _, path := range os.Args[1:] {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var r report
		if xml.Unmarshal(data, &r) != nil {
			continue
		}
		for _, c := range r.Cases {
			if c.Failure != nil {
				failed[c.Name] = strings.TrimSpace(*c.Failure)
			}
		}
	}
	fmt.Println(summarize(failed))
}

// summarize formats failing tests, keyed by name to their failure output.
func summarize(failed map[string]string) string {
	var names []string
	for name := range failed {
		if !hasFailedSubtest(name, failed) {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return "No failing tests were recorded, so the failure was probably in the build, " +
			"test setup, or the test run itself (timeout, panic, or crash)."
	}
	sort.Strings(names)

	var out []string
	for i, name := range names {
		if i == maxTests {
			out = append(out, fmt.Sprintf("…and %d more", len(names)-maxTests))
			break
		}
		line := "• `" + name + "`"
		if d := detail(failed[name]); d != "" {
			line += ": " + d
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func hasFailedSubtest(name string, failed map[string]string) bool {
	for other := range failed {
		if strings.HasPrefix(other, name+"/") {
			return true
		}
	}
	return false
}

// detail returns the line of a failure's output which best explains it.
func detail(output string) string {
	var lines []string
	for _, line := range strings.Split(output, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}
	if len(lines) == 0 {
		return ""
	}
	line := lines[0]
	for _, l := range lines {
		if detailRe.MatchString(l) {
			line = l
			break
		}
	}
	if r := []rune(line); len(r) > maxDetail {
		line = string(r[:maxDetail-1]) + "…"
	}
	return line
}
