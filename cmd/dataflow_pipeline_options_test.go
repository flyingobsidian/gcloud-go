package cmd

import (
	"encoding/json"
	"reflect"
	"testing"

	dataflow "google.golang.org/api/dataflow/v1b3"
)

func TestDataflowJobsRunPipelineOptionsFlag(t *testing.T) {
	if dataflowJobsRunCmd.Flags().Lookup("additional-pipeline-options") == nil {
		t.Fatal("jobs run missing --additional-pipeline-options")
	}
}

func TestDataflowTemplateRequestBody(t *testing.T) {
	req := &dataflow.CreateJobFromTemplateRequest{
		GcsPath:     "gs://b/t",
		JobName:     "j",
		Environment: &dataflow.RuntimeEnvironment{MaxWorkers: 3},
		Parameters:  map[string]string{"in": "x"},
	}
	body, err := dataflowTemplateRequestBody(req, []string{"a=1", "b=2"})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(body)
	var got map[string]any
	_ = json.Unmarshal(raw, &got)
	want := map[string]any{
		"gcsPath":    "gs://b/t",
		"jobName":    "j",
		"parameters": map[string]any{"in": "x"},
		"environment": map[string]any{
			"maxWorkers":                float64(3),
			"additionalPipelineOptions": []any{"a=1", "b=2"},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("body = %v\nwant %v", got, want)
	}

	body, err = dataflowTemplateRequestBody(&dataflow.CreateJobFromTemplateRequest{GcsPath: "gs://b/t"}, []string{"a=1"})
	if err != nil {
		t.Fatal(err)
	}
	env, _ := body["environment"].(map[string]any)
	if !reflect.DeepEqual(env["additionalPipelineOptions"], []string{"a=1"}) {
		t.Errorf("nil environment: %v", body)
	}
}
