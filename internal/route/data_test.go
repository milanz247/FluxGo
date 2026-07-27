package route_test

import (
	"testing"

	Route "fluxgo/internal/route"
)

func TestDataSetAndMerge(t *testing.T) {
	data := Route.Data{}.
		Set("name", "FluxGo").
		Merge(Route.Data{
			"active": true,
			"count":  2,
		})

	if data["name"] != "FluxGo" || data["active"] != true || data["count"] != 2 {
		t.Fatalf("unexpected data: %#v", data)
	}
}
