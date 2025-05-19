package exchange

import (
	"bufio"
	"log"
	"net"
)

func ListenExchange(addr string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		log.Println(line)
	}
	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
}
