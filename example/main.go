package main

import "fmt"

type Example2 struct {
	Example1 string
	Example5 string
}

type Example struct {
	Example2

	Example3 string
	Example4 int
}

func main() {
	example := Example{
		Example2: Example2{
			Example1: "example1",
			Example5: "example5",
		},
		Example3: "example3",
		Example4: 4,
	}

	//组合模式下，修改嵌套结构体的字段
	example.Example1 = "ddd"
	// 直接修改嵌套结构体的字段
	example.Example2.Example1 = "d2"

	fmt.Println(example)
}
