package main

import (
	"fmt"
	"net"
	"syscall"
)

func main() {
	// create a lil IPv4 UDP socket
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, 0)
	if err != nil{
		panic(err)
	}
	defer syscall.Close(fd) // close the file descriptor

	// parse the server IP as IPv4 (hardcoded)
	ip := net.ParseIP().To4()

	// create a local socket IPv4 at port 5000
	addr := &syscall.SockaddrInet4{
		Port: 5000,
	}

	// copy the destination IP into socket addr info struct stuff
	copy(addr.Addr[:], ip)

	message := []byte("Hello from Go!")
	
	// send the message to destination address
	err = syscall.Sendto(fd, message, 0, addr)
	if err != nil{
		panic(err)
	}

	// create buffer of 1024 bytes
	buffer := make([]byte, 1024)

	// receive data from incoming packet within buffer
	n, _, err := syscall.Recvfrom(fd, buffer, 0)
	if err != nil{
		panic(err)
	}

	fmt.Println(string[buffer[:n]))
}
