package domain

// Scooter — электросамокат с инвентарными данными и расходом энергии.
// Создавайте его через NewScooter: нулевое значение не является корректной карточкой.
type Scooter struct {
	vehicleData
	energyData
}

// NewScooter проверяет характеристики и создаёт электросамокат.
// operational описывает состояние техники; неисправность допустима до техосмотра.
func NewScooter(name, inventoryNumber string, beginnerScore int, dailyEnergyKWh float64, operational bool) (*Scooter, error) {
	vehicle, err := newVehicleData(name, inventoryNumber, beginnerScore, operational)
	if err != nil {
		return nil, err
	}
	energy, err := newEnergyData(dailyEnergyKWh)
	if err != nil {
		return nil, err
	}
	return &Scooter{vehicleData: vehicle, energyData: energy}, nil
}

// Bicycle — обычный велосипед, который не потребляет электрическую энергию.
// Создавайте его через NewBicycle.
type Bicycle struct {
	vehicleData
}

// NewBicycle проверяет характеристики и создаёт обычный велосипед.
func NewBicycle(name, inventoryNumber string, beginnerScore int, operational bool) (*Bicycle, error) {
	vehicle, err := newVehicleData(name, inventoryNumber, beginnerScore, operational)
	if err != nil {
		return nil, err
	}
	return &Bicycle{vehicleData: vehicle}, nil
}

// ElectricBicycle — электровелосипед с расходом электрической энергии.
// Создавайте его через NewElectricBicycle.
type ElectricBicycle struct {
	vehicleData
	energyData
}

// NewElectricBicycle проверяет характеристики и создаёт электровелосипед.
func NewElectricBicycle(name, inventoryNumber string, beginnerScore int, dailyEnergyKWh float64, operational bool) (*ElectricBicycle, error) {
	vehicle, err := newVehicleData(name, inventoryNumber, beginnerScore, operational)
	if err != nil {
		return nil, err
	}
	energy, err := newEnergyData(dailyEnergyKWh)
	if err != nil {
		return nil, err
	}
	return &ElectricBicycle{vehicleData: vehicle, energyData: energy}, nil
}

// Компилятор проверит соответствие типов нужным интерфейсам.
var (
	_ Vehicle        = (*Scooter)(nil)
	_ EnergyConsumer = (*Scooter)(nil)
	_ Vehicle        = (*Bicycle)(nil)
	_ Vehicle        = (*ElectricBicycle)(nil)
	_ EnergyConsumer = (*ElectricBicycle)(nil)
)
