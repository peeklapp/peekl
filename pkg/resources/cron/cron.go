package cron

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"slices"

	"github.com/mitchellh/mapstructure"
	"github.com/peeklapp/peekl/pkg/models"
	"github.com/peeklapp/peekl/pkg/resources"
	"github.com/peeklapp/peekl/pkg/utils"
	"github.com/sirupsen/logrus"
)

var (
	validNameRegex         = regexp.MustCompile("^[a-zA-Z0-9_-]+$")
	validMinuteRegex       = regexp.MustCompile("^(5[0-9]|[1234][0-9]|[0-9])$")
	validMinuteRangeRegex  = regexp.MustCompile("^(5[0-9]|[1234][0-9]|[0-9])-(5[0-9]|[1234][0-9]|[0-9])$")
	validHourRegex         = regexp.MustCompile("^(2[0123]|1[0-9]|[0-9])$")
	validHourRangeRegex    = regexp.MustCompile("^(2[0123]|1[0-9]|[0-9])-(2[0123]|1[0-9]|[0-9])$")
	validDayRegex          = regexp.MustCompile("^(3[01]|[12][0-9]|[1-9])$")
	validDayRangeRegex     = regexp.MustCompile("^(3[01]|[12][0-9]|[0-9])-(3[01]|[12][0-9]|[0-9])$")
	validMonthRegex        = regexp.MustCompile("^([1-9]|1[0-2]|JAN|FEB|MAR|APR|MAY|JUN|JUL|AUG|SEP|OCT|NOV|DEC)$")
	validMonthRangeRegex   = regexp.MustCompile("^(JAN|FEB|MAR|APR|MAY|JUN|JUL|AUG|SEP|OCT|NOV|DEC|[1-9]|1[0-2])-(JAN|FEB|MAR|APR|MAY|JUN|JUL|AUG|SEP|OCT|NOV|DEC|[1-9]|1[0-2])$")
	validWeekdayRegex      = regexp.MustCompile("^([0-6])|(MON|TUE|WED|THU|FRI|SAT|SUN)$")
	validWeekdayRangeRegex = regexp.MustCompile("^((MON|TUE|WED|THU|FRI|SAT|SUN)|[0-6])-((MON|TUE|WED|THU|FRI|SAT|SUN)|[0-6])$")
)

type CronData struct {
	Name       string `mapstructure:"name"`
	Command    string `mapstructure:"command"`
	User       string `mapstructure:"user"`
	Minute     string `mapstructure:"minute"`
	Hour       string `mapstructure:"hour"`
	Day        string `mapstructure:"day"`
	Month      string `mapstructure:"month"`
	Weekday    string `mapstructure:"weekday"`
	CronFolder string `mapstructure:"cron_folder"`
}

type CronResource struct {
	resources.CommonFieldResource
	Data CronData
}

func (c *CronResource) Process(context *models.ResourceContext) (models.ResourceResult, error) {
	var result models.ResourceResult

	if !utils.FileExist(c.Data.CronFolder) {
		result.Failed = true
		return result, fmt.Errorf(
			"The cron folder provided does not seem to exist : %s",
			c.Data.CronFolder,
		)
	}

	cronFilePath := path.Join(c.Data.CronFolder, c.Data.Name)

	if c.Present {
		fileContent := []byte("# MANAGED BY PEEKL\n")
		fileContent = fmt.Appendf(
			fileContent,
			"%s %s %s %s %s %s %s",
			c.Data.Minute,
			c.Data.Hour,
			c.Data.Day,
			c.Data.Month,
			c.Data.Weekday,
			c.Data.User,
			c.Data.Command,
		)

		if !utils.FileExist(cronFilePath) {
			logrus.Info(
				fmt.Sprintf("[%s] Cron (%s) does not exist but should", c.String(), c.Data.Name),
			)
			err := os.WriteFile(cronFilePath, fileContent, 0644)
			if err != nil {
				result.Failed = true
				return result, err
			}
			logrus.Info(
				fmt.Sprintf("[%s] Cron (%s) has been created", c.String(), c.Data.Name),
			)
			result.Created = true
			return result, nil
		} else {
			oldFileContent, err := os.ReadFile(cronFilePath)
			if err != nil {
				result.Failed = true
				return result, err
			}
			if !slices.Equal(fileContent, oldFileContent) {
				logrus.Info(
					fmt.Sprintf("[%s] Cron (%s) does not have the correct value", c.String(), c.Data.Name),
				)
				err := os.WriteFile(cronFilePath, fileContent, 0644)
				if err != nil {
					result.Failed = true
					return result, err
				}
				logrus.Info(
					fmt.Sprintf("[%s] Cron (%s) values have been updated", c.String(), c.Data.Name),
				)
				result.Updated = true
				return result, nil
			}
		}
	}

	if utils.FileExist(cronFilePath) {
		os.Remove(cronFilePath)
	}
	result.Deleted = true
	return result, nil
}

func (c *CronResource) String() string {
	return fmt.Sprintf("%s / %s", c.Type, c.Title)
}

func (c *CronResource) When() string {
	return c.WhenCondition
}

func (c *CronResource) Register() string {
	return c.RegisterVariable
}

