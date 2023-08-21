![Easy Transfer](https://github.com/ivanrg99/easytransfer/assets/47028390/17705219-6164-4637-b162-0d1e01f77c3f)

Easily transfer and stream files in a network over TCP

[![Go Report Card](https://goreportcard.com/badge/github.com/ivanrg99/easytransfer)](https://goreportcard.com/report/github.com/ivanrg99/easytransfer)
# Usage
Start the server on localhost, port 8080, with a chunk size of 500MB and a destination folder of . (the current directory) 
```bash
easytransfer -serve -address localhost:8080 -chunk 500 -destination .
```
To start the client and send the files to the address 192.168.10.87 and port 8080, with a chunk size of 500MB
```bash
easytransfer -address 192.168.10.87:8080 -chunk 500 file1 file2 file3...
```
To get more help:
```bash
easytransfer --help
```
# Overview
The server will put all received files into the destination path specified at startup. The server can handle multiple concurrent requests.
The client will send all specified files to the easytransfer server running on the given address and port number at the same time.

### What is the chunk size?
The chunk size of the server is the buffer that easytransfer will allocate for each concurrent file upload request. The bigger the chunk size, the bigger the buffer and the faster it will perform. Easytransfer reads data until the buffer is full, and then flushes the data into the destination file on disk. Disk I/O is expensive (so is logging, and we'll have an option to disable it in the future), so a bigger chunk size makes it more performant by writing to disk less often. 

⚠️ For each concurrent file upload, a buffer is allocated of the specified chunksize. If you specify a chunk size of 1GB and send 10 files at the same time from the client, the server will allocate 10GB of memory!
Easytransfer gives you control of the chunk size used to stream the file, so use that to your advantage when passing files from computers. You can restart and spawn the server easily to readjust the chunk size as needed, as well as specify a different destination folder.

The same is true for the chunk size of the client. For each file that it sends concurrently, it will allocate a buffer with the given chunk size, and use it to stream the file to the server over the network.

# Building
Just run 
```bash
go build .
```
# Releases
You can find pre-built binaries for Windows amd64 and Linux amd64 (built on WSL) on the releases page

# Roadmap
- Make a GUI version (already in the works)
- Support sending folders
- Add better error handling (timeouts when writing to server for ex)
- Add documentation and tests
- Add better progress report inside buffer reading (instead of 1 update per full buffer read)
- Cleanup code a bit more
- Check in header that both client and server have the same chunk size
