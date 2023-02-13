package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func run(branch string) {
	// find all tree on this branch that are not on the other branch
	trees := map[string]string{}

	stdout, err := exec.Command("git", "log", "--format=%H %T", "^"+branch, "HEAD").Output()
	assertNoError(err)

	for _, line := range strings.Split(string(stdout), "\n") {
		if commit, tree, found := strings.Cut(line, " "); found {
			trees[tree] = commit
		}
	}

	// go through commit on other side, find first matching commit
	stdout, err = exec.Command("git", "log", "--format=%H %T", branch, "^HEAD").Output()
	assertNoError(err)
	for _, line := range strings.Split(string(stdout), "\n") {
		if commit, tree, found := strings.Cut(line, " "); found {
			if _, ok := trees[tree]; ok {
				msg := fmt.Sprintln()
				msg += fmt.Sprintln("Post-squash merge of", branch)
				msg += fmt.Sprintln("")
				msg += fmt.Sprintln("Commit", commit[:7], "on", branch, "has the same tree as")
				msg += fmt.Sprintln("commit", trees[tree][:7], ".")

				err := exec.Command("git", "merge", "-s", "ours", commit, "-m", msg).Run()
				assertNoError(err)
				os.Exit(0)
			}
		}
	}

	println("Could not find a suitable squash merge commit on", branch)
	os.Exit(1)
}

func assertNoError(err error) {
	if err == nil {
		return
	}
	if exit, ok := err.(*exec.ExitError); ok {
		if status, ok := exit.Sys().(syscall.WaitStatus); ok {
			log.Printf("Exit Status: %d", status.ExitStatus())
		}
	} else {
		log.Fatal(err)
	}
}

func main() {
	flag.Usage = func() {
		fmt.Println("usage: git-post-squash [-h] branch")
		fmt.Println()
		fmt.Println("Post-squash merge command")
		fmt.Println()
		fmt.Println("positional arguments:")
		fmt.Println("  branch      Branch that contains the squash-merge commit (usually master)")
		fmt.Println()
		fmt.Println("optional arguments:")
		fmt.Println("  -h, --help  show this help message and exit")
	}
	flag.Parse()
	branch := flag.Arg(0)
	if branch == "" {
		println("git-post-squash: error: the following arguments are required: branch")
		os.Exit(1)
	}

	run(branch)
}
