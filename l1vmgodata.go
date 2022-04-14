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
	"bufio"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
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

var maxdata int64 = 10000 // max data number
var server_port string = "2000"
var pdata *[]data

var dmutex sync.Mutex // data mutex

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
	dmutex.Lock()
	for i = 0; i < maxdata; i++ {
		if !(*pdata)[i].used {
			dmutex.Unlock()
			return i
		}
	}
	dmutex.Unlock()
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
	dmutex.Lock()
	(*pdata)[i].used = true
	(*pdata)[i].key = key
	(*pdata)[i].value = value
	dmutex.Unlock()
	return 0
}

func get_data_key(key string) string {
	var i int64
	var match bool

	skey := strings.Trim(key, "\n")
	regexp := regexp.MustCompile(skey)

	dmutex.Lock()
	for i = 0; i < maxdata; i++ {
		if (*pdata)[i].used {
			match = regexp.Match([]byte((*pdata)[i].key))
			if match {
				dmutex.Unlock()
				nvalue := strings.Trim((*pdata)[i].value, "\n")
				return nvalue
			}
		}
	}
	dmutex.Unlock()
	// no matching key found, return empty string
	return ""
}

func get_data_value(value string) string {
	var i int64
	var match bool

	// svalue := strings.Trim(value, "\n")
	regexp := regexp.MustCompile(value)

	dmutex.Lock()
	for i = 0; i < maxdata; i++ {
		if (*pdata)[i].used {
			match = regexp.Match([]byte((*pdata)[i].value))
			if match {
				dmutex.Unlock()
				nvalue := strings.Trim((*pdata)[i].key, "\n")
				return nvalue
			}
		}
	}
	dmutex.Unlock()
	// no matching value found, return empty string
	return ""
}

func remove_data(key string) string {
	var i int64
	var match bool
	var value string

	skey := strings.Trim(key, "\n")
	regexp := regexp.MustCompile(skey)

	dmutex.Lock()
	for i = 0; i < maxdata; i++ {
		if (*pdata)[i].used {
			match = regexp.Match([]byte((*pdata)[i].key))
			if match {
				value = (*pdata)[i].value
				(*pdata)[i].used = false
				(*pdata)[i].key = ""
				(*pdata)[i].value = ""
				dmutex.Unlock()
				return value
			}
		}
	}
	dmutex.Unlock()
	// no matching key found, return empty string
	return ""
}

func save_data(file_path string) int {
	var i int64 = 0
	// create file
	f, err := os.Create(file_path)
	if err != nil {
		fmt.Println("Error opening database file: " + file_path + err.Error())
		return 1
	}
	// remember to close the file
	defer f.Close()

	// write header
	_, err = f.WriteString("l1vmgodata database\n")
	if err != nil {
		fmt.Println("Error writing database file:", err.Error())
		return 1
	}

	// write data loop
	for i = 0; i < maxdata; i++ {
		if (*pdata)[i].used {
			dmutex.Lock()
			value_save := strings.Trim((*pdata)[i].value, "\n")
			_, err = f.WriteString(":" + (*pdata)[i].key + " \"" + value_save + "\"\n")
			dmutex.Unlock()
			if err != nil {
				fmt.Println("Error writing database file:", err.Error())
				return 1
			}
		}
	}
	return 0
}

func load_data(file_path string) int {
	var i int64 = 0
	var header_line = 0
	var key string
	var value string
	// load database file
	file, err := os.Open(file_path)
	if err != nil {
		fmt.Println("Error opening database file: " + file_path + " " + err.Error())
		return 1
	}
	// remember to close the file
	defer file.Close()

	// read and check header
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if i < maxdata {
			// enough data memory, load data
			if header_line == 0 {
				if line != "l1vmgodata database" {
					fmt.Println("Error opening database file: " + file_path + " not a l1vmgodata database!")
					return 1
				}
				header_line = 1
			} else {
				fmt.Println("read: " + line + "\n")
				key, value = split_data(line)
				if key != "" {
					// store data
					dmutex.Lock()
					(*pdata)[i].used = true
					(*pdata)[i].key = key
					(*pdata)[i].value = value
					dmutex.Unlock()
					i++
				}
			}
		} else {
			fmt.Println("Error reading data base: out of memory: entries overflow!")
			return 1
		}
	}
	return 0
}

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

func split_data(input string) (string, string) {
	var i int = 0
	var j int = 0
	var search bool = true
	var copy bool = true
	var inkey string = ""
	var invalue string = ""
	var inplen int = 0
	inplen = len(input)
	for search {
		if input[i] == ':' {
			// store chars until next space char
			if i >= inplen {
				copy = false
				search = false
				return "", ""
			}
			i++
			for copy {
				if input[i] != ' ' {
					inkey = inkey + string(input[i])
				} else {
					copy = false
					search = false
				}
				i++
				if i >= inplen {
					copy = false
					search = false
				}
			}
		}
		i++
	}

	// read chars into data
	for j = i; j < inplen; j++ {
		if input[i] != '"' {
			invalue = invalue + string(input[i])
		}
		i++
	}
	return inkey, invalue
}

func split_key(input string) string {
	var i int = 0
	var search bool = true
	var copy bool = true
	var inkey string = ""
	var inplen int = 0
	inplen = len(input)
	for search {
		if input[i] == ':' {
			// store chars until next space char
			if i >= inplen {
				copy = false
				search = false
				return ""
			}
			i++
			for copy {
				if input[i] != ' ' {
					inkey = inkey + string(input[i])
				} else {
					copy = false
					search = false
				}
				i++
				if i >= inplen {
					copy = false
					search = false
				}
			}
		}
		i++
	}
	return inkey
}

func split_value(input string) string {
	var i int = 0
	var search bool = true
	var copy bool = true
	var invalue string = ""
	var inplen int = 0
	inplen = len(input)
	for search {
		if input[i] == '"' {
			// store chars until next quote char
			if i >= inplen {
				copy = false
				search = false
				return ""
			}
			i++
			for copy {
				if input[i] != '"' {
					invalue = invalue + string(input[i])
				} else {
					copy = false
					search = false
				}
				i++
				if i >= inplen {
					copy = false
					search = false
				}
			}
		}
		i++
	}
	return invalue
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
			fmt.Println("Error reading:", err.Error())
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
				} else {
					_, err = connection.Write([]byte("ERROR\n"))
				}
			} else {
				_, err = connection.Write([]byte("ERROR\n"))
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
					_, err = connection.Write([]byte("\n"))
				} else {
					_, err = connection.Write([]byte("ERROR\n"))
				}
			} else {
				_, err = connection.Write([]byte("ERROR\n"))
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
					_, err = connection.Write([]byte("\n"))
				} else {
					_, err = connection.Write([]byte("ERROR\n"))
				}
			} else {
				_, err = connection.Write([]byte("ERROR\n"))
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
					_, err = connection.Write([]byte("\n"))
				} else {
					_, err = connection.Write([]byte("ERROR\n"))
				}
			} else {
				_, err = connection.Write([]byte("ERROR\n"))
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
				} else {
					_, err = connection.Write([]byte("OK\n"))
				}
			} else {
				_, err = connection.Write([]byte("ERROR\n"))
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
				} else {
					_, err = connection.Write([]byte("OK\n"))
				}
			} else {
				_, err = connection.Write([]byte("ERROR\n"))
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
		maxdata = user_maxdata
	}
	fmt.Println("allocating ", maxdata, " space for data")
	servdata := make([]data, maxdata) // make serverdata spice
	pdata = &servdata

	init_data()
	run_server()
}
