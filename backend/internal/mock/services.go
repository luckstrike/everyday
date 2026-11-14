// Package mock keeps each source separate so real adapters can later replace it.
package mock

import (
	"everyday/backend/internal/domain"
	"time"
)

func weatherState(now time.Time) domain.Weather {
	return domain.Weather{Daily: []domain.ForecastDay{{Date: now, High: 78, Low: 62, Condition: "Sunny", Icon: domain.IconClear, Precipitation: 5}, {Date: now.AddDate(0, 0, 1), High: 76, Low: 61, Condition: "Partly cloudy", Icon: domain.IconPartlyCloud, Precipitation: 15}, {Date: now.AddDate(0, 0, 2), High: 74, Low: 59, Condition: "Showers", Icon: domain.IconRain, Precipitation: 70}, {Date: now.AddDate(0, 0, 3), High: 73, Low: 60, Condition: "Sunny", Icon: domain.IconClear, Precipitation: 10}, {Date: now.AddDate(0, 0, 4), High: 75, Low: 61, Condition: "Partly cloudy", Icon: domain.IconPartlyCloud, Precipitation: 15}}, Location: "Dallas, TX", Temperature: 72, Condition: "Clear skies", Icon: domain.IconClear, High: 78, Low: 62, Precipitation: 10, Tomorrow: "A little warmer. Sunshine with a high of 74°.", Hourly: []domain.Hour{{Label: "Now", Temperature: 72, Condition: "Clear"}, {Label: "+1h", Temperature: 67, Condition: "Cloudy"}, {Label: "+2h", Temperature: 65, Condition: "Cloudy"}, {Label: "+3h", Temperature: 63, Condition: "Clear"}, {Label: "+4h", Temperature: 61, Condition: "Clear"}}, Alert: ""}
}
func financeState(now time.Time) domain.Finance {
	return domain.Finance{Bills: []domain.Bill{{Name: "Electricity", Amount: 84.50, DueAt: now.Add(5 * 24 * time.Hour)}, {Name: "Internet", Amount: 65, DueAt: now.Add(4 * 24 * time.Hour), Paid: true}, {Name: "Rent", Amount: 1850, DueAt: now.Add(20 * 24 * time.Hour)}}}
}
func calendarState(now time.Time) []domain.CalendarEvent {
	return []domain.CalendarEvent{{Title: "Dinner with Alex", Location: "At home", StartsAt: now.Add(25 * time.Hour), EndsAt: now.Add(27 * time.Hour)}, {Title: "Weekly planning", Location: "At home", StartsAt: now.Add(48 * time.Hour), EndsAt: now.Add(49 * time.Hour)}}
}

// anchor is fixed for the process lifetime, so polling does not move mock dates.
func Build(anchor, now time.Time) domain.State {
	return domain.State{GeneratedAt: now, Sources: domain.Sources{Weather: domain.SourceMock, Finance: domain.SourceMock, Calendar: domain.SourceMock}, Weather: weatherState(anchor), Finance: financeState(anchor), Calendar: calendarState(anchor)}
}
