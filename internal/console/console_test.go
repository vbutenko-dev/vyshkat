package console_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"vyshkat/internal/console"
	"vyshkat/internal/domain"
	"vyshkat/internal/service"
)

func TestRunAddsAllObjectTypesAndShowsReports(t *testing.T) {
	fleet := newFleet(t)
	input := strings.Join([]string{
		"1", "1", "Самокат", "S-001", "5", "1.5", "да",
		"1", "2", "Велосипед", "B-001", "6", "да",
		"1", "3", "Электровелосипед", "E-001", "7", "2.25", "да",
		"2", "1", "Шлем", "H-001",
		"2", "2", "Зарядная станция", "C-001", "4.5",
		"3", "4", "5", "0",
	}, "\n")

	output := runConsole(t, fleet, input)
	wantReport := service.FleetReport{
		VehicleCount:     3,
		ItemCount:        2,
		TotalCount:       5,
		VehicleEnergyKWh: 3.75,
		ItemEnergyKWh:    4.5,
		TotalEnergyKWh:   8.25,
	}
	if got := fleet.Report(); got != wantReport {
		t.Errorf("Report() = %+v, want %+v", got, wantReport)
	}
	for _, want := range []string{
		"Транспорт принят в парк: Самокат (S-001).",
		"Транспорт принят в парк: Велосипед (B-001).",
		"Транспорт принят в парк: Электровелосипед (E-001).",
		"Вещь поставлена на баланс: Шлем (H-001).",
		"Вещь поставлена на баланс: Зарядная станция (C-001).",
		"Транспорт: 3",
		"Вещи: 2",
		"Всего: 5",
		"Энергия транспорта: 3.750 кВт·ч/сутки",
		"Энергия вещей: 4.500 кВт·ч/сутки",
		"Общая энергия: 8.250 кВт·ч/сутки",
		"До свидания.",
	} {
		assertContains(t, output, want)
	}
	if strings.Contains(output, "Ошибка:") {
		t.Errorf("valid session returned an error:\n%s", output)
	}

	// Отдельный запрос инвентаризации проверяет её вывод без сообщений приёмки.
	inventoryOutput := runConsole(t, fleet, "5\n0\n")
	for _, item := range fleet.Inventory() {
		assertContains(t, inventoryOutput, item.Name())
		assertContains(t, inventoryOutput, item.InventoryNumber())
	}
}

func TestRunRejectsFailedInspectionAndDuplicateNumber(t *testing.T) {
	fleet := newFleet(t)
	input := strings.Join([]string{
		"1", "1", "Неисправный самокат", "S-001", "9", "1.5", "нет",
		// После отказа тот же номер остаётся свободным.
		"1", "2", "Исправный велосипед", "S-001", "6", "да",
		"2", "1", "Шлем с занятым номером", "S-001",
		"0",
	}, "\n")

	output := runConsole(t, fleet, input)
	assertContains(t, output, "Ошибка: "+service.ErrInspectionRejected.Error())
	assertContains(t, output, "транспорт неисправен")
	assertContains(t, output, "Ошибка: "+service.ErrDuplicateInventoryNumber.Error())
	assertContains(t, output, "Транспорт принят в парк: Исправный велосипед (S-001).")
	items := fleet.Inventory()
	if len(items) != 1 {
		t.Fatalf("inventory length = %d, want 1", len(items))
	}
	if got := items[0].Name(); got != "Исправный велосипед" {
		t.Errorf("accepted item = %q, want Исправный велосипед", got)
	}
	if got := fleet.Report().TotalEnergyKWh; got != 0 {
		t.Errorf("energy = %v, want 0; rejected scooter must not be counted", got)
	}
}

func TestRunBeginnerListUsesInclusiveThreshold(t *testing.T) {
	fleet := newFleet(t)
	for _, tc := range []struct {
		name   string
		number string
		score  int
	}{
		{"Сложный велосипед", "B-005", 5},
		{"Простой велосипед", "B-006", 6},
	} {
		bike, err := domain.NewBicycle(tc.name, tc.number, tc.score, true)
		if err != nil {
			t.Fatal(err)
		}
		if err := fleet.AddVehicle(bike); err != nil {
			t.Fatal(err)
		}
	}
	helm, err := domain.NewHelmet("Шлем новичка", "H-001")
	if err != nil {
		t.Fatal(err)
	}
	if err := fleet.AddItem(helm); err != nil {
		t.Fatal(err)
	}

	output := runConsole(t, fleet, "4\n0\n")
	assertContains(t, output, "Простой велосипед")
	assertContains(t, output, "B-006")
	for _, unwanted := range []string{"Сложный велосипед", "B-005", "Шлем новичка", "H-001"} {
		if strings.Contains(output, unwanted) {
			t.Errorf("beginner list contains %q:\n%s", unwanted, output)
		}
	}
}

