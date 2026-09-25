import React, { useEffect, useMemo, useRef, useState } from 'react';
import { AlertTriangle, ArrowUpRight, CheckCircle2, ChevronDown, CircleAlert, ClipboardList, ExternalLink, Loader2, Maximize2, MessageSquare, Minimize2, Plus, RefreshCw, Send, X } from 'lucide-react';
import type { Module, Node, NodeComment, NodeType, WorkItem } from '../../domain/types';
import { NODE_TYPES } from '../../config/nodeTypes';
import { STATUS_CONFIG } from '../../config/status';
import { useStore } from '../../store/useStore';
import BackendTab from './BackendTab';
import BisnisTab from './BisnisTab';
import FrontendTab from './FrontendTab';
import UiuxTab from './UiuxTab';
import { DetailTab, detailTabs, facetReadiness, isOpenWorkItem, isOverdue, nextPaths, nodeCompleteness, nodeRisk, nodeWorkItems } from './nodeDetail';

type SurfaceMode = 'drawer' | 'modal' | 'page';

interface NodeDetailSurfaceProps {
  node: Node;
  module: Module;
  mode: SurfaceMode;
  onClose: () => void;
  onModeChange: (mode: SurfaceMode) => void;
}

const saveLabels = {
  idle: 'Belum ada perubahan', saving: 'Menyimpan', saved: 'Tersimpan', offline: 'Offline', failed: 'Gagal', conflict: 'Konflik',
} as const;

