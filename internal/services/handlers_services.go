package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"todo_restapi/internal/constants"
	"todo_restapi/internal/models"
)

func NextDate(now time.Time, date string, repeat string) (string, error) {

	if repeat == "" {
		return "", errors.New("repeat cannot be empty")
	}

	dateParse, err := time.Parse(constants.DateFormat, date)
	if err != nil {
		return "", fmt.Errorf("date parse error: %w", err)
	}

	repeatType, firstRepeatPattern, _ := parseRepeat(repeat)

	switch repeatType {

	case "d":

		if len(firstRepeatPattern) == 0 {
			return "", errors.New("\"d\" parameter is empty")
		}

		if firstRepeatPattern[0] > 400 {
			return "", errors.New("invalid \"d\" value (400 is max)")
		}

		dateParse = dateParse.AddDate(0, 0, firstRepeatPattern[0])
		for dateParse.Before(now) {
			dateParse = dateParse.AddDate(0, 0, firstRepeatPattern[0])
		}

		return dateParse.Format(constants.DateFormat), nil

	case "y":

		dateParse = dateParse.AddDate(1, 0, 0)
		for dateParse.Before(now) {
			dateParse = dateParse.AddDate(1, 0, 0)
		}

		return dateParse.Format(constants.DateFormat), nil

	default:
		return "", errors.New("invalid repeat value")
	}
}

func parseNumbers(input string) []int {

	stringNums := strings.Split(input, ",")
	output := make([]int, 0, len(stringNums))

	for _, num := range stringNums {
		intNum, err := strconv.Atoi(num)
		if err != nil {
			fmt.Println("parseNumbers: string to int conversion error")
		}
		output = append(output, intNum)
	}
	return output
}

func parseRepeat(repeat string) (string, []int, []int) {

	var firstRepeatPattern []int
	var secondRepeatPattern []int

	repeatParse := strings.Fields(repeat)

	if len(repeatParse) > 1 {
		firstRepeatPattern = parseNumbers(repeatParse[1])
	}

	if len(repeatParse) > 2 {
		secondRepeatPattern = parseNumbers(repeatParse[2])
	}

	repeatType := repeatParse[0]

	return repeatType, firstRepeatPattern, secondRepeatPattern
}

func ValidateTaskRequest(newTask *models.Task, now string) error {

	if newTask.Title == "" {
		return errors.New("title is empty")
	}

	if newTask.Date == "" {
		newTask.Date = now
	}

	_, err := time.Parse(constants.DateFormat, newTask.Date)
	if err != nil {
		return errors.New("invalid date format")
	}

	if newTask.Date < now {
		if newTask.Repeat == "" {
			newTask.Date = now
		} else {
			dateCalculation, err := NextDate(time.Now(), newTask.Date, newTask.Repeat)
			if err != nil {
				return fmt.Errorf("NextDate: function error: %w", err)
			}
			newTask.Date = dateCalculation
		}
	}
	return nil
}

func WriteJSONError(write http.ResponseWriter, statusCode int, errMsg string) {

	log.Println(errMsg)

	write.Header().Set("Content-Type", "application/json")
	write.WriteHeader(statusCode)

	response := map[string]string{"error": errMsg}

	if err := json.NewEncoder(write).Encode(response); err != nil {
		log.Printf("failed to encode JSON response: %v", err)
	}
}
