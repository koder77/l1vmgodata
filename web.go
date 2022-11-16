// web.go - database in go
/*
 * This file web.go is part of L1VMgodata.
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

package main

import (
	"fmt"
	"flag"
    "log"
	"net/http"
)

func parse_web(w http.ResponseWriter, command string, key string, value string) {
	var key_ret string
	var value_ret string
	var used_elements uint64

	switch command {
		case STORE_DATA:
			if store_data(key, value) != 0 {
				fmt.Fprintf(w, "ERROR can't store data!\n")
			} else {
				fmt.Fprintf(w, "data stored!\n");
			}

		case GET_DATA_KEY:
			value_ret = get_data_key(key)
			fmt.Fprintf(w, "key: %s, value: %s\n", key, value_ret)

		case GET_DATA_VALUE:
			key_ret = get_data_value(value)
			fmt.Fprintf(w, "key: %s, value: %s\n", key_ret, value)

		case REMOVE_DATA:
			value_ret = remove_data (key)
			fmt.Fprintf(w, "%s\n", value_ret)

		case SAVE_DATA:
			if save_data (value) != 0 {
				fmt.Fprintf(w, "ERROR can't save database %s !\n", value)
			} else {
				fmt.Fprintf(w, "database %s saved!\n", value)
			}

		case LOAD_DATA:
			if load_data (value) != 0 {
				fmt.Fprintf(w, "ERROR can't load database %s !\n", value)
			} else {
				fmt.Fprintf(w, "database %s loaded!\n", value)
			}

		case SAVE_DATA_JSON:
			if save_data_json (value) != 0 {
				fmt.Fprintf(w, "ERROR can't save JSON database %s !\n", value)
			} else {
				fmt.Fprintf(w, "JSON database %s saved!\n", value)
			}

		case LOAD_DATA_JSON:
			if load_data_json (value) != 0 {
				fmt.Fprintf(w, "ERROR can't load JSON database %s !\n", value)
			} else {
				fmt.Fprintf(w, "JSON database %s loaded!\n", value)
			}

		case ERASE_DATA:
			init_data()
			fmt.Fprintf(w, "ALL DATA ERASED!\n");

		case GET_USED_ELEMENTS:
			used_elements = get_used_elements()
			fmt.Fprintf(w, "usage: %d of %d\n", used_elements, maxdata)

		default:
			fmt.Fprintf(w, "ERROR! UNKNOWN COMMAND!")
	}
}

func hello(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	switch r.Method {
	case "GET":
		 http.ServeFile(w, r, "form.html")
	case "POST":
		// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		//fmt.Fprintf(w, "Post from website! r.PostFrom = %v\n", r.PostForm)
		key := r.FormValue("key")
		value := r.FormValue("value")
		command := r.FormValue("command")
		//fmt.Fprintf(w, "key = %s\n", key)
		//fmt.Fprintf(w, "value = %s\n", value)
		//fmt.Fprintf(w, "command = %s\n", command)
		parse_web(w, command, key, value)
	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}

func handle_http_request() {
	fmt.Println("Web Listening on " + server_host + ":" + server_http_port)
	fmt.Println("Waiting for client...")

	port := flag.String("p", server_http_port, "port to serve on")
	// directory := flag.String("d", "./statics", "the directory of static file to host")
	flag.Parse()

	// http.Handle("/", http.FileServer(http.Dir(*directory)))
	http.HandleFunc("/", hello)
	// log.Printf("Serving %s on HTTP port: %s\n", *directory, *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
