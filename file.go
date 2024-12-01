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
	"crypto/sha256"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func save_data(file_path string) int {
	var i uint64 = 0
	var l uint64 = 0
	var linkslen uint64 = 0

	os.Remove(file_path)
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
			linkslen = uint64(len((*pdata)[i].links))
			_, err = f.WriteString(":link" + " \"" + strconv.FormatInt(int64(linkslen), 10) + "\"\n")
			dmutex.Unlock()
			if err != nil {
				fmt.Println("Error writing database file:", err.Error())
				return 1
			}

			// save links
			if linkslen > 0 {
				for l = 0; l < linkslen; l++ {
					dmutex.Lock()
					_, err = f.WriteString(":link" + " \"" + strconv.FormatInt(int64((*pdata)[i].links[l]), 10) + "\"\n")
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
	var linkslen uint64 = 0
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

	//fmt.Println("DEBUG: load: i start =", i)

	// read and check header
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		//fmt.Println("DEBUG: i:", i, " line:", line)

		if i < maxdata {
			// enough data memory, load data
			if header_line == 0 {
				if line != "l1vmgodata database" {
					fmt.Println("Error opening database file: " + file_path + " not a l1vmgodata database!")
					return 1
				}
				header_line = 1
			} else {
				//fmt.Println("load_data: '" + line + "'\n")
				key, value = split_data(line)

				//fmt.Println("load_data: key: '" + key + "' value: '" + value + "'\n\n")

				if key != "" && key != "link" {
					// store data
					dmutex.Lock()
					(*pdata)[i].used = true
					(*pdata)[i].key = key
					(*pdata)[i].value = value
					dmutex.Unlock()
				}

				if key == "link" {
					// get links number

					linkslen, _ = strconv.ParseUint(value, 10, 64)

					//fmt.Printf ("load: links: %d\n", linkslen)

					if linkslen > 0 {
						// there are links, load them
						for l = 0; l < linkslen; l++ {
							scanner.Scan()
							line := scanner.Text()
							key, value = split_data(line)
							link, _ = strconv.ParseUint(value, 10, 64)

							//fmt.Printf("load: link: %d\n", link)

							dmutex.Lock()
							(*pdata)[i].links = append((*pdata)[i].links, link)
							dmutex.Unlock()

							// DEBUG
							fmt.Println("got link\n")
						}
					}
					i++
				}
			}
		} else {
			fmt.Println("Error reading database: out of memory: entries overflow!")
			fmt.Println("Failed to load index:", i, "into maxdata:", maxdata)
			return 1
		}
	}
	data_index = i
	free_index = i // set next free index
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
		} else {
			fmt.Println("Error reading database: out of memory: entries overflow!")
			return 1
		}

	}
	data_index = i
	return 0
}

// export CSV table
func save_data_table_csv(file_path string) int {
	// save CSV table in the format of:
	// :1-1-substance "water"
	// :link '0'
	// :1-2-chemical "H2O"
	// :link '0'
	// :1-3-boiling "100"
	// :link '0'
	// :2-1-substance "iron"
	// :link '0'
	// :2-2-chemical "Fe"
	// :link '0'
	// :2-3-boiling "3070"
	// :link '0'
	//
	// is saved as:
	// substance, chemical, boiling
	// water, H2O, 100
	// iron, Fe, 3070

	var key uint64 = 1
	var keymax uint64 = 0
	var index uint64 = 1
	var keystr string = ""
	var valuestr string = ""
	var search bool = true
	var save bool = true
	// create file
	f, err := os.Create(file_path)
	if err != nil {
		fmt.Println("Error opening database file: " + file_path + err.Error())
		return 1
	}
	// remember to close the file
	defer f.Close()

	// get start key
	for search {
		keystr, valuestr = get_table_key(index, key)

		// DEBUG
		//fmt.Printf("save tavble csv: index: %d, key: %d\n", index, key )
		//fmt.Println("save table csv: keystr: '" + keystr + "' valuestr: '" + valuestr + "’")

		if keystr != "" {
			if key > 1 {
				_, err = f.WriteString(", ")
				if err != nil {
					fmt.Println("Error writing database file:", err.Error())
					return 1
				}
			}

			_, err = f.WriteString(keystr)
			if err != nil {
				fmt.Println("Error writing database file:", err.Error())
				return 1
			}
			key = key + 1
		} else {
			search = false
		}
	}
	_, err = f.WriteString("\n")
	if err != nil {
		fmt.Println("Error writing database file:", err.Error())
		return 1
	}

	keymax = key

	index = 1
	for save {
		for key = 1; key <= keymax; key++ {
			keystr, valuestr = get_table_key(index, key)
			if keystr != "" {
				if key > 1 {
					_, err = f.WriteString(", ")
					if err != nil {
						fmt.Println("Error writing database file:", err.Error())
						return 1
					}
				}

				_, err = f.WriteString(valuestr)
				if err != nil {
					fmt.Println("Error writing database file:", err.Error())
					return 1
				}
			} else {
				if key == 1 {
					// no new data entry, exit save loop!
					save = false
				}
			}
		}
		_, err = f.WriteString("\n")
		if err != nil {
			fmt.Println("Error writing database file:", err.Error())
			return 1
		}

		index++
	}

	return 0
}

