package main

import (
	"fmt"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:8888")
	for {
		//time.Sleep(time.Duration(10)*time.Second)
		var ins string
		fmt.Scan(&ins)
		if err != nil {
			fmt.Println(err)
		}
		conn.Write([]byte(ins))
	}
}

/*http://127.0.0.1:8888/api/camera/get_ptz?camera_id=1324566666789876543*/
