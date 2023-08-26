package main

import (
	"fmt"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type DateSelector struct {
	Month      *widget.Select
	Date       *widget.Select
	Year       *widget.Entry
	Hour       *widget.Select
	Minute     *widget.Select
	ActualDate time.Time
}

func MakeDateWithDate(when time.Time) *DateSelector {
	d := DateSelector{ActualDate: when}
	return &d
}

func MakeDateWithDateAndWidget(when time.Time) *DateSelector {
	d := DateSelector{ActualDate: when}
	d.CreateWidget()
	return &d
}

func (d *DateSelector) GetMonthDays() []string {
	tempDate := time.Date(
		d.ActualDate.Year(),
		d.ActualDate.Month()+1,
		0,
		d.ActualDate.Hour(),
		d.ActualDate.Minute(),
		0,
		0,
		time.Local,
	)
	lastDay := tempDate.Day() + 1
	options := []string{}
	for i := 1; i < lastDay; i++ {
		options = append(options, fmt.Sprintf("%02d", i))
	}
	return options
}

func (d *DateSelector) CreateWidget() *fyne.Container {
	months := map[string]int{
		"Jan": 1, "Feb": 2, "Mar": 3, "Apr": 4, "May": 5, "Jun": 6,
		"Jul": 7, "Aug": 8, "Sep": 9, "Oct": 10, "Nov": 11, "Dec": 12,
	}
	d.Date = widget.NewSelect(
		d.GetMonthDays(),
		func(date string) {
			dt, _ := strconv.Atoi(date)
			d.ActualDate = time.Date(
				d.ActualDate.Year(),
				d.ActualDate.Month(),
				dt,
				d.ActualDate.Hour(),
				d.ActualDate.Minute(),
				0,
				0,
				time.Local,
			)
		},
	)
	d.Date.SetSelected(fmt.Sprintf("%02d", d.ActualDate.Day()))
	d.Date.PlaceHolder = "--"
	d.Date.Refresh()
	allMonths := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	d.Month = widget.NewSelect(
		allMonths,
		func(month string) {
			d.ActualDate = time.Date(
				d.ActualDate.Year(),
				time.Month(months[month]),
				d.ActualDate.Day(),
				d.ActualDate.Hour(),
				d.ActualDate.Minute(),
				0,
				0,
				time.Local,
			)
			tempDate := time.Date(
				d.ActualDate.Year(),
				time.Month(months[month])+1,
				0,
				d.ActualDate.Hour(),
				d.ActualDate.Minute(),
				0,
				0,
				time.Local,
			)
			d.Date.Options = d.GetMonthDays()
			selectedDate, _ := strconv.Atoi(d.Date.Selected)
			if selectedDate < tempDate.Day()+1 {
				d.Date.SetSelected(fmt.Sprintf("%d", selectedDate))
			}
			d.Date.Refresh()
		},
	)
	d.Month.SetSelected(allMonths[d.ActualDate.Month()])
	d.Month.PlaceHolder = "---"
	d.Year = widget.NewEntry()
	d.Year.SetText(fmt.Sprintf("%d", d.ActualDate.Year()))
	d.Year.OnChanged = func(year string) {
		d.Month.OnChanged(d.Month.Selected)
		d.Date.OnChanged(d.Date.Selected)
	}
	d.Hour = widget.NewSelect(
		[]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23"},
		func(sel string) {},
	)
	d.Hour.PlaceHolder = "--"
	d.Hour.SetSelected(fmt.Sprintf("%d", d.ActualDate.Hour()))
	mins := []string{}
	for i := 0; i < 60; i++ {
		mins = append(mins, fmt.Sprintf("%02d", i))
	}
	d.Minute = widget.NewSelect(
		mins,
		func(sel string) {},
	)
	d.Minute.PlaceHolder = "--"
	d.Minute.SetSelected(fmt.Sprintf("%02d", d.ActualDate.Minute()))
	return container.NewVBox(
		container.NewHBox(
			d.Date,
			d.Month,
			d.Year,
		),
		container.NewHBox(
			d.Hour,
			widget.NewLabel(":"),
			d.Minute,
		),
	)
}

func (d *DateSelector) SetDate(when time.Time)  {}
func (d *DateSelector) GetDateAsString() string { return "" }
