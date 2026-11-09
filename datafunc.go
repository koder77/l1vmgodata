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
	"github.com/cespare/xxhash/v2"
	"github.com/pbnjay/memory"
	"regexp"
	"strconv"
	"strings"
	"unsafe"
)

// index_add records a key index. The caller must hold dmutex.
func index_add(key string, i uint64) {
	key_index[key] = append(key_index[key], i)
}

// index_remove removes one index for a key. The caller must hold dmutex.
func index_remove(key string, i uint64) {
	indices := key_index[key]
	for n := 0; n < len(indices); n++ {
		if indices[n] == i {
			indices = append(indices[:n], indices[n+1:]...)
			break
		}
	}
	if len(indices) == 0 {
		delete(key_index, key)
	} else {
		key_index[key] = indices
	}
}

// find_key_index returns the first data index for an exact key match.
// The caller must hold dmutex.
func find_key_index(key string) (uint64, bool) {
	for _, i := range key_index[key] {
		if (*pdata)[i].used && (*pdata)[i].key == key {
			return i, true
		}
	}
	return 0, false
}

// store_data_at writes a data entry and updates the key index.
// The caller must hold dmutex.
func store_data_at(i uint64, key string, value string) {
	(*pdata)[i].used = true
	(*pdata)[i].key = key
	(*pdata)[i].value = value
	(*pdata)[i].key_hash = xxhash.Sum64String(key)
	(*pdata)[i].value_hash = xxhash.Sum64String(value)
	index_add(key, i)
	used_count++
}

// search if key was already set and return 1, or 0 if not already set!
func search_key(search_key string) (int, uint64) {
	dmutex.Lock()
	defer dmutex.Unlock()

	if i, found := find_key_index(search_key); found {
		// key already set
		return 1, i
	}
	// key not found
	return 0, maxdata
}

func init_data() {
	var i uint64

	dmutex.Lock()
	defer dmutex.Unlock()

	key_index = make(map[string][]uint64)
	used_count = 0
	data_index = 0
	free_index = 0

	if pdata == nil {
		return
	}

	for i = 0; i < maxdata; i++ {
		(*pdata)[i].used = false
		(*pdata)[i].key = ""
		(*pdata)[i].value = ""
		(*pdata)[i].key_hash = 0
		(*pdata)[i].value_hash = 0
		(*pdata)[i].links = nil
	}
}

func get_free_space() (int, uint64) {
	var i uint64
	dmutex.Lock()
	// check if free_index is free
	if free_index < maxdata-1 {
		if !(*pdata)[free_index].used {
			// free_index is free, increase it by 1 to point to next place
			free_index = free_index + 1
			dmutex.Unlock()
			return 0, free_index - 1
		}
	}

	// search for free place
	for i = 0; i < maxdata; i++ {
		if !(*pdata)[i].used {
			if free_index < maxdata-1 {
				free_index = i + 1
			}
			dmutex.Unlock()
			return 0, i
		}
	}
	dmutex.Unlock()
	// no free space found, return error code 1
	return 1, i
}

func try_to_allocate_more_space() int {
	var new_datasize uint64
	var free_system_ram uint64
	var max_alloc_size uint64
	var one_data data
	var add_size uint64 = 20000

	const one_data_size = uint64(unsafe.Sizeof(one_data))

	// get the free system RAM
	free_system_ram = memory.FreeMemory()
	max_alloc_size = free_system_ram - (one_data_size * add_size)
	if max_alloc_size < (one_data_size * add_size) {
		// error out of memory!
		fmt.Println("error: try_to_allocate_more_space: out of memory! dumping 'temp-data.db'")

		// save the data into the temp data file
		if save_data("temp-data.db") == 1 {
			fmt.Println("error: try_to_allocate_more_space: can't save 'temp-data.db' file!")
			return 1
		}

		return 1
	}

	// save the data into the temp data file
	if save_data("temp-data.db") == 1 {
		fmt.Println("error: try_to_allocate_more_space: can't save 'temp-data.db' file!")
		return 1
	}

	// clear pdata
	pdata = nil

	// allocate bigger data slice
	new_datasize = maxdata + add_size
	newdata := make([]data, new_datasize) // make serverdata slice
	pdata = &newdata
	maxdata = new_datasize

	fmt.Println("new datasize:", maxdata)

	init_data()

	// load temporary saved database
	data_index = 0

	if load_data("temp-data.db") == 1 {
		fmt.Println("error: try_to_allocate_more_space: can't load 'temp-data.db' file into new data slice!")
		return 1
	}

	return 0
}

