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
	// Split at the last colon, so IPv6 addresses like [::1]:2000 work too.
	i := strings.LastIndexByte(address, ':')
	if i < 0 {
		return address
	}
	ip := address[:i]
	if len(ip) >= 2 && ip[0] == '[' && ip[len(ip)-1] == ']' {
		return ip[1 : len(ip)-1]
	}
	return ip
}

func check_data(input string) int {
	inplen := len(input)

	if inplen <= 1 {
		// ERROR string to short
		return 1
	}

	colon_pos := strings.IndexByte(input, ':')
	quote_start := strings.IndexByte(input, '\'')
	quote := strings.Count(input, "'")

	quote_end := -1
	if quote_start >= 0 {
		if pos := strings.IndexByte(input[quote_start+1:], '\''); pos >= 0 {
			quote_end = quote_start + 1 + pos
		}
	}

	// check if everything makes sense
	if colon_pos <= 0 {
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
	inplen := len(input)
	if inplen <= 1 {
		return "", ""
	}

	colon_pos := strings.IndexByte(input, ':')
	if colon_pos < 0 {
		return "", ""
	}

	key_start := colon_pos + 1
	key_end := strings.IndexByte(input[key_start:], ' ')

	var inkey string
	value_start := inplen
	if key_end < 0 {
		inkey = input[key_start:]
	} else {
		key_end += key_start
		inkey = input[key_start:key_end]
		// The original parser skips the space and the opening quote.
		value_start = key_end + 2
		if value_start > inplen {
			value_start = inplen
		}
	}

	var invalue strings.Builder
	invalue.Grow(inplen - value_start)
	for i := value_start; i < inplen; i++ {
		c := input[i]
		if c != '\'' && c != '"' {
			invalue.WriteByte(c)
		}
	}

	return inkey, invalue.String()
}

func split_data_json(input string) (string, string) {
	key_pos := strings.Index(input, "\"key\":")
	if key_pos == -1 {
		// error: key part not found in string line
		return "", ""
	}

	// get key
	i := key_pos + len("\"key\":")
	quote_start := strings.IndexByte(input[i:], '"')
	if quote_start < 0 {
		return "", ""
	}
	i += quote_start + 1
	quote_end := strings.IndexByte(input[i:], '"')
	if quote_end < 0 {
		return "", ""
	}
	inkey := input[i : i+quote_end]

	value_pos := strings.Index(input, "\"value\":")
	if value_pos == -1 {
		// error: value part not found in string line
		return "", ""
	}

	// get value
	i = value_pos + len("\"value\":")
	quote_start = strings.IndexByte(input[i:], '"')
	if quote_start < 0 {
		return "", ""
	}
	i += quote_start + 1
	quote_end = strings.IndexByte(input[i:], '"')
	if quote_end < 0 {
		return "", ""
	}
	invalue := input[i : i+quote_end]

	return inkey, invalue
}

func split_data_csv(input string) (string, string) {
	comma_pos := strings.IndexByte(input, ',')
	if comma_pos <= 0 {
		return "", ""
	}

	// get key before the comma, dropping leading spaces
	inkey := strings.TrimLeft(input[:comma_pos], " ")
	// get part after the comma to the end of line
	invalue := input[comma_pos+1:]

	return inkey, invalue
}

func split_data_csv_table(input string, start int) (string, int, int) {
	inplen := len(input)

	if start == inplen {
		return "", 0, 0
	}

	comma_pos := strings.IndexByte(input[start:], ',')
	if comma_pos != -1 {
		comma_pos += start
		if comma_pos == 0 {
			return "", 0, 0
		}
		value := strings.TrimLeft(input[start:comma_pos], " ")
		return value, comma_pos, comma_pos + 1
	}

	if start < inplen {
		value := strings.TrimLeft(input[start:], " ")
		return value, 0, inplen
	}

	return "", 0, 0
}

func check_input_key(input string) int {
	if strings.IndexByte(input, ':') < 0 {
		// error no colon
		return 1
	}
	return 0
}

func check_input_value(input string) int {
	if strings.Count(input, "'") != 2 {
		// error no two single quotes
		return 1
	}
	return 0
}

func split_key(input string) string {
	if check_input_key(input) == 1 {
		// error
		return ""
	}

	colon_pos := strings.IndexByte(input, ':')
	key_start := colon_pos + 1
	if key_start >= len(input) {
		return ""
	}

	key_end := strings.IndexByte(input[key_start:], ' ')
	if key_end < 0 {
		return input[key_start:]
	}
	return input[key_start : key_start+key_end]
}

func split_value(input string) string {
	if check_input_value(input) == 1 {
		// error
		return ""
	}

	quote_start := strings.IndexByte(input, '\'')
	if quote_start < 0 {
		return ""
	}
	rest := input[quote_start+1:]
	quote_end := strings.IndexByte(rest, '\'')
	if quote_end < 0 {
		return ""
	}
	return rest[:quote_end]
}
