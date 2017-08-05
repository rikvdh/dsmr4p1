// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dsmr4p1

const (
	poly = 0xA001
)

// Table is a 256-word table representing the polynomial for efficient processing.
type Table struct {
	entries  [256]uint16
	reversed bool
}

var crcTable = makeTable(poly)

func makeTable(poly uint16) *Table {
	t := &Table{
		reversed: false,
	}
	for i := 0; i < 256; i++ {
		crc := uint16(i)
		for j := 0; j < 8; j++ {
			if crc&1 == 1 {
				crc = (crc >> 1) ^ poly
			} else {
				crc >>= 1
			}
		}
		t.entries[i] = crc
	}
	return t
}

func calcChecksum(data []byte) uint16 {
	var crc uint16
	for _, v := range data {
		crc = crcTable.entries[byte(crc)^v] ^ (crc >> 8)
	}

	return crc
}
