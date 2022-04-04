package main

import (
	"fmt"
	"runtime"
)

func main() {
	pc := make([]uintptr, 1024)

	for skip := 0; ; skip++ {
		n := runtime.Callers(skip, pc)
		if n <= 0 {
			break
		}

		fmt.Printf("n = %d, skip = %v, pc = %v\n", n, skip, pc[:n])
	}
}

// 输出：
// 	n = 4, skip = 0, pc = [5954463 5954440 5592759 5753345]
// 	n = 3, skip = 1, pc = [5954440 5592759 5753345]
// 	n = 2, skip = 2, pc = [5592759 5753345]
// 	n = 1, skip = 3, pc = [5753345]

// func Callers(skip int, pc []uintptr) int
//	作用：
//		用来返回调用栈的程序计数器，放到一个 unitprt 中。
//
//	参数:
//		skip
//			0 则表示 runtime.Callers 自身；这和 Caller 的参数意义不一样,历史原因造成的。
//			1 和 Caller 的 0 对应。
//
//	返回值：
//		该函数返回写入到 pc 切片中的项数，受切片的容量限制。