package utils

import "container/list"

// List2Array list对象转数组
func List2Array(list *list.List) []interface{} {
	var len = list.Len()
	if len == 0 {
		return nil
	}
	var arr []interface{}
	for e := list.Front(); e != nil; e = e.Next() {
		arr = append(arr, e.Value)
	}
	return arr
}
