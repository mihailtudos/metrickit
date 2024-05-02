package flags

import (
	"fmt"
	"strconv"
	"strings"
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
	amountOfSeconds, ok := strings.CutSuffix(flagsValue, "s")
	if df.DurationType == time.Second && !ok {
		return fmt.Errorf("invalid interval format")
	}

	interval, err := strconv.Atoi(amountOfSeconds)
	if err != nil || interval < 0 {
		return fmt.Errorf("invalid interval value")
	}

	df.Length = interval
	return nil
}

func (df *DurationFlag) GetDuration() *time.Duration {
	val := df.DurationType * time.Duration(df.Length)
	return &val
}
