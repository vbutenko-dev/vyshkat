package domain

import (
	"errors"
	"math"
	"strings"
)

var (
	// ErrEmptyName означает, что название пусто после удаления пробелов.
	ErrEmptyName = errors.New("название не должно быть пустым")
	// ErrEmptyInventoryNumber означает, что инвентарный номер пуст.
	ErrEmptyInventoryNumber = errors.New("инвентарный номер не должен быть пустым")
	// ErrInvalidBeginnerScore означает, что оценка находится вне диапазона 1–10.
	ErrInvalidBeginnerScore = errors.New("оценка простоты должна быть от 1 до 10")
	// ErrInvalidDailyEnergy означает отрицательный, бесконечный расход или NaN.
	ErrInvalidDailyEnergy = errors.New("суточное энергопотребление должно быть конечным неотрицательным числом")
)

// inventoryData хранит общие данные транспорта и вещей.
type inventoryData struct {
	name            string
	inventoryNumber string
}

func newInventoryData(name, inventoryNumber string) (inventoryData, error) {
	name = strings.TrimSpace(name)
	inventoryNumber = strings.TrimSpace(inventoryNumber)
	if name == "" {
		return inventoryData{}, ErrEmptyName
	}
	if inventoryNumber == "" {
		return inventoryData{}, ErrEmptyInventoryNumber
	}
	return inventoryData{name: name, inventoryNumber: inventoryNumber}, nil
}

func (data inventoryData) Name() string {
	return data.name
}

func (data inventoryData) InventoryNumber() string {
	return data.inventoryNumber
}

// vehicleData добавляет к инвентарным данным характеристики транспорта.
type vehicleData struct {
	inventoryData
	beginnerScore int
	operational   bool
}

func newVehicleData(name, inventoryNumber string, beginnerScore int, operational bool) (vehicleData, error) {
	inventory, err := newInventoryData(name, inventoryNumber)
	if err != nil {
		return vehicleData{}, err
	}
	if beginnerScore < 1 || beginnerScore > 10 {
		return vehicleData{}, ErrInvalidBeginnerScore
	}
	return vehicleData{
		inventoryData: inventory,
		beginnerScore: beginnerScore,
		operational:   operational,
	}, nil
}

func (data vehicleData) BeginnerScore() int {
	return data.beginnerScore
}

func (data vehicleData) IsOperational() bool {
	return data.operational
}

// energyData применяется только к потребителям электрической энергии.
type energyData struct {
	dailyEnergyKWh float64
}

func newEnergyData(dailyEnergyKWh float64) (energyData, error) {
	if dailyEnergyKWh < 0 || math.IsNaN(dailyEnergyKWh) || math.IsInf(dailyEnergyKWh, 0) {
		return energyData{}, ErrInvalidDailyEnergy
	}
	return energyData{dailyEnergyKWh: dailyEnergyKWh}, nil
}

func (data energyData) DailyEnergyKWh() float64 {
	return data.dailyEnergyKWh
}
