/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useRef, useEffect } from 'react';
import { useStore, AppView, NotificationItem } from '../store/useStore';
import {
  Map,
  Grid,
  FileText,
  Calendar as CalIcon,
  TrendingUp,
  Download,
  Bell,
  CheckCircle,
  FileSpreadsheet,
  FileCode,
  FileOutput,
  Edit2,
  Trash2,
  Check,
  X,
  Kanban,
  Users
} from 'lucide-react';
import { exportToJson, exportToMarkdown, exportToOpenApi, exportToCsv } from '../services/exportService';
import { motion, AnimatePresence } from 'motion/react';

export default function Topbar() {
  const {
    modules,
    activeId,
    view,
    setView,
    renameModule,
    notifications,
    markAllNotificationsRead,
    clearNotifications,
    addNotification,
  } = useStore();

  const activeModule = modules.find((m) => m.id === activeId);
  const unreadCount = notifications.filter((n) => !n.read).length;

  const [isEditingName, setIsEditingName] = useState(false);
  const [editedName, setEditedName] = useState('');
  const [editedDesc, setEditedDesc] = useState('');
  
  const [showNotifMenu, setShowNotifMenu] = useState(false);
  const [showExportMenu, setShowExportMenu] = useState(false);

  const notifRef = useRef<HTMLDivElement>(null);
  const exportRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (activeModule) {
      setEditedName(activeModule.name);
      setEditedDesc(activeModule.description || '');
    }
  }, [activeId, activeModule]);

  // Handle outside clicks to close menus
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (notifRef.current && !notifRef.current.contains(event.target as Node)) {
        setShowNotifMenu(false);
      }
      if (exportRef.current && !exportRef.current.contains(event.target as Node)) {
        setShowExportMenu(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  if (!activeModule) {
    return (
      <div className="h-16 border-b flex items-center px-6 bg-white dark:bg-slate-900 border-slate-200 dark:border-slate-800 text-slate-500">
        Pilih atau buat modul untuk memulai...
      </div>
    );
  }

  const handleSaveRename = () => {
    if (editedName.trim()) {
      renameModule(activeModule.id, editedName.trim(), editedDesc.trim());
      setIsEditingName(false);
      addNotification(
        'Modul Diperbarui',
        `Nama modul berhasil diubah menjadi "${editedName}".`,
        'success'
      );
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSaveRename();
    } else if (e.key === 'Escape') {
      setIsEditingName(false);
      setEditedName(activeModule.name);
      setEditedDesc(activeModule.description || '');
    }
  };

  const views: { id: AppView; label: string; icon: any }[] = [
    { id: 'canvas', label: 'Kanvas', icon: Map },
    { id: 'status', label: 'Tabel Status', icon: Grid },
    { id: 'doc', label: 'Spesifikasi PDF/Dokumen', icon: FileText },
    { id: 'calendar', label: 'Jadwal Tim', icon: CalIcon },
    { id: 'analytics', label: 'Analitik', icon: TrendingUp },
    { id: 'kanban', label: 'Kanban', icon: Kanban },
    { id: 'team', label: 'Anggota Tim', icon: Users },
  ];

  return (
    <div className="h-16 px-6 bg-[#111113]/90 backdrop-blur-md border-b border-white/5 flex items-center justify-between z-10 select-none shadow-[0_1px_10px_rgba(0,0,0,0.3)]">
      
      {/* Module Title Section */}
      <div className="flex-1 max-w-sm mr-4">
        {isEditingName ? (
          <div className="flex items-center space-x-2 bg-[#131315] p-2 rounded-lg border border-white/10">
            <div className="flex-col flex-1">
              <input
                type="text"
                value={editedName}
                onChange={(e) => setEditedName(e.target.value)}
                onKeyDown={handleKeyPress}
                className="text-xs font-bold w-full bg-transparent text-white border-none outline-none focus:ring-0 p-0"
                placeholder="Nama Modul"
                autoFocus
              />
              <input
                type="text"
                value={editedDesc}
                onChange={(e) => setEditedDesc(e.target.value)}
                onKeyDown={handleKeyPress}
                className="text-[10px] text-gray-400 w-full bg-transparent border-none outline-none focus:ring-0 p-0 mt-0.5"
                placeholder="Tambah Deskripsi singkat..."
              />
            </div>
            <button
              onClick={handleSaveRename}
              className="p-1 rounded bg-[#C5A267] text-black hover:bg-[#B38F52] cursor-pointer transition-colors"
            >
              <Check className="w-3.5 h-3.5" />
            </button>
            <button
              onClick={() => setIsEditingName(false)}
              className="p-1 rounded bg-white/5 text-gray-300 hover:bg-white/10 cursor-pointer transition-colors"
            >
              <X className="w-3.5 h-3.5" />
            </button>
          </div>
        ) : (
          <div className="group cursor-pointer flex items-start space-x-2" onClick={() => setIsEditingName(true)}>
            <div className="text-left">
              <div className="flex items-center space-x-1.5">
                <h2 className="text-sm font-bold text-white tracking-tight font-sans">
                  {activeModule.name}
                </h2>
                <Edit2 className="w-3 h-3 text-[#C5A267] opacity-0 group-hover:opacity-100 transition-opacity" />
              </div>
              <p className="text-[11px] text-gray-400 truncate max-w-xs block mt-0.5">
                {activeModule.description || 'Tidak ada deskripsi modul (Klik untuk menambahkan)'}
              </p>
            </div>
          </div>
        )}
      </div>

      {/* Center Section: App View Switcher */}
      <div className="flex space-x-0.5 bg-[#131315] p-1.5 rounded-xl border border-white/5">
        {views.map((v) => {
          const isActive = view === v.id;
          const Icon = v.icon;
          return (
            <button
              key={v.id}
              onClick={() => setView(v.id)}
              className={`flex items-center space-x-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold tracking-wide transition-all duration-200 cursor-pointer ${
                isActive
                  ? 'bg-[#C5A267]/10 text-[#C5A267] border border-[#C5A267]/20 shadow-[0_0_10px_rgba(197,162,103,0.1)]'
                  : 'text-gray-400 hover:text-white hover:bg-[#1A1A1D]'
              }`}
            >
              <Icon className={`w-3.5 h-3.5 ${isActive ? 'scale-110 text-[#C5A267]' : ''}`} />
              <span className="hidden md:inline">{v.label}</span>
            </button>
          );
        })}
      </div>

      {/* Right Section: Exports & Notifications */}
      <div className="flex items-center space-x-3 ml-4">
        
        {/* Export Dropdown Menu */}
        <div className="relative" ref={exportRef}>
          <button
            onClick={() => setShowExportMenu(!showExportMenu)}
            className="flex items-center space-x-1.5 px-3.5 py-2 rounded-xl text-xs font-bold text-gray-200 bg-[#131315] border border-white/10 hover:border-[#C5A267]/40 hover:bg-white/5 cursor-pointer shadow transition-all"
          >
            <Download className="w-3.5 h-3.5 text-[#C5A267] animate-bounce-slow" />
            <span>Ekspor Data</span>
          </button>

          <AnimatePresence>
            {showExportMenu && (
              <motion.div
                initial={{ opacity: 0, y: 10, scale: 0.95 }}
                animate={{ opacity: 1, y: 0, scale: 1 }}
                exit={{ opacity: 0, y: 10, scale: 0.95 }}
                className="absolute right-0 mt-2.5 w-64 rounded-xl shadow-2xl border bg-[#131315] border-white/10 z-50 p-1.5 divide-y divide-white/5"
              >
                <div className="p-2">
                  <p className="text-[10px] uppercase font-bold text-[#C5A267] font-mono tracking-wider mb-1.5 px-1.5">Dokumen & Laporan</p>
                  
                  <button
                    onClick={() => {
                      exportToMarkdown(activeModule);
                      setShowExportMenu(false);
                    }}
                    className="w-full text-left flex items-center space-x-2.5 px-3 py-2 text-xs text-gray-300 hover:bg-white/5 rounded-lg cursor-pointer transition text-left"
                  >
                    <FileText className="w-3.5 h-3.5 text-blue-400 flex-shrink-0" />
                    <div>
                      <p className="font-semibold text-[11px] text-gray-200">Ekspor Laporan (Markdown/PDF)</p>
                      <p className="text-[9px] text-gray-500">Siap salin / cetak ke PDF</p>
                    </div>
                  </button>

                  <button
                    onClick={() => {
                       exportToCsv(activeModule);
                       setShowExportMenu(false);
                    }}
                    className="w-full text-left flex items-center space-x-2.5 px-3 py-2 text-xs text-gray-300 hover:bg-white/5 rounded-lg cursor-pointer transition mt-0.5 text-left"
                  >
                    <FileSpreadsheet className="w-3.5 h-3.5 text-emerald-400 flex-shrink-0" />
                    <div>
                      <p className="font-semibold text-[11px] text-gray-200">Ekspor Laporan (Excel/CSV)</p>
                      <p className="text-[9px] text-gray-500">Sesuai format spreadsheet</p>
                    </div>
                  </button>
                </div>

                <div className="p-2">
                  <p className="text-[10px] uppercase font-bold text-[#C5A267] font-mono tracking-wider mb-1.5 px-1.5 mt-1">Struktur Developer</p>
                  
                  <button
                    onClick={() => {
                      exportToJson(activeModule);
                      setShowExportMenu(false);
                    }}
                    className="w-full text-left flex items-center space-x-2.5 px-3 py-2 text-xs text-gray-300 hover:bg-white/5 rounded-lg cursor-pointer transition text-left"
                  >
                    <FileCode className="w-3.5 h-3.5 text-purple-400 flex-shrink-0" />
                    <div>
                      <p className="font-semibold text-[11px] text-gray-200">JSON Model Kanonik</p>
                      <p className="text-[9px] text-gray-500">Model data portabel lengkap</p>
                    </div>
                  </button>

                  <button
                    onClick={() => {
                      exportToOpenApi(activeModule);
                      setShowExportMenu(false);
                    }}
                    className="w-full text-left flex items-center space-x-2.5 px-3 py-2 text-xs text-gray-300 hover:bg-white/5 rounded-lg cursor-pointer transition mt-0.5 text-left"
                  >
                    <FileOutput className="w-3.5 h-3.5 text-amber-500 flex-shrink-0" />
                    <div>
                      <p className="font-semibold text-[11px] text-gray-200">Kontrak OpenAPI 3.1</p>
                      <p className="text-[9px] text-gray-500">File json siap pakai Swagger</p>
                    </div>
                  </button>
                </div>
              </motion.div>
            )}
          </AnimatePresence>
        </div>

        {/* Real-time Notifications Bell */}
        <div className="relative" ref={notifRef}>
          <button
            onClick={() => {
              setShowNotifMenu(!showNotifMenu);
              if (!showNotifMenu) markAllNotificationsRead();
            }}
            className="p-2 rounded-xl border bg-[#131315] hover:bg-white/5 border-white/10 text-gray-300 cursor-pointer shadow transition relative"
          >
            <Bell className="w-4 h-4" />
            {unreadCount > 0 && (
              <span className="absolute -top-1 -right-1 bg-[#C5A267] text-black text-[9px] font-bold px-1.5 py-0.5 rounded-full flex items-center justify-center animate-bounce shadow-[0_0_8px_#C5A267]">
                {unreadCount}
              </span>
            )}
          </button>

          <AnimatePresence>
            {showNotifMenu && (
              <motion.div
                initial={{ opacity: 0, y: 10, scale: 0.95 }}
                animate={{ opacity: 1, y: 0, scale: 1 }}
                exit={{ opacity: 0, y: 10, scale: 0.95 }}
                className="absolute right-0 mt-2.5 w-80 rounded-xl shadow-2xl border bg-[#131315] border-white/10 z-50 p-2"
              >
                <div className="flex items-center justify-between p-2 pb-2.5 border-b border-white/5">
                  <h3 className="text-[10px] font-bold text-gray-200 uppercase tracking-widest font-sans flex items-center">
                    <Bell className="w-3.5 h-3.5 text-[#C5A267] mr-1.5" />
                    Koordinasi Tim
                  </h3>
                  <div className="flex items-center space-x-2">
                    <button
                      onClick={clearNotifications}
                      className="text-[9px] text-gray-400 hover:text-red-400 cursor-pointer uppercase tracking-wider font-mono"
                    >
                      Hapus
                    </button>
                    <span className="text-white/10">|</span>
                    <button
                      onClick={() => {
                        addNotification(
                          'Simulasi Koordinasi',
                          'Rian mengubah status UI/UX langkah Persetujuan Manager menjadi DONE!',
                          'info'
                        );
                      }}
                      className="text-[9px] text-[#C5A267] hover:text-[#E2C392] cursor-pointer uppercase tracking-wider font-mono font-bold"
                    >
                      Picu
                    </button>
                  </div>
                </div>

                <div className="max-h-72 overflow-y-auto pt-1 space-y-1.5 scrollbar-thin">
                  {notifications.length === 0 ? (
                    <div className="p-6 text-center text-xs text-gray-500">
                      Tidak ada koordinasi baru saat ini.
                    </div>
                  ) : (
                    notifications.map((n) => (
                      <div
                        key={n.id}
                        className={`p-2.5 rounded-lg text-xs transition border text-left ${
                          n.type === 'success'
                            ? 'bg-emerald-950/10 border-emerald-900/40 text-emerald-300'
                            : n.type === 'warning'
                            ? 'bg-rose-950/10 border-rose-900/40 text-rose-300'
                            : 'bg-white/5 border-white/5 text-gray-300'
                        }`}
                      >
                        <div className="flex items-center justify-between font-bold mb-1">
                          <span className="font-semibold text-[11px]">{n.title}</span>
                          <span className="text-[8px] font-mono opacity-50 font-normal">{n.timestamp}</span>
                        </div>
                        <p className="text-[10px] leading-relaxed opacity-80">{n.message}</p>
                      </div>
                    ))
                  )}
                </div>
              </motion.div>
            )}
          </AnimatePresence>
        </div>

      </div>
    </div>
  );
}
