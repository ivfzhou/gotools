/*
 * Copyright (c) 2023 ivfzhou
 * gotools is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

package gotools_test

import (
	"testing"
	"time"

	"gitee.com/ivfzhou/gotools"
)

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
		t.Error(err)
	}
	if string(stderr) != "your input is hello\n" {
		t.Error(string(stderr))
	}
	if string(stdout) != "test echo\nbegin test\nyour input is hello\n" {
		t.Error(string(stdout))
	}
}
