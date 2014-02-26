package main

import "github.com/sqp/godock/libs/packages"

import "github.com/kr/pretty"

func DEBUG(args ...interface{}) {
	for _, arg := range args {
		pretty.Printf("%# v\n", arg)
	}
}

func main() {
	packages.ListDownload()
}
