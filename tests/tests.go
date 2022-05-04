package tests

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func Sdump(value interface{}) string {
	if value == nil {
		return fmt.Sprintf("(%[1]T) %[1]v\n", value)
	}
	typ := reflect.TypeOf(value)
	switch typ.Kind() {
	case reflect.Slice:
		val := reflect.ValueOf(value)
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("(%+v len=%d cap=%d) {",
			typ, val.Len(), val.Cap()))
		sz := val.Len()
		if sz > 0 {
			sb.WriteString("\n")
		} else {
			sb.WriteString("}\n")
			return sb.String()
		}
		for i := 0; i < sz; i++ {
			sb.WriteString("#")
			sb.WriteString(strconv.Itoa(i))
			sb.WriteString("  | ")
			index := val.Index(i)
			if index.Kind() != reflect.Interface && index.Kind() != reflect.Ptr {
				sb.WriteString(Sdump(index.Interface()))
				continue
			}
			elem := val.Index(i).Elem()
			var iface interface{}
			if elem.IsValid() && elem.CanInterface() {
				iface = elem.Interface()
			}
			sb.WriteString(Sdump(iface))
		}
		sb.WriteString(fmt.Sprintf("= %#v\n", value))
		sb.WriteString("}\n")
		return sb.String()
	case reflect.Map:
		val := reflect.ValueOf(value)
		keys := val.MapKeys()
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("(%+v len=%d) {", typ, len(keys)))
		if len(keys) == 0 {
			sb.WriteString("}\n")
			return sb.String()
		}
		sb.WriteString("\n")
		for _, k := range keys {
			sb.WriteString(fmt.Sprintf("%#v:", k))
			var iface interface{}
			vkind := val.MapIndex(k).Kind()
			if vkind == reflect.Ptr || vkind == reflect.Interface {
				elem := val.MapIndex(k).Elem()
				if elem.IsValid() && elem.CanInterface() {
					iface = elem.Interface()
				}
				sb.WriteString(Sdump(iface))
			} else {
				v := val.MapIndex(k).Interface()
				sb.WriteString(Sdump(v))
			}
		}
		sb.WriteString(fmt.Sprintf("= %#v\n", value))
		sb.WriteString("}\n")
		return sb.String()
	default:
		return fmt.Sprintf("(%+v) %+v\n", typ, value)
	}
}
