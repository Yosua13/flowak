package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"backend/config"
)

// Gemini API REST payload structs
type GeminiPart struct {
	Text string `json:"text"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiSystemInstruction struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiConfig struct {
	ResponseMimeType string `json:"responseMimeType,omitempty"`
	ResponseSchema   any    `json:"responseSchema,omitempty"`
}

type GeminiRequest struct {
	Contents          []GeminiContent          `json:"contents"`
	SystemInstruction *GeminiSystemInstruction `json:"systemInstruction,omitempty"`
	GenerationConfig  *GeminiConfig            `json:"generationConfig,omitempty"`
}

type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// CallGemini REST helper
func CallGemini(systemInstruction, prompt string, schema any) (string, error) {
	apiKey := config.ActiveConfig.GeminiAPIKey
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY is empty")
	}

	model := "gemini-3.5-flash"
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, apiKey)

	reqPayload := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{{Text: prompt}},
			},
		},
	}

	if systemInstruction != "" {
		reqPayload.SystemInstruction = &GeminiSystemInstruction{
			Parts: []GeminiPart{{Text: systemInstruction}},
		}
	}

	if schema != nil {
		reqPayload.GenerationConfig = &GeminiConfig{
			ResponseMimeType: "application/json",
			ResponseSchema:   schema,
		}
	}

	reqBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	// Set User-Agent to match typical builder requests
	req.Header.Set("User-Agent", "aistudio-build")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Gemini API error (Status %d): %s", resp.StatusCode, string(respBytes))
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(respBytes, &geminiResp); err != nil {
		return "", err
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty candidates in Gemini response")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

// 1. Generate Flow Handler (POST /api/ai/generate-flow)
func AiGenerateFlowHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error": "Method not allowed"}`))
		return
	}

	var req struct {
		Prompt string `json:"prompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Prompt == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Permintaan prompt tidak boleh kosong"}`))
		return
	}

	// Fallback to offline demo flow if key is empty
	if config.ActiveConfig.GeminiAPIKey == "" {
		w.Write(getOfflineDemoFlow(req.Prompt))
		return
	}

	// Call Gemini with schema
	systemInstruction := `Anda adalah ahli pemodelan proses bisnis dan arsitek sistem teknis Flowak.
Tugas Anda mendesain alur kerja (workflow) lengkap dalam bentuk graf Node (langkah proses) dan Edge (sambungan/aliran) berdasarkan deskripsi pengguna.

Aturan Penting:
1. Buat 4 hingga 8 langkah (nodes) yang logis, berurutan, dan menyertakan aspek bisnis (actor, input, process, output, sla) serta aspek teknis untuk UI/UX, Frontend, dan Backend.
2. Tempatkan koordinat x dan y (grid kelipatan 40px) sedemikian rupa agar membentuk diagram alir yang indah dan tidak tumpang tindih. (Misal: Node pertama di x: 80, y: 150. Node berikutnya bergeser ke kanan x: 380, y: 150, dst. Jika ada percabangan, Anda bisa menyebarkan y).
3. Sambungkan node dengan Edge yang merujuk pada id node yang sesuai.
4. Pastikan response strictly valid JSON sesuai tipe skema yang diberikan.`

	promptTemplate := fmt.Sprintf(`Deskripsi Kebutuhan: "%s"

Silakan buatkan struktur node proses bisnis dan edges sesuai dengan deskripsi tersebut.`, req.Prompt)

	schema := map[string]any{
		"type": "OBJECT",
		"properties": map[string]any{
			"name": map[string]any{
				"type":        "STRING",
				"description": "Nama modul alur kerja yang representatif dan profesional",
			},
			"description": map[string]any{
				"type":        "STRING",
				"description": "Deskripsi singkat mengenai alur kerja fungsional ini",
			},
			"nodes": map[string]any{
				"type":        "ARRAY",
				"description": "Daftar node dalam diagram alir",
				"items": map[string]any{
					"type": "OBJECT",
					"properties": map[string]any{
						"tempId":    map[string]any{"type": "STRING", "description": "ID sementara node (contoh: 'node_1', 'node_2')"},
						"type":      map[string]any{"type": "STRING", "description": "Tipe node: 'terminator' | 'process' | 'decision' | 'actor' | 'system'"},
						"label":     map[string]any{"type": "STRING", "description": "Label singkat representatif langkah alur (contoh: 'Validasi OTP')"},
						"x":         map[string]any{"type": "NUMBER", "description": "Koordinat horizontal di kanvas (kisaran 100 s.d 1400)"},
						"y":         map[string]any{"type": "NUMBER", "description": "Koordinat vertikal di kanvas (kisaran 100 s.d 800)"},
						"doc": map[string]any{
							"type":        "OBJECT",
							"description": "Aspek analisis fungsional bisnis",
							"properties": map[string]any{
								"actor":   map[string]any{"type": "STRING", "description": "Aktor utama (contoh: 'Pelanggan', 'Sistem', 'Admin')"},
								"input":   map[string]any{"type": "STRING"},
								"process": map[string]any{"type": "STRING"},
								"output":  map[string]any{"type": "STRING"},
								"rules":   map[string]any{"type": "STRING"},
								"system":  map[string]any{"type": "STRING"},
								"sla":     map[string]any{"type": "STRING"},
							},
						},
						"roles": map[string]any{
							"type": "OBJECT",
							"properties": map[string]any{
								"uiux": map[string]any{
									"type": "OBJECT",
									"properties": map[string]any{
										"assignee": map[string]any{"type": "STRING"},
										"status":   map[string]any{"type": "STRING"},
										"screen":   map[string]any{"type": "STRING"},
									},
								},
								"frontend": map[string]any{
									"type": "OBJECT",
									"properties": map[string]any{
										"assignee":  map[string]any{"type": "STRING"},
										"status":    map[string]any{"type": "STRING"},
										"component": map[string]any{"type": "STRING"},
										"route":     map[string]any{"type": "STRING"},
										"framework": map[string]any{"type": "STRING"},
									},
								},
								"backend": map[string]any{
									"type": "OBJECT",
									"properties": map[string]any{
										"assignee":   map[string]any{"type": "STRING"},
										"status":     map[string]any{"type": "STRING"},
										"method":     map[string]any{"type": "STRING"},
										"endpoint":   map[string]any{"type": "STRING"},
										"statusCode": map[string]any{"type": "STRING"},
									},
								},
							},
						},
					},
					"required": []string{"tempId", "type", "label", "x", "y"},
				},
			},
			"edges": map[string]any{
				"type":        "ARRAY",
				"description": "Daftar sambungan atau garis penghubung antar node alur kerja",
				"items": map[string]any{
					"type": "OBJECT",
					"properties": map[string]any{
						"from":  map[string]any{"type": "STRING"},
						"to":    map[string]any{"type": "STRING"},
						"label": map[string]any{"type": "STRING"},
					},
					"required": []string{"from", "to"},
				},
			},
		},
		"required": []string{"name", "description", "nodes", "edges"},
	}

	resultText, err := CallGemini(systemInstruction, promptTemplate, schema)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err.Error())))
		return
	}

	// Parse result JSON and map temporary IDs to production UUIDs
	var rawFlow struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Nodes       []struct {
			TempID string `json:"tempId"`
			Type   string `json:"type"`
			Label  string `json:"label"`
			X      int    `json:"x"`
			Y      int    `json:"y"`
			Doc    struct {
				Actor   string `json:"actor"`
				Input   string `json:"input"`
				Process string `json:"process"`
				Output  string `json:"output"`
				Rules   string `json:"rules"`
				System  string `json:"system"`
				Sla     string `json:"sla"`
			} `json:"doc"`
			Roles struct {
				Uiux struct {
					Assignee string `json:"assignee"`
					Status   string `json:"status"`
					Screen   string `json:"screen"`
				} `json:"uiux"`
				Frontend struct {
					Assignee  string `json:"assignee"`
					Status    string `json:"status"`
					Component string `json:"component"`
					Route     string `json:"route"`
					Framework string `json:"framework"`
				} `json:"frontend"`
				Backend struct {
					Assignee   string `json:"assignee"`
					Status     string `json:"status"`
					Method     string `json:"method"`
					Endpoint   string `json:"endpoint"`
					StatusCode string `json:"statusCode"`
				} `json:"backend"`
			} `json:"roles"`
		} `json:"nodes"`
		Edges []struct {
			From  string `json:"from"`
			To    string `json:"to"`
			Label string `json:"label"`
		} `json:"edges"`
	}

	if err := json.Unmarshal([]byte(resultText), &rawFlow); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Failed to parse Gemini output"}`))
		return
	}

	// Map tempIds to secure node UUIDs
	idMap := make(map[string]string)
	type ResponseNode struct {
		ID    string `json:"id"`
		Type  string `json:"type"`
		Label string `json:"label"`
		X     int    `json:"x"`
		Y     int    `json:"y"`
		Doc   any    `json:"doc"`
		Roles any    `json:"roles"`
	}

	respNodes := []ResponseNode{}
	for _, n := range rawFlow.Nodes {
		generatedID := "node_" + GenerateUUID()
		idMap[n.TempID] = generatedID

		// Default assignees
		uiuxAssignee := n.Roles.Uiux.Assignee
		if uiuxAssignee == "" {
			uiuxAssignee = "Rian"
		}
		feAssignee := n.Roles.Frontend.Assignee
		if feAssignee == "" {
			feAssignee = "Siti"
		}
		beAssignee := n.Roles.Backend.Assignee
		if beAssignee == "" {
			beAssignee = "Budi"
		}

		respNodes = append(respNodes, ResponseNode{
			ID:    generatedID,
			Type:  n.Type,
			Label: n.Label,
			X:     n.X,
			Y:     n.Y,
			Doc: map[string]string{
				"actor":   n.Doc.Actor,
				"input":   n.Doc.Input,
				"process": n.Doc.Process,
				"output":  n.Doc.Output,
				"rules":   n.Doc.Rules,
				"system":  n.Doc.System,
				"sla":     n.Doc.Sla,
			},
			Roles: map[string]any{
				"uiux": map[string]string{
					"assignee": uiuxAssignee,
					"status":   "planned",
					"screen":   n.Roles.Uiux.Screen,
					"link":     "",
				},
				"frontend": map[string]string{
					"assignee":  feAssignee,
					"status":    "planned",
					"component": n.Roles.Frontend.Component,
					"route":     n.Roles.Frontend.Route,
					"framework": n.Roles.Frontend.Framework,
					"link":      "",
				},
				"backend": map[string]string{
					"assignee":   beAssignee,
					"status":     "planned",
					"method":     n.Roles.Backend.Method,
					"endpoint":   n.Roles.Backend.Endpoint,
					"auth":       "",
					"request":    "",
					"response":   "",
					"statusCode": n.Roles.Backend.StatusCode,
				},
			},
		})
	}

	type ResponseEdge struct {
		ID    string `json:"id"`
		From  string `json:"from"`
		To    string `json:"to"`
		Label string `json:"label"`
	}

	respEdges := []ResponseEdge{}
	for _, e := range rawFlow.Edges {
		fromID := idMap[e.From]
		toID := idMap[e.To]
		if fromID == "" || toID == "" {
			continue
		}
		respEdges = append(respEdges, ResponseEdge{
			ID:    "edge_" + GenerateUUID(),
			From:  fromID,
			To:    toID,
			Label: e.Label,
		})
	}

	w.Write([]byte(fmt.Sprintf(`{
		"status": "complete",
		"name": %q,
		"description": %q,
		"nodes": %s,
		"edges": %s
	}`, rawFlow.Name, rawFlow.Description, toJSONString(respNodes), toJSONString(respEdges))))
}

