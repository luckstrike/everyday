// Package mock keeps each source separate so real adapters can later replace it.
package mock

import (
	"apartment-display/backend/internal/domain"
	"time"
)

func weatherState() domain.Weather {
	return domain.Weather{Location: "Chicago, IL", Temperature: 68, Condition: "Partly cloudy", High: 72, Low: 57, Precipitation: 10, Tomorrow: "A little warmer. Sunshine with a high of 74°.", Hourly: []domain.Hour{{Label: "Now", Temperature: 68, Condition: "Partly cloudy"}, {Label: "+1h", Temperature: 67, Condition: "Cloudy"}, {Label: "+2h", Temperature: 65, Condition: "Cloudy"}, {Label: "+3h", Temperature: 63, Condition: "Clear"}, {Label: "+4h", Temperature: 61, Condition: "Clear"}}, Alert: ""}
}
func sportsState(now time.Time) domain.Sports {
	return domain.Sports{League: "NBA · Preseason", Home: "Bulls", Away: "Bucks", Status: "scheduled", Detail: "United Center · Chicago", StartsAt: now.Add(48 * time.Hour)}
}
func financeState(now time.Time) domain.Finance {
	return domain.Finance{Bills: []domain.Bill{{Name: "Electricity", Amount: 84.50, DueAt: now.Add(5 * 24 * time.Hour), Autopay: true}, {Name: "Internet", Amount: 65, DueAt: now.Add(9 * 24 * time.Hour), Autopay: true}, {Name: "Rent", Amount: 1850, DueAt: now.Add(20 * 24 * time.Hour), Autopay: false}}, MonthlyBudget: 3200, Spent: 2146.80}
}
func calendarState(now time.Time) []domain.CalendarEvent {
	return []domain.CalendarEvent{{Title: "Dinner with friends", Location: "Lula Cafe", StartsAt: now.Add(25 * time.Hour), EndsAt: now.Add(27 * time.Hour)}, {Title: "Weekly planning", Location: "At home", StartsAt: now.Add(48 * time.Hour), EndsAt: now.Add(49 * time.Hour)}}
}

// anchor is fixed for the process lifetime, so polling does not move mock dates.
func Build(anchor, now time.Time, scenario string, started time.Time) domain.State {
	state := domain.State{GeneratedAt: now, Mock: true, Scenario: "normal", Weather: weatherState(), Sports: sportsState(anchor), Finance: financeState(anchor), Calendar: calendarState(anchor), Events: []domain.AttentionEvent{}}
	if scenario == "normal" || !now.Before(started.Add(2*time.Minute)) {
		return state
	}
	state.Scenario = scenario
	var widget, kind string
	switch scenario {
	case "rain":
		widget, kind = "weather", "rain"
		state.Weather.Condition = "Rain in 15 minutes"
		state.Weather.Precipitation = 85
	case "severe":
		widget, kind = "weather", "severe"
		state.Weather.Condition = "Thunderstorms"
		state.Weather.Precipitation = 100
		state.Weather.Alert = "Simulated severe thunderstorm warning. Stay indoors, away from windows."
	case "live":
		widget, kind = "sports", "live"
		state.Sports.Status = "live"
		state.Sports.HomeScore = 86
		state.Sports.AwayScore = 82
		state.Sports.Detail = "Q3 · 04:32 remaining"
	case "starting":
		widget, kind = "sports", "starting"
		state.Sports.StartsAt = started.Add(40 * time.Minute)
	case "final":
		widget, kind = "sports", "final"
		state.Sports.Status = "final"
		state.Sports.HomeScore = 112
		state.Sports.AwayScore = 108
		state.Sports.Detail = "Final · Bulls win"
	case "bill":
		widget, kind = "finance", "tomorrow"
		state.Finance.Bills[0].DueAt = started.Add(24 * time.Hour)
	case "financial":
		widget, kind = "finance", "alert"
	case "calendar":
		widget, kind = "calendar", "soon"
		state.Calendar[0].StartsAt = started.Add(20 * time.Minute)
		state.Calendar[0].EndsAt = started.Add(2 * time.Hour)
	default:
		return state
	}
	state.Events = append(state.Events, domain.AttentionEvent{ID: scenario + ":" + started.Format(time.RFC3339Nano), Widget: widget, Kind: kind, StartsAt: started, ExpiresAt: started.Add(2 * time.Minute), Decay: scenario == "final"})
	return state
}
func ValidScenario(s string) bool {
	switch s {
	case "normal", "rain", "severe", "live", "starting", "final", "bill", "financial", "calendar":
		return true
	}
	return false
}
