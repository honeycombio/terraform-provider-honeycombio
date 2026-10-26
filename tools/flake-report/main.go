// Command flake-report reports which tests fail most often in the scheduled
// CI runs.
//
// Usage: go run ./tools/flake-report [COUNT]
//
// It looks at the last COUNT (default 30) scheduled runs of the CI workflow
// on main, reads the logs of each failed run, and prints how often every test
// failed. A test which fails in some runs and passes in others is a flake.
// Requires the GitHub CLI (gh) with access to this repository.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	workflow = "ci.yaml"
	noTest   = "(no failing test: build, setup, lint, or infrastructure)"
)

var failRe = regexp.MustCompile(`--- FAIL: (\S+)`)

// failure is a test which failed in a CI job.
type failure struct{ test, job string }

func gh(args ...string) string {
	out, err := exec.Command("gh", args...).Output()
	if err != nil {
		log.Fatalf("gh %s: %v", strings.Join(args, " "), err)
	}
	return string(out)
}

func main() {
	count := 30
	if len(os.Args) > 1 {
		var err error
		if count, err = strconv.Atoi(os.Args[1]); err != nil {
			log.Fatalf("invalid COUNT %q", os.Args[1])
		}
	}

	var runs []struct {
		ID         int    `json:"databaseId"`
		Conclusion string `json:"conclusion"`
		CreatedAt  string `json:"createdAt"`
	}
	list := gh("run", "list", "--workflow", workflow, "--event", "schedule", "--branch", "main",
		"--limit", strconv.Itoa(count), "--json", "databaseId,conclusion,createdAt")
	if err := json.Unmarshal([]byte(list), &runs); err != nil {
		log.Fatal(err)
	}

	finished, failedRuns := 0, 0
	dates := map[failure][]string{} // the dates of the runs each test failed in
	for _, run := range runs {
		if run.Conclusion != "success" && run.Conclusion != "failure" {
			continue
		}
		finished++
		if run.Conclusion != "failure" {
			continue
		}
		failedRuns++
		failures := innermost(parseFailures(gh("run", "view", strconv.Itoa(run.ID), "--log-failed")))
		if len(failures) == 0 {
			failures = []failure{{test: noTest}}
		}
		for _, f := range failures {
			dates[f] = append(dates[f], run.CreatedAt[:10])
		}
	}

	fmt.Printf("%d of the last %d scheduled runs failed\n\n", failedRuns, finished)

	keys := make([]failure, 0, len(dates))
	for f := range dates {
		keys = append(keys, f)
	}
	sort.Slice(keys, func(i, j int) bool {
		if a, b := len(dates[keys[i]]), len(dates[keys[j]]); a != b {
			return a > b
		}
		return keys[i].test < keys[j].test
	})
	for _, f := range keys {
		where := ""
		if f.job != "" {
			where = " [" + f.job + "]"
		}
		latest := dates[f][0] // runs are listed newest first
		fmt.Printf("%3d  %s%s  (latest %s)\n", len(dates[f]), f.test, where, latest)
	}
}

// parseFailures returns the failures found in the output of `gh run view
// --log-failed`, whose lines are prefixed with the job name and a tab.
func parseFailures(logs string) []failure {
	seen := map[failure]bool{}
	var failures []failure
	for _, line := range strings.Split(logs, "\n") {
		job, rest, _ := strings.Cut(line, "\t")
		if m := failRe.FindStringSubmatch(rest); m != nil {
			f := failure{test: m[1], job: job}
			if !seen[f] {
				seen[f] = true
				failures = append(failures, f)
			}
		}
	}
	return failures
}

// innermost drops the parents of failing subtests: a parent fails whenever
// one of its subtests does.
func innermost(failures []failure) []failure {
	var out []failure
	for _, f := range failures {
		parent := false
		for _, other := range failures {
			if other.job == f.job && strings.HasPrefix(other.test, f.test+"/") {
				parent = true
				break
			}
		}
		if !parent {
			out = append(out, f)
		}
	}
	return out
}