export default function NodeDetailSurface({ node, module, mode, onClose, onModeChange }: NodeDetailSurfaceProps) {
  const { activeProjectId, token, saveStatus, saveError, updateNode, retryActiveModuleSave, reloadActiveProject } = useStore();
  const [tab, setTab] = useState<DetailTab>('summary');
  const [items, setItems] = useState<WorkItem[]>([]);
  const [comments, setComments] = useState<NodeComment[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [newTask, setNewTask] = useState('');
  const [newComment, setNewComment] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [showMenu, setShowMenu] = useState(false);
  const originRef = useRef<HTMLElement | null>(null);
  const surfaceRef = useRef<HTMLElement | null>(null);

  const headers = useMemo(() => token ? { Authorization: `Bearer ${token}` } : {}, [token]);
  const nodeItems = useMemo(() => nodeWorkItems(items, node.id), [items, node.id]);
  const openItems = useMemo(() => nodeItems.filter(isOpenWorkItem), [nodeItems]);
  const overdueItems = useMemo(() => openItems.filter((item) => isOverdue(item)), [openItems]);
  const readiness = (['uiux', 'frontend', 'backend'] as const).map((facet) => ({ facet, value: facetReadiness(node, facet) }));

  const refresh = async () => {
    if (!activeProjectId || !token) {
      setLoading(false);
      return;
    }
    setLoading(true);
    setError('');
    try {
      const [workResponse, commentResponse] = await Promise.all([
        fetch(`/api/projects/${activeProjectId}/work-items?node_id=${encodeURIComponent(node.id)}`, { headers }),
        fetch(`/api/nodes/${node.id}/comments`, { headers }),
      ]);
      if (!workResponse.ok || !commentResponse.ok) throw new Error('Data detail tidak dapat dimuat.');
      setItems(await workResponse.json());
      setComments(await commentResponse.json());
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : 'Data detail tidak dapat dimuat.');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    originRef.current = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    void refresh();
    // The node identity is the resource boundary for all detail data.
  }, [node.id, activeProjectId, token]);

  useEffect(() => () => originRef.current?.focus(), []);

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        event.preventDefault();
        close();
        return;
      }
      if (event.key !== 'Tab' || mode === 'drawer' || !surfaceRef.current) return;
      const focusable = Array.from(surfaceRef.current.querySelectorAll<HTMLElement>('button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), a[href]')) as HTMLElement[];
      if (!focusable.length) return;
      const first = focusable[0];
      const last = focusable[focusable.length - 1];
      if (event.shiftKey && document.activeElement === first) {
        event.preventDefault();
        last.focus();
      } else if (!event.shiftKey && document.activeElement === last) {
        event.preventDefault();
        first.focus();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [mode]);

  const close = () => onClose();
  const changeLabel = (event: React.ChangeEvent<HTMLInputElement>) => updateNode(node.id, { label: event.target.value });
  const changeType = (event: React.ChangeEvent<HTMLSelectElement>) => updateNode(node.id, { type: event.target.value as NodeType });

  const createTask = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!activeProjectId || !newTask.trim()) return;
    setSubmitting(true);
    try {
      const response = await fetch(`/api/projects/${activeProjectId}/work-items`, {
        method: 'POST', headers: { ...headers, 'Content-Type': 'application/json' },
        body: JSON.stringify({ module_id: module.id, node_id: node.id, type: 'Task', title: newTask.trim(), priority: 'medium' }),
      });
      if (!response.ok) throw new Error('Task tidak dapat dibuat.');
      setNewTask('');
      await refresh();
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : 'Task tidak dapat dibuat.');
    } finally { setSubmitting(false); }
  };

  const createComment = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!newComment.trim()) return;
    setSubmitting(true);
    try {
      const response = await fetch(`/api/nodes/${node.id}/comments`, {
        method: 'POST', headers: { ...headers, 'Content-Type': 'application/json' }, body: JSON.stringify({ body: newComment.trim() }),
      });
      if (!response.ok) throw new Error('Komentar tidak dapat dikirim.');
      setNewComment('');
      await refresh();
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : 'Komentar tidak dapat dikirim.');
    } finally { setSubmitting(false); }
  };

  const panelClass = mode === 'drawer'
    ? 'h-full w-full min-w-[min(100vw,420px)] max-w-[640px] border-l'
    : mode === 'modal'
      ? 'w-full max-w-5xl h-[min(88vh,900px)] rounded-2xl border shadow-2xl'
      : 'fixed inset-0 z-[100] h-screen w-screen border-0';

  const content = () => {
    if (loading) return <LoadingState />;
    if (error) return <ErrorState message={error} onRetry={refresh} />;
    if (tab === 'summary') return <Summary node={node} module={module} readiness={readiness} openItems={openItems} overdueItems={overdueItems} onTasks={() => setTab('tasks')} />;
    if (tab === 'business') return <BisnisTab node={node} />;
    if (tab === 'uiux') return <UiuxTab node={node} />;
    if (tab === 'frontend') return <FrontendTab node={node} />;
    if (tab === 'backend') return <BackendTab node={node} />;
    if (tab === 'tasks') return <TaskSection items={nodeItems} value={newTask} submitting={submitting} onChange={setNewTask} onSubmit={createTask} />;
    if (tab === 'discussion') return <DiscussionSection comments={comments} value={newComment} submitting={submitting} onChange={setNewComment} onSubmit={createComment} />;
    return <ActivitySection />;
  };

  return (
    <div className={mode === 'modal' ? 'fixed inset-0 z-[100] flex items-center justify-center bg-black/65 p-4' : mode === 'page' ? '' : 'h-full'} role={mode === 'modal' ? 'dialog' : undefined} aria-modal={mode === 'modal' || mode === 'page' ? true : undefined} aria-label={`Detail ${node.label}`}>
      <section ref={surfaceRef} className={`${panelClass} flex flex-col overflow-hidden bg-[#111113] border-white/10 text-left`}>
        <header className="sticky top-0 z-20 border-b border-white/10 bg-[#111113]/95 p-4 backdrop-blur">
          <div className="flex gap-3">
            <div className="min-w-0 flex-1 space-y-2">
              <div className="flex flex-wrap items-center gap-2 text-[10px] font-mono">
                <select aria-label="Tipe node" value={node.type} onChange={changeType} className="rounded-md border border-white/10 bg-[#1A1A1D] px-2 py-1 text-gray-200">
                  {Object.entries(NODE_TYPES).map(([value, config]) => <option key={value} value={value}>{config.label}</option>)}
                </select>
                <span className="rounded-md border border-[#C5A267]/30 bg-[#C5A267]/10 px-2 py-1 text-[#E2C392]">{node.id}</span>
                <span className="text-gray-400">Owner: {node.doc.actor || node.roles.uiux?.assignee || node.roles.frontend?.assignee || node.roles.backend?.assignee || 'Belum ditentukan'}</span>
              </div>
              <input aria-label="Nama node" value={node.label} onChange={changeLabel} className="w-full border-b border-transparent bg-transparent pb-1 text-lg font-bold text-white outline-none focus:border-[#C5A267]" />
              <div className="flex flex-wrap gap-2 text-[10px]">
                <Badge label={`Risiko: ${nodeRisk(node)}`} tone={nodeRisk(node) === 'high' || nodeRisk(node) === 'critical' ? 'danger' : 'neutral'} />
                <Badge label={`Lengkap ${nodeCompleteness(node)}%`} tone="neutral" />
                <Badge label={`${openItems.length} task terbuka`} tone={openItems.length ? 'warning' : 'success'} />
                <Badge label={saveLabels[saveStatus]} tone={saveStatus === 'conflict' || saveStatus === 'failed' ? 'danger' : saveStatus === 'offline' ? 'warning' : saveStatus === 'saved' ? 'success' : 'neutral'} />
              </div>
            </div>
            <div className="flex h-fit items-center gap-1">
              <div className="relative">
                <button onClick={() => setShowMenu((value) => !value)} className="rounded-lg p-2 text-gray-400 hover:bg-white/5 hover:text-white" aria-label="Menu detail"><ChevronDown className="h-4 w-4" /></button>
                {showMenu && <div className="absolute right-0 top-9 z-30 w-48 rounded-xl border border-white/10 bg-[#1A1A1D] p-1 text-xs shadow-xl">
                  <button onClick={() => { onModeChange('modal'); setShowMenu(false); }} className="w-full rounded-lg px-3 py-2 text-left hover:bg-white/5">Pratinjau modal</button>
                  <button onClick={() => { onModeChange('page'); setShowMenu(false); }} className="w-full rounded-lg px-3 py-2 text-left hover:bg-white/5">Buka halaman penuh</button>
                </div>}
              </div>
              <button onClick={() => onModeChange(mode === 'page' ? 'drawer' : 'page')} className="rounded-lg p-2 text-gray-400 hover:bg-white/5 hover:text-white" title="Buka halaman penuh">{mode === 'page' ? <Minimize2 className="h-4 w-4" /> : <Maximize2 className="h-4 w-4" />}</button>
              <button onClick={close} className="rounded-lg p-2 text-gray-400 hover:bg-white/5 hover:text-white" aria-label="Tutup detail"><X className="h-5 w-5" /></button>
            </div>
          </div>
          {saveStatus === 'conflict' && <div className="mt-3 flex flex-wrap items-center gap-2 rounded-lg border border-rose-500/30 bg-rose-950/25 p-2 text-xs text-rose-100"><CircleAlert className="h-4 w-4" /> Versi graph berubah di server. <button onClick={() => void reloadActiveProject()} className="underline">Muat ulang</button><button onClick={() => void retryActiveModuleSave()} className="underline">Coba lagi</button><span className="text-rose-200/70">Muat ulang untuk membandingkan versi server sebelum melanjutkan.</span></div>}
          {saveStatus === 'failed' && saveError && <p role="alert" className="mt-3 rounded-lg border border-rose-500/30 bg-rose-950/25 p-2 text-xs text-rose-100">{saveError}</p>}
          {node.legacyNotes && Object.keys(node.legacyNotes).length > 0 && <div className="mt-3 rounded-lg border border-amber-500/30 bg-amber-950/20 p-2 text-xs text-amber-100"><strong>Nilai legacy perlu ditinjau:</strong> {Object.entries(node.legacyNotes).map(([field, note]) => <span key={field} className="ml-2">{field}: {note}</span>)}</div>}
        </header>
        <nav className="flex shrink-0 gap-1 overflow-x-auto border-b border-white/10 bg-[#0A0A0B] p-2" aria-label="Bagian detail node">
          {detailTabs().map((item) => <button key={item.id} onClick={() => setTab(item.id)} className={`whitespace-nowrap rounded-lg px-3 py-2 text-xs font-semibold ${tab === item.id ? 'bg-[#C5A267]/15 text-[#E2C392]' : 'text-gray-400 hover:bg-white/5 hover:text-white'}`}>{item.label}</button>)}
        </nav>
        <main className="min-h-0 flex-1 overflow-y-auto p-4 scrollbar-thin">{content()}</main>
      </section>
    </div>
  );
}

