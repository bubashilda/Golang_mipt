//go:build !solution

package hotelbusiness

import "sort"

type Guest struct {
	CheckInDate  int
	CheckOutDate int
}

type Load struct {
	StartDate  int
	GuestCount int
}

type Event struct {
	Close bool
	Date  int
}

func ComputeLoad(guests []Guest) []Load {
	var ans []Load
	var events []Event

	for _, g := range guests {
		events = append(events, Event{Close: true, Date: g.CheckOutDate})
		events = append(events, Event{Close: false, Date: g.CheckInDate})
	}

	sort.Slice(events, func(i, j int) bool {
		if events[i].Date == events[j].Date {
			return events[i].Close
		}
		return events[i].Date < events[j].Date
	})

	countGuests := 0
	for i := 0; i < len(events); i++ {
		record := Load{events[i].Date, countGuests}
		j := i
		for j < len(events) && events[j].Date == events[i].Date {
			if events[j].Close {
				record.GuestCount -= 1
			} else {
				record.GuestCount += 1
			}
			j += 1
		}
		i = j - 1
		if record.GuestCount != countGuests {
			countGuests = record.GuestCount
			ans = append(ans, record)
		}
	}
	return ans
}
