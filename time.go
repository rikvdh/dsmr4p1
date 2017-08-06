package dsmr4p1

import (
	"errors"
	"strings"
	"time"
)

// Timestamp is the timestamp format used in the dutch smartmeters. Do
// note this function assumes the CET/CEST timezone.
type Timestamp time.Time

func (t *Timestamp) UnmarshalJSON(b []byte) error {
	// The format for the timestamp is:
	// YYMMDDhhmmssX
	// The value used for X determines whether DST is active.
	// S (summer?) means yes, W (winter?) means no.
	timestamp := strings.Trim(string(b), "\"")
	var timezone string
	switch timestamp[len(timestamp)-1] {
	case 'S':
		timezone = summerTimezone
	case 'W':
		timezone = winterTimezone
	default:
		return errors.New("parsing timestamp: missing DST indicator")
	}

	// To make sure parsing is always consistent and indepentent of the the local
	// timezone of the host this code is running on, let's for now assume Dutch
	// time.
	loc, _ := time.LoadLocation("Europe/Amsterdam")

	ts, err := time.ParseInLocation("060102150405 MST", timestamp[:len(timestamp)-1]+" "+timezone, loc)
	if err != nil {
		return err
	}
	*t = Timestamp(ts)
	return nil
}
