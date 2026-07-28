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

	// registry: in-memory list of clients connected so far
	registry := make(map[string]*syscall.SockaddrInet4)

	// repeatedly check for new data in buffer indefinitely (are there new messages?)
	for {
		// get the incoming socket's address length and the address itself
		n, clientAddr, err := syscall.Recvfrom(fd, buffer, 0)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// ascertain the incoming socket is IPv4, ignore otherwise
		v4Addr, ok := clientAddr.(*syscall.SockaddrInet4)
		if !ok {
			fmt.Println("ignoring non-IPv4 sender")
			continue
		}

		// create a "ip:port" key for the sender as seen publicly
		key := fmt.Sprintf("%d.%d.%d.%d:%d",
			v4Addr.Addr[0], v4Addr.Addr[1], v4Addr.Addr[2], v4Addr.Addr[3], v4Addr.Port)

		fmt.Printf("Received %q from %s\n", string(buffer[:n]), key)

		// register client if the ip:port is not seen before
		if _, exists := registry[key]; !exists {
			registry[key] = v4Addr
			fmt.Printf("Registered new peer: %s (waiting: %d)\n", key, len(registry))
		}

		// after two unique clients registered, introduce them to one another
		if len(registry) >= 2 {
			var keys []string
			var addrs []*syscall.SockaddrInet4
			for k, a := range registry {
				keys = append(keys, k)
				addrs = append(addrs, a)
				if len(keys) == 2 {
					break
				}
			}

			// simple text protocol: "PEER <ip>:<port>" tells a client the
			// address of the other client it should try to reach directly.
			msgFor0 := []byte(fmt.Sprintf("PEER %s", keys[1]))
			msgFor1 := []byte(fmt.Sprintf("PEER %s", keys[0]))

			if err := syscall.Sendto(fd, msgFor0, 0, addrs[0]); err != nil {
				fmt.Println("send err:", err)
			}
			if err := syscall.Sendto(fd, msgFor1, 0, addrs[1]); err != nil {
				fmt.Println("send err:", err)
			}

			fmt.Printf("Introduced %s <-> %s\n", keys[0], keys[1])

			// remove this pair from the registry so the server is ready to
			// pair up the next two unique clients that show up, server no longer involved
			delete(registry, keys[0])
			delete(registry, keys[1])
		}

	}
}
