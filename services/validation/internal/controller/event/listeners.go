package event

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/MurmurationsNetwork/MurmurationsServices/pkg/logger"
	"github.com/MurmurationsNetwork/MurmurationsServices/pkg/messaging"
	"github.com/MurmurationsNetwork/MurmurationsServices/pkg/redis"
	"github.com/MurmurationsNetwork/MurmurationsServices/services/validation/internal/model"
	"github.com/MurmurationsNetwork/MurmurationsServices/services/validation/internal/service"
	"github.com/MurmurationsNetwork/MurmurationsServices/services/validation/internal/validation"
)

// NodeHandler provides an interface for handling node events.
type NodeHandler interface {
	NewNodeCreatedListener() error
}

type nodeHandler struct {
	redis             redis.Redis
	validationService service.ValidationService
}

// NewNodeHandler creates a new NodeHandler with the provided validation service.
func NewNodeHandler(
	redis redis.Redis,
	validationService service.ValidationService,
) NodeHandler {
	return &nodeHandler{
		redis:             redis,
		validationService: validationService,
	}
}

// NewNodeCreatedListener starts a listener for node-created events.
func (handler *nodeHandler) NewNodeCreatedListener() error {
	return messaging.QueueSubscribe(
		messaging.NodeCreated,
		validation.QueueGroup,
		handler.newNodeCreatedHandler,
	)
}

// Temp counter for debugging.
var counter uint64

// newNodeCreatedHandler handles the logic for node-created messages.
func (handler *nodeHandler) newNodeCreatedHandler(msg *nats.Msg) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(
				fmt.Sprintf("Panic occurred in nodeCreated handler: %v", err),
				errors.New("panic"),
			)
		}
	}()

	defer func() {
		// Acknowledge the message regardless of error.
		if err := msg.Ack(); err != nil {
			logger.Error("Error when acknowledging message", err)
		}
	}()

	// Increment the counter.
	atomic.AddUint64(&counter, 1)
	logger.Info(fmt.Sprintf("Receiving new node created event no: %d", counter))

	var nodeCreatedData messaging.NodeCreatedData
	if err := json.Unmarshal(msg.Data, &nodeCreatedData); err != nil {
		logger.Error("Error when trying to parse nodeCreatedData", err)
		return
	}

	nodeKey := fmt.Sprintf(
		"%s:%d",
		nodeCreatedData.ProfileURL,
		nodeCreatedData.Version,
	)
	exists, err := handler.redis.Get(nodeKey)
	if err != nil {
		logger.Error("Error getting key from Redis", err)
		return
	}

	if exists == "" {
		handler.validationService.ValidateNode(&model.Node{
			ProfileURL: nodeCreatedData.ProfileURL,
			Version:    nodeCreatedData.Version,
		})
		err := handler.redis.Set(nodeKey, "processed", 1*time.Hour)
		if err != nil {
			logger.Error("Error setting key in Redis", err)
		}
	} else {
		logger.Info(fmt.Sprintf("Duplicate node created event: %s", nodeKey))
	}
}
