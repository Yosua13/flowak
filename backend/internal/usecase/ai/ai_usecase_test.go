package ai_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	domainAI "backend/internal/domain/ai"
	"backend/internal/infra/rabbitmq/producer"
	usecaseAI "backend/internal/usecase/ai"
)

// In-memory mock repository for AI jobs
type mockAIJobRepo struct {
	jobs map[string]*domainAI.AIJob
}

func newMockAIJobRepo() *mockAIJobRepo {
	return &mockAIJobRepo{jobs: make(map[string]*domainAI.AIJob)}
}

func (m *mockAIJobRepo) CreateJob(ctx context.Context, job *domainAI.AIJob) error {
	m.jobs[job.ID] = job
	return nil
}

func (m *mockAIJobRepo) GetJobByID(ctx context.Context, id string) (*domainAI.AIJob, error) {
	return m.jobs[id], nil
}

func (m *mockAIJobRepo) UpdateJobStatus(ctx context.Context, id string, status string, result any, errorMessage string) error {
	if job, ok := m.jobs[id]; ok {
		job.Status = status
		job.Result = result
		job.ErrorMessage = errorMessage
		job.UpdatedAt = time.Now().UTC()
	}
	return nil
}

func TestExecuteEnqueueAIFlow_AsyncMode(t *testing.T) {
	repo := newMockAIJobRepo()
	ctx := context.Background()

	// With nil channel, publisher gracefully succeeds
	pub := producer.NewAIJobPublisher(nil)

	job, err := usecaseAI.ExecuteEnqueueAIFlow(ctx, repo, pub, "Rancang alur login pengguna", "usr_test", true)
	if err != nil {
		t.Fatalf("unexpected error enqueuing AI flow: %v", err)
	}

	if job == nil || job.ID == "" {
		t.Fatal("expected valid job returned, got nil or empty ID")
	}

	if job.Status != domainAI.StatusPending {
		t.Fatalf("expected job status 'pending', got: %s", job.Status)
	}

	// Verify job status retrieval
	retrieved, err := usecaseAI.ExecuteGetAIJobStatus(ctx, repo, job.ID)
	if err != nil {
		t.Fatalf("unexpected error getting job status: %v", err)
	}
	if retrieved == nil || retrieved.ID != job.ID {
		t.Fatalf("expected retrieved job ID %s, got: %v", job.ID, retrieved)
	}
}

func TestExecuteProcessAIFlowJob_WithMockAICaller(t *testing.T) {
	repo := newMockAIJobRepo()
	ctx := context.Background()

	// Setup custom mock AI caller
	mockResponse := map[string]any{
		"name":        "Alur Autentikasi Pengguna",
		"description": "Deskripsi alur login dan verifikasi OTP",
		"nodes": []map[string]any{
			{
				"tempId": "node_1",
				"type":   "process",
				"label":  "Kirim Kredensial",
				"x":      100,
				"y":      200,
				"doc": map[string]any{
					"actor":   "Pengguna",
					"input":   "Email & Password",
					"process": "Validasi akun",
					"output":  "OTP token",
					"rules":   "Format email valid",
					"system":  "Auth Service",
					"sla":     "2 detik",
				},
				"roles": map[string]any{
					"uiux":     map[string]string{"assignee": "Rian", "status": "planned", "screen": "LoginScreen"},
					"frontend": map[string]string{"assignee": "Siti", "status": "planned", "component": "LoginForm", "route": "/login", "framework": "React"},
					"backend":  map[string]string{"assignee": "Budi", "status": "planned", "method": "POST", "endpoint": "/api/auth/login", "statusCode": "200"},
				},
			},
		},
		"edges": []map[string]any{},
	}
	rawJSON, _ := json.Marshal(mockResponse)

	originalCaller := usecaseAI.DefaultAICaller
	usecaseAI.DefaultAICaller = func(systemInstruction, prompt string, schema any) (string, error) {
		return string(rawJSON), nil
	}
	defer func() { usecaseAI.DefaultAICaller = originalCaller }()

	jobPayload := producer.AIJobPayload{
		JobID:     "job_exec_123",
		Type:      "generate_flow",
		Prompt:    "Buatkan alur login",
		UserID:    "usr_test",
		Timestamp: time.Now().UTC(),
	}

	resultJob, err := usecaseAI.ExecuteProcessAIFlowJob(ctx, repo, jobPayload)
	if err != nil {
		t.Fatalf("unexpected error processing AI flow: %v", err)
	}

	if resultJob.Status != domainAI.StatusCompleted {
		t.Fatalf("expected status completed, got: %s", resultJob.Status)
	}

	flowResult, ok := resultJob.Result.(usecaseAI.GeneratedFlowResult)
	if !ok {
		t.Fatalf("expected GeneratedFlowResult type, got: %T", resultJob.Result)
	}

	if flowResult.Name != "Alur Autentikasi Pengguna" {
		t.Fatalf("expected flow name 'Alur Autentikasi Pengguna', got: %s", flowResult.Name)
	}
	if len(flowResult.Nodes) != 1 {
		t.Fatalf("expected 1 node, got: %d", len(flowResult.Nodes))
	}
}
