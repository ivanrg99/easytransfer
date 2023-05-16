package main

import (
	"flag"
	"log"
	"net"
	"os"
	"strings"
)

const toBytes = 1048576

type flags struct {
	addr       *string
	chunkSize  *int
	destFolder *string
	isServer   *bool
}

func main() {
	flags := parseFlags()
	if *flags.isServer {
		StartServer(&flags)
	} else {
		StartClient(&flags)
	}
}

func parseFlags() flags {
	flags := flags{}
	flags.isServer = flag.Bool("serve", false, "Start easytransfer in server mode. If not, it will start as a client")

	// Get address (todo! Make this use actual default address by using dial and show it when initializing the server)
	flags.addr = flag.String("address", getDefaultAddress(), "Address:Port at which the server will bind to / client will connect to")

	// Set buffer size for streaming
	flags.chunkSize = flag.Int("chunk", 100, "Size in MB of chunks size to be used as the streaming buffer (bigger might improve performance)")

	// Get destination folder for this run
	dest, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Failed to find user home directory to set as default destination folder")
	}
	dest = dest + string(os.PathSeparator) + "Downloads"

	flags.destFolder = flag.String("destination", dest, "Destination folder where the files will be written to (client only)")

	flag.Parse()

	return flags
}

func getDefaultAddress() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal("Could not get default IP for a possible server")
	}
	defer conn.Close()

	addr := conn.LocalAddr().(*net.UDPAddr).String()
	i := strings.Index(addr, ":")

	return addr[:i] + ":3287"
}
