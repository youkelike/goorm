package main

import (
	"fmt"
	"unsafe"
)

type EmbeddedStruct struct {
	A int8
	B int64
}

type OuterStruct struct {
	X int16
	EmbeddedStruct
	Y int64
}

func main() {
	fmt.Println(unsafe.Sizeof(EmbeddedStruct{})) // 估计EmbeddedStruct的大小
	fmt.Println(unsafe.Sizeof(OuterStruct{}))    // 估计OuterStruct的大小
}
