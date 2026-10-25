package domain_test

import (
	"errors"
	"math"
	"testing"

	"vyshkat/internal/domain"
)

// Эти адаптеры позволяют проверить общие правила через все публичные конструкторы.
func inventoryConstructors() []struct {
	name   string
	create func(string, string) (domain.InventoryItem, error)
} {
	return []struct {
		name   string
		create func(string, string) (domain.InventoryItem, error)
	}{
		{"scooter", func(name, number string) (domain.InventoryItem, error) {
			return domain.NewScooter(name, number, 6, 1.5, true)
		}},
		{"bicycle", func(name, number string) (domain.InventoryItem, error) {
			return domain.NewBicycle(name, number, 6, true)
		}},
		{"electric_bicycle", func(name, number string) (domain.InventoryItem, error) {
			return domain.NewElectricBicycle(name, number, 6, 1.5, true)
		}},
		{"helmet", func(name, number string) (domain.InventoryItem, error) {
			return domain.NewHelmet(name, number)
		}},
		{"charging_station", func(name, number string) (domain.InventoryItem, error) {
			return domain.NewChargingStation(name, number, 1.5)
		}},
	}
}

func TestConstructorsNormalizeInventoryData(t *testing.T) {
	for _, constructor := range inventoryConstructors() {
		t.Run(constructor.name, func(t *testing.T) {
			item, err := constructor.create(" \tТехника ВышКат\n", "\u00a0INV-001 \t")
			if err != nil {
				t.Fatalf("create: %v", err)
			}
			if got := item.Name(); got != "Техника ВышКат" {
				t.Errorf("Name() = %q, want %q", got, "Техника ВышКат")
			}
			if got := item.InventoryNumber(); got != "INV-001" {
				t.Errorf("InventoryNumber() = %q, want %q", got, "INV-001")
			}
		})
	}
}

func TestConstructorsRejectEmptyInventoryData(t *testing.T) {
	cases := []struct {
		name      string
		itemName  string
		number    string
		wantError error
	}{
		{"empty_name", "", "INV-001", domain.ErrEmptyName},
		{"blank_name", " \t\n\u00a0", "INV-001", domain.ErrEmptyName},
		{"empty_number", "Техника", "", domain.ErrEmptyInventoryNumber},
		{"blank_number", "Техника", " \t\n\u00a0", domain.ErrEmptyInventoryNumber},
	}
	for _, constructor := range inventoryConstructors() {
		for _, tc := range cases {
			t.Run(constructor.name+"/"+tc.name, func(t *testing.T) {
				_, err := constructor.create(tc.itemName, tc.number)
				if !errors.Is(err, tc.wantError) {
					t.Fatalf("create error = %v, want %v", err, tc.wantError)
				}
			})
		}
	}
}

func TestVehiclesValidateBeginnerScoreAndPreserveCondition(t *testing.T) {
	constructors := []struct {
		name   string
		create func(int, bool) (domain.Vehicle, error)
	}{
		{"scooter", func(score int, operational bool) (domain.Vehicle, error) {
			return domain.NewScooter("Самокат", "S-001", score, 1.5, operational)
		}},
		{"bicycle", func(score int, operational bool) (domain.Vehicle, error) {
			return domain.NewBicycle("Велосипед", "B-001", score, operational)
		}},
		{"electric_bicycle", func(score int, operational bool) (domain.Vehicle, error) {
			return domain.NewElectricBicycle("Электровелосипед", "E-001", score, 1.5, operational)
		}},
	}
	cases := []struct {
		name  string
		score int
		valid bool
	}{
		{"negative", -1, false},
		{"zero", 0, false},
		{"lower_boundary", 1, true},
		{"below_beginner_threshold", 5, true},
		{"beginner_threshold", 6, true},
		{"upper_boundary", 10, true},
		{"above_upper_boundary", 11, false},
	}
	for _, constructor := range constructors {
		for _, tc := range cases {
			t.Run(constructor.name+"/"+tc.name, func(t *testing.T) {
				for _, operational := range []bool{false, true} {
					vehicle, err := constructor.create(tc.score, operational)
					if !tc.valid {
						if !errors.Is(err, domain.ErrInvalidBeginnerScore) {
							t.Fatalf("score %d: error = %v, want %v", tc.score, err, domain.ErrInvalidBeginnerScore)
						}
						continue
					}
					if err != nil {
						t.Fatalf("score %d, operational %t: %v", tc.score, operational, err)
					}
					if got := vehicle.BeginnerScore(); got != tc.score {
						t.Errorf("BeginnerScore() = %d, want %d", got, tc.score)
					}
					if got := vehicle.IsOperational(); got != operational {
						t.Errorf("IsOperational() = %t, want %t", got, operational)
					}
				}
			})
		}
	}
}

