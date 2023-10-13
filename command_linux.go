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
	"bufio"
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

// RunShellCommand 交互式运行shell命令bin，args为每次输入的命令。
// Deprecated: 使用RunCommand替代。
func RunShellCommand(bin string, args ...string) (std []byte, estd []byte, err error) {
	cmd := exec.Command("sh")
	writer, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, err
	}
	reader, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	errReader, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, err
	}

	cmdCh := make(chan string)
	closeCh := make(chan struct{})
	output := bytes.Buffer{}
	eoutput := bytes.Buffer{}
	go func() {
		defer errReader.Close()
		scanner := bufio.NewScanner(errReader)
		for scanner.Scan() {
			s := scanner.Bytes()
			eoutput.Write(s)
			select {
			case <-closeCh:
				return
			default:
			}
		}
	}()
	go func() {
		defer reader.Close()
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			s := scanner.Bytes()
			output.Write(s)
			select {
			case <-closeCh:
				return
			default:
			}
			if len(args) <= 0 {
				cmdCh <- "exit"
				close(closeCh)
				close(cmdCh)
			} else {
				cmdCh <- args[0]
				args = args[1:]
			}
		}
	}()
	go func() {
		defer writer.Close()
		_, _ = writer.Write([]byte(bin + "\n"))
		for {
			select {
			case s := <-cmdCh:
				_, _ = writer.Write([]byte(s + "\n"))
			case <-closeCh:
				return
			}
		}
	}()
	if err = cmd.Start(); err != nil {
		return nil, nil, err
	}
	err = cmd.Wait()
	return output.Bytes(), eoutput.Bytes(), err
}
