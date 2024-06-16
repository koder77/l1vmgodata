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
	"fmt"
	"strconv"
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
	var l uint64
	var linkslen uint64

	dmutex.Lock()
	for i = 0; i < maxdata; i++ {
		(*pdata)[i].used = false
		(*pdata)[i].key = ""
		(*pdata)[i].value = ""

		linkslen = uint64(len((*pdata)[i].links))
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

func try_to_allocate_more_space() int {
	var newdatasize uint64

	// save the data into the temp data file
	if save_data("temp-data.db") == 1 {
		fmt.Println("error: try_to_allocate_more_space: can't save 'temp-data' file!")
		return 1
	}

	// clear pdata
	pdata = nil

	// allocate bigger data slice
	newdatasize = maxdata + 10000
	newdata := make([]data, newdatasize) // make serverdata slice
	pdata = &newdata
	maxdata = newdatasize

	// load temporary saved database
	data_index = 0

	if load_data("temp-data.db") == 1 {
		fmt.Println("error: try_to_allocate_more_space: can't load 'temp-data' file into new data slice!")
		return 1
	}

	return 0
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
			// try to allocate bigger array
			err = try_to_allocate_more_space()
			if err == 1 {
				fmt.Println("error: can't allocate more space for data!")
				return 1
			}
			err, i = get_free_space()
	    	if err == 1 {
	            fmt.Println("error: can't get free space for data!")
				return 1
			}
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
	dmutex.Lock()
	for i = 0; i < maxdata; i++ {
		if (*pdata)[i].used {
			match = strings.Contains((*pdata)[i].key, skey)
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

	svalue := strings.Trim(value, "\n")

	dmutex.Lock()
	for i = 0; i < maxdata; i++ {
		if (*pdata)[i].used {
			match = strings.Contains((*pdata)[i].value, svalue)
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
				fmt.Println("remove data: found match...")
				for j = 0; j < maxdata; j++ {
					if (*pdata)[j].used {
						linkslen = uint64(len((*pdata)[j].links))
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

// get info about data base usage, return used space
func get_used_elements() uint64 {
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
	if retstr == "" {
		// key not found
		// return error code
		return 1
	}

	retstr, l = get_data_key_compare(keylink)
	if retstr == "" {
		// key not found
		// return error code
		return 1
	}

	// both key and keylink are found
	// check if link was already set
	dmutex.Lock()
	for i = 0; i < uint64(len((*pdata)[k].links)); i++ {
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

func remove_element_by_index[T any](slice []T, index uint64) []T {
	return append(slice[:index], slice[index+1:]...)
}

func remove_link(key string, keylink string) int {
	// set link between key and keylink data entries

	var k uint64
	var l uint64
	var i uint64
	var retstr string

	retstr, k = get_data_key_compare(key)
	if retstr == "" {
		// key not found
		// return error code
		return 1
	}

	retstr, l = get_data_key_compare(keylink)
	if retstr == "" {
		// keylink not found
		// return error code
		return 1
	}

	dmutex.Lock()

	// both key and keylink are found
	// search the keylink string index in links
	for i = 0; i < uint64(len((*pdata)[k].links)); i++ {
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
	if retstr == "" {
		// key not found
		// return error code
		return 1, ""
	}

	linkslen = uint64(len((*pdata)[k].links))

	return linkslen, retstr
}

func get_link(key string, link_index uint64) string {
	var linkslen uint64
	var k uint64
	var l uint64
	var retstr string

	retstr, k = get_data_key_compare(key)
	if retstr == "" {
		// key not found
		// return error code
		return ""
	}

	linkslen = uint64(len((*pdata)[k].links))
	if link_index < 0 || link_index >= linkslen {
		// error link index out of range
		return ""
	}

	l = (*pdata)[k].links[link_index]
	retstr = (*pdata)[l].key

	return retstr
}

func get_table_key(index uint64, key uint64) (string, string) {
	var i uint64
	var search_index string = ""
	var search_index_len uint64 = 0
	var keystr string = ""
	var start uint64 = 0
	var end uint64 = 0
	var match bool

	search_index = strconv.FormatUint(index, 10)
	search_index = search_index + "-"
	search_index = search_index + strconv.FormatUint(key, 10)
	search_index = search_index + "-"

	search_index_len = uint64(len(search_index))

	// DEBUG
	// fmt.Println ("get_table_key: " + search_index)

	dmutex.Lock()
	for i = 0; i < maxdata; i++ {
		if (*pdata)[i].used {
			match = strings.Contains((*pdata)[i].key, search_index)
			if match {
				// fmt.Println ("get_table_key: found match!")
				start = uint64(search_index_len)
				end = uint64(len((*pdata)[i].key))
				keystr = (*pdata)[i].key[start:end]
				dmutex.Unlock()
				return keystr, (*pdata)[i].value
			}
		}
	}
	dmutex.Unlock()
	return "", ""
}
