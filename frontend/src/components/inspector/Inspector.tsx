/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from 'react';
import { useStore } from '../../store/useStore';
import { NodeType, RoleKey } from '../../domain/types';
import { NODE_TYPES } from '../../config/nodeTypes';
import { ROLES } from '../../config/roles';
import { STATUS_CONFIG } from '../../config/status';
import BisnisTab from './BisnisTab';
import UiuxTab from './UiuxTab';
import FrontendTab from './FrontendTab';
import BackendTab from './BackendTab';
import { X, Trash2, ArrowUpDown, ChevronRight, Link2 } from 'lucide-react';
import { motion } from 'motion/react';

export default function Inspector() {
  const {
    modules,
    activeId,
    selectedNodeId,
    selectNode,
    updateNode,
    deleteNode,
    setConnectFrom,
  } = useStore();

  const activeModule = modules.find((m) => m.id === activeId);
  const node = activeModule?.nodes.find((n) => n.id === selectedNodeId);

  const [activeTab, setActiveTab] = useState<'bisnis' | 'uiux' | 'frontend' | 'backend'>('bisnis');

  if (!activeModule || !node) {
    return null; // Don't show if nothing is selected
  }

  const handleLabelChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    updateNode(node.id, { label: e.target.value });
  };

  const handleTypeChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    updateNode(node.id, { type: e.target.value as NodeType });
  };

  const handleDeleteNode = () => {
    const confirmed = window.confirm(`Apakah Anda yakin ingin menghapus langkah "${node.label}"? Semua sambungan alur kerja yang menyentuh langkah ini juga akan terhapus.`);
    if (confirmed) {
      deleteNode(node.id);
    }
  };

  const handleConnectInitiate = () => {
    setConnectFrom(node.id);
  };

  // Check if facet is filled to display highlight dot marker
  const isFacetFilled = (type: 'uiux' | 'frontend' | 'backend'): boolean => {
    if (type === 'uiux') {
      const f = node.roles.uiux;
      return !!(f && (f.assignee || f.status || f.screen || f.link));
    }
    if (type === 'frontend') {
      const f = node.roles.frontend;
      const legacy = f as typeof f & { component?: string; link?: string };
      return !!(f && (f.assignee || f.status || f.page || legacy?.component || f.route || f.interaction || f.validation || f.state || f.handoffLink || legacy?.link));
    }
    if (type === 'backend') {
      const f = node.roles.backend;
      return !!(f && (f.assignee || f.status || f.endpoint || f.request || f.response));
    }
    return false;
  };

  return (
    <div className="w-96 h-screen border-l bg-[#111113] border-white/5 flex flex-col flex-shrink-0 z-10 shadow-2xl relative">
      
      {/* Inspector Header */}
      <div className="p-4 border-b border-white/5 flex items-start justify-between">
        <div className="space-y-2.5 flex-1 mr-3 text-left">
          <div className="flex items-center space-x-2">
            <span className="text-[10px] font-mono font-bold px-2 py-0.5 rounded-full bg-[#C5A267]/10 text-[#C5A267] border border-[#C5A267]/20">
              ID: {node.id}
            </span>
            <select
              value={node.type}
              onChange={handleTypeChange}
              className="text-[10px] font-bold font-mono tracking-wider text-gray-300 bg-[#1A1A1D] hover:bg-white/5 border border-white/5 rounded-lg px-2.5 py-1.5 outline-none transition cursor-pointer"
            >
              {Object.entries(NODE_TYPES).map(([typeKey, cfg]) => (
                <option key={typeKey} value={typeKey} className="bg-[#131315]">
                  {cfg.label.toUpperCase()}
                </option>
              ))}
            </select>
          </div>
          
          <input
            type="text"
            value={node.label}
            onChange={handleLabelChange}
            placeholder="Ketik Nama Langkah..."
            className="text-sm font-bold text-white outline-none w-full border-b border-white/5 focus:border-[#C5A267] pb-1.5 font-sans"
          />
        </div>

        <button
          onClick={() => selectNode(null)}
          className="p-1 rounded-xl hover:bg-white/5 text-gray-500 hover:text-white transition cursor-pointer"
          title="Tutup Panel"
        >
          <X className="w-5 h-5" />
        </button>
      </div>

      {/* Tabs list (Bisnis, UI/UX, FE, BE) */}
      <div className="flex border-b border-white/5 bg-[#0A0A0B] p-1 gap-1">
        {(['bisnis', 'uiux', 'frontend', 'backend'] as const).map((tab) => {
          const isActive = activeTab === tab;
          const hasData = tab !== 'bisnis' && isFacetFilled(tab);

          return (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className={`flex-1 py-2 rounded-lg text-xs font-bold leading-none tracking-wide transition relative flex items-center justify-center space-x-1 cursor-pointer ${
                isActive
                  ? 'bg-[#131315] text-[#C5A267] shadow-lg border border-white/10'
                  : 'text-gray-400 hover:text-white hover:bg-white/5'
              }`}
            >
              <span className="capitalize">{tab === 'uiux' ? 'UI/UX' : tab === 'bisnis' ? 'Bisnis' : tab === 'frontend' ? 'FE' : 'BE'}</span>
              {hasData && (
                <span
                  style={{ backgroundColor: tab === 'uiux' ? '#EC4899' : tab === 'frontend' ? '#06B6D4' : '#F97316' }}
                  className="w-1.5 h-1.5 rounded-full block animate-pulse"
                />
              )}
            </button>
          );
        })}
      </div>

      {/* Accordion/Tab content wrapper (Scrollable area) */}
      <div className="flex-1 overflow-y-auto p-4 scrollbar-thin">
        {activeTab === 'bisnis' && <BisnisTab node={node} />}
        {activeTab === 'uiux' && <UiuxTab node={node} />}
        {activeTab === 'frontend' && <FrontendTab node={node} />}
        {activeTab === 'backend' && <BackendTab node={node} />}
      </div>

      {/* Absolute Bottom Action Footer */}
      <div className="p-4 border-t border-white/5 bg-[#0D0D0E] grid grid-cols-2 gap-2.5">
        <button
          onClick={handleConnectInitiate}
          className="flex items-center justify-center space-x-1.5 py-2.5 px-3 rounded-xl border border-[#C5A267] text-[#C5A267] hover:bg-[#C5A267]/10 text-[10px] font-mono font-bold tracking-widest cursor-pointer transition shadow-xl"
          title="Sambungkan langkah ini ke langkah lain"
        >
          <Link2 className="w-3.5 h-3.5" />
          <span>HUBUNGKAN</span>
        </button>
        <button
          onClick={handleDeleteNode}
          className="flex items-center justify-center space-x-1.5 py-2.5 px-3 rounded-xl border border-red-950/40 hover:border-red-500/50 hover:bg-red-950/30 text-red-400 text-[10px] font-mono font-bold tracking-widest cursor-pointer transition shadow-xl"
          title="Hapus langkah ini permanen"
        >
          <Trash2 className="w-3.5 h-3.5" />
          <span>HAPUS</span>
        </button>
      </div>

    </div>
  );
}
