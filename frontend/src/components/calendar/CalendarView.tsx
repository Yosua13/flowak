import React, { useMemo, useState } from 'react';
import { CalendarDays, ChevronLeft, ChevronRight, Loader2, TriangleAlert } from 'lucide-react';
import { useStore } from '../../store/useStore';
import { useDerivedViewData } from '../../services/derivedViewData';

type EventKind = 'start' | 'due' | 'review' | 'baseline';
interface CalendarEvent { id: string; date: string; title: string; detail: string; kind: EventKind; }
const monthNames = ['Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni', 'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'];
const colors: Record<EventKind, string> = { start: 'border-sky-400 text-sky-200', due: 'border-amber-400 text-amber-200', review: 'border-violet-400 text-violet-200', baseline: 'border-emerald-400 text-emerald-200' };

export default function CalendarView() {
  const { modules, activeId, activeProjectId } = useStore();
  const module = modules.find((entry) => entry.id === activeId);
  const { data, loading, error } = useDerivedViewData(activeProjectId, module?.id);
  const today = new Date(); const [cursor, setCursor] = useState(new Date(today.getFullYear(), today.getMonth(), 1));
  const events = useMemo<CalendarEvent[]>(() => {
    if (!data) return [];
    return [
      ...data.work_items.flatMap((item) => [item.start_date && { id: `${item.id}-start`, date: item.start_date, title: item.title, detail: `${item.key} · mulai`, kind: 'start' as const }, item.due_date && { id: `${item.id}-due`, date: item.due_date, title: item.title, detail: `${item.key} · ${item.status}`, kind: 'due' as const }].filter(Boolean) as CalendarEvent[]),
      ...data.facet_reviews.map((review) => ({ id: `${review.node_id}-${review.role_key}`, date: review.due_date, title: review.node_label, detail: `${review.role_key.toUpperCase()} · review ${review.readiness}`, kind: 'review' as const })),
      ...data.baselines.map((baseline) => ({ id: `${baseline.module_id}-${baseline.version}`, date: baseline.created_at, title: `Baseline v${baseline.version}`, detail: 'Versi modul tercatat', kind: 'baseline' as const })),
    ];
  }, [data]);
  if (!module) return <State message="Pilih modul untuk melihat jadwal." />;
  if (loading) return <State message="Memuat event jadwal..." loading />;
  if (error) return <State message={`Kalender belum dapat dimuat: ${error}`} error />;
  const year = cursor.getFullYear(), month = cursor.getMonth(), offset = new Date(year, month, 1).getDay(), days = new Date(year, month + 1, 0).getDate();
  const dayEvents = (day: number) => events.filter((event) => { const date = new Date(event.date); return date.getFullYear() === year && date.getMonth() === month && date.getDate() === day; });
  return <main className="flex-1 overflow-y-auto bg-[#0A0A0B] p-6"><header className="mb-6 flex flex-wrap items-center justify-between gap-3"><div><h1 className="flex items-center gap-2 text-lg font-bold text-white"><CalendarDays className="h-5 w-5 text-[#C5A267]" />Kalender delivery</h1><p className="mt-1 text-sm text-gray-400">Start/due work item, review facet, dan baseline versi memakai data proyek yang sama.</p></div><div className="flex items-center gap-2"><button aria-label="Bulan sebelumnya" onClick={() => setCursor(new Date(year, month - 1, 1))} className="rounded border border-white/10 p-2"><ChevronLeft className="h-4 w-4" /></button><span className="w-40 text-center text-sm font-semibold text-white">{monthNames[month]} {year}</span><button aria-label="Bulan berikutnya" onClick={() => setCursor(new Date(year, month + 1, 1))} className="rounded border border-white/10 p-2"><ChevronRight className="h-4 w-4" /></button></div></header><section className="overflow-hidden rounded-lg border border-white/10 bg-[#131315]"><div className="grid grid-cols-7 border-b border-white/10 text-center text-xs text-gray-500">{['Min','Sen','Sel','Rab','Kam','Jum','Sab'].map((day) => <div key={day} className="p-3">{day}</div>)}</div><div className="grid grid-cols-7">{Array.from({ length: offset + days }, (_, index) => { const day = index - offset + 1; if (day < 1) return <div key={index} className="min-h-28 border-b border-r border-white/5 bg-black/10" />; return <div key={day} className="min-h-28 border-b border-r border-white/5 p-2"><span className="text-xs text-gray-400">{day}</span><div className="mt-1 space-y-1">{dayEvents(day).map((event) => <div key={event.id} title={`${event.title}: ${event.detail}`} className={`truncate border-l-2 bg-white/[.03] px-1 py-0.5 text-[10px] ${colors[event.kind]}`}>{event.kind === 'start' ? 'Mulai' : event.kind === 'due' ? 'Due' : event.kind === 'review' ? 'Review' : 'Baseline'} · {event.title}</div>)}</div></div>; })}</div></section><section className="mt-4 flex flex-wrap gap-4 text-xs text-gray-400">{(Object.keys(colors) as EventKind[]).map((kind) => <span key={kind} className={`border-l-2 pl-2 ${colors[kind]}`}>{kind === 'start' ? 'Mulai task' : kind === 'due' ? 'Deadline task' : kind === 'review' ? 'Review facet' : 'Baseline versi'}</span>)}</section>{events.length === 0 && <p className="mt-6 rounded-lg border border-dashed border-white/10 p-6 text-center text-sm text-gray-500">Belum ada tanggal work item, review facet, atau baseline pada modul ini.</p>}</main>;
}
function State({ message, loading, error }: { message: string; loading?: boolean; error?: boolean }) { const Icon = loading ? Loader2 : error ? TriangleAlert : CalendarDays; return <main className="flex flex-1 items-center justify-center bg-[#0A0A0B] text-sm text-gray-400"><Icon className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} />{message}</main>; }
