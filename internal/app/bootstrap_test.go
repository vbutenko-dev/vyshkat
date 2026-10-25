package app_test

import (
	"errors"
	"testing"

	"vyshkat/internal/app"
	"vyshkat/internal/domain"
	"vyshkat/internal/service"
)

func TestBuildFleetWiresBasicInspection(t *testing.T) {
	fleet, err := app.BuildFleet()
	if err != nil {
		t.Fatalf("BuildFleet: %v", err)
	}
	if got := fleet.Report(); got != (service.FleetReport{}) {
		t.Fatalf("new fleet is not empty: %+v", got)
	}
	broken, err := domain.NewScooter("Неисправный", "S-001", 10, 10, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := fleet.AddVehicle(broken); !errors.Is(err, service.ErrInspectionRejected) {
		t.Fatalf("basic inspection must reject broken transport: %v", err)
	}
	healthy, err := domain.NewScooter("Исправный", "S-001", 6, 1.5, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := fleet.AddVehicle(healthy); err != nil {
		t.Fatalf("basic inspection must accept operational transport: %v", err)
	}
	want := service.FleetReport{
		VehicleCount:     1,
		TotalCount:       1,
		VehicleEnergyKWh: 1.5,
		TotalEnergyKWh:   1.5,
	}
	if got := fleet.Report(); got != want {
		t.Errorf("built fleet report = %+v, want %+v", got, want)
	}
}

func TestBuildFleetCreatesIndependentParks(t *testing.T) {
	first, err := app.BuildFleet()
	if err != nil {
		t.Fatal(err)
	}
	second, err := app.BuildFleet()
	if err != nil {
		t.Fatal(err)
	}
	vehicle, err := domain.NewBicycle("Велосипед", "B-001", 6, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := first.AddVehicle(vehicle); err != nil {
		t.Fatal(err)
	}
	if got := second.Report(); got != (service.FleetReport{}) {
		t.Errorf("second fleet shares state with the first: %+v", got)
	}
	if err := second.AddVehicle(vehicle); err != nil {
		t.Errorf("independent fleet must have its own inventory numbers: %v", err)
	}
}
