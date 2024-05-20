//go:build generate

package w32api

//go:generate go run golang.org/x/sys/windows/mkwinsyscall -output zw32api.go w32api.go
