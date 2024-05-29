//go:build windows
// +build windows

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/Snshadow/ntfs-ads"
)

func main() {
	var flagStdin bool
	var flagSrcFile, flagTargetFile, flagADSName string


	flag.BoolVar(&flagStdin, "stdin", false, "write data from standard input")
	flag.StringVar(&flagSrcFile, "source-file", "", "source file for data being written")
	flag.StringVar(&flagTargetFile, "target-file", "", "target path for writing ADS")
	flag.StringVar(&flagADSName, "ads-name", "", "name of the ADS to write data into")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "")

		flag.PrintDefaults()
	}

	flag.Parse()


	stdRd := bufio.NewReader(os.Stdin)

	
}
