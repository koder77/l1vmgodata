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
	"strconv"
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
	var l uint64
	var linkslen uint64

	dmutex.Lock()
	for i = 0; i < maxdata; i++ {
		(*pdata)[i].used = false
		(*pdata)[i].key = ""
		(*pdata)[i].value = ""

		linkslen = uint64 (len ((*pdata)[i].links))
		if linkslen > 0 {
			// there are links set, remove all:
			for l = 0; l < linkslen; l++ {
				(*pdata)[i].links = remove_element_by_index((*pdata)[i].links, l)
			}
		}
	}
	dmutex.Unlock()
	data_index = 0
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

	err, i = search_key(key)
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
				nvalue := strings.Trim((*pdata)[i].value, "'\n")
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
				nvalue := strings.Trim((*pdata)[i].key, "'\n")
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
	var j uint64
	var l uint64
	//var linkindex uint64
	var match bool
	var value string
	var linkslen uint64 = 0
	skey := strings.Trim(key, "\n")
	// regexp := regexp.MustCompile(skey)

	dmutex.Lock()
	for i = 0; i < maxdata; i++ {
		if (*pdata)[i].used {
			// match = regexp.Match([]byte((*pdata)[i].key))
			match = (*pdata)[i].key == skey
			if match {
				fmt.Println ("remove data: found match...")
				for j = 0; j < maxdata; j++ {
					if (*pdata)[j].used {
						linkslen = uint64 (len ((*pdata)[j].links))
						if linkslen > 0 {
							for l = 0; l < linkslen; l++ {
								// try remove the data link, if possible
								// linkindex = (*pdata)[j].links[l]
								//if (*pdata)[linkindex].key == skey {
									dmutex.Unlock()
									fmt.Println("remove data: key: " + (*pdata)[j].key)
									_ = remove_link((*pdata)[j].key, skey)
									dmutex.Lock()
								//}
							}
						}
					}
				}
				value = (*pdata)[i].value
				(*pdata)[i].used = false
				(*pdata)[i].key = ""
				(*pdata)[i].value = ""

				nvalue := strings.Trim(value, "'\n")
				dmutex.Unlock()
				return nvalue
			}
		}
	}
	dmutex.Unlock()
	// no matching key found, return empty string
	return ""
}

func save_data(file_path string) int {
	var i uint64 = 0
	var l uint64 = 0
	var linkslen uint64  = 0
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
			value_save := strings.Trim((*pdata)[i].value, "'\n")
			_, err = f.WriteString(":" + (*pdata)[i].key + " \"" + value_save + "\"\n")
			if err != nil {
				fmt.Println("Error writing database file:", err.Error())
				dmutex.Unlock()
				return 1
			}

			// save links number
			linkslen = uint64 (len ((*pdata)[i].links))
			_, err = f.WriteString(":link" + " \"" + strconv.FormatInt (int64 (linkslen), 10) + "\"\n")
			dmutex.Unlock()
			if err != nil {
				fmt.Println("Error writing database file:", err.Error())
				return 1
			}

			// save links
			if linkslen > 0 {
				for l = 0; l < linkslen; l++  {
					dmutex.Lock()
					_, err = f.WriteString(":link" + " \"" + strconv.FormatInt (int64 ((*pdata)[i].links[l]), 10) + "\"\n")
					if err != nil {
						fmt.Println("Error writing database file:", err.Error())
						dmutex.Unlock()
						return 1
					}
					dmutex.Unlock()
			    }
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
	var l uint64 = 0
	var linkslen uint64  = 0
	var link uint64 = 0

	// load database file
	file, err := os.Open(file_path)
	if err != nil {
		fmt.Println("Error opening database file: " + file_path + " " + err.Error())
		return 1
	}
	// remember to close the file
	defer file.Close()

	// set i to data_index, so we can load more than one database. And don't start on zero index again!
	i = data_index

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
				// fmt.Println("read: " + line)
				key, value = split_data(line)

				//fmt.Println("load: key: " + key)

				if  key != "" && key != "link" {
					// store data
					dmutex.Lock()
					(*pdata)[i].used = true
					(*pdata)[i].key = key
					(*pdata)[i].value = value
					dmutex.Unlock()
				}

				// get links number
				scanner.Scan()
				line := scanner.Text()
				key, value = split_data(line)
				linkslen, _  =  strconv.ParseUint (value, 10, 64)

				//fmt.Printf ("load: links: %d\n", linkslen)

				if linkslen > 0 {
					// there are links, load them
					for l = 0; l < linkslen; l++ {
						scanner.Scan()
						line := scanner.Text()
						key, value = split_data(line)
						link, _  =  strconv.ParseUint (value, 10, 64)

						//fmt.Printf("load: link: %d\n", link)

						dmutex.Lock()
						(*pdata)[i].links = append ((*pdata)[i].links, link)
						dmutex.Unlock()
					}
				}
				i++
			}
		} else {
			fmt.Println("Error reading database: out of memory: entries overflow!")
			return 1
		}
	}
	data_index = i
	return 0
}

