package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

var wg sync.WaitGroup

const maxLengthUsername int = 20

/*
Client : client class
@string IP : ip of the server
@string PORT : port of the server
@string Username : username of the client
@struct net.Conn Conn : connection of the client
@bool isConnected : is client still connected or not
*/
type Client struct {
	IP          string
	PORT        string
	Username    string
	Conn        net.Conn
	IsConnected bool
}

/*
New  : "Constructor" of the "class" Client
@string IP : ip of the server
@string PORT : port of the server
@return struct Client : instance of the client
*/
func New(IP string, PORT string) Client {
	var client Client
	client.IP, client.PORT = IP, PORT
	return client
}

/*
usernameHandle  : handle client username input
*/
func (client *Client) usernameHandle() {

	for {
		username := client.getUsernameInput()
		lengthUsername := len(strings.TrimSuffix(username, "\n"))

		if lengthUsername > maxLengthUsername || lengthUsername == 0 {
			fmt.Println("[ERROR] Your username must not be empty or exceed", maxLengthUsername, "characters")
			continue
		}

		client.Conn.Write([]byte(username))  // send the username to the server that will check it
		serverResponse, err := client.read() // read the check response of the server
		client.check(err)

		if client.isUsernameGood(serverResponse) {
			client.Username = strings.TrimSuffix(username, "\n")
			fmt.Print("[SUCCESS] You are successfully connected!\n")
			break
		}
	}
}

/*
getUsernameInput : run username input
@return string : username input
*/
func (client *Client) getUsernameInput() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your username : ")
	username, err := reader.ReadString('\n')

	client.check(err)
	return username
}

/*
isUsernameGood : Check if username is good based on the server response
@string username : response of the server
@return bool : good or bad username
*/
func (client *Client) isUsernameGood(response string) bool {
	if response == "badUsername" {
		fmt.Println("[ERROR] Your username already exists in the server, please enter another username")
		return false
	} else {
		fmt.Println("[SUCCESS] Your username is accepted by the server")
		return true
	}
}

/*
connect  : connect to the server
*/
func (client *Client) connect() {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", client.IP, client.PORT))
	client.check(err)
	fmt.Println("Connecting to", conn.RemoteAddr(), "SERVER ...")
	client.IsConnected = true
	client.Conn = conn
}

/*
check  : exit and close client connection if there is an error
@error err: the error the check
*/
func (client *Client) check(err error) {
	if err != nil {
		client.IsConnected = false
		if client.Conn != nil {
			client.Conn.Close()
		}
		fmt.Println(err)
		fmt.Println("You are now disconnected !")
		os.Exit(2)
	}
}

/*
send : get user input and send it to the server
*/
func (client *Client) send() {
	defer wg.Done()
	for {
		reader := bufio.NewReader(os.Stdin)
		message, err := reader.ReadString('\n')
		if !client.IsConnected {
			break
		}
		client.check(err)
		client.Conn.Write([]byte(message))
	}
}

/*
read : catch message send by the server and return it
@return string : message of the server
*/
func (client *Client) read() (string, error) {
	messageBuffer := make([]byte, 4096)
	length, err := client.Conn.Read(messageBuffer)
	if err != nil {
		fmt.Println("[INFO] Server is down, click Enter to close the session")
		client.IsConnected = false
	}
	message := string(messageBuffer[:length])
	return message, err
}

/*
receive : catch all messages send by the server
*/
func (client *Client) receive() {
	defer wg.Done()
	for {
		message, err := client.read()
		if !client.IsConnected {
			break
		}
		if err != nil {
			fmt.Println("[INFO] Server is down, click Enter to close the session")
			client.IsConnected = false
			break
		}
		fmt.Print(message)
	}
}

/*
Run : start connection to the server and begin the conversation with others clients of the server
*/
func (client *Client) Run() {
	client.connect()
	client.usernameHandle()
	wg.Add(2)
	go client.send()
	go client.receive()
	wg.Wait()
	fmt.Println("You are now disconnected !")
}
