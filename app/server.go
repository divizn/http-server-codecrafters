package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()

	for {
		connection, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}
		go handleConnection(connection)
	}
}

func handleConnection(connection net.Conn) {
	defer connection.Close()

	requestBuffer := make([]byte, 1024)

	n, err := connection.Read(requestBuffer)
	if err != nil {
		fmt.Println("Failed to read the request:", err)
		return
	}
	fmt.Printf("Request: %s\n", requestBuffer[:n])

	request := string(requestBuffer[:n])
	path := strings.Split(request, " ")[1]

	if path == "/" {
		connection.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))

	} else if isPath("GET /echo", request) {
		message := strings.Split(path, "/")[2]
		connection.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(message), message)))

	} else if isPath("GET /user-agent", request) {
		agent := strings.Split(strings.Split(request, "User-Agent: ")[1], "\r\n")[0]
		print("agent = ", agent, "\nlength = ", len(agent))
		connection.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(agent), agent)))

	} else if isPath("GET /files", request) {
		dir := os.Args[2]
		filename := strings.Split(path, "/")[2]
		println(filename, "- filename")
		file, err := os.ReadFile(dir + "/" + filename)
		if err != nil {
			println("Error: ", err.Error())
			connection.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
			connection.Close()
			return
		}
		fileLength := len(file)
		fileContents := string(file)
		connection.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", fileLength, fileContents)))

	} else if isPath("POST /files", request) {
		filename := strings.Split(path, "/")[2]
		body := strings.Split(request, "\r\n")
		dir := os.Args[2]
		file, err := os.Create(filepath.Join(dir, filepath.Base(filename)))
		if err != nil {
			connection.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
		}
		file.WriteString(body[len(body)-1])
		file.Close()

		connection.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))

	} else {
		connection.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}

func isPath(comparison string, request string) bool {
	if strings.Contains(request, comparison) {
		return true
	} else {
		return false
	}
}
