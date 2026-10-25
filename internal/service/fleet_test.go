package service_test

import (
	"errors"
	"math"
	"strings"
	"testing"

	"vyshkat/internal/domain"
	"vyshkat/internal/service"
)

// Fake задаёт решение и запоминает обращения, не проверяя исправность.
type fakeServiceCenter struct {
	result    service.InspectionResult
	calls     int
	inspected domain.Vehicle
}

func (fake *fakeServiceCenter) Inspect(vehicle domain.Vehicle) service.InspectionResult {
	fake.calls++
	fake.inspected = vehicle
	return fake.result
}

var _ service.ServiceCenter = (*fakeServiceCenter)(nil)

func newTestFleet(t *testing.T, center service.ServiceCenter) *service.FleetService {
	t.Helper()
	fleet, err := service.NewFleetService(center)
	if err != nil {
		t.Fatalf("NewFleetService: %v", err)
	}
	return fleet
}

func assertFleetSize(t *testing.T, fleet *service.FleetService, vehicles, inventory int) {
	t.Helper()
	if got := len(fleet.Vehicles()); got != vehicles {
		t.Fatalf("vehicle count = %d, want %d", got, vehicles)
	}
	if got := len(fleet.Inventory()); got != inventory {
		t.Fatalf("inventory count = %d, want %d", got, inventory)
	}
}

func TestNewFleetServiceRequiresInspectionDependency(t *testing.T) {
	var typedNil *fakeServiceCenter
	for _, center := range []service.ServiceCenter{nil, typedNil} {
		fleet, err := service.NewFleetService(center)
		if !errors.Is(err, service.ErrMissingServiceCenter) {
			t.Errorf("error = %v, want %v", err, service.ErrMissingServiceCenter)
		}
		if fleet != nil {
			t.Error("missing dependency must not produce a fleet")
		}
	}
}

func TestFleetFollowsInjectedInspectionDecision(t *testing.T) {
	cases := []struct {
		name        string
		operational bool
		result      service.InspectionResult
	}{
		{"healthy_accepted", true, service.InspectionResult{Accepted: true}},
		{"healthy_rejected", true, service.InspectionResult{Reason: "отказ тестового техосмотра"}},
		{"broken_accepted_by_fake", false, service.InspectionResult{Accepted: true}},
		{"broken_rejected", false, service.InspectionResult{Reason: "транспорт неисправен"}},
		{"rejection_without_reason", true, service.InspectionResult{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fake := &fakeServiceCenter{result: tc.result}
			fleet := newTestFleet(t, fake)
			assertFleetSize(t, fleet, 0, 0)
			vehicle, err := domain.NewScooter("Самокат", "S-001", 6, 1.5, tc.operational)
			if err != nil {
				t.Fatal(err)
			}
			err = fleet.AddVehicle(vehicle)
			if tc.result.Accepted {
				if err != nil {
					t.Fatalf("AddVehicle: %v", err)
				}
				assertFleetSize(t, fleet, 1, 1)
				if fleet.Vehicles()[0] != vehicle || fleet.Inventory()[0] != vehicle {
					t.Error("fleet must contain the inspected vehicle")
				}
			} else {
				if !errors.Is(err, service.ErrInspectionRejected) {
					t.Fatalf("error = %v, want %v", err, service.ErrInspectionRejected)
				}
				if !strings.Contains(err.Error(), tc.result.Reason) {
					t.Errorf("rejection reason was lost: %v", err)
				}
				assertFleetSize(t, fleet, 0, 0)
			}
			if fake.calls != 1 || fake.inspected != vehicle {
				t.Errorf("inspection calls = %d; expected one call with the added vehicle", fake.calls)
			}
		})
	}
}

