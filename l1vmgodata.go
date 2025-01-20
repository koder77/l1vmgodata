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
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

// socket
const (
	SERVER_TYPE = "tcp"
)

// data base commands
const (
	STORE_DATA            = "store data"
	STORE_DATA_NEW        = "store data new"
	GET_DATA_KEY          = "get key"
	GET_DATA_VALUE        = "get value"
	GET_DATA_REGEXP_KEY   = "get regex key"
	GET_DATA_REGEXP_VALUE = "get regex value"
	REMOVE_DATA           = "remove"
	CLOSE_CONNECTION      = "close"
	SAVE_DATA             = "save"
	LOAD_DATA             = "load"
	SAVE_DATA_JSON        = "json-export"
	LOAD_DATA_JSON        = "json-import"
	SAVE_DATA_CSV         = "csv-export"
	LOAD_DATA_CSV         = "csv-import"
	SAVE_DATA_TABLE_CSV   = "csv-table-export"
	LOAD_DATA_TABLE_CSV   = "csv-table-import"
	ERASE_DATA            = "erase all"
	GET_USED_ELEMENTS     = "usage"
	SET_LINK              = "set-link"
	REMOVE_LINK           = "rem-link"
	GET_LINKS_NUMBER      = "get-links-number"
	GET_LINK_NAME         = "get-link-name"
	EXIT                  = "exit"
	AUTH                  = "login"
)

// config files
const (
	USER_FILE             = "config/users.config"
	WHITELIST             = "config/whitelist.config"
	SETTINGS              = "config/settings.l1db"
)

type data struct {
	used  bool
	key   string
	value string
	links []uint64
}

var maxdata uint64 = 10000 // max data number
var data_index uint64 = 0  // for loading multiple databases into one big database
var free_index uint64 = 0  // next index of free data entry
var server_port string = "2000"
var server_http_port string = ""
var server_host = "localhost"
var server net.Listener
var pdata *[]data
var tls_flag string = ""
var tls_sock bool = false // set to true if TLS/SSL socket used

var user_file string = USER_FILE
var database_root string = ""

// ip addresses whitelist
var ip_whitelist []string
var ip_whitelist_ind uint64 = 0

// for blacklist.go, store blacklisted IP addresses
// after 3 failed logins put IP into blacklist for banning
var blacklist_ip []string
var blacklist_ip_ind uint64 = 0

var dmutex sync.Mutex      // data mutex
var server_run bool = true // set to false by "exit" command

func read_ip_whitelist() bool {
	// load database file
	file, err := os.Open(WHITELIST)
	if err != nil {
		fmt.Println("Error opening file: "+ WHITELIST + " " + err.Error())
		return false
	}
	// remember to close the file
	defer file.Close()

	// read one IP per line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// store ip
		if len(line) >= 2 {
			ip_whitelist = append(ip_whitelist, line)
			fmt.Println("whitelist:", ip_whitelist[ip_whitelist_ind])
			ip_whitelist_ind++
		}
	}
	return true
}

func check_whitelist(ip string) bool {
	var i uint64 = 0
	for i = 0; i < ip_whitelist_ind; i++ {
		if len(ip_whitelist[i]) > 1 && ip_whitelist[i] == ip {
			// found ip in whitelist, return true
			return true
		}
	}
	// ip not found in whitelist, return false
	return false
}

func run_server() {
	var client_ip string

	fmt.Println("run_server...")
	if server_http_port != "off" {
		go handle_http_request()
	}
	server, err := net.Listen(SERVER_TYPE, server_host+":"+server_port)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer server.Close()
	fmt.Println("Listening on " + server_host + ":" + server_port)
	fmt.Println("Waiting for client...")
	for {
		connection, err := server.Accept()
		defer connection.Close()
		if err != nil {
			fmt.Println("Error accepting:", err.Error())
			os.Exit(1)
		}
		client_ip = get_client_ip(connection.RemoteAddr().String())
		if check_whitelist(client_ip) {
			if check_blacklist(client_ip) {
				fmt.Println("Error: IP:", client_ip, "is blacklisted! Connection blocked!")
			} else {
				fmt.Println("client connected:", client_ip)
				if process_client(connection) == 1 {
					// shutdown
					return
				}
			}
		} else {
			fmt.Println("access denied!", client_ip)
		}
	}
}

