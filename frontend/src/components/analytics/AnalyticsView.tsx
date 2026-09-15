import React, { useMemo } from 'react';
import { Activity, BarChart3, Clock3, Loader2, TriangleAlert } from 'lucide-react';
import { useStore } from '../../store/useStore';
import { calculateDerivedMetrics } from '../../services/derivedMetrics';
import { useDerivedViewData } from '../../services/derivedViewData';

export default function AnalyticsView() {
  const { modules, activeId, activeProjectId } = useStore();
  const module = modules.find((entry) => entry.id === activeId);
  const { data, loading, error } = useDerivedViewData(activeProjectId, module?.id);
  const metrics = useMemo(() => module && data ? calculateDerivedMetrics(data.work_items, data.history, module) : null, [data, module]);
  if (!module) return <State message="Pilih modul untuk melihat analitik delivery." />;
  if (loading) return <State message="Memuat history pekerjaan..." loading />;
  if (error) return <State message={`Analitik belum dapat dimuat: ${error}`} error />;
  if (!metrics || !data?.work_items.length) return <main className="flex flex-1 items-center justify-center bg-[#0A0A0B] text-sm text-gray-500">Belum ada work item untuk dihitung. Tambahkan work item dari Kanban.</main>;
  const percent = (value: number) => `${Math.round(value * 100)}%`;
  return <main className="flex-1 overflow-y-auto bg-[#0A0A0B] p-6"><header className="mb-6"><h1 className="flex items-center gap-2 text-lg font-bold text-white"><BarChart3 className="h-5 w-5 text-[#C5A267]" />Analitik delivery</h1><p className="mt-1 text-sm text-gray-400">Metrik agregat dari status history dan traceability work item. Tidak ada ranking individu.</p></header><section className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4"><Metric label="Throughput" value={`${metrics.throughput} selesai`} hint="transisi ke Done" /><Metric label="Cycle time" value={hours(metrics.cycleTimeHours)} hint="Backlog hingga Done" /><Metric label="Blocked time" value={hours(metrics.blockedTimeHours)} hint="durasi status Blocked" /><Metric label="Aging WIP" value={hours(metrics.agingWipHours)} hint="rata-rata pekerjaan aktif" /><Metric label="Overdue" value={String(metrics.overdue)} hint="due date terlewati" danger={metrics.overdue > 0} /><Metric label="Review rework" value={String(metrics.reviewRework)} hint="In Review kembali ke In Progress" /><Metric label="Kelengkapan" value={percent(metrics.completeness)} hint="readiness facet terisi" /><Metric label="Traceability" value={percent(metrics.traceability)} hint="work item terkait node" /></section><section className="mt-6 rounded-lg border border-white/10 bg-[#131315] p-4"><h2 className="flex items-center gap-2 text-sm font-semibold text-white"><Activity className="h-4 w-4 text-[#C5A267]" />Cara membaca</h2><p className="mt-2 text-sm leading-6 text-gray-400">Cycle time dan blocked time dibangun dari urutan perubahan status yang tercatat, bukan dari status readiness facet atau parsing SLA teks bebas. Nilai kosong berarti belum ada history yang cukup untuk menghitung metrik tersebut.</p></section></main>;
}
function hours(value: number | null) { return value === null ? 'Belum cukup data' : `${value.toFixed(value >= 10 ? 0 : 1)} jam`; }
function Metric({ label, value, hint, danger }: { label: string; value: string; hint: string; danger?: boolean }) { return <div className="rounded-lg border border-white/10 bg-[#131315] p-4"><p className="text-xs text-gray-500">{label}</p><p className={`mt-1 text-xl font-bold ${danger ? 'text-rose-300' : 'text-white'}`}>{value}</p><p className="mt-1 text-xs text-gray-500">{hint}</p></div>; }
function State({ message, loading, error }: { message: string; loading?: boolean; error?: boolean }) { const Icon = loading ? Loader2 : error ? TriangleAlert : Clock3; return <main className="flex flex-1 items-center justify-center bg-[#0A0A0B] text-sm text-gray-400"><Icon className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} />{message}</main>; }
