package main

import (
	"github.com/MurmurationsNetwork/MurmurationsServices/pkg/logger"
	"github.com/MurmurationsNetwork/MurmurationsServices/services/cronjob/schemaparser/config"
	"github.com/MurmurationsNetwork/MurmurationsServices/services/cronjob/schemaparser/internal/adapter/mongodb"
	"github.com/MurmurationsNetwork/MurmurationsServices/services/cronjob/schemaparser/internal/adapter/redisadapter"
	"github.com/MurmurationsNetwork/MurmurationsServices/services/cronjob/schemaparser/internal/repository/db"
	"github.com/MurmurationsNetwork/MurmurationsServices/services/cronjob/schemaparser/internal/service"
)

func init() {
	config.Init()
	mongodb.Init()
}

func main() {
	svc := service.NewSchemaService(
		db.NewSchemaRepository(),
		redisadapter.NewClient(),
	)

	url := config.Conf.Github.BranchURL
	branchInfo, err := svc.GetBranchInfo(url)
	if err != nil {
		logger.Error(
			"Error when trying to get last_commit and schema_list from: "+url,
			err,
		)
		return
	}

	hasNewCommit, err := svc.HasNewCommit(
		branchInfo.Commit.InnerCommit.Author.Date,
	)
	if err != nil {
		logger.Error("Error when trying to get schemas:lastCommit", err)
		return
	}
	if !hasNewCommit {
		return
	}

	err = svc.UpdateSchemas(branchInfo.Commit.Sha)
	if err != nil {
		logger.Error("Error when trying to update schemas", err)
		return
	}

	err = svc.SetLastCommit(branchInfo.Commit.InnerCommit.Author.Date)
	if err != nil {
		logger.Panic("Error when trying to set schemas:lastCommit", err)
		return
	}

	logger.Info("Library repo schemas loaded successfully")
}
