package main

import (
	"FTC/util"
	"fmt"
	"log"
	"net"
)

func main() {

	addr1 := "0.0.0.0:6001"
	addr2 := "0.0.0.0:6002"

	go initSrv(1, addr1)
	go initSrv(2, addr2)

	for {

	}
}

func initSrv(itype int, addr string) {

	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)

	if err != nil {
		log.Fatalf("net.ResovleTCPAddr fail:%s", addr)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Fatalf("listen %s fail: %s", addr, err)
	} else {

		log.Println("listening", addr)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("listener.Accept error:", err)
			continue
		}

		switch itype {
		case 1:
			go handleConnection1(conn)
		case 2:
			go handleConnection2(conn)
		}

	}
}

func handleConnection1(conn net.Conn) {
	defer conn.Close()
	var Num = 1
	for {
		Num += 1
		var buffer []byte = util.Int2Bytes_B(Num)
		conn.Write(buffer)

		fmt.Printf("write:%d .\r\n", Num)

		util.MySleep_s(2)
	}
}

func handleConnection2(conn net.Conn) {
	defer conn.Close()
	var Num = 10000
	for {
		Num += 1
		var buffer []byte = util.Int2Bytes_B(Num)
		conn.Write(buffer)

		fmt.Printf("write:%d .\r\n", Num)

		util.MySleep_s(5)
	}
}