func run_server_tls() {
	var client_ip string

	fmt.Println("run_server...")
	if server_http_port != "off" {
		go handle_http_request()
	}

	certFile := flag.String("cert", "cert.pem", "certificate PEM file")
	keyFile := flag.String("key", "cert.pem", "key PEM file")
	flag.Parse()

	cert, err := tls.LoadX509KeyPair(*certFile, *keyFile)
	if err != nil {
		fmt.Println("Error on TLS certificate: ", err.Error())
		os.Exit(1)
	}
	config := &tls.Config{Certificates: []tls.Certificate{cert}}

	server, err := tls.Listen(SERVER_TYPE, server_host+":"+server_port, config)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer server.Close()
	fmt.Println("Listening on " + server_host + ":" + server_port)
	fmt.Println("Waiting for client...")
	for {
		connection, err := server.Accept()
		defer connection.Close()
		if err != nil {
			fmt.Println("Error accepting:", err.Error())
			os.Exit(1)
		}
		client_ip = get_client_ip(connection.RemoteAddr().String())
		if check_whitelist(client_ip) {
			if check_blacklist(client_ip) {
				fmt.Println("Error: IP:", client_ip, "is blacklisted! Connection blocked!")
			} else {
				fmt.Println("client connected:", client_ip)
				if process_client(connection) == 1 {
					// shutdown
					return
				}
			}
		} else {
			fmt.Println("access denied!", client_ip)
		}
	}
}

