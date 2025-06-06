package core

import (
	"bufio"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

func RandomStr(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz1234567890")
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
func RandInt64(min, max int64) int64 {
	if min >= max || min == 0 || max == 0 {
		return max
	}
	return rand.Int63n(max-min) + min
}

// LinesInFile 读取文件 返回每行的数组
func LinesInFile(fileName string) ([]string, error) {
	result := []string{}
	f, err := os.Open(fileName)
	if err != nil {
		return result, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			result = append(result, line)
		}
	}
	return result, nil
}

// LinesReaderInFile 读取文件，返回行数
func LinesReaderInFile(filename string) (int, error) {
	f, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	// 使用更大的缓冲区减少IO操作
	buf := make([]byte, 32*1024)
	count := 0

	for {
		readSize, err := f.Read(buf)
		if readSize == 0 {
			break
		}

		// 直接遍历缓冲区计数换行符
		for i := 0; i < readSize; i++ {
			if buf[i] == '\n' {
				count++
			}
		}

		if err != nil {
			if err == io.EOF {
				// 处理文件末尾没有换行符的情况
				if readSize > 0 && (count == 0 || buf[readSize-1] != '\n') {
					count++
				}
				return count, nil
			}
			return count, err
		}
	}

	// 处理空文件或只有一行没有换行符的文件
	if count == 0 {
		count = 1
	}

	return count, nil
}

func FileExists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func GetWindowWith() int {
	w, _, err := terminal.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 0
	}
	return w
}

func IsContain(items []string, item string) bool {
	for _, eachItem := range items {
		if eachItem == item {
			return true
		}
	}
	return false
}

func SliceToString(items []string) string {
	ret := strings.Builder{}
	ret.WriteString("[")
	ret.WriteString(strings.Join(items, ","))
	ret.WriteString("]")
	return ret.String()
}
func HasStdin() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	if fi.Mode()&os.ModeNamedPipe == 0 {
		return false
	}
	return true
}
