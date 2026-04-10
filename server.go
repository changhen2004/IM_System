package main

import (
	"fmt"
	"net"
	"sync"
	"strconv"
	"time"
)

type Server struct {
	IP string
	Port int
	//在线用户列表
	OnlineMap map[string]*User
	mapLock sync.RWMutex

	//广播消息channel，用于广播消息给所有在线用户
	Message chan string

}

// NewServer 创建一个新的服务器
func NewServer(ip string, port int) *Server{
	server := &Server{
		IP: ip,
		Port: port,
		OnlineMap: make(map[string]*User),
		Message: make(chan string),
	}
	return server
}

// 监听广播消息的方法
func(s *Server) ListenMessage(){
	for{
		msg := <-s.Message
		s.mapLock.Lock()
		// 遍历所有在线用户，将消息广播给所有用户
		for _, user := range s.OnlineMap {
			user.C <- msg
		}
		s.mapLock.Unlock()
	}
}

//广播消息的方法
func(s *Server) Broadcast(user *User, msg string){
	sendMsg := "[" +user.Addr + "]" + msg
	s.Message <- sendMsg
}

// handle 处理客户端连接
func (s *Server)handle(conn net.Conn){
	defer conn.Close()
	//fmt.Println("连接成功")
	user := NewUser(conn, s)
	user.OnLine()

	//监听用户是否活跃的channel
	isLive := make(chan bool)

	//接收客户端发送的消息
	go func(){
			for{
				buf := make([]byte, 4096)
			n, err :=conn.Read(buf)
			if n == 0{
				user.OffLine()
				return
			}
			if err!= nil{
				fmt.Printf("读取客户端消息失败 err:%v \n",err)
				return
			}
			// 转换为字符串
			msg := string(buf[:n])
			// 用户广播消息
			user.DoMessage(msg)

			// 标记用户活跃
			isLive <- true
		}
	}()

	//当前handle堵塞
	for{
		select{		
		case <- isLive:
			//当前用户活跃，重置超时时间
		case <-time.After(time.Second *10):
			
			user.C <- "用户" + user.Addr + "已超时下线\n"
			close(user.C)
			conn.Close()
			return
		}
	}
}

func (s *Server) Start(){
	// socket listen
	listen,err := net.Listen("tcp", s.IP+":"+strconv.Itoa(s.Port))
	if err != nil{
		fmt.Printf("监听失败 err:%v \n",err)
		return
	}
	defer listen.Close()
	// 启动监听消息的goroutine
	go s.ListenMessage()
	// accept
	for{
		conn,err := listen.Accept()
		if err != nil{
			fmt.Printf("接受连接失败 err:%v \n",err)
			return
		}
		// handle
		go s.handle(conn)
	}
}
