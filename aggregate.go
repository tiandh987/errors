package errors

import "errors"

// MessageCountMap 包含每个错误消息的出现次数。
type MessageCountMap map[string]int

//==========================================================
// Aggregate 表示一个包含多个错误的对象，但不一定具有单一的语义。
// 聚合可以与 `errors.Is()` 一起使用来检查特定错误类型的发生。
// 不支持 Errors.As()，因为调用者可能关心与给定类型匹配的潜在多个特定错误。
type Aggregate interface {
	error
	Errors() []error
	Is(error) bool
}

// NewAggregate 将 错误切片 转换为 Aggregate 接口，
// 该接口本身就是错误接口的实现。 如果切片为空，则返回 nil。
//
// 它将检查输入错误列表的任何元素是否为 nil，以避免在调用 Error() 时出现 nil 指针 panic。
func NewAggregate(errlist []error) Aggregate {
	if len(errlist) == 0 {
		return nil
	}

	var errs []error
	for _, e := range errlist {
		if e != nil {
			errs = append(errs, e)
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return aggregate(errs)
}

//=====================================================
// 这个 helper 实现了 error 和 Errors 接口。
// 保持私有可以防止人们产生 0 个错误的聚合，这不是错误，但确实满足错误接口。
type aggregate []error

func (agg aggregate) visit(f func(err error) bool) bool {
	for _, err := range agg {
		switch err := err.(type) {
		case aggregate:
			if match := err.visit(f); match {
				return match
			}
		case Aggregate:
			for _, nestedErr := range err.Errors() {
				if match := f(nestedErr); match {
					return match
				}
			}
		default:
			if match := f(err); match {
				return match
			}
		}
	}

	return false
}

// Errors is part of the Aggregate interface.
func (agg aggregate) Errors() []error {
	return []error(agg)
}

// Error is part of the error interface.
func (agg aggregate) Error() string {
	if len(agg) == 0 {
		// This should never happen, really.
		return ""
	}

	if len(agg) == 1 {
		return agg[0].Error()
	}

	seenerrs := NewString()
	result := ""
	agg.visit(func(err error) bool {
		msg := err.Error()
		if seenerrs.Has(msg) {
			return false
		}
		seenerrs.Insert(msg)
		if len(seenerrs) > 1 {
			result += ", "
		}
		result += msg
		return false
	})

	if len(seenerrs) == 1 {
		return result
	}

	return "[" + result + "]"
}

func (agg aggregate) Is(target error) bool {
	return agg.visit(func(err error) bool {
		return errors.Is(err, target)
	})
}

//=====================================================
// Matcher 用于匹配 error。 如果匹配，返回 true。
type Matcher func(error) bool

// FilterOut 从输入错误中删除与任何匹配器匹配的所有错误。
// 如果输入是单一错误，则仅测试该错误。
// 如果输入实现了 Aggregate 接口，则错误列表将被递归处理。
//
// 例如，这可用于从错误列表中删除 known-OK 错误
//（例如 io.EOF 或 os.PathNotFound）。
func FilterOut(err error, fns ...Matcher) error {
	if err == nil {
		return nil
	}

	if agg, ok := err.(Aggregate); ok {
		return NewAggregate(filterErrors(agg.Errors(), fns...))
	}

	if !matchesError(err, fns...) {
		return err
	}

	return nil
}

// matchesError returns true if any Matcher returns true
func matchesError(err error, fns ...Matcher) bool {
	for _, fn := range fns {
		if fn(err) {
			return true
		}
	}
	return false
}

// filterErrors 返回所有 fns 都返回 false 的任何错误（或嵌套错误，如果列表包含嵌套错误）。
// 如果没有错误存在，则返回 nil 列表。
// 作为副作用，生成的 silec 将所有嵌套切片展平。
func filterErrors(list []error, fns ...Matcher) []error {
	result := []error{}
	for _, err := range list {
		r := FilterOut(err, fns...)
		if r != nil {
			result = append(result, r)
		}
	}
	return result
}