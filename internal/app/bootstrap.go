// Package app собирает зависимости приложения при запуске.
package app

import (
	"fmt"

	"go.uber.org/dig"

	"vyshkat/internal/service"
)

// BuildFleet создаёт парк через DI-контейнер с базовым техосмотром.
// Контейнер остаётся внутри точки сборки; сервисам он не передаётся.
func BuildFleet() (*service.FleetService, error) {
	container := dig.New()
	if err := container.Provide(
		service.NewBasicServiceCenter,
		dig.As(new(service.ServiceCenter)),
	); err != nil {
		return nil, fmt.Errorf("регистрация техосмотра: %w", err)
	}
	if err := container.Provide(service.NewFleetService); err != nil {
		return nil, fmt.Errorf("регистрация сервиса парка: %w", err)
	}

	var fleet *service.FleetService
	if err := container.Invoke(func(value *service.FleetService) {
		fleet = value
	}); err != nil {
		return nil, fmt.Errorf("сборка сервиса парка: %w", err)
	}
	return fleet, nil
}
