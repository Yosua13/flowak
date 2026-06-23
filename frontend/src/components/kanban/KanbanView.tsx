/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from 'react';
import { useStore } from '../../store/useStore';
import { Node, RoleKey, Status } from '../../domain/types';
import { Kanban, Sparkles, Filter, ChevronRight, User, Cpu, ArrowRightLeft, Check, CheckCircle2 } from 'lucide-react';

interface KanbanTask {
  nodeId: string;
  nodeLabel: string;
  roleKey: RoleKey;
  assignee: string;
  status: Status;
  detail: any;
}

export default function KanbanView() {
  const { modules, activeId, updateRole, teamMembers } = useStore();
  const activeModule = modules.find((m) => m.id === activeId);

  const [roleFilter, setRoleFilter] = useState<string>('all');

  if (!activeModule) {
    return (
      <div className="flex-1 p-8 bg-[#0A0A0B] text-gray-500 text-center flex items-center justify-center">
        <p>Silakan buat atau pilih modul alur kerja untuk melihat Papan Kanban.</p>
      </div>
    );
  }

  // Gather all tasks of the active module
  const tasks: KanbanTask[] = [];
  activeModule.nodes.forEach((node) => {
    ['uiux', 'frontend', 'backend'].forEach((role) => {
      const facet = node.roles[role as RoleKey];
      if (facet) {
        tasks.push({
          nodeId: node.id,
          nodeLabel: node.label,
          roleKey: role as RoleKey,
          assignee: facet.assignee || 'Unassigned',
          status: facet.status || 'planned',
          detail: facet,
        });
      }
    });
  });

  // Filtered tasks
  const filteredTasks = tasks.filter((t) => {
    if (roleFilter === 'all') return true;
    return t.roleKey === roleFilter;
  });

  const columns: { id: Status; label: string; color: string; border: string }[] = [
    { id: 'planned', label: 'Terencana (Backlog)', color: 'border-l-4 border-l-gray-400 bg-gray-950/20', border: 'border-white/5' },
    { id: 'in_progress', label: 'Pengerjaan (In Progress)', color: 'border-l-4 border-l-blue-400 bg-blue-950/5', border: 'border-blue-900/10' },
    { id: 'review', label: 'Peninjauan (In Review)', color: 'border-l-4 border-l-purple-400 bg-purple-950/5', border: 'border-purple-900/10' },
    { id: 'done', label: 'Selesai (Completed)', color: 'border-l-4 border-l-emerald-400 bg-emerald-950/5', border: 'border-emerald-900/10' },
  ];

  const handleMoveStatus = (nodeId: string, roleKey: RoleKey, nextStatus: Status) => {
    updateRole(nodeId, roleKey, { status: nextStatus });
  };

  const getRoleLabel = (r: RoleKey) => {
    switch (r) {
      case 'uiux': return 'UI/UX Design';
      case 'frontend': return 'Frontend FE';
      case 'backend': return 'Backend BE';
    }
  };

  const getRoleBadge = (r: RoleKey) => {
    switch (r) {
      case 'uiux': return 'bg-pink-500/10 text-pink-400 border-pink-500/20';
      case 'frontend': return 'bg-sky-500/10 text-sky-400 border-sky-500/20';
      case 'backend': return 'bg-purple-500/10 text-purple-400 border-purple-500/20';
    }
  };

  return (
    <div className="flex-1 bg-[#0A0A0B] p-6 overflow-y-auto scrollbar-thin select-none text-left">
      
      {/* View Header */}
      <div className="mb-6 flex flex-col md:flex-row items-start md:items-center justify-between border-b border-white/5 pb-4">
        <div>
          <h2 className="text-sm font-bold text-white flex items-center uppercase tracking-widest" style={{ fontFamily: 'Georgia, serif' }}>
            <Kanban className="w-5 h-5 mr-2 text-[#C5A267]" />
            Papan Kanban Transparansi Tugas
          </h2>
          <p className="text-xs text-gray-400 mt-1">
            Pantau dan gerakkan status pengerjaan spesifikasi UI/UX, Frontend, dan Backend per langkah dari alur kerja secara kolaboratif.
          </p>
        </div>

        {/* Filter controls */}
        <div className="flex items-center space-x-2 mt-4 md:mt-0">
          <Filter className="w-3.5 h-3.5 text-gray-400" />
          <span className="text-[10px] font-bold font-mono text-gray-400 uppercase tracking-widest">Filter Tim:</span>
          <select
            value={roleFilter}
            onChange={(e) => setRoleFilter(e.target.value)}
            className="text-xs border border-white/10 rounded-xl px-3 py-1.5 bg-[#131315] text-white outline-none focus:ring-1 focus:ring-[#C5A267]"
          >
            <option value="all">Semua Peran</option>
            <option value="uiux">UI/UX Designer</option>
            <option value="frontend">Frontend Engineer</option>
            <option value="backend">Backend Engineer</option>
          </select>
        </div>
      </div>

      {/* Kanban Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-5 items-start">
        {columns.map((col) => {
          const colTasks = filteredTasks.filter((t) => t.status === col.id);
          return (
            <div
              key={col.id}
              className={`rounded-2xl border p-4 flex flex-col min-h-[500px] h-full ${col.color} ${col.border}`}
            >
              {/* Column Header */}
              <div className="flex items-center justify-between border-b border-white/5 pb-2.5 mb-4">
                <span className="text-xs font-bold text-white uppercase tracking-wider">{col.label}</span>
                <span className="bg-white/5 text-gray-400 font-mono text-[10px] font-bold px-2 py-0.5 rounded-full">
                  {colTasks.length}
                </span>
              </div>

              {/* Column Cards */}
              <div className="flex-1 flex flex-col space-y-3.5 overflow-y-auto max-h-[700px] scrollbar-none pb-4">
                {colTasks.map((task, idx) => (
                  <div
                    key={`${task.nodeId}_${task.roleKey}`}
                    className="p-4 bg-[#131315]/95 border border-white/5 rounded-2xl shadow-xl hover:border-white/15 hover:bg-[#131315]/100 transition duration-150 flex flex-col space-y-3 pb-3"
                  >
                    {/* Node and Role headers */}
                    <div className="flex items-start justify-between gap-1">
                      <span className="text-xs font-bold text-white tracking-tight line-clamp-2">{task.nodeLabel}</span>
                      <span className={`text-[9px] font-bold font-mono tracking-wider px-2 py-0.5 rounded border uppercase flex-shrink-0 ${getBadgeBadge(task.roleKey)}`}>
                        {task.roleKey.toUpperCase()}
                      </span>
                    </div>

                    {/* Member Assignee name view */}
                    <div className="flex items-center space-x-2 text-gray-400 text-[11px] font-sans">
                      <User className="w-3.5 h-3.5 text-gray-500" />
                      <span>{task.assignee}</span>
                    </div>

                    {/* Status Mover Quick Switcher */}
                    <div className="pt-2 border-t border-white/3 flex items-center justify-between">
                      <span className="text-[9px] font-mono font-bold text-gray-550 uppercase tracking-widest">Kirim Ke:</span>
                      <div className="flex items-center space-x-1">
                        {columns.map((c) => {
                          if (c.id === task.status) return null;
                          return (
                            <button
                              key={c.id}
                              onClick={() => handleMoveStatus(task.nodeId, task.roleKey, c.id)}
                              className="text-[8px] font-mono font-bold px-1.5 py-1 rounded bg-white/5 hover:bg-[#C5A267] hover:text-black text-gray-300 transition duration-100 cursor-pointer uppercase"
                              title={`Ubah status ke ${c.label}`}
                            >
                              {c.id === 'planned' ? 'PLAN' : c.id === 'in_progress' ? 'PROG' : c.id === 'review' ? 'REV' : 'DONE'}
                            </button>
                          );
                        })}
                      </div>
                    </div>
                  </div>
                ))}

                {colTasks.length === 0 && (
                  <div className="flex-1 flex flex-col items-center justify-center p-6 border-2 border-dashed border-white/5 rounded-2xl text-gray-600 mt-2">
                    <p className="text-[10px] uppercase tracking-widest font-mono">KOSONG</p>
                  </div>
                )}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

// Inline helper for bad nomenclature errors
function getBadgeBadge(r: RoleKey) {
  switch (r) {
    case 'uiux': return 'bg-pink-500/10 text-pink-400 border-pink-500/20';
    case 'frontend': return 'bg-sky-500/10 text-sky-400 border-sky-500/20';
    case 'backend': return 'bg-purple-500/10 text-purple-400 border-purple-500/20';
  }
}
