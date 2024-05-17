package schedule

import (
	"time"
)

// CheckDailyScheduleMatch determines if the `schedule` time's daily repetition happened between `since` and `now`.
// We assume that `since` is before `now`.
//
// An important edge case here is if midnight falls in between `since` and `now`. We want to convert `schedule` to the
// correct day, but there's now two possible candidates.
//
// We handle this case by always initially converting `schedule` to `now`'s day. If `schedule` is then ahead of `now`,
// we convert it to `since`'s day instead and then try again. Suppose `since` of 23:50 and `now` of 00:10. If `since`
// is 00:05, we correct identify it on the first pass, since it being on `now`'s day puts it between the two times.
// If since is 23:55, we correctly identify it on the second pass.
//
// This edge case handling is correct in all cases because this function is concerned just with daily repetitions.
// We could brute-force check all possible dates here and this function wouldn't be incorrect.
func CheckDailyScheduleMatch(schedule time.Time, since time.Time, now time.Time) bool {
	scheduleLocalized := schedule.In(now.Location())
	scheduleToUse := replaceDateOfTime(scheduleLocalized, now)

	if scheduleToUse.After(now) {
		scheduleToUse = replaceDateOfTime(scheduleLocalized, since)
	}
	return scheduleToUse.After(since) && scheduleToUse.Before(now)
}

// IsWeekendDay returns if the date occurred on a weekend day. The same time will be different days in different
// timezones so be careful if you're passing UTC timestamps to this function--see the tests for an example.
func IsWeekendDay(date time.Time) bool {
	return date.Weekday() == time.Sunday || date.Weekday() == time.Saturday
}

// replaceDateOfTime returns the same time in the same timezone on a different date.
func replaceDateOfTime(timeToUse time.Time, dateToUse time.Time) time.Time {
	return time.Date(
		dateToUse.Year(),
		dateToUse.Month(),
		dateToUse.Day(),
		timeToUse.Hour(),
		timeToUse.Minute(),
		timeToUse.Second(),
		timeToUse.Nanosecond(),
		timeToUse.Location(),
	)
}
