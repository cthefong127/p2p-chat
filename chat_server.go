package main

import (
	"fmt"
	"syscall"
)

func main() {
	// create a IPv4 UDP socket
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, 0)
	if err != nil { // if error not null then throw error
		panic(err)
	}
	defer syscall.Close(fd) // close the fd

	addr := &syscall.SockaddrInet4{
		Port: 5000,
	}

	// Binds UDP socket
	err = syscall.Bind(fd, addr)
	if err != nil {
		panic(err)
	}

	fmt.Println("Listening on UDP port 5000")

	// create buffer of 1024 bytes
	buffer := make([]byte, 1024)

	// repeatedly check for new data in buffer indefinitely (are there new messages?)
	for {
		// get the incoming socket's address length and the address itself
		n, clientAddr, err := syscall.Recvfrom(fd, buffer, 0)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// print out received data
		fmt.Printf("Received: %s\n", string(buffer[:n]))

		response := []byte("Packet received!")

		// RETURN TO SENDER
		err = syscall.Sendto(fd, response, 0, clientAddr)
		if err != nil {
			fmt.Println(err)
		}
	}
}
