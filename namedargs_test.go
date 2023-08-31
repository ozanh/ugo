package ugo

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/ozanh/ugo/tests"
	"github.com/stretchr/testify/assert"
)

func TestNamedArgs_All(t *testing.T) {
	tests := []struct {
		args    Map
		vargs   Map
		wantRet Map
	}{
		{Map{}, Map{}, Map{}},
		{Map{"a": True}, Map{}, Map{"a": True}},
		{Map{"a": True}, Map{"b": False}, Map{"a": True, "b": False}},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			n := NewNamedArgs(tt.args.Items(), tt.vargs.Items())
			assert.Equalf(t, tt.wantRet, n.Map(), "All()")
		})
	}
}

func TestNamedArgs_CheckNames(t *testing.T) {
	tests := []struct {
		args    Map
		vargs   Map
		accept  []string
		wantErr bool
	}{
		{Map{}, Map{}, nil, false},
		{Map{"a": True}, Map{}, nil, true},
		{Map{"a": True}, Map{}, []string{"a"}, false},
		{Map{"a": True}, Map{"b": False}, []string{"a"}, true},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			n := NewNamedArgs(tt.args.Items(), tt.vargs.Items())
			if err := n.CheckNames(tt.accept...); err == nil {
				if tt.wantErr {
					t.Error("want error, but not got")
					t.Failed()
				}
			} else if !tt.wantErr {
				t.Error("not want error, but got=" + err.Error())
				t.Failed()
			}
		})
	}
}

func TestNamedArgs_CheckNamesFromSet(t *testing.T) {
	tests := []struct {
		args    Map
		vargs   Map
		accept  []string
		wantErr bool
	}{
		{Map{}, Map{}, nil, false},
		{Map{"a": True}, Map{}, nil, true},
		{Map{"a": True}, Map{}, []string{"a"}, false},
		{Map{"a": True}, Map{"b": False}, []string{"a"}, true},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			n := NewNamedArgs(tt.args.Items(), tt.vargs.Items())
			set := make(map[string]interface{}, len(tt.accept))
			for _, v := range tt.accept {
				set[v] = nil
			}
			if err := n.CheckNamesFromSet(set); err == nil {
				if tt.wantErr {
					t.Error("want error, but not got")
					t.Failed()
				}
			} else if !tt.wantErr {
				t.Error("not want error, but got=" + err.Error())
				t.Failed()
			}
		})
	}
}

func TestNamedArgs_Get(t *testing.T) {
	tests := []struct {
		args    Map
		vargs   Map
		dst     []*NamedArgVar
		wantErr bool
	}{
		{Map{}, Map{}, nil, false},
		{Map{"a": True}, Map{}, nil, true},
		{Map{"a": True}, Map{}, []*NamedArgVar{{Name: "a"}}, false},
		{Map{"a": True}, Map{}, []*NamedArgVar{{Name: "a", AcceptTypes: []string{"int"}}}, true},
		{Map{"a": True}, Map{}, []*NamedArgVar{{Name: "a", AcceptTypes: []string{"bool"}}}, false},
		{Map{"a": True}, Map{"b": False}, []*NamedArgVar{{Name: "a"}}, true},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			n := NewNamedArgs(tt.args.Items(), tt.vargs.Items())
			if err := n.Get(tt.dst...); err == nil {
				if tt.wantErr {
					t.Error("want error, but not got")
					t.Failed()
				} else {
					for _, dst := range tt.dst {
						if dst.Value != n.GetValue(dst.Name) {
							t.Errorf("bad value of %q: want=%v, got=%v", dst.Name, dst.Value, n.GetValue(dst.Name))
							t.Failed()
						}
					}
				}
			} else if !tt.wantErr {
				t.Error("not want error, but got=" + err.Error())
				t.Failed()
			}
		})
	}
}

func TestNamedArgs_GetVar(t *testing.T) {
	tests_ := []struct {
		args    Map
		vargs   Map
		dst     []*NamedArgVar
		other   Map
		wantErr bool
	}{
		{Map{}, Map{}, nil, Map{}, false},
		{Map{"a": True}, Map{}, nil, Map{"a": True}, false},
		{Map{"a": True}, Map{}, []*NamedArgVar{{Name: "a"}}, Map{}, false},
		{Map{"a": True}, Map{}, []*NamedArgVar{{Name: "a", AcceptTypes: []string{"int"}}}, Map{}, true},
		{Map{"a": True}, Map{}, []*NamedArgVar{{Name: "a", AcceptTypes: []string{"bool"}}}, Map{}, false},
		{Map{"a": True}, Map{"b": False}, []*NamedArgVar{{Name: "a"}}, Map{"b": False}, false},
		{Map{"a": True, "c": Int(1)}, Map{"b": False}, []*NamedArgVar{{Name: "a"}}, Map{"c": Int(1), "b": False}, false},
	}
	for i, tt := range tests_ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			n := NewNamedArgs(tt.args.Items(), tt.vargs.Items())
			if other, err := n.GetVar(tt.dst...); err == nil {
				if tt.wantErr {
					t.Error("want error, but not got")
					t.Failed()
				} else {
					for _, dst := range tt.dst {
						if dst.Value != n.GetPassedValue(dst.Name) {
							t.Errorf("bad value of %q: want=%v, got=%v", dst.Name, dst.Value, n.GetValue(dst.Name))
							t.Failed()
						}
					}

					if !reflect.DeepEqual(other, tt.other) {
						t.Fatalf("Objects not equal:\nExpected:\n%s\nGot:\n%s\n",
							tests.Sdump(tt.other), tests.Sdump(other))
					}
				}
			} else if !tt.wantErr {
				t.Error("not want error, but got=" + err.Error())
				t.Failed()
			}
		})
	}
}
