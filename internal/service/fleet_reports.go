package service

import "vyshkat/internal/domain"

const beginnerScoreThreshold = 6

// FleetReport содержит количества объектов и расход энергии за сутки в кВт·ч.
type FleetReport struct {
	VehicleCount int
	ItemCount    int
	TotalCount   int

	VehicleEnergyKWh float64
	ItemEnergyKWh    float64
	TotalEnergyKWh   float64
}

// Report вычисляет сводку по транспорту и вещам, уже поставленным на баланс.
func (fleet *FleetService) Report() FleetReport {
	report := FleetReport{
		VehicleCount: len(fleet.vehicles),
		ItemCount:    len(fleet.inventory) - len(fleet.vehicles),
		TotalCount:   len(fleet.inventory),
	}
	for _, item := range fleet.inventory {
		consumer, consumesEnergy := item.(domain.EnergyConsumer)
		if !consumesEnergy {
			continue
		}
		if _, isVehicle := item.(domain.Vehicle); isVehicle {
			report.VehicleEnergyKWh += consumer.DailyEnergyKWh()
		} else {
			report.ItemEnergyKWh += consumer.DailyEnergyKWh()
		}
	}
	report.TotalEnergyKWh = report.VehicleEnergyKWh + report.ItemEnergyKWh
	return report
}

// BeginnerVehicles возвращает новый список принятого транспорта с простотой ≥ 6.
// Порядок соответствует порядку приёмки; вещи в этот список не входят.
func (fleet *FleetService) BeginnerVehicles() []domain.Vehicle {
	vehicles := make([]domain.Vehicle, 0)
	for _, vehicle := range fleet.vehicles {
		if vehicle.BeginnerScore() >= beginnerScoreThreshold {
			vehicles = append(vehicles, vehicle)
		}
	}
	return vehicles
}
