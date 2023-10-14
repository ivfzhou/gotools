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

package gotools

import (
	"bytes"
	"io"
	"os/exec"
	"strings"
	"sync/atomic"
)

type Command struct {
	cmd            *exec.Cmd
	err            error
	end            atomic.Int32
	stdout, stderr bytes.Buffer
	readCursor     int64
	writeClose     io.WriteCloser
}

func (c *Command) Read() []byte {
	bs := c.stdout.Bytes()[c.readCursor:]
	c.readCursor += int64(len(bs))
	return bs
}

func (c *Command) Write(input string) error {
	if c.IsExit() {
		return c.err
	}
	if !strings.HasSuffix(input, "\n") {
		input += "\n"
	}
	_, err := c.writeClose.Write([]byte(input))
	if err != nil {
		c.err = err
		c.end.Store(1)
		c.cmd.Process.Kill()
	}
	return err
}

func (c *Command) IsExit() bool {
	return c.end.Load() > 0
}

func (c *Command) Out() (stdout, stderr []byte, err error) {
	return c.stdout.Bytes(), c.stderr.Bytes(), c.err
}

// RunCommand 交互式运行cmd命令。
func RunCommand(cmd string) *Command {
	command := exec.Command(cmd)
	c := &Command{cmd: command}
	command.Stderr = &c.stderr
	command.Stdout = &c.stdout
	c.writeClose, c.err = command.StdinPipe()
	if c.err != nil {
		c.end.Store(1)
		return c
	}
	if c.err = command.Start(); c.err != nil {
		c.end.Store(1)
		return c
	}
	go func() {
		c.err = command.Wait()
		c.end.Store(1)
	}()
	return c
}
