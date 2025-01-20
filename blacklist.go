// l1vmgodata.go - database in go
/*
 * This file blacklist.go is part of L1VMgodata.
 *
 * (c) Copyright Stefan Pietzonke (jay-t@gmx.net), 2025
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
		"bufio"
		"fmt"
		"os"
)

// config file
const (
	BLACKLIST = "blacklist.config"
)

func set_blacklist_ip(ip string) {
	dmutex.Lock()
	blacklist_ip = append (blacklist_ip, ip)
	blacklist_ip_ind++
	dmutex.Unlock()
}

func check_blacklist(ip string) bool {
	var i uint64

	if blacklist_ip_ind > 0 {
		dmutex.Lock()
		for i = 0; i < blacklist_ip_ind; i++ {
			if blacklist_ip[i] == ip {
				dmutex.Unlock()
				return true
			}
		}
		dmutex.Unlock()
	}

	return false
}

func read_ip_blacklist() bool {
	// load database file
	file, err := os.Open(BLACKLIST)
	if err != nil {
		fmt.Println("Error opening file: " + BLACKLIST + " " + err.Error())
		return false
	}
	// remember to close the file
	defer file.Close()

	// read one IP per line
	scanner := bufio.NewScanner(file)
	dmutex.Lock()
	for scanner.Scan() {
		line := scanner.Text()
		// store ip
		if len(line) >= 2 {
			blacklist_ip = append(blacklist_ip, line)
			fmt.Println("blacklist:", blacklist_ip[blacklist_ip_ind])
			blacklist_ip_ind++
		}
	}
	dmutex.Unlock()
	return true
}

func write_ip_blacklist() bool {
	var i uint64 = 0
	// save database file
	file, err :=  os.Create(BLACKLIST)
	if err != nil {
		fmt.Println("Error opening file: " + BLACKLIST + " " + err.Error())
		return false
	}
	// remember to close the file
	defer file.Close()

	dmutex.Lock()
	for i = 0; i < blacklist_ip_ind; i++ {
		_, err = file.WriteString(blacklist_ip[i] + "\n")
		if err != nil {
			fmt.Println("Error writing blacklist file:", err.Error())
		    dmutex.Unlock()
			return false
		}
	}
	dmutex.Unlock()
	return true
}
