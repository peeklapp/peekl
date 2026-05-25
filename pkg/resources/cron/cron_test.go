package cron

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type regexTestCase struct {
	Data   string
	Result bool
}

func TestValidNameRegex(t *testing.T) {
	testCases := []regexTestCase{
		{"valid_name-with-dash-123", true},
		{"invalid name because of spaces", false},
		{"invalid_name_because_of_+", false},
	}
	for _, testCase := range testCases {
		assert.Equalf(t, testCase.Result, validNameRegex.MatchString(testCase.Data), fmt.Sprintf("Test case: %+v", testCase.Data))
	}
}

func TestValidMinuteRegex(t *testing.T) {
	testCases := []regexTestCase{
		{"0", true},
		{"1", true},
		{"2", true},
		{"3", true},
		{"4", true},
		{"5", true},
		{"6", true},
		{"7", true},
		{"8", true},
		{"9", true},
		{"10", true},
		{"11", true},
		{"12", true},
		{"13", true},
		{"14", true},
		{"15", true},
		{"16", true},
		{"17", true},
		{"18", true},
		{"19", true},
		{"20", true},
		{"21", true},
		{"22", true},
		{"23", true},
		{"24", true},
		{"25", true},
		{"26", true},
		{"27", true},
		{"28", true},
		{"29", true},
		{"30", true},
		{"31", true},
		{"32", true},
		{"33", true},
		{"34", true},
		{"35", true},
		{"36", true},
		{"37", true},
		{"38", true},
		{"39", true},
		{"40", true},
		{"41", true},
		{"42", true},
		{"43", true},
		{"44", true},
		{"45", true},
		{"46", true},
		{"47", true},
		{"48", true},
		{"49", true},
		{"50", true},
		{"51", true},
		{"52", true},
		{"53", true},
		{"54", true},
		{"55", true},
		{"56", true},
		{"57", true},
		{"58", true},
		{"59", true},
		{"60", false},
	}
	for _, testCase := range testCases {
		assert.Equalf(t, testCase.Result, validMinuteRegex.MatchString(testCase.Data), fmt.Sprintf("Test case: %+v", testCase.Data))
	}
}

func TestValidMinuteRangeRegex(t *testing.T) {
	testCases := []regexTestCase{
		{"0-10", true},
		{"0-59", true},
		{"1-60", false},
		{"60-60", false},
	}
	for _, testCase := range testCases {
		assert.Equalf(t, testCase.Result, validMinuteRangeRegex.MatchString(testCase.Data), fmt.Sprintf("Test case: %+v", testCase.Data))
	}
}

func TestValidHourRegex(t *testing.T) {
	testCases := []regexTestCase{
		{"0", true},
		{"1", true},
		{"2", true},
		{"3", true},
		{"4", true},
		{"5", true},
		{"6", true},
		{"7", true},
		{"8", true},
		{"9", true},
		{"10", true},
		{"11", true},
		{"12", true},
		{"13", true},
		{"14", true},
		{"15", true},
		{"16", true},
		{"17", true},
		{"18", true},
		{"19", true},
		{"20", true},
		{"21", true},
		{"22", true},
		{"23", true},
		{"24", false},
	}
	for _, testCase := range testCases {
		assert.Equalf(t, testCase.Result, validHourRegex.MatchString(testCase.Data), fmt.Sprintf("Test case: %+v", testCase.Data))
	}
}

func TestValidHourRangeRegex(t *testing.T) {
	testCases := []regexTestCase{
		{"0-23", true},
		{"10-20", true},
		{"0-24", false},
		{"25-10", false},
	}
	for _, testCase := range testCases {
		assert.Equalf(t, testCase.Result, validHourRangeRegex.MatchString(testCase.Data), fmt.Sprintf("Test case: %+v", testCase.Data))
	}
}

func TestValidDayRegex(t *testing.T) {
	testCases := []regexTestCase{
		{"0", false},
		{"1", true},
		{"2", true},
		{"3", true},
		{"4", true},
		{"5", true},
		{"6", true},
		{"7", true},
		{"8", true},
		{"9", true},
		{"10", true},
		{"11", true},
		{"12", true},
		{"13", true},
		{"14", true},
		{"15", true},
		{"16", true},
		{"17", true},
		{"18", true},
		{"19", true},
		{"20", true},
		{"21", true},
		{"22", true},
		{"23", true},
		{"24", true},
		{"25", true},
		{"26", true},
		{"27", true},
		{"28", true},
		{"29", true},
		{"30", true},
		{"31", true},
		{"32", false},
		{"60", false},
	}
	for _, testCase := range testCases {
		assert.Equalf(t, testCase.Result, validDayRegex.MatchString(testCase.Data), fmt.Sprintf("Test case: %+v", testCase.Data))
	}
}

func TestValidDayRangeRegex(t *testing.T) {
	testCases := []regexTestCase{
		{"0-31", true},
		{"10-21", true},
		{"0-32", false},
		{"60-10", false},
	}
	for _, testCase := range testCases {
		assert.Equalf(t, testCase.Result, validDayRangeRegex.MatchString(testCase.Data), fmt.Sprintf("Test case: %+v", testCase.Data))
	}
}

func TestValidMonthRegex(t *testing.T) {
	testCases := []regexTestCase{
		{"0", false},
		{"1", true},
		{"2", true},
		{"3", true},
		{"4", true},
		{"5", true},
		{"6", true},
		{"7", true},
		{"8", true},
		{"9", true},
		{"10", true},
		{"11", true},
		{"12", true},
		{"13", false},
		{"JAN", true},
		{"FEB", true},
		{"MAR", true},
		{"APR", true},
		{"MAY", true},
		{"JUN", true},
		{"JUL", true},
		{"AUG", true},
		{"SEP", true},
		{"OCT", true},
		{"NOV", true},
		{"DEC", true},
		{"AGR", false},
	}
	for _, testCase := range testCases {
		assert.Equalf(t, testCase.Result, validMonthRegex.MatchString(testCase.Data), fmt.Sprintf("Test case: %+v", testCase.Data))
	}
}

func TestValidMonthRangeRegex(t *testing.T) {
	testCases := []regexTestCase{
		{"JAN-FEB", true},
		{"JAN-AGR", false},
		{"1-12", true},
		{"0-12", false},
	}
	for _, testCase := range testCases {
		assert.Equalf(t, testCase.Result, validMonthRangeRegex.MatchString(testCase.Data), fmt.Sprintf("Test case: %+v", testCase.Data))
	}
}

func TestValidWeekdayRegex(t *testing.T) {
	testCases := []regexTestCase{
		{"0", true},
		{"1", true},
		{"2", true},
		{"3", true},
		{"4", true},
		{"5", true},
		{"6", true},
		{"7", false},
		{"MON", true},
		{"TUE", true},
		{"WED", true},
		{"THU", true},
		{"FRI", true},
		{"SAT", true},
		{"SUN", true},
		{"GLO", false},
	}
	for _, testCase := range testCases {
		assert.Equalf(t, testCase.Result, validWeekdayRegex.MatchString(testCase.Data), fmt.Sprintf("Test case: %+v", testCase.Data))
	}
}

func TestValidWeekdayRangeRegex(t *testing.T) {
	testCases := []regexTestCase{
		{"MON-SUN", true},
		{"MON-GLO", false},
		{"0-6", true},
		{"0-7", false},
	}
	for _, testCase := range testCases {
		assert.Equalf(t, testCase.Result, validWeekdayRangeRegex.MatchString(testCase.Data), fmt.Sprintf("Test case: %+v", testCase.Data))
	}
}
