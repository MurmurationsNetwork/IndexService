package revalidatenode

import (
	"sync"

	"github.com/MurmurationsNetwork/MurmurationsServices/pkg/logger"
	mongodb "github.com/MurmurationsNetwork/MurmurationsServices/pkg/mongo"
	"github.com/MurmurationsNetwork/MurmurationsServices/pkg/natsclient"
	"github.com/MurmurationsNetwork/MurmurationsServices/services/revalidatenode/config"
	"github.com/MurmurationsNetwork/MurmurationsServices/services/revalidatenode/internal/repository/mongo"
	"github.com/MurmurationsNetwork/MurmurationsServices/services/revalidatenode/internal/service"
)

// NodeRevalidationCron handles the initialization and running of the node revalidation cron job.
type NodeRevalidationCron struct {
	// Ensures cleanup is run only once.
	runCleanup sync.Once
}

// NewCronJob initializes a new NodeRevalidationCron instance with necessary configurations.
func NewCronJob() *NodeRevalidationCron {
	config.Init()

	// Initialize MongoDB client.
	uri := mongodb.GetURI(
		config.Conf.Mongo.USERNAME,
		config.Conf.Mongo.PASSWORD,
		config.Conf.Mongo.HOST,
	)
	if err := mongodb.NewClient(uri, config.Conf.Mongo.DBName); err != nil {
		logger.Panic("error when trying to connect to MongoDB", err)
	}

	// Check MongoDB connection.
	if err := mongodb.Client.Ping(); err != nil {
		logger.Panic("error when trying to ping the MongoDB", err)
	}

	// Initialize NATS client.
	setupNATS()

	return &NodeRevalidationCron{}
}

// setupNATS initializes Nats service.
func setupNATS() {
	err := natsclient.Initialize(config.Conf.Nats.URL)
	if err != nil {
		logger.Panic("Failed to create Nats client", err)
	}
}

// Run executes the node revalidation process.
func (nc *NodeRevalidationCron) Run() error {
	// Create and run node service for revalidation.
	nodeService := service.NewNodeService(
		mongo.NewNodeRepository(mongodb.Client.GetClient()),
	)

	if err := nodeService.RevalidateNodes(); err != nil {
		return err
	}

	// Perform cleanup after running the service.
	nc.cleanup()

	return nil
}

// cleanup disconnects MongoDB and NATS clients.
func (nc *NodeRevalidationCron) cleanup() {
	nc.runCleanup.Do(func() {
		var errOccurred bool

		// Disconnect from MongoDB.
		mongodb.Client.Disconnect()

		// Disconnect from NATS.
		if err := natsclient.GetInstance().Disconnect(); err != nil {
			logger.Error("Error disconnecting from NATS: %v", err)
			errOccurred = true
		}

		// Log based on whether an error occurred.
		if errOccurred {
			logger.Info("Revalidate node service stopped with errors.")
		} else {
			logger.Info("Revalidate node service stopped gracefully.")
		}
	})
}
