package elasticsearch

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/MurmurationsNetwork/MurmurationsServices/common/logger"
	"github.com/cenkalti/backoff"
	"github.com/olivere/elastic"
)

const docType = "_doc"

var (
	Client esClientInterface = &esClient{}
)

type esClientInterface interface {
	setClient(*elastic.Client)

	Index(string, interface{}) (*elastic.IndexResponse, error)
	IndexWithID(string, string, interface{}) (*elastic.IndexResponse, error)
	Search(string, elastic.Query) (*elastic.SearchResult, error)
	Delete(string, string) error
}

type esClient struct {
	client *elastic.Client
}

func Init() {
	var client *elastic.Client

	op := func() error {
		log := logger.GetLogger()

		var err error
		client, err = elastic.NewClient(
			elastic.SetURL(os.Getenv("ELASTICSEARCH_URL")),
			elastic.SetHealthcheckInterval(10*time.Second),
			elastic.SetErrorLog(log),
			elastic.SetInfoLog(log),
		)
		if err != nil {
			return err
		}

		return nil
	}
	notify := func(err error, time time.Duration) {
		logger.Info(fmt.Sprintf("trying to re-connect ElasticSearch %s \n", err))
	}
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 2 * time.Minute

	err := backoff.RetryNotify(op, b, notify)
	if err != nil {
		logger.Panic("error when trying to ping the ElasticSearch", err)
	}

	err = createMappings(client)
	if err != nil {
		logger.Panic("error when trying to create index for ElasticSearch", err)
		return
	}

	Client.setClient(client)
}

func (c *esClient) setClient(client *elastic.Client) {
	c.client = client
}

func (c *esClient) Index(index string, doc interface{}) (*elastic.IndexResponse, error) {
	ctx := context.Background()
	result, err := c.client.Index().
		Index(index).
		Type(docType).
		BodyJson(doc).
		Do(ctx)
	if err != nil {
		logger.Error(fmt.Sprintf("error when trying to index document in index %s", index), err)
		return nil, err
	}

	return result, nil
}

func (c *esClient) IndexWithID(index string, id string, doc interface{}) (*elastic.IndexResponse, error) {
	ctx := context.Background()
	result, err := c.client.Index().
		Index(index).
		Id(id).
		Type("_doc").
		BodyJson(doc).
		Do(ctx)
	if err != nil {
		logger.Error(fmt.Sprintf("error when trying to index a document in index %s", index), err)
		return nil, err
	}

	return result, nil
}

func (c *esClient) Search(index string, query elastic.Query) (*elastic.SearchResult, error) {
	ctx := context.Background()
	result, err := c.client.Search(index).
		Query(query).
		RestTotalHitsAsInt(true).
		Do(ctx)
	if err != nil {
		logger.Error(fmt.Sprintf("error when trying to search documents in index %s", index), err)
		return nil, err
	}

	return result, nil
}

func (c *esClient) Delete(index string, id string) error {
	ctx := context.Background()
	_, err := c.client.Delete().
		Index(index).
		Type(docType).
		Id(id).
		Do(ctx)
	if err != nil {
		// Don't need to tell the client data doesn't exist.
		if elastic.IsNotFound(err) {
			return nil
		}
		logger.Error(fmt.Sprintf("error when trying to delete a document in index %s", index), err)
		return err
	}
	return nil
}
