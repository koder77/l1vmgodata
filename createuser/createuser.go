// createuser.go - database in go
/*
 * This file createuser.go is part of L1VMgodata.
 *
 * (c) Copyright Stefan Pietzonke (jay-t@gmx.net), 2024
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
	"os"
	"crypto/sha256"
	"fmt"
)

func main() {
	var salt string = ""

	fmt.Println("createuser <username> <password>")

	fmt.Println("args: ", len(os.Args))

	// check error case:
	if len(os.Args) <= 2 {
		fmt.Println("Arguments error! Need username and password!")
		os.Exit(1)
	}

	// init random number generator
	randomSeed()
	salt = randomString (64)

	user := os.Args[1]
	password := os.Args[2]

	password = password + salt

	// create hash
	password_string := fmt.Sprintf("%s", sha256.Sum256([]byte(password)))

    fmt.Println("insert this into the users.config file:")
	fmt.Printf("%v, %x\n", user, password_string)
	fmt.Printf("%v, %s\n", user, salt)

	os.Exit(0)
}