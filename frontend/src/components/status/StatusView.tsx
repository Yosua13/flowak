import React from 'react';
import { CircleDot, LayoutGrid, Loader2, TriangleAlert } from 'lucide-react';
import { useStore } from '../../store/useStore';
import { useDerivedViewData } from '../../services/derivedViewData';

const readinessFacets = ['uiux', 'frontend', 'backend'] as const;
const activeStatuses = ['Ready', 'In Progress', 'In Review', 'Blocked'] as const;

export default function StatusView() {
  const { modules, activeId, activeProjectId, selectNode, setView } = useStore();
  const module = modules.find((entry) => entry.id === activeId);
  const { data, loading, error } = useDerivedViewData(activeProjectId, module?.id);
  if (!module) return <Empty message="Pilih modul untuk melihat status kesiapan dan pekerjaan." />;
  if (loading) return <Empty message="Memuat status work item..." loading />;
  if (error) return <Empty message={`Status belum dapat dimuat: ${error}`} error />;
  const items = data?.work_items || [];
  const facets = module.nodes.flatMap((node) => readinessFacets.map((key) => ({ node, key, value: node.roles[key] })).filter((entry) => entry.value));
  const complete = facets.filter((entry) => entry.value?.readiness === 'done' || entry.value?.status === 'done').length;
  const done = items.filter((item) => item.status === 'Done').length;
  const active = items.filter((item) => activeStatuses.includes(item.status as typeof activeStatuses[number])).length;
  return <main className="flex-1 overflow-y-auto bg-[#0A0A0B] p-6">
    <header className="mb-6"><h1 className="flex items-center gap-2 text-lg font-bold text-white"><LayoutGrid className="h-5 w-5 text-[#C5A267]" />Status delivery</h1><p className="mt-1 text-sm text-gray-400">Kesiapan spesifikasi dan progres pekerjaan adalah dua sinyal yang berbeda.</p></header>
    <section className="mb-6 grid gap-3 md:grid-cols-4"><Metric label="Kesiapan facet" value={facets.length ? `${Math.round((complete / facets.length) * 100)}%` : '-'} detail={`${complete}/${facets.length} artefak siap`} /><Metric label="Work item selesai" value={`${done}/${items.length}`} detail="berdasarkan status work item" /><Metric label="Pekerjaan aktif" value={String(active)} detail="Ready, in progress, review, blocked" /><Metric label="Tertunda" value={String(items.filter((item) => item.status === 'Blocked').length)} detail="memerlukan tindak lanjut" /></section>
    <section className="overflow-hidden rounded-lg border border-white/10 bg-[#131315]"><div className="border-b border-white/10 p-4"><h2 className="text-sm font-semibold text-white">Matriks readiness spesifikasi</h2><p className="mt-1 text-xs text-gray-500">Readiness tidak dihitung sebagai task atau progres Kanban.</p></div><table className="w-full text-left text-xs"><thead className="bg-black/20 text-gray-500"><tr><th className="p-3">Node</th><th className="p-3">UI/UX</th><th className="p-3">Frontend</th><th className="p-3">Backend</th></tr></thead><tbody>{module.nodes.map((node) => <tr key={node.id} onClick={() => { selectNode(node.id); setView('canvas'); }} className="cursor-pointer border-t border-white/5 hover:bg-white/[.03]"><td className="p-3 font-medium text-white">{node.label}</td>{readinessFacets.map((key) => <td key={key} className="p-3"><Readiness value={node.roles[key]?.readiness || node.roles[key]?.status} /></td>)}</tr>)}</tbody></table>{module.nodes.length === 0 && <p className="p-6 text-center text-sm text-gray-500">Belum ada node pada modul ini.</p>}</section>
    <section className="mt-6 overflow-hidden rounded-lg border border-white/10 bg-[#131315]"><div className="border-b border-white/10 p-4"><h2 className="text-sm font-semibold text-white">Progres work item</h2><p className="mt-1 text-xs text-gray-500">Data ini sama dengan kartu pada Kanban.</p></div>{items.length ? <div className="divide-y divide-white/5">{items.map((item) => <div key={item.id} className="flex flex-wrap items-center gap-3 p-3 text-xs"><span className="font-mono text-[#C5A267]">{item.key}</span><span className="min-w-48 flex-1 text-gray-200">{item.title}</span><Badge value={item.status} /><span className="text-gray-500">{item.due_date ? `Due ${item.due_date.slice(0, 10)}` : 'Tanpa deadline'}</span></div>)}</div> : <p className="p-6 text-center text-sm text-gray-500">Belum ada work item untuk filter modul ini.</p>}</section>
  </main>;
}
function Metric({ label, value, detail }: { label: string; value: string; detail: string }) { return <div className="rounded-lg border border-white/10 bg-[#131315] p-4"><p className="text-xs text-gray-500">{label}</p><p className="mt-1 text-2xl font-bold text-white">{value}</p><p className="mt-1 text-xs text-gray-500">{detail}</p></div>; }
function Readiness({ value }: { value?: string }) { return <span className={`rounded px-2 py-1 font-mono text-[10px] ${value === 'done' ? 'bg-emerald-500/15 text-emerald-300' : value === 'review' ? 'bg-violet-500/15 text-violet-300' : 'bg-white/5 text-gray-400'}`}>{(value || 'belum ada').toUpperCase()}</span>; }
function Badge({ value }: { value: string }) { return <span className={`rounded px-2 py-1 font-mono text-[10px] ${value === 'Done' ? 'bg-emerald-500/15 text-emerald-300' : value === 'Blocked' ? 'bg-rose-500/15 text-rose-300' : 'bg-[#C5A267]/15 text-[#DCC182]'}`}>{value}</span>; }
function Empty({ message, loading, error }: { message: string; loading?: boolean; error?: boolean }) { const Icon = loading ? Loader2 : error ? TriangleAlert : CircleDot; return <main className="flex flex-1 items-center justify-center bg-[#0A0A0B] text-sm text-gray-400"><Icon className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} />{message}</main>; }
