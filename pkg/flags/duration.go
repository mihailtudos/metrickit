package flags

import (
	"errors"
	"strconv"
	"time"
)

type DurationFlag struct {
	DurationType time.Duration
	Length       int
}

func NewDurationFlag(durationType time.Duration, length int) *DurationFlag {
	return &DurationFlag{
		DurationType: durationType,
		Length:       length,
	}
}

func (df *DurationFlag) String() string {
	return df.GetDuration().String()
}

func (df *DurationFlag) Set(flagsValue string) error {
	interval, err := strconv.Atoi(flagsValue)
	if err != nil || interval < 0 {
		return errors.New("invalid interval specified")
	}

	df.Length = interval
	return nil
}

func (df *DurationFlag) GetDuration() time.Duration {
	return df.DurationType * time.Duration(df.Length)
}
