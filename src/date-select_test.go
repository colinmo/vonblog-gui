package main

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestDateSelector_GetMonthDays(t *testing.T) {
	oct, _ := time.Parse("2006-01-02", "2023-10-01")
	octDates := []string{}
	for i := 1; i < 32; i++ {
		octDates = append(octDates, fmt.Sprintf("%02d", i))
	}
	feb1, _ := time.Parse("2006-01-02", "2023-02-01")
	feb1Dates := []string{}
	for i := 1; i < 29; i++ {
		feb1Dates = append(feb1Dates, fmt.Sprintf("%02d", i))
	}
	feb2, _ := time.Parse("2006-01-02", "2020-02-01")
	feb2Dates := []string{}
	for i := 1; i < 30; i++ {
		feb2Dates = append(feb2Dates, fmt.Sprintf("%02d", i))
	}
	tests := []struct {
		name string
		d    *DateSelector
		want []string
	}{
		// TODO: Add test cases.
		{
			"Basic",
			MakeDateWithDate(oct),
			octDates,
		},
		{
			"Feb",
			MakeDateWithDate(feb1),
			feb1Dates,
		},
		{
			"FebLeap",
			MakeDateWithDate(feb2),
			feb2Dates,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.GetMonthDays(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DateSelector.GetMonthDays() = %v, want %v", got, tt.want)
			}
		})
	}
}
