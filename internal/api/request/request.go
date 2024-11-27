package request

const (
	OkReq             = "Ok"
	badReq            = "Bad Request"
	InternalServerReq = "Internal Server Error"
)

type OkResponse struct {
	Description string `json:"description"`
}

type ErrorResponse struct {
	Description string `json:"description"`
	Error       string `json:"error"`
}

func Ok() *OkResponse {
	return &OkResponse{Description: OkReq}
}

func BadRequest(err string) *ErrorResponse {
	return &ErrorResponse{Description: badReq, Error: err}
}

func InternalServer(err string) *ErrorResponse {
	return &ErrorResponse{Description: InternalServerReq, Error: err}
}
