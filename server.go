package main

import (
	"fmt"
	"net"
	"sync"
	"strconv"
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
	user := NewUser(conn)
	//用户上线，将用户添加到在线用户列表中
	s.mapLock.Lock()
	s.OnlineMap[user.Addr] = user
	s.mapLock.Unlock()
	//用户上线，广播用户上线消息
	s.Broadcast(user, "上线了")

	//接收客户端发送的消息
	for{
		buf := make([]byte, 4096)
		n, err :=conn.Read(buf)
		if err!= nil{
			fmt.Printf("读取客户端消息失败 err:%v \n",err)
			continue
		}
		// 转换为字符串
		msg := string(buf[:n])
		// 广播消息
		s.Broadcast(user, msg)
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
			continue
		}
		// handle
		go s.handle(conn)
	}
}
