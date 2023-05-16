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

func StartServer(f *flags) {
	// Start listening
	l, err := net.Listen("tcp", *f.addr)
	if err != nil {
		log.Fatal("Failed to bind to address: ", *f.addr)
	}

	fmt.Printf("EasyTransfer running on %s with chunk size of %dMB, destination folder: \"%s\"\t...\n", *f.addr, *f.chunkSize, *f.destFolder)

	// Accept concurrent connections
	for {
		conn, err := l.Accept()
		log.Println("Connection established...")
		if err != nil {
			log.Fatal(err)
		}
		fs := NewFileServer(conn, *f.destFolder, *f.chunkSize)
		go fs.HandleFile()
	}
}

type FileInfo struct {
	fileSize int64
	nameSize int64
	fileName string
}

type FileServer struct {
	conn         net.Conn
	dest         string
	info         FileInfo
	file         *os.File
	fileContents []byte
}

func NewFileServer(conn net.Conn, dest string, chunkSize int) *FileServer {
	return &FileServer{
		conn:         conn,
		dest:         dest,
		fileContents: make([]byte, chunkSize*toBytes),
		info: FileInfo{
			fileSize: 0,
			nameSize: 0,
			fileName: "",
		},
	}
}

func (fs *FileServer) Close() {
	_ = fs.file.Close()
	_ = fs.conn.Close()
}

func (fs *FileServer) fetchFileInfoSizes() {
	// Read sizes
	err := binary.Read(fs.conn, binary.LittleEndian, &fs.info.fileSize)
	if err != nil {
		log.Panicln("Could not read file size from incoming connection")
	}
	err = binary.Read(fs.conn, binary.LittleEndian, &fs.info.nameSize)
	if err != nil {
		log.Panicln("Could not read fileName size from incoming connection")
	}
}

func (fs *FileServer) fetchFileInfoName() {
	// Get fileName
	fileName := new(bytes.Buffer)
	for {
		n, err := io.CopyN(fileName, fs.conn, fs.info.nameSize)
		if err != nil {
			log.Panicln("Could not read fileName from incoming connection")
		}
		if n == fs.info.nameSize {
			break
		}
	}
	fs.info.fileName = fileName.String()
}

func (fs *FileServer) initFileInfo() {
	fs.fetchFileInfoSizes()
	fs.fetchFileInfoName()
}

func (fs *FileServer) createNewFile() {
	// Create destination file
	fName := fs.dest + string(os.PathSeparator) + fs.info.fileName

	var err error
	fs.file, err = os.Create(fName)
	if err != nil {
		log.Panicln("Could not create destination file in: ", fName)
	}
}

func (fs *FileServer) parseHeader() {
	// File format is:
	// 64 bits: File size = Z
	// 64 bits: Filename size = N
	// N  bits: Filename string
	// Z  bits: File contents

	fs.initFileInfo()

	log.Printf("Receiving file [%s]\n\t| Size: %.2fMB\n", fs.info.fileName, float64(fs.info.fileSize)/1_000_000)

	fs.createNewFile()
}

func (fs *FileServer) HandleFile() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("TERMINATED READING FILE")
		}
	}()
	defer fs.Close()

	fs.parseHeader()
	fs.parseBody()
}

// Make this into a struct method for a new buffer struct (maybe)
func (fs *FileServer) parseBody() {
	// Get file contents
	var totalRead int64
	var totalWritten int64
	var totalBuffer int

	for totalWritten < fs.info.fileSize {
		for totalBuffer < cap(fs.fileContents) {
			n, err := fs.conn.Read(fs.fileContents[totalBuffer:])
			if err != nil {
				if err != io.EOF {
					log.Panicf("Could not read file contents for file [%s]\n", fs.info.fileName)
				}
			}
			totalBuffer += n
			totalRead += int64(n)
			if totalRead == fs.info.fileSize {
				break
			}
		}
		totalBufferLimit := totalBuffer
		totalBuffer = 0

		log.Printf("File [%s] \n\t| %.2f%%", fs.info.fileName, (float64(totalRead)/float64(fs.info.fileSize))*100.0)

		for totalBuffer < cap(fs.fileContents) {
			n, err := fs.file.Write(fs.fileContents[totalBuffer:totalBufferLimit])
			totalBuffer += n
			totalWritten += int64(n)
			if err != nil {
				log.Panicf("Could not write file contents for file [%s]\n", fs.info.fileName)
			}
			if totalWritten == fs.info.fileSize {
				break
			}
		}
		totalBuffer = 0
	}
	fs.conn.Write([]byte("y"))
	log.Printf("File [%s] DONE\n", fs.info.fileName)
}
