package errors

import (
	"fmt"
	"io"
	"runtime"
	"strconv"
	"strings"
)

// 文件内容：
//	1、type stack []uintptr
//		pc 计数器切片
//
//	2、type Frame uintptr
//		栈帧
//
//  3、type StackTrace []Frame
//		调用栈

// ====================================================
// 程序计数器切片
type stack []uintptr

// callers 获取程序计数器切片
func callers() *stack {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])

	var st stack = pcs[0:n]
	return &st
}

func (s *stack) Format(st fmt.State, verb rune)  {
	switch verb {
	case 'v':
		switch {
		case st.Flag('+'):
			for _, pc := range *s {
				f := Frame(pc)
				fmt.Fprintf(st, "\n%+v", f)
			}

		}
	}
}

// StackTrack 将 pc 计数器的值 转化为 调用栈帧
func (s *stack) StackTrace() StackTrace {
	f := make([]Frame, len(*s))
	for i := 0; i < len(f); i++ {
		f[i] = Frame((*s)[i])
	}
	return f
}

// =======================================================
// Frame 表示堆栈帧内的程序计数器。
// 由于历史原因，如果 Frame 被解释为 uintptr，则其值表示程序计数器 + 1。
type Frame uintptr

// pc 返回此帧的程序计数器；
// 多个帧可能具有相同的 PC 值。
func (f Frame) pc() uintptr {
	return uintptr(f) - 1
}

// file 返回包含此 Frame 的 pc 函数的文件的完整路径。
func (f Frame) file() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return "unknow"
	}
	file, _ := fn.FileLine(f.pc())
	return file
}

// line 返回此 Frame 的 pc 的函数源代码的行号。
func (f Frame) line() int {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return 0
	}
	_, line := fn.FileLine(f.pc())
	return line
}

// name 返回这个函数的名字，如果知道的话。
func (f Frame) name() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return "unknown"
	}
	return fn.Name()
}

// Format 根据 fmt.Formatter 接口格式化帧。
//
// 	%s 源文件
// 	%d 源代码行
// 	%n 函数名
// 	%v 等价于 %s:%d
//
// Format 接受改变某些动词打印的标志，如下所示：
//
// 	%+s 函数名和源文件
//		相对于编译时的路径 GOPATH 用 \n\t 分隔
//		(<funcname>\n\t<path>)
//	%+v 等价于 %+s:%d
func (f Frame) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		switch {
		case s.Flag('+'):
			io.WriteString(s, f.name())
			io.WriteString(s, "\n\t")
			io.WriteString(s, f.file())
		default:
			io.WriteString(s, f.file())
		}
	case 'd':
		io.WriteString(s, strconv.Itoa(f.line()))
	case 'n':
		io.WriteString(s, funcname(f.name()))
	case 'v':
		f.Format(s, 's')
		io.WriteString(s, ":")
		f.Format(s, 'd')
	}
}

// MarshalText 将堆栈跟踪帧格式化为文本字符串。
// 输出与 fmt.Sprintf("%+v", f) 的输出相同，但没有换行符或制表符。
func (f Frame) MarshalText() ([]byte, error) {
	name := f.name()
	if name == "unknown" {
		return []byte(name), nil
	}

	return []byte(fmt.Sprintf("%s %s:%d", name, f.file(), f.line())), nil
}

// funcname 删除由 func.Name() 报告的函数名称的路径前缀组件。
func funcname(name string) string {
	i := strings.LastIndex(name, "/")
	name = name[i+1:]
	i = strings.Index(name, ".")
	return name[i+1:]
}

// =======================================================
// StackTrace 是从最内层（最新）到最外层（最旧）的帧堆栈。
type StackTrace []Frame

func (st StackTrace) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case s.Flag('+'):
			for _, f := range st {
				io.WriteString(s, "\n")
				f.Format(s, verb)
			}
		case s.Flag('#'):
			fmt.Fprintf(s, "%#v", []Frame(st))
		default:
			st.formatSlice(s, verb)
		}
	case 's':
		st.formatSlice(s, verb)
	}
}

// formatSlice 会将这个 StackTrace 格式化到给定的缓冲区，
// 作为 Frame 的切片，仅在使用 '%s' 或 '%v' 调用时有效。
func (st StackTrace) formatSlice(s fmt.State, verb rune) {
	io.WriteString(s, "[")
	for i, f := range st {
		if i > 0 {
			io.WriteString(s, " ")
		}
		f.Format(s, verb)
	}
	io.WriteString(s, "]")
}