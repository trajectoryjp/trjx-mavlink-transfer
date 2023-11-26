package fileio

import (
	"bufio"
	"fmt"
	"os"
)

func ReadPassword() string {
	file, err := os.Open("setting/password.txt")
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		return line
	}
	return ""
}

func WritePassword(password string) {
	file, err := os.OpenFile("setting/password.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return
	}
	defer file.Close()

	fmt.Fprintln(file, password)
}
