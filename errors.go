package errors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// 文件内容：
//	1、type withCode struct
//		WithCode()
//
//	2、type withMessage struct
//		WithMessage()
//		WithMessagef()
//
//	3、type withStack struct
//		WithStack()
//
//	4、type fundamental struct
//		New()
//		Errorf()
//
// 	5、Wrap()、WrapC()、Wrapf()
//
//	6、Cause()

//==============================================================
// withCode 引入一种新的错误类型，
// 该错误类型记录错误码、stack、cause、具体的错误信息。
type withCode struct {
	err    error // error 错误
	code   int   // 业务错误码
	cause  error // cause error
	*stack       // 错误堆栈
}

// Error 返回外部安全的错误信息
func (w *withCode) Error() string {
	return fmt.Sprintf("%v", w)
}

// Cause 返回 withCode 错误的根因
func (w *withCode) Cause() error {
	return w.cause
}

// Unwrap provides compatibility for Go 1.13 error chains.
func (w withCode) Unwrap() error {
	return w.cause
}

// Format 实现 fmt.Formatter。 https://golang.org/pkg/fmt/#hdr-Printing
//
// Verbs：
// 		%s - 如果没有指定错误消息, 返回映射到错误代码的用户安全错误字符串。
// 		%v - %s 的别名
//
// Flags：
// 		# JSON 格式的输出，用于记录日志
// 		- 输出调用者详细信息，对故障排除有用
// 		+ 输出完整的错误堆栈详细信息，对调试很有用
//
// Examples：
//		%s:    error for internal read B
//      %v:    error for internal read B
//      %-v:   error for internal read B - #0 [/home/lk/workspace/golang/src/github.com/marmotedu/iam/main.go:12 (main.main)] (#100102) Internal Server Error
//      %+v:   error for internal read B - #0 [/home/lk/workspace/golang/src/github.com/marmotedu/iam/main.go:12 (main.main)] (#100102) Internal Server Error; error for internal read A - #1 [/home/lk/workspace/golang/src/github.com/marmotedu/iam/main.go:35 (main.newErrorB)] (#100104) Validation failed
//      %#v:   [{"error":"error for internal read B"}]
//      %#-v:  [{"caller":"#0 /home/lk/workspace/golang/src/github.com/marmotedu/iam/main.go:12 (main.main)","error":"error for internal read B","message":"(#100102) Internal Server Error"}]
//      %#+v:  [{"caller":"#0 /home/lk/workspace/golang/src/github.com/marmotedu/iam/main.go:12 (main.main)","error":"error for internal read B","message":"(#100102) Internal Server Error"},{"caller":"#1 /home/lk/workspace/golang/src/github.com/marmotedu/iam/main.go:35 (main.newErrorB)","error":"error for internal read A","message":"(#100104) Validation failed"}]
func (w *withCode) Format(state fmt.State, verb rune) {
	switch verb {
	case 'v':
		str := bytes.NewBuffer([]byte{})
		jsonData := []map[string]interface{}{}

		var (
			flagDetail bool
			flagTrace  bool
			modeJSON   bool
		)

		if state.Flag('#') {
			modeJSON = true
		}

		if state.Flag('-') {
			flagDetail = true
		}

		if state.Flag('+') {
			flagTrace = true
		}

		sep := ""
		errs := list(w)
		length := len(errs)
		for k, e := range errs {
			finfo := buildFormatInfo(e)
			jsonData, str = format(length-k-1, jsonData, str, finfo, sep, flagDetail, flagTrace, modeJSON)
			sep = "; "

			if !flagTrace {
				break
			}

			if !flagDetail && !flagTrace && !modeJSON {
				break
			}
		}

		if modeJSON {
			var byts []byte
			byts, _ = json.Marshal(jsonData)

			str.Write(byts)
		}

		fmt.Fprintf(state, "%s", strings.Trim(str.String(), "\r\n\t"))
	default:
		finfo := buildFormatInfo(w)
		fmt.Fprintf(state, finfo.message)
	}
}

// WithCode 函数创建新的 withCode 类型的错误
func WithCode(code int, format string, args ...interface{}) error {
	return &withCode{
		err:   fmt.Errorf(format, args...),
		code:  code,
		stack: callers(),
	}
}