func TestFleetRejectsDuplicateVehicleBeforeInspection(t *testing.T) {
	fake := &fakeServiceCenter{result: service.InspectionResult{Accepted: true}}
	fleet := newTestFleet(t, fake)
	first, err := domain.NewBicycle("Первый", "B-001", 6, true)
	if err != nil {
		t.Fatal(err)
	}
	second, err := domain.NewScooter("Второй", "B-001", 8, 1.5, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := fleet.AddVehicle(first); err != nil {
		t.Fatal(err)
	}
	if err := fleet.AddVehicle(second); !errors.Is(err, service.ErrDuplicateInventoryNumber) {
		t.Fatalf("error = %v, want %v", err, service.ErrDuplicateInventoryNumber)
	}
	assertFleetSize(t, fleet, 1, 1)
	if fleet.Vehicles()[0] != first || fake.calls != 1 {
		t.Error("duplicate must not replace the original or trigger another inspection")
	}
}

func TestFleetRejectionDoesNotReserveInventoryNumber(t *testing.T) {
	fake := &fakeServiceCenter{result: service.InspectionResult{Accepted: true}}
	fleet := newTestFleet(t, fake)
	existing, err := domain.NewBicycle("Уже в парке", "B-BASE", 6, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := fleet.AddVehicle(existing); err != nil {
		t.Fatal(err)
	}
	fake.result = service.InspectionResult{Reason: "отказ"}
	vehicle, err := domain.NewBicycle("Велосипед", "B-001", 6, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := fleet.AddVehicle(vehicle); !errors.Is(err, service.ErrInspectionRejected) {
		t.Fatalf("first admission: %v", err)
	}
	assertFleetSize(t, fleet, 1, 1)
	if fleet.Vehicles()[0] != existing || fleet.Inventory()[0] != existing {
		t.Fatal("rejection must preserve the existing fleet")
	}
	fake.result = service.InspectionResult{Accepted: true}
	if err := fleet.AddVehicle(vehicle); err != nil {
		t.Fatalf("admission after approval: %v", err)
	}
	assertFleetSize(t, fleet, 2, 2)
	if fake.calls != 3 {
		t.Errorf("inspection calls = %d, want 3", fake.calls)
	}
}

func TestFleetSharesInventoryNumbersBetweenVehiclesAndItems(t *testing.T) {
	for _, vehicleFirst := range []bool{false, true} {
		t.Run(map[bool]string{false: "item_first", true: "vehicle_first"}[vehicleFirst], func(t *testing.T) {
			fake := &fakeServiceCenter{result: service.InspectionResult{Accepted: true}}
			fleet := newTestFleet(t, fake)
			vehicle, err := domain.NewBicycle("Велосипед", "INV-001", 6, true)
			if err != nil {
				t.Fatal(err)
			}
			helmet, err := domain.NewHelmet("Шлем", "INV-001")
			if err != nil {
				t.Fatal(err)
			}
			if vehicleFirst {
				if err := fleet.AddVehicle(vehicle); err != nil {
					t.Fatal(err)
				}
				if err := fleet.AddItem(helmet); !errors.Is(err, service.ErrDuplicateInventoryNumber) {
					t.Fatalf("AddItem: %v", err)
				}
				assertFleetSize(t, fleet, 1, 1)
				if fake.calls != 1 {
					t.Errorf("inspection calls = %d, want 1", fake.calls)
				}
			} else {
				if err := fleet.AddItem(helmet); err != nil {
					t.Fatal(err)
				}
				if err := fleet.AddVehicle(vehicle); !errors.Is(err, service.ErrDuplicateInventoryNumber) {
					t.Fatalf("AddVehicle: %v", err)
				}
				if err := fleet.AddItem(helmet); !errors.Is(err, service.ErrDuplicateInventoryNumber) {
					t.Fatalf("duplicate AddItem: %v", err)
				}
				assertFleetSize(t, fleet, 0, 1)
				if fake.calls != 0 {
					t.Errorf("inspection calls = %d, want 0", fake.calls)
				}
			}
		})
	}
}

func TestFleetAddsItemsWithoutInspectionAndPreventsBypass(t *testing.T) {
	fake := &fakeServiceCenter{result: service.InspectionResult{Reason: "отказ"}}
	fleet := newTestFleet(t, fake)
	helmet, err := domain.NewHelmet("Шлем", "H-001")
	if err != nil {
		t.Fatal(err)
	}
	station, err := domain.NewChargingStation("Станция", "C-001", 1.5)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range []domain.InventoryItem{helmet, station} {
		if err := fleet.AddItem(item); err != nil {
			t.Fatal(err)
		}
	}
	vehicle, err := domain.NewBicycle("Велосипед", "B-001", 6, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := fleet.AddItem(vehicle); !errors.Is(err, service.ErrVehicleRequiresInspection) {
		t.Fatalf("bypass attempt: %v", err)
	}
	assertFleetSize(t, fleet, 0, 2)
	if fake.calls != 0 {
		t.Errorf("inspection calls = %d, want 0", fake.calls)
	}
}

// Типы позволяют передать некорректные данные в обход конструкторов domain.
type uncheckedVehicle struct {
	name   string
	number string
	score  int
}

func (v uncheckedVehicle) Name() string            { return v.name }
func (v uncheckedVehicle) InventoryNumber() string { return v.number }
func (v uncheckedVehicle) BeginnerScore() int      { return v.score }
func (v uncheckedVehicle) IsOperational() bool     { return true }

type uncheckedEnergyVehicle struct {
	uncheckedVehicle
	energy float64
}

func (v uncheckedEnergyVehicle) DailyEnergyKWh() float64 { return v.energy }

func TestFleetRejectsInvalidVehiclesBeforeInspection(t *testing.T) {
	valid := uncheckedVehicle{name: "Транспорт", number: "INV-001", score: 6}
	var nilScooter *domain.Scooter
	cases := []struct {
		name    string
		vehicle domain.Vehicle
		wantErr error
	}{
		{"nil", nil, service.ErrMissingInventoryItem},
		{"typed_nil", nilScooter, service.ErrMissingInventoryItem},
		{"zero_value", &domain.Scooter{}, domain.ErrEmptyName},
		{"blank_name", uncheckedVehicle{" \t", "INV-001", 6}, domain.ErrEmptyName},
		{"blank_number", uncheckedVehicle{"Транспорт", " \t", 6}, domain.ErrEmptyInventoryNumber},
		{"zero_score", uncheckedVehicle{"Транспорт", "INV-001", 0}, domain.ErrInvalidBeginnerScore},
		{"high_score", uncheckedVehicle{"Транспорт", "INV-001", 11}, domain.ErrInvalidBeginnerScore},
		{"negative_energy", uncheckedEnergyVehicle{valid, -1}, domain.ErrInvalidDailyEnergy},
		{"nan_energy", uncheckedEnergyVehicle{valid, math.NaN()}, domain.ErrInvalidDailyEnergy},
		{"infinite_energy", uncheckedEnergyVehicle{valid, math.Inf(1)}, domain.ErrInvalidDailyEnergy},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fake := &fakeServiceCenter{result: service.InspectionResult{Accepted: true}}
			fleet := newTestFleet(t, fake)
			if err := fleet.AddVehicle(tc.vehicle); !errors.Is(err, tc.wantErr) {
				t.Fatalf("error = %v, want %v", err, tc.wantErr)
			}
			assertFleetSize(t, fleet, 0, 0)
			if fake.calls != 0 {
				t.Errorf("inspection calls = %d, want 0", fake.calls)
			}
		})
	}
}

func TestFleetRejectsMissingAndInvalidItems(t *testing.T) {
	var nilHelmet *domain.Helmet
	fake := &fakeServiceCenter{}
	fleet := newTestFleet(t, fake)
	for _, item := range []domain.InventoryItem{nil, nilHelmet} {
		if err := fleet.AddItem(item); !errors.Is(err, service.ErrMissingInventoryItem) {
			t.Errorf("error = %v, want %v", err, service.ErrMissingInventoryItem)
		}
	}
	if err := fleet.AddItem(&domain.Helmet{}); !errors.Is(err, domain.ErrEmptyName) {
		t.Errorf("zero-value item: %v", err)
	}
	assertFleetSize(t, fleet, 0, 0)
}

func TestFleetSupportsNewVehicleTypesAndNormalizesNumberLookup(t *testing.T) {
	fake := &fakeServiceCenter{result: service.InspectionResult{Accepted: true}}
	fleet := newTestFleet(t, fake)
	vehicle := uncheckedVehicle{name: "Новый вид транспорта", number: " INV-001 ", score: 6}
	if err := fleet.AddVehicle(vehicle); err != nil {
		t.Fatalf("new vehicle type: %v", err)
	}
	helmet, err := domain.NewHelmet("Шлем", "INV-001")
	if err != nil {
		t.Fatal(err)
	}
	if err := fleet.AddItem(helmet); !errors.Is(err, service.ErrDuplicateInventoryNumber) {
		t.Fatalf("normalized duplicate: %v", err)
	}
	assertFleetSize(t, fleet, 1, 1)
}

func TestFleetReturnsIndependentLists(t *testing.T) {
	fleet := newTestFleet(t, service.NewBasicServiceCenter())
	vehicle, err := domain.NewBicycle("Велосипед", "B-001", 6, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := fleet.AddVehicle(vehicle); err != nil {
		t.Fatal(err)
	}
	vehicles := fleet.Vehicles()
	inventory := fleet.Inventory()
	vehicles[0] = nil
	inventory[0] = nil
	if fleet.Vehicles()[0] != vehicle || fleet.Inventory()[0] != vehicle {
		t.Error("changing returned lists must not change the fleet")
	}
}
