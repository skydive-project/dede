package main

import "github.com/skydive-project/dede/dede"

func main() {
	dede.InitServer()
	dede.ListenAndServe()
}
