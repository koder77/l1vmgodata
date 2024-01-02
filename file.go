// file.go - database in go
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
// data file save/load export/import functions

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"strconv"
)

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

// export CSV
func save_data_csv(file_path string) int {
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
	_, err = f.WriteString("key, value\n")
	if err != nil {
		fmt.Println("Error writing database file:", err.Error())
		return 1
	}

	// write data loop
	dmutex.Lock()
	for i = 0; i < maxdata; i++ {
		if (*pdata)[i].used {
			value_save := strings.Trim((*pdata)[i].value, "'\n")
			_, err = f.WriteString((*pdata)[i].key + ", " + value_save + "\n")
			if err != nil {
				fmt.Println("Error writing database file:", err.Error())
				dmutex.Unlock()
				return 1
			}
		}
	}
	dmutex.Unlock()
	return 0
}

func load_data_csv(file_path string) int {
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

	// set i to data_index, so we can load more than one database. And don't start on zero index again!
	i = data_index

	// read and check header
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if i < maxdata {
			// enough data memory, load data
			if header_line == 0 {
				// skip header line
				header_line = 1
				continue
			}

			// fmt.Println("read: " + line)
			key, value = split_data_csv(line)

			//fmt.Println("load: key: " + key)
			// store data
			dmutex.Lock()
			(*pdata)[i].used = true
			(*pdata)[i].key = key
			(*pdata)[i].value = value
			dmutex.Unlock()
		}
	}
	data_index = i
	return 0
}
