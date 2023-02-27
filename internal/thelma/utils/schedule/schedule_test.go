package schedule

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestCheckDailyScheduleMatch(t *testing.T) {
	type args struct {
		schedule time.Time
		since    time.Time
		now      time.Time
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "same day before miss",
			args: args{
				schedule: testTimeFactory(t, "2000-01-01T17:20:00-05:00"),
				since:    testTimeFactory(t, "2023-02-24T16:55:00-05:00"),
				now:      testTimeFactory(t, "2023-02-24T17:15:00-05:00"),
			},
			want: false,
		},
		{
			name: "same day hit",
			args: args{
				schedule: testTimeFactory(t, "2000-01-01T17:20:00-05:00"),
				since:    testTimeFactory(t, "2023-02-24T17:10:00-05:00"),
				now:      testTimeFactory(t, "2023-02-24T17:30:00-05:00"),
			},
			want: true,
		},
		{
			name: "same day after miss",
			args: args{
				schedule: testTimeFactory(t, "2000-01-01T17:20:00-05:00"),
				since:    testTimeFactory(t, "2023-02-24T17:25:00-05:00"),
				now:      testTimeFactory(t, "2023-02-24T17:45:00-05:00"),
			},
			want: false,
		},
		{
			name: "hit across timezones",
			args: args{
				schedule: testTimeFactory(t, "2000-01-01T17:20:00-06:00"),
				since:    testTimeFactory(t, "2023-02-24T18:10:00-05:00"),
				now:      testTimeFactory(t, "2023-02-24T18:30:00-05:00"),
			},
			want: true,
		},
		{
			name: "miss before midnight",
			args: args{
				schedule: testTimeFactory(t, "2000-01-01T23:40:00-05:00"),
				since:    testTimeFactory(t, "2023-02-24T23:50:00-05:00"),
				now:      testTimeFactory(t, "2023-02-25T00:10:00-05:00"),
			},
			want: false,
		},
		{
			name: "hit before midnight",
			args: args{
				schedule: testTimeFactory(t, "2000-01-01T23:55:00-05:00"),
				since:    testTimeFactory(t, "2023-02-24T23:50:00-05:00"),
				now:      testTimeFactory(t, "2023-02-25T00:10:00-05:00"),
			},
			want: true,
		},
		{
			name: "hit at midnight",
			args: args{
				schedule: testTimeFactory(t, "2000-01-01T00:00:00-05:00"),
				since:    testTimeFactory(t, "2023-02-24T23:50:00-05:00"),
				now:      testTimeFactory(t, "2023-02-25T00:10:00-05:00"),
			},
			want: true,
		},
		{
			name: "hit after midnight",
			args: args{
				schedule: testTimeFactory(t, "2000-01-01T00:05:00-05:00"),
				since:    testTimeFactory(t, "2023-02-24T23:50:00-05:00"),
				now:      testTimeFactory(t, "2023-02-25T00:10:00-05:00"),
			},
			want: true,
		},
		{
			name: "miss after midnight",
			args: args{
				schedule: testTimeFactory(t, "2000-01-01T00:15:00-05:00"),
				since:    testTimeFactory(t, "2023-02-24T23:50:00-05:00"),
				now:      testTimeFactory(t, "2023-02-25T00:10:00-05:00"),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckDailyScheduleMatch(tt.args.schedule, tt.args.since, tt.args.now); got != tt.want {
				t.Errorf("CheckDailyScheduleMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsWeekendDay(t *testing.T) {
	type args struct {
		date time.Time
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "friday false",
			args: args{date: testTimeFactory(t, "2023-02-24T12:00:00-05:00")},
			want: false,
		},
		{
			name: "saturday true",
			args: args{date: testTimeFactory(t, "2023-02-25T12:00:00-05:00")},
			want: true,
		},
		{
			name: "sunday true",
			args: args{date: testTimeFactory(t, "2023-02-26T12:00:00-05:00")},
			want: true,
		},
		{
			name: "monday false",
			args: args{date: testTimeFactory(t, "2023-02-27T12:00:00-05:00")},
			want: false,
		},
		{
			name: "friday night false",
			args: args{date: testTimeFactory(t, "2023-02-24T23:00:00-05:00")},
			want: false,
		},
		{
			name: "friday night in UTC true",
			args: args{date: testTimeFactory(t, "2023-02-24T23:00:00-05:00").In(time.UTC)},
			want: true,
		},
		{
			name: "sunday night true",
			args: args{date: testTimeFactory(t, "2023-02-26T23:00:00-05:00")},
			want: true,
		},
		{
			name: "sunday night in UTC false",
			args: args{date: testTimeFactory(t, "2023-02-26T23:00:00-05:00").In(time.UTC)},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsWeekendDay(tt.args.date); got != tt.want {
				t.Errorf("IsWeekendDay() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_replaceDateOfTime(t *testing.T) {
	type args struct {
		timeToUse time.Time
		dateToUse time.Time
	}
	tests := []struct {
		args args
		want string
	}{
		{
			args: args{
				timeToUse: testTimeFactory(t, "2023-02-24T12:00:00-05:00"),
				dateToUse: testTimeFactory(t, "1999-01-02T06:00:00-06:00"),
			},
			want: "1999-01-02T12:00:00-05:00",
		},
		{
			args: args{
				timeToUse: testTimeFactory(t, "1999-01-02T06:00:00-06:00"),
				dateToUse: testTimeFactory(t, "2023-02-24T12:00:00-05:00"),
			},
			want: "2023-02-24T06:00:00-06:00",
		},
	}
	for idx, tt := range tests {
		t.Run(fmt.Sprintf("%d", idx+1), func(t *testing.T) {
			if got := replaceDateOfTime(tt.args.timeToUse, tt.args.dateToUse).Format(time.RFC3339); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("replaceDateOfTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func testTimeFactory(t *testing.T, rfc3339 string) time.Time {
	ret, err := time.Parse(time.RFC3339, rfc3339)
	if err != nil {
		t.Errorf("couldn't parse %s to time: %v", rfc3339, err)
		t.Fail()
	}
	return ret
}
