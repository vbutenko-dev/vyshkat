package service

import (
	"errors"
	"fmt"
	"strings"

	"vyshkat/internal/domain"
)

var (
	ErrMissingServiceCenter      = errors.New("сервис техосмотра не передан")
	ErrMissingInventoryItem      = errors.New("объект не передан")
	ErrDuplicateInventoryNumber  = errors.New("инвентарный номер уже занят")
	ErrInspectionRejected        = errors.New("транспорт не прошёл техосмотр")
	ErrVehicleRequiresInspection = errors.New("транспорт нужно добавлять через AddVehicle с техосмотром")
)

// FleetService хранит парк в памяти и использует переданный техосмотр.
// Создавайте сервис через NewFleetService.
type FleetService struct {
	center           ServiceCenter
	vehicles         []domain.Vehicle
	inventory        []domain.InventoryItem
	inventoryNumbers map[string]struct{}
}

// NewFleetService получает зависимость извне, не выбирая реализацию техосмотра.
func NewFleetService(center ServiceCenter) (*FleetService, error) {
	if isNil(center) {
		return nil, ErrMissingServiceCenter
	}
	return &FleetService{
		center:           center,
		inventoryNumbers: make(map[string]struct{}),
	}, nil
}

// AddVehicle добавляет транспорт только после положительного решения техосмотра.
// При ошибке или отказе содержимое парка не меняется.
func (fleet *FleetService) AddVehicle(vehicle domain.Vehicle) error {
	if err := validateInventoryItem(vehicle); err != nil {
		return err
	}
	if score := vehicle.BeginnerScore(); score < 1 || score > 10 {
		return domain.ErrInvalidBeginnerScore
	}
	number := strings.TrimSpace(vehicle.InventoryNumber())
	if _, exists := fleet.inventoryNumbers[number]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicateInventoryNumber, number)
	}

	result := fleet.center.Inspect(vehicle)
	if !result.Accepted {
		if reason := strings.TrimSpace(result.Reason); reason != "" {
			return fmt.Errorf("%w: %s", ErrInspectionRejected, reason)
		}
		return ErrInspectionRejected
	}

	// Изменяем состояние только после завершения всех проверок.
	fleet.vehicles = append(fleet.vehicles, vehicle)
	fleet.inventory = append(fleet.inventory, vehicle)
	fleet.inventoryNumbers[number] = struct{}{}
	return nil
}

// AddItem добавляет вещь без техосмотра и проверяет общий реестр номеров.
// Транспорт нельзя добавить этим методом, обходя AddVehicle.
func (fleet *FleetService) AddItem(item domain.InventoryItem) error {
	if err := validateInventoryItem(item); err != nil {
		return err
	}
	if _, isVehicle := item.(domain.Vehicle); isVehicle {
		return ErrVehicleRequiresInspection
	}
	number := strings.TrimSpace(item.InventoryNumber())
	if _, exists := fleet.inventoryNumbers[number]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicateInventoryNumber, number)
	}
	fleet.inventory = append(fleet.inventory, item)
	fleet.inventoryNumbers[number] = struct{}{}
	return nil
}

// Vehicles возвращает копию списка принятого транспорта.
func (fleet *FleetService) Vehicles() []domain.Vehicle {
	vehicles := make([]domain.Vehicle, len(fleet.vehicles))
	copy(vehicles, fleet.vehicles)
	return vehicles
}

// Inventory возвращает копию общего списка транспорта и вещей в порядке приёмки.
func (fleet *FleetService) Inventory() []domain.InventoryItem {
	items := make([]domain.InventoryItem, len(fleet.inventory))
	copy(items, fleet.inventory)
	return items
}
