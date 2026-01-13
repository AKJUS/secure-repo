package runnerlabel

import (
	"io/ioutil"
	"path"
	"testing"
)

func TestReplaceRunnerLabels(t *testing.T) {
	const inputDirectory = "../../../testfiles/runnerLabel/input"
	const outputDirectory = "../../../testfiles/runnerLabel/output"

	tests := []struct {
		name        string
		inputFile   string
		outputFile  string
		labelMap    map[string]string
		wantUpdated bool
		wantErr     bool
	}{
		{
			name:       "single job with ubuntu-latest",
			inputFile:  "singleJob.yml",
			outputFile: "singleJob.yml",
			labelMap: map[string]string{
				"ubuntu-latest": "step-ubuntu-24",
			},
			wantUpdated: true,
			wantErr:     false,
		},
		{
			name:       "multiple jobs with different ubuntu versions",
			inputFile:  "multipleJobs.yml",
			outputFile: "multipleJobs.yml",
			labelMap: map[string]string{
				"ubuntu-22.04":  "step-ubuntu-22",
				"ubuntu-24.04":  "step-ubuntu-24",
				"ubuntu-latest": "step-ubuntu-24",
			},
			wantUpdated: true,
			wantErr:     false,
		},
		{
			name:       "array of runners",
			inputFile:  "arrayRunners.yml",
			outputFile: "arrayRunners.yml",
			labelMap: map[string]string{
				"ubuntu-latest": "step-ubuntu-24",
			},
			wantUpdated: true,
			wantErr:     false,
		},
		{
			name:       "multiple array items to replace",
			inputFile:  "multipleArrayItems.yml",
			outputFile: "multipleArrayItems.yml",
			labelMap: map[string]string{
				"ubuntu-latest":  "step-ubuntu-24",
				"windows-latest": "step-windows",
			},
			wantUpdated: true,
			wantErr:     false,
		},
		{
			name:       "inline array syntax",
			inputFile:  "inlineArray.yml",
			outputFile: "inlineArray.yml",
			labelMap: map[string]string{
				"ubuntu-latest": "step-ubuntu-24",
			},
			wantUpdated: true,
			wantErr:     false,
		},
		{
			name:       "compact ubuntu version numbers",
			inputFile:  "compactVersions.yml",
			outputFile: "compactVersions.yml",
			labelMap: map[string]string{
				"ubuntu-22": "step-ubuntu-22",
				"ubuntu-24": "step-ubuntu-24",
			},
			wantUpdated: true,
			wantErr:     false,
		},
		{
			name:       "no changes needed - already using custom runners",
			inputFile:  "noChangesNeeded.yml",
			outputFile: "noChangesNeeded.yml",
			labelMap: map[string]string{
				"ubuntu-latest": "step-ubuntu-24",
			},
			wantUpdated: false,
			wantErr:     false,
		},
		{
			name:       "comprehensive test with all scenarios",
			inputFile:  "comprehensive.yml",
			outputFile: "comprehensive.yml",
			labelMap: map[string]string{
				"ubuntu-latest":  "step-ubuntu-24",
				"ubuntu-24":      "step-ubuntu-24",
				"ubuntu-22":      "step-ubuntu-22",
				"windows-latest": "step-windows",
			},
			wantUpdated: true,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read input file
			input, err := ioutil.ReadFile(path.Join(inputDirectory, tt.inputFile))
			if err != nil {
				t.Fatalf("error reading input file: %v", err)
			}

			// Run the function
			got, updated, err := ReplaceRunnerLabels(string(input), tt.labelMap)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("ReplaceRunnerLabels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check if updated flag matches
			if updated != tt.wantUpdated {
				t.Errorf("ReplaceRunnerLabels() updated = %v, wantUpdated %v", updated, tt.wantUpdated)
			}

			// Read expected output file
			expectedOutput, err := ioutil.ReadFile(path.Join(outputDirectory, tt.outputFile))
			if err != nil {
				t.Fatalf("error reading expected output file: %v", err)
			}

			// Compare output with expected
			if got != string(expectedOutput) {
				t.Errorf("ReplaceRunnerLabels() output mismatch\nGot:\n%s\n\nWant:\n%s", got, string(expectedOutput))
			}
		})
	}
}

func TestReplaceRunnerLabels_InvalidYAML(t *testing.T) {
	invalidYaml := `name: Test Workflow
on: [push
jobs:
  test:
    runs-on: ubuntu-latest
`
	labelMap := map[string]string{
		"ubuntu-latest": "step-ubuntu-24",
	}

	_, _, err := ReplaceRunnerLabels(invalidYaml, labelMap)
	if err == nil {
		t.Errorf("ReplaceRunnerLabels() expected error for invalid YAML, got nil")
	}
}

func TestReplaceRunnerLabels_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		inputYaml   string
		labelMap    map[string]string
		wantUpdated bool
		wantErr     bool
	}{
		{
			name: "empty label map",
			inputYaml: `name: Test Workflow
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
`,
			labelMap:    map[string]string{},
			wantUpdated: false,
			wantErr:     false,
		},
		{
			name: "workflow without jobs",
			inputYaml: `name: Test Workflow
on: [push]
`,
			labelMap: map[string]string{
				"ubuntu-latest": "step-ubuntu-24",
			},
			wantUpdated: false,
			wantErr:     false,
		},
		{
			name: "job without runs-on",
			inputYaml: `name: Test Workflow
on: [push]
jobs:
  test:
    container: ubuntu:latest
    steps:
      - uses: actions/checkout@v2
`,
			labelMap: map[string]string{
				"ubuntu-latest": "step-ubuntu-24",
			},
			wantUpdated: false,
			wantErr:     false,
		},
		{
			name: "no matching labels",
			inputYaml: `name: Test Workflow
on: [push]
jobs:
  test:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v2
`,
			labelMap: map[string]string{
				"ubuntu-latest": "step-ubuntu-24",
			},
			wantUpdated: false,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, updated, err := ReplaceRunnerLabels(tt.inputYaml, tt.labelMap)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("ReplaceRunnerLabels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check if updated flag matches
			if updated != tt.wantUpdated {
				t.Errorf("ReplaceRunnerLabels() updated = %v, wantUpdated %v", updated, tt.wantUpdated)
			}

			// For edge cases where no changes expected, input should equal output
			if !tt.wantUpdated && got != tt.inputYaml {
				t.Errorf("ReplaceRunnerLabels() expected no changes but output differs from input")
			}
		})
	}
}
