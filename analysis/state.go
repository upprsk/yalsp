package analysis

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"yalsp/lsp"
)

type State struct {
	Documents map[DocumentURI]Document

	yalcPath string
	logger   *log.Logger

	fs *TmpFS
}

type Document struct {
	URI     string
	tmpPath string
}

type DocumentURI = string

type DiagnosticCallback func(uri DocumentURI, ds []lsp.Diagnostic, err error)

func NewState(logger *log.Logger, yalcPath string) State {
	return State{
		Documents: map[DocumentURI]Document{},
		yalcPath:  yalcPath,
		logger:    logger,
	}
}

func (s *State) Open() error {
	fs, err := OpenTmpFs("yalsp")
	if err != nil {
		return fmt.Errorf("open tmpfs: %w", err)
	}

	s.fs = fs

	return nil
}

func (s *State) Close() {
	if s.fs != nil {
		s.fs.Close()
		s.fs = nil
	}
}

func (s *State) OpenDocument(uri DocumentURI, text string) error {
	p, err := s.fs.AddNewFile(uri, text)
	if err != nil {
		return fmt.Errorf("addNewFile: %w", err)
	}

	s.Documents[uri] = Document{
		URI:     uri,
		tmpPath: p,
	}

	s.logger.Printf("for file with uri=%+v got path: %+v", uri, p)
	return nil
}

func (s *State) UpdateDocument(uri DocumentURI, text string) error {
	v, ok := s.Documents[uri]
	if !ok {
		return errors.New("document not found")
	}

	if _, err := s.fs.AddNewFile(v.URI, text); err != nil {
		return fmt.Errorf("addNewFile: %w", err)
	}

	return nil
}

func (s *State) StartDiagnostics(cb DiagnosticCallback) {
	for uri, doc := range s.Documents {
		diagnostics, err := s.startDiagnosticsFor(doc)
		cb(uri, diagnostics, err)
	}
}

func (s *State) startDiagnosticsFor(doc Document) ([]lsp.Diagnostic, error) {
	// FIXME: can not use the filesystem here
	logs, err := runYalcSingleFile(s.yalcPath, doc.tmpPath)
	if err != nil {
		return nil, fmt.Errorf("run yalc: %w", err)
	}

	diags := make([]lsp.Diagnostic, len(logs))
	for idx, log := range logs {
		r, err := s.calcRange(doc.tmpPath, log.Span[0], log.Span[1])
		if err != nil {
			s.logger.Printf("failed to calculate range for %s: %v", doc.URI, err)
		}

		var severity int = 1
		switch log.Prefix {
		case "bug":
			severity = 1
		case "error":
			severity = 1
		case "warn":
			severity = 2
		case "note":
			severity = 3
		case "debug":
			severity = 4
		}

		s.logger.Println("got error message:", log.Message)

		diags[idx] = lsp.Diagnostic{
			Range:    r,
			Severity: severity,
			Source:   "yalsp",
			Message:  log.Message,
		}
	}

	return diags, nil
}

func (s *State) calcRange(path string, start, end int) (lsp.Range, error) {
	f, err := s.fs.OpenFileFromPath(path)
	if err != nil {
		return lsp.Range{}, fmt.Errorf("open file from name: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	var r lsp.Range

	for line, idx := 0, 0; scanner.Scan(); line++ {
		text := scanner.Text()
		endIdx := idx + len(text) + 1 // add newline
		if start >= idx && start <= endIdx {
			r.Start = lsp.Position{Line: line, Character: start - idx}
			s.logger.Printf("found start location: %v (line=%d, idx=%d)", r.Start, line, idx)
		}

		if end >= idx && end <= endIdx {
			r.End = lsp.Position{Line: line, Character: end - idx}
			s.logger.Printf("found end location: %v (line=%d, idx=%d)", r.End, line, idx)
			break
		}

		idx = endIdx
	}

	if err := scanner.Err(); err != nil {
		return r, fmt.Errorf("scanner: %w", err)
	}

	return r, nil
}
