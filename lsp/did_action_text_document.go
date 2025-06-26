package lsp

type DidOpenTextDocumentNotification struct {
	Notification
	Params DidOpenTextDocumentParams `json:"params"`
}

type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

type DidChangeTextDocumentNotification struct {
	Notification
	Params DidChangeTextDocumentParams `json:"params"`
}

type DidChangeTextDocumentParams struct {
	TextDocument   VersionTextDocumentIdentifier    `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

type HoverTextDocumentRequest struct {
	Request
	Params HoverTextDocumentParams `json:"params"`
}

type HoverTextDocumentParams struct {
	TextDocumentPositionParams
}

type HoverTextDocumentResponse struct {
	Response
	Result HoverTextDocumentResult `json:"result"`
}

type HoverTextDocumentResult struct {
	Contents string `json:"contents"`
}

type PublicDiagnosticsNotification struct {
	Notification
	Params PublicDiagnosticsParams `json:"params"`
}

type PublicDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type Diagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity"`
	Source   string `json:"source"`
	Message  string `json:"message"`
}

func NewHoverTextDocumentResponse(id int, contents string) HoverTextDocumentResponse {
	return HoverTextDocumentResponse{
		Response: NewResponse(id),
		Result: HoverTextDocumentResult{
			Contents: contents,
		},
	}
}

func NewPublicDiagnosticsNotification(
	uri string,
	diagnostics []Diagnostic,
) PublicDiagnosticsNotification {
	return PublicDiagnosticsNotification{
		Notification: NewNotification("textDocument/publishDiagnostics"),
		Params: PublicDiagnosticsParams{
			URI:         uri,
			Diagnostics: diagnostics,
		},
	}
}
