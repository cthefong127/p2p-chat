package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

func main() {
	// create a lil IPv4 UDP socket
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, 0)
	if err != nil {
		panic(err)
	}
	defer syscall.Close(fd) // close the file descriptor

	// step 1: register w/ bootstrap server using same fd
	// fd used for receiving PEER msg & talking to peer to keep local port constant
	// send the message to destination address
	if err := syscall.Sendto(fd, []byte("Hello from Go!"), 0, &syscall.SockaddrInet4{
		Port: 5000,
		Addr: [4]byte{35, 252, 138, 18},
	}); err != nil {
		panic(err)
	}

	// create buffer of 1024 bytes
	buffer := make([]byte, 1024)

	var peerAddr *syscall.SockaddrInet4

	// step 2: wait for bootstrap server to introduce a peer
	// any packet not matching:
	//			"PEER <ip>:<port>"
	// is ignored for now (later implement retries/timeouts)
	// receive data from incoming packet within buffer
	for peerAddr == nil {
		n, from, err := syscall.Recvfrom(fd, buffer, 0)
		if err != nil {
			panic(err)
		}

		msg := string(buffer[:n]) // parse buffer as string
		fmt.Printf("Received %q from %v\n", msg, from)

		if strings.HasPrefix(msg, "PEER ") {
			parsed, err := parsePeerAddr(strings.TrimPrefix(msg, "PEER "))
			if err != nil {
				fmt.Println("could not parse peer address:", err)
				continue
			}
			peerAddr = parsed
			fmt.Printf("Learned peer address: %d.%d.%d.%d:%d\n",
				peerAddr.Addr[0], peerAddr.Addr[1], peerAddr.Addr[2], peerAddr.Addr[3], peerAddr.Port)
		}
	}

	// step 3: peer connection established, no longer thru server
	// "PUNCH" packet sent as placeholder for real hole punching, peer does same thing simultaneously
	// fine if first packet silently dropped by peer's NAT
	if err := syscall.Sendto(fd, []byte("PUNCH"), 0, peerAddr); err != nil {
		panic(err)
	}
	fmt.Println("Sent punch packet directly to peer")

	// step 4: consistently listen to whatever is sent back in concurrent goroutine (yay p2p)
	go func() {
		recvBuf := make([]byte, 1024) // separate buffer since this now runs concurrently with the sender
		for {
			n, from, err := syscall.Recvfrom(fd, recvBuf, 0)
			if err != nil {
				fmt.Println(err)
				continue
			}
			sa, ok := from.(*syscall.SockaddrInet4)
			if !ok {
				fmt.Println("not an IPv4")
			}

			incoming_ip := fmt.Sprintf("%d.%d.%d.%d:%d", sa.Addr[0], sa.Addr[1], sa.Addr[2], sa.Addr[3], sa.Port)
			fmt.Printf("\nPeer says: %s (from %s)\n> ", string(recvBuf[:n]), incoming_ip)
		}
	}()

	// step 4.5: infinite for loop to check for sending messages
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			continue // don't send empty lines
		}

		if err := syscall.Sendto(fd, []byte(input), 0, peerAddr); err != nil {
			panic(err)
		}
	}
}

// parsePeerAddr: turns an "ip:port" string (as sent by bootstrap server in
// a "PEER <ip:<port>" message) into a *sysscall.SockaddrInet4 to pass to Sendto
func parsePeerAddr(s string) (*syscall.SockaddrInet4, error) {
	host, portStr, found := strings.Cut(s, ":")
	if !found {
		return nil, fmt.Errorf("invalid peer address %q: missing port", s)
	}

	octets := strings.Split(host, ".")
	if len(octets) != 4 {
		return nil, fmt.Errorf("invalid ip %q", host)
	}

	var addr [4]byte
	for i, o := range octets {
		v, err := strconv.Atoi(o)
		if err != nil || v < 0 || v > 255 {
			return nil, fmt.Errorf("invalid ip octet %q", o)
		}
		addr[i] = byte(v)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port %q", portStr)
	}

	return &syscall.SockaddrInet4{Port: port, Addr: addr}, nil
}
