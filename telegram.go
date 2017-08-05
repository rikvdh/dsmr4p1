package dsmr4p1

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Telegram holds the a P1 telegram. It is essentially a slice of bytes.
type Telegram struct {
	data  []byte
	lines map[string][]string

	Identifier string

	// Data stored in database for every reading.
	Timestamp                     string `json:",omitempty" p1:"0-0:1.0.0"`
	ElectricityEquipmentID        string `json:",omitempty" p1:"0-0:96.1.1"`
	ElectricityDelivered1         string `json:",omitempty" p1:"1-0:1.8.1"`
	ElectricityReturned1          string `json:",omitempty" p1:"1-0:2.8.1"`
	ElectricityDelivered2         string `json:",omitempty" p1:"1-0:1.8.2"`
	ElectricityReturned2          string `json:",omitempty" p1:"1-0:2.8.2"`
	ElectricityCurrentlyDelivered string `json:",omitempty" p1:"1-0:1.7.0"`
	ElectricityCurrentlyReturned  string `json:",omitempty" p1:"1-0:2.7.0"`

	PhaseCurrentlyDeliveredL1 string `json:",omitempty" p1:"1-0:21.7.0"`
	PhaseCurrentlyDeliveredL2 string `json:",omitempty" p1:"1-0:41.7.0"`
	PhaseCurrentlyDeliveredL3 string `json:",omitempty" p1:"1-0:61.7.0"`

	// Static data, stored in database but only record of the last reading is preserved.
	DsmrVersion                string `json:",omitempty" p1:"1-3:0.2.8"`
	ElectricityTariff          string `json:",omitempty" p1:"0-0:96.14.0"`
	PowerFailureCount          string `json:",omitempty" p1:"0-0:96.7.21"`
	LongPowerFailureCount      string `json:",omitempty" p1:"0-0:96.7.9"`
	InstantaneousCurrentL1     string `json:",omitempty" p1:"1-0:31.7.0"`
	InstantaneousCurrentL2     string `json:",omitempty" p1:"1-0:51.7.0"`
	InstantaneousCurrentL3     string `json:",omitempty" p1:"1-0:71.7.0"`
	InstantaneousActivePowerL1 string `json:",omitempty" p1:"1-0:22.7.0"`
	InstantaneousActivePowerL2 string `json:",omitempty" p1:"1-0:42.7.0"`
	InstantaneousActivePowerL3 string `json:",omitempty" p1:"1-0:62.7.0"`
	VoltageSagCountL1          string `json:",omitempty" p1:"1-0:32.32.0"`
	VoltageSagCountL2          string `json:",omitempty" p1:"1-0:52.32.0"`
	VoltageSagCountL3          string `json:",omitempty" p1:"1-0:72.32.0"`
	VoltageSwellCountL1        string `json:",omitempty" p1:"1-0:32.36.0"`
	VoltageSwellCountL2        string `json:",omitempty" p1:"1-0:52.36.0"`
	VoltageSwellCountL3        string `json:",omitempty" p1:"1-0:72.36.0"`

	// Gas meter information
	GasEquipmentID string `json:",omitempty" p1:"0-1:96.1.0"`
	GasTimeValue   string `json:",omitempty" p1:"0-1:24.2.1"`
}

// Parse attempts to parse the telegram. It returns a map of strings to string
// slices. The keys in the map are the ID-codes, the strings in the slice are
// are the value between brackets for that ID-code.
func (t *Telegram) parse() error {
	// Parse the telegram in a relatively naive way. Of course this
	// is not properly langsec approved :)

	lines := strings.Split(string(t.data), "\r\n")

	if len(lines) < 2 {
		return errors.New("unexpected too few line in telegram")
	}

	// Some additional checks
	if lines[0][0] != '/' {
		return errors.New("expected '/' missing in first line of telegram")
	}
	if len(lines[1]) != 0 {
		return errors.New("missing separating new line (CR+LF) between identifier and data in telegram")
	}

	// According to the documentation, the telegram starts with:
	// "/XXXZ Ident CR LF CR LF", followed by the data.
	i := bytes.Index(t.data, []byte("\r\n\r\n"))
	t.Identifier = string(t.data[5:i])

	for i, l := range lines[2 : len(lines)-1] {
		found := false
		idCodeEnd := strings.Index(l, "(")
		if idCodeEnd == -1 {
			return errors.New("expected '(', not found on line" + string(i))
		}

		idCode := l[:idCodeEnd]
		ty := reflect.TypeOf(*t)
		for i := 0; i < ty.NumField(); i++ {
			f := ty.Field(i)
			if f.Tag.Get("p1") == idCode {
				reflect.ValueOf(t).Elem().FieldByName(f.Name).SetString(l[idCodeEnd+1 : len(l)-1])
				found = true
			}
		}
		if !found {
			fmt.Printf("Skipping: %s\n", l)
		}
	}
	//t.lines = result
	return nil
}
