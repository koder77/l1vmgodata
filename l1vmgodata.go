// l1vmgodata.go - database in go
// work in progress, currently only echoes client message

/*
 * This file l1vmgodata.go is part of l1vmgodata.
 *
 * (c) Copyright Stefan Pietzonke (jay-t@gmx.net), 2022
 *
 * L1vm is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * L1vm is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with L1vm.  If not, see <http://www.gnu.org/licenses/>.
 */

// tutorial code for tcp/ip sockets taken from: https://www.developer.com/languages/intro-socket-programming-go/

package main

import (
	"fmt"
	"net"
	"os"
)

const (
	MACHINE_BIG_ENDIAN = "0"

	SOCKET_OPEN   = "1"
	SOCKET_CLOSE  = "0"
	SOCKET_SERVER = "0"
	SOCKET_CLIENT = "1"

	ERR_FILE_OK     = "0"
	ERR_FILE_OPEN   = "-1"
	ERR_FILE_CLOSE  = "-2"
	ERR_FILE_READ   = "-3"
	ERR_FILE_WRITE  = "-4"
	ERR_FILE_NUMBER = "-5"
	ERR_FILE_EOF    = "-6"
	ERR_FILE_FPOS   = "-7"
)

const (
	SERVER_HOST = "localhost"
	SERVER_PORT = "2000"
	SERVER_TYPE = "tcp"
)

type mem struct {
	string
	int64
	float64
}

type data struct {
	dtype int
	name  string
	size  int64
	mem   mem
}

func run_server() {
	fmt.Println("run_server...")
	server, err := net.Listen(SERVER_TYPE, SERVER_HOST+":"+SERVER_PORT)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer server.Close()
	fmt.Println("Listening on " + SERVER_HOST + ":" + SERVER_PORT)
	fmt.Println("Waiting for client...")
	for {
		connection, err := server.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		fmt.Println("client connected")
		go processClient(connection)
	}
}

func processClient(connection net.Conn) {
	buffer := make([]byte, 1024)
	mLen, err := connection.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}
	fmt.Println("Received: ", string(buffer[:mLen]))
	_, err = connection.Write([]byte("Thanks! Got your message:" + string(buffer[:mLen])))
	connection.Close()
}

func main() {
	fmt.Println("l1vmgodata start...")
	run_server()
}
