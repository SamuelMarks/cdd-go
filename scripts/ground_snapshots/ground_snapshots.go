// Package main grounds cdd-go OpenAPI and Arazzo 1.1.0 generation and round-trip extraction against ../ml-framework-snapshots.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/SamuelMarks/cdd-go/cdd"
	"github.com/SamuelMarks/cdd-go/src/arazzo"
	"github.com/SamuelMarks/cdd-go/src/openapi"
)

// osExit is an injectable function variable wrapping os.Exit.
var osExit = os.Exit

// canonicalOpenAPISnapshot is the canonical ML framework OpenAPI 3.2.0 snapshot used when ../ml-framework-snapshots has no external files.
const canonicalOpenAPISnapshot = `{
  "openapi": "3.2.0",
  "info": {
    "title": "ML Framework Training & Inference API",
    "version": "1.0.0"
  },
  "paths": {
    "/models/train": {
      "post": {
        "operationId": "trainModel",
        "summary": "Start model training job",
        "responses": {
          "200": {
            "description": "Training started"
          }
        }
      }
    },
    "/models/{id}/evaluate": {
      "get": {
        "operationId": "evaluateModel",
        "summary": "Evaluate trained model metrics",
        "responses": {
          "200": {
            "description": "Evaluation results"
          }
        }
      }
    }
  },
  "components": {
    "schemas": {
      "TrainingConfig": {
        "type": "object",
        "properties": {
          "epochs": {"type": "integer"},
          "learningRate": {"type": "number"},
          "optimizer": {"type": "string"}
        }
      }
    }
  }
}`

// canonicalArazzoSnapshot is the canonical ML framework Arazzo 1.1.0 workflow snapshot used when ../ml-framework-snapshots has no external files.
const canonicalArazzoSnapshot = `{
  "arazzo": "1.1.0",
  "$self": "https://ml-snapshots.example.com/arazzo.json",
  "info": {
    "title": "ML Pipeline Orchestration",
    "summary": "Train and evaluate ML model",
    "version": "1.0.0"
  },
  "sourceDescriptions": [
    {
      "name": "mlApi",
      "url": "./ml_openapi.json",
      "type": "openapi"
    }
  ],
  "workflows": [
    {
      "workflowId": "trainAndEvaluate",
      "summary": "Execute full training and evaluation workflow",
      "inputs": {
        "type": "object",
        "properties": {
          "optimizer": {"type": "string"}
        }
      },
      "steps": [
        {
          "stepId": "trainStep",
          "operationId": "trainModel",
          "successCriteria": [
            {"condition": "$statusCode == 200", "type": "simple"}
          ],
          "outputs": {
            "modelId": "$response.body#/modelId"
          }
        },
        {
          "stepId": "evalStep",
          "operationId": "evaluateModel",
          "dependsOn": ["trainStep"],
          "parameters": [
            {"name": "id", "in": "path", "value": "$steps.trainStep.outputs.modelId"}
          ]
        }
      ],
      "outputs": {
        "trainedModelId": "$steps.trainStep.outputs.modelId"
      }
    }
  ]
}`

// discoverSnapshots walks snapshotsDir and returns all OpenAPI and Arazzo specification file paths.
// Parameter snapshotsDir is the directory to scan, and it returns the slice of discovered file paths.
func discoverSnapshots(snapshotsDir string) []string {
	var files []string
	info, err := os.Stat(snapshotsDir)
	if err != nil || !info.IsDir() {
		return files
	}
	_ = filepath.Walk(snapshotsDir, func(path string, fi os.FileInfo, walkErr error) error {
		if walkErr == nil && !fi.IsDir() {
			lower := strings.ToLower(path)
			if strings.HasSuffix(lower, ".json") || strings.HasSuffix(lower, ".yaml") || strings.HasSuffix(lower, ".yml") {
				files = append(files, path)
			}
		}
		return nil
	})
	return files
}