// 2. Generate AI Mock Payload (POST /api/ai/mock-payload)
func AiMockPayloadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error": "Method not allowed"}`))
		return
	}

	var req struct {
		Actor    string `json:"actor"`
		Label    string `json:"label"`
		Endpoint string `json:"endpoint"`
		Process  string `json:"process"`
		Method   string `json:"method"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Invalid request body"}`))
		return
	}

	// Fallback to offline mock response if key is empty
	if config.ActiveConfig.GeminiAPIKey == "" {
		w.Write([]byte(fmt.Sprintf(`{
			"request": "{\n  \"deviceId\": \"MOCK_DEVICE_ID\",\n  \"channel\": \"Web\",\n  \"action\": %q\n}",
			"response": "{\n  \"success\": true,\n  \"timestamp\": %q,\n  \"data\": {\n    \"step\": %q,\n    \"actor\": %q,\n    \"simulated\": true\n  }\n}"
		}`, req.Label, time.Now().Format(time.RFC3339), req.Label, req.Actor)))
		return
	}

	prompt := fmt.Sprintf(`Definisikan payload JSON Request Body dan JSON Response Body fungsional, realistis untuk API Endpoint berikut:
- Tipe Langkah / Proses Bisnis: %s
- Aktor Utama: %s
- Jalur Endpoint: %s
- Metoda HTTP: %s
- Deskripsi Alur: %s

Keluarkan data dalam format skema JSON yang berisi dua properti string: 'request' (berisi teks payload JSON request) dan 'response' (berisi teks payload JSON response).`,
		req.Label, req.Actor, req.Endpoint, req.Method, req.Process)

	schema := map[string]any{
		"type": "OBJECT",
		"properties": map[string]any{
			"request":  map[string]any{"type": "STRING", "description": "JSON string request body, dikosongkan jika GET"},
			"response": map[string]any{"type": "STRING", "description": "JSON string response dari server"},
		},
		"required": []string{"request", "response"},
	}

	resultText, err := CallGemini("", prompt, schema)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err.Error())))
		return
	}

	w.Write([]byte(resultText))
}

