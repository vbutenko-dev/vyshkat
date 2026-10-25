package service

import (
	"math"
	"reflect"
	"strings"

	"vyshkat/internal/domain"
)

// Интерфейс может содержать nil-указатель и при этом не равняться nil.
// Проверяем это до вызова методов, чтобы не получить панику.
func isNil(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}

// Приёмка проверяет и объекты, созданные в обход конструкторов domain.
func validateInventoryItem(item domain.InventoryItem) error {
	if isNil(item) {
		return ErrMissingInventoryItem
	}
	if strings.TrimSpace(item.Name()) == "" {
		return domain.ErrEmptyName
	}
	if strings.TrimSpace(item.InventoryNumber()) == "" {
		return domain.ErrEmptyInventoryNumber
	}
	if consumer, ok := item.(domain.EnergyConsumer); ok {
		energy := consumer.DailyEnergyKWh()
		if energy < 0 || math.IsNaN(energy) || math.IsInf(energy, 0) {
			return domain.ErrInvalidDailyEnergy
		}
	}
	return nil
}
