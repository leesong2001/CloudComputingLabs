package main

import (
	"fmt"
	"net"
)

var coordinatorIPPort = "192.168.136.1:8001"

func main() {
	conn, err := net.Dial("tcp", coordinatorIPPort)
	if err != nil {
		fmt.Println("dial error:", err)
		return
	}
	testRESPArraysStr := []string{
		"*3\r\n$3\r\nSET\r\n$5\r\nkeynm\r\n$7\r\nvaluenm\r\n",
		"*2\r\n$3\r\nGET\r\n$8\r\nnoneitem\r\n",
		"*2\r\n$3\r\nDEL\r\n$4\r\nkey1\r\n",
		"*3\r\n$3\r\nDEL\r\n$4\r\nkey1\r\n$4\r\nkey2\r\n",
		"*2\r\n$3\r\nGET\r\n$5\r\nkeynm\r\n",
	}
	for _, RESPArrays := range testRESPArraysStr {
		//发送指令消息给协同者
		conn.Write([]byte(RESPArrays))
		fmt.Println("send msg : " + RESPArrays)

		//接收服务端反馈
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("rcv err msg from server: " + RESPArrays)
			return
		}
		rcvData := string(buffer[:n])
		fmt.Println("raw rcv data: \n" + rcvData)

	}
}
