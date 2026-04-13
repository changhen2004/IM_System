package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"io"
)

type Client struct{
	ServerIp string
	ServerPort int
	Name string
	conn net.Conn
	flag int
}

func NewClient(ServerIp string, ServerPort int) *Client{
	//创建客户端链接
	client:= &Client{
		ServerIp: ServerIp,
		ServerPort: ServerPort,
		flag: 999, 
	}
	// 连接服务器
	conn, err := net.Dial("tcp", ServerIp + ":" + strconv.Itoa(ServerPort))
	if err != nil {
		panic(err)
	}
	client.conn = conn
	
	return client
}

//处理server回应的消息，直接显示到标准输出
func (client *Client) DealResponse(){
	//一旦client.conn有数据可读，就直接显示到标准输出，永久阻塞监听
	io.Copy(os.Stdout, client.conn)
}

func (client *Client) menu() bool{
	var flag int
	fmt.Println(">>>>>>客户端菜单：")
	fmt.Println("1. 公聊模式")
	fmt.Println("2. 私聊模式")
	fmt.Println("3. 更新用户名")
	fmt.Println("0. 退出")
	
	fmt.Scanln(&flag)
	if flag >=0 && flag <=3{
		client.flag = flag
		return true
	}else{
		fmt.Println(">>>>>>请输入合法范围内的数字...")
		return false
	}
}

//查询在线用户
func (client *Client) SelectUser(){
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil{
		fmt.Println(">>>>>>查询在线用户失败...")
		return
	}
}

//私聊模式
func (client *Client) PrivateChat(){
	var targetName string
	var chatMsg string
	client.SelectUser()
	fmt.Println(">>>>>>请输入聊天对象[用户名],exit退出")
	fmt.Scanln(&targetName)

	for targetName != "exit"{
		fmt.Println(">>>>>>请输入聊天内容[exit退出]：")
		fmt.Scanln(&chatMsg)
		for chatMsg != "exit"{
			if len(chatMsg) != 0{
				sendMsg := "to|" + targetName + "|" + chatMsg + "\n\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil{
					fmt.Println(">>>>>>发送消息失败...")
					break
				}
			}
			chatMsg = ""
			fmt.Println(">>>>>>请输入聊天内容[exit退出]：")
			fmt.Scanln(&chatMsg)
		}
		client.SelectUser()
		fmt.Println(">>>>>>请输入聊天对象[用户名],exit退出")
		fmt.Scanln(&targetName)
	}
}

//公聊模式
func (client *Client) PublicChat(){
	fmt.Println(">>>>>>请输入聊天内容，输入exit退出")
	var chatMsg string
	fmt.Scanln(&chatMsg)
	for chatMsg != "exit"{
		if len(chatMsg) != 0{
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil{
				fmt.Println(">>>>>>发送消息失败...")
				break
			}
		}
		chatMsg = ""
		fmt.Println(">>>>>>请输入聊天内容，输入exit退出")
		fmt.Scanln(&chatMsg)
	}
}

func (client *Client) UpdateName() bool{
	fmt.Println(">>>>>>请输入用户名：")
	fmt.Scanln(&client.Name)
	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil{
		fmt.Println(">>>>>>更新用户名失败...")
		return false
	}
	
	return true
}

func (client *Client) Run(){
	for client.flag != 0{
		for client.menu() != true{
		}

		switch client.flag{
		case 1:
			// 公聊模式
			fmt.Println(">>>>>>公聊模式")
			client.PublicChat()
			
		case 2:
			// 私聊模式
			fmt.Println(">>>>>>私聊模式")
			client.PrivateChat()
		case 3:
			// 更新用户名
			client.UpdateName()
		}
	}
}
		

var serverIp string
var serverPort int

// ./client -ip 127.0.0.1 -port 8080
func init(){
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP（默认127.0.0.1）")
	flag.IntVar(&serverPort, "port", 8080, "设置服务器端口（默认8080）")
}

func main(){
	//命令行解析
	flag.Parse()
	
	client := NewClient(serverIp, serverPort)
	if client.conn == nil{
		fmt.Println(">>>>>>连接服务器失败...")
		return
	}
	// 启动处理server回应的消息的goroutine
	go client.DealResponse()
	fmt.Println(">>>>>>服务器连接成功...")

	//启动客户端的业务
	client.Run()
}