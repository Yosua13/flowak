/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from 'react';
import { Node, Status } from '../../domain/types';
import { useStore } from '../../store/useStore';
import { ExternalLink, Code, Sparkles, Loader2 } from 'lucide-react';

interface FrontendTabProps {
  node: Node;
}

export default function FrontendTab({ node }: FrontendTabProps) {
  const { updateRole, teamMembers, addNotification, token } = useStore();
  const fe = node.roles?.frontend || { assignee: '', status: 'planned', component: '', route: '', framework: 'React', link: '' };

  // AI Boilerplate generation states
  const [aiGeneratingCode, setAiGeneratingCode] = useState(false);
  const [selectedFramework, setSelectedFramework] = useState('React + Tailwind');
  const [generatedCode, setGeneratedCode] = useState('');
  const [copiedCode, setCopiedCode] = useState(false);

  const handleFieldChange = (key: string, val: string) => {
    updateRole(node.id, 'frontend', { [key]: val });
  };

  const handleGenerateBoilerplate = async () => {
    setAiGeneratingCode(true);
    setGeneratedCode('');
    try {
      const res = await fetch('/api/ai/generate-code', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({
          type: 'frontend',
          nodeLabel: node.label,
          nodeType: node.type,
          detail: fe,
          framework: selectedFramework,
        }),
      });

      const data = await res.json();
      if (data.error) throw new Error(data.error);

      setGeneratedCode(data.code);
      addNotification('Komponen Terbuat', `Boilerplate frontend ${selectedFramework} untuk "${node.label}" berhasil dibuat.`, 'success');
    } catch (err: any) {
      console.error(err);
      setGeneratedCode(`// Gagal menghasilkan kode:\n// ${err.message}\n\ntry {\n  // Mohon hubungkan GEMINI_API_KEY\n}`);
    } finally {
      setAiGeneratingCode(false);
    }
  };

  const handleCopyCode = () => {
    navigator.clipboard.writeText(generatedCode);
    setCopiedCode(true);
    setTimeout(() => setCopiedCode(false), 2000);
  };

  const feMembers = teamMembers.filter((m) => m.role === 'frontend' || m.role === 'pm');

  return (
    <div className="space-y-4 text-left select-text pb-6">
      <div>
        <span className="text-[10px] font-bold font-mono tracking-widest text-[#C5A267] uppercase block mb-1">
          Spesifikasi Frontend (Komponen & Rute)
        </span>
        <p className="text-[11px] text-gray-400 mb-4 font-sans">
          Kelola rincian pengerjaan kode presentasi antarmuka/komponen klien.
        </p>
      </div>

      <div className="space-y-3.5">
        {/* Assignee selection */}
        <div className="space-y-1">
          <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">
            Penanggung Jawab (FE Engineer)
          </label>
          <select
            value={fe.assignee || ''}
            onChange={(e) => handleFieldChange('assignee', e.target.value)}
            className="w-full text-xs border border-white/5 rounded-xl px-3 py-2 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267] transition font-medium"
          >
            <option value="" className="bg-[#131315]">-- Pilih Anggota Tim --</option>
            {feMembers.map((m) => (
              <option key={m.id} value={m.name} className="bg-[#131315]">
                {m.name} ({m.email})
              </option>
            ))}
          </select>
        </div>

        {/* Status selection */}
        <div className="space-y-1">
          <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">
            Status Kesiapan Komponen
          </label>
          <select
            value={fe.status || 'planned'}
            onChange={(e) => handleFieldChange('status', e.target.value as Status)}
            className="w-full text-xs border border-white/5 rounded-xl px-3 py-2 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267] transition font-medium"
          >
            <option value="planned" className="bg-[#131315]">Planned (Terencana)</option>
            <option value="in_progress" className="bg-[#131315]">In Progress (Pengerjaan)</option>
            <option value="review" className="bg-[#131315]">In Review (Peninjauan)</option>
            <option value="done" className="bg-[#131315]">Done (Selesai)</option>
          </select>
        </div>

        {/* Component name */}
        <div className="space-y-1">
          <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">
            Nama File Komponen (Font Mono)
          </label>
          <input
            type="text"
            value={fe.component || ''}
            onChange={(e) => handleFieldChange('component', e.target.value)}
            placeholder="Contoh: LeaveForm.tsx, SidebarRouter.tsx..."
            className="w-full text-xs font-mono border border-white/5 rounded-xl px-3 py-2 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267] transition font-medium font-mono"
          />
        </div>

        {/* Route endpoint name */}
        <div className="space-y-1">
          <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">
            Jalur Rute Klien (URL Route)
          </label>
          <input
            type="text"
            value={fe.route || ''}
            onChange={(e) => handleFieldChange('route', e.target.value)}
            placeholder="Contoh: /dashboard/leaves, /apply..."
            className="w-full text-xs font-mono border border-white/5 rounded-xl px-3 py-2 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267] transition font-medium font-mono"
          />
        </div>

        {/* Framework layout */}
        <div className="space-y-1">
          <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">
            Kerangka Kerja / Library
          </label>
          <input
            type="text"
            value={fe.framework || 'React + Tailwind'}
            onChange={(e) => handleFieldChange('framework', e.target.value)}
            placeholder="Contoh: React, Angular, Tailwind CSS..."
            className="w-full text-xs border border-white/5 rounded-xl px-3 py-2 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267] transition font-medium"
          />
        </div>

        {/* Storybook / Repo link */}
        <div className="space-y-1">
          <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">
            Tautan Repositori / Storybook Komponen
          </label>
          <div className="flex space-x-2">
            <input
              type="url"
              value={fe.link || ''}
              onChange={(e) => handleFieldChange('link', e.target.value)}
              placeholder="https://github.com/flowak/fe/..."
              className="flex-1 text-xs border border-white/5 rounded-xl px-3 py-2 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267] transition font-medium"
            />
            {fe.link && (
              <a
                href={fe.link}
                target="_blank"
                rel="noreferrer"
                className="p-2 border border-white/10 hover:bg-white/5 rounded-xl flex items-center justify-center text-gray-400 hover:text-white cursor-pointer"
                title="Buka Repositori"
              >
                <ExternalLink className="w-4 h-4" />
              </a>
            )}
          </div>
        </div>

        {/* AI Boilerplate Code Generator Section C */}
        <div className="pt-4 mt-2 border-t border-white/5">
          <span className="text-[10px] font-bold font-mono tracking-widest text-[#C5A267] uppercase block mb-2">
            AI Frontend Code Builder
          </span>
          <div className="flex space-x-2 mb-3">
            <select
              value={selectedFramework}
              onChange={(e) => setSelectedFramework(e.target.value)}
              className="flex-1 text-xs border border-white/5 rounded-xl px-3 py-2 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267]"
            >
              <option value="React + Tailwind TSX">React + Tailwind Functional component</option>
              <option value="Vue 3 SFC Composition">Vue 3 Single File Component (Composition API)</option>
              <option value="HTML5 Native UI CSS">Vanilla HTML Canvas + CSS Components</option>
            </select>
            <button
              onClick={handleGenerateBoilerplate}
              disabled={aiGeneratingCode || !fe.component}
              className="flex items-center justify-center space-x-1.5 px-4 bg-white/5 border border-white/10 hover:bg-white/15 rounded-xl text-xs font-semibold text-white cursor-pointer transition disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {aiGeneratingCode ? (
                <Loader2 className="w-3.5 h-3.5 animate-spin whitespace-nowrap" />
              ) : (
                <Code className="w-3.5 h-3.5" />
              )}
              <span className="whitespace-nowrap">BUAT FE KODE</span>
            </button>
          </div>

          {generatedCode && (
            <div className="mt-3 bg-[#0A0A0B] border border-[#C5A267]/20 rounded-xl p-3 text-[10px] font-mono leading-relaxed relative">
              <div className="flex items-center justify-between border-b border-white/5 pb-2 mb-2">
                <span className="text-white font-mono text-[9px] uppercase font-bold text-gray-400">{selectedFramework}</span>
                <button
                  onClick={handleCopyCode}
                  className="text-[#C5A267] font-mono text-[9px] hover:underline"
                >
                  {copiedCode ? 'TERSALIN' : 'SALIN'}
                </button>
              </div>
              <pre className="text-blue-400 overflow-x-auto whitespace-pre leading-relaxed max-h-56 leading-normal p-1 scrollbar-thin">
                {generatedCode}
              </pre>
            </div>
          )}
        </div>

      </div>
    </div>
  );
}
