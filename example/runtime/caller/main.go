package main

import (
	"fmt"
	"runtime"
)

func main() {
	for skip := 0; ; skip++ {
		pc, file, line, ok := runtime.Caller(skip)
		if !ok {
			break
		}
		fmt.Printf("skip = %v, pc = %v, file = %v, line = %v\n",
			skip, pc, file, line)
	}
}

// 输出：
// 	skip = 0, pc = 6216648, file = D:/golang/workspace/src/github.com/tiandh987/errors/example/runtime/caller/main.go, line = 10
//	skip = 1, pc = 5854902, file = C:/Program Files/Go/src/runtime/proc.go, line = 255
//	skip = 2, pc = 6015488, file = C:/Program Files/Go/src/runtime/asm_amd64.s, line = 1581

// func Caller(skip int) (pc uintptr, file string, line int, ok bool)
//	作用：
//		返回函数调用栈的某一层的程序计数器、文件信息、行号。
//
//	参数：
//		skip 是要提升的堆栈帧数，
//		0 代表当前函数，也就是调用 runtime.Caller 的函数；
//		1 代表上一层调用者；
//		.... 以此类推
//
// 	返回值：
// 		pc   是 uintptr 这个返回的是函数指针
// 		file 是函数所在文件名目录
// 		line 所在行号
// 		ok   是否可以获取到信息