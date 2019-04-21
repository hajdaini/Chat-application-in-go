package server

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

/*
Server : server class
@string IP : ip of the server
@string PORT : port of the server
@string Username : username of the client
@struct net.Conn Conn : connection of the client
@bool isConnected : is client still connected or not
*/
type Server struct {
	IP       string
	PORT     string
	Listener net.Listener
	clients  map[net.Conn]string // @key : client and @value : username
}

/*
New  : create new Server
@string IP : ip of the server
@string PORT : port of the server
@return struct Server : instance of the server
*/
func New(IP string, PORT string) Server {
	var server Server
	server.IP, server.PORT = IP, PORT
	server.clients = make(map[net.Conn]string) // initialize the Map of clients
	return server
}

/*
startConnection : starting the server listener
*/
func (server *Server) startConnection() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", server.IP, server.PORT))
	server.check(err)
	fmt.Println("Server is running ...")
	server.Listener = listener
}

/*
check  : exit and close client connection if there is an error
@error err: the error the check
*/
func (server *Server) check(err error) {
	if err != nil {
		if server.Listener != nil {
			server.Listener.Close()
		}
		fmt.Println("Server is shutdown")
		fmt.Println(err)
		os.Exit(2)
	}
}

/*
addLog : add a line in the log file
@string line : the line to add to the log file
*/
func (server *Server) addLog(line string) {
	file, err := os.OpenFile("logs.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	defer file.Close()
	if err != nil {
		fmt.Println("ERROR :", err)
	}

	line = datetimeLine(line)
	_, err = file.WriteString(line)
	if err != nil {
		fmt.Println("ERROR :", err)
	}
	fmt.Print(line)
}

/*
formatedLog : add current datetime to the line
@return string : the new line with the datetime
*/
func datetimeLine(text string) string {
	datetimeNow := time.Now().Format("02/01/2006 15:04:05") // Format MDD-MM-YYYY hh:mm:ss
	return fmt.Sprintf("[%s] %s", datetimeNow, text)
}

/*
addClient : add client to the server Map()
@struct net.Conn client : the client to add
*/
func (server *Server) addClient(client net.Conn) {
	if !server.checkUsernameClient(client) {
		return // exit the goroutine because the client interrupt the username input
	}

	message := fmt.Sprintf("[INFO] %s join the server\n", server.clients[client])

	if len(server.clients) > 1 {
		client.Write([]byte("List of usernames in the server:\n"))
		for c := range server.clients {
			client.Write([]byte(fmt.Sprintf("-> %s\n", server.clients[c])))
		}
	}

	client.Write([]byte("You can start the discussion with guests ...\n\n"))
	server.sendToAll(client, message, true)
	server.addLog(fmt.Sprintf("%s connected from %s\n", server.clients[client], client.RemoteAddr()))
	server.receive(client)
}

/*
addClient : check if the client username already exist in the clients Map()
@struct net.Conn client : the client to to check
@return bool : client username accepted or not
*/
func (server *Server) checkUsernameClient(client net.Conn) bool {
	var (
		username string
		err      error
	)

	for {
		username, err = server.catchClientUsername(client)
		if err != nil {
			return false
		}

		if server.isUsernameExists(username, client) {
			client.Write([]byte("badUsername"))
		} else {
			client.Write([]byte("goodUsername"))
			break
		}
	}

	server.clients[client] = username //add client the clients Map()
	return true
}

/*
isUsernameExists : check if the username of the client already exists in the Map()
@string username : username to check
@string username : client to check
*/
func (server *Server) isUsernameExists(username string, client net.Conn) bool {
	findUsername := false
	for c := range server.clients {
		if server.clients[c] == username && client != c { // we ignore the client itself comparaison because he is already in the clients Map()
			fmt.Println(datetimeLine(fmt.Sprintf("The username %s already exist !", username)))
			findUsername = true
			break
		}
	}
	return findUsername
}

/*
catchClientUsername : catch the username send by the client
@struct net.Conn client : client who sent the username
@return string : username of the client
@return error : error from user input
*/
func (server *Server) catchClientUsername(client net.Conn) (string, error) {
	var err error

	usernameBuffer := make([]byte, 4096)
	length, err := client.Read(usernameBuffer)

	username := string(usernameBuffer[:length]) // remove all unused bytes in the buffer and convert it to string

	if err != nil { // if the client interrupt the username input
		server.addLog(fmt.Sprintf("Client from %s interrupt the username input\n", client.RemoteAddr()))
		username = "error"
		err = errors.New("client interrupt input")
	}

	return strings.TrimSuffix(username, "\n"), err
}

/*
sendToAll : send message to all clients
@struct net.Conn client : client to ignore if @ignoreItself is true
@string message : message that will be send
@bool ignoreItself : ignore the client variable
*/
func (server *Server) sendToAll(client net.Conn, message string, ignoreItself bool) {
	for c := range server.clients {
		if ignoreItself {
			if c != client { // we do not send the message back to the same sender
				c.Write([]byte(message))
			}
		} else {
			c.Write([]byte(message))
		}
	}
}

/*
receive : receive message of of the client
@struct net.Conn client : client that sent the message
*/
func (server *Server) receive(client net.Conn) {
	buf := bufio.NewReader(client)
	for {
		if message, err := buf.ReadString('\n'); err != nil {
			server.removeClient(client)
			break
		} else {
			message = fmt.Sprintf("%s : %s", server.clients[client], message)
			server.sendToAll(client, message, true)
		}
	}
}

/*
removeClient : remove client from the clients Map()
@struct net.Conn client : client to delete
*/
func (server *Server) removeClient(client net.Conn) {
	message := fmt.Sprintf("[INFO] %s is now disconnected\n", server.clients[client])
	server.sendToAll(client, message, true)
	server.addLog(fmt.Sprintf("%s is disconnected [total client %d]\n", server.clients[client], len(server.clients)-1))
	delete(server.clients, client)
}

/*
Run : start server connection and wait for clients
*/
func (server *Server) Run() {
	server.startConnection()
	for {
		client, err := server.Listener.Accept()
		server.check(err)
		go server.addClient(client)
	}
}
