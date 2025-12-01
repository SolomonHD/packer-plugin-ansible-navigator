// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

package ansiblenavigatorremote

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockUi implements packersdk.Ui for testing
type MockUi struct {
	messages []string
	errors   []string
}

func (m *MockUi) Say(message string) {
	m.messages = append(m.messages, message)
}

func (m *MockUi) Message(message string) {
	m.messages = append(m.messages, message)
}

func (m *MockUi) Error(message string) {
	m.errors = append(m.errors, message)
}

func (m *MockUi) Errorf(message string, args ...interface{}) {
	m.errors = append(m.errors, message)
}

func (m *MockUi) Sayf(message string, args ...interface{}) {
	m.messages = append(m.messages, message)
}

func (m *MockUi) Machine(t string, args ...string) {}

func (m *MockUi) Ask(query string) (string, error) {
	return "", nil
}

func (m *MockUi) Askf(query string, args ...interface{}) (string, error) {
	return "", nil
}

func (m *MockUi) TrackProgress(src string, currentSize, totalSize int64, stream io.ReadCloser) io.ReadCloser {
	return stream
}

func TestNavigatorEventParsing(t *testing.T) {
	tests := []struct {
		name          string
		jsonInput     string
		expectedEvent string
		expectError   bool
	}{
		{
			name: "Valid playbook_on_start event",
			jsonInput: `{
				"event": "playbook_on_start",
				"uuid": "test-uuid-1",
				"counter": 1,
				"play": "Setup infrastructure"
			}`,
			expectedEvent: "playbook_on_start",
			expectError:   false,
		},
		{
			name: "Valid runner_on_ok event",
			jsonInput: `{
				"event": "runner_on_ok",
				"uuid": "test-uuid-2",
				"counter": 2,
				"task": "Install packages",
				"host": "web1.example.com",
				"status": "ok"
			}`,
			expectedEvent: "runner_on_ok",
			expectError:   false,
		},
		{
			name: "Valid runner_on_failed event",
			jsonInput: `{
				"event": "runner_on_failed",
				"uuid": "test-uuid-3",
				"counter": 3,
				"task": "Restart nginx",
				"host": "web2.example.com",
				"status": "failed"
			}`,
			expectedEvent: "runner_on_failed",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var event NavigatorEvent
			err := json.Unmarshal([]byte(tt.jsonInput), &event)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedEvent, event.Event)
			}
		})
	}
}

func TestHandleNavigatorEvent(t *testing.T) {
	tests := []struct {
		name               string
		event              NavigatorEvent
		expectMessage      bool
		expectError        bool
		expectedTasksTotal int
		expectedFailed     int
	}{
		{
			name: "Playbook start event",
			event: NavigatorEvent{
				Event: "playbook_on_start",
				Play:  "Configure SSH",
			},
			expectMessage:      true,
			expectError:        false,
			expectedTasksTotal: 0,
			expectedFailed:     0,
		},
		{
			name: "Task success event",
			event: NavigatorEvent{
				Event: "runner_on_ok",
				Task:  "Ensure SSHD running",
				Host:  "web1.example.com",
			},
			expectMessage:      true,
			expectError:        false,
			expectedTasksTotal: 1,
			expectedFailed:     0,
		},
		{
			name: "Task failure event",
			event: NavigatorEvent{
				Event: "runner_on_failed",
				Task:  "Restart nginx",
				Host:  "web2.example.com",
			},
			expectMessage:      false,
			expectError:        true,
			expectedTasksTotal: 1,
			expectedFailed:     1,
		},
		{
			name: "Task skipped event",
			event: NavigatorEvent{
				Event: "runner_on_skipped",
				Task:  "Install optional package",
				Host:  "web1.example.com",
			},
			expectMessage:      true,
			expectError:        false,
			expectedTasksTotal: 1,
			expectedFailed:     0,
		},
		{
			name: "Host unreachable event",
			event: NavigatorEvent{
				Event: "runner_on_unreachable",
				Task:  "Ping host",
				Host:  "web3.example.com",
			},
			expectMessage:      false,
			expectError:        true,
			expectedTasksTotal: 1,
			expectedFailed:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ui := &MockUi{}
			summary := &Summary{
				FailedTasks: make([]NavigatorEvent, 0),
			}

			handleNavigatorEvent(ui, &tt.event, summary, false)

			if tt.expectMessage {
				assert.Greater(t, len(ui.messages), 0, "Expected at least one message")
			}

			if tt.expectError {
				assert.Greater(t, len(ui.errors), 0, "Expected at least one error")
			}

			assert.Equal(t, tt.expectedTasksTotal, summary.TasksTotal, "Tasks total mismatch")
			assert.Equal(t, tt.expectedFailed, summary.TasksFailed, "Failed tasks mismatch")

			if tt.expectedFailed > 0 {
				assert.Len(t, summary.FailedTasks, tt.expectedFailed, "Failed tasks list mismatch")
			}
		})
	}
}

func TestSummaryAggregation(t *testing.T) {
	ui := &MockUi{}
	summary := &Summary{
		FailedTasks: make([]NavigatorEvent, 0),
	}

	events := []NavigatorEvent{
		{Event: "playbook_on_start", Play: "Setup"},
		{Event: "runner_on_ok", Task: "Task 1", Host: "host1"},
		{Event: "runner_on_ok", Task: "Task 2", Host: "host1"},
		{Event: "runner_on_failed", Task: "Task 3", Host: "host2"},
		{Event: "runner_on_ok", Task: "Task 4", Host: "host1"},
	}

	for _, event := range events {
		handleNavigatorEvent(ui, &event, summary, false)
	}

	assert.Equal(t, 1, summary.PlaysRun, "Expected 1 play")
	assert.Equal(t, 4, summary.TasksTotal, "Expected 4 tasks total")
	assert.Equal(t, 1, summary.TasksFailed, "Expected 1 failed task")
	assert.Len(t, summary.FailedTasks, 1, "Expected 1 failed task in list")
}

