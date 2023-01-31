package projectpath

import (
	"fmt"
	"path/filepath"
	"runtime"
)

var Root string

func init() {
	_, b, _, _ := runtime.Caller(0)
	Root = filepath.Join(filepath.Dir(b), "../..")
	fmt.Println("Just set the ROOT as ", Root)
}
