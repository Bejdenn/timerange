package timerange

import (
	"fmt"
	"strings"
	"time"
)

const TimeOnlyNoSeconds = "15:04"

// Max returns the time that is further in the future.
func Max(t, u time.Time) time.Time {
	if t.After(u) {
		return t
	}
	return u
}

// Min returns the time that is further in the past.
func Min(t, u time.Time) time.Time {
	if t.Before(u) {
		return t
	}
	return u
}

type TimeRange struct {
	Start, End time.Time
}

// NewUnbound creates a TimeRange for a given start.
// That is, the end of the time range is set to be one day later to the start, at 23:59.
func NewUnbound(start time.Time) (TimeRange, error) {
	return New(start, time.Date(start.Year(), start.Month(), start.Day()+1, 23, 59, 0, 0, start.Location()))
}

// New creates a TimeRange for the given start and end.
func New(start, end time.Time) (TimeRange, error) {
	if start.IsZero() {
		return TimeRange{}, fmt.Errorf("start is zero time")
	} else if end.IsZero() {
		return TimeRange{}, fmt.Errorf("end is zero time")
	} else if start.After(end) {
		return TimeRange{}, fmt.Errorf("start (%v) is after end (%v)", start, end)
	}

	return TimeRange{start, end}, nil
}

func (tr TimeRange) Duration() time.Duration {
	return tr.End.Sub(tr.Start)
}

// Sub returns tr-u.
// The difference of two time ranges is defined as the complement between the two.
// The returned slice contains all partial time ranges of tr that do not overlap with u.
func (tr TimeRange) Sub(u TimeRange) []TimeRange {
	res := make([]TimeRange, 0)
	// tr.Start  tr.End
	//  |---------|
	//     |---------|
	// u.Start      u.End
	if tr.Start.Before(u.Start) {
		res = append(res, TimeRange{tr.Start, Min(u.Start, tr.End)})
	}
	// tr.Start    tr.End
	//     |---------|
	//  |---------|
	// u.Start  u.End
	if tr.End.After(u.End) {
		res = append(res, TimeRange{Max(u.End, tr.Start), tr.End})
	}
	return res
}

// SubMulti performs Sub on all us.
//
// If us is empty, the returned slice will only contain tr.
//
// If us contains one element, the call will act like Sub and the returned slice will contain tr.Sub(u).
//
// If us contains more than one element, SubMulti will apply every u in us on tr and its Sub results.
func (tr TimeRange) SubMulti(us []TimeRange) []TimeRange {
	if len(us) == 0 {
		return []TimeRange{tr}
	}

	out := []TimeRange{}
	for _, s := range tr.Sub(us[0]) {
		out = append(out, s.SubMulti(us[min(1, len(us)):])...)
	}
	return out
}

// Parse parses a list of time ranges from a list of strings.
// The individual strings are expected to be in the format "HH:MM-HH:MM".
// As the time is internally parsed without knowing the date, the date is taken from a reference time, ref.
func Parse(ref time.Time, values []string) ([]TimeRange, error) {
	trs := make([]TimeRange, 0, len(values))
	for _, v := range values {
		parts := strings.Split(v, "-")
		if len(parts) != 2 {
			return nil, fmt.Errorf("exclude did not contain hyphen delimiter: %s", v)
		}

		start, err := time.Parse(TimeOnlyNoSeconds, parts[0])
		if err != nil {
			return nil, fmt.Errorf("could not parse start time: %v", err)
		}

		end, err := time.Parse(TimeOnlyNoSeconds, parts[1])
		if err != nil {
			return nil, fmt.Errorf("could not parse end time: %v", err)
		}

		trs = append(trs, TimeRange{Start: Normalize(ref, start), End: Normalize(ref, end)})
	}
	return trs, nil
}

// Normalize takes a reference time and a target time and returns the target time with the date of the reference time.
func Normalize(ref, t time.Time) time.Time {
	return time.Date(ref.Year(), ref.Month(), ref.Day(), t.Hour(), t.Minute(), 0, 0, ref.Location())
}