func (c *CronResource) validateMinuteField() bool {
	minuteValid := false
	if c.Data.Minute == "*" {
		minuteValid = true
	} else if validMinuteRegex.MatchString(c.Data.Minute) {
		minuteValid = true
	} else if validMinuteRangeRegex.MatchString(c.Data.Minute) {
		minuteValid = true
	}
	return minuteValid
}

func (c *CronResource) validateHourField() bool {
	hourValid := false
	if c.Data.Hour == "*" {
		hourValid = true
	} else if validHourRegex.MatchString(c.Data.Hour) {
		hourValid = true
	} else if validHourRangeRegex.MatchString(c.Data.Hour) {
		hourValid = true
	}
	return hourValid
}

func (c *CronResource) validateDayField() bool {
	dayValid := false
	if c.Data.Day == "*" {
		dayValid = true
	} else if validDayRegex.MatchString(c.Data.Day) {
		dayValid = true
	} else if validDayRangeRegex.MatchString(c.Data.Day) {
		dayValid = true
	}
	return dayValid
}

func (c *CronResource) validateMonthField() bool {
	monthValid := false
	if c.Data.Month == "*" {
		monthValid = true
	} else if validMonthRegex.MatchString(c.Data.Month) {
		monthValid = true
	} else if validMonthRangeRegex.MatchString(c.Data.Month) {
		monthValid = true
	}
	return monthValid
}

func (c *CronResource) validateWeekdayField() bool {
	weekdayValid := false
	if c.Data.Weekday == "*" {
		weekdayValid = true
	} else if validWeekdayRegex.MatchString(c.Data.Weekday) {
		weekdayValid = true
	} else if validWeekdayRangeRegex.MatchString(c.Data.Weekday) {
		weekdayValid = true
	}
	return weekdayValid
}

func (c *CronResource) Validate() error {
	validationErrors := []models.ValidationError{}

	if c.Data.Command == "" {
		validationErrors = append(
			validationErrors,
			models.ValidationError{
				FieldName:    "command",
				ViolatedRule: "Field cannot be empty",
			},
		)
	}

	if c.Data.Name == "" {
		validationErrors = append(
			validationErrors,
			models.ValidationError{
				FieldName:    "name",
				ViolatedRule: "Field cannot be empty",
			},
		)
	} else {
		if !validNameRegex.MatchString(c.Data.Name) {
			validationErrors = append(
				validationErrors,
				models.ValidationError{
					FieldName:    "name",
					ViolatedRule: "Field can only contain alphanumerical symbols, underscores, and dash",
				},
			)
		}
	}

	if c.Data.User == "" {
		validationErrors = append(
			validationErrors,
			models.ValidationError{
				FieldName:    "user",
				ViolatedRule: "Field cannot be empty",
			},
		)
	}

	if !c.validateMinuteField() {
		validationErrors = append(
			validationErrors,
			models.ValidationError{
				FieldName:    "minute",
				ViolatedRule: "Field is not a valid minute value : *, 0 -> 59, 0->59-0->59",
			},
		)
	}

	if !c.validateHourField() {
		validationErrors = append(
			validationErrors,
			models.ValidationError{
				FieldName:    "hour",
				ViolatedRule: "Field is not a valid hour value : *, 0 -> 23, 0->23-0->23",
			},
		)
	}

	if !c.validateDayField() {
		validationErrors = append(
			validationErrors,
			models.ValidationError{
				FieldName:    "day",
				ViolatedRule: "Field is not a valid day value: *, 0->31, 0->31-0->31",
			},
		)
	}

	if !c.validateMonthField() {
		validationErrors = append(
			validationErrors,
			models.ValidationError{
				FieldName:    "month",
				ViolatedRule: "Field is not a valid month value : *, 1->12, 1->12-1->12, JAN->DEC, JAN->DEC-JAN->DEC",
			},
		)
	}

	if !c.validateWeekdayField() {
		validationErrors = append(
			validationErrors,
			models.ValidationError{
				FieldName:    "weekday",
				ViolatedRule: "Field is not a valid weekday value : *, 0->6, 0->6-0->6, MON->SUN, MON->SUN-MON->SUN",
			},
		)
	}

	if len(validationErrors) > 0 {
		return models.ResourceValidationError{
			Type:             c.Type,
			Title:            c.Title,
			ValidationErrors: validationErrors,
		}
	}

	return nil
}

func NewCronResource(resource *models.Resource, dataField any, roleContext *models.RoleContext) (*CronResource, error) {
	var cronResource CronResource

	defaults := map[string]any{
		"minute":      "*",
		"hour":        "*",
		"day":         "*",
		"month":       "*",
		"weekday":     "*",
		"user":        "root",
		"cron_folder": "/etc/cron.d",
	}

	var cronData CronData

	err := mapstructure.Decode(defaults, &cronData)
	if err != nil {
		return nil, err
	}

	err = mapstructure.Decode(dataField, &cronData)
	if err != nil {
		return nil, err
	}

	cronResource.Title = resource.Title
	cronResource.Type = resource.Type
	cronResource.Present = *resource.Present
	cronResource.WhenCondition = resource.When
	cronResource.RegisterVariable = resource.Register
	cronResource.Data = cronData

	return &cronResource, nil
}
