// Package console реализует консольный ввод и вывод приложения «ВышКат».
package console

import (
	"bufio"
	"errors"
	"fmt"
	"io"

	"vyshkat/internal/service"
)

const menu = `
ВышКат — учёт парка
1 — Добавить транспорт
2 — Добавить вещь
3 — Сводный отчёт
4 — Транспорт для новичков
5 — Инвентарь
0 — Выход
`

// Console связывает команды пользователя с готовым сервисом парка.
type Console struct {
	fleet  *service.FleetService
	input  *bufio.Scanner
	output io.Writer
}

// New получает готовый парк, источник ввода и место вывода.
// В приложении используются os.Stdin/os.Stdout, в тестах — строки и буфер.
func New(fleet *service.FleetService, input io.Reader, output io.Writer) *Console {
	return &Console{fleet: fleet, input: bufio.NewScanner(input), output: output}
}

// Run обрабатывает команды до выхода или завершения ввода.
func (c *Console) Run() error {
	for {
		fmt.Fprint(c.output, menu)
		command, err := c.readLine("Команда: ")
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("чтение ввода: %w", err)
		}
		switch command {
		case "0":
			fmt.Fprintln(c.output, "До свидания.")
			return nil
		case "1":
			err = c.addVehicle()
		case "2":
			err = c.addItem()
		case "3":
			c.printReport()
		case "4":
			c.printBeginners()
		case "5":
			c.printInventory()
		default:
			fmt.Fprintln(c.output, "Неизвестная команда. Выберите номер из меню.")
		}
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			if readErr := c.input.Err(); readErr != nil {
				return fmt.Errorf("чтение ввода: %w", readErr)
			}
			fmt.Fprintf(c.output, "Ошибка: %v\n", err)
		}
	}
}

func (c *Console) addVehicle() error {
	vehicle, err := c.readVehicle()
	if err != nil {
		return err
	}
	if err := c.fleet.AddVehicle(vehicle); err != nil {
		return err
	}
	fmt.Fprintf(c.output, "Транспорт принят в парк: %s (%s).\n", vehicle.Name(), vehicle.InventoryNumber())
	return nil
}

func (c *Console) addItem() error {
	item, err := c.readItem()
	if err != nil {
		return err
	}
	if err := c.fleet.AddItem(item); err != nil {
		return err
	}
	fmt.Fprintf(c.output, "Вещь поставлена на баланс: %s (%s).\n", item.Name(), item.InventoryNumber())
	return nil
}
