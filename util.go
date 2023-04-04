package golog

import (
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/unicode/norm"
)

const (
	DefaultDateFormat = "2006-01-02 15:04:05"
)

// GetQueryParameter is a helper function to get query parameters by name and return
// error ONLY if they are required and not present and DO NOT have a defined default value.
//
// Example:
// ```
// qpApple,ok := GetQueryParameter(r,"apple",false,true,"Granny Smith")
//
//	if err != nil {
//			HandleError(w, http.StatusInternalServerError.RequestURI, err)
//			return
//	}
//
// ```
func GetQueryParameter(r *http.Request, name string, required bool, defaultOk bool, def string) (string, error) {

	var (
		ok   bool
		keys []string
	)

	name = strings.ToLower(name)

	// Check the request for the query parameter
	keys, ok = r.URL.Query()[name]
	//Log.Debugf("URL Query Values [%s]: %+v", r.URL.RawQuery, r.URL.Query())
	if !ok {
		keys = append(keys, "")

		// Is the missing query parameter required?
		if required {

			// Can we set a default value for the parameter?
			if !defaultOk {
				return "", fmt.Errorf("required query parameter [%s] is not present & default value not allowed", name)
			}

			keys[0] = def
			Log.Warningf("Required query parameter [%s] was not present, set to default value [%s]", name, def)
		}

		//param not present but not required, return default or empty
		if keys == nil || len(keys[0]) < 1 {
			// Can we set a default value for the parameter?
			if defaultOk {
				keys[0] = def
				Log.Debugf("Optional query parameter [%s] was not present, set to default value [%s]", name, def)
			} else {
				Log.Debugf("Optional query parameter [%s] was not present, default not allowed", name)
				keys[0] = ""
			}
		}
	} else {
		Log.Debugf("Query parameter [%s] is present with a value of [%v]", name, keys[0])
	}

	// Query()[name] will return an array of items,
	// we only want the single item.
	key := keys[0]

	// check for nil
	// if keys[0] == "" {
	// 	Log.Debugf("Returning URL Param [%s] of empty [nil]", name)
	// } else {
	// 	Log.Debugf("Returning URL Param [%s] value of [%s]", name, string(key))
	// }

	return key, nil
}

func ConvertToBoolean(s string) bool {
	s = strings.ToLower(s)
	if s == "1" || s == "y" || s == "yes" || s == "t" || s == "true" {
		return true
	}
	return false
}

// StringArrayContains helper function to return true or false is a given string exists in the provided string array
func StringArrayContains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func StringsToInts(sList string) (iList []int, err error) {
	var i, j int
	var s string

	array := strings.Split(sList, ",")
	for i, s = range array {
		if j, err = strconv.Atoi(strings.TrimSpace(s)); err != nil {
			err = fmt.Errorf("item %d: %q invalid number", i+1, s)
			return
		}
		iList = append(iList, j)
	}

	sort.Ints(iList)
	return
}

// Used to bucket names for alpha selector #, 0-9, A-Z
func AlphaSortNormalize(s string) (n string) {
	if len(s) == 0 {
		return "#"
	}

	// Remove leading "The"
	re := regexp.MustCompile(`^The `)
	s = re.ReplaceAllString(s, "")

	for _, b := range s {

		// Convert e or é to E
		c := strings.ToUpper(norm.NFD.String(string(b))[0:1])

		if c >= "0" && c <= "9" || c >= "A" && c <= "Z" || c == " " {
			// Keep 0-9, A-Z, space as is
			n += c
		} else {
			// Make other symbols into #
			n += "#"
		}
	}
	return
}

func parseUnitValue(unitRegex *regexp.Regexp, sub []string, unit string, currentValue *int, unitFound *bool) error {
	if unitRegex.MatchString(sub[2]) {
		if *unitFound {
			return fmt.Errorf("duplicate unit %q", unit)
		}
		value, err := strconv.Atoi(sub[1])
		if err != nil {
			return fmt.Errorf("failed to convert value %s", sub[1])
		}
		*currentValue = value
		*unitFound = true
	}
	return nil
}