//=========================================================
type withMessage struct {
	cause error
	msg   string
}

func (w *withMessage) Error() string {
	return w.msg
}

func (w *withMessage) Cause() error {
	return w.cause
}

func (w *withMessage) Unwrap() error {
	return w.cause
}

func (w *withMessage) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v\n", w.Cause())
			io.WriteString(s, w.msg)
			return
		}
		fallthrough
	case 's', 'q':
		io.WriteString(s, w.Error())
	}
}

// WithMessage 使用一个新的 message 注解一个 err
// 对 err 进行包装
func WithMessage(err error, message string) error {
	if err == nil {
		return nil
	}
	return &withMessage{
		cause: err,
		msg:   message,
	}
}

func WithMessagef(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	return &withMessage{
		cause: err,
		msg:   fmt.Sprintf(format, args...),
	}
}

//========================================================
type withStack struct {
	error
	*stack
}

func (w *withStack) Cause() error {
	return w.error
}

func (w *withStack) Unwrap() error {
	if e, ok := w.error.(interface{ Unwrap() error }); ok {
		return e.Unwrap()
	}

	return w.error
}

func (w withStack) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprint(s, "%+v", w.Cause())
			w.stack.Format(s, verb)
			return
		}
		fallthrough
	case 's':
		io.WriteString(s, w.Error())
	case 'q':
		fmt.Fprintf(s, "%q", w.Error())
	}
}

// WithStack 在调用 WithStack 时使用堆栈跟踪注释 err。
// 如果 err 为 nil，WithStack 返回 nil。
func WithStack(err error) error {
	if err == nil {
		return nil
	}

	if e, ok := err.(*withCode); ok {
		return &withCode{
			err:   e.err,
			code:  e.code,
			cause: err,
			stack: callers(),
		}
	}

	return &withStack{
		error: err,
		stack: callers(),
	}
}

//===================================================
// fundamental 是一个错误，它有一个消息和一个堆栈，但没有调用者。
// 作为最基本错误使用。
type fundamental struct {
	msg string
	*stack
}

func (f *fundamental) Error() string {
	return f.msg
}

func (f *fundamental) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			io.WriteString(s, f.msg)
			f.stack.Format(s, verb)
			return
		}
		fallthrough
	case 's':
		io.WriteString(s, f.msg)
	case 'q':
		fmt.Fprintf(s, "%q", f.msg)
	}
}

// New 使用提供的消息返回错误。
// New 还会记录调用时的堆栈跟踪。
func New(message string) error {
	return &fundamental{
		msg:   message,
		stack: callers(),
	}
}

// Errorf 根据格式说明符进行格式化，并将字符串作为满足错误的值返回。
// Errorf 还会记录调用时的堆栈跟踪。
func Errorf(format string, args ...interface{}) error {
	return &fundamental{
		msg:   fmt.Sprintf(format, args...),
		stack: callers(),
	}
}

// =======================================================
// Wrap 返回一个错误，在调用 Wrap 时使用堆栈跟踪注释 err，并提供消息。
// 如果 err 为 nil，Wrap 返回 nil。
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}

	if e, ok := err.(*withCode); ok {
		return &withCode{
			err:   fmt.Errorf(message),
			code:  e.code,
			cause: err,
			stack: callers(),
		}
	}

	err = &withMessage{
		cause: err,
		msg:   message,
	}

	return &withStack{
		error: err,
		stack: callers(),
	}
}

func WrapC(err error, code int, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	return &withCode{
		err:   fmt.Errorf(format, args...),
		code:  code,
		cause: err,
		stack: callers(),
	}
}

// Wrapf 返回一个错误，在调用 Wrapf 时使用堆栈跟踪注释 err，以及格式说明符。
// 如果 err 为 nil，Wrapf 返回 nil。
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	if e, ok := err.(*withCode); ok {
		return &withCode{
			err:   fmt.Errorf(format, args...),
			code:  e.code,
			cause: err,
			stack: callers(),
		}
	}

	err = &withMessage{
		cause: err,
		msg:   fmt.Sprintf(format, args...),
	}
	return &withStack{
		err,
		callers(),
	}
}
