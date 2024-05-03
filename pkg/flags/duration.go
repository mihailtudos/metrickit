package flags

import (
	"fmt"
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
	return fmt.Sprintf("%v", df.GetDuration())
}

func (df *DurationFlag) Set(flagsValue string) error {
	interval, err := strconv.Atoi(flagsValue)
	if err != nil {
		return fmt.Errorf("invalid interval specified")
	}

	df.Length = interval
	return nil
}

func (df *DurationFlag) GetDuration() *time.Duration {
	val := df.DurationType * time.Duration(df.Length)
	return &val
}
