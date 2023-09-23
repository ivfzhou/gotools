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
	"os/exec"
)

// RunShellCommand 交互式运行shell命令bin，args为每次输入的命令。
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
