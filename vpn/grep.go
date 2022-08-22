package vpn

import (
	"bufio"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func findPattern(file *os.File, pattern string) []string {
	fileReader := bufio.NewReader(file)
	lines := 0
	lineIdx := 0
	match := []string{}

	for {
		line, err := fileReader.ReadString('\n')
		lineIdx++
		if strings.Contains(line, pattern) {
			match = append(match, line)
			lines++
			continue
		}
		if err == io.EOF {
			return match
		}
	}
}

func find(pattern string, filenames []string) []string {
	match := []string{}
	for _, filename := range filenames {
		file, err := os.OpenFile(filename, os.O_RDONLY, os.ModePerm)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		for _, s := range findPattern(file, pattern) {
			match = append(match, s)
		}

	}
	return match
}

func findInDir(pattern string, dirnames []string) []string {
	match := []string{}
	for _, dirname := range dirnames {
		files, err := ioutil.ReadDir(dirname)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		for _, file := range files {
			if file.IsDir() {
				m := findInDir(pattern, []string{filepath.Join(dirname, file.Name())})
				for _, s := range m {
					match = append(match, s)
				}
				continue
			}
			filePath := filepath.Join(dirname, file.Name())
			for _, s := range find(pattern, []string{filePath}) {
				match = append(match, s)
			}
		}
	}
	return match
}
