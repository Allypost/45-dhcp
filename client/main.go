package client

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/netip"
	"os"
	"os/signal"
	"strings"
	"time"

	"DHCP5x9/util/ports"
)

type clientConfig struct {
	Port int16
}

func randBetween(min, max int) int {
	return rand.Intn(max-min) + min
}

func parseArgs() (*clientConfig, error) {
	if len(os.Args) != 2 {
		log.Fatalln("Usage: client <server port>")
	}

	// Parse PORT
	port, err := ports.ParsePort(os.Args[1])
	if err != nil {
		return nil, err
	}

	return &clientConfig{
		Port: int16(port),
	}, nil
}

func connectToServer(config *clientConfig) (*net.UDPConn, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("localhost:%d", config.Port))
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, err
	}

	return conn, err
}

func send(conn *net.UDPConn, msg string) error {
	fmt.Printf(">> %s\n", msg)

	_, err := conn.Write([]byte(msg))

	return err
}

func recv(conn *net.UDPConn) string {
	buffer := make([]byte, 512)
	n, _, _ := conn.ReadFromUDP(buffer)
	buffer = buffer[:n]

	fmt.Printf("<< %s\n", buffer)

	return string(buffer)
}

func validateResponse(respParts []string) error {
	if len(respParts) != 2 {
		return fmt.Errorf("invalid response: %s", strings.Join(respParts, " "))
	}

	if respParts[0] != "OFFER" {
		return fmt.Errorf("invalid response: %s", strings.Join(respParts, " "))
	}

	if _, err := netip.ParseAddr(respParts[1]); err != nil {
		return err
	}

	return nil
}

func waitToDie(sleepFor time.Duration) {
	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt, os.Kill)

	select {
	case <-interrupts:
	case <-time.After(sleepFor * time.Second):
	}
}

func startClient(config *clientConfig) error {
	conn, err := connectToServer(config)
	if err != nil {
		return err
	}
	defer conn.Close()

	err = send(conn, "REQUEST")
	if err != nil {
		return err
	}

	resp := recv(conn)
	respParts := strings.SplitN(resp, " ", 2)
	if err := validateResponse(respParts); err != nil {
		return err
	}

	err = send(conn, "CONFIRM")

	fmt.Printf("IP: %s\n", respParts[1])

	sleepFor := randBetween(5, 10)
	waitToDie(time.Duration(sleepFor))

	return send(conn, "RELEASE")
}

func Run() {
	config, err := parseArgs()
	if err != nil {
		log.Fatalln(err)
	}

	err = startClient(config)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Client exited")
}
