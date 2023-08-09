package tests

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func Sdump(value interface{}) string {
	var sb strings.Builder
	sdump("", &sb, value)
	return sb.String()
}
func sdump(prefix string, sb *strings.Builder, value interface{}) {
	const indent = "  "
	if value == nil {
		sb.WriteString(fmt.Sprintf("(%[1]T) %[1]v\n", value))
		return
	}
	typ := reflect.TypeOf(value)
	switch typ.Kind() {
	case reflect.Slice:
		val := reflect.ValueOf(value)
		if val.IsNil() {
			sb.WriteString(fmt.Sprintf("(%+v nil)\n", typ))
			return
		}
		sb.WriteString(fmt.Sprintf("(%+v len=%d cap=%d) {",
			typ, val.Len(), val.Cap()))
		sz := val.Len()
		if sz > 0 {
			sb.WriteString("\n")
		} else {
			sb.WriteString("}\n")
			return
		}
		for i := 0; i < sz; i++ {
			sb.WriteString(prefix + indent + "#")
			sb.WriteString(strconv.Itoa(i))
			sb.WriteString(" ")
			index := val.Index(i)
			if index.Kind() != reflect.Interface && index.Kind() != reflect.Ptr {
				sdump(prefix+indent, sb, index.Interface())
				continue
			}
			elem := val.Index(i).Elem()
			var iface interface{}
			if elem.IsValid() && elem.CanInterface() {
				iface = elem.Interface()
			}
			sdump(prefix+indent, sb, iface)
		}
		sb.WriteString(prefix + "}\n")
	case reflect.Map:
		val := reflect.ValueOf(value)
		if val.IsNil() {
			sb.WriteString(fmt.Sprintf("(%+v nil)\n", typ))
			return
		}
		keys := val.MapKeys()
		fmt.Fprintf(sb, "(%+v len=%d) {", typ, len(keys))
		if len(keys) == 0 {
			sb.WriteString("}\n")
			return
		}
		sb.WriteString("\n")
		for _, k := range keys {
			sb.WriteString(prefix + indent + fmt.Sprintf("%#v: ", k))
			var iface interface{}
			vkind := val.MapIndex(k).Kind()
			if vkind == reflect.Ptr || vkind == reflect.Interface {
				elem := val.MapIndex(k).Elem()
				if elem.IsValid() && elem.CanInterface() {
					iface = elem.Interface()
				}
				sdump(prefix+indent, sb, iface)
			} else {
				v := val.MapIndex(k).Interface()
				sdump(prefix+indent, sb, v)
			}
		}
		sb.WriteString(prefix + "}\n")
		return
	default:
		fmt.Fprintf(sb, "(%+v) %+v\n", typ, value)
	}
}
