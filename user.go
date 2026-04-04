package main

import "net"


type User struct{
	Name string
	Addr string
	Conn net.Conn	
	C chan string
	server *Server
}
//创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User{
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name: userAddr,
		Addr: userAddr,
		Conn: conn,
		C: make(chan string),
		server: server,
	}
	// 启动监听消息的goroutine
	go user.ListenMessage()
	return user
}
func(u *User) OnLine(){
	u.server.mapLock.Lock()
	u.server.OnlineMap[u.Addr] = u	
	u.server.mapLock.Unlock()
	u.server.Broadcast(u, "用户" + u.Addr + "已上线")
}
func(u *User) OffLine(){
	u.server.mapLock.Lock()
	delete(u.server.OnlineMap, u.Addr)
	u.server.mapLock.Unlock()
	u.server.Broadcast(u, "用户" + u.Addr + "已下线")
}
func(u *User) DoMessage(msg string){
	u.server.Broadcast(u, msg)
}


//监听当前User channel的方法，一旦有消息，就接收并发送给客户端
func(u *User) ListenMessage(){
	for{
		msg := <-u.C
		u.Conn.Write([]byte(msg))
	}
}
