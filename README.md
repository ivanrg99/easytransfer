![Easy Transfer](https://github.com/ivanrg99/easytransfer/assets/47028390/17705219-6164-4637-b162-0d1e01f77c3f)

Easily transfer and stream files in a network over TCP

# Usage
Start the server on localhost, port 8080, with a chunk size of 500MB and a destination folder of . (the current directory) 
```bash
easytransfer -address localhost:8080 -chunk 500 -destination .
```

# Building
Just run 
```bash
go build .
```
on the backend folder for the server binary, or the client folder for the client binary

# Releases
You can find pre-built binaries for Windows amd64 and Linux amd64 (built on WSL) on the releases page
