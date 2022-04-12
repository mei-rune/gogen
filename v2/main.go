package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/runner-mei/gogen/v2/gengen"
)

func usage() {
	fmt.Printf(`使用方法: %s 子命令 <filename> (try -h)
	有如下子命令: server, client`, os.Args[0])
	os.Exit(1)
}

func main() {

	// var (
	// 	outdirrel = flag.String("target-dir", ".", "base directory to emit into")
	// )

	// flag.Usage = usage
	// flag.Parse()
	// args := flag.Args()
	// if len(args) == 0 {
	// 	usage()
	// 	return
	// }

	// fset := flag.NewFlagSet(args[0], flag.ExitOnError)

	fset := flag.CommandLine
	args := os.Args

	var gen = &gengen.Generator{}
	gen.Flags(fset)
	err := fset.Parse(args[1:])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := gen.Run(fset.Args()); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
