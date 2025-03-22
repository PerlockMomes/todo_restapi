package services

import (
	"errors"
	"time"

	"todo_restapi/internal/constants"
)

func IsDate(searchQuery string) (string, error) {

	isTime, err := time.Parse("02.01.2006", searchQuery)
	if err != nil {
		return "", errors.New("invalid date format")
	}

	return isTime.Format(constants.DateFormat), nil
}