// 3. AI Spec Reviewer & Auditor (POST /api/ai/audit-flow)
func AiAuditFlowHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error": "Method not allowed"}`))
		return
	}

	var req struct {
		Module any `json:"module"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Module == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Modul harus disertakan."}`))
		return
	}

	// Fallback to offline mock response if key is empty
	if config.ActiveConfig.GeminiAPIKey == "" {
		w.Write([]byte(`{
			"score": 85,
			"summary": "Audit offline demo dilakukan. Struktur modul valid namun integrasi AI sesungguhnya tidak aktif karena API KEY kosong.",
			"issues": [
				{
					"severity": "warning",
					"message": "Rancang deskripsi alur kerja yang mendalam untuk meningkatkan akurasi koding.",
					"nodeName": "Evaluasi Awal",
					"type": "Dokumentasi"
				}
			]
		}`))
		return
	}

	moduleJSON, _ := json.Marshal(req.Module)
	prompt := fmt.Sprintf(`Analisis kualitas rancangan spesifikasi fungsional modul alur kerja berikut ini.
%s

Deteksi potensi kecacatan perancangan (desain flaw) seperti:
- SLA Bottleneck: Langkah manual atau terlampau lambat.
- Security Flaw: Endpoint backend sensitif yang tidak memiliki otorisasi token.
- Dead End: Langkah alur yang tidak terhubung ke node akhir mana pun terputus.
- Missing Info: Node yang aspek proses bisnisnya kosong.

Berikan keluaran hasil audit Anda dalam format skema JSON terstruktur berisi:
1. 'score' (angka integer 0-100 penilaian kelayakan rancangan)
2. 'summary' (paragraf ringkasan naratif tinjauan kualitatif)
3. 'issues' (array berisi daftar masalah/saran perbaikan)`, string(moduleJSON))

	schema := map[string]any{
		"type": "OBJECT",
		"properties": map[string]any{
			"score":   map[string]any{"type": "INTEGER", "description": "Skor audit rancangan antara 0 - 100"},
			"summary": map[string]any{"type": "STRING", "description": "Ringkasan deskriptif analitis tingkat tinggi perihal rancangan alur ini"},
			"issues": map[string]any{
				"type": "ARRAY",
				"items": map[string]any{
					"type": "OBJECT",
					"properties": map[string]any{
						"severity": map[string]any{"type": "STRING", "description": "Tingkat keparahan: 'error' (Kritikal) | 'warning' (Peringatan) | 'info' (Catatan)"},
						"message":  map[string]any{"type": "STRING", "description": "Rincian detail temuan audit serta rekomendasi solusi pemulihan"},
						"nodeName": map[string]any{"type": "STRING", "description": "Nama langkah terkait jika temuan spesifik ke suatu node (opsional)"},
						"type":     map[string]any{"type": "STRING", "description": "Kategori masalah: 'SLA' | 'Keamanan' | 'Logika Alur' | 'Dokumentasi'"},
					},
					"required": []string{"severity", "message", "type"},
				},
			},
		},
		"required": []string{"score", "summary", "issues"},
	}

	resultText, err := CallGemini("", prompt, schema)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err.Error())))
		return
	}

	w.Write([]byte(resultText))
}

