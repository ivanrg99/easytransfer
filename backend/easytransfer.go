package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func handleFile(conn net.Conn, dest string, chunkSize int64) {
	defer conn.Close()
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("TERMINATED READING FILE")
		}
	}()
	// File format is:
	// 64 bits: File size = Z
	// 64 bits: Filename size = N
	// N  bits: Filename string
	// Z  bits: File contents
	// TODO! Also send checksum and check it

	log.Println("Connection established...")

	// Read sizes
	var size int64
	var nameSize int64
	err := binary.Read(conn, binary.LittleEndian, &size)
	if err != nil {
		log.Panicln("Could not read file size from incoming connection")
	}
	err = binary.Read(conn, binary.LittleEndian, &nameSize)
	if err != nil {
		log.Panicln("Could not read fileName size from incoming connection")
	}

	// Get fileName
	fileName := new(bytes.Buffer)
	for {
		n, err := io.CopyN(fileName, conn, nameSize)
		if err != nil {
			log.Panicln("Could not read fileName from incoming connection")
		}
		if n == nameSize {
			break
		}
	}

	log.Printf("Receiving file [%s]\n\t| Size: %.2fMB\n", fileName, float64(size)/1_000_000)

	// Create destination file
	fName := dest + string(os.PathSeparator) + fileName.String()

	f, err := os.Create(fName)
	if err != nil {
		log.Panicln(err)
	}
	defer f.Close()

	// Get file contents (stream 50MB)
	fileContents := make([]byte, chunkSize*1048576)
	var totalRead int64
	for {
		n, err := conn.Read(fileContents)
		if err != nil {
			log.Panicf("Could not read file contents for file [%s]\n", fileName)
		}

		totalRead += int64(n)
		log.Printf("File [%s] \n\t| %.2f%%", fileName, (float64(totalRead)/float64(size))*100.0)

		// TODO! Might handle this later
		_, err = f.Write(fileContents[:n])
		if err != nil {
			log.Panicf("Could not write file contents for file [%s]\n", fileName)
		}
		if totalRead == size {
			break
		}
	}
	log.Printf("File [%s] DONE\n", fileName)
	conn.Write([]byte("ok"))
}

func main() {
	// Get address
	addr := flag.String("address", "localhost:3287", "Address:Port at which the server will bind")

	// Set buffer size for streaming
	chunkSize := flag.Int64("chunk", 50, "Size in MB of chunks size to be used as the streaming buffer")

	// Get destination folder for this run
	dest, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	dest = dest + string(os.PathSeparator) + "Downloads"

	folder := flag.String("destination", dest, "Destination folder where the files will be written to")

	flag.Parse()

	// Start listening
	l, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("EasyTransfer running on %s with streaming size of %dMB, destination folder: %s   | ...\n", *addr, *chunkSize, *folder)

	// Accept concurrent connections
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleFile(conn, *folder, *chunkSize)
	}
}
