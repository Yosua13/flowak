/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { Node, Status } from '../../domain/types';
import { useStore } from '../../store/useStore';
import { ExternalLink, Palette, Smartphone } from 'lucide-react';

interface UiuxTabProps {
  node: Node;
}

export default function UiuxTab({ node }: UiuxTabProps) {
  const { updateRole, teamMembers } = useStore();
  const uiux = node.roles?.uiux || {
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

        {/* Placeholder wireframe mockup generator */}
        <div className="pt-2">
          <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider mb-2">
            Pratinjau Mockup Layar (Wireframe)
          </label>
          <div className="border border-white/5 rounded-2xl bg-[#0A0A0B] p-4 flex flex-col items-center justify-center text-center">
            {uiux.screen ? (
              <div className="w-full max-w-48 bg-[#131315] border border-white/10 rounded-2xl p-3 shadow-2xl font-sans">
                {/* Simulated Phone UI */}
                <div className="w-8 h-1 bg-white/10 rounded-full mx-auto mb-3" />
                <div className="border-b border-white/5 pb-2 mb-2 flex items-center justify-between">
                  <span className="text-[10px] font-bold text-white truncate w-24 block text-left">
                    {uiux.screen}
                  </span>
                  <Palette className="w-3" style={{ color: '#C5A267' }} />
                </div>
                {/* Simulated content blocks */}
                <div className="space-y-2">
                  <div className="h-6 rounded bg-[#1A1A1D] border border-white/5 flex items-center px-1.5 justify-between">
                    <div className="w-8 h-2 bg-white/10 rounded" />
                    <div className="w-3 h-3 rounded-full bg-white/10" />
                  </div>
                  <div className="h-10 rounded bg-[#1A1A1D] border border-white/5 flex-col flex justify-center p-1.5 space-y-1.5">
                    <div className="w-12 h-1.5 bg-white/15 rounded" />
                    <div className="w-20 h-1.5 bg-white/10 rounded" />
                  </div>
                  <div className="h-6 rounded-md bg-[#C5A267] text-black text-[8px] font-mono font-bold flex items-center justify-center">
                    KIRIM APPROVAL
                  </div>
                </div>
              </div>
            ) : (
              <div className="py-6 text-gray-500">
                <Smartphone className="w-8 h-8 mx-auto mb-2 text-white/5" />
                <p className="text-[10px] font-medium max-w-40 mx-auto">Tulis nama layar diatas untuk merender wireframe interaktif otomatis.</p>
              </div>
            )}
          </div>
        </div>

      </div>
    </div>
  );
}