// 4. AI Boilerplate Code Generator (POST /api/ai/generate-code)
func AiGenerateCodeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error": "Method not allowed"}`))
		return
	}

	var req struct {
		Type      string `json:"type"`
		NodeLabel string `json:"nodeLabel"`
		NodeType  string `json:"nodeType"`
		Detail    any    `json:"detail"`
		Framework string `json:"framework"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Invalid request body"}`))
		return
	}

	// Fallback to offline mock response if key is empty
	if config.ActiveConfig.GeminiAPIKey == "" {
		var code string
		if req.Type == "frontend" {
			code = fmt.Sprintf(`// Offline Demo React Component for %s
import React from 'react';

export default function %sView() {
  return (
    <div className="p-6 bg-slate-900 text-white rounded-xl border border-white/5">
      <h3 className="text-lg font-bold mb-2">%s</h3>
      <p className="text-xs text-slate-400">Tipe Node: %s</p>
      <button className="mt-4 px-4 py-2 bg-emerald-600 rounded text-xs font-bold hover:bg-emerald-500">
        Jalankan Aksi
      </button>
    </div>
  );
}`, req.NodeLabel, strings.ReplaceAll(req.NodeLabel, " ", ""), req.NodeLabel, req.NodeType)
		} else {
			code = fmt.Sprintf(`// Offline Demo Express API for %s
import { Request, Response } from 'express';

export async function handle%s(req: Request, res: Response) {
  try {
    console.log("Simulasi memproses %s");
    return res.status(200).json({
      success: true,
      message: "Proses berhasil disimulasikan",
      timestamp: new Date().toISOString()
    });
  } catch (error) {
    return res.status(500).json({ error: "Internal server error" });
  }
}`, req.NodeLabel, strings.ReplaceAll(req.NodeLabel, " ", ""), req.NodeLabel)
		}

		w.Write([]byte(fmt.Sprintf(`{
			"code": %q,
			"language": %q
		}`, code, func() string {
			if req.Type == "frontend" {
				return "tsx"
			}
			return "typescript"
		}())))
		return
	}

	systemInstruction := ""
	prompt := ""
	detailJSON, _ := json.Marshal(req.Detail)

	if req.Type == "frontend" {
		systemInstruction = "Anda adalah pengembang Frontend kelas dunia ahli React, Typescript, dan Tailwind CSS."
		prompt = fmt.Sprintf(`Tulis kode boilerplate komponen React yang fungsional, bersih, interaktif, dan terstyling indah menggunakan Tailwind CSS.
Langkah Proses: "%s" (Tipe: %s)
Detail: %s
Target Layout/Kerangka: "%s"

Desainlah komponen ini agar menyertakan visualisasi form input, status loading, dan tombol aksi interaktif yang relevan dengan proses bisnis di atas secara simulatif dan modern.`,
			req.NodeLabel, req.NodeType, string(detailJSON), req.Framework)
	} else {
		systemInstruction = "Anda adalah pengembang API Backend kelas dunia ahli Node.js Express.js dan TypeScript."
		prompt = fmt.Sprintf(`Tulis kode boilerplate file controller / endpoint Express.js TypeScript yang bersih, aman, dan modular.
Langkah Proses: "%s" (Tipe: %s)
Detail: %s

Sertakan parsing, handling validasi sederhana, dan try-catch error response block.`,
			req.NodeLabel, req.NodeType, string(detailJSON))
	}

	// For general text output generation without schema constraints, we don't pass schema object
	resultText, err := CallGemini(systemInstruction, prompt, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err.Error())))
		return
	}

	// Parse text to extract code from markdown block if returned
	code := resultText
	// Strip markdown blocks if returned
	if strings.Contains(code, "```") {
		parts := strings.Split(code, "```")
		for _, part := range parts {
			// Find the block containing tsx/typescript/javascript code
			if strings.HasPrefix(part, "tsx\n") || strings.HasPrefix(part, "typescript\n") || strings.HasPrefix(part, "javascript\n") {
				lines := strings.Split(part, "\n")
				code = strings.Join(lines[1:], "\n")
				break
			}
		}
	}

	w.Write([]byte(fmt.Sprintf(`{
		"code": %q,
		"language": %q
	}`, code, func() string {
		if req.Type == "frontend" {
			return "tsx"
		}
		return "typescript"
	}())))
}

