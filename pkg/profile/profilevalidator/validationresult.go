package profilevalidator

// ValidationResult is the results from a schema validation operation.
type ValidationResult struct {
	// Whether the validation passed.
	Valid bool
	// Titles of the validation errors.
	ErrorMessages []string
	// Detailed descriptions of the validation errors.
	Details []string
	// Sources indicates the failing pieces of data.
	Sources [][]string
	// HTTP status codes associated with each error.
	ErrorStatus []int
}

// NewValidationResult initializes a new ValidationResult object with default values.
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Valid:         true,
		ErrorMessages: make([]string, 0),
		Details:       make([]string, 0),
		Sources:       make([][]string, 0),
		ErrorStatus:   make([]int, 0),
	}
}

// AppendError adds a single error to the ValidationResult.
func (vr *ValidationResult) AppendError(
	errorMessage, detail string,
	source []string,
	status int,
) {
	vr.ErrorMessages = append(vr.ErrorMessages, errorMessage)
	vr.Details = append(vr.Details, detail)
	vr.Sources = append(vr.Sources, source)
	vr.ErrorStatus = append(vr.ErrorStatus, status)
	vr.Valid = false
}

// AppendErrors adds multiple errors to the ValidationResult.
func (vr *ValidationResult) AppendErrors(
	errorMessages, details []string,
	sources [][]string,
	status []int,
) {
	if vr == nil || len(errorMessages) == 0 {
		return
	}

	for i, errorMessage := range errorMessages {
		var detail string
		var source []string
		var stat int

		// Use details if available.
		if i < len(details) {
			detail = details[i]
		}

		// Use sources if available.
		if i < len(sources) {
			source = sources[i]
		}

		// Use status if available.
		if i < len(status) {
			stat = status[i]
		}

		vr.AppendError(errorMessage, detail, source, stat)
	}
}

// Merge combines another ValidationResult into the current one.
func (vr *ValidationResult) Merge(other *ValidationResult) *ValidationResult {
	if vr == nil || other == nil || other.Valid {
		return vr
	}

	vr.AppendErrors(
		other.ErrorMessages,
		other.Details,
		other.Sources,
		other.ErrorStatus,
	)

	return vr
}
