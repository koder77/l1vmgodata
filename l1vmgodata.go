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
	"sync"
)

// socket
const (
	SERVER_TYPE = "tcp"
)

// data base commands
const (
	STORE_DATA          = "store data"
	GET_DATA_KEY        = "get key"
	GET_DATA_VALUE      = "get value"
	REMOVE_DATA         = "remove"
	CLOSE_CONNECTION    = "close"
	SAVE_DATA           = "save"
	LOAD_DATA           = "load"
	SAVE_DATA_JSON      = "json-export"
	LOAD_DATA_JSON      = "json-import"
	SAVE_DATA_CSV       = "csv-export"
	LOAD_DATA_CSV       = "csv-import"
	SAVE_DATA_TABLE_CSV = "csv-table-export"
	LOAD_DATA_TABLE_CSV = "csv-table-import"
	ERASE_DATA          = "erase all"
	GET_USED_ELEMENTS   = "usage"
	SET_LINK            = "set-link"
	REMOVE_LINK         = "rem-link"
	GET_LINKS_NUMBER    = "get-links-number"
	GET_LINK_NAME       = "get-link-name"
	EXIT                = "exit"
)

type data struct {
	used  bool
	key   string
	value string
	links []uint64
}

var maxdata uint64 = 10000 // max data number
var data_index uint64 = 0  // for loading multiple databases into one big database
var server_port string = "2000"
var server_http_port string = ""
var server_host = "localhost"
var server net.Listener
var pdata *[]data

// ip addresses whitelist
var ip_whitelist []string
var ip_whitelist_ind uint64 = 0

var dmutex sync.Mutex // data mutex

func read_ip_whitelist() bool {
	// load database file
	file, err := os.Open("whitelist.txt")
	if err != nil {
		fmt.Println("Error opening file: whitelist.txt " + err.Error())
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
		if err != nil {
			fmt.Println("Error accepting:", err.Error())
			os.Exit(1)
		}
		client_ip = get_client_ip(connection.RemoteAddr().String())
		if check_whitelist(client_ip) {
			fmt.Println("client connected:", client_ip)
			go process_client(connection)
		} else {
			fmt.Println("access denied!", client_ip)
		}
	}
}

