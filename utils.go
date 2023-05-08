package misc

import (
	"fmt"
	"log"
	"path"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"unsafe"
)

//go:linkname Nanotime runtime.nanotime1
func Nanotime() int64

func String2Bytes(s string) []byte {
	var buf = *(*[]byte)(unsafe.Pointer(&s))
	(*reflect.SliceHeader)(unsafe.Pointer(&buf)).Cap = len(s)
	return buf
}

func Bytes2String(bts []byte) string {
	return *(*string)(unsafe.Pointer(&bts))
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
