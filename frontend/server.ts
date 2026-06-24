/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import express from "express";
import path from "path";
import { createServer as createViteServer } from "vite";
import { GoogleGenAI, Type } from "@google/genai";
import dotenv from "dotenv";

dotenv.config();

// Custom UID generator for backend matching
function uid() {
  return Math.random().toString(36).substring(2, 9);
}

async function startServer() {
  const app = express();
  const PORT = Number(process.env.FRONTEND_SERVER_PORT || 3002);

  // Use raw and JSON middleware
  app.use(express.json({ limit: "10mb" }));

  // Initialize Google GenAI
  const apiKey = process.env.GEMINI_API_KEY;
  let ai: GoogleGenAI | null = null;
  if (apiKey) {
    ai = new GoogleGenAI({
      apiKey,
      httpOptions: {
        headers: {
          "User-Agent": "aistudio-build",
        },
      },
    });
  }

  // --- API ENDPOINTS ---

  // 1. Generate flow with AI
  app.post("/api/ai/generate-flow", async (req, res) => {
    try {
      const { prompt } = req.body;
      if (!prompt) {
        return res.status(400).json({ error: "Permintaan prompt tidak boleh kosong" });
      }

      if (!ai) {
        return res.status(200).json({
          status: "mock_demo",
          error: "GEMINI_API_KEY tidak terkonfigurasi. Menggunakan data demo offline.",
          modules: getDemoFlow(prompt),
        });
      }

      const systemInstruction = `Anda adalah ahli pemodelan proses bisnis dan arsitek sistem teknis Flowak.
Tugas Anda mendesain alur kerja (workflow) lengkap dalam bentuk graf Node (langkah proses) dan Edge (sambungan/aliran) berdasarkan deskripsi pengguna.

Aturan Penting:
1. Buat 4 hingga 8 langkah (nodes) yang logis, berurutan, dan menyertakan aspek bisnis (actor, input, process, output, sla) serta aspek teknis untuk UI/UX, Frontend, dan Backend.
2. Tempatkan koordinat x dan y (grid kelipatan 40px) sedemikian rupa agar membentuk diagram alir yang indah dan tidak tumpang tindih. (Misal: Node pertama di x: 80, y: 150. Node berikutnya bergeser ke kanan x: 380, y: 150, dst. Jika ada percabangan, Anda bisa menyebarkan y).
3. Sambungkan node dengan Edge yang merujuk pada id node yang sesuai.
4. Pastikan response strictly valid JSON sesuai tipe skema yang diberikan.`;

      const promptTemplate = `Deskripsi Kebutuhan: "${prompt}"

Silakan buatkan struktur node proses bisnis dan edges sesuai dengan deskripsi tersebut.`;

      const responseSchema = {
        type: Type.OBJECT,
        properties: {
          name: {
            type: Type.STRING,
            description: "Nama modul alur kerja yang representatif dan profesional",
          },
          description: {
            type: Type.STRING,
            description: "Deskripsi singkat mengenai alur kerja fungsional ini",
          },
          nodes: {
            type: Type.ARRAY,
            description: "Daftar node dalam diagram alir",
            items: {
              type: Type.OBJECT,
              properties: {
                tempId: {
                  type: Type.STRING,
                  description: "ID sementara node (contoh: 'node_1', 'node_2')",
                },
                type: {
                  type: Type.STRING,
                  description: "Tipe node: 'terminator' | 'process' | 'decision' | 'actor' | 'system'",
                },
                label: {
                  type: Type.STRING,
                  description: "Label singkat representatif langkah alur (contoh: 'Validasi OTP')",
                },
                x: {
                  type: Type.NUMBER,
                  description: "Koordinat horizontal di kanvas (kisaran 100 s.d 1400)",
                },
                y: {
                  type: Type.NUMBER,
                  description: "Koordinat vertikal di kanvas (kisaran 100 s.d 800)",
                },
                doc: {
                  type: Type.OBJECT,
                  description: "Aspek analisis fungsional bisnis",
                  properties: {
                    actor: { type: Type.STRING, description: "Aktor utama (contoh: 'Pelanggan', 'Sistem', 'Admin')" },
                    input: { type: Type.STRING, description: "Input data proses" },
                    process: { type: Type.STRING, description: "Uraian logika proses bisnis" },
                    output: { type: Type.STRING, description: "Output pasca proses selesai" },
                    rules: { type: Type.STRING, description: "Kebijakan/Aturan validasi proses" },
                    system: { type: Type.STRING, description: "Subsistem/Aplikasi terkait" },
                    sla: { type: Type.STRING, description: "Perkiraan Service Level Agreement SLA (contoh: '5 detik', '1 menit')" },
                  },
                },
                roles: {
                  type: Type.OBJECT,
                  description: "Aspek korelasi per peran spesifik",
                  properties: {
                    uiux: {
                      type: Type.OBJECT,
                      properties: {
                        assignee: { type: Type.STRING, description: "Rian (UI/UX Designer)" },
                        status: { type: Type.STRING, description: "Status: 'planned'" },
                        screen: { type: Type.STRING, description: "Nama rancangan layar figma/mockup" },
                      },
                    },
                    frontend: {
                      type: Type.OBJECT,
                      properties: {
                        assignee: { type: Type.STRING, description: "Siti (FE Engineer)" },
                        status: { type: Type.STRING, description: "Status: 'planned'" },
                        component: { type: Type.STRING, description: "Nama komponen file (misal: OtpForm.tsx)" },
                        route: { type: Type.STRING, description: "Jalur URL rute (misal: /auth/otp)" },
                        framework: { type: Type.STRING, description: "React + Tailwind" },
                      },
                    },
                    backend: {
                      type: Type.OBJECT,
                      properties: {
                        assignee: { type: Type.STRING, description: "Budi (BE Engineer)" },
                        status: { type: Type.STRING, description: "Status: 'planned'" },
                        method: { type: Type.STRING, description: "HTTP Method: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'" },
                        endpoint: { type: Type.STRING, description: "Jalur API: (misal: /api/v1/auth/verify)" },
                        statusCode: { type: Type.STRING, description: "200" },
                      },
                    },
                  },
                },
              },
              required: ["tempId", "type", "label", "x", "y"],
            },
          },
          edges: {
            type: Type.ARRAY,
            description: "Daftar sambungan atau garis penghubung antar node alur kerja",
            items: {
              type: Type.OBJECT,
              properties: {
                from: { type: Type.STRING, description: "tempId node asal (terdefinisi di nodes)" },
                to: { type: Type.STRING, description: "tempId node tujuan (terdefinisi di nodes)" },
                label: { type: Type.STRING, description: "Label sambungan opsional (contoh: 'Sukses', 'Gagal')" },
              },
              required: ["from", "to"],
            },
          },
        },
        required: ["name", "description", "nodes", "edges"],
      };

      const response = await ai.models.generateContent({
        model: "gemini-3.5-flash",
        contents: promptTemplate,
        config: {
          systemInstruction,
          responseMimeType: "application/json",
          responseSchema,
        },
      });

      const parsedData = JSON.parse(response.text || "{}");

      // Map tempIds from AI output to real app-safe IDs
      const idMap: Record<string, string> = {};
      const finalNodes = (parsedData.nodes || []).map((n: any) => {
        const generatedId = `node_${uid()}`;
        idMap[n.tempId || n.id] = generatedId;

        // Assure healthy template structure
        return {
          id: generatedId,
          type: n.type || "process",
          label: n.label || "Langkah Alur",
          x: n.x || 150,
          y: n.y || 150,
          doc: {
            actor: n.doc?.actor || "Sistem",
            input: n.doc?.input || "",
            process: n.doc?.process || "",
            output: n.doc?.output || "",
            rules: n.doc?.rules || "",
            system: n.doc?.system || "Aplikasi Portal",
            sla: n.doc?.sla || "Instan",
          },
          roles: {
            uiux: {
              assignee: n.roles?.uiux?.assignee || "Rian",
              status: n.roles?.uiux?.status || "planned",
              screen: n.roles?.uiux?.screen || "",
              link: "",
            },
            frontend: {
              assignee: n.roles?.frontend?.assignee || "Siti",
              status: n.roles?.frontend?.status || "planned",
              component: n.roles?.frontend?.component || "",
              route: n.roles?.frontend?.route || "",
              framework: n.roles?.frontend?.framework || "React + Tailwind CSS",
              link: "",
            },
            backend: {
              assignee: n.roles?.backend?.assignee || "Budi",
              status: n.roles?.backend?.status || "planned",
              method: n.roles?.backend?.method || "POST",
              endpoint: n.roles?.backend?.endpoint || "",
              auth: "",
              request: "",
              response: "",
              statusCode: n.roles?.backend?.statusCode || "200",
            },
          },
        };
      });

      const finalEdges = (parsedData.edges || [])
        .map((e: any) => {
          const fromId = idMap[e.from];
          const toId = idMap[e.to];
          if (!fromId || !toId) return null;
          return {
            id: `edge_${uid()}`,
            from: fromId,
            to: toId,
            label: e.label || "",
          };
        })
        .filter(Boolean);

      return res.json({
        status: "complete",
        name: parsedData.name || "Modul Alur AI",
        description: parsedData.description || "Alur kerja terbuat melalui asisten AI",
        nodes: finalNodes,
        edges: finalEdges,
      });
    } catch (error: any) {
      console.error("AI Generate Flow Error", error);
      res.status(500).json({ error: error.message });
    }
  });

  // 2. Generate AI Mock Payload for Backend tab
  app.post("/api/ai/mock-payload", async (req, res) => {
    try {
      const { actor, label, endpoint, process, method } = req.body;
      if (!ai) {
        return res.json({
          request: JSON.stringify({ deviceId: "MOCK_DEVICE_ID", channel: "Web" }, null, 2),
          response: JSON.stringify({ success: true, timestamp: new Date().toISOString(), data: { step: label, simulated: true } }, null, 2),
        });
      }

      const prompt = `Definisikan payload JSON Request Body dan JSON Response Body fungsional, realistis untuk API Endpoint berikut:
- Tipe Langkah / Proses Bisnis: ${label}
- Aktor Utama: ${actor}
- Jalur Endpoint: ${endpoint}
- Metoda HTTP: ${method}
- Deskripsi Alur: ${process}

Keluarkan data dalam format skema JSON yang berisi dua properti string: 'request' (berisi teks payload JSON request) dan 'response' (berisi teks payload JSON response).`;

      const responseSchema = {
        type: Type.OBJECT,
        properties: {
          request: { type: Type.STRING, description: "JSON string representatif body request, dikosongkan jika GET" },
          response: { type: Type.STRING, description: "JSON string representatif body respon dari server" },
        },
        required: ["request", "response"],
      };

      const response = await ai.models.generateContent({
        model: "gemini-3.5-flash",
        contents: prompt,
        config: {
          responseMimeType: "application/json",
          responseSchema,
        },
      });

      const parsedJSON = JSON.parse(response.text || "{}");
      return res.json(parsedJSON);
    } catch (error: any) {
      console.error("Mock Payload Error", error);
      res.json({ error: error.message });
    }
  });

  // 3. AI Spec Reviewer & Auditor
  app.post("/api/ai/audit-flow", async (req, res) => {
    try {
      const { module } = req.body;
      if (!module) return res.status(400).json({ error: "Modul harus disertakan." });

      if (!ai) {
        return res.json({
          score: 85,
          summary: "Audit offline demo dilakukan. Struktur modul valid namun integrasi AI sesungguhnya tidak aktif karena API KEY kosong.",
          issues: [
            {
              severity: "warning",
              message: "Rancang deskripsi alur kerja yang mendalam untuk meningkatkan akurasi koding.",
              nodeName: "Evaluasi Awal",
              type: "Dokumentasi",
            },
          ],
        });
      }

      const prompt = `Analisis kualitas rancangan spesifikasi fungsional modul alur kerja berikut ini.
Nama Modul: ${module.name}
Deskripsi Modul: ${module.description || "N/A"}

Daftar Node:
${JSON.stringify(
        (module.nodes || []).map((n: any) => ({
          label: n.label,
          type: n.type,
          actor: n.doc?.actor,
          process: n.doc?.process,
          sla: n.doc?.sla,
          endpoint: n.roles?.backend?.endpoint,
          auth: n.roles?.backend?.auth,
        })),
        null,
        2
      )}

Daftar Sambungan (Edges):
${JSON.stringify(module.edges || [])}

Deteksi potensi kecacatan perancangan (desain flaw) seperti:
- SLA Bottleneck: Langkah manual atau terlampau lambat.
- Security Flaw: Endpoint backend sensitif yang tidak memiliki otorisasi token.
- Dead End: Langkah alur yang tidak terhubung ke node akhir mana pun terputus.
- Missing Info: Node yang aspek proses bisnisnya kosong.

Berikan keluaran hasil audit Anda dalam format skema JSON terstruktur berisi:
1. 'score' (angka integer 0-100 penilaian kelayakan rancangan)
2. 'summary' (paragraf ringkasan naratif tinjauan kualitatif)
3. 'issues' (array berisi daftar masalah/saran perbaikan)`;

      const responseSchema = {
        type: Type.OBJECT,
        properties: {
          score: { type: Type.INTEGER, description: "Skor audit rancangan antara 0 - 100" },
          summary: { type: Type.STRING, description: "Ringkasan deskriptif analitis tingkat tinggi perihal rancangan alur ini" },
          issues: {
            type: Type.ARRAY,
            items: {
              type: Type.OBJECT,
              properties: {
                severity: { type: Type.STRING, description: "Tingkat keparahan: 'error' (Kritikal) | 'warning' (Peringatan) | 'info' (Catatan)" },
                message: { type: Type.STRING, description: "Rincian detail temuan audit serta rekomendasi solusi pemulihan" },
                nodeName: { type: Type.STRING, description: "Nama langkah terkait jika temuan spesifik ke suatu node (opsional)" },
                type: { type: Type.STRING, description: "Kategori masalah: 'SLA' | 'Keamanan' | 'Logika Alur' | 'Dokumentasi'" },
              },
              required: ["severity", "message", "type"],
            },
          },
        },
        required: ["score", "summary", "issues"],
      };

      const response = await ai.models.generateContent({
        model: "gemini-3.5-flash",
        contents: prompt,
        config: {
          responseMimeType: "application/json",
          responseSchema,
        },
      });

      const auditResult = JSON.parse(response.text || "{}");
      return res.json(auditResult);
    } catch (error: any) {
      console.error("AI Audit Flow Error", error);
      res.status(500).json({ error: error.message });
    }
  });

  // 4. AI Boilerplate Code Generator
  app.post("/api/ai/generate-code", async (req, res) => {
    try {
      const { type, nodeLabel, nodeType, detail, framework } = req.body;
      if (!ai) {
        return res.json({
          code: `// Offline demo mockup code\nexport default function ${nodeLabel.replace(/\s+/g, "")}Mock() {\n  return <div>Offline Demo</div>;\n}`,
          language: "typescript",
        });
      }

      let systemInstruction = "";
      let prompt = "";

      if (type === "frontend") {
        systemInstruction = "Anda adalah pengembang Frontend kelas dunia ahli React, Typescript, dan Tailwind CSS.";
        prompt = `Tulis kode boilerplate komponen React yang fungsional, bersih, interaktif, dan terstyling indah menggunakan Tailwind CSS.
Langkah Proses: "${nodeLabel}" (Tipe: ${nodeType})
Nama Komponen: "${detail.component || nodeLabel.replace(/\s+/g, "") + ".tsx"}"
Rute Link Klien: "${detail.route || "/"}"
Target Layout/Kerangka: "${framework || "React + Tailwind"}"

Desainlah komponen ini agar menyertakan visualisasi form input, status loading, dan tombol aksi interaktif yang relevan dengan proses bisnis di atas secara simulatif dan modern.`;
      } else {
        systemInstruction = "Anda adalah pengembang API Backend kelas dunia ahli Node.js Express.js dan TypeScript.";
        prompt = `Tulis kode boilerplate file controller / endpoint Express.js TypeScript yang bersih, aman, dan modular.
Langkah Proses: "${nodeLabel}" (Tipe: ${nodeType})
Metoda HTTP: "${detail.method || "POST"}"
API Endpoint: "${detail.endpoint || "/api/route"}"
StatusCode Terkait: "${detail.statusCode || "200"}"

Gunakan middleware otorisasi jika kolom auth didefinisikan: "${detail.auth || ""}".
Sertakan penanganan parameter body request:
\`\`\`json
${detail.request || "{}"}
\`\`\`
Dan kembalikan respon sukses berupa JSON matching dengan skema respon berikut:
\`\`\`json
${detail.response || "{}"}
\`\`\`
Sertakan parsing, handling validasi sederhana, dan try-catch error response block.`;
      }

      const response = await ai.models.generateContent({
        model: "gemini-3.5-flash",
        contents: prompt,
        config: {
          systemInstruction,
        },
      });

      return res.json({
        code: response.text || `// Gagal menghasilkan kode`,
        language: type === "frontend" ? "tsx" : "typescript",
      });
    } catch (error: any) {
      console.error("AI Code Generator Error", error);
      res.status(500).json({ error: error.message });
    }
  });

  // Serve static files and mount Vite middleware
  if (process.env.NODE_ENV !== "production") {
    const vite = await createViteServer({
      server: { middlewareMode: true },
      appType: "spa",
    });
    app.use(vite.middlewares);
  } else {
    const distPath = path.join(process.cwd(), "dist");
    app.use(express.static(distPath));
    app.get("*", (req, res) => {
      res.sendFile(path.join(distPath, "index.html"));
    });
  }

  app.listen(PORT, "0.0.0.0", () => {
    console.log(`Server running on port ${PORT}`);
  });
}

