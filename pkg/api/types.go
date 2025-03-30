package api

type (
	Response struct {
		Data  interface{} `json:"data,omitempty"`
		Meta  *Meta       `json:"meta,omitempty"`
		Error *Error      `json:"error,omitempty"`
	}

	Meta struct {
		Total int `json:"total,omitempty"`
		Page  int `json:"page,omitempty"`
		Limit int `json:"limit,omitempty"`
	}

	Error struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	}
)

func SuccessResponse(data interface{}, meta *Meta) Response {
	return Response{
		Data: data,
		Meta: meta,
	}
}

func ErrorResponse(message, code string) Response {
	return Response{
		Error: &Error{Message: message, Code: code},
	}
}
