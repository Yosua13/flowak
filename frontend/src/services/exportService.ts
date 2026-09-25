import { Module } from '../domain/types';
import { useStore } from '../store/useStore';
import type { DerivedViewData } from './derivedViewData';

const getDisplayAssignee = (name?: string): string => {
  if (!name) return 'Belum ditunjuk';
  const { teamMembers } = useStore.getState();
  return teamMembers.some((m) => m.name === name) ? name : 'Belum ditunjuk';
};

const isRegistered = (name?: string): boolean => {
  if (!name) return false;
  const { teamMembers } = useStore.getState();
  return teamMembers.some((m) => m.name === name);
};

const formatBusinessRules = (rules: Module['nodes'][number]['doc']['rules']): string =>
  (rules || []).map((rule) => rule.description).filter(Boolean).join('; ');

/**
 * Utility to download files in the browser
 */
function downloadFile(content: string, filename: string, contentType: string) {
  const blob = new Blob([content], { type: contentType });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}

/**
 * Export module to raw canonical JSON
 */
export function exportToJson(module: Module) {
  const jsonString = JSON.stringify({ schemaVersion: module.schemaVersion, exportedAt: new Date().toISOString(), module: redactSensitive(module) }, null, 2);
  downloadFile(jsonString, `${module.name.toLowerCase().replace(/\s+/g, '_')}_canonical.json`, 'application/json');
}

/**
 * Export module to polished readable Markdown (PDF printable text)
 */
export function generateMarkdown(module: Module, derived?: DerivedViewData | null): string {
  module = redactSensitive(module) as Module;
  let md = `# Alur Kerja: ${module.name}\n\n`;
  if (module.description) {
    md += `> ${module.description}\n\n`;
  }
  const baseline = derived?.baselines[0];
  md += `- **Schema version**: ${module.schemaVersion}\n`;
  md += baseline ? `- **Baseline modul**: v${baseline.version} (${new Date(baseline.created_at).toISOString()})\n\n` : `- **Baseline modul**: belum dipublikasikan\n\n`;

  md += `## 1. Ringkasan Langkah Bisnis\n\n`;
  module.nodes.forEach((node, index) => {
    md += `### ${index + 1}. [${node.type.toUpperCase()}] ${node.label}\n\n`;
    if (node.doc.actor) md += `- **Aktor Utama**: ${node.doc.actor}\n`;
    if (node.doc.trigger) md += `- **Pemicu**: ${node.doc.trigger}\n`;
    if (node.doc.system) md += `- **Sistem Utama**: ${node.doc.system}\n`;
    if (node.doc.sla) md += `- **SLA Estimasi**: ${node.doc.sla}\n`;
    if (node.doc.priority || node.doc.riskLevel) md += `- **Prioritas/Risiko**: ${node.doc.priority || 'medium'} / ${node.doc.riskLevel || 'medium'}\n`;
    if (node.doc.process) md += `- **Deskripsi Proses**: ${node.doc.process}\n`;
    const businessRules = formatBusinessRules(node.doc.rules);
    if (businessRules) md += `- **Aturan Bisnis**: ${businessRules}\n`;
    if (node.doc.exceptionPath) md += `- **Alur Pengecualian**: ${node.doc.exceptionPath}\n`;
    if (node.doc.acceptanceCriteria) md += `- **Kriteria Selesai**: ${node.doc.acceptanceCriteria}\n`;
    if (node.doc.input) md += `- **Input**: ${node.doc.input}\n`;
    if (node.doc.output) md += `- **Output**: ${node.doc.output}\n`;
    md += `\n`;

    // Facets Kesiapan
    md += `#### Peran & Kesiapan Implementasi\n\n`;
    if (node.roles.uiux) {
      md += `- **UI/UX**: Assigned to **${getDisplayAssignee(node.roles.uiux.assignee)}** | Status: \`${node.roles.uiux.status?.toUpperCase() || 'PLANNED'}\` | Screen: *${node.roles.uiux.screen || 'N/A'}*\n`;
    }
    if (node.roles.frontend) {
      const legacyFrontend = node.roles.frontend as typeof node.roles.frontend & { component?: string };
      const pageName = node.roles.frontend.page || legacyFrontend.component || 'N/A';
      md += `- **Frontend Handoff**: Assigned to **${getDisplayAssignee(node.roles.frontend.assignee)}** | Status: \`${node.roles.frontend.status?.toUpperCase() || 'PLANNED'}\` | Page: \`${pageName}\` | Route: \`${node.roles.frontend.route || 'N/A'}\`\n`;
    }
    if (node.roles.backend) {
      md += `- **Backend**: Assigned to **${getDisplayAssignee(node.roles.backend.assignee)}** | Status: \`${node.roles.backend.status?.toUpperCase() || 'PLANNED'}\` | API Contract: \`${node.roles.backend.method || 'GET'} ${node.roles.backend.endpoint || 'N/A'}\` | Status code: \`${node.roles.backend.statusCode || 'N/A'}\`\n`;
    }
    md += `\n---\n\n`;
  });

  md += `## 2. Alur Koneksi (Urutan Hubungan)\n\n`;
  if (module.edges.length === 0) {
    md += `*Belum ada koneksi antar-langkah diatur.*\n`;
  } else {
    module.edges.forEach((edge) => {
      const fromNode = module.nodes.find((n) => n.id === edge.from)?.label || edge.from;
      const toNode = module.nodes.find((n) => n.id === edge.to)?.label || edge.to;
      const labelStr = edge.label ? ` --[ "${edge.label}" ]--> ` : ' ----> ';
      md += `- \`${fromNode}\`${labelStr}\`${toNode}\`\n`;
    });
  }

  md += `\n## 3. Ringkasan Work Item\n\n`;
  if (!derived?.work_items.length) md += `*Belum ada work item pada modul ini.*\n`;
  else derived.work_items.forEach((item) => { md += `- **${item.key}** [${item.status}] ${item.title}${item.due_date ? ` (due ${item.due_date})` : ''}\n`; });
  if (derived?.comments.length) {
    md += `\n## 4. Keputusan dan Komentar\n\n`;
    derived.comments.forEach((comment) => { md += `- ${comment.body}\n`; });
  }

  md += `\n\n*Dokumen dicetak otomatis via Flowak Workspace pd ${new Date().toLocaleDateString('id-ID')}*\n`;
  return md;
}

