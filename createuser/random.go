// random.go - database in go
/*
 * This file random.go is part of L1VMgodata.
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
    crypto_rand "crypto/rand"
    "encoding/binary"
    math_rand "math/rand"
    "strings"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-+*/[]{}~&#><=|"

func randomString(n int) string {
    sb := strings.Builder{}
    sb.Grow(n)
    for i := 0; i < n; i++ {
        sb.WriteByte(charset[math_rand.Intn(len(charset))])
    }
    return sb.String()
}

func randomSeed() {
    var b [8]byte
    _, err := crypto_rand.Read(b[:])
    if err != nil {
        panic("cannot seed math/rand package with cryptographically secure random number generator")
    }
    math_rand.Seed(int64(binary.LittleEndian.Uint64(b[:])))
}