function Badge({ label, tone }: { label: string; tone: 'neutral' | 'success' | 'warning' | 'danger' }) {
  const colors = { neutral: 'border-white/10 bg-white/5 text-gray-300', success: 'border-emerald-400/25 bg-emerald-500/10 text-emerald-300', warning: 'border-amber-400/25 bg-amber-500/10 text-amber-200', danger: 'border-rose-400/25 bg-rose-500/10 text-rose-200' };
  return <span className={`rounded-full border px-2 py-1 ${colors[tone]}`}>{label}</span>;
}

function Summary({ node, module, readiness, openItems, overdueItems, onTasks }: { node: Node; module: Module; readiness: Array<{ facet: 'uiux' | 'frontend' | 'backend'; value?: string }>; openItems: WorkItem[]; overdueItems: WorkItem[]; onTasks: () => void }) {
  const paths = nextPaths(node, module.edges, module.nodes);
  return <div className="space-y-5">
    <section className="rounded-xl border border-white/10 bg-[#131315] p-4"><h2 className="text-sm font-bold text-white">Ringkasan kesiapan</h2><p className="mt-2 text-sm leading-relaxed text-gray-300">{node.doc.outcome || node.doc.output || 'Outcome belum didokumentasikan.'}</p><dl className="mt-4 grid gap-3 sm:grid-cols-2"><Info label="Aktor" value={node.doc.actor} /><Info label="Pemicu" value={node.doc.trigger} /><Info label="Proses" value={node.doc.process} /><Info label="SLA" value={node.doc.sla} /></dl></section>
    <section className="grid gap-3 md:grid-cols-3">{readiness.map(({ facet, value }) => <div key={facet} className="rounded-xl border border-white/10 bg-[#131315] p-3"><p className="text-[10px] font-mono uppercase text-gray-500">{facet}</p><p className="mt-1 text-sm font-semibold text-white">{value ? STATUS_CONFIG[value as keyof typeof STATUS_CONFIG]?.label || value : 'Belum siap'}</p></div>)}</section>
    <section className="rounded-xl border border-white/10 bg-[#131315] p-4"><div className="flex items-center justify-between"><h2 className="text-sm font-bold text-white">Pekerjaan aktif</h2><button onClick={onTasks} className="text-xs text-[#E2C392] hover:underline">Lihat task</button></div><p className="mt-2 text-sm text-gray-300">{openItems.length ? `${openItems.length} task terbuka${overdueItems.length ? `, ${overdueItems.length} melewati tenggat` : ''}.` : 'Tidak ada task terbuka.'}</p></section>
    <section className="rounded-xl border border-white/10 bg-[#131315] p-4"><h2 className="text-sm font-bold text-white">Jalur berikutnya</h2>{paths.length ? <ul className="mt-2 space-y-2">{paths.map((path) => <li key={path.id} className="flex items-center gap-2 text-sm text-gray-300"><ArrowUpRight className="h-4 w-4 text-[#C5A267]" />{path.label}{path.edgeLabel && <span className="text-xs text-gray-500">({path.edgeLabel})</span>}</li>)}</ul> : <p className="mt-2 text-sm text-gray-500">Belum ada jalur keluar.</p>}</section>
  </div>;
}