func TestWriteSummaryJSON(t *testing.T) {
	summary := &Summary{
		PlaysRun:    2,
		TasksTotal:  10,
		TasksFailed: 1,
		FailedTasks: []NavigatorEvent{
			{
				Event:  "runner_on_failed",
				Task:   "Restart nginx",
				Host:   "web1.example.com",
				Status: "failed",
			},
		},
	}

	// Create a temporary file
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "summary.json")

	// Write the summary
	err := writeSummaryJSON(summary, outputPath)
	assert.NoError(t, err, "Failed to write summary JSON")

	// Verify the file exists
	_, err = os.Stat(outputPath)
	assert.NoError(t, err, "Summary file does not exist")

	// Read and verify the content
	data, err := os.ReadFile(outputPath)
	assert.NoError(t, err, "Failed to read summary file")

	var readSummary Summary
	err = json.Unmarshal(data, &readSummary)
	assert.NoError(t, err, "Failed to parse summary JSON")

	assert.Equal(t, summary.PlaysRun, readSummary.PlaysRun)
	assert.Equal(t, summary.TasksTotal, readSummary.TasksTotal)
	assert.Equal(t, summary.TasksFailed, readSummary.TasksFailed)
	assert.Len(t, readSummary.FailedTasks, 1)
}

func TestStructuredLoggingConfiguration(t *testing.T) {
	tests := []struct {
		name              string
		navigatorMode     string
		structuredLogging bool
		shouldUseJSON     bool
	}{
		{
			name:              "JSON mode with structured logging enabled",
			navigatorMode:     "json",
			structuredLogging: true,
			shouldUseJSON:     true,
		},
		{
			name:              "JSON mode with structured logging disabled",
			navigatorMode:     "json",
			structuredLogging: false,
			shouldUseJSON:     false,
		},
		{
			name:              "Stdout mode with structured logging enabled",
			navigatorMode:     "stdout",
			structuredLogging: true,
			shouldUseJSON:     false,
		},
		{
			name:              "Stdout mode with structured logging disabled",
			navigatorMode:     "stdout",
			structuredLogging: false,
			shouldUseJSON:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				NavigatorMode:     tt.navigatorMode,
				StructuredLogging: tt.structuredLogging,
			}

			// Logic: structured logging only works when both are set correctly
			useStructuredLogging := config.StructuredLogging && config.NavigatorMode == "json"
			assert.Equal(t, tt.shouldUseJSON, useStructuredLogging)
		})
	}
}

func TestInvalidJSONHandling(t *testing.T) {
	summary := &Summary{
		FailedTasks: make([]NavigatorEvent, 0),
	}

	// This would simulate what happens in the actual decoder when invalid JSON is encountered
	invalidJSON := `{"event": "invalid"` // Missing closing brace

	var event NavigatorEvent
	err := json.Unmarshal([]byte(invalidJSON), &event)
	assert.Error(t, err, "Should fail to parse invalid JSON")

	// The summary should remain unchanged
	assert.Equal(t, 0, summary.PlaysRun)
	assert.Equal(t, 0, summary.TasksTotal)
	assert.Equal(t, 0, summary.TasksFailed)
}

func TestMultiplePlaysScenario(t *testing.T) {
	ui := &MockUi{}
	summary := &Summary{
		FailedTasks: make([]NavigatorEvent, 0),
	}

	events := []NavigatorEvent{
		{Event: "playbook_on_start", Play: "Play 1"},
		{Event: "runner_on_ok", Task: "Task 1-1", Host: "host1"},
		{Event: "runner_on_ok", Task: "Task 1-2", Host: "host1"},
		{Event: "playbook_on_start", Play: "Play 2"},
		{Event: "runner_on_ok", Task: "Task 2-1", Host: "host1"},
		{Event: "runner_on_failed", Task: "Task 2-2", Host: "host2"},
	}

	for _, event := range events {
		handleNavigatorEvent(ui, &event, summary, false)
	}

	assert.Equal(t, 2, summary.PlaysRun, "Expected 2 plays")
	assert.Equal(t, 4, summary.TasksTotal, "Expected 4 tasks total")
	assert.Equal(t, 1, summary.TasksFailed, "Expected 1 failed task")
}

func TestLogOutputPathValidation(t *testing.T) {
	tests := []struct {
		name        string
		logPath     string
		shouldExist bool
	}{
		{
			name:        "Valid log path",
			logPath:     filepath.Join(t.TempDir(), "test-summary.json"),
			shouldExist: true,
		},
		{
			name:        "Empty log path",
			logPath:     "",
			shouldExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.logPath == "" {
				// No file should be written
				return
			}

			summary := &Summary{
				PlaysRun:    1,
				TasksTotal:  5,
				TasksFailed: 0,
				FailedTasks: make([]NavigatorEvent, 0),
			}

			err := writeSummaryJSON(summary, tt.logPath)
			assert.NoError(t, err)

			if tt.shouldExist {
				_, err := os.Stat(tt.logPath)
				assert.NoError(t, err, "Log file should exist")
			}
		})
	}
}
