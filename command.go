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
	"sync"
)

type Command struct {
	cmd            *exec.Cmd
	err            error
	stdout, stderr bytes.Buffer
	readCursor     int64
	writeClose     io.WriteCloser
	exit           chan struct{}
	once           sync.Once
}

// Read 读取标准输出。
func (c *Command) Read() []byte {
	bs := c.stdout.Bytes()[c.readCursor:]
	c.readCursor += int64(len(bs))
	return bs
}

// Write 响应输入。
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
		c.once.Do(func() { close(c.exit) })
		c.cmd.Process.Kill()
	}
	return err
}

// IsExit 是否已运行完毕。
func (c *Command) IsExit() bool {
	select {
	case <-c.exit:
		return true
	default:
		return false
	}
}

// Out 等待运行结束，返回运行结果
func (c *Command) Out() (stdout, stderr []byte, err error) {
	<-c.exit
	return c.stdout.Bytes(), c.stderr.Bytes(), c.err
}

// RunCommand 交互式运行cmd命令。
func RunCommand(cmd string) *Command {
	command := exec.Command(cmd)
	c := &Command{cmd: command, exit: make(chan struct{})}
	command.Stderr = &c.stderr
	command.Stdout = &c.stdout
	c.writeClose, c.err = command.StdinPipe()
	if c.err != nil {
		c.once.Do(func() { close(c.exit) })
		return c
	}
	if c.err = command.Start(); c.err != nil {
		c.once.Do(func() { close(c.exit) })
		return c
	}
	go func() {
		c.err = command.Wait()
		c.once.Do(func() { close(c.exit) })
	}()
	return c
}

// RunCommandAndPrompt 运行命令并响应输入prompts。
func RunCommandAndPrompt(cmd string, prompts ...string) (stdout, stderr []byte, err error) {
	command := RunCommand(cmd)
	for _, v := range prompts {
		if err = command.Write(v); err != nil {
			return
		}
	}
	stdout, stderr, err = command.Out()
	return
}
