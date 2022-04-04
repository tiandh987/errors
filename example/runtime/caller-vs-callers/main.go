package main

import (
	"fmt"
	"runtime"
)

// runtime.Caller 与 runtime.Callers 有哪些关联和差异？

func main() {
	// Caller
	for skip := 0; ; skip++ {
		pc, file, line, ok := runtime.Caller(skip)
		if !ok {
			break
		}
		fmt.Printf("skip = %v, pc = %v, file = %v, line = %v\n",
			skip, pc, file, line)
	}

	// Callers
	pc := make([]uintptr, 1024)
	for skip := 0; ; skip++ {
		n := runtime.Callers(skip, pc)
		if n <= 0 {
			break
		}
		fmt.Printf("n = %d, skip = %v, pc = %v\n", n, skip, pc[:n])
	}
}

// Caller 输出：
//	skip = 0, pc = 1367001, file = D:/golang/workspace/src/github.com/tiandh987/errors/example/runtime/caller-vs-callers/main.go, line = 13
//	skip = 1, pc = 1005238, file = C:/Program Files/Go/src/runtime/proc.go, line = 255
//	skip = 2, pc = 1165824, file = C:/Program Files/Go/src/runtime/asm_amd64.s, line = 1581

// Callers 输出：
//	n = 4, skip = 0, pc = [1367301 1367274 1005239 1165825]
// 	n = 3, skip = 1, pc = [1367274 1005239 1165825]
//	n = 2, skip = 2, pc = [1005239 1165825]
//	n = 1, skip = 3, pc = [1165825]
// (由于历史原因，如果 Frame 被解释为 uintptr，则其值表示程序计数器 + 1。)


// 比如两个输出结果可以发现, 1005239 和 1165825 两个 pc 值是相同的.
// 它们分别对应 runtime.main 和 runtime.goexit 函数.
//
// runtime.Caller 输出的 1367001 和 runtime.Callers 输出的 1367274 并不相同.
// 这是因为, 这两个函数的调用位置并不相同, 因此导致了 pc 值也不完全相同.
//
// 最后就是 runtime.Callers 多输出一个 1367301 值,
// 对应 runtime.Callers 内部的调用位置.