func TestRunRetriesMalformedNumbers(t *testing.T) {
	fleet := newFleet(t)
	input := strings.Join([]string{
		"1", "1", "Самокат", "S-001", "шесть", "6", "много", "1.25", "да",
		"0",
	}, "\n")

	output := runConsole(t, fleet, input)
	assertContains(t, output, "Транспорт принят в парк: Самокат (S-001).")
	vehicles := fleet.Vehicles()
	if len(vehicles) != 1 {
		t.Fatalf("vehicle count = %d, want 1", len(vehicles))
	}
	if got := vehicles[0].BeginnerScore(); got != 6 {
		t.Errorf("score = %d, want 6", got)
	}
	if got := fleet.Report().TotalEnergyKWh; got != 1.25 {
		t.Errorf("energy = %v, want 1.25", got)
	}
}

func TestRunReportsInvalidCardsAndReturnsToMenu(t *testing.T) {
	for _, tc := range []struct {
		name string
		card []string
		want error
	}{
		{"empty_name", []string{"2", "1", "   ", "H-001"}, domain.ErrEmptyName},
		{"empty_number", []string{"2", "1", "Шлем", "   "}, domain.ErrEmptyInventoryNumber},
		{"invalid_score", []string{"1", "2", "Велосипед", "B-001", "0", "да"}, domain.ErrInvalidBeginnerScore},
		{"negative_energy", []string{"2", "2", "Станция", "C-001", "-1"}, domain.ErrInvalidDailyEnergy},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fleet := newFleet(t)
			input := strings.Join(append(tc.card, "3", "0"), "\n")

			output := runConsole(t, fleet, input)
			assertContains(t, output, "Ошибка: "+tc.want.Error())
			assertContains(t, output, "Всего: 0")
			assertContains(t, output, "До свидания.")
			if got := len(fleet.Inventory()); got != 0 {
				t.Errorf("inventory length = %d, want 0", got)
			}
			if strings.Contains(output, "Неизвестная команда") {
				t.Errorf("card answers leaked into the main menu:\n%s", output)
			}
		})
	}
}

func TestRunEmptyReportsAndUnknownCommand(t *testing.T) {
	output := runConsole(t, newFleet(t), "неизвестно\n3\n4\n5\n0\n")
	for _, want := range []string{
		"Неизвестная команда",
		"Транспорт: 0",
		"Вещи: 0",
		"Всего: 0",
		"Общая энергия: 0.000 кВт·ч/сутки",
		"Подходящего транспорта нет.",
		"Инвентарь пуст.",
		"До свидания.",
	} {
		assertContains(t, output, want)
	}
}

func TestRunStopsAtEOFWithoutAddingPartialCards(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input string
	}{
		{"main_menu", ""},
		{"vehicle_kind", "1\n"},
		{"vehicle_condition", "1\n1\nСамокат\nS-001\n6\n1.5\n"},
		{"item_number", "2\n1\nШлем\n"},
		{"item_energy", "2\n2\nСтанция\nC-001\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fleet := newFleet(t)
			runConsole(t, fleet, tc.input)
			if got := len(fleet.Inventory()); got != 0 {
				t.Errorf("inventory length = %d, want 0", got)
			}
		})
	}
}

func TestRunReturnsInputReadError(t *testing.T) {
	want := errors.New("input unavailable")
	var output bytes.Buffer
	app := console.New(newFleet(t), failingReader{err: want}, &output)
	if err := app.Run(); !errors.Is(err, want) {
		t.Errorf("Run() error = %v, want %v", err, want)
	}
}

func TestRunPreservesReadErrorAfterBufferedInvalidCard(t *testing.T) {
	sourceErr := errors.New("input disconnected")
	fleet := newFleet(t)
	var output bytes.Buffer
	input := &dataAndErrorReader{
		data: []byte("2\n1\n   \nH-001\n"),
		err:  sourceErr,
	}
	app := console.New(fleet, input, &output)
	if err := app.Run(); !errors.Is(err, sourceErr) {
		t.Errorf("Run() error = %v, want source error %v", err, sourceErr)
	}
	if got := len(fleet.Inventory()); got != 0 {
		t.Errorf("inventory length = %d, want 0", got)
	}
}

type dataAndErrorReader struct {
	data []byte
	err  error
}

func (reader *dataAndErrorReader) Read(buffer []byte) (int, error) {
	n := copy(buffer, reader.data)
	reader.data = reader.data[n:]
	if len(reader.data) > 0 {
		return n, nil
	}
	return n, reader.err
}

type failingReader struct {
	err error
}

func (reader failingReader) Read(_ []byte) (int, error) {
	return 0, reader.err
}

func newFleet(t *testing.T) *service.FleetService {
	t.Helper()
	fleet, err := service.NewFleetService(service.NewBasicServiceCenter())
	if err != nil {
		t.Fatal(err)
	}
	return fleet
}

func runConsole(t *testing.T, fleet *service.FleetService, input string) string {
	t.Helper()
	var output bytes.Buffer
	app := console.New(fleet, strings.NewReader(input), &output)
	if err := app.Run(); err != nil {
		t.Fatalf("Run(): %v", err)
	}
	return output.String()
}

func assertContains(t *testing.T, output, want string) {
	t.Helper()
	if !strings.Contains(output, want) {
		t.Errorf("output does not contain %q:\n%s", want, output)
	}
}
