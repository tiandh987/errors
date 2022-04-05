package errors

import (
	"reflect"
	"sort"
)

// 文件内容：
//		1、type String map[string]Empty
//			字符串 Set 集合
//			(1)两种创建方法：NewString()、StringKeySet()
//			(2)插入：Insert()
//			(3)删除：Delete()
//			(4)查询：Has()、HasAll()、HasAny()
//					List()、UnsortedList()
//					PopAny()
//			(5)Len()

//			两个 Set
//			(6)Diffenrence()
//			(7)并集：Union()
//			(8)交集：Intersection()
//			(9)IsSuperSet()
//			(10)Equal()
//
//		2、type sortableSliceOfString []string
//			(1) Len()
//			(2) Less()
//			(3) Swap()


// Empty 是公共的，因为它被一些内部 API 对象用于 “外部字符串数组” 和 “内部集合” 之间的转换，
// 现在转换逻辑需要公共类型。
type Empty struct{}

// String 是一组字符串，通过 map[string]struct{} 实现以最小化内存消耗。
type String map[string]Empty

// NewString creates a String from a list of values.
func NewString(items ...string) String {
	ss := String{}
	ss.Insert(items...)
	return ss
}

// StringKeySet 从 map[string](? extends interface{}) 的键创建一个 String。
// 如果传入的值实际上不是一个 map，这会 panic。
func StringKeySet(theMap interface{}) String {
	v := reflect.ValueOf(theMap)
	ret := String{}

	for _, keyValue := range v.MapKeys() {
		ret.Insert(keyValue.Interface().(string))
	}
	return ret
}

// Insert 添加 item 到 set
func (s String) Insert(items ...string) String {
	for _, item := range items {
		s[item] = Empty{}
	}

	return s
}

// Delete 从 set 中删除 items。
func (s String) Delete(items ...string) String {
	for _, item := range items {
		delete(s, item)
	}
	return s
}

// Has 返回 true，如果 item 在 set 中
func (s String) Has(item string) bool {
	_, contained := s[item]
	return contained
}

// HasAll 返回 true，如果所有 item 在 set 中
func (s String) HasAll(items ...string) bool {
	for _, item := range items {
		if !s.Has(item) {
			return false
		}
	}
	return true
}

// HasAny 返回 true，如果任意一个 item 在 set 中
func (s String) HasAny(items ...string) bool {
	for _, item := range items {
		if s.Has(item) {
			return true
		}
	}
	return false
}

// Difference returns a set of objects that are not in s2
// For example:
// 	s = {a1, a2, a3}
// 	s2 = {a1, a2, a4, a5}
//
// s.Difference(s2) = {a3}
// s2.Difference(s) = {a4, a5}
func (s String) Difference(s2 String) String {
	result := NewString()
	for key := range s {
		if !s2.Has(key) {
			result.Insert(key)
		}
	}
	return result
}

// Union returns a new set which includes items in either s or s2.
// For example:
// 	s = {a1, a2}
// 	s2 = {a3, a4}
//
// 	s.Union(s2) = {a1, a2, a3, a4}
// 	s2.Union(s) = {a1, a2, a3, a4}
func (s String) Union(s2 String) String {
	result := NewString()
	for key := range s {
		result.Insert(key)
	}
	for key := range s2 {
		result.Insert(key)
	}
	return result
}

// Intersection returns a new set which includes the item in BOTH s and s2
// For example:
// 	s = {a1, a2}
// 	s2 = {a2, a3}
//
// 	s.Intersection(s2) = {a2}
func (s String) Intersection(s2 String) String {
	var walk, other String
	result := NewString()
	if s.Len() < s2.Len() {
		walk = s
		other = s2
	} else {
		walk = s2
		other = s
	}
	for key := range walk {
		if other.Has(key) {
			result.Insert(key)
		}
	}
	return result
}

// IsSuperset returns true if and only if s is a superset of s2.
func (s String) IsSuperset(s2 String) bool {
	for item := range s2 {
		if !s.Has(item) {
			return false
		}
	}
	return true
}

// 当且仅当 s 等于（作为一个集合）s2 时，Equal 才返回 true。
// 如果两个集合的成员相同，则它们相等。
// （实际上，这意味着相同的元素，顺序无关紧要）
func (s String) Equal(s2 String) bool {
	return len(s) == len(s2) && s.IsSuperset(s2)
}

// List returns the contents as a sorted string slice.
func (s String) List() []string {
	res := make(sortableSliceOfString, 0, len(s))
	for key := range s {
		res = append(res, key)
	}
	sort.Sort(res)
	return []string(res)
}

// UnsortedList returns the slice with contents in random order.
func (s String) UnsortedList() []string {
	res := make([]string, 0, len(s))
	for key := range s {
		res = append(res, key)
	}
	return res
}

// PopAny returns a single element from the set.
func (s String) PopAny() (string, bool) {
	for key := range s {
		s.Delete(key)
		return key, true
	}
	var zeroValue string
	return zeroValue, false
}

// Len returns the size of the set.
func (s String) Len() int {
	return len(s)
}

//===========================================================
type sortableSliceOfString []string

func (s sortableSliceOfString) Len() int {
	return len(s)
}

func (s sortableSliceOfString) Less(i, j int) bool {
	return lessString(s[i], s[j])
}

func (s sortableSliceOfString) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func lessString(lhs, rhs string) bool {
	return lhs < rhs
}