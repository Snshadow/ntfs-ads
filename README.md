# ntfs-ads
Access NTFS(New Technology File System) ADS(Alternate Data Stream) using golang.

This package provides access data streams in NTFS for files and directories with names(a.k.a Alternate Data Stream) which can be accessed by appending ":\[stream name\]" after file name when accessing.
This also appiles to directories, which is normally not available with cmd with commonly known methods. Also, extracting data from alternative stream is a bit complicated with cmd.

## Query ADS from file
_Query name and size of ADS from file_
```go
import (
	"fmt"

	"github.com/Snshadow/ntfs-ads"
)

func main() {
	targetPath := "test.txt"

	strmMap, err := ntfs_ads.GetFileADS(targetPath)
	if err != nil {
		panic(err)
	}

	for name, size := range strmMap {
		fmt.Printf("name: %s, size: %d\n", name, size)
	}
}
```

## Executables

This package has two executables for accessing ADS from file. Binary files can be found in release page.

※ _"github.com/josephspurrier/goversioninfo" package used to set information for exe._

_P.S.: Microsoft Defender has a tendency to flag golang compiled exe as trojan malware, but it's a false positive. If you are concerned, you can look at the source file in the cmd directory and build it yourself._

```
query-ads.exe queries ADS(Alternate Data Stream) from the named file. Read and write its content if requested.
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
write-ads.exe writes data info the specified ADS(Alternate Data Stream). Can read data from file or stdin.
Usage:
Write data from file: write-ads.exe [source file] [target file] [ADS name] or write-ads.exe -source-file [source-file] -target-file [target file] -ads-name [ADS name]
Write data from stdin: echo "[data]" | write-ads.exe --stdin [target file] [ADS name]
Remove ADS from file: write-ads.exe -remove -target-file [target file] -ads-name [ADS name]

  -ads-name string
        name of the ADS to write data or remove
  -append
        append data into specified stream
  -remove
        remove specified ADS
  -source-file string
        source file of data being written
  -stdin
        read data from standard input
  -target-file string
        target path for writing ADS
```
