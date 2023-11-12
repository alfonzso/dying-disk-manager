package input

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

func Manager(filename string) string {

	var in io.Reader
	if filename != "" {
		log.Debug("filename: ", filename)
		f, err := os.Open(filename)
		if err != nil {
			fmt.Println("[ ERROR ] cannot open file: err:", err)
			os.Exit(1)
		}
		defer f.Close()
		in = f
	} else {
		in = os.Stdin
	}

	if read := readFromFile(filename, in); len(read) > 0 {
		return read
	}

	if read := readFromStdin(in); len(read) > 0 {
		return read
	}

	fmt.Println("[ ERROR ] no inpunt given")
	os.Exit(1)

	return ""

}

func readFromFile(filename string, in io.Reader) string {
	if filename != "" {
		log.Debug("reading from file: ", filename)
		return readFromInput(in)
	}
	return ""
}

func readFromStdin(in io.Reader) string {
	if stat, _ := os.Stdin.Stat(); (stat.Mode() & os.ModeCharDevice) == 0 {
		log.Debug("reading from stdin")
		return readFromInput(in)
	}
	return ""
}

func readFromInput(in io.Reader) string {
	buf := bufio.NewScanner(in)

	var sss []string
	for buf.Scan() {
		sss = append(sss, buf.Text())
	}

	if err := buf.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error reading: err:", err)
	}

	return strings.Join(sss, "\n")
}
