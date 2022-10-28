package utils

import (
	"container/list"
	"errors"
	"reflect"
)

// List2Array list对象转数组
func List2Array(list *list.List, receiver any) error {
	rvPtr := reflect.ValueOf(receiver)
	if rvPtr.Kind() != reflect.Ptr || rvPtr.IsNil() {
		return errors.New(reflect.TypeOf(receiver).Name() + "not ptr")
	}

	if list.Len() == 0 {
		return nil
	}
	rv := rvPtr.Elem()
	for e := list.Front(); e != nil; e = e.Next() {
		rv.Set(reflect.Append(rv, reflect.ValueOf(e.Value)))
	}
	return nil
}
