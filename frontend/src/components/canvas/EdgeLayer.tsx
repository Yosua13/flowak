/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from 'react';
import { useStore } from '../../store/useStore';
import { Edge } from '../../domain/types';
import { Trash2, Edit2, Check, X } from 'lucide-react';

export default function EdgeLayer() {
  const { modules, activeId, deleteEdge, updateEdgeLabel } = useStore();
  const activeModule = modules.find((m) => m.id === activeId);

  const [editingEdgeId, setEditingEdgeId] = useState<string | null>(null);
  const [tempLabel, setTempLabel] = useState('');

  if (!activeModule || !activeModule.edges || activeModule.edges.length === 0) {
    return null;
  }

  // Estimated node dimensions
  const nodeWidth = 220;
  const nodeHeight = 135;

  const handleSaveLabel = (id: string) => {
    updateEdgeLabel(id, tempLabel.trim());
    setEditingEdgeId(null);
  };

  return (
    <svg className="absolute top-0 left-0 w-full h-full pointer-events-none select-none overflow-visible z-0" style={{ minWidth: 1500, minHeight: 1000 }}>
      <defs>
        <marker
          id="arrowhead"
          markerWidth="8"
          markerHeight="8"
          refX="6"
          refY="4"
          orient="auto"
          markerUnits="userSpaceOnUse"
        >
          <path d="M0,0 L8,4 L0,8 Z" fill="#C5A267" />
        </marker>
        <marker
          id="arrowhead-hover"
          markerWidth="8"
          markerHeight="8"
          refX="6"
          refY="4"
          orient="auto"
          markerUnits="userSpaceOnUse"
        >
          <path d="M0,0 L8,4 L0,8 Z" fill="#EF4444" />
        </marker>
      </defs>

      {activeModule.edges.map((edge) => {
        const fromNode = activeModule.nodes.find((n) => n.id === edge.from);
        const toNode = activeModule.nodes.find((n) => n.id === edge.to);

        if (!fromNode || !toNode) return null;

        // Center calculation
        const x1 = fromNode.x + nodeWidth / 2;
        const y1 = fromNode.y + nodeHeight / 2;
        const x2 = toNode.x + nodeWidth / 2;
        const y2 = toNode.y + nodeHeight / 2;

        const dx = x2 - x1;
        const dy = y2 - y1;
        const dist = Math.sqrt(dx * dx + dy * dy);

        if (dist === 0) return null;

        const ux = dx / dist;
        const uy = dy / dist;

        // Edge Pullback: arrow should end on the edge boundary
        // We calculate box intersection pullback
        const pullback = 115; // slightly shorter than half of width to touch the border
        const targetX = x2 - ux * pullback;
        const targetY = y2 - uy * pullback;

        // Midpoint for placing text labels and delete button
        const midX = x1 + dx * 0.5;
        const midY = y1 + dy * 0.5;

        const isEditing = editingEdgeId === edge.id;

        return (
          <g key={edge.id} className="pointer-events-auto group">
            {/* Clickable transparent thick path for easy select/delete */}
            <path
              d={`M ${x1} ${y1} L ${targetX} ${targetY}`}
              fill="none"
              stroke="transparent"
              strokeWidth="15"
              className="cursor-pointer"
              onClick={(e) => {
                e.stopPropagation();
                if (!isEditing) {
                  setEditingEdgeId(edge.id);
                  setTempLabel(edge.label || '');
                }
              }}
            />

            {/* Visual thin line with arrow */}
            <path
              d={`M ${x1} ${y1} L ${targetX} ${targetY}`}
              fill="none"
              stroke="#C5A267"
              strokeWidth="2.5"
              markerEnd="url(#arrowhead)"
              className="group-hover:stroke-red-500 transition-colors duration-150"
            />

            {/* Editing Box or Label */}
            {isEditing ? (
              <foreignObject
                x={midX - 85}
                y={midY - 16}
                width={170}
                height={32}
                className="z-50 pointer-events-auto"
              >
                <div
                  className="flex items-center space-x-1.5 bg-[#131315]/95 backdrop-blur-md border border-[#C5A267] rounded-xl p-1 shadow-2xl h-full"
                  onClick={(e) => e.stopPropagation()}
                >
                  <input
                    type="text"
                    value={tempLabel}
                    onChange={(e) => setTempLabel(e.target.value)}
                    className="w-full text-[10px] text-white bg-transparent outline-none px-2 font-mono"
                    placeholder="Nama label..."
                    autoFocus
                    onKeyDown={(e) => {
                      if (e.key === 'Enter') {
                        e.stopPropagation();
                        handleSaveLabel(edge.id);
                      } else if (e.key === 'Escape') {
                        e.stopPropagation();
                        setEditingEdgeId(null);
                      }
                    }}
                  />
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      handleSaveLabel(edge.id);
                    }}
                    className="p-1 bg-[#C5A267] text-black hover:bg-[#B38F52] rounded-lg cursor-pointer flex items-center justify-center transition-colors"
                  >
                    <Check className="w-3 h-3 text-black" />
                  </button>
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      setEditingEdgeId(null);
                    }}
                    className="p-1 bg-white/5 text-gray-400 hover:text-white rounded-lg cursor-pointer flex items-center justify-center transition-colors"
                    title="Batal"
                  >
                    <X className="w-3 h-3" />
                  </button>
                </div>
              </foreignObject>
            ) : (
              <>
                {/* Optional label */}
                {edge.label ? (
                  <g transform={`translate(${midX}, ${midY})`} className="pointer-events-none">
                    <rect
                      x={-edge.label.length * 3.5 - 6}
                      y={-9}
                      width={edge.label.length * 7 + 12}
                      height={18}
                      rx={4}
                      fill="#131315"
                      stroke="rgba(197, 162, 103, 0.3)"
                      strokeWidth="1"
                    />
                    <text
                      textAnchor="middle"
                      y={3}
                      className="text-[9px] font-mono font-bold fill-gray-300 pointer-events-none"
                    >
                      {edge.label}
                    </text>
                  </g>
                ) : null}

                {/* Hover control bar for Edit Label and Delete Connection */}
                <foreignObject
                  x={midX - 35}
                  y={edge.label ? midY - 32 : midY - 14}
                  width={70}
                  height={26}
                  className="opacity-0 group-hover:opacity-100 transition-opacity duration-150 pointer-events-auto"
                >
                  <div 
                    className="flex bg-[#131315]/95 border border-white/10 rounded-lg p-0.5 shadow-xl justify-around items-center h-full"
                    onClick={(e) => e.stopPropagation()}
                  >
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        setEditingEdgeId(edge.id);
                        setTempLabel(edge.label || '');
                      }}
                      className="p-1 text-[#C5A267] hover:text-[#E2C392] hover:bg-white/5 rounded transition-colors cursor-pointer"
                      title="Edit Label Hubungan"
                    >
                      <Edit2 className="w-3 h-3" />
                    </button>
                    <div className="w-px bg-white/5 h-3" />
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        if (window.confirm('Hapus koneksi alur kerja ini?')) {
                          deleteEdge(edge.id);
                        }
                      }}
                      className="p-1 text-red-400 hover:text-red-500 hover:bg-white/5 rounded transition-colors cursor-pointer"
                      title="Hapus Koneksi"
                    >
                      <Trash2 className="w-3 h-3" />
                    </button>
                  </div>
                </foreignObject>
              </>
            )}
          </g>
        );
      })}
    </svg>
  );
}