// Simple Offline demo fallback generator
function getDemoFlow(prompt: string) {
  const desc = prompt.toLowerCase();
  if (desc.includes("otp") || desc.includes("login") || desc.includes("auth")) {
    return {
      name: "Sistem Otentikasi & Verifikasi OTP WA",
      description: "Alur terpadu autentikasi user beserta lapisan pencegahan fraud via kode OTP WhatsApp.",
      nodes: [
        {
          tempId: "n_1",
          type: "terminator",
          label: "Halaman Login Dimulai",
          x: 100,
          y: 200,
          doc: { actor: "User", input: "Email & Sandi", process: "User membuka app dan ketik kredensial login", output: "Submit Form", sla: "Instan" },
        },
        {
          tempId: "n_2",
          type: "process",
          label: "Validasi Database",
          x: 350,
          y: 200,
          doc: { actor: "Sistem", input: "Kredensial login", process: "Mengecek kecocokan email dan hash password di DB", output: "Status Valid", sla: "1 detik" },
        },
        {
          tempId: "n_3",
          type: "system",
          label: "Kirim OTP via WA",
          x: 600,
          y: 250,
          doc: { actor: "Gateway WA", input: "Nomor handphone", process: "Kirim SMS/WA API OTP token 6 digit ke HP terdaftar", output: "WhatsApp Sent", sla: "3 detik" },
        },
        {
          tempId: "n_4",
          type: "terminator",
          label: "Berhasil dashboard",
          x: 850,
          y: 200,
          doc: { actor: "User", input: "Token OTP", process: "User ketik OTP, validasi sukses, arahkan ke dasbor", output: "Landing Dashboard", sla: "Instan" },
        },
      ],
      edges: [
        { from: "n_1", to: "n_2" },
        { from: "n_2", to: "n_3", label: "Kredensial Cocok" },
        { from: "n_3", to: "n_4", label: "OTP Benar" },
      ],
    };
  }

  // General demo flow
  return {
    name: "Alur Kerja Proses Bisnis " + prompt.substring(0, 25),
    description: "Alur adaptif instan terbuat melalui AI berdasarkan deskripsi fungsional.",
    nodes: [
      {
        tempId: "n_1",
        type: "terminator",
        label: "Mulai Proses",
        x: 120,
        y: 220,
        doc: { actor: "User", input: "Inisiasi", process: "Memulai alur bisnis utama", output: "Alur Terbuka", sla: "Instan" },
      },
      {
        tempId: "n_2",
        type: "process",
        label: "Pemberkasan / Pemrosesan",
        x: 380,
        y: 220,
        doc: { actor: "Sistem", input: "Dokumen Alur", process: "Melakukan integrasi internal dan pengecekan manual", output: "Validasi Lolos", sla: "10 menit" },
      },
      {
        tempId: "n_3",
        type: "terminator",
        label: "Selesai",
        x: 640,
        y: 220,
        doc: { actor: "Sistem", input: "Berita Acara", process: "Selesai dan sinkronkan log status ke database", output: "Alur Ditutup", sla: "Instan" },
      },
    ],
    edges: [
      { from: "n_1", to: "n_2" },
      { from: "n_2", to: "n_3" },
    ],
  };
}

startServer();
