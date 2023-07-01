package gotools

import (
	"bufio"
	"bytes"
	"os/exec"
	"regexp"
)

func RunPSCommand(bin string, args ...string) ([]byte, error) {
	cmd := exec.Command("PowerShell.exe", "-NoLogo", "-NoExit")
	writer, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	reader, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	errReader, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	cmdCh := make(chan string)
	closeCh := make(chan struct{})
	re := regexp.MustCompile(`^PS [A-Z]:(\\[^\\/:*?"<>|]*?)*?> `)
	output := bytes.Buffer{}

	ranCmd := make(map[string]struct{}, len(args)+2)
	ranCmd["exit"] = struct{}{}
	ranCmd[bin] = struct{}{}
	for _, v := range args {
		ranCmd[v] = struct{}{}
	}

	go func() {
		defer errReader.Close()
		scanner := bufio.NewScanner(errReader)
		for scanner.Scan() {
			s := scanner.Bytes()
			if re.Match(s) {
				s = re.ReplaceAll(s, nil)
			}
			if _, ok := ranCmd[string(s)]; ok {
				s = nil
			}
			output.Write(s)
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
			if _, ok := ranCmd[string(s)]; ok {
				s = nil
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
		return nil, err
	}
	err = cmd.Wait()
	return output.Bytes(), err
}

func RunCmdCommand(bin string, args ...string) ([]byte, error) {
	cmd := exec.Command("cmd.exe", "/Q", "/S", "/K")
	writer, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	reader, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	errReader, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	cmdCh := make(chan string)
	closeCh := make(chan struct{})
	output := bytes.Buffer{}
	re := regexp.MustCompile(`^[A-Z]:(\\[^\\/:*?"<>|]*?)*?>`)
	go func() {
		defer errReader.Close()
		scanner := bufio.NewScanner(errReader)
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
		return nil, err
	}
	err = cmd.Wait()
	return output.Bytes(), err
}

func RunShellCommand(bin string, args ...string) ([]byte, error) {
	cmd := exec.Command("sh")
	writer, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	reader, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	errReader, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	cmdCh := make(chan string)
	closeCh := make(chan struct{})
	output := bytes.Buffer{}
	go func() {
		defer errReader.Close()
		scanner := bufio.NewScanner(errReader)
		for scanner.Scan() {
			s := scanner.Bytes()
			output.Write(s)
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
		return nil, err
	}
	err = cmd.Wait()
	return output.Bytes(), err
}