func TestEnergyConsumersValidateDailyEnergy(t *testing.T) {
	constructors := []struct {
		name   string
		create func(float64) (domain.EnergyConsumer, error)
	}{
		{"scooter", func(energy float64) (domain.EnergyConsumer, error) {
			return domain.NewScooter("Самокат", "S-001", 6, energy, true)
		}},
		{"electric_bicycle", func(energy float64) (domain.EnergyConsumer, error) {
			return domain.NewElectricBicycle("Электровелосипед", "E-001", 6, energy, true)
		}},
		{"charging_station", func(energy float64) (domain.EnergyConsumer, error) {
			return domain.NewChargingStation("Станция", "C-001", energy)
		}},
	}
	cases := []struct {
		name   string
		energy float64
		valid  bool
	}{
		{"zero", 0, true},
		{"positive", 1.5, true},
		{"negative", -0.1, false},
		{"nan", math.NaN(), false},
		{"positive_infinity", math.Inf(1), false},
		{"negative_infinity", math.Inf(-1), false},
	}
	for _, constructor := range constructors {
		for _, tc := range cases {
			t.Run(constructor.name+"/"+tc.name, func(t *testing.T) {
				consumer, err := constructor.create(tc.energy)
				if !tc.valid {
					if !errors.Is(err, domain.ErrInvalidDailyEnergy) {
						t.Fatalf("energy %v: error = %v, want %v", tc.energy, err, domain.ErrInvalidDailyEnergy)
					}
					return
				}
				if err != nil {
					t.Fatalf("energy %v: %v", tc.energy, err)
				}
				if got := consumer.DailyEnergyKWh(); got != tc.energy {
					t.Errorf("DailyEnergyKWh() = %v, want %v", got, tc.energy)
				}
			})
		}
	}
}

func TestBicycleAndHelmetHaveOnlyTheirCapabilities(t *testing.T) {
	bicycle, err := domain.NewBicycle("Велосипед", "B-001", 6, true)
	if err != nil {
		t.Fatal(err)
	}
	if _, consumesEnergy := any(bicycle).(domain.EnergyConsumer); consumesEnergy {
		t.Error("ordinary bicycle must not implement EnergyConsumer")
	}
	helmet, err := domain.NewHelmet("Шлем", "H-001")
	if err != nil {
		t.Fatal(err)
	}
	if _, consumesEnergy := any(helmet).(domain.EnergyConsumer); consumesEnergy {
		t.Error("helmet must not implement EnergyConsumer")
	}
	if _, isVehicle := any(helmet).(domain.Vehicle); isVehicle {
		t.Error("helmet must not implement Vehicle")
	}
	station, err := domain.NewChargingStation("Станция", "C-001", 1.5)
	if err != nil {
		t.Fatal(err)
	}
	if _, isVehicle := any(station).(domain.Vehicle); isVehicle {
		t.Error("charging station must not implement Vehicle")
	}
}

func TestNewScooterReturnsNilOnInvalidInput(t *testing.T) {
	scooter, err := domain.NewScooter("Самокат", "S-001", 0, 1.5, true)
	if !errors.Is(err, domain.ErrInvalidBeginnerScore) {
		t.Fatalf("error = %v, want %v", err, domain.ErrInvalidBeginnerScore)
	}
	if scooter != nil {
		t.Fatal("invalid input must not produce a scooter")
	}
}