export function exportToMarkdown(module: Module) {
  const content = generateMarkdown(module);
  downloadFile(content, `${module.name.toLowerCase().replace(/\s+/g, '_')}_workflow.md`, 'text/markdown');
}

/**
 * Export module Backend facets to OpenAPI 3.1.0 contract
 */
export function generateOpenApi(module: Module): string {
  module = redactSensitive(module) as Module;
  const openapi: any = {
    openapi: '3.1.0',
    info: {
      title: `API Contract - ${module.name}`,
      description: module.description || `Generated API contract from Flowak nodes.`,
      version: '1.0.0',
    },
    paths: {},
    components: {
      securitySchemes: {
        BearerAuth: {
          type: 'http',
          scheme: 'bearer',
          bearerFormat: 'JWT',
        },
      },
    },
  };

  module.nodes.forEach((node) => {
    const be = node.roles.backend;
    if (be && be.endpoint) {
      const path = be.endpoint.split('?')[0]; // strip query params
      const method = (be.method || 'GET').toLowerCase();

      if (!openapi.paths[path]) {
        openapi.paths[path] = {};
      }

      const operation: any = {
        summary: node.label,
        description: `Implementasi proses bisnis untuk langkah: "${node.label}". [Aktor: ${node.doc.actor || 'N/A'}] [Aturan Bisnis: ${formatBusinessRules(node.doc.rules) || 'N/A'}]`,
        responses: {},
      };

      const code = be.statusCode || '200';
      operation.responses[code] = {
        description: `Respon contoh kesuksesan (Status ${code})`,
      };

      // Response examples can contain production data; a contract exports its typed shape only.
      operation.responses[code].content = { 'application/json': { schema: { type: 'object' } } };

      if (method !== 'get' && be.request) {
        try {
          const parsedReq = JSON.parse(be.request);
          operation.requestBody = {
            content: {
              'application/json': {
                schema: {
                  type: 'object',
                },
                example: redactSensitive(parsedReq),
              },
            },
          };
        } catch (e) {
          operation.requestBody = {
            content: {
              'application/json': {
                example: '[redacted non-JSON request example]',
              },
            },
          };
        }
      }

      if (be.auth) {
        operation.security = [{ BearerAuth: [] }];
      }

      openapi.paths[path][method] = operation;
    }
  });

  return JSON.stringify(openapi, null, 2);
}

