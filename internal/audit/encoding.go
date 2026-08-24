package audit

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

type Document struct {
	Events  []Event `json:"events"`
	Summary Summary `json:"summary"`
}

func NewDocument(events []Event) Document {
	cloned := make([]Event, len(events))
	for index, event := range events {
		cloned[index] = event.Clone()
	}
	return Document{Events: cloned, Summary: Summarize(cloned)}
}

func WriteJSON(writer io.Writer, events []Event) error {
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(NewDocument(events)); err != nil {
		return fmt.Errorf("encode audit document: %w", err)
	}
	return nil
}

func WriteJSONLines(writer io.Writer, events []Event) error {
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	for _, event := range events {
		if err := encoder.Encode(event); err != nil {
			return fmt.Errorf("encode audit event %d: %w", event.Sequence, err)
		}
	}
	return nil
}

func ReadJSONLines(reader io.Reader, maximum int) ([]Event, error) {
	if maximum < 0 {
		maximum = 0
	}
	scanner := bufio.NewScanner(reader)
	buffer := make([]byte, 64*1024)
	scanner.Buffer(buffer, 4*1024*1024)
	events := make([]Event, 0)
	line := 0
	for scanner.Scan() {
		line++
		if maximum > 0 && len(events) == maximum {
			return nil, fmt.Errorf("read audit events: maximum of %d exceeded", maximum)
		}
		var event Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return nil, fmt.Errorf("decode audit event on line %d: %w", line, err)
		}
		events = append(events, event.Clone())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan audit events: %w", err)
	}
	return events, nil
}
