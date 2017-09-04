// +build !wrap

package main

var Handler = ExampleHandler

type ExampleInput struct {
	A int
	B int
}

func ExampleHandler(in ExampleInput) int {
	return in.A + in.B
}