const sensitiveKey = /authorization|cookie|password|secret|token|api[_-]?key/i;
export function redactSensitive(value: unknown): unknown {
  if (Array.isArray(value)) return value.map(redactSensitive);
  if (value && typeof value === 'object') return Object.fromEntries(Object.entries(value).map(([key, item]) => {
    if (sensitiveKey.test(key) || key === 'curl') return [key, '{{REDACTED}}'];
    if (key === 'auth') return [key, typeof item === 'string' && (/^\{\{[A-Z][A-Z0-9_]*\}\}$/.test(item) || item === 'none' || item === 'inherit') ? item : '{{API_TOKEN}}'];
    if (key === 'endpoint' && typeof item === 'string') return [key, item.split('?')[0].split('#')[0]];
    return [key, redactSensitive(item)];
  }));
  if (typeof value === 'string') {
    const trimmed = value.trim();
    if (trimmed.startsWith('{') || trimmed.startsWith('[')) {
      try { return JSON.stringify(redactSensitive(JSON.parse(trimmed))); } catch { /* retain non-JSON text after redaction */ }
    }
    return value.replace(/Bearer\s+[^\s"']+/gi, 'Bearer {{REDACTED}}');
  }
  return value;
}

export function generateRedactedCurl(module: Module): string {
  return module.nodes.flatMap((node) => {
    const contract = node.roles.backend;
    if (!contract?.endpoint) return [];
    const method = contract.method || 'GET';
    const auth = contract.auth ? " -H 'Authorization: Bearer {{API_TOKEN}}'" : '';
    const body = method === 'GET' || !contract.request ? '' : " -H 'Content-Type: application/json' --data '{{REQUEST_BODY}}'";
    return [`# ${node.label}\ncurl -X ${method} '{{BASE_URL}}${contract.endpoint.split('?')[0]}'${auth}${body}`];
  }).join('\n\n');
}

export function exportToCurl(module: Module) {
  downloadFile(generateRedactedCurl(module), `${module.name.toLowerCase().replace(/\s+/g, '_')}_requests.sh`, 'text/plain');
}

export function exportToOpenApi(module: Module) {
  const content = generateOpenApi(module);
  downloadFile(content, `${module.name.toLowerCase().replace(/\s+/g, '_')}_openapi.json`, 'application/json');
}

/**
 * Export module to Excel-suitable CSV format
 */
export function exportToCsv(module: Module) {
  module = redactSensitive(module) as Module;
  // Column definitions for the Excel export
  const headers = [
    'ID Langkah',
    'Label Langkah',
    'Tipe Langkah',
    'Aktor',
    'Pemicu',
    'Sistem',
    'Estimasi SLA',
    'Prioritas',
    'Risiko',
    'Deskripsi Proses Bisnis',
    'Aturan Bisnis',
    'Alur Pengecualian',
    'Kriteria Selesai',
    'UX Assignee',
    'UX Status',
    'UX Laman/Layar',
    'UX Figma Link',
    'FE Assignee',
    'FE Status',
    'FE Halaman',
    'FE Jalur Rute',
    'FE Interaksi',
    'FE Validasi',
    'FE State Handling',
    'FE Target Selesai',
    'BE Assignee',
    'BE Status',
    'BE HTTP Metode',
    'BE Endpoint',
    'BE Status Code',
  ];

  const rows = module.nodes.map((node) => {
    return [
      node.id,
      node.label,
      node.type,
      node.doc.actor || '',
      node.doc.trigger || '',
      node.doc.system || '',
      node.doc.sla || '',
      node.doc.priority || '',
      node.doc.riskLevel || '',
      (node.doc.process || '').replace(/"/g, '""'), // escape quotes in CSV
      (node.doc.rules || []).map((rule) => rule.description).join('; ').replace(/"/g, '""'),
      (node.doc.exceptionPath || '').replace(/"/g, '""'),
      (node.doc.acceptanceCriteria || '').replace(/"/g, '""'),
      node.roles.uiux?.assignee && isRegistered(node.roles.uiux.assignee) ? node.roles.uiux.assignee : '',
      node.roles.uiux?.status || '',
      node.roles.uiux?.screen || '',
      node.roles.uiux?.link || '',
      node.roles.frontend?.assignee && isRegistered(node.roles.frontend.assignee) ? node.roles.frontend.assignee : '',
      node.roles.frontend?.status || '',
      node.roles.frontend?.page || (node.roles.frontend as typeof node.roles.frontend & { component?: string } | undefined)?.component || '',
      node.roles.frontend?.route || '',
      node.roles.frontend?.interaction || '',
      node.roles.frontend?.validation || '',
      node.roles.frontend?.state || '',
      node.roles.frontend?.dueDate || '',
      node.roles.backend?.assignee && isRegistered(node.roles.backend.assignee) ? node.roles.backend.assignee : '',
      node.roles.backend?.status || '',
      node.roles.backend?.method || '',
      node.roles.backend?.endpoint || '',
      node.roles.backend?.statusCode || '',
    ];
  });

  const csvContent = [
    headers.join(','),
    ...rows.map((row) => row.map((val) => `"${val.replace(/[\r\n]+/g, ' ')}"`).join(',')),
  ].join('\n');

  downloadFile('\uFEFF' + csvContent, `${module.name.toLowerCase().replace(/\s+/g, '_')}_report.csv`, 'text/csv;charset=utf-8;');
}
