package main

import "fmt"

func main() {
	Reverse(&floor{})
}

type floor struct {
}

func (f *floor) ABC() {
	fmt.Println("abc")
}

func Reverse(c Close) {
	if _, ok := c.(Open); ok {
		fmt.Println("Open ok")
	}
}

type Close interface {
	ABC()
}

type Open interface {
	CBA()
}
