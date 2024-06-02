//go:generate go run github.com/josephspurrier/goversioninfo/cmd/goversioninfo query_ads.json

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

var (
	getNameSizePad = func(m map[string]int64) (name int, size int) {
		for k, v := range m {
			if l := len(k); l > name {
				name = l
			}
			if l := len(fmt.Sprintf("%d", v)); l > size {
				size = l
			}
		}

		return
	}
)

func main() {
	var flagStdout bool
	var flagFileName, flagTargetAds, flagOutFileName string

	flag.BoolVar(&flagStdout, "stdout", false, "write ads content to stdout")

	flag.StringVar(&flagFileName, "filename", "", "name of a file to query ADS")
	flag.StringVar(&flagTargetAds, "ads-name", "", "name of a ADS to read data")
	flag.StringVar(&flagOutFileName, "out-file", "", "name of a file to output ADS data, default to ADS name")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s queries ADS(Alternate Data Stream) from the named file. Read and write its content if requested.\nUsage:\nQuery all ADS name from file: %s [filename]\nWrite ADS content to file: %s -filename [file name] -ads-name [ADS name] -out-file [outfile name]\nWrite ADS content to stdout(for piping output): %s -filename [filename] -ads-name [ADS name] -stdout | (process output)\n\n", os.Args[0], os.Args[0], os.Args[0], os.Args[0])
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
		// query all ADS name(s) and size(s)
		strmMap, err := ntfs_ads.GetFileADS(flagFileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not query ADS from file \"%s\": %v\n", flagFileName, err)

			return
		}

		namePad, sizePad := getNameSizePad(strmMap)

		fmt.Printf("ADS of %s:\n(name : byte size)\n", flagFileName)
		for name, size := range strmMap {
			fmt.Printf("%*s : %*d\n", namePad, name, sizePad, size)
		}
	} else {
		var err error

		strmHnd, sErr := ntfs_ads.OpenFileADS(flagFileName, flagTargetAds, os.O_RDONLY)
		if sErr != nil {
			fmt.Fprintf(os.Stderr, "Could not open ADS with name \"%s\" from file \"%s\": %v\n", flagTargetAds, flagFileName, sErr)
			os.Exit(2)
		}

		// use buffered io in case of large sized data stored in ADS
		bw := bufio.NewReader(strmHnd)
		rdBuf := make([]byte, 4096)

		var outFileName string
		if !flagStdout {
			if flagOutFileName == "" {
				outFileName = flagTargetAds
			} else {
				outFileName = flagOutFileName
			}
		}

		var outFile *os.File

		if outFileName != "" {
			var sErr error
			outFile, sErr = os.Create(outFileName)
			if sErr != nil {
				err = fmt.Errorf("could not prepare file for writing ADS data: %v", sErr)

				goto EXIT
			}
		}

		for {
			n, rdErr := bw.Read(rdBuf)
			if rdErr != nil {
				if rdErr != io.EOF {
					err = fmt.Errorf("read error: %v", rdErr)
				}
				break
			}

			if n < len(rdBuf) {
				rdBuf = rdBuf[:n]
			}

			if flagStdout {
				os.Stdout.Write(rdBuf)
			} else {
				_, wrErr := outFile.Write(rdBuf)
				if wrErr != nil {
					goto EXIT
				}
			}
		}

		if outFileName != "" && err == nil {
			fmt.Printf("Wrote ADS data into file \"%s\"\n", outFileName)
		}

	EXIT:
		if outFile != nil {
			outFile.Close()
		}
		strmHnd.Close()

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while reading data from ADS: %v", err)
			os.Exit(2)
		}
	}
}
