package runnerlabel

import (
	"fmt"
	"strings"

	"github.com/step-security/secure-repo/remediation/workflow/permissions"
	"gopkg.in/yaml.v3"
)

// RunnerLabelMapping represents the replacement to be performed
type RunnerLabelMapping struct {
	jobName    string
	oldLabel   string
	newLabel   string
	lineNum    int
	columnNum  int
	isArray    bool
	arrayIndex int
}

// findRunsOnNode finds the runs-on node for a job, handling both string and array formats
func findRunsOnNode(jobNode *yaml.Node) *yaml.Node {
	for i := 0; i < len(jobNode.Content); i += 2 {
		keyNode := jobNode.Content[i]
		if keyNode.Value == "runs-on" && i+1 < len(jobNode.Content) {
			return jobNode.Content[i+1]
		}
	}
	return nil
}

// ReplaceRunnerLabels replaces runner labels in a workflow based on the provided label map
// labelMap: map of old labels to new labels (e.g., "ubuntu-latest" -> "step-ubuntu-24")
// Returns: updated YAML string, bool indicating if changes were made, error if any
func ReplaceRunnerLabels(inputYaml string, labelMap map[string]string) (string, bool, error) {
	if len(labelMap) == 0 {
		return inputYaml, false, nil
	}

	// Parse the YAML into a tree structure
	t := yaml.Node{}
	err := yaml.Unmarshal([]byte(inputYaml), &t)
	if err != nil {
		return "", false, fmt.Errorf("unable to parse yaml: %v", err)
	}

	// Find all jobs node
	jobsNode := permissions.IterateNode(&t, "jobs", "!!map", 0)
	if jobsNode == nil {
		// No jobs found
		return inputYaml, false, nil
	}

	// Collect all the replacements we need to make
	var replacements []RunnerLabelMapping

	// Iterate through each job
	for i := 0; i < len(jobsNode.Content); i += 2 {
		jobNameNode := jobsNode.Content[i]
		jobNode := jobsNode.Content[i+1]

		jobName := jobNameNode.Value

		// Find the runs-on node for this job
		runsOnNode := findRunsOnNode(jobNode)
		if runsOnNode == nil {
			continue
		}

		// Handle both string and array formats
		switch runsOnNode.Kind {
		case yaml.ScalarNode:
			// Single runner label
			oldLabel := runsOnNode.Value
			if newLabel, ok := labelMap[oldLabel]; ok {
				replacements = append(replacements, RunnerLabelMapping{
					jobName:   jobName,
					oldLabel:  oldLabel,
					newLabel:  newLabel,
					lineNum:   runsOnNode.Line - 1, // Convert to 0-based
					columnNum: runsOnNode.Column - 1,
					isArray:   false,
				})
			}
		case yaml.SequenceNode:
			// Array of runner labels
			for idx, labelNode := range runsOnNode.Content {
				oldLabel := labelNode.Value
				if newLabel, ok := labelMap[oldLabel]; ok {
					replacements = append(replacements, RunnerLabelMapping{
						jobName:    jobName,
						oldLabel:   oldLabel,
						newLabel:   newLabel,
						lineNum:    labelNode.Line - 1, // Convert to 0-based
						columnNum:  labelNode.Column - 1,
						isArray:    true,
						arrayIndex: idx,
					})
				}
			}
		}
	}

	if len(replacements) == 0 {
		// No changes needed
		return inputYaml, false, nil
	}

	// Apply the replacements
	inputLines := strings.Split(inputYaml, "\n")
	updated := false

	for _, r := range replacements {
		if r.lineNum >= len(inputLines) {
			continue
		}

		oldLine := inputLines[r.lineNum]

		// Get the prefix (indentation + key)
		prefix := oldLine[:r.columnNum]

		// Replace the old label with the new one
		// We need to preserve any quotes, comments, etc.
		oldLineAfterColumn := oldLine[r.columnNum:]

		// Simple replacement - replace the first occurrence of the old label
		newLineAfterColumn := strings.Replace(oldLineAfterColumn, r.oldLabel, r.newLabel, 1)

		inputLines[r.lineNum] = prefix + newLineAfterColumn
		updated = true
	}

	output := strings.Join(inputLines, "\n")
	return output, updated, nil
}
