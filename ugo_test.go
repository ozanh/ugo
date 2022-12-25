package ugo_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	. "github.com/ozanh/ugo"
)

func TestToObject(t *testing.T) {
	err := errors.New("test error")
	fn := func(...Object) (Object, error) { return nil, nil }

	testCases := []struct {
		iface   interface{}
		want    Object
		wantErr bool
	}{
		{iface: nil, want: Undefined},
		{iface: "a", want: String("a")},
		{iface: int64(-1), want: Int(-1)},
		{iface: int(1), want: Int(1)},
		{iface: uint(1), want: Uint(1)},
		{iface: uint64(1), want: Uint(1)},
		{iface: uintptr(1), want: Uint(1)},
		{iface: true, want: True},
		{iface: false, want: False},
		{iface: rune(1), want: Char(1)},
		{iface: byte(1), want: Char(1)},
		{iface: float64(1), want: Float(1)},
		{iface: float32(1), want: Float(1)},
		{iface: []byte(nil), want: Bytes{}},
		{iface: []byte("a"), want: Bytes{'a'}},
		{iface: map[string]Object(nil), want: Map{}},
		{iface: map[string]Object{"a": Int(1)}, want: Map{"a": Int(1)}},
		{iface: map[string]interface{}{"a": 1}, want: Map{"a": Int(1)}},
		{iface: map[string]interface{}{"a": uint32(1)}, wantErr: true},
		{iface: []Object(nil), want: Array{}},
		{iface: []Object{Int(1), Char('a')}, want: Array{Int(1), Char('a')}},
		{iface: []interface{}{Int(1), Char('a')}, want: Array{Int(1), Char('a')}},
		{iface: []interface{}{uint32(1)}, wantErr: true},
		{iface: Object(nil), want: Undefined},
		{iface: String("a"), want: String("a")},
		{iface: CallableFunc(nil), want: Undefined},
		{iface: fn, want: &Function{Value: fn}},
		{iface: err, want: &Error{Message: err.Error(), Cause: err}},
		{iface: error(nil), want: Undefined},
		{iface: uint16(1), wantErr: true},
	}

	for _, tC := range testCases {
		t.Run(fmt.Sprintf("%[1]T:%[1]v", tC.iface), func(t *testing.T) {
			got, err := ToObject(tC.iface)
			if (err != nil) != tC.wantErr {
				t.Errorf("ToObject() error = %v, wantErr %v", err, tC.wantErr)
				return
			}
			if fn, ok := tC.iface.(CallableFunc); ok && fn != nil {
				require.NotNil(t, tC.want.(*Function).Value)
				return
			}
			if !reflect.DeepEqual(got, tC.want) {
				t.Errorf("ToObject() = %v, want %v", got, tC.want)
			}
		})
	}
}

func TestToInterface(t *testing.T) {

	testCases := []struct {
		object Object
		want   interface{}
	}{
		{object: nil, want: nil},
		{object: Undefined, want: nil},
		{object: Int(1), want: int64(1)},
		{object: String(""), want: ""},
		{object: String("a"), want: "a"},
		{object: Bytes(nil), want: []byte(nil)},
		{object: Bytes(""), want: []byte{}},
		{object: Bytes("a"), want: []byte{'a'}},
		{object: Array(nil), want: []interface{}{}},
		{object: Array{}, want: []interface{}{}},
		{object: Array{Int(1)}, want: []interface{}{int64(1)}},
		{object: Array{Undefined}, want: []interface{}{nil}},
		{object: Map(nil), want: map[string]interface{}{}},
		{object: Map{}, want: map[string]interface{}{}},
		{object: Map{"a": Undefined}, want: map[string]interface{}{"a": nil}},
		{object: Map{"a": Int(1)}, want: map[string]interface{}{"a": int64(1)}},
		{object: Uint(1), want: uint64(1)},
		{object: Char(1), want: rune(1)},
		{object: Float(1), want: float64(1)},
		{object: True, want: true},
		{object: False, want: false},
		{object: (*SyncMap)(nil), want: map[string]interface{}{}},
		{
			object: &SyncMap{Value: Map{"a": Int(1)}},
			want:   map[string]interface{}{"a": int64(1)},
		},
	}
	for _, tC := range testCases {
		t.Run(fmt.Sprintf("%T", tC.object), func(t *testing.T) {
			if got := ToInterface(tC.object); !reflect.DeepEqual(got, tC.want) {
				t.Errorf("ToInterface() = %v, want %v", got, tC.want)
			}
		})
	}
}

func TestToObjectAlt(t *testing.T) {
	err := errors.New("test error")
	fn := func(...Object) (Object, error) { return nil, nil }

	testCases := []struct {
		iface   interface{}
		want    Object
		wantErr bool
	}{
		{iface: nil, want: Undefined},
		{iface: "a", want: String("a")},
		{iface: int64(-1), want: Int(-1)},
		{iface: int32(-1), want: Int(-1)},
		{iface: int16(-1), want: Int(-1)},
		{iface: int8(-1), want: Int(-1)},
		{iface: int(1), want: Int(1)},
		{iface: uint(1), want: Uint(1)},
		{iface: uint64(1), want: Uint(1)},
		{iface: uint32(1), want: Uint(1)},
		{iface: uint16(1), want: Uint(1)},
		{iface: uint8(1), want: Uint(1)},
		{iface: uintptr(1), want: Uint(1)},
		{iface: true, want: True},
		{iface: false, want: False},
		{iface: rune(1), want: Int(1)},
		{iface: byte(2), want: Uint(2)},
		{iface: float64(1), want: Float(1)},
		{iface: float32(1), want: Float(1)},
		{iface: []byte(nil), want: Bytes{}},
		{iface: []byte("a"), want: Bytes{'a'}},
		{iface: map[string]Object(nil), want: Map{}},
		{iface: map[string]Object{"a": Int(1)}, want: Map{"a": Int(1)}},
		{iface: map[string]interface{}{"a": 1}, want: Map{"a": Int(1)}},
		{iface: map[string]interface{}{"a": uint32(1)}, want: Map{"a": Uint(1)}},
		{iface: []Object(nil), want: Array{}},
		{iface: []Object{Int(1), Char('a')}, want: Array{Int(1), Char('a')}},
		{iface: []interface{}{Int(1), Char('a')}, want: Array{Int(1), Char('a')}},
		{iface: []interface{}{uint32(1)}, want: Array{Uint(1)}},
		{iface: Object(nil), want: Undefined},
		{iface: String("a"), want: String("a")},
		{iface: CallableFunc(nil), want: Undefined},
		{iface: fn, want: &Function{Value: fn}},
		{iface: err, want: &Error{Message: err.Error(), Cause: err}},
		{iface: error(nil), want: Undefined},
		{iface: struct{}{}, wantErr: true},
	}

	for _, tC := range testCases {
		t.Run(fmt.Sprintf("%[1]T:%[1]v", tC.iface), func(t *testing.T) {
			got, err := ToObjectAlt(tC.iface)
			if (err != nil) != tC.wantErr {
				t.Errorf("ToObjectAlt() error = %v, wantErr %v", err, tC.wantErr)
				return
			}
			if fn, ok := tC.iface.(CallableFunc); ok && fn != nil {
				require.NotNil(t, tC.want.(*Function).Value)
				return
			}
			if !reflect.DeepEqual(got, tC.want) {
				t.Errorf("ToObjectAlt() = %[1]v (%[1]T), want %[2]v (%[2]T)", got, tC.want)
			}
		})
	}
}
