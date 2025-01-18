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

func set_blacklist_ip(max_blacklist_ip uint64, ip string) {
	dmutex.Lock()
	if blacklist_ip_ind < max_blacklist_ip {
		blacklist_ip[blacklist_ip_ind] = ip
		blacklist_ip_ind++
	}
	dmutex.Unlock()
}

func check_blacklist(ip string) bool {
	var i uint64

	if blacklist_ip_ind > 0 {
		dmutex.Lock()
		for i = 0; i <= blacklist_ip_ind; i++ {
			if blacklist_ip[i] == ip {
				dmutex.Unlock()
				return true
			}
		}
		dmutex.Unlock()
	}

	return false
}
