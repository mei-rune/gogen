package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/runner-mei/gogen/gengen"
)

func usage() {
	fmt.Printf("Usage: %s <filename> (try -h)", os.Args[0])
}

func main() {

	// var (
	// 	outdirrel = flag.String("target-dir", ".", "base directory to emit into")
	// )

	flag.Usage = usage

	gen := &gengen.Generator{}
	gen.Flags(flag.CommandLine)
	flag.Parse()

	args := flag.Args()
	gen.Run(args)

	// for _, filename := range args {
	// 	log.Println("#", filename)
	// 	_, err := gengen.ParseFile(filename)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 		return
	// 	}
	// }
}
