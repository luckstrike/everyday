package mock

import (
	"testing"
	"time"
)

func TestScenariosExpireAndMockDatesStayStable(t *testing.T) {
	anchor := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	for _, scenario := range []string{"rain", "severe", "live", "starting", "final", "bill", "financial", "calendar"} {
		state := Build(anchor, anchor, scenario, anchor)
		if len(state.Events) != 1 || state.Scenario != scenario {
			t.Fatalf("missing %s", scenario)
		}
		expired := Build(anchor, anchor.Add(2*time.Minute), scenario, anchor)
		if len(expired.Events) != 0 || expired.Scenario != "normal" {
			t.Fatalf("unexpired %s", scenario)
		}
	}
	first := Build(anchor, anchor, "normal", anchor)
	later := Build(anchor, anchor.Add(time.Hour), "normal", anchor)
	if first.Calendar[0].StartsAt != later.Calendar[0].StartsAt {
		t.Fatal("mock date moved during polling")
	}
}
