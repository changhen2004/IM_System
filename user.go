package main

import "net"


type User struct{
	Name string
	Addr string
	Conn net.Conn	
	C chan string
}
//创建一个用户的API
func NewUser(conn net.Conn) *User{
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name: userAddr,
		Addr: userAddr,
		Conn: conn,
		C: make(chan string),
	}
	// 启动监听消息的goroutine
	go user.ListenMessage()
	return user
}
//监听当前User channel的方法，一旦有消息，就接收并发送给客户端
func(u *User) ListenMessage(){
	for{
		msg := <-u.C
		u.Conn.Write([]byte(msg))
	}
}
