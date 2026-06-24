import { Module } from '../domain/types';
import { useStore } from '../store/useStore';

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
  const jsonString = JSON.stringify(module, null, 2);
  downloadFile(jsonString, `${module.name.toLowerCase().replace(/\s+/g, '_')}_canonical.json`, 'application/json');
}

/**
 * Export module to polished readable Markdown (PDF printable text)
 */
export function generateMarkdown(module: Module): string {
  let md = `# Alur Kerja: ${module.name}\n\n`;
  if (module.description) {
    md += `> ${module.description}\n\n`;
  }

  md += `## 1. Ringkasan Langkah Bisnis\n\n`;
  module.nodes.forEach((node, index) => {
    md += `### ${index + 1}. [${node.type.toUpperCase()}] ${node.label}\n\n`;
    if (node.doc.actor) md += `- **Aktor Utama**: ${node.doc.actor}\n`;
    if (node.doc.system) md += `- **Sistem Utama**: ${node.doc.system}\n`;
    if (node.doc.sla) md += `- **SLA Estimasi**: ${node.doc.sla}\n`;
    if (node.doc.process) md += `- **Deskripsi Proses**: ${node.doc.process}\n`;
    if (node.doc.rules) md += `- **Aturan Bisnis**: ${node.doc.rules}\n`;
    if (node.doc.input) md += `- **Input**: ${node.doc.input}\n`;
    if (node.doc.output) md += `- **Output**: ${node.doc.output}\n`;
    md += `\n`;

    // Facets Kesiapan
    md += `#### Peran & Kesiapan Implementasi\n\n`;
    if (node.roles.uiux) {
      md += `- **UI/UX**: Assigned to **${getDisplayAssignee(node.roles.uiux.assignee)}** | Status: \`${node.roles.uiux.status?.toUpperCase() || 'PLANNED'}\` | Screen: *${node.roles.uiux.screen || 'N/A'}*\n`;
    }
    if (node.roles.frontend) {
      md += `- **Frontend**: Assigned to **${getDisplayAssignee(node.roles.frontend.assignee)}** | Status: \`${node.roles.frontend.status?.toUpperCase() || 'PLANNED'}\` | Component: \`${node.roles.frontend.component || 'N/A'}\` | Route: \`${node.roles.frontend.route || 'N/A'}\`\n`;
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
        description: `Implementasi proses bisnis untuk langkah: "${node.label}". [Aktor: ${node.doc.actor || 'N/A'}] [Aturan Bisnis: ${node.doc.rules || 'N/A'}]`,
        responses: {},
      };

      const code = be.statusCode || '200';
      operation.responses[code] = {
        description: `Respon contoh kesuksesan (Status ${code})`,
      };

      if (be.response) {
        try {
          const parsedRes = JSON.parse(be.response);
          operation.responses[code].content = {
            'application/json': {
              example: parsedRes,
            },
          };
        } catch (e) {
          operation.responses[code].content = {
            'application/json': {
              example: { rawResponse: be.response },
            },
          };
        }
      }

      if (method !== 'get' && be.request) {
        try {
          const parsedReq = JSON.parse(be.request);
          operation.requestBody = {
            content: {
              'application/json': {
                schema: {
                  type: 'object',
                },
                example: parsedReq,
              },
            },
          };
        } catch (e) {
          operation.requestBody = {
            content: {
              'application/json': {
                example: be.request,
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

export function exportToOpenApi(module: Module) {
  const content = generateOpenApi(module);
  downloadFile(content, `${module.name.toLowerCase().replace(/\s+/g, '_')}_openapi.json`, 'application/json');
}

/**
 * Export module to Excel-suitable CSV format
 */
export function exportToCsv(module: Module) {
  // Column definitions for the Excel export
  const headers = [
    'ID Langkah',
    'Label Langkah',
    'Tipe Langkah',
    'Aktor',
    'Sistem',
    'Estimasi SLA',
    'Deskripsi Proses Bisnis',
    'Aturan Bisnis',
    'UX Assignee',
    'UX Status',
    'UX Laman/Layar',
    'UX Figma Link',
    'FE Assignee',
    'FE Status',
    'FE Komponen',
    'FE Jalur Rute',
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
      node.doc.system || '',
      node.doc.sla || '',
      (node.doc.process || '').replace(/"/g, '""'), // escape quotes in CSV
      (node.doc.rules || '').replace(/"/g, '""'),
      node.roles.uiux?.assignee && isRegistered(node.roles.uiux.assignee) ? node.roles.uiux.assignee : '',
      node.roles.uiux?.status || '',
      node.roles.uiux?.screen || '',
      node.roles.uiux?.link || '',
      node.roles.frontend?.assignee && isRegistered(node.roles.frontend.assignee) ? node.roles.frontend.assignee : '',
      node.roles.frontend?.status || '',
      node.roles.frontend?.component || '',
      node.roles.frontend?.route || '',
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
