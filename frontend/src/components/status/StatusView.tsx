/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { useStore } from '../../store/useStore';
import { ROLES } from '../../config/roles';
import { STATUS_CONFIG } from '../../config/status';
import { NODE_TYPES } from '../../config/nodeTypes';
import { PanelLeftClose, AlertCircle, LayoutGrid, CheckCircle } from 'lucide-react';

export default function StatusView() {
  const { modules, activeId, selectNode, setView } = useStore();
  const activeModule = modules.find((m) => m.id === activeId);

  if (!activeModule) {
    return (
      <div className="flex-1 p-8 bg-[#0A0A0B] text-gray-400 font-sans text-center flex items-center justify-center">
        Pilih modul terlebih dahulu...
      </div>
    );
  }

  const nodes = activeModule.nodes || [];

  const handleRowClick = (nodeId: string) => {
    selectNode(nodeId);
    setView('canvas'); // Transition to canvas mode
  };

  const getStatusStyle = (status: string) => {
    switch (status) {
      case 'planned':
        return {
          color: '#8A8A93',
          background: 'rgba(138, 138, 147, 0.08)',
          borderColor: 'rgba(138, 138, 147, 0.25)',
        };
      case 'in_progress':
        return {
          color: '#4B96F3',
          background: 'rgba(75, 150, 243, 0.08)',
          borderColor: 'rgba(75, 150, 243, 0.25)',
        };
      case 'review':
        return {
          color: '#906CF2',
          background: 'rgba(144, 108, 242, 0.08)',
          borderColor: 'rgba(144, 108, 242, 0.25)',
        };
      case 'done':
        return {
          color: '#C5A267',
          background: 'rgba(197, 162, 103, 0.08)',
          borderColor: 'rgba(197, 162, 103, 0.25)',
        };
      default:
        return {
          color: '#8A8A93',
          background: 'rgba(138, 138, 147, 0.08)',
          borderColor: 'rgba(138, 138, 147, 0.25)',
        };
    }
  };

  // Calculations for total statistics
  const totalTasksCount = nodes.length * 3;
  let doneTasksCount = 0;
  let inProgressCount = 0;
  let reviewCount = 0;
  let plannedCount = 0;

  nodes.forEach((n) => {
    ['uiux', 'frontend', 'backend'].forEach((role) => {
      const status = n.roles[role as 'uiux' | 'frontend' | 'backend']?.status || 'planned';
      if (status === 'done') doneTasksCount++;
      else if (status === 'in_progress') inProgressCount++;
      else if (status === 'review') reviewCount++;
      else plannedCount++;
    });
  });

  const percentComplete = totalTasksCount > 0 ? Math.round((doneTasksCount / totalTasksCount) * 100) : 0;

  return (
    <div className="flex-1 flex flex-col bg-[#0A0A0B] p-6 overflow-y-auto scrollbar-thin select-none">
      
      {/* Matrix Statistics Ribbon Header */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <div className="bg-[#131315] border border-white/5 p-4 rounded-2xl shadow-xl text-left">
          <span className="text-[9px] font-bold text-gray-400 font-mono uppercase tracking-widest block">Progres Keseluruhan</span>
          <div className="flex items-baseline space-x-2 mt-1">
            <span className="text-2xl font-black text-[#C5A267] font-sans tracking-tight">{percentComplete}%</span>
            <span className="text-[11px] text-gray-400 font-sans">Kesiapan Artefak</span>
          </div>
          <div className="w-full bg-white/5 h-2 rounded-full mt-2.5 overflow-hidden">
            <div style={{ width: `${percentComplete}%` }} className="bg-[#C5A267] h-full rounded-full transition-all duration-500" />
          </div>
        </div>

        <div className="bg-[#131315] border border-white/5 p-4 rounded-2xl shadow-xl text-left">
          <span className="text-[9px] font-bold text-emerald-400 font-mono uppercase tracking-widest block">Selesai (DONE)</span>
          <span className="text-2xl font-black text-emerald-400 mt-1 block tracking-tight">{doneTasksCount} <span className="text-sm font-normal text-gray-500">/ {totalTasksCount} tugas</span></span>
          <p className="text-[10px] text-gray-500 mt-1.5 font-sans">Siap diintegrasikan ke sandbox produksi.</p>
        </div>

        <div className="bg-[#131315] border border-white/5 p-4 rounded-2xl shadow-xl text-left">
          <span className="text-[9px] font-bold text-[#906CF2] font-mono uppercase tracking-widest block">Sedang Dikerjakan</span>
          <span className="text-2xl font-black text-white mt-1 block tracking-tight">{(inProgressCount + reviewCount)} <span className="text-sm font-normal text-gray-500">tugas aktif</span></span>
          <p className="text-[10px] text-gray-500 mt-1.5 font-sans">{inProgressCount} Pengembangan, {reviewCount} Peninjauan.</p>
        </div>

        <div className="bg-[#131315] border border-white/5 p-4 rounded-2xl shadow-xl text-left">
          <span className="text-[9px] font-bold text-gray-400 font-mono uppercase tracking-widest block">Belum Mulai (PLANNED)</span>
          <span className="text-2xl font-black text-gray-400 mt-1 block tracking-tight">{plannedCount} <span className="text-sm font-normal text-gray-500">tugas antrean</span></span>
          <p className="text-[10px] text-gray-500 mt-1.5 font-sans">Tugas belum dirilis ke sprint/papan kontributor.</p>
        </div>
      </div>

      {/* Matrix Table */}
      <div className="bg-[#131315] border border-white/5 rounded-2xl overflow-hidden shadow-2xl">
        <div className="p-4 border-b border-white/5 flex items-center justify-between">
          <h3 className="text-sm font-bold text-white uppercase tracking-widest flex items-center" style={{ fontFamily: 'Georgia, serif' }}>
            <LayoutGrid className="w-4 h-4 mr-1.5 text-[#C5A267]" />
            Matriks Keterlacakan Antar Peran Tim
          </h3>
          <span className="text-[10px] text-gray-500 font-mono tracking-wider italic">
            *Klik baris untuk membuka inspector di kanvas interaktif
          </span>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse">
            <thead>
              <tr className="bg-[#0D0D0E] border-b border-white/5 text-[9px] font-bold font-mono text-gray-400 uppercase tracking-widest">
                <th className="py-3 px-6 h-10">Langkah Proses Bisnis</th>
                <th className="py-3 px-6 h-10">Tipe</th>
                <th className="py-3 px-6 h-10 border-l border-white/5">UI / UX</th>
                <th className="py-3 px-6 h-10 border-l border-white/5">Frontend (FE)</th>
                <th className="py-3 px-6 h-10 border-l border-white/5">Backend (BE)</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5">
              {nodes.length === 0 ? (
                <tr>
                  <td colSpan={5} className="py-8 text-center text-xs text-gray-500">
                    Belum ada langkah alur yang diatur. Silakan tambahkan langkah baru di Lensa Kanvas.
                  </td>
                </tr>
              ) : (
                nodes.map((node) => (
                  <tr
                    key={node.id}
                    onClick={() => handleRowClick(node.id)}
                    className="hover:bg-white/2 cursor-pointer transition-colors"
                  >
                    {/* Node label & actor */}
                    <td className="py-3.5 px-6 font-medium text-white text-xs">
                      <div>
                        <p className="font-bold text-white">{node.label}</p>
                        <p className="text-[9px] text-gray-500 mt-0.5 font-sans">Actor: {node.doc.actor || 'Sistem'}</p>
                      </div>
                    </td>

                    {/* Step Type badge */}
                    <td className="py-3.5 px-6">
                      <span
                        style={{ color: NODE_TYPES[node.type]?.color, backgroundColor: 'rgba(197, 162, 103, 0.05)', borderColor: 'rgba(197, 162, 103, 0.15)' }}
                        className="text-[9px] font-mono font-bold px-2 py-0.5 rounded-full border opacity-90 uppercase"
                      >
                        {node.type}
                      </span>
                    </td>

                    {/* UI/UX facet cell */}
                    <td className="py-3.5 px-6 border-l border-white/5">
                      {node.roles.uiux && (node.roles.uiux.assignee || node.roles.uiux.status) ? (
                        <div className="flex flex-col space-y-1">
                          <span
                            style={getStatusStyle(node.roles.uiux.status || 'planned')}
                            className="text-[9px] font-mono font-bold px-2 py-0.5 rounded-md border text-center self-start"
                          >
                            {(node.roles.uiux.status || 'planned').toUpperCase()}
                          </span>
                          <span className="text-[10px] text-gray-400">
                            {node.roles.uiux.assignee || 'Belum ditunjuk'}
                          </span>
                        </div>
                      ) : (
                        <span className="text-[10px] text-gray-650 font-mono italic">Belum Ada</span>
                      )}
                    </td>

                    {/* FE facet cell */}
                    <td className="py-3.5 px-6 border-l border-white/5">
                      {node.roles.frontend && (node.roles.frontend.assignee || node.roles.frontend.status) ? (
                        <div className="flex flex-col space-y-1">
                          <span
                            style={getStatusStyle(node.roles.frontend.status || 'planned')}
                            className="text-[9px] font-mono font-bold px-2 py-0.5 rounded-md border text-center self-start"
                          >
                            {(node.roles.frontend.status || 'planned').toUpperCase()}
                          </span>
                          <span className="text-[10px] text-gray-400">
                            {node.roles.frontend.assignee || 'Belum ditunjuk'}
                          </span>
                        </div>
                      ) : (
                        <span className="text-[10px] text-gray-650 font-mono italic">Belum Ada</span>
                      )}
                    </td>

                    {/* BE facet cell */}
                    <td className="py-3.5 px-6 border-l border-white/5">
                      {node.roles.backend && (node.roles.backend.assignee || node.roles.backend.status) ? (
                        <div className="flex flex-col space-y-1">
                          <span
                            style={getStatusStyle(node.roles.backend.status || 'planned')}
                            className="text-[9px] font-mono font-bold px-2 py-0.5 rounded-md border text-center self-start"
                          >
                            {(node.roles.backend.status || 'planned').toUpperCase()}
                          </span>
                          <span className="text-[10px] text-gray-400">
                            {node.roles.backend.assignee || 'Belum ditunjuk'}
                          </span>
                        </div>
                      ) : (
                        <span className="text-[10px] text-gray-650 font-mono italic">Belum Ada</span>
                      )}
                    </td>

                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Legend & explanations */}
      <div className="mt-4 p-4 bg-[#131315] border border-white/5 rounded-2xl flex items-center flex-wrap gap-4 justify-between">
        <div className="flex items-center space-x-2 text-xs text-gray-400">
          <AlertCircle className="w-3.5 h-3.5 text-[#C5A267]" />
          <span className="font-semibold font-sans">Legenda Status Kesiapan:</span>
        </div>
        <div className="flex space-x-3.5">
          {Object.entries(STATUS_CONFIG).map(([key, value]) => {
            const calculatedStyles = getStatusStyle(key);
            return (
              <div key={key} className="flex items-center space-x-1.5 text-xs">
                <span
                  style={{ backgroundColor: calculatedStyles.color }}
                  className="w-2 h-2 rounded-full inline-block"
                />
                <span className="text-[10px] font-bold font-mono text-gray-400 uppercase tracking-wider">{value.label}</span>
              </div>
            );
          })}
        </div>
      </div>

    </div>
  );
}
