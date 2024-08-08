//go:generate go run github.com/josephspurrier/goversioninfo/cmd/goversioninfo write_ads.json

//go:build windows
// +build windows

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Snshadow/ntfs-ads"
)

func main() {
	var flagStdin, flagAppend, flagRemove bool
	var flagSourceFile, flagTargetFile, flagADSName string

	flag.BoolVar(&flagStdin, "stdin", false, "read data from standard input")
	flag.BoolVar(&flagAppend, "append", false, "append data into specified stream")
	flag.BoolVar(&flagRemove, "remove", false, "remove specified ADS")

	flag.StringVar(&flagSourceFile, "source-file", "", "source file of data being written")
	flag.StringVar(&flagTargetFile, "target-file", "", "target path for writing ADS")
	flag.StringVar(&flagADSName, "ads-name", "", "name of the ADS to write data or remove")
	
	progName := filepath.Base(os.Args[0])

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s writes data info the specified ADS(Alternate Data Stream). Can read data from file or stdin.\nUsage:\nWrite data from file: %s [target file] [source file] [ADS name] or %s -source-file [source-file] -target-file [target file] -ads-name [ADS name]\nWrite data from stdin: echo \"[data]\" | %s --stdin [target file] [ADS name]\nRemove ADS from file: %s -remove -target-file [target file] -ads-name [ADS name]\n\n", progName, progName, progName, progName, progName)

		flag.PrintDefaults()
	}

	flag.Parse()

	if flagTargetFile == "" {
		if flagTargetFile = flag.Arg(0); flagTargetFile == "" {
			flag.Usage()
			os.Exit(1)
		}
	}

	var src *os.File

	if flagStdin {
		src = os.Stdin
	} else if !flagRemove {
		if flagSourceFile == "" {
			if flagSourceFile = flag.Arg(1); flagSourceFile == "" {
				flag.Usage()
				os.Exit(1)
			}
		}
		fd, err := os.Open(flagSourceFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not open file for reading: %v\n", err)
			os.Exit(2)
		}
		src = fd
		defer fd.Close()
	}

	if flagADSName == "" {
		if flagStdin || flagRemove {
			flagADSName = flag.Arg(1)
		} else {
			flagADSName = flag.Arg(2)
		}
		if flagADSName == "" {
			flag.Usage()
			os.Exit(1)
		}
	}

	if flagRemove {
		err := os.Remove(flagTargetFile + ":" + flagADSName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not remove ADS from \"%s\" with name \"%s\": %v\n", flagTargetFile, flagADSName, err)
			os.Exit(2)
		}
		fmt.Printf("Removed ADS \"%s\" from file \"%s\"\n", flagADSName, flagTargetFile)

		return
	}

	openFlag := os.O_WRONLY
	if flagAppend {
		openFlag |= os.O_APPEND
	} else {
		openFlag |= os.O_CREATE | os.O_TRUNC
	}

	strmHnd, err := ntfs_ads.OpenFileADS(flagTargetFile, flagADSName, openFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open ADS for writing: %v\n", err)
		os.Exit(2)
	}

	rdBuf := make([]byte, 4096)

	for {
		n, rdErr := src.Read(rdBuf)
		if rdErr != nil {
			if rdErr != io.EOF {
				err = fmt.Errorf("failed to read data from file: %v", rdErr)
			}
			break
		}

		if n < len(rdBuf) {
			rdBuf = rdBuf[:n]
		}

		_, sErr := strmHnd.Write(rdBuf)
		if sErr != nil {
			err = fmt.Errorf("write error: %v", sErr)
			goto EXIT
		}

	}

EXIT:
	strmHnd.Close()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while writing data into ADS: %v\n", err)
		os.Exit(2)
	} else {
		fmt.Printf("Wrote data into ADS \"%s:%s\"\n", flagTargetFile, flagADSName)
	}
}