// export to .json data file
func save_data_json(file_path string) int {
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
	_, err = f.WriteString("{ \"l1vmgodata database\" :[\n")
	if err != nil {
		fmt.Println("Error writing database file:", err.Error())
		return 1
	}

	// write data loop
	for i = 0; i < maxdata; i++ {
		if (*pdata)[i].used {
			dmutex.Lock()
			value_save := strings.Trim((*pdata)[i].value, "\n")
			_, err = f.WriteString("{ \"key\": \"" + (*pdata)[i].key + "\", \"value\": \"" + value_save + "\" },\n")
			dmutex.Unlock()
			if err != nil {
				fmt.Println("Error writing database file:", err.Error())
				return 1
			}
		}
	}

	// remove last comma to create valid json file:
	_, err = f.Seek(-2, 1)
	if err != nil {
		fmt.Println("Error seeking database file:", err.Error())
		return 1
	}

	_, err = f.WriteString("\n]\n}\n")
	if err != nil {
		fmt.Println("Error writing database file:", err.Error())
		return 1
	}

	return 0
}

// import .json file
func load_data_json(file_path string) int {
	var i uint64 = 0
	var header_line = 0
	var key string
	var value string
	// open file
	file, err := os.Open(file_path)
	if err != nil {
		fmt.Println("Error opening database file: " + file_path + " " + err.Error())
		return 1
	}
	// remember to close the file
	defer file.Close()

	// set i to data_index, so we can load more than one database. And don't start on zero index again!
	i = data_index

	// read and check header
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if i < maxdata {
			// enough data memory, load data
			if header_line == 0 {
				if line != "{ \"l1vmgodata database\" :[" {
					fmt.Println("Error opening database file: " + file_path + " not a json l1vmgodata database!")
					return 1
				}
				header_line = 1
			} else {
				// fmt.Println("read: " + line)
				key, value = split_data_json(line)
				// fmt.Println("key: " +key + " value: " + value +"\n")

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
			fmt.Println("Error reading database: out of memory: entries overflow!")
			return 1
		}
	}
	data_index = i
	return 0
}

// get info about data base usage, return used space
func get_used_elements() (uint64) {
	var i uint64
	var free uint64 = 0

	dmutex.Lock()
	for i = 0; i < maxdata; i++ {
		if (*pdata)[i].used {
			free++
		}
	}
	dmutex.Unlock()
	// return free data space
	return free
}


// link functions ==============================================================
func get_data_key_compare(key string) (string, uint64) {
	// don't use regex to compare, using normal string compare to find exact match
	var i uint64

	skey := strings.Trim(key, "\n")

	dmutex.Lock()
	for i = 0; i < maxdata; i++ {
		if (*pdata)[i].used {
			if skey == (*pdata)[i].key {
				dmutex.Unlock()
				nvalue := strings.Trim((*pdata)[i].value, "'\n")
				return nvalue, i
			}
		}
	}
	dmutex.Unlock()
	// no matching key found, return empty string
	return "", i
}

func set_link(key string, keylink string) int {
	// set link between key and keylink data entries

	var k uint64
	var l uint64
	var i uint64
	var retstr string

	retstr, k = get_data_key_compare(key)
	if (retstr == ""){
		// key not found
		// return error code
		return 1
	}

	retstr, l = get_data_key_compare(keylink)
	if (retstr == ""){
		// key not found
		// return error code
		return 1
	}

	// both key and keylink are found
	// check if link was already set
	dmutex.Lock()
	for i = 0; i < uint64 (len ((*pdata)[k].links)); i++ {
		if (*pdata)[k].links[i] == l {
			// error return, link was already set!
			dmutex.Unlock()
			return 1
		}
	}

	// set the link
	(*pdata)[k].links = append((*pdata)[k].links, l)
	dmutex.Unlock()

	return 0
}

func remove_element_by_index[T any](slice []T, index uint64 ) []T {
	return append(slice[:index], slice[index+1:]...)
}

func remove_link(key string, keylink string) int {
	// set link between key and keylink data entries

	var k uint64
	var l uint64
	var i uint64
	var retstr string

	retstr, k = get_data_key_compare(key)
	if (retstr == ""){
		// key not found
		// return error code
		return 1
	}

	retstr, l = get_data_key_compare(keylink)
	if (retstr == ""){
		// keylink not found
		// return error code
		return 1
	}

	dmutex.Lock()

	// both key and keylink are found
	// search the keylink string index in links
	for i = 0; i <  uint64 (len ((*pdata)[k].links)); i++  {
		if (*pdata)[k].links[i] == l {
			// found matching index in the keys link list
			// rermove it
			(*pdata)[k].links = remove_element_by_index((*pdata)[k].links, i)
			dmutex.Unlock()
            return 0
		}
	}

	// link not found
	dmutex.Unlock()
	return 1
}

func get_number_of_links(key string) (uint64, string) {
	var linkslen uint64
	var k uint64
	var retstr string

	retstr, k = get_data_key_compare(key)
	if (retstr == ""){
		// key not found
		// return error code
		return 1, ""
	}

	linkslen = uint64 (len ((*pdata)[k].links))

	return linkslen, retstr
}

func get_link(key string, link_index uint64) string {
	var linkslen uint64
	var k uint64
	var l uint64
	var retstr string

	retstr, k = get_data_key_compare(key)
	if (retstr == ""){
		// key not found
		// return error code
		return ""
	}

	linkslen = uint64 (len ((*pdata)[k].links))
    if link_index < 0 || link_index >= linkslen {
		// error link index out of range
		return ""
	}

	l = (*pdata)[k].links[link_index]
	retstr = (*pdata)[l].key

	return retstr
}
