// +build go1.13

package errors

import (
	stderrors "errors"
)

// Is 报告 err 链中是否有任何 error 与 target 匹配。
//
// 该链由 err 本身 和 通过重复调用 Unwrap 获得的错误序列组成。
//
// 如果 error 与 target 相等，
// 或者如果它实现方法 Is(error) bool 使得 Is(target) 返回 true，
// 则认为 error 与 target 匹配。
func Is(err, target error) bool {
	return stderrors.Is(err, target)
}

// As 找到 err 链中与 target 匹配的第一个错误，如果是，则将 target 设置为该错误值并返回 true。
//
// 该链由 err 本身和通过重复调用 Unwrap 获得的错误序列组成。
//
// 如果 error 的具体值可分配给 target 指向的值，
// 或者如果 error 具有方法 As(interface{}) bool 使得 As(target) 返回 true，则错误匹配 target。
// 在后一种情况下，As 方法负责设置目标。
//
// 如果 target 不是指向实现 error 的类型 或 任何接口类型的非零指针，则会出现 panic。
// 如果 err 为 nil，则 As 返回 false。
func As(err error, target interface{}) bool {
	return stderrors.As(err, target)
}

// Unwrap 返回 对 err 调用 Unwrap 方法的结果，如果 err 的类型包含 Unwrap 方法返回错误。
// 否则，Unwrap 返回 nil。
func Unwrap(err error) error {
	return stderrors.Unwrap(err)
}
