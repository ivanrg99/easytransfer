package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

func sendFile(filePath string, addr string, chunkSize int64) {
	// Get file details
	f, err := os.Open(filePath)
	defer f.Close()
	if err != nil {
		log.Fatalf("Failed to open file: %s\n", filePath)
	}

	info, _ := f.Stat()
	if info.IsDir() {
		log.Fatalf("%s is a directory, not a file!", filePath)
	}

	// Connect to server
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to connect to: %s\n", addr)
	}

	// File format is:
	// 64 bits: File size = Z
	// 64 bits: Filename size = N
	// N  bits: Filename string
	// Z  bits: File contents
	// TODO! Also send checksum and check it
	err = binary.Write(conn, binary.LittleEndian, info.Size())
	if err != nil {
		log.Fatalf("Failed to write file size of [%s] to: %s\n", filePath, addr)
	}

	err = binary.Write(conn, binary.LittleEndian, int64(len(info.Name())))
	if err != nil {
		log.Fatalf("Failed to write name size of [%s] to: %s\n", filePath, addr)
	}

	_, err = fmt.Fprintf(conn, info.Name())
	if err != nil {
		log.Fatalf("Failed to write filename of [%s] to: %s\n", filePath, addr)
	}

	// Start streaming file in chunks of 50MB
	fileContents := make([]byte, chunkSize*1048576)
	var totalRead int64

	for {
		n, err := f.Read(fileContents)
		if err != nil {
			log.Panicf("Failed to read file contents of [%s] into the buffer\n", filePath)
		}
		totalRead += int64(n)
		log.Printf("File [%s]\n\t| %.2f%%\n", filePath, (float64(totalRead)/float64(info.Size()))*100.0)

		err = binary.Write(conn, binary.LittleEndian, fileContents[:n])
		if err != nil {
			log.Panicf("Failed to write file contents of [%s] to: %s\n", filePath, addr)
		}
		if totalRead == info.Size() {
			log.Printf("File [%s] DONE\n", filePath)
			break
		}
	}
	ack := make([]byte, 2)
	conn.Read(ack)
	if string(ack) != "ok" {
		log.Panicln("Server did not ack")
	}
}

func main() {
	// Get address and streaming chunk size
	addr := flag.String("address", "localhost:3287", "Address:Port at which the client will try to connect to send the files")
	chunkSize := flag.Int64("chunk", 50, "Size in MB of chunks size to be streamed to the address")

	flag.Parse()
	filenames := flag.Args()

	if len(filenames) == 0 {
		fmt.Println("Error: At least one filename must be provided")
		os.Exit(1)
	}

	done := make(chan bool)
	for _, filename := range filenames {
		go func(filename string) {
			sendFile(filename, *addr, *chunkSize)
			done <- true
		}(filename)
	}
	for i := 0; i < len(filenames); i++ {
		<-done
	}
}
