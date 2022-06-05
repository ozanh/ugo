#!/bin/bash
set -x
set -e
set -u
set -o pipefail

# This script is used to create repl.gif under docs to simulate key presses and
# record the terminal window as a gif. Recording as a gif should be handled
# separately.

# To install xdotool on Debian/Ubuntu
## sudo apt install xdotool

# Note that you need run to this script on X window system.

function echoerr {
  echo "$1"
  exit 1
}

function xtype {
  xdotool windowfocus "$WID"
  xdotool type --delay 150ms "$1" &&
  sleep .3
  xdotool windowfocus "$WID"
  xdotool key --clearmodifiers "Return"
  sleep .3
}

GOBIN=$(go env GOBIN)
if test -z "$GOBIN"; then
  mkdir -p "$HOME/go/bin"
  export GOBIN="$HOME/go/bin"
fi

go install .

gnome-terminal --geometry -0+0 --working-directory "$GOBIN"

sleep 2

WID=$(xdotool search --class gnome-terminal | tail -1)
test -z "$WID" && echoerr "Cannot find window id" || echo "Window ID:$WID"

xtype 'clear && uname -a'
xtype './ugo'
xtype 'sum := func(...args) { var t=0; for v in args { t+=v }; return t }'
xtype 'a := 11; b := 22'
xtype 'c := sum(a, b); printf("%d + %d = %d\n", a, b, c)'
xtype 'try { \'
xtype '  a/0 \'
xtype '} catch err { \'
xtype '  println("Caught error:", err) \'
xtype '}'
xtype '.commands'
xtype '.builtins'
xtype 'typeName(a)'
xtype 'arr'
xtype 'arr := [a, b, c]'
xtype 'arr = append(arr, "str")'
xtype 'arr'
xtype 'arr2 := arr[:1]'
xtype 'm := {key1: arr, "key 2": arr2}'
xtype 'm'
xtype 'fmt:=import("fmt")'
xtype 'json:=import("json")'
xtype '_ := fmt.Printf("%s", json.MarshalIndent(m, "", " "))'
xtype '.exit'

exit 0
