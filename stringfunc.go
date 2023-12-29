// stringfunc.go - database in go
/*
 * This file stringfunc.go is part of L1VMgodata.
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
	"strings"
)

func get_client_ip(address string) string {
	var i int = 0
	var ip_str string = ""
	// get ip part before port: 192.168.0.1:4000
	for address[i] != ':' {
		ip_str = ip_str + string(address[i])
		i++
	}
	return ip_str
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

func split_data_json(input string) (string, string) {
	var i int = 0
	var search bool = true
	var copy bool = true
	var inkey string = ""
	var invalue string = ""
	var inplen int = 0
	var pos int = 0
	var asciicode int = 0
	inplen = len(input)
	// search for: "key":
	pos = strings.Index(input, "\"key\":")
	if pos == -1 {
		// error: key part not found in string line
		return "", ""
	}
	// get key
	i = pos + 6
	for search {
		asciicode = int(input[i])
		if asciicode == 34 {
			// found quote char
			if i >= inplen {
				copy = false
				search = false
				return "", ""
			}
			i++
			for copy {
				asciicode = int(input[i])
				if asciicode != 34 {
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

	search = true
	copy = true
	pos = strings.Index(input, "\"value\":")
	if pos == -1 {
		// error: key part not found in string line
		return "", ""
	}
	// get value
	i = pos + 8
	for search {
		asciicode = int(input[i])
		if asciicode == 34 {
			// found quote char
			if i >= inplen {
				copy = false
				search = false
				return "", ""
			}
			i++
			for copy {
				asciicode = int(input[i])
				if asciicode != 34 {
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
		if input[i] == ':' && i < inplen {
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
		if i >= inplen {
			search = false
		}
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
		if input[i] == '\'' && i < inplen {
			// store chars until next quote char
			if i >= inplen {
				copy = false
				search = false
				return ""
			}
			i++
			for copy {
				if input[i] != '\'' {
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
		if i >= inplen {
			search = false
		}
	}
	return invalue
}
