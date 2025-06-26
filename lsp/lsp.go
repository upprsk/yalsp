package lsp

type Request struct {
	RPC    string `json:"jsonrpc"`
	Id     int    `json:"id"`
	Method string `json:"method"`
}

type Response struct {
	RPC string `json:"jsonrpc"`
	Id  int    `json:"id,omitempty"`
}

type Notification struct {
	RPC    string `json:"jsonrpc"`
	Method string `json:"method"`
}

func NewResponse(id int) Response {
	return Response{
		RPC: "2.0",
		Id:  id,
	}
}

func NewNotification(method string) Notification {
	return Notification{
		RPC:    "2.0",
		Method: method,
	}
}
