/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from 'react';
import { useStore } from '../../store/useStore';
import { TrendingUp, Award, Star, CheckCircle, BarChart2, ShieldCheck, Clock, UserCheck, AlertOctagon, Sparkles, Loader2, Play, ChevronRight, AlertTriangle, Info, Check } from 'lucide-react';
import { STATUS_CONFIG } from '../../config/status';

export default function AnalyticsView() {
  const { modules, activeId, teamMembers, addNotification } = useStore();
  const activeModule = modules.find((m) => m.id === activeId);

  // AI Auditor states
  const [auditing, setAuditing] = useState(false);
  const [auditResult, setAuditResult] = useState<{
    score: number;
    summary: string;
    issues: Array<{ severity: 'error' | 'warning' | 'info'; message: string; nodeName?: string; type: string }>;
  } | null>(null);

  if (!activeModule) {
    return (
      <div className="flex-1 p-8 bg-[#0A0A0B] text-gray-500 text-center select-none flex items-center justify-center">
        <p>Silakan buat atau pilih modul alur kerja untuk mengakses panel Evaluasi Analitik.</p>
      </div>
    );
  }

  const nodes = activeModule.nodes || [];

  // Aggregation of states
  const totalSteps = nodes.length;
  const totalTasks = totalSteps * 3;

  let doneTaskCount = 0;
  let inProgressTaskCount = 0;
  let reviewTaskCount = 0;
  let plannedTaskCount = 0;

  nodes.forEach((n) => {
    ['uiux', 'frontend', 'backend'].forEach((role) => {
      const facet = n.roles[role as 'uiux' | 'frontend' | 'backend'];
      const status = facet?.status || 'planned';

      if (status === 'done') doneTaskCount++;
      else if (status === 'in_progress') inProgressTaskCount++;
      else if (status === 'review') reviewTaskCount++;
      else plannedTaskCount++;
    });
  });

  const completionRate = totalTasks > 0 ? Math.round((doneTaskCount / totalTasks) * 100) : 0;

  // Deterministic statements for evaluation
  let evaluationTitle = 'Butuh Koordinasi Segera';
  let evaluationStatement = 'Mayoritas tugas masih dalam tahap perencanaan atau pengerjaan awal. Disarankan melakukan sinkronisasi harian.';
  let badgeColor = 'bg-amber-100 text-amber-800 dark:bg-amber-950/20 dark:text-amber-400 border-amber-950/40';

  if (completionRate >= 85) {
    evaluationTitle = 'Kinerja Sangat Baik (Sinergi Sempurna)';
    evaluationStatement = 'Tim berkolaborasi dengan efisiensi mendekati 100%. Mayoritas artefak selesai disusun dan siap rilis produksi.';
    badgeColor = 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950/20 dark:text-emerald-400 border-emerald-950/40';
  } else if (completionRate >= 40) {
    evaluationTitle = 'Produktivitas Stabil';
    evaluationStatement = 'Alur kerja berjalan dengan lancar. Tetap pertahankan momentum harian menjelang uji coba rilis.';
    badgeColor = 'bg-indigo-100 text-indigo-805 dark:bg-indigo-950/20 dark:text-indigo-400 border-indigo-950/40';
  }

  // SLA parsing
  let totalEstimasiHari = 0;
  nodes.forEach((n) => {
    const sla = (n.doc.sla || '').toLowerCase();
    if (sla.includes('jam')) {
      const hours = parseInt(sla) || 0;
      totalEstimasiHari += hours / 8;
    } else if (sla.includes('hari')) {
      const days = parseInt(sla) || 0;
      totalEstimasiHari += days;
    } else if (sla.includes('menit')) {
      const mins = parseInt(sla) || 0;
      totalEstimasiHari += mins / 480;
    }
  });

  const estimatedLeadTime = totalEstimasiHari > 0 ? (totalEstimasiHari).toFixed(1) : '1.5';

  // Trigger real AI analysis flow
  const handlePerformAudit = async () => {
    setAuditing(true);
    try {
      const res = await fetch('/api/ai/audit-flow', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ module: activeModule }),
      });

      const data = await res.json();
      if (data.error) throw new Error(data.error);

      setAuditResult(data);
      addNotification('Audit Spesifikasi Selesai', `Alur kerja berhasil di-audit dengan skor rancangan ${data.score}/100.`, 'success');
    } catch (err: any) {
      console.error(err);
      addNotification('Audit AI Gagal', 'Gagal memperoleh data audit API.', 'warning');
    } finally {
      setAuditing(false);
    }
  };

  const getSeverityStyle = (sev: string) => {
    switch (sev) {
      case 'error': return 'text-red-400 bg-red-950/25 border-red-900/40';
      case 'warning': return 'text-amber-400 bg-amber-950/25 border-amber-900/40';
      default: return 'text-sky-400 bg-sky-950/25 border-sky-900/40';
    }
  };

  return (
    <div className="flex-1 bg-[#0A0A0B] p-6 overflow-y-auto scrollbar-thin select-none pb-12 text-left">
      
      {/* Dashboard Header banner with AI Audit Button */}
      <div className="mb-6 flex flex-col md:flex-row items-start md:items-center justify-between border-b border-white/5 pb-4 gap-4">
        <div>
          <h2 className="text-sm font-bold text-white flex items-center uppercase tracking-widest" style={{ fontFamily: 'Georgia, serif' }}>
            <BarChart2 className="w-5 h-5 mr-2 text-[#C5A267]" />
            Analitik Evaluasi Mingguan Tim
          </h2>
          <p className="text-xs text-gray-400 mt-1 leading-snug">
            Metrik agregat progres perancangan, pengerjaan tugas kontributor, dan sisa backlog proses.
          </p>
        </div>

        {/* AI Audit Action */}
        <button
          onClick={handlePerformAudit}
          disabled={auditing || nodes.length === 0}
          className="flex items-center space-x-2 px-4 py-2.5 rounded-xl bg-[#C5A267] hover:bg-[#B59155] disabled:bg-white/5 disabled:text-gray-500 text-black text-xs font-bold font-mono tracking-wider transition duration-150 cursor-pointer disabled:cursor-not-allowed shadow-lg"
        >
          {auditing ? (
            <>
              <Loader2 className="w-3.5 h-3.5 animate-spin" />
              <span>MENGAUDIT ALUR...</span>
            </>
          ) : (
            <>
              <Sparkles className="w-3.5 h-3.5 animate-pulse text-black" />
              <span>AUDIT SPESIFIKASI VIA AI</span>
            </>
          )}
        </button>
      </div>

      {/* AI Spec Review / Audit Result Block */}
      {auditResult && (
        <div className="mb-6 border border-[#C5A267]/20 p-5 rounded-3xl bg-[#131315]/80 backdrop-blur-sm shadow-2xl relative">
          <div className="absolute right-4.5 top-4.5 flex items-center space-x-1.5 bg-[#C5A267] text-black font-mono text-[9px] font-bold px-2.5 py-1 rounded-xl shadow">
            <Sparkles className="w-3 h-3 text-black" />
            <span>AI AUDIT LAPORAN</span>
          </div>

          <div className="flex flex-col md:flex-row gap-6 items-start">
            {/* Score circle */}
            <div className="flex-shrink-0 flex flex-col items-center justify-center p-4 bg-white/2 border border-white/5 rounded-2xl w-28 h-28 text-center self-center md:self-start">
              <span className="text-[10px] font-bold font-mono text-[#C5A267]">QUALITY SCORE</span>
              <span className="text-3xl font-black text-white mt-1">{auditResult.score}</span>
              <span className="text-[9px] text-gray-500 font-bold">dari 100</span>
            </div>

            {/* Qualitative summary */}
            <div className="flex-1 space-y-3.5">
              <h3 className="text-white text-xs font-bold uppercase tracking-wider font-sans">Hasil Tinjauan Desain Fungsional</h3>
              <p className="text-gray-300 text-xs leading-relaxed font-sans">{auditResult.summary}</p>
              
              {/* Detailed checklist warnings */}
              {auditResult.issues.length > 0 ? (
                <div className="space-y-2.5 pt-2">
                  <span className="text-[10px] font-bold font-mono text-gray-400 uppercase tracking-widest block">Potensi Hambatan Desain & Keamanan ({auditResult.issues.length})</span>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-2 max-h-48 overflow-y-auto scrollbar-thin pr-1 pb-1">
                    {auditResult.issues.map((issue, idx) => (
                      <div key={idx} className={`p-3 border rounded-xl flex items-start space-x-2.5 text-xs inline-block ${getSeverityStyle(issue.severity)}`}>
                        {issue.severity === 'error' ? (
                          <AlertOctagon className="w-4 h-4 flex-shrink-0 mt-0.5 text-red-400" />
                        ) : issue.severity === 'warning' ? (
                          <AlertTriangle className="w-4 h-4 flex-shrink-0 mt-0.5 text-amber-400" />
                        ) : (
                          <Info className="w-4 h-4 flex-shrink-0 mt-0.5 text-sky-400" />
                        )}
                        <div className="space-y-0.5">
                          <p className="font-bold font-sans">
                            {issue.nodeName ? `[${issue.nodeName}] ` : ''}{issue.type}
                          </p>
                          <p className="text-[11px] text-gray-300 leading-normal font-sans">{issue.message}</p>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              ) : (
                <div className="text-xs text-emerald-400 bg-emerald-950/20 border border-emerald-900/40 p-3 rounded-xl flex items-center space-x-2 font-mono">
                  <Check className="w-4 h-4 text-emerald-400" />
                  <span>Rancangan Alur Sempurna! Tidak mendeteksi anomali keamanan maupun SLA bottleneck.</span>
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Evaluation Statement Banner */}
      <div className={`border p-4.5 rounded-2xl flex flex-col md:flex-row items-start md:items-center justify-between gap-4 mb-6 shadow-2xl ${badgeColor}`}>
        <div className="text-left font-sans">
          <div className="flex items-center space-x-2 text-[#C5A267]">
            <Award className="w-5 h-5 animate-pulse text-[#C5A267]" />
            <h4 className="text-xs font-bold uppercase tracking-wider font-sans">{evaluationTitle}</h4>
          </div>
          <p className="text-xs mt-1 max-w-2xl text-gray-300 leading-relaxed font-sans">{evaluationStatement}</p>
        </div>
        <div className="text-right flex-shrink-0">
          <span className="text-[9px] font-bold font-mono tracking-widest block text-[#C4A166]">NILAI EVALUASI</span>
          <span className="text-xl font-black font-sans text-white">{completionRate}% Selesai</span>
        </div>
      </div>

      {/* Bento Metric Charts Layout Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-6">
        
        {/* Progress distribution bar charts */}
        <div className="lg:col-span-2 bg-[#131315] border border-white/5 p-5 rounded-2xl shadow-xl text-left">
          <h4 className="text-[10px] font-bold text-gray-400 font-mono uppercase tracking-widest border-b border-white/5 pb-2 mb-4 flex items-center">
            <Clock className="w-4 h-4 mr-1.5 text-[#C5A267]" />
            Progres Distribusi Status Tugas
          </h4>

          <div className="space-y-4 font-sans">
            <div>
              <div className="flex justify-between text-xs font-semibold text-gray-300 mb-1">
                <span>Done (Artefak Selesai)</span>
                <span className="font-mono text-emerald-400">{doneTaskCount} / {totalTasks} ({completionRate}%)</span>
              </div>
              <div className="w-full bg-white/5 h-2.5 rounded-full overflow-hidden">
                <div style={{ width: `${completionRate}%` }} className="bg-emerald-500 h-full rounded-full transition-all duration-305" />
              </div>
            </div>

            <div>
              <div className="flex justify-between text-xs font-semibold text-gray-300 mb-1">
                <span>In Review (Peninjauan Kualitas)</span>
                <span className="font-mono text-purple-400">{reviewTaskCount} / {totalTasks} ({totalTasks > 0 ? Math.round((reviewTaskCount / totalTasks) * 100) : 0}%)</span>
              </div>
              <div className="w-full bg-white/5 h-2.5 rounded-full overflow-hidden">
                <div style={{ width: `${totalTasks > 0 ? (reviewTaskCount / totalTasks) * 100 : 0}%` }} className="bg-purple-500 h-full rounded-full transition-all duration-305" />
              </div>
            </div>

            <div>
              <div className="flex justify-between text-xs font-semibold text-gray-300 mb-1">
                <span>In Progress (Pemrograman/Sketsa)</span>
                <span className="font-mono text-blue-400">{inProgressTaskCount} / {totalTasks} ({totalTasks > 0 ? Math.round((inProgressTaskCount / totalTasks) * 100) : 0}%)</span>
              </div>
              <div className="w-full bg-white/5 h-2.5 rounded-full overflow-hidden">
                <div style={{ width: `${totalTasks > 0 ? (inProgressTaskCount / totalTasks) * 100 : 0}%` }} className="bg-blue-500 h-full rounded-full transition-all duration-305" />
              </div>
            </div>

            <div>
              <div className="flex justify-between text-xs font-semibold text-gray-300 mb-1">
                <span>Planned (Antrean Kerja)</span>
                <span className="font-mono text-gray-400">{plannedTaskCount} / {totalTasks} ({totalTasks > 0 ? Math.round((plannedTaskCount / totalTasks) * 100) : 0}%)</span>
              </div>
              <div className="w-full bg-white/5 h-2.5 rounded-full overflow-hidden">
                <div style={{ width: `${totalTasks > 0 ? (plannedTaskCount / totalTasks) * 100 : 0}%` }} className="bg-gray-600 h-full rounded-full transition-all duration-305" />
              </div>
            </div>
          </div>
        </div>

        {/* SLA and Complexity summaries */}
        <div className="bg-[#131315] border border-white/5 p-5 rounded-2xl shadow-xl text-left">
          <h4 className="text-[10px] font-bold text-gray-400 font-mono uppercase tracking-widest border-b border-white/5 pb-2 mb-4 flex items-center">
            <ShieldCheck className="w-4 h-4 mr-1.5 text-[#C5A267]" />
            Kerumitan Rencana (SLA)
          </h4>

          <div className="space-y-4 pt-1.5">
            <div className="flex items-center justify-between border-b border-white/5 pb-3">
              <div>
                <p className="text-[10px] font-bold text-gray-500 font-mono uppercase tracking-wider">ESTIMASI LEAD TIME ALUR</p>
                <p className="text-white text-base font-black tracking-tight">{estimatedLeadTime} Hari Kerja</p>
              </div>
              <span className="bg-white/5 border border-white/10 p-2 rounded-xl text-white font-mono text-[10px] font-bold">
                8h/Hari
              </span>
            </div>

            <div className="flex items-center justify-between">
              <div>
                <p className="text-[10px] font-bold text-gray-500 font-mono uppercase tracking-wider">KONTRAK ENDPOINT API</p>
                <p className="text-white text-base font-black tracking-tight">
                  {nodes.filter((n) => n.roles.backend?.endpoint).length} Endpoint
                </p>
              </div>
              <span className="bg-[#C5A267]/10 border border-[#C5A267]/20 p-2 rounded-xl text-[#C5A267] font-mono text-[10px] font-bold uppercase tracking-wider">
                API-Ready
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* Assignee Productivity Performance Chart/Directory */}
      <div className="bg-[#131315] border border-white/5 p-5 rounded-2xl shadow-xl text-left">
        <h4 className="text-[10px] font-bold text-gray-400 font-mono uppercase tracking-widest border-b border-white/5 pb-2 mb-4 flex items-center">
          <UserCheck className="w-4 h-4 mr-1.5 text-[#C5A267]" />
          Pembagian Beban Kerja & Produktivitas Anggota Tim
        </h4>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 pt-2">
          {Object.entries(
            modules.reduce((acc: Record<string, { done: number; total: number }>, m) => {
              m.nodes.forEach((n) => {
                const roles = ['uiux', 'frontend', 'backend'] as const;
                roles.forEach((r) => {
                  const facet = n.roles[r];
                  if (facet && facet.assignee) {
                    const name = facet.assignee;
                    if (!acc[name]) acc[name] = { done: 0, total: 0 };
                    acc[name].total++;
                    if (facet.status === 'done') acc[name].done++;
                  }
                });
              });
              return acc;
            }, {})
          ).map(([name, stat]) => {
            const pct = Math.round((stat.done / stat.total) * 100);
            return (
              <div key={name} className="border border-white/5 rounded-xl p-3.5 bg-white/2">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-xs font-bold text-white font-sans">{name}</span>
                  <span className="text-[10px] font-mono font-bold text-[#C5A267]">{stat.done}/{stat.total} Selesai</span>
                </div>
                <div className="flex items-center space-x-2.5 font-sans">
                  <div className="flex-1 bg-white/5 h-2 rounded-full overflow-hidden">
                    <div style={{ width: `${pct}%` }} className="bg-[#C5A267] h-full rounded-full transition-all duration-350" />
                  </div>
                  <span className="text-[10px] font-mono font-semibold text-gray-400">{pct}%</span>
                </div>
              </div>
            );
          })}

          {Object.keys(
            modules.reduce((acc: Record<string, { done: number; total: number }>, m) => {
              m.nodes.forEach((n) => {
                ['uiux', 'frontend', 'backend'].forEach((r) => {
                  const facet = n.roles[r as 'uiux' | 'frontend' | 'backend'];
                  if (facet && facet.assignee) acc[facet.assignee] = { done: 0, total: 0 };
                });
              });
              return acc;
            }, {})
          ).length === 0 && (
            <div className="col-span-full py-8 text-center text-xs text-gray-500 font-sans">
              Belum ada tugas yang dikaitkan ke desainer/pemrogram mana pun. Buka kanvas, gunakan Inspector, dan tentukan penanggung jawab.
            </div>
          )}
        </div>
      </div>

    </div>
  );
}
