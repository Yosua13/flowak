/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { Node, Status } from '../../domain/types';
import { useStore } from '../../store/useStore';
import { ExternalLink, FileText, MousePointer2, Route } from 'lucide-react';

interface FrontendTabProps {
  node: Node;
}

export default function FrontendTab({ node }: FrontendTabProps) {
  const { updateRole, teamMembers } = useStore();
  const fe = node.roles?.frontend || {
    assignee: '',
    status: 'planned',
    page: '',
    route: '',
    interaction: '',
    validation: '',
    state: '',
    handoffLink: '',
    dueDate: '',
    notes: '',
  };

  const legacyFe = fe as typeof fe & { component?: string; link?: string };
  const pageName = fe.page || legacyFe.component || '';
  const handoffLink = fe.handoffLink || legacyFe.link || '';

  const handleFieldChange = (key: string, val: string) => {
    updateRole(node.id, 'frontend', { [key]: val });
  };

  const feMembers = teamMembers.filter((m) => m.role === 'frontend' || m.role === 'pm');

  return (
    <div className="space-y-4 text-left select-text pb-6">
      <div>
        <span className="text-[10px] font-bold font-mono tracking-widest text-[#C5A267] uppercase block mb-1">
          Handoff Frontend (Halaman & Interaksi)
        </span>
        <p className="text-[11px] text-gray-400 mb-4 font-sans">
          Catat kebutuhan pengalaman pengguna yang harus diterjemahkan menjadi layar, rute, validasi, dan state aplikasi.
        </p>
      </div>

      <div className="space-y-3.5">
        <div className="space-y-1">
          <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">
            Penanggung Jawab
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

        <div className="grid grid-cols-2 gap-2">
          <div className="space-y-1">
            <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">
              Status Handoff
            </label>
            <select
              value={fe.status || 'planned'}
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
              value={fe.dueDate || ''}
              onChange={(e) => handleFieldChange('dueDate', e.target.value)}
              className="w-full text-xs border border-white/5 rounded-xl px-3 py-2 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267] transition font-medium"
            />
          </div>
        </div>

        <div className="space-y-1">
          <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">
            Nama Halaman / Pengalaman
          </label>
          <div className="relative">
            <FileText className="w-3.5 h-3.5 text-gray-500 absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              value={pageName}
              onChange={(e) => handleFieldChange('page', e.target.value)}
              placeholder="Contoh: Form Pengajuan Cuti"
              className="w-full text-xs border border-white/5 rounded-xl pl-9 pr-3 py-2 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267] transition font-medium"
            />
          </div>
        </div>

        <div className="space-y-1">
          <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">
            Rute / Jalur Navigasi
          </label>
          <div className="relative">
            <Route className="w-3.5 h-3.5 text-gray-500 absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              value={fe.route || ''}
              onChange={(e) => handleFieldChange('route', e.target.value)}
              placeholder="/pengajuan/cuti"
              className="w-full text-xs font-mono border border-white/5 rounded-xl pl-9 pr-3 py-2 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267] transition font-medium"
            />
          </div>
        </div>

        <div className="space-y-1">
          <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">
            Perilaku Interaksi
          </label>
          <textarea
            value={fe.interaction || ''}
            onChange={(e) => handleFieldChange('interaction', e.target.value)}
            placeholder="Contoh: submit menampilkan loading, sukses membuka ringkasan, gagal menampilkan pesan validasi."
            rows={3}
            className="w-full text-xs border border-white/5 rounded-xl px-3 py-2 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267] transition font-medium h-20 resize-none"
          />
        </div>

        <div className="space-y-1">
          <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">
            Validasi & Pesan Pengguna
          </label>
          <textarea
            value={fe.validation || ''}
            onChange={(e) => handleFieldChange('validation', e.target.value)}
            placeholder="Contoh: tanggal mulai wajib lebih besar dari hari ini; alasan wajib diisi."
            rows={3}
            className="w-full text-xs border border-white/5 rounded-xl px-3 py-2 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267] transition font-medium h-20 resize-none"
          />
        </div>

        <div className="space-y-1">
          <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">
            State Handling
          </label>
          <div className="relative">
            <MousePointer2 className="w-3.5 h-3.5 text-gray-500 absolute left-3 top-3" />
            <textarea
              value={fe.state || ''}
              onChange={(e) => handleFieldChange('state', e.target.value)}
              placeholder="Loading, empty state, retry, success, error, dan edge case utama."
              rows={3}
              className="w-full text-xs border border-white/5 rounded-xl pl-9 pr-3 py-2 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267] transition font-medium h-20 resize-none"
            />
          </div>
        </div>

        <div className="space-y-1">
          <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">
            Link Handoff / Preview
          </label>
          <div className="flex space-x-2">
            <input
              type="url"
              value={handoffLink}
              onChange={(e) => handleFieldChange('handoffLink', e.target.value)}
              placeholder="https://..."
              className="flex-1 text-xs border border-white/5 rounded-xl px-3 py-2 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267] transition font-medium"
            />
            {handoffLink && (
              <a
                href={handoffLink}
                target="_blank"
                rel="noreferrer"
                className="p-2 border border-white/10 hover:bg-white/5 rounded-xl flex items-center justify-center text-gray-400 hover:text-white cursor-pointer"
                title="Buka Handoff"
              >
                <ExternalLink className="w-4 h-4" />
              </a>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
