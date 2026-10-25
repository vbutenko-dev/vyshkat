package console

import "fmt"

func (c *Console) printReport() {
	report := c.fleet.Report()
	fmt.Fprintf(c.output, "Транспорт: %d\nВещи: %d\nВсего: %d\n", report.VehicleCount, report.ItemCount, report.TotalCount)
	fmt.Fprintf(c.output, "Энергия транспорта: %.3f кВт·ч/сутки\n", report.VehicleEnergyKWh)
	fmt.Fprintf(c.output, "Энергия вещей: %.3f кВт·ч/сутки\n", report.ItemEnergyKWh)
	fmt.Fprintf(c.output, "Общая энергия: %.3f кВт·ч/сутки\n", report.TotalEnergyKWh)
}

func (c *Console) printBeginners() {
	vehicles := c.fleet.BeginnerVehicles()
	if len(vehicles) == 0 {
		fmt.Fprintln(c.output, "Подходящего транспорта нет.")
		return
	}
	fmt.Fprintln(c.output, "Транспорт для новичков:")
	for _, vehicle := range vehicles {
		fmt.Fprintf(c.output, "%s — %s (простота %d)\n", vehicle.InventoryNumber(), vehicle.Name(), vehicle.BeginnerScore())
	}
}

func (c *Console) printInventory() {
	items := c.fleet.Inventory()
	if len(items) == 0 {
		fmt.Fprintln(c.output, "Инвентарь пуст.")
		return
	}
	fmt.Fprintln(c.output, "Инвентарный номер — Наименование:")
	for _, item := range items {
		fmt.Fprintf(c.output, "%s — %s\n", item.InventoryNumber(), item.Name())
	}
}
