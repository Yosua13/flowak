/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from 'react';
import { useStore } from '../store/useStore';
import { Folder, Plus, Trash2, Users, LayoutDashboard, ChevronLeft, ChevronRight, Moon, Sun, Map, Upload } from 'lucide-react';
import { motion, AnimatePresence } from 'motion/react';

export default function Sidebar() {
  const {
    modules,
    activeId,
    selectModule,
    addModule,
    deleteModule,
    teamMembers,
    darkMode,
    toggleDarkMode,
    importModule,
    addNotification,
    selectProject,
  } = useStore();

  const [isAdding, setIsAdding] = useState(false);
  const [newModuleName, setNewModuleName] = useState('');
  const [newModuleDesc, setNewModuleDesc] = useState('');
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);

  const activeModule = modules.find((m) => m.id === activeId);

  const handleAddModule = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newModuleName.trim()) return;
    addModule(newModuleName.trim(), newModuleDesc.trim());
    setNewModuleName('');
    setNewModuleDesc('');
    setIsAdding(false);
  };

  const handleImportJson = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = (event) => {
      try {
        const parsed = JSON.parse(event.target?.result as string);
        importModule(parsed);
        addNotification('Modul Diimpor', `Modul "${parsed.name || 'Alur Kerja'}" berhasil diimpor ke sistem.`, 'success');
      } catch (err) {
        addNotification('Gagal Impor', 'Format file JSON tidak sah.', 'warning');
      }
    };
    reader.readAsText(file);
    // Reset file input value
    e.target.value = '';
  };

  const handleDeleteClick = (id: string, name: string, e: React.MouseEvent) => {
    e.stopPropagation();
    const confirmed = window.confirm(`Apakah Anda yakin ingin menghapus modul "${name}" secara permanen beserta seluruh langkah kegiatannya?`);
    if (confirmed) {
      deleteModule(id);
    }
  };

  return (
    <div
      className={`h-screen flex flex-col border-r bg-[#111113] text-gray-200 border-white/10 transition-all duration-300 relative ${
        sidebarCollapsed ? 'w-16' : 'w-72'
      }`}
    >
      {/* Brand Header */}
      <div className="p-4 flex items-center justify-between border-b border-white/5">
        {!sidebarCollapsed && (
          <div className="flex items-center space-x-2 p-1 text-left">
            <div className="w-3 h-3 rounded-full bg-[#C5A267] shadow-[0_0_10px_#C5A267] flex-shrink-0 animate-pulse"></div>
            <div>
              <h1 className="text-2xl font-light tracking-tighter text-white flex items-center gap-1.5" style={{ fontFamily: 'Georgia, serif' }}>
                <span className="italic">Flowak</span>
              </h1>
              <span className="text-[9px] text-[#C5A267] font-mono tracking-widest block uppercase leading-none mt-1">
                Traceability Spine
              </span>
            </div>
          </div>
        )}
        {sidebarCollapsed && (
          <div className="w-3 h-3 rounded-full bg-[#C5A267] shadow-[0_0_10px_#C5A267] mx-auto animate-pulse"></div>
        )}

        <button
          onClick={() => setSidebarCollapsed(!sidebarCollapsed)}
          className="absolute -right-3 top-16 bg-[#C5A267] hover:bg-[#B38F52] text-black rounded-full p-1 border shadow-md border-white/10 z-50 cursor-pointer transition-all duration-200"
        >
          {sidebarCollapsed ? <ChevronRight className="w-3.5 h-3.5" /> : <ChevronLeft className="w-3.5 h-3.5" />}
        </button>
      </div>

      {/* Module Navigation */}
      <div className="flex-1 overflow-y-auto p-3 space-y-4 scrollbar-thin">
        {/* Kembali ke Dashboard */}
        {!sidebarCollapsed && (
          <button
            onClick={() => selectProject(null)}
            className="w-full flex items-center space-x-2 px-3 py-2 rounded-lg text-xs font-semibold text-gray-400 hover:text-white hover:bg-white/5 cursor-pointer transition mb-2"
          >
            <ChevronLeft className="w-4 h-4 text-[#C5A267]" />
            <span>Kembali ke Galeri</span>
          </button>
        )}
        <div>
          <div className="flex items-center justify-between px-2 mb-2">
            {!sidebarCollapsed && <span className="text-xs font-semibold text-gray-400 uppercase tracking-widest">Daftar Modul</span>}
            {!sidebarCollapsed && (
              <div className="flex items-center space-x-1">
                {/* Impor JSON button trigger */}
                <button
                  onClick={() => document.getElementById('sidebar-import-file')?.click()}
                  className="text-[#C5A267] hover:text-[#E2C392] p-1 rounded-md hover:bg-white/5 transition-colors cursor-pointer"
                  title="Impor Modul (JSON)"
                >
                  <Upload className="w-3.5 h-3.5" />
                </button>
                <input
                  type="file"
                  id="sidebar-import-file"
                  accept=".json"
                  onChange={handleImportJson}
                  className="hidden"
                />
                
                <button
                  onClick={() => setIsAdding(true)}
                  className="text-[#C5A267] hover:text-[#E2C392] p-1 rounded-md hover:bg-white/5 transition-colors cursor-pointer"
                  title="Tambah Modul"
                >
                  <Plus className="w-4 h-4" />
                </button>
              </div>
            )}
          </div>

          {isAdding && !sidebarCollapsed && (
            <motion.form
              initial={{ opacity: 0, y: -10 }}
              animate={{ opacity: 1, y: 0 }}
              onSubmit={handleAddModule}
              className="mb-4 bg-[#131315] p-3 rounded-xl border border-[#C5A267]/20 space-y-2.5"
            >
              <input
                type="text"
                required
                placeholder="Nama Modul Baru..."
                value={newModuleName}
                onChange={(e) => setNewModuleName(e.target.value)}
                className="w-full text-xs bg-[#1A1A1D] border border-white/10 rounded-lg p-2 text-white focus:ring-1 focus:ring-[#C5A267] focus:outline-none"
              />
              <textarea
                placeholder="Deskripsi singkat..."
                value={newModuleDesc}
                onChange={(e) => setNewModuleDesc(e.target.value)}
                className="w-full text-xs bg-[#1A1A1D] border border-white/10 rounded-lg p-2 text-white focus:ring-1 focus:ring-[#C5A267] focus:outline-none h-12 resize-none"
              />
              <div className="flex justify-end space-x-1.5 pt-1">
                <button
                  type="button"
                  onClick={() => setIsAdding(false)}
                  className="text-[10px] uppercase font-mono tracking-wider text-gray-400 hover:text-white px-2 py-1 rounded"
                >
                  Batal
                </button>
                <button
                  type="submit"
                  className="text-[10px] bg-[#C5A267] text-[#0A0A0B] font-bold uppercase tracking-wider px-2.5 py-1 rounded shadow cursor-pointer hover:bg-[#B38F52] transition-colors"
                >
                  Simpan
                </button>
              </div>
            </motion.form>
          )}

          <div className="space-y-1">
            {modules.map((m) => {
              const isActive = m.id === activeId;
              const nodeCount = m.nodes ? m.nodes.length : 0;
              return (
                <div
                  key={m.id}
                  onClick={() => selectModule(m.id)}
                  className={`w-full flex items-center justify-between text-left px-3 py-2.5 rounded-lg text-sm transition-all relative group cursor-pointer ${
                    isActive
                      ? 'bg-white/5 border border-white/10 text-white font-medium border-l-4 border-[#C5A267] shadow-[0_0_12px_rgba(197,162,103,0.15)]'
                      : 'text-gray-400 hover:bg-white/5 hover:text-white'
                  }`}
                >
                  <div className="flex items-center space-x-2 overflow-hidden">
                    <Folder className={`w-4 h-4 flex-shrink-0 ${isActive ? 'text-[#C5A267]' : 'text-gray-500'}`} />
                    {!sidebarCollapsed && <span className="truncate pr-1 text-xs">{m.name}</span>}
                  </div>

                  {!sidebarCollapsed && (
                    <div className="flex items-center space-x-1.5 flex-shrink-0">
                      <span className={`text-[10px] font-mono px-1.5 py-0.5 rounded-full ${
                        isActive ? 'bg-[#C5A267]/15 text-[#C5A267] border border-[#C5A267]/25' : 'bg-[#1A1A1D] text-gray-500 border border-white/5'
                      }`}>
                        {nodeCount}
                      </span>
                      {modules.length > 1 && (
                        <button
                          onClick={(e) => handleDeleteClick(m.id, m.name, e)}
                          className="opacity-0 group-hover:opacity-100 text-gray-500 hover:text-red-400 p-0.5 rounded hover:bg-white/5 transition"
                          title="Hapus Modul"
                        >
                          <Trash2 className="w-3.5 h-3.5" />
                        </button>
                      )}
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        </div>

        {/* Team Members Widget */}
        {!sidebarCollapsed && (
          <div className="pt-4 border-t border-white/5">
            <span className="text-xs font-semibold text-gray-400 uppercase tracking-widest px-2 mb-2 block flex items-center">
              <Users className="w-3.5 h-3.5 mr-1 text-[#C5A267]" />
              Anggota Tim (Online)
            </span>
            <div className="space-y-2 px-2 pt-1.5">
              {teamMembers.map((member) => (
                <div key={member.id} className="flex items-center justify-between text-xs text-gray-400">
                  <div className="flex items-center space-x-2">
                    <div className="w-6 h-6 rounded-full flex items-center justify-center text-[10px] font-bold text-white uppercase bg-gradient-to-tr from-gray-700 to-gray-500">
                      {member.name.substring(0, 2)}
                    </div>
                    <div className="text-left">
                      <p className="font-semibold text-gray-300 text-[11px] leading-tight">{member.name}</p>
                      <p className="text-[9px] text-gray-500 uppercase tracking-wider">{member.role}</p>
                    </div>
                  </div>
                  {/* Active gold dot helper */}
                  <span className="flex h-2 w-2 relative">
                    <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-[#C5A267] opacity-75"></span>
                    <span className="relative inline-flex rounded-full h-2 w-2 bg-[#C5A267]"></span>
                  </span>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* Sync Status / Info indicator widget above footer */}
      {!sidebarCollapsed && (
        <div className="px-3 pb-2 pt-1">
          <div className="p-3.5 bg-[#1A1A1C] border border-white/5 rounded-xl text-left">
            <p className="text-[9px] uppercase tracking-widest text-[#C5A267] mb-2 font-mono font-bold">Sync Status</p>
            <div className="flex items-center gap-2">
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-500"></span>
              <span className="text-[11px] text-gray-300">Real-time Active</span>
            </div>
          </div>
        </div>
      )}

      {/* Footer view */}
      <div className="p-4 border-t border-white/5 bg-[#0D0D0F] flex items-center justify-between">
        {!sidebarCollapsed && (
          <div className="text-[10px] text-gray-500 font-mono flex items-center space-x-1 uppercase tracking-wider">
            <span className="text-[#C5A267] font-bold">•</span>
            <span>Versi SPA 1.0</span>
          </div>
        )}
        <button
          onClick={toggleDarkMode}
          className={`flex items-center justify-center p-2 rounded-lg text-gray-400 hover:text-white hover:bg-white/5 cursor-pointer transition ${
            sidebarCollapsed ? 'mx-auto' : ''
          }`}
          title={darkMode ? 'Ganti ke Mode Terang' : 'Ganti ke Mode Gelap'}
        >
          {darkMode ? <Sun className="w-4 h-4 text-[#C5A267] animate-spin-slow" /> : <Moon className="w-4 h-4" />}
        </button>
      </div>
    </div>
  );
}
