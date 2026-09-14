// Package harbor parses Harbor job directories (config.json, result.json,
// trial subdirectories) into the metadata vRI persists at ingest.
package harbor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Job struct {
	Name        string   `json:"name"`
	Dir         string   `json:"dir"` // absolute path of the job directory
	Agent       string   `json:"agent"`
	Model       string   `json:"model,omitempty"`
	Environment string   `json:"environment"`
	TaskPaths   []string `json:"task_paths"`
	Trials      []Trial  `json:"trials"`
}

type Trial struct {
	Name      string   `json:"name"` // directory name, e.g. fail-close-verifier__jnmXFCb
	Task      string   `json:"task"` // task id: the name prefix before "__"
	Reward    *float64 `json:"reward,omitempty"`
	Exception string   `json:"exception,omitempty"`
}

type jobConfig struct {
	JobName     string `json:"job_name"`
	Environment struct {
		Type string `json:"type"`
	} `json:"environment"`
	Agents []struct {
		Name      string `json:"name"`
		ModelName string `json:"model_name"`
	} `json:"agents"`
	Tasks []struct {
		Path string `json:"path"`
	} `json:"tasks"`
}

type trialResult struct {
	TrialName string `json:"trial_name"`
	Config    struct {
		Agent struct {
			Name      string `json:"name"`
			ModelName string `json:"model_name"`
		} `json:"agent"`
	} `json:"config"`
	VerifierResult *struct {
		Rewards map[string]float64 `json:"rewards"`
	} `json:"verifier_result"`
	ExceptionInfo *struct {
		ExceptionType string `json:"exception_type"`
	} `json:"exception_info"`
}

// ParseJob reads a Harbor job directory. config.json and result.json at the
// job level must exist; every subdirectory containing a result.json is a
// trial named <task>__<suffix>.
func ParseJob(dir string) (*Job, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	cfgBytes, err := os.ReadFile(filepath.Join(abs, "config.json"))
	if err != nil {
		return nil, fmt.Errorf("harbor: %s: reading job config: %w", dir, err)
	}
	var cfg jobConfig
	if err := json.Unmarshal(cfgBytes, &cfg); err != nil {
		return nil, fmt.Errorf("harbor: %s: parsing job config.json: %w", dir, err)
	}
	if _, err := os.Stat(filepath.Join(abs, "result.json")); err != nil {
		return nil, fmt.Errorf("harbor: %s: missing job result.json", dir)
	}

	job := &Job{Dir: abs, Environment: cfg.Environment.Type}
	job.Name = cfg.JobName
	if job.Name == "" {
		job.Name = filepath.Base(abs)
	}
	for _, t := range cfg.Tasks {
		job.TaskPaths = append(job.TaskPaths, t.Path)
	}
	if len(cfg.Agents) > 0 {
		job.Agent = cfg.Agents[0].Name
		job.Model = cfg.Agents[0].ModelName
	}

	entries, err := os.ReadDir(abs)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if !e.IsDir() || !strings.Contains(e.Name(), "__") {
			continue
		}
		trialDir := filepath.Join(abs, e.Name())
		if _, err := os.Stat(filepath.Join(trialDir, "result.json")); err != nil {
			continue
		}
		trial, err := parseTrial(trialDir, e.Name())
		if err != nil {
			return nil, err
		}
		job.Trials = append(job.Trials, *trial)
	}
	sort.Slice(job.Trials, func(i, j int) bool { return job.Trials[i].Name < job.Trials[j].Name })

	// Oracle and nop jobs carry no job-level agents list; the trial config
	// still names the agent that ran.
	if job.Agent == "" && len(job.Trials) > 0 {
		var tr trialResult
		if err := readJSON(filepath.Join(abs, job.Trials[0].Name, "result.json"), &tr); err == nil {
			job.Agent = tr.Config.Agent.Name
			job.Model = tr.Config.Agent.ModelName
		}
	}
	return job, nil
}

func parseTrial(dir, name string) (*Trial, error) {
	trial := &Trial{Name: name, Task: name[:strings.Index(name, "__")]}
	var tr trialResult
	if err := readJSON(filepath.Join(dir, "result.json"), &tr); err != nil {
		return nil, fmt.Errorf("harbor: trial %s: %w", name, err)
	}
	if tr.VerifierResult != nil {
		if r, ok := tr.VerifierResult.Rewards["reward"]; ok {
			trial.Reward = &r
		}
	}
	if tr.ExceptionInfo != nil {
		trial.Exception = tr.ExceptionInfo.ExceptionType
	}
	if trial.Reward == nil {
		// Fall back to the verifier's reward file when the trial result
		// carries no verifier_result (e.g. an interrupted run).
		if b, err := os.ReadFile(filepath.Join(dir, "verifier", "reward.txt")); err == nil {
			if r, err := strconv.ParseFloat(strings.TrimSpace(string(b)), 64); err == nil {
				trial.Reward = &r
			}
		}
	}
	return trial, nil
}

// SemanticArtifacts lists the artifact paths (relative to the job dir) that
// ingest copies into the object store: the job-level config and result, then
// per trial config.json, result.json, verifier/*, and agent/trajectory.json.
// Large logs stay at their source path per the v0 contract.
func SemanticArtifacts(job *Job) ([]string, error) {
	paths := []string{"config.json", "result.json"}
	for _, t := range job.Trials {
		for _, f := range []string{"config.json", "result.json"} {
			paths = append(paths, t.Name+"/"+f)
		}
		verifierDir := filepath.Join(job.Dir, t.Name, "verifier")
		entries, err := os.ReadDir(verifierDir)
		if err != nil {
			return nil, fmt.Errorf("harbor: trial %s: reading verifier dir: %w", t.Name, err)
		}
		for _, e := range entries {
			if e.Type().IsRegular() {
				paths = append(paths, t.Name+"/verifier/"+e.Name())
			}
		}
		traj := t.Name + "/agent/trajectory.json"
		if _, err := os.Stat(filepath.Join(job.Dir, filepath.FromSlash(traj))); err == nil {
			paths = append(paths, traj)
		}
	}
	return paths, nil
}

func readJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}
