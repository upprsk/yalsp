package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"yalsp/analysis"
	"yalsp/lsp"
	"yalsp/rpc"
)

func main() {
	var logPath string
	flag.StringVar(&logPath, "logPath", "", "path to the file to log messages to")

	var yalcPath string
	flag.StringVar(&yalcPath, "yalcPath", "yalc", "path to the YAL compiler (yalc) executable")

	logger, err := getLogger(logPath)
	if err != nil {
		log.Panic("getLogger:", err)
	}

	logger.Println("started")
	defer logger.Println("exiting")

	output := os.Stdout

	state := analysis.NewState(logger, yalcPath)
	if err := state.Open(); err != nil {
		logger.Fatal("failed to open state:", err)
	}
	defer state.Close()

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(rpc.SplitMessage)

	for scanner.Scan() {
		msg := scanner.Bytes()
		method, contents, err := rpc.DecodeMessage(msg)
		if err != nil {
			logger.Println("error:", err)
			continue
		}

		if err := handleMessage(logger, output, &state, method, contents); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			logger.Println("message error:", err)
		}
	}

	if err := scanner.Err(); err != nil {
		logger.Println("scanner error:", err)
		logger.Println("scanner data:", scanner.Text())
	}
}

func handleMessage(
	logger *log.Logger,
	output io.Writer,
	state *analysis.State,
	method string,
	contents []byte,
) error {
	logger.Printf("received msg with method '%s' (%d bytes)", method, len(contents))

	switch method {
	case "initialize":
		var request lsp.InitializeRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			return fmt.Errorf("json unmarshal: %w", err)
		}

		logger.Printf(
			"connected to: %s %s",
			request.Params.ClientInfo.Name,
			request.Params.ClientInfo.Version,
		)

		response := lsp.NewInitializeResponse(request.Id)
		writeResponse(output, response)
		logger.Print("sent initialize reply")

	case "initialized": // nothing
		logger.Print("initialized")

	case "shutdown": // nothing
		logger.Print("finishing")
		return io.EOF

	case "textDocument/didOpen":
		var request lsp.DidOpenTextDocumentNotification
		if err := json.Unmarshal(contents, &request); err != nil {
			return fmt.Errorf("json unmarshal: %w", err)
		}

		state.OpenDocument(request.Params.TextDocument.URI, request.Params.TextDocument.Text)
		logger.Printf("opened: %s", request.Params.TextDocument.URI)

		go state.StartDiagnostics(func(uri string, ds []lsp.Diagnostic, err error) {
			if err != nil {
				logger.Println("diagnostics error:", err)
				return
			}

			// send the diagnostics
			notification := lsp.NewPublicDiagnosticsNotification(uri, ds)
			writeResponse(output, notification)
		})

	case "textDocument/didChange":
		var request lsp.DidChangeTextDocumentNotification
		if err := json.Unmarshal(contents, &request); err != nil {
			return fmt.Errorf("json unmarshal: %w", err)
		}

		for _, changes := range request.Params.ContentChanges {
			state.UpdateDocument(request.Params.TextDocument.URI, changes.Text)
			logger.Printf("changed: %s", request.Params.TextDocument.URI)
		}

		go state.StartDiagnostics(func(uri string, ds []lsp.Diagnostic, err error) {
			if err != nil {
				logger.Println("diagnostics error:", err)
				return
			}

			// send the diagnostics
			notification := lsp.NewPublicDiagnosticsNotification(uri, ds)
			writeResponse(output, notification)
		})

	case "textDocument/hover":
		var request lsp.HoverTextDocumentRequest
		if err := json.Unmarshal(contents, &request); err != nil {
			return fmt.Errorf("json unmarshal: %w", err)
		}

		response := lsp.NewHoverTextDocumentResponse(
			request.Id,
			fmt.Sprintf("hello world! position=%v", request.Params.Position),
		)
		writeResponse(output, response)

		logger.Print("sent hover reply")

	default:
		return fmt.Errorf("unknown method: %s", method)
	}

	return nil
}

func writeResponse(w io.Writer, msg any) {
	reply := rpc.EncodeMessage(msg)
	w.Write([]byte(reply))
}

func getLogger(filename string) (*log.Logger, error) {
	var logFile *os.File

	if filename == "" {
		logFile = os.Stderr
	} else {
		var err error
		logFile, err = os.Create(filename)
		if err != nil {
			return nil, fmt.Errorf("create: %w", err)
		}
	}

	return log.New(logFile, "[yalsp]", log.Ldate|log.Ltime|log.Lshortfile), nil
}
