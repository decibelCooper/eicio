package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/decibelcooper/eicio/go-eicio"
)

var (
	outFile = flag.String("o", "", "file to save output to")
	keep = flag.Bool("k", false, "keep only the specified collections, rather than stripping them away")
	decompress = flag.Bool("d", false, "decompress the stdin input with gzip")
	compress = flag.Bool("c", false, "compress the stdout output with gzip")
)

func printUsage() {
	fmt.Fprintf(os.Stderr,
		`Usage: eicio-strip [options] <eicio-input-file> <collections>...
options:
`,
	)
	flag.PrintDefaults()
}

func main() {
	flag.Usage = printUsage
	flag.Parse()

	if flag.NArg() < 1 {
		printUsage()
		log.Fatal("Invalid arguments")
	}

	var reader *eicio.Reader
	var err error

	filename := flag.Arg(0)
	if filename == "-" {
		stdin := bufio.NewReader(os.Stdin)
		if *decompress {
			reader, err = eicio.NewGzipReader(stdin)
		} else {
			reader = eicio.NewReader(stdin)
		}
	} else {
		reader, err = eicio.Open(filename)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	var writer *eicio.Writer
	if *outFile == "" {
		if *compress {
			writer = eicio.NewGzipWriter(os.Stdout)
		} else {
			writer = eicio.NewWriter(os.Stdout)
		}
	} else {
		writer, err = eicio.Create(*outFile)
		if err != nil {
			log.Fatal(err)
		}
	}
	defer writer.Close()

	var colls []string
	for i := 1; i < flag.NArg(); i++ {
		colls = append(colls, flag.Arg(i))
	}

	nEventsRead := 0

	for event := range reader.Events() {
		if reader.Err != nil {
			log.Print(reader.Err)
		}

		for _, collName := range event.GetNames() {
			if *keep {
				keepThis := false
				for _, keepName := range colls {
					if collName == keepName {
						keepThis = true
						break
					}
				}
				if !keepThis {
					event.Remove(collName)
				}
			} else {
				for _, removeName := range colls {
					if collName == removeName {
						event.Remove(collName)
					}
				}
			}
		}

		if err := writer.Push(event); err != nil {
			log.Fatal(err)
		}

		nEventsRead++
	}

	if (reader.Err != nil && reader.Err != io.EOF) || nEventsRead == 0 {
		log.Print(reader.Err)
	}
}
