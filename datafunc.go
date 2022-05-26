// datafunc.go - database in go
/*
 * This file datafunc.go is part of L1VMgodata.
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
	"os"
	"regexp"
	"strings"
)

// search if key was already set and return 1, or 0 if not already set!
func search_key(search_key string) (int, uint64) {
	var i uint64

	dmutex.Lock()
	for i = 0; i < maxdata; i++ {
		if (*pdata)[i].used {
			if (*pdata)[i].key == search_key {
				dmutex.Unlock()
				// key already set
				return 1, i
			}
		}
	}
	dmutex.Unlock()
	// key not found
	return 0, i
}

func init_data() {
	var i uint64
	dmutex.Lock()
	for i = 0; i < maxdata; i++ {
		(*pdata)[i].used = false
		(*pdata)[i].key = ""
		(*pdata)[i].value = ""
	}
	dmutex.Unlock()
}

func get_free_space() (int, uint64) {
	var i uint64
	dmutex.Lock()
	for i = 0; i < maxdata; i++ {
		if !(*pdata)[i].used {
			dmutex.Unlock()
			return 0, i
		}
	}
	dmutex.Unlock()
	// no free space found, return error code 1
	return 1, i
}

func store_data(key string, value string) uint64 {
	var i uint64 = 0
	var err int = 0

	err, i = search_key (key)
	if err == 0 {
		// key not already used, get free space
		err, i = get_free_space()
		if err == 1 {
			// error: no free space
			return 1
		}
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
	var i uint64
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
	var i uint64
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
	var i uint64
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
	var i uint64 = 0
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
	var i uint64 = 0
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
