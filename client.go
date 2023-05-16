package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

func StartClient(f *flags) {
	filenames := flag.Args()

	if len(filenames) == 0 {
		fmt.Println("Error: At least one filename must be provided")
		os.Exit(1)
	}

	done := make(chan bool)
	for _, filename := range filenames {
		go func(filename string) {
			fc := NewFileClient(filename, *f.addr, *f.chunkSize)
			fc.SendFile()
			done <- true
		}(filename)
	}
	for i := 0; i < len(filenames); i++ {
		<-done
	}
}

type FileClient struct {
	conn         net.Conn
	info         os.FileInfo
	f            *os.File
	fileContents []byte
	s            string
}

func NewFileClient(filePath, addr string, chunkSize int) *FileClient {
	fc := FileClient{}
	fc.initialize(filePath, addr, chunkSize)
	return &fc
}

func (fc *FileClient) SendFile() {
	defer fc.Close()

	fc.sendHeader()
	fc.sendBody()
	fc.waitForAck()
}

func (fc *FileClient) Close() {
	fc.conn.Close()
	fc.f.Close()
}

func (fc *FileClient) initialize(filePath, addr string, chunkSize int) {
	// Get file details
	var err error
	fc.f, err = os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open file: %s\n", filePath)
	}

	fc.info, _ = fc.f.Stat()
	if fc.info.IsDir() {
		fc.f.Close()
		log.Fatalf("%s is a directory, not a file!", filePath)
	}

	// Connect to server
	d := net.Dialer{Timeout: 10 * time.Second}
	fc.conn, err = d.Dial("tcp", addr)
	if err != nil {
		fc.f.Close()
		log.Fatalf("Failed to connect to: %s\n", addr)
	}
	fc.fileContents = make([]byte, chunkSize*toBytes)
}

func (fc *FileClient) sendHeader() {
	// File format is:
	// 64 bits: File size = Z
	// 64 bits: Filename size = N
	// N  bits: Filename string
	// Z  bits: File contents
	err := binary.Write(fc.conn, binary.LittleEndian, fc.info.Size())
	if err != nil {
		log.Fatalf("Failed to write file size of [%s] to server\n", fc.info.Name())
	}

	err = binary.Write(fc.conn, binary.LittleEndian, int64(len(fc.info.Name())))
	if err != nil {
		log.Fatalf("Failed to write filename size of [%s] to server\n", fc.info.Name())
	}

	_, err = fmt.Fprintf(fc.conn, fc.info.Name())
	if err != nil {
		log.Fatalf("Failed to write filename of [%s] to server\n", fc.info.Name())
	}
}

func (fc *FileClient) sendBody() {
	// Start streaming file in chunks
	var totalRead int64
	for {
		n, err := fc.f.Read(fc.fileContents)
		if err != nil {
			log.Panicf("Failed to read file contents of [%s] into the buffer\n", fc.info.Name())
		}
		totalRead += int64(n)
		log.Printf("File [%s]\n\t| %.2f%%\n", fc.info.Name(), (float64(totalRead)/float64(fc.info.Size()))*100.0)

		err = binary.Write(fc.conn, binary.LittleEndian, fc.fileContents[:n])
		if err != nil {
			log.Panicf("Failed to write file contents of [%s] to server\n", fc.info.Name())
		}
		if totalRead == fc.info.Size() {
			log.Printf("File [%s] Waiting for server to confirm that it got the file...\n", fc.info.Name())
			break
		}
	}
}

func (fc *FileClient) waitForAck() {
	ack := make([]byte, 1)
	_, err := fc.conn.Read(ack)
	if err != nil {
		log.Panicln("Could not receive ack from server")
	}
	if string(ack) != "y" {
		log.Panicln("Server did not ack")
	}
	log.Printf("File [%s] DONE\n", fc.info.Name())
}
