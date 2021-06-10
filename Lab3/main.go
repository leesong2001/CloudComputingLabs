package main

import (
	"fmt"
	"net"
)

func main() {
	finCnt2PC = 0
	database = make(map[string]string)
	readConfig()
	var localIPPort string
	if mode == "coordinator" {
		localIPPort = coordinatorIPPort
	} else if mode == "participant" {
		localIPPort = participantIPPortArr[0]
	}

	//绑定IP:Port
	l, err := net.Listen("tcp", localIPPort)
	if err != nil {
		fmt.Println(mode, " listen error:", err)
		return
	}
	defer l.Close()
	if mode == "coordinator" {
		start_coordinator(l)
	} else if mode == "participant" {
		//监听端口，accept客户端的连接请求
		for {
			conn, err := l.Accept()
			if err != nil {
				fmt.Println("coordinatorIPPort accept error:", err)
				return
			}

			go coordinatorHandle(conn)
			//处理coordinator的请求
			//现在还需要为其他参与者进行恢复 添加go 关键字
		}
	}
}
