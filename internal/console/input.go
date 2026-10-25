package console

import (
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"vyshkat/internal/domain"
)

func (c *Console) readLine(prompt string) (string, error) {
	fmt.Fprint(c.output, prompt)
	if c.input.Scan() {
		return strings.TrimSpace(c.input.Text()), nil
	}
	if err := c.input.Err(); err != nil {
		return "", err
	}
	return "", io.EOF
}

func (c *Console) readChoice(prompt string, options ...string) (string, error) {
	for {
		value, err := c.readLine(prompt)
		if err != nil {
			return "", err
		}
		value = strings.ToLower(value)
		if slices.Contains(options, value) {
			return value, nil
		}
		fmt.Fprintf(c.output, "Выберите один из вариантов: %s.\n", strings.Join(options, ", "))
	}
}

func (c *Console) readInt(prompt string) (int, error) {
	for {
		text, err := c.readLine(prompt)
		if err != nil {
			return 0, err
		}
		value, err := strconv.Atoi(text)
		if err == nil {
			return value, nil
		}
		fmt.Fprintln(c.output, "Введите целое число.")
	}
}

func (c *Console) readEnergy() (float64, error) {
	for {
		text, err := c.readLine("Энергия за сутки (кВт·ч, например 1.5): ")
		if err != nil {
			return 0, err
		}
		value, err := strconv.ParseFloat(text, 64)
		if err == nil {
			return value, nil
		}
		fmt.Fprintln(c.output, "Введите число с точкой, например 1.5.")
	}
}

func (c *Console) readInventoryFields() (string, string, error) {
	name, err := c.readLine("Наименование: ")
	if err != nil {
		return "", "", err
	}
	number, err := c.readLine("Инвентарный номер: ")
	if err != nil {
		return "", "", err
	}
	return name, number, nil
}

func (c *Console) readVehicle() (domain.Vehicle, error) {
	kind, err := c.readChoice("Вид (1 — электросамокат, 2 — велосипед, 3 — электровелосипед): ", "1", "2", "3")
	if err != nil {
		return nil, err
	}
	name, number, err := c.readInventoryFields()
	if err != nil {
		return nil, err
	}
	score, err := c.readInt("Простота для новичка (1–10): ")
	if err != nil {
		return nil, err
	}
	var energy float64
	if kind != "2" {
		energy, err = c.readEnergy()
		if err != nil {
			return nil, err
		}
	}
	condition, err := c.readChoice("Исправен (да/нет): ", "да", "нет")
	if err != nil {
		return nil, err
	}
	operational := condition == "да"
	if kind == "1" {
		return domain.NewScooter(name, number, score, energy, operational)
	}
	if kind == "2" {
		return domain.NewBicycle(name, number, score, operational)
	}
	return domain.NewElectricBicycle(name, number, score, energy, operational)
}

func (c *Console) readItem() (domain.InventoryItem, error) {
	kind, err := c.readChoice("Вид (1 — шлем, 2 — зарядная станция): ", "1", "2")
	if err != nil {
		return nil, err
	}
	name, number, err := c.readInventoryFields()
	if err != nil {
		return nil, err
	}
	if kind == "1" {
		return domain.NewHelmet(name, number)
	}
	energy, err := c.readEnergy()
	if err != nil {
		return nil, err
	}
	return domain.NewChargingStation(name, number, energy)
}