func process_client(connection net.Conn) {
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

		// check close
		regexp_close := regexp.MustCompile(CLOSE_CONNECTION)
		match = regexp_close.Match([]byte(buffer[:mLen]))
		if match {
			_, err = connection.Write([]byte("OK\n"))
			run_loop = false
			continue
		}

		// check exit
		regexp_exit := regexp.MustCompile(EXIT)
		match = regexp_exit.Match([]byte(buffer[:mLen]))
		if match {
			_, err = connection.Write([]byte("OK\n"))

			// cleanup
			fmt.Println("cleaning up and exit!")
			init_data()
			pdata = nil
			server.Close()

			os.Exit(0)
		}

		// store data
		regexp_store := regexp.MustCompile(STORE_DATA)
		match = regexp_store.Match([]byte(buffer[:mLen]))
		if match {
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
		regexp_save := regexp.MustCompile(SAVE_DATA)
		match = regexp_save.Match([]byte(buffer[:mLen]))
		if match {
			// try to find matching path name
			value = split_value(string(buffer[:mLen]))
			if value != "" {
				if save_data(value) != 0 {
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
		regexp_load := regexp.MustCompile(LOAD_DATA)
		match = regexp_load.Match([]byte(buffer[:mLen]))
		if match {
			// try to find matching path name
			value = split_value(string(buffer[:mLen]))
			if value != "" {
				if load_data(value) != 0 {
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
		regexp_save_json := regexp.MustCompile(SAVE_DATA_JSON)
		match = regexp_save_json.Match([]byte(buffer[:mLen]))
		if match {
			// try to find matching path name
			value = split_value(string(buffer[:mLen]))
			if value != "" {
				if save_data_json(value) != 0 {
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
		regexp_load_json := regexp.MustCompile(LOAD_DATA_JSON)
		match = regexp_load_json.Match([]byte(buffer[:mLen]))
		if match {
			// try to find matching path name
			value = split_value(string(buffer[:mLen]))
			if value != "" {
				if load_data_json(value) != 0 {
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
		regexp_save_csv := regexp.MustCompile(SAVE_DATA_CSV)
		match = regexp_save_csv.Match([]byte(buffer[:mLen]))
		if match {
			// try to find matching path name
			value = split_value(string(buffer[:mLen]))
			if value != "" {
				if save_data_csv(value) != 0 {
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

		regexp_load_csv := regexp.MustCompile(LOAD_DATA_CSV)
		match = regexp_load_csv.Match([]byte(buffer[:mLen]))
		if match {
			// try to find matching path name
			value = split_value(string(buffer[:mLen]))
			if value != "" {
				if load_data_csv(value) != 0 {
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
		regexp_save_table_csv := regexp.MustCompile(SAVE_DATA_TABLE_CSV)
		match = regexp_save_table_csv.Match([]byte(buffer[:mLen]))
		if match {
			// try to find matching path name
			value = split_value(string(buffer[:mLen]))
			if value != "" {
				if save_data_table_csv(value) != 0 {
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
		regexp_load_table_csv := regexp.MustCompile(LOAD_DATA_TABLE_CSV)
		match = regexp_load_table_csv.Match([]byte(buffer[:mLen]))
		if match {
			// try to find matching path name
			value = split_value(string(buffer[:mLen]))
			if value != "" {
				if load_data_table_csv(value) != 0 {
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
		regexp_erase_all := regexp.MustCompile(ERASE_DATA)
		match = regexp_erase_all.Match([]byte(buffer[:mLen]))
		if match {
			init_data()
			_, err = connection.Write([]byte("OK\n"))
			if err != nil {
				fmt.Println("process_client: Error writing:", err.Error())
			}
			continue
		}

		// check get free elements
		regexp_get_used_elements := regexp.MustCompile(GET_USED_ELEMENTS)
		match = regexp_get_used_elements.Match([]byte(buffer[:mLen]))
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

		regexp_set_link := regexp.MustCompile(SET_LINK)
		match = regexp_set_link.Match([]byte(buffer[:mLen]))
		if match {
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

		regexp_remove_link := regexp.MustCompile(REMOVE_LINK)
		match = regexp_remove_link.Match([]byte(buffer[:mLen]))
		if match {
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

		regexp_get_number_of_links := regexp.MustCompile(GET_LINKS_NUMBER)
		match = regexp_get_number_of_links.Match([]byte(buffer[:mLen]))
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

		regexp_get_link := regexp.MustCompile(GET_LINK_NAME)
		match = regexp_get_link.Match([]byte(buffer[:mLen]))
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
}

func main() {
	fmt.Println("l1vmgodata <ip> <port> <http-port | off> [number of data entries]")
	fmt.Println("l1vmgodata start 0.9.1 ...")

	if len(os.Args) == 4 || len(os.Args) == 5 {
		server_host = os.Args[1]
		server_port = os.Args[2]
		server_http_port = os.Args[3]
	}
	if len(os.Args) == 5 {
		// get maxdata from command line
		user_maxdata, err := strconv.ParseInt(os.Args[4], 10, 64)
		if err != nil {
			panic(err)
		}
		maxdata = uint64(user_maxdata)
	}
	// check error case:
	if len(os.Args) <= 2 {
		fmt.Println("Arguments error! Need ip and ports!")
		os.Exit(1)
	}

	if !read_ip_whitelist() {
		os.Exit(1)
	}

	fmt.Println("allocating ", maxdata, " space for data")
	servdata := make([]data, maxdata) // make serverdata splice
	pdata = &servdata

	init_data()
	run_server()
}
