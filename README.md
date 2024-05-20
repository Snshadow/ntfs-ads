# ntfs-ads
Access NTFS(New Technology File System) ADS(Alternate Data Stream) using golang.

This package provides access data streams in NTFS for files and directories with names(a.k.a Alternate Data Stream) which default data stream does not have.
This also appiles to directories, which is normally not available with cmd with commonly known methods. Also, extracting data from alternative stream is a bit complicated with cmd.
