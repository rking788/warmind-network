package charlemagne

// BaseResponse contains the field that are present in all response from the
// Charlemagne API.
type BaseResponse struct {
	ErrorCode int    `json:"errorCode"`
	ErrorMsg  string `json:"errorMessage"`
}
