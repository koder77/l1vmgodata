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
	"fmt"
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

func check_data(input string) int {
	var i int = 0
	var colon_pos int = 0
	var chars_start = 0
	var quote_start int = 0
	var quote_end int = 0
	var quote int = 0
	var inplen int = 0
	inplen = len(input)

	if inplen <= 1 {
		// ERROR string to short
		return 1
	}

	for i = 0; i < inplen; i++ {
		if colon_pos != 0 {
			if input[i] != ' ' {
				if chars_start == 0 {
					chars_start = i
				}
			}
		}
		if input[i] == ':' {
			colon_pos = i
		}
		if input[i] == '\'' {
			if quote_start != 0 {
				if quote_end == 0 {
					quote_end = i
				}
			}
			if quote_start == 0 {
				quote_start = i
			}
			quote = quote + 1
		}
	}

	// check if everything makes sense
	if colon_pos == 0 {
		// no colon found, error
		fmt.Println("check_data: error no colon found!")
		return 1
	}
	if quote != 2 {
		// no two quotes found, error
		fmt.Println("check_data: error no two quotes found!")
		return 1
	}

	if colon_pos > quote_start || colon_pos > quote_end {
		// colon not before quotes
		fmt.Println("check_data: error colon after quotes!")
		return 1
	}

	// all ok!
	return 0
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
	if inplen <= 1 {
		return "", ""
	}

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
		if input[i] != '\'' {
			if input[i] != '"' {
				invalue = invalue + string(input[i])
			}
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

func split_data_csv(input string) (string, string) {
	var i int = 0
	var inkey string = ""
	var invalue string = ""
	var inplen int = 0
	var comma_pos int = 0
	var got_char bool = false
	inplen = len(input)

	// check for comma
	for i = 0; i < inplen; i++ {
		if input[i] == ',' {
			comma_pos = i
			break
		}
	}

	if comma_pos == 0 {
		inkey = ""
		invalue = ""
		return inkey, invalue
	}

	// get key before the comma
	for i = 0; i < comma_pos; i++ {
		if input[i] == ' ' {
			if got_char == true {
				inkey = inkey + string(input[i])
			}
		} else {
			inkey = inkey + string(input[i])
			got_char = true
		}
	}

	// get part after the comma to the end of line
	for i = comma_pos + 1; i < inplen; i++ {
		invalue = invalue + string(input[i])
	}

	return inkey, invalue
}

func split_data_csv_table(input string, start int) (string, int, int) {
	var i int = 0
	var v int = 0
	var value string = ""
	var inplen int = 0
	var comma_pos int = 0
	var got_char bool = false
	inplen = len(input)

	if start == inplen {
		return "", 0, 0
	}

	// check for comma
	comma_pos = 0
	for i = start; i < inplen; i++ {
		if input[i] == ',' {
			comma_pos = i
			//start = comma_pos + 1  // for next comma search start
			break
		}
	}

	value = ""
	if comma_pos != 0 {
		for v = start; v < comma_pos; v++ {
			if input[v] == ' ' {
				if got_char == true {
					value = value + string(input[v])
				}
			} else {
				value = value + string(input[v])
				got_char = true
			}
		}
		return value, comma_pos, comma_pos + 1
	} else {
		if i == inplen {
			for v = start; v < inplen; v++ {
				if input[v] == ' ' {
					if got_char == true {
						value = value + string(input[v])
					}
				} else {
					value = value + string(input[v])
					got_char = true
				}
			}
			return value, comma_pos, inplen
		} else {
			return "", 0, 0
		}
	}
}

func check_input_key(input string) int {
	var i int = 0
	var search bool = true
	var inplen int = 0
	var colon bool = false

	inplen = len(input)
	for search {
		if input[i] == ':' {
			colon = true
		}

		i++
		if i >= inplen {
			search = false
		}
	}
	if colon == false {
		// error no two single quotes
		return 1
	}
	return 0
}

func check_input_value(input string) int {
	var i int = 0
	var search bool = true
	var inplen int = 0
	var single_quote = 0

	inplen = len(input)
	for search {
		if input[i] == '\'' {
			single_quote++
		}
		i++
		if i >= inplen {
			search = false
		}
	}
	if single_quote != 2 {
		// error no two single quotes
		return 1
	}
	return 0
}

func split_key(input string) string {
	var i int = 0
	var search bool = true
	var copy bool = true
	var inkey string = ""
	var inplen int = 0
	inplen = len(input)
	if check_input_key(input) == 1 {
		// error
		return ""
	}
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
	if check_input_value(input) == 1 {
		// error
		return ""
	}
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