func store_data(key string, value string) uint64 {
	var i uint64 = 0
	var err int = 0
	var new_key bool = false

	err, i = search_key(key)
	if err == 0 {
		// key not already used, get free space
		new_key = true
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
	if new_key {
		store_data_at(i, key, value)
	} else {
		(*pdata)[i].used = true
		(*pdata)[i].key = key
		(*pdata)[i].value = value
		(*pdata)[i].key_hash = xxhash.Sum64String(key)
		(*pdata)[i].value_hash = xxhash.Sum64String(value)
	}
	dmutex.Unlock()
	return 0
}

func store_data_new(key string, value string) uint64 {
	// don't check if key is already in database
	var i uint64 = 0
	var err int = 0

	// get free space
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

	// store data at index i
	dmutex.Lock()
	store_data_at(i, key, value)
	dmutex.Unlock()
	return 0
}

func get_data_key_regexp(key string) string {
	var i uint64

	skey := strings.Trim(key, "\n")
	// Compile the expression once, not once per database entry.
	re, err := regexp.Compile(skey)
	if err != nil {
		return ""
	}

	dmutex.Lock()
	defer dmutex.Unlock()
	for i = 0; i < maxdata; i++ {
		if (*pdata)[i].used {
			if re.MatchString((*pdata)[i].key) {
				nvalue := strings.Trim((*pdata)[i].key, "'\n")
				return nvalue
			}
		}
	}
	// no matching key found, return empty string
	return ""
}

func get_data_value_regexp(value string) string {
	var i uint64

	svalue := strings.Trim(value, "\n")
	// Compile the expression once, not once per database entry.
	re, err := regexp.Compile(svalue)
	if err != nil {
		return ""
	}

	dmutex.Lock()
	defer dmutex.Unlock()
	for i = 0; i < maxdata; i++ {
		if (*pdata)[i].used {
			if re.MatchString((*pdata)[i].value) {
				nvalue := strings.Trim((*pdata)[i].key, "'\n")
				return nvalue
			}
		}
	}
	// no matching value found, return empty string
	return ""
}

func get_data_key(key string) string {
	skey := strings.Trim(key, "\n")

	dmutex.Lock()
	defer dmutex.Unlock()

	if i, found := find_key_index(skey); found {
		return strings.Trim((*pdata)[i].value, "'\n")
	}
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
	var j uint64
	var value string
	skey := strings.Trim(key, "\n")

	dmutex.Lock()
	i, found := find_key_index(skey)
	if !found {
		dmutex.Unlock()
		// no matching key found, return empty string
		return ""
	}

	value = (*pdata)[i].value

	// remove all links pointing to this key
	for j = 0; j < maxdata; j++ {
		if !(*pdata)[j].used || len((*pdata)[j].links) == 0 {
			continue
		}
		links := (*pdata)[j].links
		kept := links[:0]
		for _, link := range links {
			if link != skey {
				kept = append(kept, link)
			}
		}
		(*pdata)[j].links = kept
	}

	(*pdata)[i].used = false
	(*pdata)[i].key = ""
	(*pdata)[i].value = ""
	(*pdata)[i].key_hash = 0
	(*pdata)[i].value_hash = 0
	(*pdata)[i].links = nil
	index_remove(skey, i)
	used_count--

	nvalue := strings.Trim(value, "'\n")
	dmutex.Unlock()
	return nvalue
}

// get info about data base usage, return used space
func get_used_elements() uint64 {
	dmutex.Lock()
	defer dmutex.Unlock()
	// return free data space
	return used_count
}

// link functions ==============================================================
func get_data_key_compare(key string) (string, uint64) {
	// don't use regex to compare, using normal string compare to find exact match
	skey := strings.Trim(key, "\n")

	dmutex.Lock()
	defer dmutex.Unlock()

	if i, found := find_key_index(skey); found {
		nvalue := strings.Trim((*pdata)[i].value, "'\n")
		return nvalue, i
	}
	// no matching key found, return empty string
	return "", maxdata
}

func set_link(key string, keylink string) int {
	// set link between key and keylink data entries

	var k uint64
	var i uint64
	var retstr string

	retstr, k = get_data_key_compare(key)
	if retstr == "" {
		// key not found
		// return error code
		return 1
	}

	retstr, _ = get_data_key_compare(keylink)
	if retstr == "" {
		// key not found
		// return error code
		return 1
	}

	// both key and keylink are found
	// check if link was already set
	dmutex.Lock()
	for i = 0; i < uint64(len((*pdata)[k].links)); i++ {
		if (*pdata)[k].links[i] != "" {
			// error return, link was already set!
			dmutex.Unlock()
			return 1
		}
	}

	// set the link
	(*pdata)[k].links = append((*pdata)[k].links, keylink)
	dmutex.Unlock()

	return 0
}

func remove_element_by_index[T any](slice []T, index uint64) []T {
	return append(slice[:index], slice[index+1:]...)
}

func remove_link(key string, keylink string) int {
	// set link between key and keylink data entries

	var k uint64
	var i uint64
	var retstr string

	retstr, k = get_data_key_compare(key)
	if retstr == "" {
		// key not found
		// return error code
		return 1
	}

	retstr, _ = get_data_key_compare(keylink)
	if retstr == "" {
		// keylink not found
		// return error code
		return 1
	}

	dmutex.Lock()

	// both key and keylink are found
	// search the keylink string index in links
	for i = 0; i < uint64(len((*pdata)[k].links)); i++ {
		if (*pdata)[k].links[i] == keylink {
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

	retstr = (*pdata)[k].links[link_index]

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
