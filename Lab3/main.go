package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
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
		//直接连到服务器
		//监听端口，accept客户端的连接请求
		time.Sleep(time.Second * 1)
		for i, participantIPPortTmp := range participantIPPortArr {
			cn, err := net.Dial("tcp", participantIPPortTmp)
			if err != nil {
				fmt.Printf("link to participant%s failed: %s\n", participantIPPortTmp, err.Error())
				connParticipant[i] = nil
				continue
			}
			connParticipant[i] = cn
			defer connParticipant[i].Close()
		}
		go heartBeatsCheck(cmdlist, status) //开启心跳
		for {
			conn, err := l.Accept()
			if err != nil {
				fmt.Println("coordinatorIPPort accept error:", err)
				return
			}
			fmt.Println("client dail: ", conn)

			clientHandle(conn)
		}
	} else if mode == "participant" {
		//监听端口，accept客户端的连接请求
		for {
			conn, err := l.Accept()
			if err != nil {
				fmt.Println("coordinatorIPPort accept error:", err)
				return
			}
			coordinatorHandle(conn) //处理coordinator的请求
			if conn != nil {
				conn.Close()
			}
		}
	}
}
