# ntfs-ads
Access NTFS(New Technology File System) ADS(Alternate Data Stream) using golang.

This package provides access data streams in NTFS for files and directories with names(a.k.a Alternate Data Stream) which can be accessed by appending ":\[stream name\]" after file name.
This also appiles to directories and reparse points, which is normally not available with cmd with commonly known methods. Also, extracting data from alternative stream is a bit complicated with cmd.

## Query ADS from file
_Query name and size of ADS from file_
```go
import (
	"fmt"

	"github.com/Snshadow/ntfs-ads"
)

func main() {
	targetPath := "test.txt"

	ads, err := ntfs_ads.GetFileADS(targetPath)
	if err != nil {
		panic(err)
	}

	for name, size := range ads.StreamInfoMap {
		fmt.Printf("name: %s, size: %d\n", name, size)
	}
}
```

## Write, remove, rename ADS from file
```go
import (
	"fmt"
	"os"

	"github.com/Snshadow/ntfs-ads"
)

func main() {
	targetPath := "test.txt"

	ads1, err := ntfs_ads.OpenFileADS(targetPath, "ads1", os.O_CREATE|os_O_WRONLY)
	ads1.Write([]byte("test ads 1"))
	ads1.Close()
	ads2, err := ntfs_ads.OpenFileADS(targetPath, "ads2", os.O_CREATE|os_O_WRONLY)
	ads2.Write([]byte("test ads 2"))
	ads2.Close()
	ads3, err := ntfs_ads.OpenFileADS(targetPath, "ads3", os.O_CREATE|os_O_WRONLY)
	ads3.Write([]byte("test ads 3"))
	ads3.Close()
	ads4, err := ntfs_ads.OpenFileADS(targetPath, "ads4", os.O_CREATE|os_O_WRONLY)
	ads4.Write([]byte("test ads 4"))
	ads4.Close()

	// create ADS handler for file
	ads, err := ntfs_ads.GetFileADS(targetPath)
	if err != nil {
		panic(err)
	}

	// rename ADS "ads1" to "renamed1"
	err = ads.RenameADS("ads1", "renamed1", true)
	if err != nil {
		panic(err)
	}

	// remove ADS "ads2"
	err = ads.RemoveADS("ads2")
	if err != nil {
		panic(err)
	}

	// remove all ADS from "test.txt"
	err = ads.RemoveAllADS()
	if err != nil {
		panic(err)
	}
}
```

## Executables

This package has two executables for accessing ADS from file. Binary files can be found in release page.

※ _"github.com/josephspurrier/goversioninfo" package used to set information for exe._

_P.S.: Microsoft Defender has a tendency to flag golang compiled exe as trojan malware, but it's a false positive. If you are concerned, you can look at the source file in the cmd directory and build it yourself._

```
query-ads.exe queries ADS(Alternate Data Stream) from the named file, reads and writes its content if requested.
Usage:
Query all ADS name from file: query-ads.exe [filename]
Write ADS content to file: query-ads.exe -filename [file name] -ads-name [ADS name] -out-file [outfile name]      
Write ADS content to stdout(for piping output): query-ads.exe -filename [filename] -ads-name [ADS name] -stdout | (process output)

  -ads-name string
        name of a ADS to read data
  -filename string
        name of a file to query ADS
  -out-file string
        name of a file to output ADS data, default to ADS name
  -stdout
        write ads content to stdout
```

```
write_ads.exe writes data info the specified ADS(Alternate Data Stream). Can read data from file or standard input.
Usage:
Write data from file: write_ads.exe [target file] [source file] [ADS name]
 or
 write_ads.exe -source-file [source-file] -target-file [target file] -ads-name [ADS name]
Write data from stdin: echo "[data]" | write_ads.exe --stdin [target file] [ADS name]
Remove ADS from file: write_ads.exe -remove -target-file [target file] -ads-name [ADS name]
Remove all ADS from file: write_ads.exe -remove-all [target-file]
Rename ADS from file: write_ads.exe -rename [target name] [ADS name] [new ADS name]

  -ads-name string
        name of the ADS to write data or remove
  -append
        append data into specified stream
  -new-ads-name string
        new name for the ADS
  -remove
        remove specified ADS
  -remove-all
        remove all ADS from specified file
  -rename
        rename specified ADS
  -source-file string
        source file of data being written
  -stdin
        read data from standard input
  -target-file string
        target path for writing ADS
```
