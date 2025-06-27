package analysis

import (
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
	lines   [][2]int
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
		lines:   calcLines(text),
	}

	s.logger.Printf("for file with uri=%+v got path: %+v", uri, p)
	return nil
}

func (s *State) UpdateDocument(uri DocumentURI, text string) error {
	v, ok := s.Documents[uri]
	if !ok {
		return errors.New("document not found")
	}

	// update line cache
	v.lines = calcLines(text)
	s.Documents[uri] = v

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
		r := doc.sanToRange(log.Span[0], log.Span[1])

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

func (doc Document) positonToSpan(r lsp.Range) [2]int {
	return [2]int{doc.lines[r.Start.Line][0], doc.lines[r.End.Line][1]}
}

func (doc Document) sanToRange(start, end int) (r lsp.Range) {
	for idx, line := range doc.lines {
		if start >= line[0] && start <= line[1] {
			r.Start = lsp.Position{Line: idx, Character: start - line[0]}
		}

		if end >= line[0] && end <= line[1] {
			r.End = lsp.Position{Line: idx, Character: end - line[0]}
		}
	}

	return
}

func calcLines(text string) [][2]int {
	var lines [][2]int

	var lineStart int
	for idx, c := range text {
		// found newline
		if c == '\n' {
			lineRange := [2]int{lineStart, idx}
			lines = append(lines, lineRange)
			lineStart = idx + 1
		}
	}

	if lineStart != len(text) {
		lines = append(lines, [2]int{lineStart, len(text)})
	}

	return lines
}
