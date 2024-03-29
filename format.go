package errors

import (
	"bytes"
	"fmt"
)

// formatInfo 包含所有的错误信息
type formatInfo struct {
	code    int
	message string
	err     string
	stack   *stack
}

// list 会将错误堆栈转换为一个简单的数组。
func list(e error) []error {
	ret := []error{}

	if e != nil {
		if w, ok := e.(interface{ Unwrap() error }); ok {
			ret = append(ret, e)
			ret = append(ret, list(w.Unwrap())...)
		} else {
			ret = append(ret, e)
		}
	}

	return ret
}

// buildFormatInfo 构建格式化信息
// 进行类型断言：fundamental、withStack、withCode、其他
func buildFormatInfo(e error) *formatInfo {
	var finfo *formatInfo

	switch err := e.(type) {
	case *fundamental:
		finfo = &formatInfo{
			code:    unknownCoder.Code(),
			message: err.msg,
			err:     err.msg,
			stack:   err.stack,
		}
	case *withStack:
		finfo = &formatInfo{
			code:    unknownCoder.Code(),
			message: err.Error(),
			err:     err.Error(),
			stack:   err.stack,
		}
	case *withCode:
		coder, ok := codes[err.code]
		if !ok {
			coder = unknownCoder
		}

		extMsg := coder.String()
		if extMsg == "" {
			extMsg = err.err.Error()
		}

		finfo = &formatInfo{
			code:    coder.Code(),
			message: extMsg,
			err:     err.err.Error(),
			stack:   err.stack,
		}
	default:
		finfo = &formatInfo{
			code:    unknownCoder.Code(),
			message: err.Error(),
			err:     err.Error(),
		}
	}

	return finfo
}

func format(k int, jsonData []map[string]interface{}, str *bytes.Buffer, finfo *formatInfo,
	sep string, flagDetail, flagTrace, modeJSON bool) ([]map[string]interface{}, *bytes.Buffer) {

	if modeJSON {
		data := map[string]interface{}{}
		if flagDetail || flagTrace {
			data = map[string]interface{}{
				"message": finfo.message,
				"code":    finfo.code,
				"error":   finfo.err,
			}

			caller := fmt.Sprintf("#%d", k)
			if finfo.stack != nil {
				f := Frame((*finfo.stack)[0])
				caller = fmt.Sprintf("%s %s:%d (%s)",
					caller,
					f.file(),
					f.line(),
					f.name(),
				)
			}
			data["caller"] = caller
		} else {
			data["error"] = finfo.message
		}
		jsonData = append(jsonData, data)
	} else {
		if flagDetail || flagTrace {
			if finfo.stack != nil {
				f := Frame((*finfo.stack)[0])
				fmt.Fprintf(str, "%s%s - #%d [%s:%d (%s)] (%d) %s",
					sep,
					finfo.err,
					k,
					f.file(),
					f.line(),
					f.name(),
					finfo.code,
					finfo.message)
			} else {
				fmt.Fprintf(str, "%s%s - #%d %s", sep, finfo.err, k, finfo.message)
			}
		} else {
			fmt.Fprintf(str, finfo.message)
		}
	}

	return jsonData, nil
}
