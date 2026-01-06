package main

import (
	"log"
	"net"
)

// GetOutboundIP is a clever way to find the local IP address of your machine.
// Instead of listing ALL network interfaces (which can be many and confusing),
// we attempt to "dial" a public IP (Google's DNS 8.8.8.8) using UDP.
// This doesn't actually send a packet or require internet; it just asks the
// OS: "If I wanted to send a packet to this address, which of my local IPs would I use?"
func GetOutboundIP() string {

	// We use UDP (connectionless) so it doesn't wait for a handshake.
	// It just picks the interface that would be used to reach the internet.
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close() // Ensure we clean up the connection resource

	// Ask the connection for its local address.
	// We cast it to *net.UDPAddr so we can access the .IP field.
	localAddr := conn.LocalAddr().(*net.UDPAddr)

	// Return the IP as a string (e.g., "192.168.1.15")
	return localAddr.IP.String()
}
