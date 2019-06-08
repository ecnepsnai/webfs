package main

import (
	"fmt"
	"os"
	"path"

	"github.com/ecnepsnai/logtic"

	"github.com/ecnepsnai/webfs"
)

func main() {
	dataDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	bindAddr := "127.0.0.1:8080"

	logtic.Log.FilePath = path.Join(dataDir, "webfs.log")
	logtic.Log.Level = logtic.LevelInfo

	args := os.Args[1:]
	i := 0
	for i < len(args) {
		arg := args[i]

		if arg == "-b" {
			value := args[i+1]
			i++
			bindAddr = value
		} else if arg == "-d" {
			value := args[i+1]
			i++
			dataDir = value
		} else if arg == "-v" {
			logtic.Log.Level = logtic.LevelDebug
		} else {
			fmt.Printf("Unknown argument '%s'\n", arg)
			printHelpAndExit()
		}

		i++
	}

	if err := logtic.Open(); err != nil {
		panic(err)
	}

	if err := webfs.Start(dataDir, bindAddr); err != nil {
		panic(err)
	}
}

func printHelpAndExit() {
	fmt.Printf("Usage %s [-b <bind address> -d <data directory> -v]\n", os.Args[0])
	fmt.Printf("-b Specify the IP address and port to listen on\n")
	fmt.Printf("-d Specify the data directory to work in\n")
	fmt.Printf("Default is to listen on 127.0.0.1:8080 and work in the current directory\n")
	os.Exit(1)
}
