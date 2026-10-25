package service_test

import (
	"errors"
	"math"
	"slices"
	"testing"

	"vyshkat/internal/domain"
	"vyshkat/internal/service"
)

func TestEmptyFleetReports(t *testing.T) {
	fake := &fakeServiceCenter{result: service.InspectionResult{Accepted: true}}
	fleet := newTestFleet(t, fake)
	if got := fleet.Report(); got != (service.FleetReport{}) {
		t.Errorf("empty fleet report = %+v, want zero values", got)
	}
	if got := len(fleet.BeginnerVehicles()); got != 0 {
		t.Errorf("beginner vehicle count = %d, want 0", got)
	}
	if fake.calls != 0 {
		t.Error("reports must not trigger inspection")
	}
}

func TestFleetReportCountsVehiclesItemsAndEnergy(t *testing.T) {
	fleet := newTestFleet(t, service.NewBasicServiceCenter())
	scooter, err := domain.NewScooter("Самокат", "S-001", 6, 1.5, true)
	if err != nil {
		t.Fatal(err)
	}
	bicycle, err := domain.NewBicycle("Велосипед", "B-001", 5, true)
	if err != nil {
		t.Fatal(err)
	}
	electricBicycle, err := domain.NewElectricBicycle("Электровелосипед", "E-001", 8, 0.75, true)
	if err != nil {
		t.Fatal(err)
	}
	helmet, err := domain.NewHelmet("Шлем", "H-001")
	if err != nil {
		t.Fatal(err)
	}
	station, err := domain.NewChargingStation("Станция", "C-001", 2.25)
	if err != nil {
		t.Fatal(err)
	}
	for _, vehicle := range []domain.Vehicle{scooter, bicycle, electricBicycle} {
		if err := fleet.AddVehicle(vehicle); err != nil {
			t.Fatal(err)
		}
	}
	for _, item := range []domain.InventoryItem{helmet, station} {
		if err := fleet.AddItem(item); err != nil {
			t.Fatal(err)
		}
	}
	want := service.FleetReport{
		VehicleCount:     3,
		ItemCount:        2,
		TotalCount:       5,
		VehicleEnergyKWh: 2.25,
		ItemEnergyKWh:    2.25,
		TotalEnergyKWh:   4.5,
	}
	if got := fleet.Report(); got != want {
		t.Errorf("report = %+v, want %+v", got, want)
	}
	if got := fleet.Report(); got != want {
		t.Errorf("repeated report = %+v, want %+v", got, want)
	}
	assertFleetSize(t, fleet, 3, 5)
}

func TestFleetReportsExcludeRejectedAndDuplicateObjects(t *testing.T) {
	fake := &fakeServiceCenter{result: service.InspectionResult{Accepted: true}}
	fleet := newTestFleet(t, fake)
	accepted, err := domain.NewScooter("Принятый", "S-001", 6, 1.5, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := fleet.AddVehicle(accepted); err != nil {
		t.Fatal(err)
	}
	fake.result = service.InspectionResult{Reason: "отказ"}
	rejected, err := domain.NewScooter("Отклонённый", "S-002", 10, 100, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := fleet.AddVehicle(rejected); !errors.Is(err, service.ErrInspectionRejected) {
		t.Fatalf("rejected vehicle: %v", err)
	}
	duplicate, err := domain.NewScooter("Дубликат", "S-001", 10, 100, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := fleet.AddVehicle(duplicate); !errors.Is(err, service.ErrDuplicateInventoryNumber) {
		t.Fatalf("duplicate vehicle: %v", err)
	}
	station, err := domain.NewChargingStation("Станция", "C-001", 2.25)
	if err != nil {
		t.Fatal(err)
	}
	if err := fleet.AddItem(station); err != nil {
		t.Fatal(err)
	}
	if err := fleet.AddItem(station); !errors.Is(err, service.ErrDuplicateInventoryNumber) {
		t.Fatalf("duplicate station: %v", err)
	}
	want := service.FleetReport{
		VehicleCount:     1,
		ItemCount:        1,
		TotalCount:       2,
		VehicleEnergyKWh: 1.5,
		ItemEnergyKWh:    2.25,
		TotalEnergyKWh:   3.75,
	}
	if got := fleet.Report(); got != want {
		t.Errorf("report = %+v, want %+v", got, want)
	}
	vehicles := fleet.BeginnerVehicles()
	if len(vehicles) != 1 || vehicles[0] != accepted {
		t.Fatal("beginner list must contain only the accepted scooter")
	}
	if fake.calls != 2 {
		t.Errorf("inspection calls = %d, want 2; reporting must not inspect again", fake.calls)
	}
}

func TestBeginnerVehiclesThresholdOrderAndIndependentList(t *testing.T) {
	fleet := newTestFleet(t, service.NewBasicServiceCenter())
	cases := []struct {
		number string
		score  int
	}{
		{"B-1", 1},
		{"B-5", 5},
		{"B-6", 6},
		{"B-10", 10},
	}
	for _, tc := range cases {
		vehicle, err := domain.NewBicycle("Велосипед", tc.number, tc.score, true)
		if err != nil {
			t.Fatal(err)
		}
		if err := fleet.AddVehicle(vehicle); err != nil {
			t.Fatal(err)
		}
	}
	helmet, err := domain.NewHelmet("Шлем", "H-001")
	if err != nil {
		t.Fatal(err)
	}
	if err := fleet.AddItem(helmet); err != nil {
		t.Fatal(err)
	}
	vehicles := fleet.BeginnerVehicles()
	numbers := make([]string, 0, len(vehicles))
	for _, vehicle := range vehicles {
		numbers = append(numbers, vehicle.InventoryNumber())
	}
	if want := []string{"B-6", "B-10"}; !slices.Equal(numbers, want) {
		t.Fatalf("beginner numbers = %v, want %v", numbers, want)
	}
	vehicles[0] = nil
	if fleet.BeginnerVehicles()[0].InventoryNumber() != "B-6" {
		t.Error("changing the returned list must not change the fleet")
	}
	assertFleetSize(t, fleet, 4, 5)
}

func TestFleetReportUsesEnergyConsumerInterfaceForNewVehicleType(t *testing.T) {
	fake := &fakeServiceCenter{result: service.InspectionResult{Accepted: true}}
	fleet := newTestFleet(t, fake)
	vehicle := uncheckedEnergyVehicle{
		uncheckedVehicle: uncheckedVehicle{name: "Новый транспорт", number: "NEW-001", score: 6},
		energy:           0.1,
	}
	if err := fleet.AddVehicle(vehicle); err != nil {
		t.Fatal(err)
	}
	station, err := domain.NewChargingStation("Станция", "C-001", 0.2)
	if err != nil {
		t.Fatal(err)
	}
	if err := fleet.AddItem(station); err != nil {
		t.Fatal(err)
	}
	report := fleet.Report()
	if report.VehicleEnergyKWh != 0.1 || report.ItemEnergyKWh != 0.2 {
		t.Errorf("energy categories = %+v, want 0.1 for vehicles and 0.2 for items", report)
	}
	// Десятичные дроби float64 сравниваем с небольшим допуском.
	if math.Abs(report.TotalEnergyKWh-0.3) > 1e-12 {
		t.Errorf("total energy = %v, want approximately 0.3", report.TotalEnergyKWh)
	}
}
