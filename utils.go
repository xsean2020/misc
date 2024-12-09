package misc

import (
	"fmt"
	"log"
	"path"
	"runtime"
	"strings"
	"sync"
	"unsafe"
)

var JSONEscape = strings.NewReplacer(
	"\f", "",
	"\b", "",
	"\n", "",
	"\r", "",
	"\t", "",
	`"`, `\"`,
	`\`, `\\`).Replace

//go:linkname Nanotime runtime.nanotime1
func Nanotime() int64

func String2Bytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func Bytes2String(bts []byte) string {
	if len(bts) == 0 {
		return ""
	}
	return unsafe.String(&bts[0], len(bts))
}

func Array2String[T any](array []T, delim string) string {
	return strings.Trim(strings.Replace(fmt.Sprint(array), " ", delim, -1), "[]")
}

var (
	m sync.Map
)

func Caller(skip int) (runtime.Frame, bool) {
	rpc := [1]uintptr{}
	n := runtime.Callers(skip+1, rpc[:])
	var (
		frame runtime.Frame
	)
	if n < 1 {
		return frame, false
	}

	if item, ok := m.Load(rpc[0]); ok {
		frame = item.(runtime.Frame)
	} else {
		frame, _ = runtime.CallersFrames(rpc[:]).Next()
		m.Store(rpc[0], frame)
	}
	return frame, true
}

// 打印调用栈
func StackInfo(skip int) string {
	var buider = new(strings.Builder)
	i := skip
	blanks := "  "
	for {
		frame, ok := Caller(i)
		if !ok {
			break
		}
		buider.WriteString(fmt.Sprintf("%s%s:%d, func: %s\n", blanks, frame.File, frame.Line, path.Base(frame.Function)))
		blanks += "  "
		i++
	}
	return buider.String()
}

func PrintPanicStack(recv func(string)) {
	if x := recover(); x != nil {
		const skip = 3
		var buider = new(strings.Builder)
		buider.WriteString(fmt.Sprintf("Err [%v] call stack: \n", x))
		i := skip
		blanks := "  "
		for {
			frame, ok := Caller(i)
			if !ok {
				break
			}
			buider.WriteString(fmt.Sprintf("%s%s:%d, func: %s\n", blanks, frame.File, frame.Line, path.Base(frame.Function)))
			blanks += "  "
			i++
		}
		msg := buider.String()
		if recv != nil {
			recv(msg)
		} else {
			log.Println(msg)
		}
	}
}

func ToCamelCase(input string) string {
	// 将输入字符串按下划线分隔成切片
	words := strings.Split(input, "_")
	// 对每个单词进行处理
	for i := 0; i < len(words); i++ {
		// 将单词的首字母转换为大写，其他字母保持不变
		words[i] = strings.Title(words[i])
	}
	// 将处理后的单词拼接成驼峰式名称
	return strings.Join(words, "")
}
