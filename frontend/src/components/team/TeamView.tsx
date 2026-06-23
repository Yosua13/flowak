/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from 'react';
import { useStore } from '../../store/useStore';
import { Users, Plus, Mail, ShieldAlert, Award, Trash2 } from 'lucide-react';

export default function TeamView() {
  const { teamMembers, addTeamMember, deleteTeamMember } = useStore();

  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [role, setRole] = useState('frontend');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (name.trim() && email.trim()) {
      addTeamMember(name.trim(), email.trim(), role);
      setName('');
      setEmail('');
    }
  };

  const capitalize = (s: string) => {
    if (!s) return '';
    return s.toUpperCase();
  };

  const getRoleBadgeColor = (r: string) => {
    switch (r) {
      case 'pm': return 'bg-amber-500/10 text-amber-400 border-amber-500/20';
      case 'uiux': return 'bg-pink-500/10 text-pink-400 border-pink-500/20';
      case 'frontend': return 'bg-sky-500/10 text-sky-400 border-sky-500/20';
      case 'backend': return 'bg-purple-500/10 text-purple-400 border-purple-500/20';
      default: return 'bg-slate-500/10 text-slate-400 border-slate-500/20';
    }
  };

  return (
    <div className="flex-1 bg-[#0A0A0B] p-6 overflow-y-auto scrollbar-thin select-none text-left">
      
      {/* View Header */}
      <div className="mb-6 border-b border-white/5 pb-4">
        <h2 className="text-sm font-bold text-white flex items-center uppercase tracking-widest" style={{ fontFamily: 'Georgia, serif' }}>
          <Users className="w-5 h-5 mr-2 text-[#C5A267]" />
          Kelola Anggota Tim Rilis (Kontributor)
        </h2>
        <p className="text-xs text-gray-400 mt-1 leading-snug">
          Manipulasi, hapus, dan tambahkan tenaga ahli yang bertugas merekayasa alur proses fungsional workspace ini secara real time.
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 items-start">
        {/* Contributors List Panel */}
        <div className="lg:col-span-2 bg-[#131315] border border-white/5 p-5 rounded-2xl shadow-xl">
          <span className="text-[10px] font-bold font-mono tracking-widest text-[#C5A267] uppercase block border-b border-white/5 pb-2 mb-4">
            Daftar Kontributor Aktif ({teamMembers.length})
          </span>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {teamMembers.map((member) => (
              <div
                key={member.id}
                className="p-4 rounded-xl border border-white/5 bg-[#1A1A1D]/80 flex items-center justify-between hover:border-white/10 transition group"
              >
                <div className="flex items-center space-x-3.5">
                  <img
                    src={member.avatar || `https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=100&fit=crop`}
                    alt={member.name}
                    aria-label={member.name}
                    referrerPolicy="no-referrer"
                    className="w-10 h-10 rounded-full border border-white/10 object-cover"
                  />
                  <div className="space-y-0.5">
                    <p className="text-xs font-bold text-white font-sans">{member.name}</p>
                    <p className="text-[10px] text-gray-400 flex items-center font-sans">
                      <Mail className="w-3 h-3 mr-1 text-gray-500" /> {member.email}
                    </p>
                    <span className={`inline-block text-[8px] font-bold font-mono px-1.5 py-0.5 rounded border mt-1.5 uppercase ${getRoleBadgeColor(member.role)}`}>
                      {capitalize(member.role)}
                    </span>
                  </div>
                </div>

                {/* Restrict deleting core seeded members to promote layout stability */}
                {member.id.startsWith('member_') ? (
                  <button
                    onClick={() => deleteTeamMember(member.id)}
                    className="p-2 border border-red-500/10 hover:bg-red-500/10 rounded-xl text-red-400 cursor-pointer opacity-0 group-hover:opacity-100 transition-opacity"
                    title="Keluarkan Kontributor"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                ) : (
                  <span className="text-[8px] font-mono font-bold text-gray-650 tracking-widest">SISTEM</span>
                )}
              </div>
            ))}
          </div>
        </div>

        {/* Add Member Form Panel */}
        <div className="bg-[#131315] border border-white/5 p-5 rounded-2xl shadow-xl">
          <span className="text-[10px] font-bold font-mono tracking-widest text-[#C5A267] uppercase block border-b border-white/5 pb-2 mb-4">
            Masukkan Kontributor Baru
          </span>

          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-1">
              <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">
                NAMA LENGKAP KONTRIBUTOR
              </label>
              <input
                type="text"
                required
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Contoh: Farhan Sanjaya, Lia Kartika..."
                className="w-full text-xs border border-white/5 rounded-xl px-3 py-2.5 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267] font-medium"
              />
            </div>

            <div className="space-y-1">
              <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">
                ALAMAT EMAIL RESMI
              </label>
              <input
                type="email"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="designer1@flowakportal.com"
                className="w-full text-xs font-mono border border-white/5 rounded-xl px-3 py-2.5 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267]"
              />
            </div>

            <div className="space-y-1">
              <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">
                DIVISI PERAN UTAMA (TIM)
              </label>
              <select
                value={role}
                onChange={(e) => setRole(e.target.value)}
                className="w-full text-xs border border-white/5 rounded-xl px-3 py-2.5 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267] font-semibold"
              >
                <option value="uiux">UI/UX Designer</option>
                <option value="frontend">Frontend Engineer (FE)</option>
                <option value="backend">Backend Engineer (BE)</option>
                <option value="pm">Product Manager (PM)</option>
              </select>
            </div>

            <button
              type="submit"
              className="w-full flex items-center justify-center space-x-1.5 py-2.5 rounded-xl bg-[#C5A267] hover:bg-[#B59155] text-black text-xs font-bold font-mono tracking-wider cursor-pointer"
            >
              <Plus className="w-3.5 h-3.5" />
              <span>DAFTARKAN KONTRIBUTOR</span>
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