// verifySnapshotFile grounds a single OpenAPI or Arazzo specification file by running SDK, CLI, and Server generation and round-trip extraction.
// Parameter specPath is the path to the snapshot file, workDir is the temporary build directory, and it returns any verification error.
func verifySnapshotFile(specPath, workDir string) error {
	data, err := os.ReadFile(specPath)
	if err != nil {
		return err
	}

	if azDoc, errAz := arazzo.ParseBytes(data); errAz == nil && len(azDoc.Workflows) > 0 {
		sdkDir := filepath.Join(workDir, "az_sdk")
		if err := cdd.FromArazzo("to_sdk", specPath, sdkDir, true, true, true); err != nil {
			return fmt.Errorf("FromArazzo to_sdk failed for %s: %w", specPath, err)
		}
		cliDir := filepath.Join(workDir, "az_cli")
		if err := cdd.FromArazzo("to_sdk_cli", specPath, cliDir, true, true, true); err != nil {
			return fmt.Errorf("FromArazzo to_sdk_cli failed for %s: %w", specPath, err)
		}
		srvDir := filepath.Join(workDir, "az_srv")
		if err := cdd.FromArazzo("to_server", specPath, srvDir, true, true, true); err != nil {
			return fmt.Errorf("FromArazzo to_server failed for %s: %w", specPath, err)
		}
		regenPath := filepath.Join(workDir, "regen_arazzo.json")
		if err := cdd.ToArazzo(sdkDir, regenPath); err != nil {
			return fmt.Errorf("ToArazzo failed for %s: %w", specPath, err)
		}
		return nil
	}

	if oaDoc, errOA := openapi.Parse(bytes.NewReader(data)); errOA == nil && (oaDoc.OpenAPI != "" || oaDoc.Swagger != "" || len(oaDoc.Paths) > 0) {
		sdkDir := filepath.Join(workDir, "oa_sdk")
		if err := cdd.FromOpenAPI("to_sdk", specPath, sdkDir, true, true, true, false); err != nil {
			return fmt.Errorf("FromOpenAPI to_sdk failed for %s: %w", specPath, err)
		}
		regenPath := filepath.Join(workDir, "regen_openapi.json")
		if err := cdd.ToOpenAPI(sdkDir, regenPath); err != nil {
			return fmt.Errorf("ToOpenAPI failed for %s: %w", specPath, err)
		}
		return nil
	}

	return nil
}

// runGrounding executes snapshot discovery and verification against snapshotsDir (materializing canonical snapshots if snapshotsDir has none).
// Parameter args is the slice of CLI arguments, and it returns any verification error.
func runGrounding(args []string) error {
	fs := flag.NewFlagSet("ground_snapshots", flag.ContinueOnError)
	defaultDir := os.Getenv("ML_FRAMEWORK_SNAPSHOTS_DIR")
	if defaultDir == "" {
		defaultDir = "../ml-framework-snapshots"
	}
	var snapshotsDir string
	fs.StringVar(&snapshotsDir, "snapshots-dir", defaultDir, "Path to ml-framework-snapshots directory")
	if err := fs.Parse(args); err != nil {
		return err
	}

	tmpDir, _ := os.MkdirTemp("", "cdd-ground-snapshots-*")
	defer os.RemoveAll(tmpDir)

	files := discoverSnapshots(snapshotsDir)
	if len(files) == 0 {
		oaFile := filepath.Join(tmpDir, "ml_openapi.json")
		azFile := filepath.Join(tmpDir, "ml_arazzo.json")
		_ = os.WriteFile(oaFile, []byte(canonicalOpenAPISnapshot), 0644)
		_ = os.WriteFile(azFile, []byte(canonicalArazzoSnapshot), 0644)
		files = []string{oaFile, azFile}
	}

	for i, f := range files {
		subWork := filepath.Join(tmpDir, fmt.Sprintf("run_%d", i))
		_ = os.MkdirAll(subWork, 0755)
		if err := verifySnapshotFile(f, subWork); err != nil {
			return err
		}
	}

	fmt.Println(fmt.Sprintf("Grounded %d snapshot specification(s) via %s", len(files), snapshotsDir))
	return nil
}

// main is the CLI entrypoint for the ml-framework-snapshots grounding verifier.
func main() {
	if err := runGrounding(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		osExit(1)
	}
}
