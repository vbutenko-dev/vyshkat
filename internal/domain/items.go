package domain

// Helmet — шлем на балансе.
// Создавайте его через NewHelmet.
type Helmet struct {
	inventoryData
}

// NewHelmet проверяет инвентарные данные и создаёт шлем.
func NewHelmet(name, inventoryNumber string) (*Helmet, error) {
	inventory, err := newInventoryData(name, inventoryNumber)
	if err != nil {
		return nil, err
	}
	return &Helmet{inventoryData: inventory}, nil
}

// ChargingStation — зарядная станция на балансе с собственным расходом энергии.
// Создавайте её через NewChargingStation.
type ChargingStation struct {
	inventoryData
	energyData
}

// NewChargingStation проверяет характеристики и создаёт зарядную станцию.
func NewChargingStation(name, inventoryNumber string, dailyEnergyKWh float64) (*ChargingStation, error) {
	inventory, err := newInventoryData(name, inventoryNumber)
	if err != nil {
		return nil, err
	}
	energy, err := newEnergyData(dailyEnergyKWh)
	if err != nil {
		return nil, err
	}
	return &ChargingStation{inventoryData: inventory, energyData: energy}, nil
}

var (
	_ InventoryItem  = (*Helmet)(nil)
	_ InventoryItem  = (*ChargingStation)(nil)
	_ EnergyConsumer = (*ChargingStation)(nil)
)
