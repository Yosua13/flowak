/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useMemo, useState } from 'react';
import { useStore } from '../../store/useStore';
import { Calendar as CalIcon, ChevronLeft, ChevronRight, Clock, User, ListFilter } from 'lucide-react';
import { STATUS_CONFIG } from '../../config/status';
import { NODE_TYPES } from '../../config/nodeTypes';
import { RoleKey, Status } from '../../domain/types';

interface CalendarTask {
  nodeId: string;
  nodeLabel: string;
  nodeType: keyof typeof NODE_TYPES;
  roleKey: RoleKey;
  assignee: string;
  status: Status;
  dueDate: string;
  actor?: string;
  sla?: string;
}

export default function CalendarView() {
  const { modules, activeId, teamMembers } = useStore();
  const activeModule = modules.find((m) => m.id === activeId);

  const today = new Date();
  const [currentMonth, setCurrentMonth] = useState<number>(today.getMonth());
  const [currentYear, setCurrentYear] = useState<number>(today.getFullYear());
  const [selectedCalendarTask, setSelectedCalendarTask] = useState<CalendarTask | null>(null);

  const monthNames = [
    'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
    'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'
  ];
  const daysOfWeek = ['Min', 'Sen', 'Sel', 'Rab', 'Kam', 'Jum', 'Sab'];

  const tasks = useMemo<CalendarTask[]>(() => {
    if (!activeModule) return [];
    const out: CalendarTask[] = [];
    const registeredNames = new Set(teamMembers.map((m) => m.name));

    activeModule.nodes.forEach((node) => {
      (['uiux', 'frontend', 'backend'] as RoleKey[]).forEach((roleKey) => {
        const facet = node.roles[roleKey];
        if (!facet?.dueDate) return;
        out.push({
          nodeId: node.id,
          nodeLabel: node.label,
          nodeType: node.type,
          roleKey,
          assignee: facet.assignee && registeredNames.has(facet.assignee) ? facet.assignee : 'Belum ditunjuk',
          status: facet.status || 'planned',
          dueDate: facet.dueDate,
          actor: node.doc.actor,
          sla: node.doc.sla,
        });
      });
    });

    return out;
  }, [activeModule, teamMembers]);

  if (!activeModule) {
    return (
      <div className="flex-1 p-8 bg-[#0A0A0B] text-gray-500 flex items-center justify-center">
        Pilih modul terlebih dahulu...
      </div>
    );
  }

  const handlePrevMonth = () => {
    if (currentMonth === 0) {
      setCurrentMonth(11);
      setCurrentYear(currentYear - 1);
    } else {
      setCurrentMonth(currentMonth - 1);
    }
  };

  const handleNextMonth = () => {
    if (currentMonth === 11) {
      setCurrentMonth(0);
      setCurrentYear(currentYear + 1);
    } else {
      setCurrentMonth(currentMonth + 1);
    }
  };

  const getTasksForDay = (day: number) => {
    return tasks.filter((task) => {
      const [year, month, date] = task.dueDate.split('-').map(Number);
      return year === currentYear && month === currentMonth + 1 && date === day;
    });
  };

  const getStatusColor = (status: Status) => STATUS_CONFIG[status]?.color || '#3B82F6';

  const startDayOffset = new Date(currentYear, currentMonth, 1).getDay();
  const totalDaysInMonth = new Date(currentYear, currentMonth + 1, 0).getDate();

  const calendarCells: Array<{ d: number | null; current: boolean; tasks: CalendarTask[] }> = [];
  for (let i = 0; i < startDayOffset; i++) {
    calendarCells.push({ d: null, current: false, tasks: [] });
  }
  for (let d = 1; d <= totalDaysInMonth; d++) {
    calendarCells.push({ d, current: true, tasks: getTasksForDay(d) });
  }

  return (
    <div className="flex-1 flex flex-col md:flex-row bg-[#0A0A0B] p-6 gap-6 overflow-y-auto scrollbar-thin">
      <div className="flex-1 bg-[#131315] border border-white/5 rounded-2xl overflow-hidden shadow-2xl flex flex-col">
        <div className="p-4 border-b border-white/5 flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <CalIcon className="w-5 h-5 text-[#C5A267]" />
            <h3 className="text-xs font-bold text-white uppercase tracking-widest font-sans" style={{ fontFamily: 'Georgia, serif' }}>
              Jadwal Delivery Artefak
            </h3>
          </div>
          <div className="flex items-center space-x-3">
            <button
              onClick={handlePrevMonth}
              className="p-1 px-2.5 rounded-lg border border-white/10 hover:bg-white/5 cursor-pointer font-bold transition text-gray-300 hover:text-white"
            >
              <ChevronLeft className="w-4 h-4" />
            </button>
            <span className="text-xs font-bold text-white font-mono w-32 text-center uppercase tracking-wider">
              {monthNames[currentMonth]} {currentYear}
            </span>
            <button
              onClick={handleNextMonth}
              className="p-1 px-2.5 rounded-lg border border-white/10 hover:bg-white/5 cursor-pointer font-bold transition text-gray-300 hover:text-white"
            >
              <ChevronRight className="w-4 h-4" />
            </button>
          </div>
        </div>

        <div className="grid grid-cols-7 border-b border-white/5 bg-white/2 p-2 text-center text-xs font-bold font-mono text-gray-400">
          {daysOfWeek.map((day) => (
            <div key={day} className="py-2.5">{day}</div>
          ))}
        </div>

        <div className="grid grid-cols-7 flex-1 min-h-96">
          {calendarCells.map((cell, idx) => {
            const isToday =
              cell.d === today.getDate() &&
              currentMonth === today.getMonth() &&
              currentYear === today.getFullYear();

            return (
              <div
                key={idx}
                className={`border-b border-r last:border-r-0 border-white/5 p-1.5 flex flex-col h-24 hover:bg-white/5 relative select-none ${
                  cell.current ? 'text-white' : 'text-gray-600 bg-white/2'
                }`}
              >
                {cell.d !== null && (
                  <span className={`text-[10px] font-bold font-mono tracking-wider ${isToday ? 'bg-[#C5A267] text-black w-5 h-5 rounded-full flex items-center justify-center' : ''}`}>
                    {cell.d}
                  </span>
                )}

                <div className="mt-1 space-y-1 overflow-y-auto max-h-16 scrollbar-thin">
                  {cell.tasks.map((task) => {
                    const nodeTypeColor = NODE_TYPES[task.nodeType]?.color || '#C5A267';
                    return (
                      <button
                        key={`${task.nodeId}_${task.roleKey}`}
                        onClick={(e) => {
                          e.stopPropagation();
                          setSelectedCalendarTask(task);
                        }}
                        style={{ borderLeftColor: nodeTypeColor }}
                        className="w-full text-[9px] font-sans font-bold bg-[#1A1A1D] hover:bg-white/5 border border-white/5 border-l-4 rounded px-1.5 py-0.5 truncate text-left text-gray-300 cursor-pointer transition shadow-sm"
                        title={`${task.nodeLabel} - ${task.roleKey.toUpperCase()}`}
                      >
                        {task.roleKey.toUpperCase()} · {task.nodeLabel}
                      </button>
                    );
                  })}
                </div>
              </div>
            );
          })}
        </div>
      </div>

      <div className="w-full md:w-64 bg-[#111113] border border-white/5 p-4 rounded-2xl shadow-xl flex flex-col">
        <h4 className="text-[10px] font-bold text-[#C5A267] uppercase tracking-widest font-mono border-b border-white/5 pb-2 mb-3.5 flex items-center">
          <Clock className="w-3.5 h-3.5 mr-1.5 text-[#C5A267]" />
          Rincian Tenggat
        </h4>

        {selectedCalendarTask ? (
          <div className="space-y-4 text-left">
            <div>
              <span
                style={{
                  color: NODE_TYPES[selectedCalendarTask.nodeType]?.color,
                  backgroundColor: 'rgba(197, 162, 103, 0.05)',
                  borderColor: 'rgba(197, 162, 103, 0.2)'
                }}
                className="text-[9px] font-mono font-bold px-2 py-0.5 rounded border uppercase"
              >
                {selectedCalendarTask.roleKey}
              </span>
              <h5 className="font-bold text-white text-sm mt-3 leading-snug">
                {selectedCalendarTask.nodeLabel}
              </h5>
            </div>

            <div className="space-y-2.5 text-xs text-gray-400">
              <div className="flex items-center space-x-1.5">
                <User className="w-3.5 h-3.5 text-gray-400" />
                <span>Owner: <strong className="text-gray-200">{selectedCalendarTask.assignee}</strong></span>
              </div>
              <div className="flex items-center space-x-1.5">
                <Clock className="w-3.5 h-3.5 text-gray-400" />
                <span>Due: <strong className="text-[#C5A267] font-mono">{selectedCalendarTask.dueDate}</strong></span>
              </div>
              {selectedCalendarTask.actor && (
                <div className="flex items-center space-x-1.5">
                  <ListFilter className="w-3.5 h-3.5 text-gray-400" />
                  <span>Aktor: <strong className="text-gray-200">{selectedCalendarTask.actor}</strong></span>
                </div>
              )}
            </div>

            <div className="pt-3 border-t border-white/5 space-y-2">
              <p className="text-[10px] font-bold text-gray-400 font-mono uppercase tracking-wider">Status Artefak:</p>
              <div className="flex items-center justify-between text-xs text-gray-300">
                <span>{selectedCalendarTask.roleKey.toUpperCase()}</span>
                <span className="font-mono text-[10px] font-bold uppercase" style={{ color: getStatusColor(selectedCalendarTask.status) }}>
                  {selectedCalendarTask.status}
                </span>
              </div>
              {selectedCalendarTask.sla && (
                <div className="flex items-center justify-between text-xs text-gray-300">
                  <span>SLA Flow</span>
                  <span className="font-mono text-[10px] text-[#C5A267]">{selectedCalendarTask.sla}</span>
                </div>
              )}
            </div>

            <button
              onClick={() => setSelectedCalendarTask(null)}
              className="w-full text-center text-[10px] font-mono font-bold tracking-widest py-2 border border-white/10 rounded-lg text-gray-400 hover:text-white hover:bg-white/5 transition cursor-pointer uppercase"
            >
              Reset Pemilihan
            </button>
          </div>
        ) : (
          <div className="flex-1 flex flex-col items-center justify-center text-gray-500 py-12">
            <CalIcon className="w-8 h-8 text-white/5 mb-2" />
            <p className="text-[10px] font-medium max-w-44 text-center leading-relaxed">
              Pilih task bertanggal di kalender.
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
