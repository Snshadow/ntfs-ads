//go:generate goversioninfo write_ads.json

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
	"github.com/Snshadow/ntfs-ads/cmd/utils"
)

func main() {
	var flagStdin, flagAppend, flagRemove, flagRemoveAll, flagRename bool
	var flagSourceFile, flagTargetFile, flagADSName, flagNewADSName string

	flag.BoolVar(&flagStdin, "stdin", false, "read data from standard input")
	flag.BoolVar(&flagAppend, "append", false, "append data into specified stream")
	flag.BoolVar(&flagRemove, "remove", false, "remove specified ADS")
	flag.BoolVar(&flagRemoveAll, "remove-all", false, "remove all ADS from specified file")
	flag.BoolVar(&flagRename, "rename", false, "rename specified ADS")

	flag.StringVar(&flagSourceFile, "source-file", "", "source file of data being written")
	flag.StringVar(&flagTargetFile, "target-file", "", "target path for writing ADS")
	flag.StringVar(&flagADSName, "ads-name", "", "name of the ADS to write data or remove")
	flag.StringVar(&flagNewADSName, "new-ads-name", "", "new name for the ADS")

	progName := filepath.Base(os.Args[0])

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s writes data info the specified ADS(Alternate Data Stream). Can read data from file or standard input.\nUsage:\nWrite data from file: %s [target file] [source file] [ADS name]\n or\n %s -source-file [source-file] -target-file [target file] -ads-name [ADS name]\nWrite data from stdin: echo \"[data]\" | %s --stdin [target file] [ADS name]\nRemove ADS from file: %s -remove -target-file [target file] -ads-name [ADS name]\nRemove all ADS from file: %s -remove-all [target-file]\nRename ADS from file: %s -rename [target name] [ADS name] [new ADS name]\n\n", progName, progName, progName, progName, progName, progName, progName)

		flag.PrintDefaults()

		// prevent window from closing immediately if the console was created for this process
		if utils.IsFromOwnConsole() {
			fmt.Println("\nPress enter to close...")
			fmt.Scanln()
		}
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
	} else if !flagRemove && !flagRemoveAll && !flagRename {
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

	if flagADSName == "" && !flagRemoveAll {
		if flagStdin || flagRemove || flagRename {
			flagADSName = flag.Arg(1)
		} else {
			flagADSName = flag.Arg(2)
		}
		if flagADSName == "" {
			flag.Usage()
			os.Exit(1)
		}
	}

	if flagRename && flagNewADSName == "" {
		if flagNewADSName = flag.Arg(2); flagNewADSName == "" {
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

	if flagRemoveAll || flagRename {
		ads, err := ntfs_ads.GetFileADS(flagTargetFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not ADS from file \"%s\": %v\n", flagTargetFile, err)
			os.Exit(2)
		}

		if flagRemoveAll {
			err = ads.RemoveAllADS()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not remove all ADS from file \"%s\": %v\n", flagTargetFile, err)
				os.Exit(2)
			}
			fmt.Printf("Removed all ADS from file \"%s\"\n", flagTargetFile)
		} else if flagRename {
			err = ads.RenameADS(flagADSName, flagNewADSName, true)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not rename ADS \"%s\" to \"%s\" from file \"%s\": %v\n", flagADSName, flagNewADSName, flagTargetFile, err)
				os.Exit(2)
			}
			fmt.Printf("Renamed ADS \"%s\" to \"%s\" from file \"%s\"\n", flagADSName, flagNewADSName, flagTargetFile)
		}

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
	defer strmHnd.Close()

	_, err = io.Copy(strmHnd, src)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while writing data into ADS: %v\n", err)
		os.Exit(2)
	} else {
		if flagAppend {
			fmt.Printf("Appended data into ADS \"%s:%s\"\n", flagTargetFile, flagADSName)
		} else {
			fmt.Printf("Wrote data into ADS \"%s:%s\"\n", flagTargetFile, flagADSName)
		}
	}
}
