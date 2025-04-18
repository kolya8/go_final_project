package api

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

const dateFormat = "20060102"

type repeatData struct {
	key   string
	day   []int
	month []int
}

func nextDateHandler(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	repeat := r.FormValue("repeat")
	if strings.TrimSpace(repeat) == "" {
		http.Error(w, "repeat parameter is required", http.StatusBadRequest)
		return
	}

	date := r.FormValue("date")

	now, err := readDateTypeParam(r, "now")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	result, err := NextDate(now, date, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
}

func readDateTypeParam(r *http.Request, p string) (time.Time, error) {
	input := r.FormValue(p)

	if input == "" && p == "now" {
		return time.Now(), nil
	}

	date, err := time.Parse(dateFormat, input)
	if err != nil {
		return date, fmt.Errorf("%s invalid format", p)
	}
	return date, nil
}

func repeatParse(input string, result *repeatData) error {
	inputData := strings.Split(input, " ")

	for i, value := range inputData {
		switch i {
		case 0:
			result.key = value
		case 1:
			var err error
			result.day, err = repeatParseDaysMonths(value)
			if err != nil {
				return err
			}
		case 2:
			var err error
			result.month, err = repeatParseDaysMonths(value)
			if err != nil {
				return err
			}
		}
	}

	sort.Ints(result.day)
	sort.Ints(result.month)

	return nil
}

func repeatParseDaysMonths(value string) ([]int, error) {
	result := make([]int, 0, len(value))
	for _, v := range strings.Split(value, ",") {
		if num, err := strconv.Atoi(v); err == nil {
			result = append(result, num)
		} else {
			return result, errors.New("repeat parameter invalid format")
		}
	}
	return result, nil
}

func findNextYear(targetDate, now time.Time) time.Time {
	for {
		targetDate = targetDate.AddDate(1, 0, 0)
		if afterNow(targetDate, now) {
			break
		}
	}
	return targetDate
}

func findNextDay(targetDate, now time.Time, day int) time.Time {
	for {
		targetDate = targetDate.AddDate(0, 0, day)
		if afterNow(targetDate, now) {
			break
		}
	}
	return targetDate
}

func handleMonthlyRepeat(date, now time.Time, repeat repeatData) time.Time {
	if len(repeat.month) == 0 {
		return handleDaysInAllMonths(date, now, repeat)
	}
	return handleDaysInSpecificMonths(date, now, repeat)
}

func handlePosDayValue(date, now time.Time, days []int) time.Time {
	var found bool

	firstDayOfMonth := firstDayOfMonth(date)
	for !found {
		for _, dateValue := range days {
			date = firstDayOfMonth.AddDate(0, 0, dateValue-1)
			if date.Day() != dateValue {
				continue
			}
			if afterNow(date, now) {
				found = true
				break
			}
		}
		firstDayOfMonth = firstDayOfMonth.AddDate(0, 1, 0)
	}
	return date
}

func handleNegDayValue(date, now time.Time, days []int) time.Time {
	var found bool

	lastDayOfMonth := lastDayOfMonth(date)
	for !found {
		for _, dateValue := range days {
			date = lastDayOfMonth.AddDate(0, 0, dateValue+1)
			if afterNow(date, now) {
				found = true
				break
			}
		}
		lastDayOfMonth = lastDayOfNextMonth(lastDayOfMonth)
	}
	return date
}

func handleDaysInAllMonths(date, now time.Time, repeat repeatData) time.Time {

	neg, pos := splitNegativesPositive(repeat.day)

	if len(pos) > 0 && len(neg) == 0 {
		date = handlePosDayValue(date, now, repeat.day)
	}

	if len(neg) > 0 && len(pos) == 0 {
		date = handleNegDayValue(date, now, repeat.day)
	}

	if len(pos) > 0 && len(neg) > 0 {
		dateNeg := handlePosDayValue(date, now, pos)
		datePos := handleNegDayValue(date, now, neg)
		if dateNeg.Before(datePos) {
			date = dateNeg
		} else {
			date = datePos
		}
	}

	return date
}

func splitNegativesPositive(days []int) (neg, pos []int) {
	for _, d := range days {
		if d < 0 {
			neg = append(neg, d)
		} else {
			pos = append(pos, d)
		}
	}
	return
}

func firstDayOfMonth(date time.Time) time.Time {
	firstDay := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	return firstDay
}

func lastDayOfMonth(date time.Time) time.Time {
	lastDay := time.Date(date.Year(), date.Month()+1, 0, 0, 0, 0, 0, date.Location())
	return lastDay
}

func handleDaysInSpecificMonths(date, now time.Time, repeat repeatData) time.Time {
	months := make(map[int]bool)
	for _, m := range repeat.month {
		months[m] = true
	}

	firstDayOfMonth := firstDayOfMonth(date)

	for {
		if _, exists := months[int(firstDayOfMonth.Month())]; !exists {
			firstDayOfMonth = firstDayOfMonth.AddDate(0, 1, 0)
			continue
		}

		for _, day := range repeat.day {
			targetDate := firstDayOfMonth.AddDate(0, 0, day-1)
			if targetDate.Day() == day && afterNow(targetDate, now) {
				return targetDate
			}
		}

		firstDayOfMonth = firstDayOfMonth.AddDate(0, 1, 0)
	}
}

func handleWeeklyRepeat(date, now time.Time, repeat repeatData) time.Time {
	for {
		for _, d := range repeat.day {
			daysUntilNext := (d - int(date.Weekday()) + 7) % 7
			targetDate := date.AddDate(0, 0, daysUntilNext)
			if afterNow(targetDate, now) {
				return targetDate
			}
		}
		date = date.AddDate(0, 0, 7)
	}
}

func NextDate(now time.Time, dstart string, repeat string) (string, error) {

	repeatData := repeatData{}
	err := repeatParse(repeat, &repeatData)
	if err != nil {
		return "", err
	}

	if err := validRepeat(&repeatData); err != nil {
		return "", err
	}

	date, err := time.Parse(dateFormat, dstart)
	if err != nil {
		return "", err
	}

	switch repeatData.key {
	case "y":
		date = findNextYear(date, now)
	case "d":
		date = findNextDay(date, now, repeatData.day[0])
	case "m":
		date = handleMonthlyRepeat(date, now, repeatData)
	case "w":
		date = handleWeeklyRepeat(date, now, repeatData)
	}

	return date.Format(dateFormat), nil
}

func lastDayOfNextMonth(d time.Time) time.Time {
	year := d.Year()
	month := d.Month()

	month++
	if month > 12 {
		month = 1
		year++
	}

	return time.Date(year, month+1, 0, 0, 0, 0, 0, d.Location())
}

func afterNow(date, now time.Time) bool {
	return date.After(now)
}

func validRepeat(r *repeatData) error {

	validValues := map[string]bool{
		"y": true,
		"m": true,
		"d": true,
		"w": true,
	}

	if !validValues[r.key] {
		return errors.New("repeat parameter invalid character")
	}

	if r.key != "y" && len(r.day) == 0 {
		return errors.New("repeat parameter no days interval")
	}

	if r.key == "d" {
		for _, v := range r.day {
			if v > 400 {
				return errors.New("repeat parameter interval exceeded")
			}
		}
	}

	if r.key != "d" {
		for _, v := range r.day {
			if v > 31 {
				return errors.New("repeat parameter invalid day of the month")
			}
		}
	}

	if r.key == "w" {
		for _, v := range r.day {
			if v > 7 {
				return errors.New("repeat parameter invalid day of the week")
			}
		}
	}

	if r.key == "m" {
		for _, v := range r.day {
			if v < -2 {
				return errors.New("repeat parameter invalid day of the month")
			}
		}
	}

	for _, v := range r.month {
		if v > 12 {
			return errors.New("repeat parameter invalid month")
		}
	}

	return nil
}