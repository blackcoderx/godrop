package main

import (
	"log"
	"net" // The Network library
)

func GetOutboundIP() string {

	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close() // Close it when we are done

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}