func process_client(connection net.Conn) int {
	var run_loop bool = true
	buffer := make([]byte, 4096)
	var key string = ""
	var value string = ""
	var used_space uint64 = 0
	var used_space_percent float64 = 0.0
	var info string = ""
	var ret uint64
	var retstring string = ""
	var link uint64 = 0

	var match bool
	var inputstr string = ""

	var tls_auth bool = false // set to true if user password matches l1vmgodata password
	var user_role string = "normal-user"
	var user_ret = 0
	var return_value int = 0

	var authenticate_retries int = 1
	var client_ip string

	for run_loop {
		mLen, err := connection.Read(buffer)
		if err != nil {
			fmt.Println("process_client: Error reading:", err.Error())
			// end for loop
			run_loop = false
			continue
		}
		// fmt.Println("Received: '", string(buffer[:mLen]), "'")
		// fmt.Println("length: ", mLen)
		// fmt.Println(strings.HasPrefix(str, "the")) // false
		// check close

		inputstr = string(buffer)

		match = strings.HasPrefix(inputstr, CLOSE_CONNECTION)
		if match {
			_, err = connection.Write([]byte("OK\n"))
			run_loop = false
			continue
		}

		// check exit
		match = strings.HasPrefix(inputstr, EXIT)
		if match {
			if user_role == "admin" {
				_, err = connection.Write([]byte("OK\n"))
				if err != nil {
					fmt.Println("process_client: Error sending OK:", err.Error())
				}

				// cleanup
				fmt.Println("cleaning up and exit!")
				//init_data()
				// server.Close()
				// pdata = nil

				run_loop = false
				return_value = 1 // exit return value, stop main program
			} else {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error sending ERROR:", err.Error())
				}
			}
			continue
		}

		// check authentication
		match = strings.HasPrefix(inputstr, AUTH)
		if match {
			fmt.Print("got login... ")

			key = split_key(string(buffer[:mLen]))
			if key == "" {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error get key:", err.Error())
				}
			}

			value = split_value(string(buffer[:mLen]))
			if value == "" {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error get value:", err.Error())
				}
			}

			// hash value user password
			user_ret, user_role = check_user(user_file, key, value)
			if user_ret == 0 {
				tls_auth = true
				fmt.Println("ok!")

				_, err = connection.Write([]byte("OK\n"))
				if err != nil {
					fmt.Println("process_client: Error authenticate:", err.Error())
				}
			} else {
				fmt.Println("access denied! ")

				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error authenticate:", err.Error())
				}

				if authenticate_retries == 3 {
					// exit run loop, authentication failed
					// set ip in blacklist
					client_ip = get_client_ip(connection.RemoteAddr().String())
					set_blacklist_ip(client_ip)
					if !write_ip_blacklist() {
						fmt.Println("ERROR: saving blacklist file!")
					}
					run_loop = false

					fmt.Println("process_client: Error too many denied logins. IP:", client_ip, "banned!")
				}
				authenticate_retries++
			}
			continue
		}

		if tls_sock == true {
			if tls_auth != true {
				continue
			}
		}

		// store new data, don't check if key already used
		// extreme speedup over store data!!!
		match = strings.HasPrefix(inputstr, STORE_DATA_NEW)
		if match {
			if user_role == "read-only" {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error writing:", err.Error())
				}
				continue
			}

			// store key/value pair
			// try to store data
			if check_data(string(buffer[:mLen])) != 0 {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error writing:", err.Error())
				}
				continue
			}
			key, value = split_data(string(buffer[:mLen]))
			if key != "" {
				if store_data_new(key, value) == 0 {
					_, err = connection.Write([]byte("OK\n"))
					if err != nil {
						fmt.Println("process_client: Error writing:", err.Error())
					}
				} else {
					_, err = connection.Write([]byte("ERROR\n"))
					if err != nil {
						fmt.Println("process_client: Error writing:", err.Error())
					}
				}
			} else {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error writing:", err.Error())
				}
			}
			continue
		}

		// store data
		match = strings.HasPrefix(inputstr, STORE_DATA)
		if match {
			if user_role == "read-only" {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error writing:", err.Error())
				}
				continue
			}

			// store key/value pair
			// try to store data
			if check_data(string(buffer[:mLen])) != 0 {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error writing:", err.Error())
				}
				continue
			}
			key, value = split_data(string(buffer[:mLen]))
			if key != "" {
				if store_data(key, value) == 0 {
					_, err = connection.Write([]byte("OK\n"))
					if err != nil {
						fmt.Println("process_client: Error writing:", err.Error())
					}
				} else {
					_, err = connection.Write([]byte("ERROR\n"))
					if err != nil {
						fmt.Println("process_client: Error writing:", err.Error())
					}
				}
			} else {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error writing:", err.Error())
				}
			}
			continue
		}

		// get data key
		match = strings.HasPrefix(inputstr, GET_DATA_KEY)
		if match {
			// try to find matching key

			key = split_key(string(buffer[:mLen]))
			if key != "" {
				value = get_data_key(key)
				if value != "" {
					_, err = connection.Write([]byte(value))
					if err != nil {
						fmt.Println("process_client: Error writing:", err.Error())
					}
					_, err = connection.Write([]byte("\n"))
					if err != nil {
						fmt.Println("process_client: Error writing:", err.Error())
					}
				} else {
					_, err = connection.Write([]byte("ERROR\n"))
					if err != nil {
						fmt.Println("process_client: Error writing:", err.Error())
					}
				}
			} else {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error writing:", err.Error())
				}
			}
			continue
		}

		// get data value
		match = strings.HasPrefix(inputstr, GET_DATA_VALUE)
		if match {
			// try to find matching value
			value = split_value(string(buffer[:mLen]))
			if value != "" {
				key = get_data_value(value)
				if key != "" {
					_, err = connection.Write([]byte(key))
					if err != nil {
						fmt.Println("process_client: Error writing:", err.Error())
					}
					_, err = connection.Write([]byte("\n"))
					if err != nil {
						fmt.Println("process_client: Error writing:", err.Error())
					}
				} else {
					_, err = connection.Write([]byte("ERROR\n"))
					if err != nil {
						fmt.Println("process_client: Error writing:", err.Error())
					}
				}
			} else {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error writing:", err.Error())
				}
			}
			continue
		}

		// remove data, send value
		match = strings.HasPrefix(inputstr, REMOVE_DATA)
		if match {
			if user_role == "read-only" {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error writing:", err.Error())
				}
				continue
			}

			// try to find matching key
			key = split_key(string(buffer[:mLen]))
			if key != "" {
				value = remove_data(key)
				if value != "" {
					_, err = connection.Write([]byte(value))
					if err != nil {
						fmt.Println("process_client: Error writing:", err.Error())
					}
					_, err = connection.Write([]byte("\n"))
					if err != nil {
						fmt.Println("process_client: Error writing:", err.Error())
					}
				} else {
					_, err = connection.Write([]byte("ERROR\n"))
					if err != nil {
						fmt.Println("process_client: Error writing:", err.Error())
					}
				}
			} else {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error writing:", err.Error())
				}
			}
			continue
		}

		// get key with regex expression
		match = strings.HasPrefix(inputstr, GET_DATA_REGEXP_KEY)
		if match {
			// try to find matching key
			key = split_key(string(buffer[:mLen]))
			if key != "" {
				value = get_data_key_regexp(key)
				if value != "" {
					_, err = connection.Write([]byte(value))
					if err != nil {
						fmt.Println("process_client: Error writing:", err.Error())
					}
					_, err = connection.Write([]byte("\n"))
					if err != nil {
						fmt.Println("process_client: Error writing:", err.Error())
					}
				} else {
					_, err = connection.Write([]byte("ERROR\n"))
					if err != nil {
						fmt.Println("process_client: Error writing:", err.Error())
					}
				}
			} else {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error writing:", err.Error())
				}
			}
			continue
		}

		// get value with regex expression
		match = strings.HasPrefix(inputstr, GET_DATA_REGEXP_VALUE)
		if match {
			// try to find matching key
			value = split_value(string(buffer[:mLen]))
			if value != "" {
				key = get_data_value_regexp(value)
				if key != "" {
					_, err = connection.Write([]byte(key))
					if err != nil {
						fmt.Println("process_client: Error writing:", err.Error())
					}
					_, err = connection.Write([]byte("\n"))
					if err != nil {
						fmt.Println("process_client: Error writing:", err.Error())
					}
				} else {
					_, err = connection.Write([]byte("ERROR\n"))
					if err != nil {
						fmt.Println("process_client: Error writing:", err.Error())
					}
				}
			} else {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error writing:", err.Error())
				}
			}
			continue
		}

		// check save
		match = strings.HasPrefix(inputstr, SAVE_DATA)
		if match {
			if user_role == "read-only" {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error writing:", err.Error())
				}
				continue
			}

			// try to find matching path name
			value = split_value(string(buffer[:mLen]))
			if value != "" {
				if save_data(database_root+value) != 0 {
					_, err = connection.Write([]byte("ERROR\n"))
					if err != nil {
						fmt.Println("process_client: Error writing:", err.Error())
					}
				} else {
					_, err = connection.Write([]byte("OK\n"))
					if err != nil {
						fmt.Println("process_client: Error writing:", err.Error())
					}
				}
			} else {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error writing:", err.Error())
				}
			}
			continue
		}

		// check load
		match = strings.HasPrefix(inputstr, LOAD_DATA)
		if match {
			// try to find matching path name
			value = split_value(string(buffer[:mLen]))
			if value != "" {
				if load_data(database_root+value) != 0 {
					_, err = connection.Write([]byte("ERROR\n"))
					if err != nil {
						fmt.Println("process_client: Error writing:", err.Error())
					}
				} else {
					_, err = connection.Write([]byte("OK\n"))
					if err != nil {
						fmt.Println("process_client: Error writing:", err.Error())
					}
				}
			} else {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error writing:", err.Error())
				}
			}
			continue
		}

		// check save
		match = strings.HasPrefix(inputstr, SAVE_DATA_JSON)
		if match {
			if user_role == "read-only" {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error writing:", err.Error())
				}
				continue
			}

			// try to find matching path name
			value = split_value(string(buffer[:mLen]))
			if value != "" {
				if save_data_json(database_root+value) != 0 {
					_, err = connection.Write([]byte("ERROR\n"))
					if err != nil {
						fmt.Println("process_client: Error writing:", err.Error())
					}
				} else {
					_, err = connection.Write([]byte("OK\n"))
					if err != nil {
						fmt.Println("process_client: Error writing:", err.Error())
					}
				}
			} else {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error writing:", err.Error())
				}
			}
			continue
		}

		// check load
		match = strings.HasPrefix(inputstr, LOAD_DATA_JSON)
		if match {
			// try to find matching path name
			value = split_value(string(buffer[:mLen]))
			if value != "" {
				if load_data_json(database_root+value) != 0 {
					_, err = connection.Write([]byte("ERROR\n"))
					if err != nil {
						fmt.Println("process_client: Error loading:", err.Error())
					}
				} else {
					_, err = connection.Write([]byte("OK\n"))
					if err != nil {
						fmt.Println("process_client: Error loading:", err.Error())
					}
				}
			} else {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error loading:", err.Error())
				}
			}
			continue
		}

		// check save CSV
		match = strings.HasPrefix(inputstr, SAVE_DATA_CSV)
		if match {
			if user_role == "read-only" {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error writing:", err.Error())
				}
				continue
			}

			// try to find matching path name
			value = split_value(string(buffer[:mLen]))
			if value != "" {
				if save_data_csv(database_root+value) != 0 {
					_, err = connection.Write([]byte("ERROR\n"))
					if err != nil {
						fmt.Println("process_client: Error writing:", err.Error())
					}
				} else {
					_, err = connection.Write([]byte("OK\n"))
					if err != nil {
						fmt.Println("process_client: Error writing:", err.Error())
					}
				}
			} else {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error writing:", err.Error())
				}
			}
			continue
		}

		match = strings.HasPrefix(inputstr, LOAD_DATA_CSV)
		if match {
			// try to find matching path name
			value = split_value(string(buffer[:mLen]))
			if value != "" {
				if load_data_csv(database_root+value) != 0 {
					_, err = connection.Write([]byte("ERROR\n"))
					if err != nil {
						fmt.Println("process_client: Error writing:", err.Error())
					}
				} else {
					_, err = connection.Write([]byte("OK\n"))
					if err != nil {
						fmt.Println("process_client: Error writing:", err.Error())
					}
				}
			} else {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error writing:", err.Error())
				}
			}
			continue
		}

		// check save table CSV
		match = strings.HasPrefix(inputstr, SAVE_DATA_TABLE_CSV)
		if match {
			if user_role == "read-only" {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error writing:", err.Error())
				}
				continue
			}

			// try to find matching path name
			value = split_value(string(buffer[:mLen]))
			if value != "" {
				if save_data_table_csv(database_root+value) != 0 {
					_, err = connection.Write([]byte("ERROR\n"))
					if err != nil {
						fmt.Println("process_client: Error writing:", err.Error())
					}
				} else {
					_, err = connection.Write([]byte("OK\n"))
					if err != nil {
						fmt.Println("process_client: Error writing:", err.Error())
					}
				}
			} else {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error writing:", err.Error())
				}
			}
			continue
		}

		// check load table CSV
		match = strings.HasPrefix(inputstr, LOAD_DATA_TABLE_CSV)
		if match {
			// try to find matching path name
			value = split_value(string(buffer[:mLen]))
			if value != "" {
				if load_data_table_csv(database_root+value) != 0 {
					_, err = connection.Write([]byte("ERROR\n"))
					if err != nil {
						fmt.Println("process_client: Error writing:", err.Error())
					}
				} else {
					_, err = connection.Write([]byte("OK\n"))
					if err != nil {
						fmt.Println("process_client: Error writing:", err.Error())
					}
				}
			} else {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error writing:", err.Error())
				}
			}
			continue
		}

		// check erase all data
		// user role check
		match = strings.HasPrefix(inputstr, ERASE_DATA)
		if match {
			// user role check

			if user_role == "admin" {
				match = strings.HasPrefix(inputstr, ERASE_DATA)
				if match {
					init_data()
					_, err = connection.Write([]byte("OK\n"))
					if err != nil {
						fmt.Println("process_client: Error writing:", err.Error())
					}
					continue
				}
			} else {
				_, err = connection.Write([]byte("ERROR not admin user! Erasing denied!\n"))
				if err != nil {
					fmt.Println("process_client: Error writing:", err.Error())
				}
				continue
			}
		}

		// check get free elements
		match = strings.HasPrefix(inputstr, GET_USED_ELEMENTS)
		if match {
			used_space = get_used_elements()
			used_space_percent = float64(100 * used_space / maxdata)

			info = "USAGE " + strconv.FormatFloat(used_space_percent, 'f', 2, 64) + "% : " + strconv.FormatUint(used_space, 10) + " of " + strconv.FormatUint(maxdata, 10) + "\n"
			_, err = connection.Write([]byte(info))
			if err != nil {
				fmt.Println("process_client: Error sending space usage.", err.Error())
			}
			continue
		}

		match = strings.HasPrefix(inputstr, SET_LINK)
		if match {
			if user_role == "read-only" {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error writing:", err.Error())
				}
				continue
			}

			key = split_key(string(buffer[:mLen]))
			if key == "" {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error set link:", err.Error())
				}
			}

			value = split_value(string(buffer[:mLen]))
			if value == "" {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error set link:", err.Error())
				}
			}

			if set_link(key, value) != 0 {
				// error
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error set link:", err.Error())
				}
			} else {
				_, err = connection.Write([]byte("OK\n"))
				if err != nil {
					fmt.Println("process_client: Error set link:", err.Error())
				}
			}
			continue
		}

		match = strings.HasPrefix(inputstr, REMOVE_LINK)
		if match {
			if user_role == "read-only" {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error writing:", err.Error())
				}
				continue
			}

			key = split_key(string(buffer[:mLen]))
			if key == "" {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error remove link:", err.Error())
				}
			}

			value = split_value(string(buffer[:mLen]))
			if value == "" {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error remove link:", err.Error())
				}
			}

			if remove_link(key, value) != 0 {
				// error
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error remove link:", err.Error())
				}
			} else {
				_, err = connection.Write([]byte("OK\n"))
				if err != nil {
					fmt.Println("process_client: Error remove link:", err.Error())
				}
			}
			continue
		}

		match = strings.HasPrefix(inputstr, GET_LINKS_NUMBER)
		if match {
			key = split_key(string(buffer[:mLen]))
			if key == "" {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error get number of links:", err.Error())
				}
			}

			ret, retstring = get_number_of_links(key)
			if retstring == "" {
				// key not found
				// error
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error get number of links:", err.Error())
				}
			} else {
				// all ok write number of links
				info = strconv.FormatUint(ret, 10) + "\n"
				_, err = connection.Write([]byte(info))
				if err != nil {
					fmt.Println("process_client: Error get number of links:.", err.Error())
				}
			}
			continue
		}

		match = strings.HasPrefix(inputstr, GET_LINK_NAME)
		if match {
			key = split_key(string(buffer[:mLen]))
			if key == "" {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error get number of links:", err.Error())
				}
			}

			value = split_value(string(buffer[:mLen]))
			if value == "" {
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error remove link:", err.Error())
				}
			}

			link, _ = strconv.ParseUint(value, 10, 64)

			retstring = get_link(key, link)
			if retstring == "" {
				// key not found
				// error
				_, err = connection.Write([]byte("ERROR\n"))
				if err != nil {
					fmt.Println("process_client: Error get number of links:", err.Error())
				}
			} else {
				// all ok write link name
				info = retstring + "\n"
				_, err = connection.Write([]byte(info))
				if err != nil {
					fmt.Println("process_client: Error get number of links:.", err.Error())
				}
			}
		}

		// no matching command
		_, err = connection.Write([]byte("ERROR! UNKNOWN COMMAND!\n"))
		if err != nil {
			fmt.Println("process_client: Error sending unknown command error:", err.Error())
		}
	}

	connection.Close()

	return (return_value)
}

