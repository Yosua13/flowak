/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from 'react';
import { useStore } from '../../store/useStore';
import { generateMarkdown } from '../../services/exportService';
import { Printer, Copy, Check, FileText } from 'lucide-react';

export default function DocView() {
  const { modules, activeId, teamMembers } = useStore();
  const activeModule = modules.find((m) => m.id === activeId);

  const getDisplayAssignee = (name?: string) => {
    if (!name) return 'Belum ditunjuk';
    return teamMembers.some((m) => m.name === name) ? name : 'Belum ditunjuk';
  };

  const [copied, setCopied] = useState(false);

  if (!activeModule) {
    return (
      <div className="flex-1 p-8 bg-[#0A0A0B] text-gray-500 font-sans text-center flex items-center justify-center">
        Pilih modul terlebih dahulu...
      </div>
    );
  }

  const handleCopyMarkdown = () => {
    const md = generateMarkdown(activeModule);
    navigator.clipboard.writeText(md);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handlePrint = () => {
    window.print(); // Triggers native browser print formatted via CSS @media print
  };

  const nodes = activeModule.nodes || [];

  return (
    <div className="flex-1 bg-[#0A0A0B] p-6 overflow-y-auto scrollbar-thin select-text print:bg-white print:p-0">
      
      {/* Print & Copy Control Ribbon Toolbar */}
      <div className="max-w-3xl mx-auto mb-6 flex items-center justify-between bg-[#131315] border border-white/5 p-3.5 rounded-2xl shadow-xl print:hidden">
        <div className="flex items-center space-x-2">
          <FileText className="w-4 h-4 text-[#C5A267]" />
          <span className="text-xs font-bold text-white uppercase tracking-widest font-sans">Dokumen Spesifikasi Fungsional</span>
        </div>
        <div className="flex space-x-2.5">
          <button
            onClick={handleCopyMarkdown}
            className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg border border-white/10 text-xs font-semibold text-gray-300 bg-white/5 hover:bg-white/10 transition cursor-pointer"
          >
            {copied ? (
              <>
                <Check className="w-3.5 h-3.5 text-emerald-500" />
                <span className="text-emerald-500 font-mono">TERCOPIMD!</span>
              </>
            ) : (
              <>
                <Copy className="w-3.5 h-3.5" />
                <span>SALIN MARKDOWN</span>
              </>
            )}
          </button>
          
          <button
            onClick={handlePrint}
            className="flex items-center space-x-1.5 px-4 py-1.5 rounded-lg bg-[#C5A267] hover:bg-[#B08E56] text-black text-xs font-bold transition shadow-sm cursor-pointer"
          >
            <Printer className="w-3.5 h-3.5 text-black" />
            <span>CETAK / EKSPOR PDF</span>
          </button>
        </div>
      </div>

      {/* Styled Printable Paper Canvas (Serif Typography) */}
      <div className="max-w-3xl mx-auto bg-[#131315] border border-white/5 p-12 rounded-3xl shadow-2xl text-gray-200 font-serif leading-relaxed print:border-none print:shadow-none print:p-0 print:text-black print:bg-white">
        
        {/* Document Header */}
        <div className="border-b border-white/5 pb-6 mb-8 text-center print:text-left print:border-slate-350">
          <p className="text-[10px] font-mono tracking-widest text-[#C5A267] uppercase font-bold mb-1">PROSES BISNIS & TRACEABILITY INDEX</p>
          <h1 className="text-3xl font-black font-sans tracking-tight text-white print:text-black" style={{ fontFamily: 'Georgia, serif' }}>
            {activeModule.name}
          </h1>
          {activeModule.description && (
            <p className="text-sm italic text-gray-400 mt-3 max-w-xl mx-auto font-sans leading-normal">
              {activeModule.description}
            </p>
          )}
        </div>

        {/* Narrative Flow Directory */}
        <div className="space-y-12">
          {nodes.length === 0 ? (
            <div className="p-8 text-center text-sm font-sans italic text-gray-500">
              Belum ada langkah yang dituangkan ke dalam alur kerja ini.
            </div>
          ) : (
            nodes.map((node, i) => (
              <div key={node.id} className="pt-2 border-b border-dashed border-white/5 pb-10 last:border-none">
                
                {/* Node Label & Index Header */}
                <div className="flex items-center space-x-2.5 mb-5 font-sans">
                  <span className="bg-[#C5A267] text-black text-xs font-bold font-mono px-2 py-0.5 rounded">
                    {String(i + 1).padStart(2, '0')}
                  </span>
                  <h2 className="text-lg font-black tracking-tight text-white print:text-black">
                    {node.label}
                  </h2>
                  <span className="text-[10px] font-mono bg-white/5 px-2 py-0.5 rounded text-[#C5A267] uppercase tracking-widest font-bold">
                    {node.type}
                  </span>
                </div>

                {/* Grid Details */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6 text-sm font-sans my-4 bg-[#0B0B0C] p-5 rounded-2xl border border-white/5 print:bg-white print:border-none print:p-0">
                  
                  {/* Left Column: Business processes (Business Facet) */}
                  <div className="space-y-2 text-left">
                    <p className="border-b border-white/5 text-[10px] font-bold tracking-wider text-[#C5A267] uppercase font-mono pb-0.5">Spesifikasi Bisnis</p>
                    <div className="text-xs space-y-1.5">
                      <p><strong className="text-gray-400">Aktor Utama:</strong> <span className="text-white print:text-black">{node.doc.actor || 'N/A'}</span></p>
                      <p><strong className="text-gray-400">Pemicu:</strong> <span className="text-white print:text-black">{node.doc.trigger || 'N/A'}</span></p>
                      <p><strong className="text-gray-400">Input Proses:</strong> <span className="text-white print:text-black">{node.doc.input || 'N/A'}</span></p>
                      <p><strong className="text-gray-400">Uraian Alur:</strong> <span className="text-gray-300 print:text-black">{node.doc.process || 'N/A'}</span></p>
                      <p><strong className="text-gray-400">Output Proses:</strong> <span className="text-white print:text-black">{node.doc.output || 'N/A'}</span></p>
                      <p><strong className="text-gray-400">Target SLA:</strong> <span className="text-white print:text-black">{node.doc.sla || 'N/A'}</span></p>
                      <p><strong className="text-gray-400">Prioritas/Risiko:</strong> <span className="text-white print:text-black">{node.doc.priority || 'medium'} / {node.doc.riskLevel || 'medium'}</span></p>
                    </div>
                  </div>

                  {/* Right Column: Implementation targets (Roles Facets) */}
                  <div className="space-y-4 text-left">
                    
                    {/* UI/UX info */}
                    {node.roles.uiux && (
                      <div className="space-y-1 text-xs">
                        <p className="border-b border-white/5 text-[9px] font-bold tracking-wider text-[#E07A5F] uppercase font-mono pb-0.5">Artefak Desain UI/UX</p>
                        <p><strong className="text-gray-400">Penanggung Jawab:</strong> <span className="text-white print:text-black">{getDisplayAssignee(node.roles.uiux.assignee)}</span></p>
                        <p><strong className="text-gray-400">Nama Layar:</strong> <span className="text-white print:text-black">{node.roles.uiux.screen || 'N/A'}</span></p>
                        <p><strong className="text-gray-400">Status Kesiapan:</strong> <span className="font-mono text-[9px] uppercase font-bold text-[#E07A5F] bg-[#E07A5F]/5 px-1.5 py-0.5 rounded border border-[#E07A5F]/20">{node.roles.uiux.status}</span></p>
                      </div>
                    )}

                    {/* FE info */}
                    {node.roles.frontend && (
                      <div className="space-y-1 text-xs">
                        <p className="border-b border-white/5 text-[9px] font-bold tracking-wider text-[#4F9D9F] uppercase font-mono pb-0.5">Handoff Frontend (Klien)</p>
                        <p><strong className="text-gray-400">Penanggung Jawab:</strong> <span className="text-white print:text-black">{getDisplayAssignee(node.roles.frontend.assignee)}</span></p>
                        <p><strong className="text-gray-400">Halaman:</strong> <code className="text-[10px] bg-white/5 px-1 rounded font-mono text-gray-300 print:bg-slate-100">{node.roles.frontend.page || (node.roles.frontend as typeof node.roles.frontend & { component?: string }).component || 'N/A'}</code></p>
                        <p><strong className="text-gray-400">Jalur Rute:</strong> <code className="text-[10px] bg-white/5 px-1 rounded font-mono text-gray-300 print:bg-slate-100">{node.roles.frontend.route || 'N/A'}</code></p>
                        <p><strong className="text-gray-400">Interaksi:</strong> <span className="text-white print:text-black">{node.roles.frontend.interaction || 'N/A'}</span></p>
                        <p><strong className="text-gray-400">Status Kesiapan:</strong> <span className="font-mono text-[9px] uppercase font-bold text-[#4F9D9F] bg-[#4F9D9F]/5 px-1.5 py-0.5 rounded border border-[#4F9D9F]/20">{node.roles.frontend.status || 'Planned'}</span></p>
                      </div>
                    )}

                    {/* BE info */}
                    {node.roles.backend && (
                      <div className="space-y-1 text-xs">
                        <p className="border-b border-white/5 text-[9px] font-bold tracking-wider text-amber-500 uppercase font-mono pb-0.5">Implementasi Backend API</p>
                        <p><strong className="text-gray-400">Penanggung Jawab:</strong> <span className="text-white print:text-black">{getDisplayAssignee(node.roles.backend.assignee)}</span></p>
                        <p><strong className="text-gray-400">Endpoint Kontrak:</strong> <code className="text-[10px] bg-[#1A1A1D] border border-white/5 text-[#C5A267] px-1 rounded font-mono font-bold">{node.roles.backend.method || 'GET'} {node.roles.backend.endpoint || 'N/A'}</code></p>
                        <p><strong className="text-gray-400">Status Kesiapan:</strong> <span className="font-mono text-[9px] uppercase font-bold text-amber-500 bg-amber-500/5 px-1.5 py-0.5 rounded border border-amber-500/20">{node.roles.backend.status || 'Planned'}</span></p>
                      </div>
                    )}

                  </div>
                </div>

                {/* Substantive prose detailing business logic policies (Fills the narrative style) */}
                {node.doc.rules && (
                  <div className="pl-4 border-l-2 border-[#C5A267]/50 mt-2 text-xs italic text-gray-400 font-serif text-left">
                    <p className="font-sans font-bold text-[9px] tracking-wider text-[#C5A267] uppercase not-italic mb-1">Kebijakan Logika & Aturan Mutlak:</p>
                    {node.doc.rules}
                  </div>
                )}

                {node.doc.exceptionPath && (
                  <div className="pl-4 border-l-2 border-red-500/50 mt-2 text-xs italic text-gray-400 font-serif text-left">
                    <p className="font-sans font-bold text-[9px] tracking-wider text-red-400 uppercase not-italic mb-1">Alur Pengecualian:</p>
                    {node.doc.exceptionPath}
                  </div>
                )}

                {node.doc.acceptanceCriteria && (
                  <div className="pl-4 border-l-2 border-emerald-500/50 mt-2 text-xs italic text-gray-400 font-serif text-left">
                    <p className="font-sans font-bold text-[9px] tracking-wider text-emerald-400 uppercase not-italic mb-1">Kriteria Selesai:</p>
                    {node.doc.acceptanceCriteria}
                  </div>
                )}

              </div>
            ))
          )}
        </div>

        {/* Connections List at absolute bottom */}
        {activeModule.edges && activeModule.edges.length > 0 && (
          <div className="mt-12 pt-8 border-t border-white/5 text-left">
            <h3 className="text-sm font-bold text-[#C5A267] uppercase font-sans tracking-wider mb-4 print:text-black">
              Indeks Relasi Urutan Sambutan Alur
            </h3>
            <ul className="text-xs space-y-1.5 font-mono list-disc list-inside text-gray-400 print:text-slate-750">
              {activeModule.edges.map((edge) => {
                const f = activeModule.nodes.find((n) => n.id === edge.from)?.label || edge.from;
                const t = activeModule.nodes.find((n) => n.id === edge.to)?.label || edge.to;
                return (
                  <li key={edge.id}>
                    [{f}] {edge.label ? `--("${edge.label}")-->` : '---->' } [{t}]
                  </li>
                );
              })}
            </ul>
          </div>
        )}

      </div>
    </div>
  );
}
