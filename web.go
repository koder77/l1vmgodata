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
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func send_form_head(w http.ResponseWriter) {
	fmt.Fprintf(w, "<!DOCTYPE html>")
	fmt.Fprintf(w, "<html>")
	fmt.Fprintf(w, "<head>")
	fmt.Fprintf(w, "<meta charset=\"UTF-8\" />")
	fmt.Fprintf(w, "<title>L1VMgodata - web</title>")
	fmt.Fprintf(w, "</head>")
	fmt.Fprintf(w, "<body>")
	fmt.Fprintf(w, "<div>")
	fmt.Fprintf(w, "<h2>L1VMgodata - web</h2>")
	fmt.Fprintf(w, "<br>")
	fmt.Fprintf(w, "<form method=\"POST\" action=\"/\">")
	fmt.Fprintf(w, "<label>command</label><input name=\"command\" type=\"text\" value=\"\"/>&nbsp;")
	fmt.Fprintf(w, "<label>key</label><input name=\"key\" type=\"text\" value=\"\"/>&nbsp;")
	fmt.Fprintf(w, "<label>value</label><input name=\"value\" type=\"text\" value=\"\"/>&nbsp;")
	fmt.Fprintf(w, "<input type=\"submit\" value=\"submit\"/>")
	fmt.Fprintf(w, "</form>")
	fmt.Fprintf(w, "<br>")
	fmt.Fprintf(w, "<br>")
}

func send_form_end(w http.ResponseWriter) {
	fmt.Fprintf(w, "</div>")
	fmt.Fprintf(w, "</body>")
	fmt.Fprintf(w, "</html>")
}

func parse_web(w http.ResponseWriter, command string, key string, value string) {
	var key_ret string
	var value_ret string
	var used_elements uint64
	var linkslen uint64
	var retstr string
	var linkindex uint64

	switch command {
	case STORE_DATA:
		send_form_head(w)

		if store_data(key, value) != 0 {
			fmt.Fprintf(w, "ERROR can't store data!\n")
		} else {
			fmt.Fprintf(w, "data stored!\n")
		}

		send_form_end(w)

	case GET_DATA_KEY:
		send_form_head(w)

		value_ret = get_data_key(key)
		fmt.Fprintf(w, "key: %s, value: %s\n", key, value_ret)

		send_form_end(w)

	case GET_DATA_VALUE:
		send_form_head(w)

		key_ret = get_data_value(value)
		fmt.Fprintf(w, "key: %s, value: %s\n", key_ret, value)

		send_form_end(w)

	case REMOVE_DATA:
		send_form_head(w)

		value_ret = remove_data(key)
		fmt.Fprintf(w, "%s\n", value_ret)

		send_form_end(w)

	case GET_DATA_REGEXP_KEY:
		send_form_head(w)

		value_ret = get_data_key_regexp(key)
		fmt.Fprintf(w, "%s\n", value_ret)

		send_form_end(w)

	case GET_DATA_REGEXP_VALUE:
		send_form_head(w)

		key_ret = get_data_value_regexp(value)
		fmt.Fprintf(w, "%s\n", key_ret)

		send_form_end(w)

	case SAVE_DATA:
		send_form_head(w)

		if save_data(value) != 0 {
			fmt.Fprintf(w, "ERROR can't save database %s !\n", value)
		} else {
			fmt.Fprintf(w, "database %s saved!\n", value)
		}

		send_form_end(w)

	case LOAD_DATA:
		send_form_head(w)

		if load_data(value) != 0 {
			fmt.Fprintf(w, "ERROR can't load database %s !\n", value)
		} else {
			fmt.Fprintf(w, "database %s loaded!\n", value)
		}

		send_form_end(w)

	case SAVE_DATA_JSON:
		send_form_head(w)

		if save_data_json(value) != 0 {
			fmt.Fprintf(w, "ERROR can't save JSON database %s !\n", value)
		} else {
			fmt.Fprintf(w, "JSON database %s saved!\n", value)
		}

		send_form_end(w)

	case LOAD_DATA_JSON:
		send_form_head(w)

		if load_data_json(value) != 0 {
			fmt.Fprintf(w, "ERROR can't load JSON database %s !\n", value)
		} else {
			fmt.Fprintf(w, "JSON database %s loaded!\n", value)
		}

		send_form_end(w)

	case ERASE_DATA:
		send_form_head(w)

		init_data()
		fmt.Fprintf(w, "ALL DATA ERASED!\n")

		send_form_end(w)

	case GET_USED_ELEMENTS:
		send_form_head(w)

		used_elements = get_used_elements()
		fmt.Fprintf(w, "usage: %d of %d\n", used_elements, maxdata)

		send_form_end(w)

	case SET_LINK:
		send_form_head(w)

		if set_link(key, value) != 0 {
			fmt.Fprintf(w, "ERROR can't set link %s !\n", value)
		} else {
			fmt.Fprintf(w, "link set!\n")
		}

		send_form_end(w)

	case REMOVE_LINK:
		send_form_head(w)

		if remove_link(key, value) != 0 {
			fmt.Fprintf(w, "ERROR can't remove link %s !\n", value)
		} else {
			fmt.Fprintf(w, "link removed!\n")
		}

		send_form_end(w)

	case GET_LINKS_NUMBER:
		send_form_head(w)

		linkslen, retstr = get_number_of_links(key)
		if retstr == "" {
			fmt.Fprintf(w, "ERROR can't get links number!\n")
		} else {
			fmt.Fprintf(w, "links: %d\n", linkslen)
		}

		send_form_end(w)

	case GET_LINK_NAME:
		send_form_head(w)

		linkindex, _ = strconv.ParseUint(value, 10, 64)

		retstr = get_link(key, linkindex)
		if retstr == "" {
			fmt.Fprintf(w, "ERROR can't get links name!\n")
		} else {
			fmt.Fprintf(w, "link name: %s\n", retstr)
		}

		send_form_end(w)

	default:
		send_form_head(w)

		fmt.Fprintf(w, "ERROR! UNKNOWN COMMAND!")

		send_form_end(w)
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
