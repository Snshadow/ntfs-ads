//go:build windows
// +build windows

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Snshadow/ntfs-ads"
)

func main() {
	var flagStdin, flagAppend bool
	var flagSourceFile, flagTargetFile, flagADSName string

	flag.BoolVar(&flagStdin, "stdin", false, "read data from standard input")
	flag.BoolVar(&flagAppend, "append", false, "append data into specified stream")

	flag.StringVar(&flagSourceFile, "source-file", "", "source file for data being written")
	flag.StringVar(&flagTargetFile, "target-file", "", "target path for writing ADS")
	flag.StringVar(&flagADSName, "ads-name", "", "name of the ADS to write data into")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "")

		flag.PrintDefaults()
	}

	flag.Parse()

	if flagTargetFile == "" {
		if flagTargetFile = flag.Arg(0); flagTargetFile == "" {
			flag.Usage()
			os.Exit(1)
		}
	}

	var src io.Reader

	if flagStdin {
		src = os.Stdin
	} else {
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
		if flagStdin {
			flagADSName = flag.Arg(1)
		} else {
			flagADSName = flag.Arg(2)
		}
		if flagADSName == "" {
			flag.Usage()
			os.Exit(1)
		}
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

	rd := bufio.NewReader(src)
	rdBuf := make([]byte, 4096)

	var wrOffset int64

	for {
		n, rdErr := rd.Read(rdBuf)
		if rdErr != nil {
			if rdErr != io.EOF {
				err = fmt.Errorf("failed to read data from file: %v", err)
			}
			break
		}

		if n < len(rdBuf) {
			rdBuf = rdBuf[:n]
		}

		if flagAppend {
			_, sErr := strmHnd.Write(rdBuf)
			if sErr != nil {
				err = fmt.Errorf("write error: %v", sErr)
				goto EXIT
			}
		} else {
			n, sErr := strmHnd.WriteAt(rdBuf, wrOffset)
			if sErr != nil {
				err = fmt.Errorf("write error: %v", sErr)
				goto EXIT
			}
			wrOffset += int64(n)
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
