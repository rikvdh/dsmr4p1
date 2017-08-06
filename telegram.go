package dsmr4p1

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Value holds a P1 value
type Value struct {
	Val  float64
	Unit string
}

func (v *Value) UnmarshalJSON(b []byte) error {
	var err error
	v.Val, v.Unit, err = parseValue(strings.Trim(string(b), "\""))
	return err
}

type GasMeterValue struct {
	Timestamp Timestamp
	Value     float64
	Unit      string
}

func (v *GasMeterValue) UnmarshalJSON(b []byte) error {
	rawVals := strings.Split(strings.Trim(string(b), "\""), ")(")
	if len(rawVals) != 2 {
		return fmt.Errorf("GasMeterValue didnt parse: %s", b)
	}
	if err := v.Timestamp.UnmarshalJSON([]byte(rawVals[0])); err != nil {
		return err
	}
	var err error
	v.Value, v.Unit, err = parseValue(rawVals[1])
	return err
}

// Telegram holds the a P1 telegram. It is essentially a slice of bytes.
type Telegram struct {
	Identifier string

	// Data stored in database for every reading.
	Timestamp                     Timestamp `json:"0-0:1.0.0,omitempty"`
	ElectricityEquipmentID        string    `json:"0-0:96.1.1,omitempty"`
	ElectricityDelivered1         Value     `json:"1-0:1.8.1,omitempty"`
	ElectricityReturned1          Value     `json:"1-0:2.8.1,omitempty"`
	ElectricityDelivered2         Value     `json:"1-0:1.8.2,omitempty"`
	ElectricityReturned2          Value     `json:"1-0:2.8.2,omitempty"`
	ElectricityCurrentlyDelivered Value     `json:"1-0:1.7.0,omitempty"`
	ElectricityCurrentlyReturned  Value     `json:"1-0:2.7.0,omitempty"`

	PhaseCurrentlyDeliveredL1 Value `json:"1-0:21.7.0,omitempty"`
	PhaseCurrentlyDeliveredL2 Value `json:"1-0:41.7.0,omitempty"`
	PhaseCurrentlyDeliveredL3 Value `json:"1-0:61.7.0,omitempty"`

	// Static data, stored in database but only record of the last reading is preserved.
	DsmrVersion                Value `json:"1-3:0.2.8,omitempty"`
	ElectricityTariff          Value `json:"0-0:96.14.0,omitempty"`
	PowerFailureCount          Value `json:"0-0:96.7.21,omitempty"`
	LongPowerFailureCount      Value `json:"0-0:96.7.9,omitempty"`
	InstantaneousCurrentL1     Value `json:"1-0:31.7.0,omitempty"`
	InstantaneousCurrentL2     Value `json:"1-0:51.7.0,omitempty"`
	InstantaneousCurrentL3     Value `json:"1-0:71.7.0,omitempty"`
	InstantaneousActivePowerL1 Value `json:"1-0:22.7.0,omitempty"`
	InstantaneousActivePowerL2 Value `json:"1-0:42.7.0,omitempty"`
	InstantaneousActivePowerL3 Value `json:"1-0:62.7.0,omitempty"`
	VoltageSagCountL1          Value `json:"1-0:32.32.0,omitempty"`
	VoltageSagCountL2          Value `json:"1-0:52.32.0,omitempty"`
	VoltageSagCountL3          Value `json:"1-0:72.32.0,omitempty"`
	VoltageSwellCountL1        Value `json:"1-0:32.36.0,omitempty"`
	VoltageSwellCountL2        Value `json:"1-0:52.36.0,omitempty"`
	VoltageSwellCountL3        Value `json:"1-0:72.36.0,omitempty"`

	// Gas meter information
	GasEquipmentID string        `json:"0-1:96.1.0,omitempty"`
	GasTimeValue   GasMeterValue `json:"0-1:24.2.1,omitempty"`
}

// Parse attempts to parse the telegram. It returns a map of strings to string
// slices. The keys in the map are the ID-codes, the strings in the slice are
// are the value between brackets for that ID-code.
func parseTelegram(data []byte) (*Telegram, error) {
	t := &Telegram{}
	// Parse the telegram in a relatively naive way. Of course this
	// is not properly langsec approved :)

	lines := strings.Split(string(data), "\r\n")

	if len(lines) < 2 {
		return nil, errors.New("unexpected too few line in telegram")
	}

	// Some additional checks
	if lines[0][0] != '/' {
		return nil, errors.New("expected '/' missing in first line of telegram")
	}
	if len(lines[1]) != 0 {
		return nil, errors.New("missing separating new line (CR+LF) between identifier and data in telegram")
	}

	// According to the documentation, the telegram starts with:
	// "/XXXZ Ident CR LF CR LF", followed by the data.
	i := bytes.Index(data, []byte("\r\n\r\n"))
	t.Identifier = string(data[5:i])

	values := map[string]string{}

	for i, l := range lines[2 : len(lines)-1] {
		idCodeEnd := strings.Index(l, "(")
		if idCodeEnd == -1 {
			return nil, errors.New("expected '(', not found on line" + string(i))
		}

		values[l[:idCodeEnd]] = l[idCodeEnd+1 : len(l)-1]
	}

	b, err := json.Marshal(values)
	if err != nil {
		fmt.Println(err)
	}

	if err := json.Unmarshal(b, t); err != nil {
		fmt.Println(err)
	}
	/*		for i := 0; i < ty.NumField(); i++ {
				f := ty.Field(i)
				if f.Tag.Get("p1") == idCode {
					fld := sv.FieldByName(f.Name)
					rawVal :=
					switch fld.Type().Name() {
					case "string":
						fld.SetString(rawVal)
					case "Time":
						tm := fld.Interface().(time.Time)
						tm, _ = parseTimestamp(rawVal)
						_ = tm
					case "float64":
						val, _, err := parseValue(rawVal)
						if err != nil {
							fmt.Println(f.Name, err)
							continue
						}
						fld.SetFloat(val)
					case "Value":
						val, unit, err := parseValue(rawVal)
						if err != nil {
							fmt.Println(f.Name, err)
							continue
						}
						fld.FieldByName("Val").SetFloat(val)
						fld.FieldByName("Unit").SetString(unit)
					case "GasMeterValue":

					default:
						fmt.Println(f.Name, fld.Type().Name(), rawVal)
					}
				}
			}
	*/
	return t, nil
}
