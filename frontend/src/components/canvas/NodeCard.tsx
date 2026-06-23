/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useRef, useState } from 'react';
import { useStore } from '../../store/useStore';
import { Node, NodeType, RoleKey } from '../../domain/types';
import { NODE_TYPES } from '../../config/nodeTypes';
import { ROLES } from '../../config/roles';
import { STATUS_CONFIG } from '../../config/status';
import * as Icons from 'lucide-react';

interface NodeCardProps {
  node: Node;
  key?: string | number;
}

export default function NodeCard({ node }: NodeCardProps) {
  const {
    activeId,
    selectedNodeId,
    selectNode,
    moveNode,
    connectFrom,
    addEdge,
    setConnectFrom,
  } = useStore();

  const isSelected = selectedNodeId === node.id;
  const isConnectingSource = connectFrom === node.id;
  const isConnectingTargetCandidate = connectFrom !== null && connectFrom !== node.id;

  const cardRef = useRef<HTMLDivElement>(null);

  // Drag state
  const [isDragging, setIsDragging] = useState(false);
  const dragStartPos = useRef({ x: 0, y: 0 });
  const nodeStartPos = useRef({ x: 0, y: 0 });
  const hasMoved = useRef(false);

  const config = NODE_TYPES[node.type] || NODE_TYPES.process;

  // Dynamically resolve icon component
  const IconComponent = (Icons as any)[config.icon] || Icons.Cpu;

  // Pointer event drag implementation
  const handlePointerDown = (e: React.PointerEvent<HTMLDivElement>) => {
    // Only drag on left click
    if (e.button !== 0) return;
    
    e.stopPropagation(); // Avoid deselecting in Canvas
    
    setIsDragging(true);
    hasMoved.current = false;
    dragStartPos.current = { x: e.clientX, y: e.clientY };
    nodeStartPos.current = { x: node.x, y: node.y };

    e.currentTarget.setPointerCapture(e.pointerId);
  };

  const handlePointerMove = (e: React.PointerEvent<HTMLDivElement>) => {
    if (!isDragging) return;
    e.stopPropagation();

    const dx = e.clientX - dragStartPos.current.x;
    const dy = e.clientY - dragStartPos.current.y;

    if (Math.abs(dx) > 2 || Math.abs(dy) > 2) {
      hasMoved.current = true;
    }

    const nextX = Math.max(0, nodeStartPos.current.x + dx);
    const nextY = Math.max(0, nodeStartPos.current.y + dy);

    moveNode(node.id, nextX, nextY);
  };

  const handlePointerUp = (e: React.PointerEvent<HTMLDivElement>) => {
    if (!isDragging) return;
    e.stopPropagation();

    setIsDragging(false);
    e.currentTarget.releasePointerCapture(e.pointerId);

    // If pointer was clicked (didn't drag far), fire selection or edge connection
    if (!hasMoved.current) {
      if (connectFrom !== null) {
        if (connectFrom !== node.id) {
          // Connect path
          addEdge(connectFrom, node.id);
        } else {
          // Cancel connect mode by clicking self
          setConnectFrom(null);
        }
      } else {
        selectNode(node.id);
      }
    }
  };

  // Render role badges (UX, FE, BE)
  const renderRoleBadges = () => {
    const roles: RoleKey[] = ['uiux', 'frontend', 'backend'];
    return (
      <div className="flex space-x-1.5 mt-2.5 pt-2 border-t border-white/5">
        {roles.map((rKey) => {
          const roleFacet = node.roles[rKey];
          const rConfig = ROLES[rKey];
          const isFilled = !!(roleFacet && (roleFacet.assignee || roleFacet.status));
          
          if (!isFilled) {
            return (
              <div
                key={rKey}
                className="flex-1 flex items-center justify-center py-1 text-[9px] font-mono font-bold rounded-md bg-white/5 text-gray-600 border border-dashed border-white/10"
                title={`${rConfig.label}: Belum Ditentukan`}
              >
                {rConfig.short}-
              </div>
            );
          }

          const status = roleFacet?.status || 'planned';
          const sConfig = STATUS_CONFIG[status];

          return (
            <div
              key={rKey}
              style={{
                backgroundColor: isSelected ? undefined : rConfig.bg,
                borderColor: rConfig.border,
                color: rConfig.color,
              }}
              className="flex-1 py-1 text-[9px] font-mono font-bold rounded-md border text-center flex flex-col items-center justify-center leading-none"
              title={`${rConfig.label}: ${roleFacet?.assignee || 'Belum ditunjuk'} (${sConfig?.label || ''})`}
            >
              <div className="flex items-center space-x-0.5 justify-center">
                <span>{rConfig.short}</span>
                <span
                  style={{ backgroundColor: sConfig?.color || '#CBD5E1' }}
                  className="w-1.5 h-1.5 rounded-full inline-block"
                />
              </div>
            </div>
          );
        })}
      </div>
    );
  };

  const getMethodColor = (method: string) => {
    switch (method) {
      case 'GET': return 'bg-emerald-500/10 text-emerald-400 border-emerald-400/20';
      case 'POST': return 'bg-blue-500/10 text-blue-400 border-blue-400/20';
      case 'PUT': return 'bg-amber-500/10 text-amber-400 border-amber-400/20';
      case 'DELETE': return 'bg-red-500/10 text-red-400 border-red-400/20';
      default: return 'bg-white/5 text-gray-400 border-white/10';
    }
  };

  return (
    <div
      ref={cardRef}
      onPointerDown={handlePointerDown}
      onPointerMove={handlePointerMove}
      onPointerUp={handlePointerUp}
      style={{
        left: node.x,
        top: node.y,
        width: 220,
        touchAction: 'none',
      }}
      className={`absolute select-none bg-[#131315] border rounded-2xl p-3.5 shadow-xl transform transition-all duration-150 cursor-grab active:cursor-grabbing ${
        isSelected
          ? 'ring-1 ring-[#C5A267] border-[#C5A267] shadow-[0_0_15px_rgba(197,162,103,0.30)] scale-[1.02] z-40'
          : isConnectingSource
          ? 'ring-1 ring-[#C5A267] border-[#C5A267]/80 bg-[#C5A267]/10 z-30 pulse-border'
          : isConnectingTargetCandidate
          ? 'hover:ring-1 hover:ring-rose-400 border-white/5 hover:border-rose-400/60 z-25'
          : 'border-white/10 hover:border-white/20 z-20 hover:shadow-2xl'
      }`}
    >
      {/* Node Header based on type */}
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center space-x-1.5 overflow-hidden">
          <div
            style={{ color: config.color, backgroundColor: config.bg }}
            className="w-5.5 h-5.5 rounded-lg flex items-center justify-center flex-shrink-0"
          >
            <IconComponent className="w-3.5 h-3.5" />
          </div>
          <span className="text-[10px] font-bold text-gray-500 uppercase tracking-widest truncate">
            {config.label}
          </span>
        </div>

        {/* API Endpoint indicator if Backend facet serves an endpoint */}
        {node.roles.backend?.endpoint && (
          <div className={`text-[8px] font-mono font-bold px-1.5 py-0.5 rounded border ${getMethodColor(node.roles.backend.method || 'GET')}`}>
            {node.roles.backend.method || 'GET'}
          </div>
        )}
      </div>

      {/* Main Node Label */}
      <h3 className="text-xs font-bold text-white tracking-tight leading-snug line-clamp-2 max-h-10 h-7 text-left">
        {node.label}
      </h3>

      {/* SLA / Owner tag if set */}
      <div className="flex items-center justify-between mt-2.5 text-[9px] font-medium text-gray-500">
        <span className="truncate max-w-24">
          Actor: <span className="font-semibold text-gray-300">{node.doc.actor || '-'}</span>
        </span>
        {node.doc.sla && (
          <span className="bg-white/5 hover:bg-white/10 text-gray-300 px-1 py-0.5 rounded font-mono border border-white/5">
            {node.doc.sla}
          </span>
        )}
      </div>

      {/* Role Badges */}
      {renderRoleBadges()}

      {/* Visual Overlay elements inside the card for connection flow context */}
      {isConnectingTargetCandidate && (
        <div className="absolute inset-0 bg-[#C5A267]/5 pointer-events-none rounded-2xl flex items-center justify-center border border-dashed border-[#C5A267]/35">
          <span className="text-[9px] bg-[#C5A267] text-black font-mono font-bold px-2 py-1 rounded-md shadow-lg tracking-wider">
            HUBUNGKAN
          </span>
        </div>
      )}
    </div>
  );
}
