package main

import (
	"log"
	"net" // The Network library
)

func GetOutboundIP() string {
	// 1. Dial a dummy connection to Google DNS
	// "udp" means we don't actually handshake or send data.
	// We just check the route.

	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close() // Close it when we are done

	// 2. Ask the connection: "Who am I?"
	localAddr := conn.LocalAddr().(*net.UDPAddr)

	// 3. Return the string (e.g., "192.168.1.50")
	return localAddr.IP.String()
}
