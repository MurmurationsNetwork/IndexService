package http

import (
	"encoding/csv"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/MurmurationsNetwork/MurmurationsServices/common/jsonapi"
	"github.com/MurmurationsNetwork/MurmurationsServices/services/dataproxy/internal/usecase"
	"github.com/gin-gonic/gin"
)

type BatchesHandler interface {
	GetBatchesByUserID(c *gin.Context)
	Validate(c *gin.Context)
	Import(c *gin.Context)
	Edit(c *gin.Context)
	Delete(c *gin.Context)
}

type batchesHandler struct {
	batchUsecase usecase.BatchUsecase
}

func NewBatchesHandler(batchService usecase.BatchUsecase) BatchesHandler {
	return &batchesHandler{
		batchUsecase: batchService,
	}
}

func (handler *batchesHandler) GetBatchesByUserID(c *gin.Context) {
	userId := c.Query("user_id")
	if len(userId) != 25 {
		errors := jsonapi.NewError([]string{"Invalid `user_id`"}, []string{"The `user_id` is not valid."}, nil, []int{http.StatusBadRequest})
		res := jsonapi.Response(nil, errors, nil, nil)
		c.JSON(errors[0].Status, res)
		return
	}

	batches, err := handler.batchUsecase.GetBatchesByUserID(userId)
	if err != nil {
		errors := jsonapi.NewError([]string{"Get Batches Failed"}, []string{"Failed to get batches by `user_id`: " + userId + " with error: " + err.Error()}, nil, []int{http.StatusBadRequest})
		res := jsonapi.Response(nil, errors, nil, nil)
		c.JSON(errors[0].Status, res)
		return
	}

	res := jsonapi.Response(batches, nil, nil, nil)
	c.JSON(http.StatusOK, res)
}

func (handler *batchesHandler) Validate(c *gin.Context) {
	file, schemas, errors := validateCsvInputs(c)
	if errors != nil {
		res := jsonapi.Response(nil, errors, nil, nil)
		c.JSON(errors[0].Status, res)
		return
	}

	records, errors := parseCsv(file)
	if errors != nil {
		res := jsonapi.Response(nil, errors, nil, nil)
		c.JSON(errors[0].Status, res)
		return
	}

	line, err := handler.batchUsecase.Validate(schemas, records)
	if err != nil {
		errors := jsonapi.NewError([]string{"CSV Validation Failed"}, []string{"Failed to validate line " + strconv.Itoa(line) + " with error: " + err.Error()}, nil, []int{http.StatusBadRequest})
		res := jsonapi.Response(nil, errors, nil, nil)
		c.JSON(errors[0].Status, res)
		return
	}

	meta := jsonapi.NewMeta("The submitted CSV file was validated successfully.", "", "")
	res := jsonapi.Response(nil, nil, nil, meta)
	c.JSON(http.StatusOK, res)
}

func (handler *batchesHandler) Import(c *gin.Context) {
	// TODO: We need to validate `user_id` from DB

	// The `user_id` is 25 characters long (cuid format)
	userId := c.PostForm("user_id")
	if len(userId) != 25 {
		errors := jsonapi.NewError([]string{"Invalid `user_id`"}, []string{"The `user_id` is not valid."}, nil, []int{http.StatusBadRequest})
		res := jsonapi.Response(nil, errors, nil, nil)
		c.JSON(errors[0].Status, res)
		return
	}

	file, schemas, errors := validateCsvInputs(c)
	if errors != nil {
		res := jsonapi.Response(nil, errors, nil, nil)
		c.JSON(errors[0].Status, res)
		return
	}

	records, errors := parseCsv(file)
	if errors != nil {
		res := jsonapi.Response(nil, errors, nil, nil)
		c.JSON(errors[0].Status, res)
		return
	}

	// Get metadata for the batch
	metaName := c.PostForm("meta_name")
	metaUrl := c.PostForm("meta_url")

	batchId, line, err := handler.batchUsecase.Import(schemas, records, userId, metaName, metaUrl)
	if err != nil {
		errors := jsonapi.NewError([]string{"CSV Import Failed"}, []string{"Failed to import line " + strconv.Itoa(line) + " in `batch_id`: " + batchId + " with error: " + err.Error()}, nil, []int{http.StatusBadRequest})
		meta := jsonapi.NewBatchMeta("", batchId)
		res := jsonapi.Response(nil, errors, nil, meta)
		c.JSON(errors[0].Status, res)
		return
	}

	meta := jsonapi.NewBatchMeta("The submitted CSV file was imported successfully.", batchId)
	res := jsonapi.Response(nil, nil, nil, meta)
	c.JSON(http.StatusOK, res)
}