func main() {
	// var user_maxdata uint64 = 0
	var server_host_set bool = false
	var server_port_set bool = false
	var tls_flag_set bool = false
	var server_http_port_set bool = false

	fmt.Println("l1vmgodata <ip> <port> <tls=on | tls=off> <http-port | off> [number of data entries]")
	fmt.Println("l1vmgodata start 0.9.6 ...")

	fmt.Println("args: ", len(os.Args))

	if len(os.Args) == 5 || len(os.Args) == 6 {
		server_host = os.Args[1]
		server_port = os.Args[2]
		tls_flag = os.Args[3]
		server_http_port = os.Args[4]

		server_host_set = true
		server_port_set = true
		tls_flag_set = true
		server_http_port_set = true
	}
	if len(os.Args) == 6 {
		// get maxdata from command line
		user_maxdata, err := strconv.ParseInt(os.Args[5], 10, 64)
		if err != nil {
			panic(err)
		}
		maxdata = uint64(user_maxdata)
	}

	if !read_ip_whitelist() {
		os.Exit(1)
	}

	if !read_ip_blacklist() {
		os.Exit(1)
	}

	fmt.Println("allocating ", maxdata, " space for data")
	servdata := make([]data, maxdata) // make serverdata splice
	pdata = &servdata

	init_data()

	// get configuration from: "settings.l1db"
	if load_data(SETTINGS) != 0 {
		fmt.Println("error: can't load config file 'settings.l1db'!")
		init_data()
		pdata = nil
		os.Exit(1)
	}

	// get database root path
	database_root = get_data_key("database-root\n")
	if database_root == "" {
		fmt.Println("can't get key ':database-root' from config file 'settings.l1db'!")
	}

	// check for missing config

	// server host
	if server_host_set == false {
		server_host = get_data_key("host\n")
		if server_host == "" {
			fmt.Println("can't get key ':host' from config file 'settings.l1db'!")
		} else {
			server_host_set = true
		}
	}

	// server port
	if server_port_set == false {
		server_port = get_data_key("port\n")
		if server_port == "" {
			fmt.Println("can't get key ':port' from config file 'settings.l1db'!")
		} else {
			server_port_set = true
		}
	}

	// tls flag
	if tls_flag_set == false {
		tls_flag = get_data_key("tls\n")
		if tls_flag == "" {
			fmt.Println("can't get key ':tls' from config file 'settings.l1db'!")
		} else {
			tls_flag_set = true
		}
	}

	// server http port flag
	if server_http_port_set == false {
		server_http_port = get_data_key("http-port\n")
		if server_http_port == "" {
			fmt.Println("can't get key ':http-port' from config file 'settings.l1db'!")
		} else {
			server_http_port_set = true
		}
	}

	// check if all needed config is set
	if server_host_set == false {
		fmt.Println("Error: no server host set!")
		init_data()
		pdata = nil
		os.Exit(1)
	}

	if server_port_set == false {
		fmt.Println("Error: no server port set!")
		init_data()
		pdata = nil
		os.Exit(1)
	}

	if tls_flag_set == false {
		fmt.Println("Error: no tls config set!")
		init_data()
		pdata = nil
		os.Exit(1)
	}

	if server_http_port_set == false {
		fmt.Println("Error: no http port config set!")
		init_data()
		pdata = nil
		os.Exit(1)
	}

	// all config stuff load, clear config data base
	init_data()

	if tls_flag == "tls=on" {
		fmt.Println("running server: TLS on!")
		tls_sock = true
		run_server_tls()
		init_data()
		pdata = nil
		os.Exit(0)
	} else {
		fmt.Println("running server: normal socket!")
		run_server()
		init_data()
		pdata = nil
		os.Exit(0)
	}
}
