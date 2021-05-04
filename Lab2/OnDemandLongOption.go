package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
)

const debug_mode = false
const rootPath = "../Lab2"
const resp404 = "HTTP/1.0 404 Not Found\r\n"
const resp501 = "HTTP/1.0 501 Not Implemented\r\n"
const resp200 = "HTTP/1.0 200 OK\r\n"
const contentHtml = "Content-type: text/html\r\n"
const contentLen = "Content-length: "

var ip string
var port string
var ThreadNum int

func readFile(filePath string) ([]byte, int) {
	f, err := os.OpenFile(filePath, os.O_RDONLY, 0600)
	defer f.Close()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		contentByte, _ := ioutil.ReadAll(f)
		return contentByte, len(contentByte)
	}
	return nil, 0
}
func response(statusCode int, conn net.Conn, req string, filePath string, attachment string) {
	/* statusCode状态码
	 * filePath 文件路径
	 * conn 客户端连接
	 * req 请求方式
	 * attachment post请求的name-id pair
	 */

	if statusCode == 501 {
		conn.Write([]byte(resp501))
		conn.Write([]byte(contentLen + strconv.Itoa(0) + "\r\n")) //首部字段
		conn.Write([]byte("\r\n"))                                //空行
	} else if statusCode == 404 {
		conn.Write([]byte(resp404))
		conn.Write([]byte(contentLen + strconv.Itoa(0) + "\r\n")) //首部字段
		conn.Write([]byte("\r\n"))                                //空行
	} else if statusCode == 200 {
		if req == "GET" {
			//conn.Write(200 OK 以及 html文件全部内容)
			data, datasize := readFile(filePath)
			conn.Write([]byte(resp200))
			conn.Write([]byte(contentHtml))                                  //首部字段
			conn.Write([]byte(contentLen + strconv.Itoa(datasize) + "\r\n")) //首部字段

			conn.Write([]byte("\r\n")) //空行
			conn.Write(data)

		} else {
			//返回200 OK并回显 "Name"-"ID" pairs
			//conn.Write(200 OK + attachment )
			conn.Write([]byte(resp200))
			conn.Write([]byte(contentLen + strconv.Itoa(len([]byte(attachment))) + "\r\n")) //首部字段
			conn.Write([]byte("\r\n" + attachment))
		}
	}
}
func fileIsExist(filePath string) (bool, error) {
	/*golang判断文件或文件夹是否存在的方法为使用os.Stat()函数返回的错误值进行判断:

	如果返回的错误为nil,说明文件或文件夹存在
	如果返回的错误类型使用os.IsNotExist()判断为true,说明文件或文件夹不存在
	如果返回的错误为其它类型,则不确定是否在存在*/
	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
func parseDataField(data string) (string, string) {
	var Name string
	var ID string
	//data:   'Name=XXXXX&ID=XXX'

	var splitIdx int
	for i := 0; i < len(data); i++ {
		if data[i] == '&' {
			splitIdx = i
			break
		}
	}

	if data[0:5] == "Name=" {
		Name = data[5:splitIdx]
		if data[splitIdx+1:splitIdx+4] == "ID=" {
			ID = data[splitIdx+4 : len(data)]
		}
	} else if data[0:3] == "ID=" {
		ID = data[3:splitIdx]
		if data[splitIdx+1:splitIdx+6] == "Name=" {
			Name = data[splitIdx+6 : len(data)]
		}
	} else {
		Name = ""
		ID = ""
	}

	return Name, ID
}
func accept_request_thread(conn net.Conn) {
	var method_bd strings.Builder
	var url_bd strings.Builder
	var data_bd strings.Builder
	var i int
	var filePath string

	// 创建一个新切片， 用作保存数据的缓冲区
	buf := make([]byte, 1024)
	n, err := conn.Read(buf) // 从conn中读取客户端发送的数据内容
	if err != nil {
		fmt.Printf("客户端退出 err=%v\n", err)
		return
	}

	i = 0
	for i < n && buf[i] != ' ' {
		//根据空格切分请求方法
		method_bd.WriteByte(buf[i])
		i++
	}
	for i < n && buf[i] == ' ' {
		i++
	} //游标移动到url field

	for i < n && buf[i] != ' ' {
		//根据空格切分请求的url
		url_bd.WriteByte(buf[i])
		i++
	}
	for i < n && buf[i] == ' ' {
		i++
	} //游标移动到http version field

	var dataFieldStart int
	for ; i < n; i++ { //连续的 \r\n\r\n 确定data field 起始
		if buf[i] == '\r' && buf[i+1] == '\n' && buf[i+2] == '\r' && buf[i+3] == '\n' {
			dataFieldStart = i + 4 //指向data field起始字节
			break
		}
	}
	i = dataFieldStart
	for i < n {
		//根据空格切分请求方法
		data_bd.WriteByte(buf[i])
		i++
	}

	method := method_bd.String()
	url := url_bd.String()
	data := data_bd.String()

	var j = 0
	for ; j < len(url); j++ {
		if url[j] == '/' {
			if j+1 < len(url) {
				if url[j+1] != '/' {
					//请求路径 /path...
					break
				}
			} else { //  url 以'/'结尾
				break
			}
		}
	}

	if j == len(url) {
		filePath = string('/')
	} else {
		filePath = url[j:len(url)] //   '/path' or '/'
	}

	filePath = rootPath + filePath

	if debug_mode {
		fmt.Println(method)
		fmt.Println(url)
		fmt.Println(filePath)
	}

	if method == "GET" {
		/*需求4-2：如果请求的url对应于目录下已经存在的html文件，则返回200 OK以及文件的全部内容。特别的是，要求能够处理带有子目录的url:*/
		if debug_mode {
			fmt.Println("len filePath:", len(filePath))
			fmt.Println("filePath[-5:]:", filePath[len(filePath)-5:len(filePath)])
		}
		if len(filePath) > 5 && filePath[len(filePath)-5:len(filePath)] == ".html" {
			fileExist, _ := fileIsExist(filePath)
			if fileExist {
				//返回200 OK 以及文件的全部内容
				//response(int statusCode,Conn conn,String req,String filePath,String attachment)
				response(200, conn, method, filePath, "")
			} else {
				//返回 404 Not Found response
				response(404, conn, "", "", "")
			}
		} else {
			//请求的是目录
			fileExist, _ := fileIsExist(filePath + "/index.html")
			if fileExist {
				//返回200 OK 以及index.html文件的全部内容
				response(200, conn, method, filePath+"/index.html", "")
			} else {
				//返回 404 Not Found response
				response(404, conn, "", "", "")
			}
		}
	} else if method == "POST" {
		if url == "/Post_show" {
			//解析data field
			if debug_mode {
				fmt.Println("data field: ", data)
			}
			Name, ID := parseDataField(data)
			attachment := "Your Name: " + Name + "\nYour ID: " + ID
			response(200, conn, method, "", attachment)

		} else {
			//返回404 Not Found response message.
			response(404, conn, "", "", "")
		}
	} else {
		//既不是GET也不是POST，返回501 Not Implemented error message
		response(501, conn, "", "", "")
	}

}

func main() {
	ipInput := flag.String("ip", "127.0.0.1", "What is your ip_address?")
	portInput := flag.String("port", "8888", "What is the port?")
	numberThreadInput := flag.Int("number-thread", 1, "How much is the thread number?")

	flag.Parse() //解析输入的参数

	ip = *ipInput
	port = *portInput
	ThreadNum = *numberThreadInput

	listen, err := net.Listen("tcp", ":"+port) // 创建用于监听的 socket
	if err != nil {
		fmt.Println("listen err=", err)
		return
	}
	fmt.Println("监听套接字，创建成功, 服务器开始监听。。。")
	defer listen.Close() // 服务器结束前关闭 listener

	// 循环等待客户端来链接
	for {
		if debug_mode {
			fmt.Println("阻塞等待客户端来链接...")
		}
		conn, err := listen.Accept() // 创建用户数据通信的socket
		if debug_mode {
			if err != nil {
				fmt.Println("Accept() err=", err)
			} else {
				fmt.Println("通信套接字，创建成功。。。")
			}
		}
		// 这里准备起一个协程，为客户端服务
		go accept_request_thread(conn)
	}
}