func (handler *batchesHandler) Edit(c *gin.Context) {
	userId := c.PostForm("user_id")
	if len(userId) != 25 {
		errors := jsonapi.NewError([]string{"Invalid `user_id`"}, []string{"The `user_id` is not valid."}, nil, []int{http.StatusBadRequest})
		res := jsonapi.Response(nil, errors, nil, nil)
		c.JSON(errors[0].Status, res)
		return
	}

	batchId := c.PostForm("batch_id")
	if len(batchId) != 25 {
		errors := jsonapi.NewError([]string{"Invalid `batch_id`"}, []string{"The `batch_id` is not valid."}, nil, []int{http.StatusBadRequest})
		res := jsonapi.Response(nil, errors, nil, nil)
		c.JSON(errors[0].Status, res)
		return
	}

	file, schemas, errors := validateCsvInputs(c)
	if errors != nil {
		res := jsonapi.Response(nil, errors, nil, nil)
		c.JSON(errors[0].Status, res)
		return
	}

	records, errors := parseCsv(file)
	if errors != nil {
		res := jsonapi.Response(nil, errors, nil, nil)
		c.JSON(errors[0].Status, res)
		return
	}

	// Get metadata for the batch
	metaName := c.PostForm("meta_name")
	metaUrl := c.PostForm("meta_url")

	line, err := handler.batchUsecase.Edit(schemas, records, userId, batchId, metaName, metaUrl)
	if err != nil {
		errors := jsonapi.NewError([]string{"CSV Edit Failed"}, []string{"Failed to edit line " + strconv.Itoa(line) + " for `batch_id`: " + batchId + " with error: " + err.Error()}, nil, []int{http.StatusBadRequest})
		meta := jsonapi.NewBatchMeta("", batchId)
		res := jsonapi.Response(nil, errors, nil, meta)
		c.JSON(errors[0].Status, res)
		return
	}

	meta := jsonapi.NewBatchMeta("The submitted CSV file was updated successfully.", batchId)
	res := jsonapi.Response(nil, nil, nil, meta)
	c.JSON(http.StatusOK, res)
}

func (handler *batchesHandler) Delete(c *gin.Context) {
	userId := c.PostForm("user_id")
	if len(userId) != 25 {
		errors := jsonapi.NewError([]string{"Invalid `user_id`"}, []string{"The `user_id` is not valid."}, nil, []int{http.StatusBadRequest})
		res := jsonapi.Response(nil, errors, nil, nil)
		c.JSON(errors[0].Status, res)
		return
	}

	batchId := c.PostForm("batch_id")
	if len(batchId) != 25 {
		errors := jsonapi.NewError([]string{"Invalid `batch_id`"}, []string{"The `batch_id` is not valid."}, nil, []int{http.StatusBadRequest})
		res := jsonapi.Response(nil, errors, nil, nil)
		c.JSON(errors[0].Status, res)
		return
	}

	// Call delete service
	err := handler.batchUsecase.Delete(userId, batchId)
	if err != nil {
		errors := jsonapi.NewError([]string{"Delete Batch Failed"}, []string{"Failed to delete `batch_id`: " + batchId + " with error: " + err.Error()}, nil, []int{http.StatusBadRequest})
		meta := jsonapi.NewBatchMeta("", batchId)
		res := jsonapi.Response(nil, errors, nil, meta)
		c.JSON(errors[0].Status, res)
		return
	}

	meta := jsonapi.NewBatchMeta("The submitted `batch_id` was successfully deleted.", batchId)
	res := jsonapi.Response(nil, nil, nil, meta)
	c.JSON(http.StatusOK, res)
}

func validateCsvInputs(c *gin.Context) (*multipart.FileHeader, []string, []jsonapi.Error) {
	// Get fields from the POST request
	file, err := c.FormFile("file")
	if err != nil {
		errors := jsonapi.NewError([]string{"Get File Error"}, []string{"The submitted document could not be parsed."}, [][]string{{"parameter", "file"}}, []int{http.StatusBadRequest})
		return nil, nil, errors
	}

	// If the file is not CSV, we cannot process it
	fileName := file.Filename
	if fileName[len(fileName)-4:] != ".csv" {
		errors := jsonapi.NewError([]string{"Get File Error"}, []string{"The submitted document is not a CSV file."}, [][]string{{"parameter", "file"}}, []int{http.StatusBadRequest})
		return nil, nil, errors
	}

	// Get schemas from the POST request
	rawSchemas := c.PostForm("schemas")
	if rawSchemas == "" {
		errors := jsonapi.NewError([]string{"Invalid Query Parameter"}, []string{"The following query parameter is not valid: schemas."}, [][]string{{"parameter", "schemas"}}, []int{http.StatusBadRequest})
		return nil, nil, errors
	}

	// make schemas to []string
	rawSchemas = strings.ReplaceAll(rawSchemas, "\"", "")
	schemas := strings.Split(rawSchemas[1:len(rawSchemas)-1], ",")

	return file, schemas, nil
}

func parseCsv(file *multipart.FileHeader) ([][]string, []jsonapi.Error) {
	// Parse CSV and put all data in service
	rawFile, err := file.Open()
	if err != nil {
		errors := jsonapi.NewError([]string{"File Open Error"}, []string{"The file is corrupted and cannot be opened."}, [][]string{{"parameter", "file"}}, []int{http.StatusBadRequest})
		return nil, errors
	}

	csvReader := csv.NewReader(rawFile)
	records, err := csvReader.ReadAll()
	if err != nil {
		errors := jsonapi.NewError([]string{"File Open Error"}, []string{"Unable to parse file as CSV."}, [][]string{{"parameter", "file"}}, []int{http.StatusBadRequest})
		return nil, errors
	}

	return records, nil
}
