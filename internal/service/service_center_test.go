package service_test

import (
	"strings"
	"testing"

	"vyshkat/internal/domain"
	"vyshkat/internal/service"
)

func TestBasicServiceCenterInspectsEveryVehicleType(t *testing.T) {
	constructors := []struct {
		name   string
		create func(bool) (domain.Vehicle, error)
	}{
		{"scooter", func(operational bool) (domain.Vehicle, error) {
			return domain.NewScooter("Самокат", "S-001", 6, 1.5, operational)
		}},
		{"bicycle", func(operational bool) (domain.Vehicle, error) {
			return domain.NewBicycle("Велосипед", "B-001", 6, operational)
		}},
		{"electric_bicycle", func(operational bool) (domain.Vehicle, error) {
			return domain.NewElectricBicycle("Электровелосипед", "E-001", 6, 1.5, operational)
		}},
	}
	cases := []struct {
		name        string
		operational bool
	}{
		{"operational", true},
		{"broken", false},
	}
	// Обращаемся к настоящей реализации через интерфейс.
	var center service.ServiceCenter = service.NewBasicServiceCenter()
	for _, constructor := range constructors {
		for _, tc := range cases {
			t.Run(constructor.name+"/"+tc.name, func(t *testing.T) {
				vehicle, err := constructor.create(tc.operational)
				if err != nil {
					t.Fatalf("create vehicle: %v", err)
				}
				result := center.Inspect(vehicle)
				if result.Accepted != tc.operational {
					t.Errorf("Accepted = %t, want %t", result.Accepted, tc.operational)
				}
				if result.Accepted && result.Reason != "" {
					t.Errorf("accepted vehicle has rejection reason: %q", result.Reason)
				}
				if !result.Accepted && strings.TrimSpace(result.Reason) == "" {
					t.Error("rejected vehicle must have a reason")
				}
			})
		}
	}
}

func TestBasicServiceCenterSeparatesInspectionFromBeginnerRule(t *testing.T) {
	cases := []struct {
		name         string
		score        int
		operational  bool
		wantAccepted bool
	}{
		{"healthy_difficult_vehicle", 1, true, true},
		{"broken_easy_vehicle", 10, false, false},
	}
	center := service.NewBasicServiceCenter()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			vehicle, err := domain.NewBicycle("Велосипед", "B-001", tc.score, tc.operational)
			if err != nil {
				t.Fatalf("create bicycle: %v", err)
			}
			result := center.Inspect(vehicle)
			if result.Accepted != tc.wantAccepted {
				t.Errorf("Accepted = %t, want %t", result.Accepted, tc.wantAccepted)
			}
		})
	}
}