func ParseDateInterval(interval string) (years, months, days int, err error) {
	var dFound, mFound, yFound bool

	d := []string{"days", "day", "d"}
	m := []string{"months", "month", "mo"}
	y := []string{"years", "year", "yr", "y"}
	all := append(d, m...)
	all = append(all, y...)
	allStr := strings.Join(all, ", ")

	// match overall pattern
	overallRgx := regexp.MustCompile(`^(-?\d+ ?\w+ ?){1,3}$`)
	splitRgx := regexp.MustCompile(`(-?\d+ ?\w+ ?)\b`)
	fieldRgx := regexp.MustCompile(`(-?\d+) ?(\w+)`)

	// Im getting a warning: should use raw string (`...`) with regexp.MustCompile to avoid having to escape twice (S1007)

	dRgx := regexp.MustCompile("^" + strings.Join(d, "|") + "$")
	mRgx := regexp.MustCompile("^" + strings.Join(m, "|") + "$")
	yRgx := regexp.MustCompile("^" + strings.Join(y, "|") + "$")

	s := strings.ToLower(strings.TrimSpace(interval))

	if !overallRgx.MatchString(s) {
		err = fmt.Errorf("invalid interval %q, must be up to 3 numbers, each followed by a unit: %s",
			interval, allStr)
		return
	}

	fields := splitRgx.FindAllString(s, 3)

	for _, field := range fields {
		sub := fieldRgx.FindStringSubmatch(field)
		if len(sub) != 3 {
			err = fmt.Errorf("invalid interval %q, failed to parse field %q",
				interval, field)
			return
		}

		err = parseUnitValue(dRgx, sub, sub[2], &days, &dFound)
		if err != nil {
			err = fmt.Errorf("invalid interval %q, %w", interval, err)
			return
		}

		err = parseUnitValue(mRgx, sub, sub[2], &months, &mFound)
		if err != nil {
			err = fmt.Errorf("invalid interval %q, %w", interval, err)
			return
		}

		err = parseUnitValue(yRgx, sub, sub[2], &years, &yFound)
		if err != nil {
			err = fmt.Errorf("invalid interval %q, %w", interval, err)
			return
		}
	}

	return
}

// IsStartDateValid is a helper function that check that provided date for NULL/ZERO or IS In the Past. If either of these
// check fail - the rereturn value is false. A StartDate MUST have a Valid Date and it must be in the PAST
func IsStartDateValid(dateT time.Time) error {
	if dateT.IsZero() {
		return fmt.Errorf("no start date")
	}

	// Get current Date values (NOW)
	year, month, day := dateT.Date()
	newDateTime := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
	//fmt.Printf("Start Date: %s\n", newDateTime.Format(DefaultDateFormat))
	if newDateTime.After(time.Now()) {
		return fmt.Errorf("start date [%s] is in future", newDateTime.Format(DefaultDateFormat))
	}

	return nil
}

// IsEndDateValid is a helper function that will check that the provided date is NULL/ZERO or IS In the Future. If either of these
// checks fail - the return value is false. A EndDate MUST have a NULL/ZERO Date OR it must be in the FUTURE
func IsEndDateValid(dateT time.Time) error {
	if dateT.IsZero() {
		return nil
	}

	year, month, day := dateT.Date()
	newDateTime := time.Date(year, month, day, 23, 59, 59, 0, time.Local)
	//fmt.Printf("End Date: %s\n", newDateTime.Format(DefaultDateFormat))
	if newDateTime.Before(time.Now()) {
		return fmt.Errorf("end date [%s] is in past", newDateTime.Format(DefaultDateFormat))
	}

	return nil
}

// IsValidDateRange - helper function used for uniform  start and end date validation
// Param startT		- Start DateTime
// Param endT		- End DateTime
// Return error 	- NIL if dates are in range otherwise MSG as to WHY it failed
func IsValidDateRange(startT, endT time.Time) error {
	var err error

	err = IsStartDateValid(startT)
	if err != nil {
		return err
	}

	err = IsEndDateValid(endT)
	if err != nil {
		return err
	}

	return nil
}

