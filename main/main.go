package main

import "fmt"

func myFunction() (result int) {
	defer func() {
		result = 42 // 尝试修改返回值
	}()

	result = 10 // 设置返回值
	return
}

func main() {
	fmt.Println(myFunction())
}