function Info({ label, value }: { label: string; value?: string }) { return <div><dt className="text-[10px] font-mono uppercase text-gray-500">{label}</dt><dd className="mt-1 text-sm text-gray-200">{value || 'Belum diisi'}</dd></div>; }
function LoadingState() { return <div className="flex min-h-48 items-center justify-center gap-2 text-sm text-gray-400"><Loader2 className="h-4 w-4 animate-spin" /> Memuat detail node…</div>; }
function ErrorState({ message, onRetry }: { message: string; onRetry: () => void }) { return <div className="rounded-xl border border-rose-500/30 bg-rose-950/20 p-4 text-sm text-rose-100"><AlertTriangle className="mb-2 h-5 w-5" />{message}<button onClick={onRetry} className="ml-3 underline">Coba lagi</button></div>; }
function TaskSection({ items, value, submitting, onChange, onSubmit }: { items: WorkItem[]; value: string; submitting: boolean; onChange: (value: string) => void; onSubmit: (event: React.FormEvent) => void }) { return <div className="space-y-4"><form onSubmit={onSubmit} className="flex gap-2"><input value={value} onChange={(event) => onChange(event.target.value)} placeholder="Tambahkan task untuk node ini" className="min-w-0 flex-1 rounded-lg border border-white/10 bg-[#1A1A1D] px-3 py-2 text-sm text-white outline-none focus:border-[#C5A267]" /><button disabled={submitting || !value.trim()} className="rounded-lg bg-[#C5A267] px-3 text-black disabled:opacity-50" aria-label="Tambah task"><Plus className="h-4 w-4" /></button></form>{items.length ? <ul className="space-y-2">{items.map((item) => <li key={item.id} className="flex items-center justify-between rounded-xl border border-white/10 bg-[#131315] p-3"><div><p className="text-sm font-semibold text-white">{item.key} · {item.title}</p><p className="mt-1 text-xs text-gray-400">{item.status}{item.due_date && ` · Tenggat ${item.due_date}`}</p></div><ClipboardList className="h-4 w-4 text-[#C5A267]" /></li>)}</ul> : <p className="rounded-xl border border-dashed border-white/15 p-5 text-sm text-gray-500">Belum ada task yang ditautkan ke node ini.</p>}</div>; }
function DiscussionSection({ comments, value, submitting, onChange, onSubmit }: { comments: NodeComment[]; value: string; submitting: boolean; onChange: (value: string) => void; onSubmit: (event: React.FormEvent) => void }) { return <div className="space-y-4"><form onSubmit={onSubmit} className="space-y-2"><label className="text-sm font-semibold text-white">Diskusi node</label><textarea value={value} onChange={(event) => onChange(event.target.value)} placeholder="Tulis komentar…" className="min-h-24 w-full rounded-lg border border-white/10 bg-[#1A1A1D] p-3 text-sm text-white outline-none focus:border-[#C5A267]" /><button disabled={submitting || !value.trim()} className="flex items-center gap-2 rounded-lg bg-[#C5A267] px-3 py-2 text-sm font-semibold text-black disabled:opacity-50"><Send className="h-4 w-4" />Kirim komentar</button></form>{comments.length ? <ul className="space-y-3">{comments.map((comment) => <li key={comment.id} className="rounded-xl border border-white/10 bg-[#131315] p-3"><p className="text-sm text-gray-200">{comment.body}</p><p className="mt-2 text-[10px] font-mono text-gray-500">{comment.author_id} · {new Date(comment.created_at).toLocaleString()}</p></li>)}</ul> : <p className="rounded-xl border border-dashed border-white/15 p-5 text-sm text-gray-500">Belum ada komentar pada node ini.</p>}</div>; }
function ActivitySection() { return <div className="rounded-xl border border-white/10 bg-[#131315] p-5 text-sm text-gray-400"><CheckCircle2 className="mb-2 h-5 w-5 text-[#C5A267]" />Aktivitas node belum tersedia dari API. Surface ini tidak menampilkan status real-time atau riwayat sintetis.</div>; }
