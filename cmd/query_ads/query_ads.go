//go:generate go run github.com/josephspurrier/goversioninfo/cmd/goversioninfo query_ads.json

//go:build windows
// +build windows

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Snshadow/ntfs-ads"
	"github.com/Snshadow/ntfs-ads/cmd/utils"
)

var (
	getNameSizePad = func(m map[string]int64) (name int, size int) {
		for k, v := range m {
			if l := len(k); l > name {
				name = l
			}
			if l := len(strconv.FormatInt(v, 10)); l > size {
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

	progName := filepath.Base(os.Args[0])

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s queries ADS(Alternate Data Stream) from the named file, reads and writes its content if requested.\nUsage:\nQuery all ADS name from file: %s [filename]\nWrite ADS content to file: %s -filename [filename] -ads-name [ADS name] -out-file [outfile name]\n or\n %s [filename] [ADS name] [outfile name]\nWrite ADS content to stdout(for piping output): %s -filename [filename] -ads-name [ADS name] -stdout | (process output)\n or\n %s -stdout [filename] [ADS name] | (process output)\n\n", progName, progName, progName, progName, progName, progName)
		flag.PrintDefaults()

		// prevent window from closing immediately if the console was created for this process
		if utils.IsFromOwnConsole() {
			fmt.Println("\nPress enter to close...")
			fmt.Scanln()
		}
	}

	flag.Parse()

	if flagFileName == "" {
		if flagFileName = flag.Arg(0); flagFileName == "" {
			flag.Usage()
			os.Exit(1)
		}
	}
	if flagTargetAds == "" {
		flagTargetAds = flag.Arg(1)
	}
	if flagOutFileName == "" && !flagStdout {
		flagOutFileName = flag.Arg(2)
	}

	if flagTargetAds == "" {
		// query all ADS name(s) and size(s)
		ads, err := ntfs_ads.GetFileADS(flagFileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not query ADS from file \"%s\": %v\n", flagFileName, err)

			return
		}

		namePad, sizePad := getNameSizePad(ads.StreamInfoMap)

		fmt.Printf("ADS of %s:\n(name : byte size)\n", flagFileName)
		for name, size := range ads.StreamInfoMap {
			fmt.Printf("%*s : %*d\n", namePad, name, sizePad, size)
		}
	} else {
		var err error

		strmHnd, sErr := ntfs_ads.OpenFileADS(flagFileName, flagTargetAds, os.O_RDONLY)
		if sErr != nil {
			fmt.Fprintf(os.Stderr, "Could not open ADS with name \"%s\" from file \"%s\": %v\n", flagTargetAds, flagFileName, sErr)
			os.Exit(2)
		}
		defer strmHnd.Close()

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
				fmt.Fprintf(os.Stderr, "could not prepare file for writing ADS data: %v", sErr)
				os.Exit(2)
			}
			defer outFile.Close()
		} else if flagStdout {
			outFile = os.Stdout
		}

		_, err = io.Copy(outFile, strmHnd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while reading data from ADS: %v", err)
			os.Exit(2)
		}

		if outFileName != "" {
			fmt.Printf("Wrote ADS data into file \"%s\"\n", outFileName)
		}
	}
}
