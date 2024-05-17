package flags

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDurationFlagCanBeCreatedWithDifferentTimeValues(t *testing.T) {
	tests := []struct {
		name      string
		value     int
		valueType time.Duration
		want      *DurationFlag
	}{
		{
			name:      "seconds based duration flags can be created",
			value:     10,
			valueType: time.Second,
			want:      &DurationFlag{DurationType: time.Second, Length: 10},
		},
		{
			name:      "minutes based duration flags can be created",
			value:     10,
			valueType: time.Minute,
			want:      &DurationFlag{DurationType: time.Minute, Length: 10},
		},
		{
			name:      "0 value flags can be created",
			value:     0,
			valueType: time.Minute,
			want:      &DurationFlag{DurationType: time.Minute, Length: 0},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			df := NewDurationFlag(test.valueType, test.value)
			assert.Equal(t, df.GetDuration(), test.want.GetDuration())
		})
	}
}

func TestDurationFlagCanBeSet(t *testing.T) {
	tests := []struct {
		name         string
		value        string
		valueType    time.Duration
		want         *DurationFlag
		invalidInput bool
	}{
		{
			name:         "seconds based duration flags can be set",
			value:        "10",
			valueType:    time.Second,
			want:         &DurationFlag{DurationType: time.Second, Length: 10},
			invalidInput: true,
		},
		{
			name:         "minutes based duration flags can be set",
			value:        "10",
			valueType:    time.Minute,
			want:         &DurationFlag{DurationType: time.Minute, Length: 10},
			invalidInput: true,
		},
		{
			name:         "0 value flags can be set",
			value:        "0",
			valueType:    time.Minute,
			want:         &DurationFlag{DurationType: time.Minute, Length: 0},
			invalidInput: true,
		},
		{
			name:         "set duration flag with invalid input",
			value:        "no valid input",
			valueType:    time.Minute,
			want:         &DurationFlag{DurationType: time.Minute, Length: 0},
			invalidInput: false,
		},
		{
			name:         "set duration flag with negative value",
			value:        "-1",
			valueType:    time.Second,
			want:         &DurationFlag{DurationType: time.Second, Length: 0},
			invalidInput: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			df := &DurationFlag{Length: 0, DurationType: test.valueType}
			err := df.Set(test.value)
			if test.invalidInput {
				require.NoError(t, err)
				assert.Equal(t, df.GetDuration(), test.want.GetDuration())
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDurationFlagStringFormat(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		valueType time.Duration
		want      string
	}{
		{
			name:      "seconds based duration flags are formatted correctly",
			value:     "10",
			valueType: time.Second,
			want:      "10s",
		},
		{
			name:      "minutes based duration flags are formatted correctly",
			value:     "10",
			valueType: time.Minute,
			want:      "10m0s",
		},
		{
			name:      "0 value flags are formatted correctly",
			value:     "0",
			valueType: time.Second,
			want:      "0s",
		},
		{
			name:      "0 value flags are formatted correctly",
			value:     "-4",
			valueType: time.Second,
			want:      "0s",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			df := NewDurationFlag(test.valueType, 0)
			_ = df.Set(test.value)
			assert.Equal(t, df.String(), test.want)
		})
	}
}
