/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from 'react';
import { useStore } from '../../store/useStore';
import { NodeType } from '../../domain/types';
import { NODE_TYPES } from '../../config/nodeTypes';
import NodeCard from './NodeCard';
import EdgeLayer from './EdgeLayer';
import { Play, Cpu, GitFork, User, Database, Plus, Trash2, ArrowRight, MousePointerClick, X, Sparkles, Loader2, ZoomIn, ZoomOut, Maximize2, LayoutGrid, Check, Eye } from 'lucide-react';
import { motion, AnimatePresence } from 'motion/react';

export default function Canvas() {
  const {
    modules,
    activeId,
    addNode,
    selectNode,
    connectFrom,
    setConnectFrom,
    loadAiGeneratedFlow,
    addNotification,
    token,
  } = useStore();

  const activeModule = modules.find((m) => m.id === activeId);

  const [newNodeType, setNewNodeType] = useState<NodeType | null>(null);
  const [newNodeLabel, setNewNodeTypeLabel] = useState('');

  // AI Flow Generator states
  const [aiExpanded, setAiExpanded] = useState(false);
  const [aiPrompt, setAiPrompt] = useState('');
  const [aiGenerating, setAiGenerating] = useState(false);
  const [aiStatusStep, setAiStatusStep] = useState('');

  // Zoom factor state
  const [zoom, setZoom] = useState(1);

  const handleCanvasClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      selectNode(null);
      setConnectFrom(null);
    }
  };

  const handleAddNodeClicked = (type: NodeType) => {
    setNewNodeType(type);
    setNewNodeTypeLabel('');
  };

  const submitNewNode = (e: React.FormEvent) => {
    e.preventDefault();
    if (newNodeType && newNodeLabel.trim()) {
      addNode(newNodeType, newNodeLabel.trim());
      setNewNodeType(null);
      setNewNodeTypeLabel('');
    }
  };

  // AI Flow Generator trigger
  const handleGenerateAiFlow = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!aiPrompt.trim()) return;

    setAiGenerating(true);
    setAiStatusStep('Menganalisis proses bisnis modul...');

    try {
      // Simulate stepped progress messages for a beautiful premium experience
      setTimeout(() => setAiStatusStep('Merumuskan rancangan arsitektur per babak...'), 2000);
      setTimeout(() => setAiStatusStep('Menghitung koordinat letak kanvas terkoordinasi...'), 4000);
      setTimeout(() => setAiStatusStep('Pemetaan kontak API & komponen Frontend...'), 6000);

      const res = await fetch('/api/ai/generate-flow', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({ prompt: aiPrompt }),
      });

      const data = await res.json();
      if (!res.ok || data.error) {
        throw new Error(data.error || 'Terjadi kesalahan eksternal');
      }

      // Success loading
      await loadAiGeneratedFlow(data.name, data.description, data.nodes, data.edges);
      
      // Cleanup
      setAiPrompt('');
      setAiExpanded(false);
      addNotification('AI Flow Berhasil', `Modul baru "${data.name}" siap digunakan!`, 'success');
    } catch (err: any) {
      console.error(err);
      addNotification('AI Gagal Terkoneksi', `Memakai Demo/Fallback lokal: ${err.message}`, 'warning');
    } finally {
      setAiGenerating(false);
      setAiStatusStep('');
    }
  };

  // Auto-Align Arrangement Algorithm
  const handleAutoAlign = () => {
    if (!activeModule || !activeModule.nodes || activeModule.nodes.length === 0) return;

    // Arrange nodes nicely in a grid or horizontal tree sequence
    const { modules } = useStore.getState();
    const updatedNodes = [...activeModule.nodes];

    // Simple sorting: sort by id, type or simple heuristics
    updatedNodes.sort((a, b) => a.id.localeCompare(b.id));

    // Arrange horizontally with wrapping
    const spacingX = 350;
    const spacingY = 250;
    const cols = 3;

    const aligned = activeModule.nodes.map((node, index) => {
      const col = index % cols;
      const row = Math.floor(index / cols);
      return {
        ...node,
        x: 100 + col * spacingX,
        y: 150 + row * spacingY,
      };
    });

    useStore.setState({
      modules: modules.map((m) =>
        m.id === activeId ? { ...m, nodes: aligned } : m
      ),
    });

    addNotification('Tata Letak Rapi', 'Kanvas proses bisnis berhasil disejajarkan secara ortogonal.', 'success');
  };

  if (!activeModule) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center p-8 bg-[#0A0A0B] text-gray-500">
        <p>Pilih atau buat modul di sidebar untuk memulai perancangan alur kerja.</p>
      </div>
    );
  }

  const paletteItems: { type: NodeType; icon: any; color: string }[] = [
    { type: 'terminator', icon: Play, color: 'text-red-400 bg-red-950/20 border border-red-950/40 hover:bg-red-950/40' },
    { type: 'process', icon: Cpu, color: 'text-blue-400 bg-blue-950/20 border border-blue-950/40 hover:bg-blue-950/40' },
    { type: 'decision', icon: GitFork, color: 'text-amber-400 bg-amber-950/20 border border-amber-950/40 hover:bg-amber-950/40' },
    { type: 'actor', icon: User, color: 'text-emerald-400 bg-emerald-950/20 border border-emerald-950/40 hover:bg-emerald-950/40' },
    { type: 'system', icon: Database, color: 'text-purple-400 bg-purple-950/20 border border-purple-950/40 hover:bg-purple-950/40' },
  ];

  const hasNodes = activeModule.nodes && activeModule.nodes.length > 0;

  return (
    <div className="flex-1 flex flex-col relative select-none overflow-hidden h-full bg-[#0A0A0B]">
      
      {/* Floating Canvas Palette with AI integration */}
      <div className="absolute top-4 left-6 z-30 flex items-center bg-[#131315]/95 backdrop-blur-md rounded-2xl p-2.5 shadow-2xl border border-white/10 flex-wrap gap-2.5 max-w-full">
        <div className="px-2.5 pr-3 py-1 border-r border-white/5">
          <span className="text-[10px] font-bold text-[#C5A267] font-mono tracking-widest block uppercase">Palette Alur</span>
          <span className="text-[9px] text-gray-400 font-semibold block leading-tight mt-0.5">Klik tipe & tambah</span>
        </div>

        <div className="flex items-center space-x-1.5 flex-wrap">
          {paletteItems.map((item) => {
            const config = NODE_TYPES[item.type];
            const Icon = item.icon;
            return (
              <button
                key={item.type}
                onClick={() => handleAddNodeClicked(item.type)}
                className={`flex items-center space-x-1.5 px-3 py-1.5 rounded-xl text-xs font-bold transition-all hover:scale-103 cursor-pointer ${item.color} shadow-sm`}
                title={`Tambah Langkah ${config.label}`}
              >
                <Icon className="w-3 h-3" />
                <span>{config.label}</span>
              </button>
            );
          })}
        </div>

        <div className="border-l border-white/5 pl-2.5 flex items-center space-x-1.5">
          <button
            onClick={() => setAiExpanded(!aiExpanded)}
            className="flex items-center space-x-1.5 px-3 py-1.5 rounded-xl text-xs font-bold bg-[#C5A267] text-black hover:bg-[#B59155] cursor-pointer shadow transition-all duration-150"
          >
            <Sparkles className="w-3 h-3 text-black animate-pulse" />
            <span>Rancang via AI</span>
          </button>

          {hasNodes && (
            <button
              onClick={handleAutoAlign}
              className="flex items-center space-x-1.5 px-3 py-1.5 rounded-xl border border-white/15 bg-white/5 hover:bg-white/10 text-white text-xs font-bold cursor-pointer transition"
              title="Rapikan tata letak semua langkah"
            >
              <LayoutGrid className="w-3.5 h-3.5" />
              <span>Rapikan</span>
            </button>
          )}
        </div>
      </div>

      {/* Slide-out AI Panel Drawer */}
      <AnimatePresence>
        {aiExpanded && (
          <motion.div
            initial={{ opacity: 0, y: -20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            className="absolute top-20 left-6 z-40 bg-[#131315]/95 backdrop-blur-md rounded-2xl p-4 shadow-2xl border border-[#C5A267]/25 w-96 text-left"
          >
            <div className="flex items-center justify-between mb-2">
              <span className="text-xs font-bold text-white uppercase tracking-widest flex items-center">
                <Sparkles className="w-4 h-4 text-[#C5A267] mr-1.5" /> Smart AI Flow Builder
              </span>
              <button
                onClick={() => setAiExpanded(false)}
                className="text-gray-400 hover:text-white p-1 rounded-full hover:bg-white/5 cursor-pointer"
              >
                <X className="w-3.5 h-3.5" />
              </button>
            </div>
            <p className="text-[11px] text-gray-400 mb-3 leading-relaxed">
              Tulis instruksi alur fungsional aplikasi (contoh: <em>"Buat sistem registrasi pelapor bencana dengan unggahan foto dan verifikasi OTP email"</em>).
            </p>

            <form onSubmit={handleGenerateAiFlow} className="space-y-3">
              <textarea
                value={aiPrompt}
                onChange={(e) => setAiPrompt(e.target.value)}
                placeholder="Buat modul pengajuan cuti karyawan dengan..."
                rows={3}
                required
                className="w-full text-xs font-medium border border-white/10 bg-[#1A1A1D] text-white rounded-xl p-3 focus:outline-none focus:ring-1 focus:ring-[#C5A267] resize-none"
              />

              <button
                type="submit"
                disabled={aiGenerating || !aiPrompt.trim()}
                className="w-full flex items-center justify-center space-x-1.5 py-2.5 rounded-xl bg-[#C5A267] hover:bg-[#B59155] disabled:bg-white/5 disabled:text-gray-500 text-black text-xs font-bold font-mono tracking-wider transition cursor-pointer"
              >
                {aiGenerating ? (
                  <>
                    <Loader2 className="w-3.5 h-3.5 animate-spin" />
                    <span>MENGHASILKAN ALUR...</span>
                  </>
                ) : (
                  <>
                    <Sparkles className="w-3.5 h-3.5 text-black" />
                    <span>BUAT MODEL PROSES</span>
                  </>
                )}
              </button>
            </form>

            {aiGenerating && (
              <div className="mt-3 bg-white/3 border border-white/5 rounded-xl p-3 flex items-center space-x-2">
                <Loader2 className="w-4 h-4 text-[#C5A267] animate-spin flex-shrink-0" />
                <div className="text-[10px] font-mono text-gray-300">
                  <p className="font-bold uppercase tracking-wider text-white">Status AI</p>
                  <p className="mt-0.5">{aiStatusStep}</p>
                </div>
              </div>
            )}
          </motion.div>
        )}
      </AnimatePresence>

      {/* Zoom HUD Controls bottom corner */}
      <div className="absolute bottom-4 right-6 z-30 flex items-center space-x-1.5 bg-[#131315]/90 backdrop-blur-md border border-white/10 p-2 rounded-xl shadow-xl">
        <button
          onClick={() => setZoom(Math.max(0.5, zoom - 0.1))}
          className="p-1.5 border border-white/10 hover:bg-white/5 rounded-lg text-gray-400 hover:text-white cursor-pointer transition"
          title="Zoom Out (Perkecil) -10%"
        >
          <ZoomOut className="w-3.5 h-3.5" />
        </button>
        <span className="text-[10px] font-mono text-gray-300 font-bold px-2 inline-block min-w-12 text-center select-none">
          {Math.round(zoom * 100)}%
        </span>
        <button
          onClick={() => setZoom(Math.min(1.5, zoom + 0.1))}
          className="p-1.5 border border-white/10 hover:bg-white/5 rounded-lg text-gray-400 hover:text-white cursor-pointer transition"
          title="Zoom In (Perbesar) +10%"
        >
          <ZoomIn className="w-3.5 h-3.5" />
        </button>
        <button
          onClick={() => setZoom(1)}
          className="p-1.5 border border-white/10 hover:bg-white/5 rounded-lg text-gray-400 hover:text-white cursor-pointer transition"
          title="Reset Zoom Ke 100%"
        >
          <Maximize2 className="w-3.5 h-3.5" />
        </button>
      </div>

      {/* Connection Mode Helper Alert */}
      {connectFrom !== null && (
        <div className="absolute top-4 right-6 z-30 flex items-center space-x-3 bg-[#C5A267] text-black rounded-2xl px-4 py-2.5 shadow-2xl border border-white/10 animate-pulse">
          <MousePointerClick className="w-4 h-4 text-black animate-ping" />
          <div className="text-xs text-left">
            <p className="font-bold uppercase tracking-wider text-[10px]">Mode Pembuatan Koneksi Aktif</p>
            <p className="text-[10px] opacity-90">Klik langkah target tujuan untuk menyambung (Atau tekan tombol ESC/klik kanvas kosong untuk membatalkan).</p>
          </div>
          <button
            onClick={() => setConnectFrom(null)}
            className="text-black hover:bg-black/10 p-1 rounded-full cursor-pointer transition-colors"
          >
            <X className="w-3.5 h-3.5" />
          </button>
        </div>
      )}

      {/* Main Graph Workbench Sandbox */}
      <div
        onClick={handleCanvasClick}
        style={{
          backgroundImage: 'radial-gradient(rgba(197, 162, 103, 0.15) 1.2px, transparent 1.2px)',
          backgroundSize: '24px 24px',
          width: '100%',
          height: '100%',
        }}
        className="flex-1 overflow-auto bg-[#0A0A0B] p-12 relative flex-shrink-0 cursor-crosshair scrollbar-thin transition-colors"
      >
        <div
          className="relative transition-all duration-200"
          style={{
            width: 1600,
            height: 1100,
            transform: `scale(${zoom})`,
            transformOrigin: 'top left',
          }}
          onClick={handleCanvasClick}
        >
          {/* Back Edge layer (draw lines first) */}
          <EdgeLayer />

          {/* Sinks the Node cards */}
          {activeModule.nodes.map((node) => (
            <NodeCard key={node.id} node={node} />
          ))}

          {/* Empty Canvas Friendly State */}
          {!hasNodes && (
            <div className="absolute inset-x-0 top-40 mx-auto max-w-sm text-center p-8 bg-[#131315] border border-white/10 rounded-3xl shadow-2xl pointer-events-none">
              <Plus className="w-10 h-10 text-[#C5A267] mx-auto mb-4 animate-bounce" />
              <h4 className="text-sm font-bold text-white mb-1.5" style={{ fontFamily: 'Georgia, serif' }}>Mulai Merancang Alur Kerja</h4>
              <p className="text-[11px] text-gray-400 leading-relaxed">
                Kanvas modul ini masih kosong. Silakan gunakan palette diatas untuk menambah langkah alur kerja pertama Anda, atau ketik permintaan Anda di tombol <strong>Smart AI Rancang via AI</strong> di atas untuk merumuskannya secara instan!
              </p>
            </div>
          )}
        </div>
      </div>

      {/* Add Node Popup Input (Inline Dialog Modul) */}
      <AnimatePresence>
        {newNodeType && (
          <div className="absolute inset-0 bg-[#0A0A0B]/80 backdrop-blur-sm flex items-center justify-center z-50 p-4">
            <motion.div
              initial={{ scale: 0.9, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              exit={{ scale: 0.9, opacity: 0 }}
              className="bg-[#131315] p-6 rounded-2xl max-w-sm w-full shadow-2xl border border-white/10 text-left"
            >
              <h3 className="text-sm font-bold text-[#C5A267] mb-3 uppercase tracking-wider font-sans">
                Langkah Baru: {NODE_TYPES[newNodeType]?.label}
              </h3>
              <form onSubmit={submitNewNode} className="space-y-4">
                <div>
                  <label className="block text-[10px] font-semibold text-gray-400 mb-1.5 font-mono uppercase tracking-widest">LABEL / NAMA LANGKAH</label>
                  <input
                    type="text"
                    required
                    placeholder="Contoh: Buat Pesanan, Cek Saldo..."
                    value={newNodeLabel}
                    onChange={(e) => setNewNodeTypeLabel(e.target.value)}
                    className="w-full text-xs font-semibold border border-white/10 bg-[#1A1A1D] text-white rounded-xl p-3 focus:outline-none focus:ring-1 focus:focus:ring-[#C5A267]"
                    autoFocus
                  />
                </div>

                <div className="flex space-x-2 justify-end pt-1">
                  <button
                    type="button"
                    onClick={() => setNewNodeType(null)}
                    className="text-[10px] font-mono font-bold tracking-widest text-gray-400 hover:text-white px-4 py-2.5 rounded-xl transition cursor-pointer hover:bg-white/5"
                  >
                    BATAL
                  </button>
                  <button
                    type="submit"
                    className="text-[10px] font-mono font-bold tracking-widest bg-[#C5A267] hover:bg-[#B38F52] text-black px-4 py-2.5 rounded-xl shadow transition cursor-pointer flex items-center space-x-1.5"
                  >
                    <span>TAMBAHKAN</span>
                    <ArrowRight className="w-3.5 h-3.5" />
                  </button>
                </div>
              </form>
            </motion.div>
          </div>
        )}
      </AnimatePresence>
    </div>
  );
}
