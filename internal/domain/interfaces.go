// Package domain описывает объекты и правила предметной области «ВышКат».
package domain

// InventoryItem — объект на балансе с названием и инвентарным номером.
type InventoryItem interface {
	Name() string
	InventoryNumber() string
}

// EnergyConsumer сообщает суточный расход электрической энергии в кВт·ч.
// Значение должно быть конечным и неотрицательным; нулевой расход допустим.
type EnergyConsumer interface {
	DailyEnergyKWh() float64
}

// Vehicle — прокатный транспорт, который можно передать на техосмотр.
type Vehicle interface {
	InventoryItem

	// BeginnerScore возвращает оценку простоты от 1 до 10 включительно.
	BeginnerScore() int

	// IsOperational сообщает исправность, а решение о приёмке принимает техосмотр.
	IsOperational() bool
}
