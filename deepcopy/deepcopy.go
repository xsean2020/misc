//go:build go1.18
// +build go1.18

package deepcopy

import (
	"reflect"
	"sync"
	"time"
)

type Interface interface {
	DeepCopy() interface{}
}

// 类型缓存 key 为 reflect.Type，value 为类型拷贝函数（或字段信息）
var copierCache sync.Map

func Copy[T any](src T) (t T) {
	original := reflect.ValueOf(src)
	switch original.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		if original.IsNil() {
			return
		}
	}
	cpy := reflect.New(original.Type()).Elem()
	deepCopyValue(original, cpy)
	return cpy.Interface().(T)
}

// 复制主逻辑：带缓存
func deepCopyValue(original, cpy reflect.Value) {
	// check if implements custom Interface
	if original.CanInterface() {
		if copier, ok := original.Interface().(Interface); ok {
			cpy.Set(reflect.ValueOf(copier.DeepCopy()))
			return
		}
	}

	typ := original.Type()
	// 使用缓存
	if copierFunc, ok := copierCache.Load(typ); ok {
		copierFunc.(func(reflect.Value, reflect.Value))(original, cpy)
		return
	}

	// 构造并缓存
	copier := createCopier(typ)
	copierCache.Store(typ, copier)
	copier(original, cpy)
}

// 创建对应类型的复制函数
func createCopier(typ reflect.Type) func(reflect.Value, reflect.Value) {
	switch typ.Kind() {
	case reflect.Chan:
		return func(original, cpy reflect.Value) {
			if original.IsNil() {
				return
			}
			cpy.Set(reflect.MakeChan(typ, original.Cap()))
		}

	case reflect.Ptr:
		return func(original, cpy reflect.Value) {
			if original.IsNil() {
				return
			}
			originalElem := original.Elem()
			newPtr := reflect.New(originalElem.Type())
			deepCopyValue(originalElem, newPtr.Elem())
			cpy.Set(newPtr)
		}
	case reflect.Interface:
		return func(original, cpy reflect.Value) {
			if original.IsNil() {
				return
			}
			elem := original.Elem()
			newVal := reflect.New(elem.Type()).Elem()
			deepCopyValue(elem, newVal)
			cpy.Set(newVal)
		}
	case reflect.Struct:
		// time.Time 特判
		if typ == reflect.TypeOf(time.Time{}) {
			return func(original, cpy reflect.Value) {
				cpy.Set(original)
			}
		}
		// 预提取所有可访问字段
		var fields []int
		for i := 0; i < typ.NumField(); i++ {
			if typ.Field(i).PkgPath == "" { // exported
				fields = append(fields, i)
			}
		}
		return func(original, cpy reflect.Value) {
			for _, i := range fields {
				deepCopyValue(original.Field(i), cpy.Field(i))
			}
		}

	case reflect.Array:
		return func(original, cpy reflect.Value) {
			for i := 0; i < original.Len(); i++ {
				deepCopyValue(original.Index(i), cpy.Index(i))
			}
		}
	case reflect.Slice:
		return func(original, cpy reflect.Value) {
			if original.IsNil() {
				return
			}
			cpy.Set(reflect.MakeSlice(typ, original.Len(), original.Cap()))
			for i := 0; i < original.Len(); i++ {
				deepCopyValue(original.Index(i), cpy.Index(i))
			}
		}
	case reflect.Map:
		return func(original, cpy reflect.Value) {
			if original.IsNil() {
				return
			}
			cpy.Set(reflect.MakeMap(typ))
			for _, key := range original.MapKeys() {
				origVal := original.MapIndex(key)
				newVal := reflect.New(origVal.Type()).Elem()
				deepCopyValue(origVal, newVal)
				newKey := reflect.ValueOf(Copy(key.Interface()))
				cpy.SetMapIndex(newKey, newVal)
			}
		}
	default:
		return func(original, cpy reflect.Value) {
			cpy.Set(original)
		}
	}
}
