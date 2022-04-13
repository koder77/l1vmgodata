// l1vmgodata.go - database in go
// work in progress, currently only echoes client message

/*
 * This file l1vmgodata.go is part of l1vmgodata.
 *
 * (c) Copyright Stefan Pietzonke (jay-t@gmx.net), 2022
 *
 * L1VMgodata is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * L1VMgodata is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with L1VMgodata.  If not, see <http://www.gnu.org/licenses/>.
 */

// tutorial code for tcp/ip sockets taken from: https://www.developer.com/languages/intro-socket-programming-go/

package main

import (
	"fmt"
	"net"
	"os"
	"regexp"
)

// misc defs
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

// socket
const (
	SERVER_HOST = "localhost"
	SERVER_PORT = "2000"
	SERVER_TYPE = "tcp"
)

type data struct {
	used  bool
	key   string
	value string
}

var maxdata int64 = 10000 // max data number
var pdata *[]data

func init_data() {
	var i int64
	for i = 0; i < maxdata; i++ {
		(*pdata)[i].used = false
		(*pdata)[i].key = ""
		(*pdata)[i].value = ""
	}
}

func get_free_space() int64 {
	var i int64
	for i = 0; i < maxdata; i++ {
		if !(*pdata)[i].used {
			return i
		}
	}
	// no free space found, return -1
	return -1
}

func store_data(key string, value string) int {
	var i int64 = 0

	i = get_free_space()
	if i < 0 {
		// error: no fre space
		return 1
	}

	// store data at index i
	(*pdata)[i].used = true
	(*pdata)[i].key = key
	(*pdata)[i].value = value

	return 0
}

func get_data_key(key string) string {
	var i int64
	var match bool

	regexp := regexp.MustCompile(key)

	for i = 0; i < maxdata; i++ {
		if !(*pdata)[i].used {
			match = regexp.Match([]byte((*pdata)[i].key))
			if match {
				return (*pdata)[i].value
			}
		}
	}
	// no matching key found, return empty string
	return ""
}

func get_data_value(value string) string {
	var i int64
	var match bool

	regexp := regexp.MustCompile(value)

	for i = 0; i < maxdata; i++ {
		if !(*pdata)[i].used {
			match = regexp.Match([]byte((*pdata)[i].value))
			if match {
				return (*pdata)[i].key
			}
		}
	}
	// no matching value found, return empty string
	return ""
}

func remove_data(key string) string {
	var i int64
	var match bool
	var value string

	regexp := regexp.MustCompile(key)

	for i = 0; i < maxdata; i++ {
		if !(*pdata)[i].used {
			match = regexp.Match([]byte((*pdata)[i].key))
			if match {
				value = (*pdata)[i].value
				(*pdata)[i].used = false
				(*pdata)[i].key = ""
				(*pdata)[i].value = ""
				return value
			}
		}
	}
	// no matching key found, return empty string
	return ""
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
	servdata := make([]data, maxdata) // make serverdata spice
	pdata = &servdata

	init_data()
	run_server()
}
