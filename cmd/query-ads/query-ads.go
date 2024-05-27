//go:build windows
// +build windows

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"

	ntfs_ads "github.com/Snshadow/ntfs-ads"
)

func main() {
	var flagStdout bool
	var flagFileName, flagTargetAds, flagOutFileName string

	flag.BoolVar(&flagStdout, "stdout", false, "write ads content to stdout")

	flag.StringVar(&flagFileName, "filename", "", "name of a file to query ADS")
	flag.StringVar(&flagTargetAds, "ads-name", "", "name of a ADS to read data")
	flag.StringVar(&flagOutFileName, "out-file", "", "name of a file to output ADS data, default to ADS name")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s queries ADS(Alternate Data Stream) from the named file. Read and write its content if requested.\nUsage\nQuery all ADS name from file: %s [filename]\nWrite ADS content to file: %s -filename [file name] -ads-name [ADS name] -out-file [outfile name]\nWrite ADS content to stdout: %s -filename [filename] -ads-name [ADS name] -stdout > [outfile]\n\n", os.Args[0], os.Args[0], os.Args[0], os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if flagFileName == "" {
		if flagFileName = flag.Arg(0); flagFileName == "" {
			flag.Usage()
			os.Exit(1)
		}
	}

	if flagTargetAds == "" {
		// query all ADS name(s)
		strmMap, err := ntfs_ads.GetFileADS(flagFileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not query ADS from file \"%s\": %v\n", flagFileName, err)
		}

		fmt.Printf("ADS of %s(name : byte size):\n", flagFileName)
		for name, size := range strmMap {
			fmt.Printf("\t%s: %d\n", name, size)
		}
	} else {
		var err error

		strmHnd, sErr := ntfs_ads.OpenFileADS(flagFileName, flagTargetAds)
		if sErr != nil {
			fmt.Fprintf(os.Stderr, "Could not open descriptor of \"%s\" from file \"%s\": %v\n", flagTargetAds, flagFileName, sErr)
			os.Exit(2)
		}
		defer strmHnd.Close()

		// use buffered io in case of large sized data stored in ADS
		bw := bufio.NewReader(strmHnd)
		rdBuf := make([]byte, 4096)

		var outFileName string
		if flagOutFileName == "" {
			outFileName = flagTargetAds
		} else {
			outFileName = flagOutFileName
		}

		var wrOffset int64
		outFile, sErr := os.Create(outFileName)
		if sErr != nil {
			err = fmt.Errorf("Could not prepare file for writing ADS data: %v", sErr)

			goto EXIT
		}
		defer outFile.Close()

		for {
			n, rdErr := bw.Read(rdBuf)
			if rdErr != nil && rdErr != io.EOF {
				err = fmt.Errorf("read error: %v", rdErr)
				break
			} else if rdErr == io.EOF {
				break
			}

			if n < len(rdBuf) {
				rdBuf = rdBuf[:n]
			}

			if flagStdout {
				os.Stdout.Write(rdBuf)
			} else {
				n, wrErr := outFile.WriteAt(rdBuf, wrOffset)
				if wrErr != nil {
					goto EXIT
				}
				wrOffset += int64(n)
			}
		}

	EXIT:
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while writing from ADS: %v", err)
			os.Exit(2)
		}
	}
}
