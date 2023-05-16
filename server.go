package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func handleFile(conn net.Conn, dest string, chunkSize int) {
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
	var totalWritten int64
	var totalBuffer int

	for totalWritten < size {
		fmt.Printf("Buffer CAP: %d\n", cap(fileContents))
		for totalBuffer < cap(fileContents) {
			log.Println("Reading")
			n, err := conn.Read(fileContents[totalBuffer:])
			if err != nil {
				if err != io.EOF {
					log.Panicf("Could not read file contents for file [%s]\n", fileName)
				}
			}
			totalBuffer += n
			totalRead += int64(n)
			if totalRead == size {
				log.Println("Read everything!")
				break
			}
		}
		log.Println("Done reading. Buffer should be full")
		totalBufferLimit := totalBuffer
		totalBuffer = 0

		log.Printf("File [%s] \n\t| %.2f%%", fileName, (float64(totalRead)/float64(size))*100.0)

		for totalBuffer < cap(fileContents) {
			log.Println("Writing...")
			n, err := f.Write(fileContents[totalBuffer:totalBufferLimit])
			fmt.Printf("Current written N: %d\n", n/1048576)
			totalBuffer += n
			totalWritten += int64(n)
			if err != nil {
				log.Panicf("Could not write file contents for file [%s]\n", fileName)
			}
			if totalWritten == size {
				log.Println("Wrote everything!")
				break
			}
			fmt.Printf("Total written: %d\n", totalWritten/1048576)
		}
		log.Println("Done writing. Buffer should be full")
		totalBuffer = 0
		fmt.Printf("Total read: %d\n", totalRead/1048576)
		fmt.Printf("Total written: %d\n", totalWritten/1048576)
		fmt.Println("Redoing loop...")
	}
	fmt.Println("DONE. Sending ACK...")
	log.Printf("File [%s] DONE\n", fileName)
	conn.Write([]byte("ok"))
}

func startServer(f *flags) {
	// Start listening
	l, err := net.Listen("tcp", *f.addr)
	if err != nil {
		log.Fatal("Failed to bind to address: ", *f.addr)
	}

	fmt.Printf("EasyTransfer running on %s with streaming size of %dMB, destination folder: %s   | ...\n", *f.addr, *f.chunkSize, *f.destFolder)

	// Accept concurrent connections
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleFile(conn, *f.destFolder, *f.chunkSize)
	}
}
