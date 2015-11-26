package main

import "eleme"
import "runtime"

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	eleme.Eleme()
}
