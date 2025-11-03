// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0

package ansible

import (
	"encoding/json"
	"fmt"
	"os"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// NavigatorEvent represents a JSON event from ansible-navigator output
type NavigatorEvent struct {
	Event   string                 `json:"event"`
	UUID    string                 `json:"uuid"`
	Counter int                    `json:"counter"`
	Task    string                 `json:"task,omitempty"`
	Play    string                 `json:"play,omitempty"`
	Host    string                 `json:"host,omitempty"`
	Status  string                 `json:"status,omitempty"`
	Stdout  string                 `json:"stdout,omitempty"`
	Stderr  string                 `json:"stderr,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// Summary contains aggregated information about ansible-navigator execution
type Summary struct {
	PlaysRun    int              `json:"plays_run"`
	TasksTotal  int              `json:"tasks_total"`
	TasksFailed int              `json:"tasks_failed"`
	FailedTasks []NavigatorEvent `json:"failed_tasks"`
}

// handleNavigatorEvent processes individual ansible-navigator JSON events
func handleNavigatorEvent(ui packersdk.Ui, e *NavigatorEvent, summary *Summary) {
	switch e.Event {
	case "playbook_on_start":
		if e.Play != "" {
			ui.Message(fmt.Sprintf("==> ansible-navigator: Play '%s' started", e.Play))
		}
		summary.PlaysRun++

	case "playbook_on_play_start":
		if e.Play != "" {
			ui.Message(fmt.Sprintf("==> ansible-navigator: Play '%s' starting", e.Play))
		}

	case "runner_on_ok":
		if e.Task != "" && e.Host != "" {
			ui.Message(fmt.Sprintf("==> ansible-navigator: Task '%s' [%s]: OK", e.Task, e.Host))
		}
		summary.TasksTotal++

	case "runner_on_failed":
		if e.Task != "" && e.Host != "" {
			ui.Error(fmt.Sprintf("==> ansible-navigator: Task '%s' [%s]: FAILED", e.Task, e.Host))
			summary.FailedTasks = append(summary.FailedTasks, *e)
			summary.TasksFailed++
		}
		summary.TasksTotal++

	case "runner_on_skipped":
		if e.Task != "" && e.Host != "" {
			ui.Message(fmt.Sprintf("==> ansible-navigator: Task '%s' [%s]: SKIPPED", e.Task, e.Host))
		}
		summary.TasksTotal++

	case "runner_on_unreachable":
		if e.Task != "" && e.Host != "" {
			ui.Error(fmt.Sprintf("==> ansible-navigator: Task '%s' [%s]: UNREACHABLE", e.Task, e.Host))
			summary.FailedTasks = append(summary.FailedTasks, *e)
			summary.TasksFailed++
		}
		summary.TasksTotal++

	case "playbook_on_stats":
		ui.Message("==> ansible-navigator: Playbook execution completed")
		// Extract stats from Data if available
		if e.Data != nil {
			if changed, ok := e.Data["changed"].(float64); ok {
				summary.TasksTotal += int(changed)
			}
			if ok, okVal := e.Data["ok"].(float64); okVal {
				summary.TasksTotal += int(ok)
			}
		}

	case "playbook_on_task_start":
		if e.Task != "" {
			ui.Message(fmt.Sprintf("==> ansible-navigator: Task '%s' starting", e.Task))
		}
	}
}

// writeSummaryJSON writes the execution summary to a JSON file
func writeSummaryJSON(summary *Summary, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create summary file: %s", err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(summary); err != nil {
		return fmt.Errorf("failed to encode summary: %s", err)
	}

	return nil
}
