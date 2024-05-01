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

func (sa *DurationFlag) String() string {
	return fmt.Sprintf("%v", sa.GetDuration())
}

func (a *DurationFlag) Set(flagsValue string) error {
	interval, err := strconv.Atoi(flagsValue)
	if err != nil {
		return fmt.Errorf("invalid interval specified")
	}

	a.Length = interval
	return nil
}

func (a *DurationFlag) GetDuration() time.Duration {
	return a.DurationType * time.Duration(a.Length)
}
