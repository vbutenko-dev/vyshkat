// Package service содержит операции над моделями предметной области «ВышКат».
package service

import "vyshkat/internal/domain"

// InspectionResult — решение техосмотра.
// При отказе Reason содержит причину; при принятии Reason пуст.
type InspectionResult struct {
	Accepted bool
	Reason   string
}

// ServiceCenter определяет контракт техосмотра.
// Inspect ожидает ненулевой транспорт с корректными характеристиками.
// Проверка входных данных и постановка на баланс относятся к сервису парка.
type ServiceCenter interface {
	Inspect(vehicle domain.Vehicle) InspectionResult
}

// BasicServiceCenter принимает исправный транспорт и отклоняет неисправный.
// Он не хранит парк и не изменяет переданный транспорт.
type BasicServiceCenter struct{}

// NewBasicServiceCenter создаёт сервис техосмотра.
func NewBasicServiceCenter() *BasicServiceCenter {
	return &BasicServiceCenter{}
}

// Inspect принимает решение по признаку исправности транспорта.
func (center *BasicServiceCenter) Inspect(vehicle domain.Vehicle) InspectionResult {
	if !vehicle.IsOperational() {
		return InspectionResult{
			Accepted: false,
			Reason:   "транспорт неисправен",
		}
	}
	return InspectionResult{Accepted: true}
}

var _ ServiceCenter = (*BasicServiceCenter)(nil)
