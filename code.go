package errors

import (
	"fmt"
	"net/http"
	"sync"
)

// 使用：
// 	在使用该 errors 包的时候，需要调用 Register 或者 MustRegister，
// 	将一个 Coder 注册到 errors 开辟的内存中，数据结构为：
//		var codes = map[int]Coder{}

// 该文件内容：
//	1、定义 Coder 接口
//
//	   Coder 注册函数 Register、MustRegister
//		区别：当重复定义同一个错误 Code 时，MustRegister会 panic，
//			这样可以防止后面注册的错误覆盖掉之前注册的错误。
//			在实际开发中，建议使用MustRegister。
//
//	2、用于存储注册 Coder 的内存空间
//
//	3、实现 Coder 接口的 defaultCoder 结构体
//
//	4、预定义 Coder unknownCoder

// codes contains a map of error codes to metadata.
var codes = map[int]Coder{}
var codeMux = &sync.Mutex{}

var (
	unknownCoder defaultCoder = defaultCoder{
		C:    0,
		HTTP: http.StatusInternalServerError,
		Ext:  "An internal server error occurred",
		Ref:  "http://github.com/tiandh987/errors/README.md",
	}
)

func init() {
	codes[unknownCoder.Code()] = unknownCoder
}

// =========================================================
type Coder interface {
	// 用于相关错误码的 HTTP 状态
	HTTPStatus() int

	// 外部（用户）可见的错误文本
	String() string

	// 返回给用户该错误对应的详细文档
	Reference() string

	// Code returns the code of the coder
	Code() int
}

// 设计技巧：
//	XXX()和MustXXX()的函数命名方式，是一种 Go 代码设计技巧，在 Go 代码中经常使用，
//	例如 Go 标准库中 regexp 包提供的 Compile 和 MustCompile 函数。
//	和 XXX 相比，MustXXX 会在某种情况不满足时 panic。
//	因此使用 MustXXX 的开发者看到函数名就会有一个心理预期：使用不当，会造成程序 panic。

// Register 注册一个用户定义的错误码
// 它将会覆盖已存在的相同 code
func Register(coder Coder) {
	if coder.Code() == 0 {
		// 0 被该 error 包保留为 unknownCode 错误码
		panic("code `0` is reserved by `github.com/tiandh987/errors` as unknownCode error code")
	}

	codeMux.Lock()
	defer codeMux.Unlock()

	codes[coder.Code()] = coder
}

// Register 注册一个用户定义的错误码
// 当相同的 Code 已经存在时，将会引发 panic
func MustRegister(coder Coder) {
	if coder.Code() == 0 {
		panic("code `0` is reserved by `github.com/tiandh987/errors` as unknownCode error code")
	}

	codeMux.Lock()
	defer codeMux.Unlock()

	if _, ok := codes[coder.Code()]; ok {
		panic(fmt.Sprintf("code: %d already exist", coder.Code()))
	}

	codes[coder.Code()] = coder
}

// =================================================
type defaultCoder struct {
	// C 指的是 ErrCode 的整数代码
	C    int

	// HTTP status用于该错误代码的 HTTP 状态码。
	HTTP int

	// External 外部（用户）可见的错误文本
	Ext  string

	// Reference 错误相关的 reference 文档
	Ref  string
}

func (coder defaultCoder) Code() int {
	return coder.C
}

func (coder defaultCoder) HTTPStatus() int {
	if coder.HTTP == 0 {
		return 500
	}
	return coder.HTTP
}

func (coder defaultCoder) String() string {
	return coder.Ext
}

func (coder defaultCoder) Reference() string {
	return coder.Ref
}

// ========================================================
// ParseCoder 解析任何 error 为 *withCode。
// nil error 将直接返回 nil
// None withStack error will be parsed as ErrUnknown.
func ParseCoder(err error) Coder {
	if err == nil {
		return nil
	}

	if v, ok := err.(*withCode); ok {
		if coder, ok := codes[v.code]; ok {
			return coder
		}
	}

	return unknownCoder
}

// IsCode 报告错误链中是否包含给定的错误代码。
func IsCode(err error, code int) bool {
	if v, ok := err.(*withCode); ok {
		if v.code == code {
			return true
		}

		if v.cause != nil {
			return IsCode(v.cause, code)
		}

		return false
	}

	return false
}