// Helpers
func toJSONString(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func getOfflineDemoFlow(prompt string) []byte {
	desc := strings.ToLower(prompt)
	var flowJSON string

	if strings.Contains(desc, "otp") || strings.Contains(desc, "login") || strings.Contains(desc, "auth") {
		flowJSON = `{
			"status": "complete",
			"name": "Sistem Otentikasi & Verifikasi OTP WA",
			"description": "Alur terpadu autentikasi user beserta lapisan pencegahan fraud via kode OTP WhatsApp.",
			"nodes": [
				{
					"id": "node_otp_1",
					"type": "terminator",
					"label": "Halaman Login Dimulai",
					"x": 100,
					"y": 200,
					"doc": { "actor": "User", "input": "Email & Sandi", "process": "User membuka app dan ketik kredensial login", "output": "Submit Form", "sla": "Instan" },
					"roles": {
						"uiux": { "assignee": "Rian", "status": "planned", "screen": "Login Form" },
						"frontend": { "assignee": "Siti", "status": "planned", "component": "LoginForm.tsx", "route": "/login" },
						"backend": { "assignee": "Budi", "status": "planned", "method": "POST", "endpoint": "/api/auth/login", "statusCode": "200" }
					}
				},
				{
					"id": "node_otp_2",
					"type": "process",
					"label": "Validasi Database",
					"x": 380,
					"y": 200,
					"doc": { "actor": "Sistem", "input": "Kredensial login", "process": "Mengecek kecocokan email dan hash password di DB", "output": "Status Valid", "sla": "1 detik" },
					"roles": {
						"uiux": { "assignee": "Rian", "status": "planned", "screen": "Loading overlay" },
						"frontend": { "assignee": "Siti", "status": "planned", "component": "Loading.tsx" },
						"backend": { "assignee": "Budi", "status": "planned", "method": "POST", "endpoint": "/api/auth/verify", "statusCode": "200" }
					}
				},
				{
					"id": "node_otp_3",
					"type": "system",
					"label": "Kirim OTP via WA",
					"x": 660,
					"y": 200,
					"doc": { "actor": "Gateway WA", "input": "Nomor handphone", "process": "Kirim SMS/WA API OTP token 6 digit ke HP terdaftar", "output": "WhatsApp Sent", "sla": "3 detik" },
					"roles": {
						"uiux": { "assignee": "Rian", "status": "planned", "screen": "" },
						"frontend": { "assignee": "Siti", "status": "planned", "component": "" },
						"backend": { "assignee": "Arief", "status": "planned", "method": "POST", "endpoint": "/api/otp/send", "statusCode": "200" }
					}
				},
				{
					"id": "node_otp_4",
					"type": "terminator",
					"label": "Landing Dashboard",
					"x": 940,
					"y": 200,
					"doc": { "actor": "User", "input": "Token OTP", "process": "User ketik OTP, validasi sukses, arahkan ke dasbor", "output": "Landing Dashboard", "sla": "Instan" },
					"roles": {
						"uiux": { "assignee": "Rian", "status": "planned", "screen": "Dashboard" },
						"frontend": { "assignee": "Siti", "status": "planned", "component": "Dashboard.tsx", "route": "/dashboard" },
						"backend": { "assignee": "Budi", "status": "planned", "method": "GET", "endpoint": "/api/dashboard/stats", "statusCode": "200" }
					}
				}
			],
			"edges": [
				{ "id": "edge_otp_1", "from": "node_otp_1", "to": "node_otp_2" },
				{ "id": "edge_otp_2", "from": "node_otp_2", "to": "node_otp_3", "label": "Kredensial Cocok" },
				{ "id": "edge_otp_3", "from": "node_otp_3", "to": "node_otp_4", "label": "OTP Benar" }
			]
		}`
	} else {
		flowJSON = fmt.Sprintf(`{
			"status": "complete",
			"name": "Alur Kerja Proses Bisnis %s",
			"description": "Alur adaptif instan terbuat melalui AI berdasarkan deskripsi fungsional.",
			"nodes": [
				{
					"id": "node_gen_1",
					"type": "terminator",
					"label": "Mulai Proses",
					"x": 100,
					"y": 200,
					"doc": { "actor": "User", "input": "Inisiasi", "process": "Memulai alur bisnis utama", "output": "Alur Terbuka", "sla": "Instan" },
					"roles": {
						"uiux": { "assignee": "Rian", "status": "planned", "screen": "Mulai" },
						"frontend": { "assignee": "Siti", "status": "planned", "component": "StartButton.tsx" },
						"backend": { "assignee": "Budi", "status": "planned", "method": "POST", "endpoint": "/api/start", "statusCode": "200" }
					}
				},
				{
					"id": "node_gen_2",
					"type": "process",
					"label": "Pemberkasan / Pemrosesan",
					"x": 400,
					"y": 200,
					"doc": { "actor": "Sistem", "input": "Dokumen Alur", "process": "Melakukan integrasi internal dan pengecekan manual", "output": "Validasi Lolos", "sla": "10 menit" },
					"roles": {
						"uiux": { "assignee": "Rian", "status": "planned", "screen": "Proses" },
						"frontend": { "assignee": "Siti", "status": "planned", "component": "ProcessPanel.tsx" },
						"backend": { "assignee": "Budi", "status": "planned", "method": "POST", "endpoint": "/api/process", "statusCode": "200" }
					}
				},
				{
					"id": "node_gen_3",
					"type": "terminator",
					"label": "Selesai",
					"x": 700,
					"y": 200,
					"doc": { "actor": "Sistem", "input": "Berita Acara", "process": "Selesai dan sinkronkan log status ke database", "output": "Alur Ditutup", "sla": "Instan" },
					"roles": {
						"uiux": { "assignee": "Rian", "status": "planned", "screen": "Selesai" },
						"frontend": { "assignee": "Siti", "status": "planned", "component": "SuccessAlert.tsx" },
						"backend": { "assignee": "Budi", "status": "planned", "method": "POST", "endpoint": "/api/finish", "statusCode": "200" }
					}
				}
			],
			"edges": [
				{ "id": "edge_gen_1", "from": "node_gen_1", "to": "node_gen_2" },
				{ "id": "edge_gen_2", "from": "node_gen_2", "to": "node_gen_3" }
			]
		}`, func() string {
			if len(prompt) > 25 {
				return prompt[:25]
			}
			return prompt
		}())
	}
	return []byte(flowJSON)
}