func IsValidDateStringsRange(startStr, endStr string) (err error) {
	var sDate, eDate time.Time

	if len(startStr) > 0 {
		sDate, err = SmartGetDateT(startStr)
		if err != nil {
			return err
		}
	}

	if len(endStr) > 0 {
		eDate, err = SmartGetDateT(endStr)
		if err != nil {
			return err
		}
	}

	return IsValidDateRange(sDate, eDate)
}

// IsZuluDate returns true if the given date string ends with a 'Z' or 'z', indicating a Zulu time format.
func IsZuluDate(s string) bool {
	if len(s) == 0 {
		return false
	}

	lastChar := s[len(s)-1]
	return lastChar == 'Z' || lastChar == 'z'
}

// SmartGetDateT helper function to get a workable time using OUR date format.
// NOTE: At the moment this is hard-coded throughout the the API and all micro services
// This function will properly parse the following date time formats that we are using in the API
//
// - 2006-01-02T15:04:05Z07:00
// - 2006-01-02 15:04:05
// - 2006-01-02
func SmartGetDateT(d string) (result time.Time, err error) {

	if len(d) == 0 {
		err = fmt.Errorf("cannot parse empty string for datetime")
		return
	}

	if IsZuluDate(d) {
		//layOut := "2006-01-02T15:04:05Z07:00" // yyyy-dd-MM
		result, err = time.Parse(time.RFC3339, d)

		if err != nil {
			err = fmt.Errorf("parsing Date [%s] using Format [2006-01-02T15:04:05Z07:00]: %v", d, err)
			return
		}
		// Success
		return
	}

	//layout := "2006-01-02"
	if len(d) == 10 {
		result, err = time.Parse("2006-01-02", d)
		if err != nil {
			err = fmt.Errorf("parsing Date [%s] using Format [2006-01-02]: %v", d, err)
			return
		}
		// Success
		return
	}

	//layout := "2006-01-02 15:04:05"
	result, err = time.Parse(DefaultDateFormat, d)
	if err != nil {
		err = fmt.Errorf("parsing Date [%s] using Format [%s]: %v", d, time.RFC3339, err)
		return
	}

	// Success
	return
}

// DateIsWithin will return TRUE if the provided date string is within the provided duration string
func DateIsWithin(date string, duration string) (result bool, err error) {
	var (
		dur time.Duration
		dt  time.Time
	)

	if len(date) == 0 || len(duration) == 0 {
		err = fmt.Errorf("DateIsWithin: Please provide valid date and duration values")
		return
	}

	// Attempt to parse the provided duration
	dur, err = time.ParseDuration(duration)
	if err != nil {
		err = fmt.Errorf("DateIsWithin: %v", err)
		return
	}

	// attempt to parse the provided date string
	dt, err = SmartGetDateT(date)
	if err != nil {
		err = fmt.Errorf("DateIsWithin: %v", err)
		return
	}

	result = !dt.Add(dur).Before(time.Now())

	return
}

// GetDateT helper function to get a workable time using OUR date format.
// NOTE: At the moment is is hard-coded throughout the the API and all micro services
//
//	This
func GetDateT(d string) time.Time {
	var (
		result time.Time
		err    error
	)

	if len(d) > 10 {
		// Long Date/Time
		result, err = time.Parse(DefaultDateFormat, d)
		if err != nil {
			Log.Warningf("Parse Long DateTime %s: %v", d, err)
		}
	} else {
		// Short Date No Time
		result, err = time.Parse("2006-01-02", d)
		if err != nil {
			Log.Warningf("Parse Short DateTime %s: %v", d, err)
		}
	}
	return result
}

// GetDateWithErrorT helper function to get a workable time using OUR date format and return any error
func GetDateWithErrorT(d string) (t time.Time, err error) {
	t, err = time.Parse(DefaultDateFormat, d)
	return
}