// import CSV table
func load_data_table_csv(file_path string) int {
	var key uint64 = 1
	var index uint64 = 1
	var keyfullstr string = ""
	var indexstr string = ""
	var keystr string = ""
	var keys_headerstr string = ""
	var key_headerstr string = ""
	var valuestr string = ""
	var key_line = true
	var i uint64
	var value_start int = 0
	var value_next int = 0
	//var value_comma_pos int
	var header_start int = 0
	var header_next int = 0
	//var header_comma_pos int
	var do_split bool = true

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

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if key_line {
			// get key names
			key_line = false
			keys_headerstr = line

			//fmt.Println("csv table import: key header: " + keys_headerstr)
		} else {
			key = 1
			do_split = true
			for do_split {
				key_headerstr, _, header_next = split_data_csv_table(keys_headerstr, header_start)

				//fmt.Println("csv table import: key: " + key_headerstr)

				if key_headerstr == "" {
					// end of header, reset
					header_start = 0
					value_start = 0
					do_split = false
					continue
				}

				indexstr = strconv.FormatUint(index, 10)
				keystr = strconv.FormatUint(key, 10)
				keyfullstr = indexstr + "-" + keystr + "-" + key_headerstr
				//fmt.Println("csv table import: key full: " + keyfullstr)

				valuestr, _, value_next = split_data_csv_table(line, value_start)
				if valuestr == "" {
					// end of header, reset
					value_start = 0
					do_split = false
					continue
				}

				//fmt.Println("csv table import: value: " + valuestr)

				if i < maxdata {
					dmutex.Lock()
					(*pdata)[i].used = true
					(*pdata)[i].key = keyfullstr
					(*pdata)[i].value = valuestr
					data_index = i
					i++
					dmutex.Unlock()
				} else {
					fmt.Println("Error reading database: out of memory: entries overflow!")
					return 1
				}

				// set next data position
				header_start = header_next
				value_start = value_next

				key++
			}
			index++
		}
	}
	return 0
}

// check user name in password file
func check_user(file_path string, user string, password string) int {
	// load database file
	var user_list string
	var password_hash string
	var password_salt string
	var salt string = ""
	var parse_loop = 0

	file, err := os.Open(file_path)
	if err != nil {
		fmt.Println("Error opening user file: " + file_path + " " + err.Error())
		return 1
	}
	// remember to close the file
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if parse_loop == 0 {
			// get first line data: username and password hash
			user_list, password_hash = split_data_csv(line)
			password_hash = fmt.Sprintf("%s", password_hash[1:])
			parse_loop = 1
		} else {
			// read second line: salt
			line = scanner.Text()
			_, salt = split_data_csv(line)

			salt = fmt.Sprintf("%s", salt[1:])
			password_salt = password + salt

			if user_list == user {
				// found valid user name, check for password match
				password_file_string := fmt.Sprintf("%x", sha256.Sum256([]byte(password_salt)))

				comp := strings.Compare(password_hash, password_file_string)
				if comp == 0 {
					return 0
				}
			}
			parse_loop = 0
		}
	}
	// user and password don't match
	return 1
}
