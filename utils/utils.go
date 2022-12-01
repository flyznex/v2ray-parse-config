package utils

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

var (
	vmessPrefix             = "vmess://"
	ErrorInvalidFormatVMess = errors.New("invalid format vmess")
)

type Parser interface {
	GenConfig(input string) error
	GetName() string
}

func CheckVmessConfigValid(s string) error {
	if !strings.Contains(s, vmessPrefix) {
		return ErrorInvalidFormatVMess
	}
	return nil
}

func ReadUpdatedAt(path string) string {
	f, err := os.Open(path)
	if err != nil {
		fmt.Println("[ERROR] Get updated time got error", err.Error())
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var line int
	for scanner.Scan() {
		if line == 0 {
			return scanner.Text()[19:]
		}
	}
	return ""
}
