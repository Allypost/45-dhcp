package server

import (
	"fmt"
	"log"
	"net"
	"net/netip"
	"os"
	"strings"
	"time"

	"DHCP5x9/util/ports"
	"DHCP5x9/util/sortedIpHeap"
)

type serverConfig struct {
	Host   string
	Port   int16
	Subnet netip.Prefix
}

type ipRent struct {
	ip        netip.Addr
	confirmed bool
}

func parseArgs() (*serverConfig, error) {
	if len(os.Args) != 3 {
		log.Fatal("Usage: server <subnet> <port>")
	}

	// Parse SUBNET
	subnet, err := netip.ParsePrefix(os.Args[1])
	if err != nil {
		return nil, err
	}

	// Parse PORT
	port, err := ports.ParsePort(os.Args[2])
	if err != nil {
		return nil, err
	}

	// Parse HOST
	host := os.Getenv("HOST")
	if _, err := netip.ParseAddr(host); err != nil {
		host = "127.0.0.1"
	}

	return &serverConfig{
		Host:   host,
		Port:   port,
		Subnet: subnet,
	}, nil
}

func ipSequence(initialAddr netip.Addr, subnet netip.Prefix) <-chan netip.Addr {
	ch := make(chan netip.Addr)
	go func() {
		for addr := initialAddr; subnet.Contains(addr); addr = addr.Next() {
			ch <- addr
		}
		close(ch)
	}()
	return ch
}

func send(conn *net.UDPConn, addr *net.UDPAddr, msg string) error {
	_, err := conn.WriteToUDP([]byte(msg), addr)

	log.Printf(">> (%s) %s\n", addr, msg)

	return err
}

func recv(conn *net.UDPConn, buff []byte) (string, *net.UDPAddr) {
	n, addr, _ := conn.ReadFromUDP(buff)
	buff = buff[:n]

	log.Printf("<< (%s) %s \n", addr, buff)

	return strings.TrimSpace(string(buff)), addr
}

func startServer(config *serverConfig) error {
	s, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", config.Host, config.Port))
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", s)
	if err != nil {
		return err
	}
	defer conn.Close()

	log.Printf("Started server with parameters: Host=%s Port=%d Subnet=%s\n", config.Host, config.Port, config.Subnet)

	ipGenerator := ipSequence(config.Subnet.Addr(), config.Subnet)
	ipBacklog := sortedIpHeap.New()
	addrToIp := make(map[string]ipRent)

	commands := make(map[string]func(*net.UDPAddr, string, string) error)

	commands["REQUEST"] = func(addr *net.UDPAddr, token string, msg string) error {
		var ip netip.Addr
		if ipBacklog.Len() > 0 {
			ip = ipBacklog.Pop().(netip.Addr)
		} else {
			ip = <-ipGenerator
		}

		if !ip.IsValid() {
			send(conn, addr, "DENY no more IPs available")
			return nil
		}

		addrToIp[token] = ipRent{
			ip,
			false,
		}
		go func() {
			time.Sleep(10 * time.Second)
			item, ok := addrToIp[token]
			if ok && !item.confirmed {
				commands["RELEASE"](addr, token, "")
			}
		}()

		payload := fmt.Sprintf("OFFER %s", ip)
		err = send(conn, addr, payload)
		if err != nil {
			return err
		}

		return nil
	}

	commands["RELEASE"] = func(addr *net.UDPAddr, token string, msg string) error {
		if item, ok := addrToIp[token]; ok {
			ipBacklog.Push(item.ip)
			delete(addrToIp, token)
		}

		return nil
	}

	commands["CONFIRM"] = func(addr *net.UDPAddr, token string, msg string) error {
		item, ok := addrToIp[token]
		if ok {
			item.confirmed = true
			addrToIp[token] = item
		}

		return nil
	}

	buffer := make([]byte, 1024)
	for {
		msg, addr := recv(conn, buffer)
		token := addr.String()
		cmd, msg, _ := strings.Cut(msg, " ")

		handler, ok := commands[cmd]
		if !ok {
			continue
		}
		err := handler(addr, token, msg)
		if err != nil {
			log.Println(err)
		}
	}
}

func Run() {
	config, err := parseArgs()
	if err != nil {
		log.Fatal(err)
	}

	err = startServer(config)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Server stopped")
}
