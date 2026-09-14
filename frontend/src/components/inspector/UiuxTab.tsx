/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { Node, Status, UiuxFacet } from '../../domain/types';
import { useStore } from '../../store/useStore';
import { ExternalLink } from 'lucide-react';

interface UiuxTabProps {
  node: Node;
}

export default function UiuxTab({ node }: UiuxTabProps) {
  const { updateRole, teamMembers } = useStore();
  const uiux: UiuxFacet = node.roles?.uiux || {
    assignee: '',
    status: 'planned',
    screen: '',
    link: '',
    wireframeUrl: '',
    stateNotes: '',
    accessibilityNotes: '',
    dueDate: '',
  };

  const handleFieldChange = (key: string, val: string) => {
    updateRole(node.id, 'uiux', { [key]: val });
  };

  const uiuxMembers = teamMembers.filter((m) => m.role === 'uiux' || m.role === 'pm');

  return (
    <div className="space-y-4 text-left select-text">
      <div>
        <span className="text-[10px] font-bold font-mono tracking-widest text-[#C5A267] uppercase block mb-1">
          Spesifikasi UI/UX (Desain Layar)
        </span>
        <p className="text-[11px] text-gray-400 mb-4 font-sans">
          Petakan visual aplikasi dan simpan navigasi mockup Figma yang relevan.
        </p>
      </div>

      <div className="space-y-3.5">
        {/* Assignee Selection */}
        <div className="space-y-1">
          <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">
            Penanggung Jawab (Desainer)
          </label>
          <select
            value={uiux.assignee || ''}
            onChange={(e) => handleFieldChange('assignee', e.target.value)}
            className="w-full text-xs border border-white/5 rounded-xl px-3 py-2 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267] transition font-medium"
          >
            <option value="" className="bg-[#131315]">-- Pilih Anggota Tim --</option>
            {uiuxMembers.map((m) => (
              <option key={m.id} value={m.name} className="bg-[#131315]">
                {m.name} ({m.email})
              </option>
            ))}
          </select>
        </div>

        <div className="grid grid-cols-2 gap-2">
          <div className="space-y-1">
            <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">
              Status Desain
            </label>
            <select
              value={uiux.status || 'planned'}
              onChange={(e) => handleFieldChange('status', e.target.value as Status)}
              className="w-full text-xs border border-white/5 rounded-xl px-3 py-2 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267] transition font-medium"
            >
              <option value="planned" className="bg-[#131315]">Planned</option>
              <option value="in_progress" className="bg-[#131315]">In Progress</option>
              <option value="review" className="bg-[#131315]">In Review</option>
              <option value="done" className="bg-[#131315]">Done</option>
            </select>
          </div>

          <div className="space-y-1">
            <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">
              Target Selesai
            </label>
            <input
              type="date"
              value={uiux.dueDate || ''}
              onChange={(e) => handleFieldChange('dueDate', e.target.value)}
              className="w-full text-xs border border-white/5 rounded-xl px-3 py-2 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267] transition font-medium"
            />
          </div>
        </div>

        {/* Screen layout name */}
        <div className="space-y-1">
          <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">
            Nama Layar / Laman
          </label>
          <input
            type="text"
            value={uiux.screen || ''}
            onChange={(e) => handleFieldChange('screen', e.target.value)}
            placeholder="Contoh: Form input cuti, view_manager_portal..."
            className="w-full text-xs border border-white/5 rounded-xl px-3 py-2 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267] transition font-medium"
          />
        </div>

        {/* Figma link */}
        <div className="space-y-1">
          <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">
            Tautan Figma File
          </label>
          <div className="flex space-x-2">
            <input
               type="url"
               value={uiux.link || ''}
               onChange={(e) => handleFieldChange('link', e.target.value)}
               placeholder="https://figma.com/file/..."
               className="flex-1 text-xs border border-white/5 rounded-xl px-3 py-2 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267] transition font-medium"
            />
            {uiux.link && (
              <a
                href={uiux.link}
                target="_blank"
                rel="noreferrer"
                className="p-2 border border-white/10 hover:bg-white/5 rounded-xl flex items-center justify-center text-gray-400 hover:text-white cursor-pointer"
                title="Buka Mockup Figma"
              >
                <ExternalLink className="w-4 h-4" />
              </a>
            )}
          </div>
        </div>

        <div className="space-y-1">
          <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">
            Link Wireframe / Preview
          </label>
          <input
            type="url"
            value={uiux.wireframeUrl || ''}
            onChange={(e) => handleFieldChange('wireframeUrl', e.target.value)}
            placeholder="https://..."
            className="w-full text-xs border border-white/5 rounded-xl px-3 py-2 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267] transition font-medium"
          />
        </div>

        <div className="space-y-1">
          <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">
            State & Aksesibilitas
          </label>
          <textarea
            value={uiux.stateNotes || ''}
            onChange={(e) => handleFieldChange('stateNotes', e.target.value)}
            placeholder="Empty/loading/error/success state utama."
            rows={2}
            className="w-full text-xs border border-white/5 rounded-xl px-3 py-2 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267] transition font-medium h-16 resize-none"
          />
          <textarea
            value={uiux.accessibilityNotes || ''}
            onChange={(e) => handleFieldChange('accessibilityNotes', e.target.value)}
            placeholder="Kontras, keyboard focus, label form, dan kebutuhan aksesibilitas lain."
            rows={2}
            className="w-full text-xs border border-white/5 rounded-xl px-3 py-2 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267] transition font-medium h-16 resize-none"
          />
        </div>

        {([
          ['userGoal', 'Tujuan pengguna'], ['surface', 'Surface / platform'], ['designVersion', 'Versi desain'],
          ['screenStates', 'Screen states'], ['interactions', 'Interaksi'], ['contentMessages', 'Konten & pesan'],
          ['responsiveIntent', 'Responsif'], ['accessibilityNotes', 'Aksesibilitas'], ['notes', 'Checklist handoff & evidence'],
        ] as const).map(([key, label]) => (
          <div className="space-y-1" key={key}>
            <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">{label}</label>
            <textarea value={uiux[key] || ''} onChange={(e) => handleFieldChange(key, e.target.value)} rows={2}
              className="w-full text-xs border border-white/5 rounded-xl px-3 py-2 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267] h-16 resize-none" />
          </div>
        ))}

      </div>
    </div>
  );
}
