/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from 'react';
import { useStore } from '../../store/useStore';
import { Users, Plus, Mail, Trash2, UserPlus } from 'lucide-react';

export default function TeamView() {
  const { 
    teamMembers, 
    addTeamMember, 
    currentUser,
    projectMembers,
    addProjectMember,
    deleteProjectMember
  } = useStore();

  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [role, setRole] = useState('frontend');
  const [selectedGlobalUser, setSelectedGlobalUser] = useState('');

  const isPM = currentUser?.role === 'pm';

  const handleRegisterUser = (e: React.FormEvent) => {
    e.preventDefault();
    if (name.trim() && email.trim()) {
      addTeamMember(name.trim(), email.trim(), role);
      setName('');
      setEmail('');
    }
  };

  const handleAddMemberToProject = (e: React.FormEvent) => {
    e.preventDefault();
    if (selectedGlobalUser) {
      addProjectMember(selectedGlobalUser);
      setSelectedGlobalUser('');
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

  // Find global users who are NOT yet project members
  const availableGlobalUsers = teamMembers.filter(
    (globalUser) => !projectMembers.some((member) => member.id === globalUser.id)
  );

  return (
    <div className="flex-1 bg-[#0A0A0B] p-6 overflow-y-auto scrollbar-thin select-none text-left">
      
      {/* View Header */}
      <div className="mb-6 border-b border-white/5 pb-4">
        <h2 className="text-sm font-bold text-white flex items-center uppercase tracking-widest" style={{ fontFamily: 'Georgia, serif' }}>
          <Users className="w-5 h-5 mr-2 text-[#C5A267]" />
          Kolaborasi & Anggota Tim Proyek
        </h2>
        <p className="text-xs text-gray-400 mt-1 leading-snug">
          Kelola siapa saja yang memiliki akses ke proyek ini, dan tambahkan kontributor baru ke sistem.
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 items-start">
        
        {/* Left Column: Project Members */}
        <div className="lg:col-span-2 space-y-6">
          <div className="bg-[#131315] border border-white/5 p-5 rounded-2xl shadow-xl">
            <span className="text-[10px] font-bold font-mono tracking-widest text-[#C5A267] uppercase block border-b border-white/5 pb-2 mb-4">
              Anggota Proyek Ini ({projectMembers.length})
            </span>

            {projectMembers.length === 0 ? (
              <p className="text-xs text-gray-500 py-6 text-center">Belum ada anggota tim tambahan di proyek ini. PM dapat menambahkan di samping/bawah.</p>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {projectMembers.map((member) => (
                  <div
                    key={member.id}
                    className="p-4 rounded-xl border border-white/5 bg-[#1A1A1D]/80 flex items-center justify-between hover:border-white/10 transition group"
                  >
                    <div className="flex items-center space-x-3.5">
                      <img
                        src={member.avatar || `https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=100&fit=crop`}
                        alt={member.name}
                        className="w-10 h-10 rounded-full border border-white/10 object-cover"
                      />
                      <div className="space-y-0.5">
                        <p className="text-xs font-bold text-white font-sans">{member.name}</p>
                        <p className="text-[10px] text-gray-400 flex items-center font-sans font-mono">
                          <Mail className="w-3 h-3 mr-1 text-gray-500" /> {member.email}
                        </p>
                        <span className={`inline-block text-[8px] font-bold font-mono px-1.5 py-0.5 rounded border mt-1.5 uppercase ${getRoleBadgeColor(member.role)}`}>
                          {capitalize(member.role)}
                        </span>
                      </div>
                    </div>

                    {isPM && currentUser && member.id !== currentUser.id && (
                      <button
                        onClick={() => deleteProjectMember(member.id)}
                        className="p-2 border border-red-500/10 hover:bg-red-500/10 rounded-xl text-red-400 cursor-pointer opacity-0 group-hover:opacity-100 transition-opacity"
                        title="Hapus dari Proyek"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* PM Area: Add Member to Project */}
          {isPM && (
            <div className="bg-[#131315] border border-white/5 p-5 rounded-2xl shadow-xl">
              <span className="text-[10px] font-bold font-mono tracking-widest text-[#C5A267] uppercase block border-b border-white/5 pb-2 mb-4">
                Tambahkan Anggota Tim ke Proyek
              </span>

              <form onSubmit={handleAddMemberToProject} className="flex flex-col sm:flex-row items-end gap-4">
                <div className="flex-1 w-full space-y-1">
                  <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">
                    PILIH KONTRIBUTOR TERDAFTAR
                  </label>
                  <select
                    value={selectedGlobalUser}
                    onChange={(e) => setSelectedGlobalUser(e.target.value)}
                    required
                    className="w-full text-xs border border-white/5 rounded-xl px-3 py-2.5 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267]"
                  >
                    <option value="">-- Pilih Kontributor --</option>
                    {availableGlobalUsers.map((user) => (
                      <option key={user.id} value={user.id}>
                        {user.name} ({capitalize(user.role)}) - {user.email}
                      </option>
                    ))}
                  </select>
                </div>
                <button
                  type="submit"
                  disabled={!selectedGlobalUser}
                  className="w-full sm:w-auto flex items-center justify-center space-x-1.5 px-6 py-2.5 rounded-xl bg-[#C5A267] hover:bg-[#B59155] text-black text-xs font-bold font-mono tracking-wider cursor-pointer disabled:opacity-45 disabled:hover:bg-[#C5A267]"
                >
                  <UserPlus className="w-3.5 h-3.5" />
                  <span>TAMBAH KE PROYEK</span>
                </button>
              </form>
            </div>
          )}
        </div>

        {/* Right Column: Register Global Contributors */}
        <div className="space-y-6">
          {isPM && (
            <div className="bg-[#131315] border border-white/5 p-5 rounded-2xl shadow-xl">
              <span className="text-[10px] font-bold font-mono tracking-widest text-[#C5A267] uppercase block border-b border-white/5 pb-2 mb-4">
                Daftarkan Akun Baru (Global)
              </span>

              <form onSubmit={handleRegisterUser} className="space-y-4">
                <div className="space-y-1">
                  <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">
                    NAMA LENGKAP KONTRIBUTOR
                  </label>
                  <input
                    type="text"
                    required
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder="Contoh: Farhan Sanjaya"
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
                    className="w-full text-xs border border-white/5 rounded-xl px-3 py-2.5 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267]"
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
                  <span>DAFTARKAN AKUN</span>
                </button>
              </form>
            </div>
          )}

          {/* Global System Users List */}
          <div className="bg-[#131315] border border-white/5 p-5 rounded-2xl shadow-xl max-h-[400px] overflow-y-auto scrollbar-none">
            <span className="text-[10px] font-bold font-mono tracking-widest text-[#C5A267] uppercase block border-b border-white/5 pb-2 mb-4">
              Semua Akun Terdaftar ({teamMembers.length})
            </span>

            <div className="space-y-3">
              {teamMembers.map((member) => (
                <div key={member.id} className="flex items-center justify-between p-3 rounded-xl bg-white/[0.01] border border-white/5">
                  <div className="space-y-0.5">
                    <p className="text-[11px] font-bold text-white">{member.name}</p>
                    <p className="text-[9px] text-gray-500 font-mono">{member.email}</p>
                  </div>
                  <span className={`text-[8px] font-mono font-bold px-2 py-0.5 rounded border ${getRoleBadgeColor(member.role)}`}>
                    {member.role.toUpperCase()}
                  </span>
                </div>
              ))}
            </div>
          </div>
        </div>

      </div>
    </div>
  );
}
