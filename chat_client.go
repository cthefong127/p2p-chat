package main

import (
	"fmt"
	"syscall"
)

func main() {
	// create a lil IPv4 UDP socket
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, 0)
	if err != nil {
		panic(err)
	}
	defer syscall.Close(fd) // close the file descriptor

	// send the message to destination address
	if err := syscall.Sendto(fd, []byte("Hello from Go!"), 0, &syscall.SockaddrInet4{
		Port: 5000,
		Addr: [4]byte{34, 83, 0, 251},
	}); err != nil {
		panic(err)
	}

	// create buffer of 1024 bytes
	buffer := make([]byte, 1024)

	// receive data from incoming packet within buffer
	n, _, err := syscall.Recvfrom(fd, buffer, 0)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(buffer[:n]))
}
