package app

import (
	"github.com/MurmurationsNetwork/MurmurationsServices/services/index/internal/controller/event"
	"github.com/MurmurationsNetwork/MurmurationsServices/services/index/internal/repository/db"
	"github.com/MurmurationsNetwork/MurmurationsServices/services/index/internal/usecase"
)

func listenToEvents() error {
	nodeHandler := event.NewNodeHandler(usecase.NewNodeService(db.NewRepository()))

	err := nodeHandler.Validated()
	if err != nil {
		return err
	}
	err = nodeHandler.ValidationFailed()
	if err != nil {
		return err
	}
	return nil
}
