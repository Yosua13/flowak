/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from 'react';
import { useStore } from '../../store/useStore';
import { Calendar as CalIcon, ChevronLeft, ChevronRight, Clock, User, CheckCircle, AlertOctagon, ListFilter } from 'lucide-react';
import { STATUS_CONFIG } from '../../config/status';
import { NODE_TYPES } from '../../config/nodeTypes';

export default function CalendarView() {
  const { modules, activeId } = useStore();
  const activeModule = modules.find((m) => m.id === activeId);

  const [currentMonth, setCurrentMonth] = useState<number>(5); // June (0-indexed is 5)
  const [currentYear, setCurrentYear] = useState<number>(2026);
  const [selectedCalendarTask, setSelectedCalendarTask] = useState<any | null>(null);

  if (!activeModule) {
    return (
      <div className="flex-1 p-8 bg-slate-50 dark:bg-slate-950 text-slate-500">
        Pilih modul terlebih dahulu...
      </div>
    );
  }

  const nodes = activeModule.nodes || [];

  // Month configurations
  const monthNames = [
    'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
    'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'
  ];

  const daysOfWeek = ['Min', 'Sen', 'Sel', 'Rab', 'Kam', 'Jum', 'Sab'];

  // June 2026 starts on Monday (index 1) and has 30 days
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

  // Dynamically map nodes to days in our calendar (June 2026) deterministically to look populated and real!
  const getTasksForDay = (day: number) => {
    // June 2026 specific task mapping
    const tasks: any[] = [];
    nodes.forEach((node, idx) => {
      // Map deterministic days: e.g. Node-0 starts on day 3, Node-1 on day 8, Node-2 on day 15...
      const scheduledDay = 3 + (idx * 5);
      if (scheduledDay === day && currentMonth === 5 && currentYear === 2026) {
        tasks.push(node);
      }
    });
    return tasks;
  };

  const getStatusColor = (status: string) => {
    const config = STATUS_CONFIG[status as any];
    return config?.color || '#3B82F6';
  };

  // Generate calendar days
  const startDayOffset = 1; // June 1st, 2026 is a Monday (1)
  const totalDaysInMonth = 30; // June

  const calendarCells: any[] = [];
  // Fill initial empty slots from prev month
  for (let i = 0; i < startDayOffset; i++) {
    calendarCells.push({ d: null, current: false });
  }
  // Fill current month days
  for (let d = 1; d <= totalDaysInMonth; d++) {
    calendarCells.push({ d, current: true, tasks: getTasksForDay(d) });
  }

  return (
    <div className="flex-1 flex flex-col md:flex-row bg-[#0A0A0B] p-6 gap-6 overflow-y-auto scrollbar-thin">
      
      {/* Calendar Master Grid Module */}
      <div className="flex-1 bg-[#131315] border border-white/5 rounded-2xl overflow-hidden shadow-2xl flex flex-col">
        {/* Calendar Header Navigator */}
        <div className="p-4 border-b border-white/5 flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <CalIcon className="w-5 h-5 text-[#C5A267]" />
            <h3 className="text-xs font-bold text-white uppercase tracking-widest font-sans" style={{ fontFamily: 'Georgia, serif' }}>
              Jadwal Koordinasi Sprint Kalender
            </h3>
          </div>
          <div className="flex items-center space-x-3">
            <button
              onClick={handlePrevMonth}
              className="p-1 px-2.5 rounded-lg border border-white/10 hover:bg-white/5 cursor-pointer font-bold transition text-gray-300 hover:text-white"
            >
              <ChevronLeft className="w-4 h-4" />
            </button>
            <span className="text-xs font-bold text-white font-mono w-28 text-center uppercase tracking-wider">
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

        {/* Days of week */}
        <div className="grid grid-cols-7 border-b border-white/5 bg-white/2 p-2 text-center text-xs font-bold font-mono text-gray-400">
          {daysOfWeek.map((day) => (
            <div key={day} className="py-2.5">{day}</div>
          ))}
        </div>

        {/* Grid Cells */}
        <div className="grid grid-cols-7 flex-1 min-h-96">
          {calendarCells.map((cell, idx) => {
            const hasTasks = cell.tasks && cell.tasks.length > 0;
            return (
              <div
                key={idx}
                className={`border-b border-r last:border-r-0 border-white/5 p-1.5 flex flex-col h-24 hover:bg-white/5 relative select-none ${
                  cell.current ? 'text-white' : 'text-gray-600 bg-white/2'
                }`}
              >
                {cell.d !== null && (
                  <span className={`text-[10px] font-bold font-mono tracking-wider ${cell.d === 22 && currentMonth === 5 ? 'bg-[#C5A267] text-black w-5 h-5 rounded-full flex items-center justify-center font-bold' : ''}`}>
                    {cell.d}
                  </span>
                )}

                {/* Day tasks blocks */}
                <div className="mt-1 space-y-1 overflow-y-auto max-h-16 scrollbar-thin">
                  {cell.tasks && cell.tasks.map((task: any) => {
                    const nodeTypeColor = NODE_TYPES[task.type]?.color || '#C5A267';
                    return (
                      <div
                        key={task.id}
                        onClick={(e) => {
                          e.stopPropagation();
                          setSelectedCalendarTask(task);
                        }}
                        style={{ borderLeftColor: nodeTypeColor }}
                        className="text-[9px] font-sans font-bold bg-[#1A1A1D] hover:bg-white/5 border border-white/5 border-l-4 rounded px-1.5 py-0.5 truncate text-left text-gray-300 cursor-pointer transition shadow-sm"
                        title={task.label}
                      >
                        {task.label}
                      </div>
                    );
                  })}
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {/* Target schedule details panel */}
      <div className="w-full md:w-64 bg-[#111113] border border-white/5 p-4 rounded-2xl shadow-xl flex flex-col">
        <h4 className="text-[10px] font-bold text-[#C5A267] uppercase tracking-widest font-mono border-b border-white/5 pb-2 mb-3.5 flex items-center">
          <Clock className="w-3.5 h-3.5 mr-1.5 text-[#C5A267]" />
          Rincian Tenggat & SLA
        </h4>

        {selectedCalendarTask ? (
          <div className="space-y-4 text-left">
            <div>
              <span
                style={{
                  color: NODE_TYPES[selectedCalendarTask.type]?.color,
                  backgroundColor: 'rgba(197, 162, 103, 0.05)',
                  borderColor: 'rgba(197, 162, 103, 0.2)'
                }}
                className="text-[9px] font-mono font-bold px-2 py-0.5 rounded border uppercase"
              >
                {selectedCalendarTask.type}
              </span>
              <h5 className="font-bold text-white text-sm mt-3 leading-snug">
                {selectedCalendarTask.label}
              </h5>
            </div>

            <div className="space-y-2.5 text-xs text-gray-400">
              <div className="flex items-center space-x-1.5">
                <User className="w-3.5 h-3.5 text-gray-400" />
                <span>Aktor: <strong className="text-gray-200">{selectedCalendarTask.doc.actor || 'Sistem'}</strong></span>
              </div>
              <div className="flex items-center space-x-1.5">
                <Clock className="w-3.5 h-3.5 text-gray-400" />
                <span>SLA Kerja: <strong className="text-[#C5A267] font-mono">{selectedCalendarTask.doc.sla || 'Instan'}</strong></span>
              </div>
            </div>

            <div className="pt-3 border-t border-white/5 space-y-2">
              <p className="text-[10px] font-bold text-gray-400 font-mono uppercase tracking-wider">Statistik Kesiapan:</p>
              
              <div className="flex items-center justify-between text-xs text-gray-300">
                <span>UI/UX Desain</span>
                <span className="font-mono text-[10px] font-bold uppercase" style={{ color: getStatusColor(selectedCalendarTask.roles.uiux?.status || 'planned') }}>
                  {selectedCalendarTask.roles.uiux?.status || 'Planned'}
                </span>
              </div>

              <div className="flex items-center justify-between text-xs text-gray-300">
                <span>Frontend Dev</span>
                <span className="font-mono text-[10px] font-bold uppercase" style={{ color: getStatusColor(selectedCalendarTask.roles.frontend?.status || 'planned') }}>
                  {selectedCalendarTask.roles.frontend?.status || 'Planned'}
                </span>
              </div>

              <div className="flex items-center justify-between text-xs text-gray-300">
                <span>Backend Dev</span>
                <span className="font-mono text-[10px] font-bold uppercase" style={{ color: getStatusColor(selectedCalendarTask.roles.backend?.status || 'planned') }}>
                  {selectedCalendarTask.roles.backend?.status || 'Planned'}
                </span>
              </div>
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
              Klik salah satu petak tugas yang terjadwal di kalender untuk merinci SLA & owner.
            </p>
          </div>
        )}
      </div>

    </div>
  );
}
