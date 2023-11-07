package gotools_test

import (
	"path/filepath"
	"testing"
	"time"

	"gitee.com/ivfzhou/gotools/v3"
)

func TestRunCommandAndPrompt(t *testing.T) {
	stdout, stderr, err := gotools.RunCommandAndPrompt("testdata"+string(filepath.Separator)+"echo", "echo")
	if err != nil {
		t.Error("command: unexpected error", err)
	}
	if string(stderr) != "your input is echo\n" {
		t.Error("command: unexpected stderr", string(stderr))
	}
	if string(stdout) != "test echo\nbegin test\nyour input is echo\n" {
		t.Error("command: unexpected stdout", string(stdout))
	}
}

func TestRunCommand(t *testing.T) {
	command := gotools.RunCommand("testdata/echo")
	times := 0
	for !command.IsExit() {
		bs := command.Read()
		if len(bs) <= 0 {
			time.Sleep(time.Second)
			continue
		}
		if times == 0 {
			if string(bs) != "test echo\nbegin test\n" {
				t.Error(string(bs))
			}
			if err := command.Write("hello"); err != nil {
				t.Error(err)
				return
			}
			times++
			continue
		}
		if times == 1 {
			if string(bs) != "your input is hello\n" {
				t.Error(string(bs))
			}
		}
	}
	stdout, stderr, err := command.Out()
	if err != nil {
		t.Error("command: unexpected error", err)
	}
	if string(stderr) != "your input is hello\n" {
		t.Error("command: unexpected stderr", string(stderr))
	}
	if string(stdout) != "test echo\nbegin test\nyour input is hello\n" {
		t.Error("command: unexpected stdout", string(stdout))
	}
}
