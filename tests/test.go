package main

import (
	"FTC/util"
	"fmt"
	"strings"
)

func main() {
	r := "/dev/ttyS12"
	index := strings.IndexAny(r, "[0123456789]")
	fmt.Println(index)
	fmt.Println(r[index:])

	util.ChkTimeOut("2019-03-01 19:40:00", 10)
}
