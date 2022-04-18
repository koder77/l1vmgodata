// l1vmgodata.go - database in go
/*
 * This file l1vmgodata.go is part of L1VMgodata.
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
	"strconv"
	"sync"
)

// socket
const (
	SERVER_HOST = "localhost"
	SERVER_TYPE = "tcp"
)

// data base commands
const (
	STORE_DATA       = "store data"
	GET_DATA_KEY     = "get key"
	GET_DATA_VALUE   = "get value"
	REMOVE_DATA      = "remove"
	CLOSE_CONNECTION = "close"
	SAVE_DATA        = "save"
	LOAD_DATA        = "load"
)

type data struct {
	used  bool
	key   string
	value string
}

var maxdata uint64 = 10000 // max data number
var server_port string = "2000"
var pdata *[]data

var dmutex sync.Mutex // data mutex

func run_server() {
	fmt.Println("run_server...")
	server, err := net.Listen(SERVER_TYPE, SERVER_HOST+":"+server_port)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer server.Close()
	fmt.Println("Listening on " + SERVER_HOST + ":" + server_port)
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
	var run_loop bool = true
	buffer := make([]byte, 4096)
	var key string = ""
	var value string = ""

	var match bool

	for run_loop {
		mLen, err := connection.Read(buffer)
		if err != nil {
			fmt.Println("processClient: Error reading:", err.Error())
		}
		// fmt.Println("Received: '", string(buffer[:mLen]), "'")
		// fmt.Println("length: ", mLen)

		// store data
		regexp_store := regexp.MustCompile(STORE_DATA)
		match = regexp_store.Match([]byte(buffer[:mLen]))
		if match {
			// store key/value pair
			// try to store data
			key, value = split_data(string(buffer[:mLen]))
			if key != "" {
				if store_data(key, value) == 0 {
					_, err = connection.Write([]byte("OK\n"))
					if err != nil {
						fmt.Println("processClient: Error writing:", err.Error())
					}
				} else {
					_, err = connection.Write([]byte("ERROR\n"))
					if err != nil {
						fmt.Println("processClient: Error writing:", err.Error())
					}
				}
			} else {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("processClient: Error writing:", err.Error())
				}
			}
		}

		// get data key
		regexp_key := regexp.MustCompile(GET_DATA_KEY)
		match = regexp_key.Match([]byte(buffer[:mLen]))
		if match {
			// try to find matching key
			key = split_key(string(buffer[:mLen]))
			if key != "" {
				value = get_data_key(key)
				if value != "" {
					_, err = connection.Write([]byte(value))
					if err != nil {
						fmt.Println("processClient: Error writing:", err.Error())
					}
					_, err = connection.Write([]byte("\n"))
					if err != nil {
						fmt.Println("processClient: Error writing:", err.Error())
					}
				} else {
					_, err = connection.Write([]byte("ERROR\n"))
					if err != nil {
						fmt.Println("processClient: Error writing:", err.Error())
					}
				}
			} else {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("processClient: Error writing:", err.Error())
				}
			}
		}

		// get data value
		regexp_value := regexp.MustCompile(GET_DATA_VALUE)
		match = regexp_value.Match([]byte(buffer[:mLen]))
		if match {
			// try to find matching value
			value = split_value(string(buffer[:mLen]))
			if value != "" {
				key = get_data_value(value)
				if key != "" {
					_, err = connection.Write([]byte(key))
					if err != nil {
						fmt.Println("processClient: Error writing:", err.Error())
					}
					_, err = connection.Write([]byte("\n"))
					if err != nil {
						fmt.Println("processClient: Error writing:", err.Error())
					}
				} else {
					_, err = connection.Write([]byte("ERROR\n"))
					if err != nil {
						fmt.Println("processClient: Error writing:", err.Error())
					}
				}
			} else {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("processClient: Error writing:", err.Error())
				}
			}
		}

		// remove data, send value
		regexp_rdata := regexp.MustCompile(REMOVE_DATA)
		match = regexp_rdata.Match([]byte(buffer[:mLen]))
		if match {
			// try to find matching key
			key = split_key(string(buffer[:mLen]))
			if key != "" {
				value = remove_data(key)
				if value != "" {
					_, err = connection.Write([]byte(value))
					if err != nil {
						fmt.Println("processClient: Error writing:", err.Error())
					}
					_, err = connection.Write([]byte("\n"))
					if err != nil {
						fmt.Println("processClient: Error writing:", err.Error())
					}
				} else {
					_, err = connection.Write([]byte("ERROR\n"))
					if err != nil {
						fmt.Println("processClient: Error writing:", err.Error())
					}
				}
			} else {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("processClient: Error writing:", err.Error())
				}
			}
		}

		// check close
		regexp_close := regexp.MustCompile(CLOSE_CONNECTION)
		match = regexp_close.Match([]byte(buffer[:mLen]))
		if match {
			_, err = connection.Write([]byte("OK\n"))
			run_loop = false
		}

		// check save
		regexp_save := regexp.MustCompile(SAVE_DATA)
		match = regexp_save.Match([]byte(buffer[:mLen]))
		if match {
			// try to find matching path name
			value = split_value(string(buffer[:mLen]))
			if value != "" {
				if save_data(value) != 0 {
					_, err = connection.Write([]byte("ERROR\n"))
					if err != nil {
						fmt.Println("processClient: Error writing:", err.Error())
					}
				} else {
					_, err = connection.Write([]byte("OK\n"))
					if err != nil {
						fmt.Println("processClient: Error writing:", err.Error())
					}
				}
			} else {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("processClient: Error writing:", err.Error())
				}
			}
		}

		// check load
		regexp_load := regexp.MustCompile(LOAD_DATA)
		match = regexp_load.Match([]byte(buffer[:mLen]))
		if match {
			// try to find matching path name
			value = split_value(string(buffer[:mLen]))
			if value != "" {
				if load_data(value) != 0 {
					_, err = connection.Write([]byte("ERROR\n"))
					if err != nil {
						fmt.Println("processClient: Error writing:", err.Error())
					}
				} else {
					_, err = connection.Write([]byte("OK\n"))
					if err != nil {
						fmt.Println("processClient: Error writing:", err.Error())
					}
				}
			} else {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("processClient: Error writing:", err.Error())
				}
			}
		}

	}

	connection.Close()
}

func main() {
	fmt.Println("l1vmgodata <port> <number of data entries>")
	fmt.Println("l1vmgodata start...")

	if len(os.Args) == 2 || len(os.Args) == 3 {
		// get port from command line
		server_port = os.Args[1]
	}
	if len(os.Args) == 3 {
		// get maxdata from command line
		user_maxdata, err := strconv.ParseInt(os.Args[2], 10, 64)
		if err != nil {
			panic(err)
		}
		maxdata = uint64(user_maxdata)
	}
	fmt.Println("allocating ", maxdata, " space for data")
	servdata := make([]data, maxdata) // make serverdata spice
	pdata = &servdata

	init_data()
	run_server()
}
