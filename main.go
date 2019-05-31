package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

func runCommand(cmdWithArgs ...string) (string, error) {
	ret, err := exec.Command(cmdWithArgs[0], cmdWithArgs[1:]...).Output()
	if err != nil {
		return "", err
	}

	return string(ret), nil
}

func getCFlags(name string) ([]string, error) {
	line, err := runCommand("pkg-config", "--cflags", name)
	if err != nil {
		return nil, fmt.Errorf("can't run pkg-config for %q: %v", name, err)
	}

	return strings.Fields(strings.TrimSpace(line)), nil
}

var tempDir string

func cleanupTempDir() {
	if tempDir == "" {
		return
	}

	if err := os.RemoveAll(tempDir); err != nil {
		log.Printf("Warning: Can't remove temporary directory %q: %v", tempDir, err)
	}
}

func main() {
	var err error

	if tempDir, err = ioutil.TempDir("", "go-wireshark"); err != nil {
		log.Fatalf("Can't create temporary directory: %v", err)
	}
	defer cleanupTempDir()

	const inputFile = "/usr/include/wireshark/epan/packet.h"

	cflags, err := getCFlags("wireshark")
	if err != nil {
		log.Fatalf("Can't get CFLAGS: %v", err)
	}

	tmpf := filepath.Join(tempDir, filepath.Base(inputFile))

	clangCmd := append([]string{"clang", "-E"}, cflags...)
	clangCmd = append(clangCmd, "-o", tmpf, inputFile)

	_, err = runCommand(clangCmd...)
	if err != nil {
		log.Fatalf("Can't run clang for preprocessing: %v", err)
	}

	spew.Dump(getFunDecls(tmpf))
}
