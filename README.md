# easytransfer
Easily transfer files in a local network over http. Multiplatform
You need to have the server open in both machines.

## Client
For the time being, just a multithreaded Rust CLI. The code is pretty straightforward, a GUI frontend could be made easily

## Server
For the time being, Spring Boot based. Can be compiled with GraalVM for immediate startups and ease of packaging. The server needs to be up in all machines that want to use EasyTransfer. You could add the server to run at startup.


### Roadmap
- Improve error handling
- Better server logging
- Make server non-blocking (could use Spring Webflux, but will probably rewrite in Go for less memory and CPU usage and ease of distribution)
- Make it more configurable (Default port is 3287, can't be changed. Defaul destination folder for files sent with EasyTransfer is the Downloads folder, can't be changed).
- As features grow, consider adding tests (not relevant now)
