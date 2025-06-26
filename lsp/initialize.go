package lsp

// method: "initialize"
type InitializeRequest struct {
	Request
	Params InitializeParams `json:"params"`
}

type InitializeResponse struct {
	Response
	Result InitializeResult `json:"result"`
}

type InitializeParams struct {
	ClientInfo            *ClientInfo `json:"clientInfo"`
	RootPath              string      `json:"rootPath"`
	RootURI               string      `json:"rootUri"`
	InitializationOptions any         `json:"initializationOptions"`
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   ServerInfo         `json:"serverInfo"`
}

type ServerCapabilities struct {
	TextDocumentSync int  `json:"textDocumentSync"`
	HoverProvider    bool `json:"hoverProvider"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func NewInitializeResponse(id int) InitializeResponse {
	return InitializeResponse{
		Response: NewResponse(id),
		Result: InitializeResult{
			Capabilities: ServerCapabilities{
				TextDocumentSync: 1,
				HoverProvider:    true,
			},
			ServerInfo: ServerInfo{
				Name:    "yalsp",
				Version: "0.0.0-alpha1",
			},
		},
	}
}
