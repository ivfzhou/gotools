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
	"regexp"
)

var psConsolePrefixReg = regexp.MustCompile(`^PS [A-Z]:(\\[^\\/:*?"<>|]*?)*?> `)

// RunPSCommand 交互式运行PowerShell命令bin，std 是标准输出数据，estd 是标准错误输出数据。
func RunPSCommand(bin string, interactive ...string) (std []byte, estd []byte, err error) {
	cmd := exec.Command("PowerShell.exe", "-NoLogo", "-NoExit")
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

	runCmd := make(map[string]struct{}, len(interactive)+2)
	runCmd["exit"] = struct{}{}
	runCmd[bin] = struct{}{}
	for _, v := range interactive {
		runCmd[v] = struct{}{}
	}

	go func() {
		defer errReader.Close()
		scanner := bufio.NewScanner(errReader)
		for scanner.Scan() {
			s := scanner.Bytes()
			if psConsolePrefixReg.Match(s) {
				s = psConsolePrefixReg.ReplaceAll(s, nil)
			}
			if _, ok := runCmd[string(s)]; ok {
				s = nil
			}
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
			if psConsolePrefixReg.Match(s) {
				s = psConsolePrefixReg.ReplaceAll(s, nil)
			}
			if _, ok := runCmd[string(s)]; ok {
				s = nil
			}
			output.Write(s)
			select {
			case <-closeCh:
				return
			default:
			}
			if len(interactive) <= 0 {
				cmdCh <- "exit"
				close(closeCh)
				close(cmdCh)
			} else {
				cmdCh <- interactive[0]
				interactive = interactive[1:]
			}
		}
	}()
	go func() {
		defer writer.Close()
		_, _ = writer.Write([]byte(bin + "\r\n"))
		for {
			select {
			case s := <-cmdCh:
				_, _ = writer.Write([]byte(s + "\r\n"))
			case <-closeCh:
				return
			}
		}
	}()
	if err = cmd.Start(); err != nil {
		return nil, nil, err
	}
	err = cmd.Wait()
	return eoutput.Bytes(), output.Bytes(), err
}

func RunCmdCommand(bin string, args ...string) (std []byte, estd []byte, err error) {
	cmd := exec.Command("cmd.exe", "/Q", "/S", "/K")
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
	re := regexp.MustCompile(`^[A-Z]:(\\[^\\/:*?"<>|]*?)*?>`)
	go func() {
		defer errReader.Close()
		scanner := bufio.NewScanner(errReader)
		for scanner.Scan() {
			s := scanner.Bytes()
			if re.Match(s) {
				s = re.ReplaceAll(s, nil)
			}
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
			if re.Match(s) {
				s = re.ReplaceAll(s, nil)
			}
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
		_, _ = writer.Write([]byte(bin + "\r\n"))
		for {
			select {
			case s := <-cmdCh:
				_, _ = writer.Write([]byte(s + "\r\n"))
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
