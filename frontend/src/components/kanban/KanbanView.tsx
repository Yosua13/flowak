import React, { useEffect, useMemo, useState } from 'react';
import { createPortal } from 'react-dom';
import { AlertTriangle, Archive, CalendarDays, ChevronLeft, ChevronRight, CircleUserRound, ClipboardList, Filter, GripVertical, MessageSquare, Paperclip, Plus, Search, Send, X, Link2, Check, ArrowUpRight, RotateCcw } from 'lucide-react';
import type { WorkItem, WorkItemStatus, WorkItemType } from '../../domain/types';
import { useStore } from '../../store/useStore';
import { workItemsApi, type WorkItemActivity, type WorkItemInput } from '../../services/workItems';
import { dialog } from '../../services/dialog';
import { artifactCountMap, BOARD_STATUSES, canTransitionWorkItem, columnMetrics, filterWorkItems, focusTrapIndex, itemByKey, keyboardTransitionTarget, optimisticStatus, type KanbanFilters } from './kanbanModel';
import WorkItemArtifacts from './WorkItemArtifacts';

const statusTone: Record<WorkItemStatus, string> = { Backlog: 'border-gray-500', Ready: 'border-cyan-500', 'In Progress': 'border-blue-500', 'In Review': 'border-violet-500', Blocked: 'border-rose-500', Done: 'border-emerald-500', Canceled: 'border-gray-700' };
const types: WorkItemType[] = ['Story', 'Task', 'Bug', 'Review', 'Research', 'Subtask'];
const points = [1, 2, 3, 5, 8, 13] as const;
type Member = { id: string; name: string };
type BoardNode = { id: string; label: string; moduleId: string; moduleName: string; doc: { outcome?: string; actor?: string } };

export default function KanbanView() {
  const store = useStore();
  const { activeProjectId, activeId, modules, projectMembers, selectedWorkItemKey } = store;
  const [items, setItems] = useState<WorkItem[]>([]);
  const [artifactCounts, setArtifactCounts] = useState<Record<string, { comments: number; attachments: number }>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [dragOver, setDragOver] = useState<WorkItemStatus | null>(null);
  const [draggedKey, setDraggedKey] = useState<string | null>(null);
  const [transitioningKey, setTransitioningKey] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);
  const [filters, setFilters] = useState<KanbanFilters>({ search: '', moduleId: activeId || 'all', assigneeId: 'all', type: 'all', facet: 'all', priority: 'all', includeCanceled: false });
  const nodes = useMemo<BoardNode[]>(() => modules.flatMap((module) => module.nodes.map((node) => ({ ...node, moduleId: module.id, moduleName: module.name }))), [modules]);
  const visible = useMemo(() => filterWorkItems(items, filters), [items, filters]);
  const selected = itemByKey(items, selectedWorkItemKey);

  const load = async () => {
    if (!activeProjectId) return;
    setLoading(true); setError('');
    try {
      const loaded = await workItemsApi.listAll(activeProjectId);
      setItems(loaded);
      void Promise.all(loaded.map(async (item) => ({ id: item.id, ...(await workItemsApi.artifacts(item.key)) })))
        .then((counts) => setArtifactCounts(artifactCountMap(counts)))
        .catch(() => undefined);
    }
    catch (requestError) { setError(requestError instanceof Error ? requestError.message : 'Work item tidak dapat dimuat.'); }
    finally { setLoading(false); }
  };
  useEffect(() => { void load(); }, [activeProjectId]);
  useEffect(() => { if (activeId) setFilters((current) => ({ ...current, moduleId: activeId })); }, [activeId]);

  const transition = async (item: WorkItem, status: WorkItemStatus) => {
    if (item.status === status) return;
    if (!canTransitionWorkItem(item.status, status)) {
      store.addNotification('Perpindahan Tidak Diizinkan', `${item.status} hanya dapat dipindahkan ke tahap berikutnya sesuai alur kerja.`, 'warning');
      return;
    }
    let note = '';
    if (status === 'Blocked') {
      const result = await dialog.prompt({
        title: 'Alasan Task Diblokir',
        subtitle: `Tuliskan alasan mengapa ${item.key} tidak dapat dilanjutkan`,
        badge: 'Status: Blocked',
        tone: 'danger',
        icon: 'blocked',
        placeholder: 'Contoh: Menunggu deployment backend API atau klarifikasi spesifikasi...',
        confirmLabel: 'Tandai Blocked',
        cancelLabel: 'Batal',
        required: true,
        quickOptions: [
          'Menunggu endpoint API Backend selesai',
          'Menunggu aset/revisi desain dari UI/UX',
          'Ditemukan bug blocker kritis di sistem',
          'Memerlukan klarifikasi dari Product Manager',
          'Terkendala dependensi library pihak ketiga',
        ],
      });
      if (!result) return;
      note = result.trim();
    }
    if (status === 'Done') {
      const result = await dialog.prompt({
        title: 'Resolusi Penyelesaian Task',
        subtitle: `Tuliskan ringkasan penyelesaian untuk task ${item.key}`,
        badge: 'Status: Done',
        tone: 'success',
        icon: 'check',
        placeholder: 'Contoh: Fitur berhasil diimplementasikan dan diverifikasi di branch staging...',
        confirmLabel: 'Selesaikan Task',
        cancelLabel: 'Batal',
        required: true,
        quickOptions: [
          'Selesai diuji dan diverifikasi (QA Passed)',
          'Pull Request sudah di-merge ke branch utama',
          'Bug berhasil diperbaiki dan lolos uji regresi',
          'Implementasi sesuai kriteria penerimaan',
          'Dokumentasi & spesifikasi telah diperbarui',
        ],
      });
      if (!result) return;
      note = result.trim();
    }
    const previous = items;
    setTransitioningKey(item.key);
      setItems((current) => optimisticStatus(current, item.id, status));
    try {
      const updated = await workItemsApi.transition(item.key, status, item.row_version, note);
      setItems((current) => current.map((candidate) => candidate.id === item.id ? updated : candidate));
    } catch (requestError) {
      setItems(previous);
      store.addNotification('Perubahan Dibatalkan', requestError instanceof Error ? requestError.message : 'Status task tidak dapat diubah.', 'warning');
    } finally {
      setTransitioningKey(null);
    }
  };

  const startDrag = (event: React.DragEvent<HTMLElement>, key: string) => {
    event.dataTransfer.effectAllowed = 'move';
    // text/plain is supported consistently by Chromium, Firefox, and Safari.
    event.dataTransfer.setData('text/plain', key);
    event.dataTransfer.setData('text/work-item', key);
    setDraggedKey(key);
  };

  const dropOnColumn = (event: React.DragEvent<HTMLElement>, status: WorkItemStatus) => {
    event.preventDefault();
    const key = event.dataTransfer.getData('text/plain') || event.dataTransfer.getData('text/work-item') || draggedKey;
    setDragOver(null);
    setDraggedKey(null);
    const item = itemByKey(items, key);
    if (item) void transition(item, status);
  };

  if (!activeProjectId) return <Empty message="Pilih proyek untuk membuka Kanban." />;
  return (
    <div className="flex-1 flex flex-col h-full overflow-hidden bg-[#0A0A0B] text-left text-gray-100">
      <header className="shrink-0 border-b border-white/10 px-6 py-2.5 bg-[#0d0d0f]/80 backdrop-blur-sm">
        <div className="flex items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            <h1 className="text-sm font-bold text-white tracking-wide">Kanban Work Item</h1>
            <span className="rounded-full bg-white/5 border border-white/10 px-2.5 py-0.5 text-[10px] font-mono text-[#C5A267]">
              {visible.length} task
            </span>
          </div>
          <button
            onClick={() => setCreating(true)}
            className="flex items-center gap-1.5 rounded-lg bg-[#C5A267] hover:bg-[#d8b478] px-3 py-1.5 text-xs font-bold text-black transition-all shadow-sm shadow-[#C5A267]/20 cursor-pointer active:scale-95"
          >
            <Plus className="h-3.5 w-3.5" />
            <span>Task baru</span>
          </button>
        </div>
        <BoardFilters filters={filters} setFilters={setFilters} modules={modules} members={projectMembers} />
      </header>

      {transitioningKey && <div className="border-b border-[#C5A267]/20 bg-[#C5A267]/10 px-4 py-2 text-xs text-[#E7CB93]">Memindahkan work item...</div>}
      {loading ? <Empty message="Memuat work item..." /> : error ? <Empty message={error} action={load} /> : (
        <div className="flex-1 overflow-x-auto p-4 scrollbar-thin">
          <div className="grid min-w-[1320px] grid-cols-6 gap-3 items-start h-full">
            {BOARD_STATUSES.map((status) => {
              const cards = visible.filter((item) => item.status === status);
              const metrics = columnMetrics(visible, status);
              return (
                <section
                  key={status}
                  onDragOver={(event) => { event.preventDefault(); event.dataTransfer.dropEffect = 'move'; }}
                  onDragEnter={() => setDragOver(status)}
                  onDragLeave={(event) => { if (!event.currentTarget.contains(event.relatedTarget as Node)) setDragOver(null); }}
                  onDrop={(event) => dropOnColumn(event, status)}
                  className={`min-h-[520px] rounded-lg border border-white/10 border-t-2 bg-[#111113] p-3 transition-colors ${statusTone[status]} ${dragOver === status ? 'ring-1 ring-[#C5A267] bg-[#C5A267]/5' : ''}`}
                >
                  <div className="mb-3 flex items-start justify-between gap-2 border-b border-white/5 pb-3">
                    <div>
                      <h2 className="text-xs font-bold text-white">{status}</h2>
                      <p className="mt-1 text-[10px] text-gray-500">{metrics.points} poin · {metrics.overdue} terlambat</p>
                    </div>
                    <span className="rounded bg-white/5 px-2 py-1 text-[10px] font-mono">{metrics.count}</span>
                  </div>
                  <div className="space-y-2">
                    {cards.map((item) => (
                      <WorkItemCard
                        key={item.id}
                        item={item}
                        counts={artifactCounts[item.id]}
                        node={nodes.find((node) => node.id === item.node_id)}
                        assignee={projectMembers.find((member) => member.id === item.assignee_id)?.name}
                        isDragging={draggedKey === item.key}
                        isTransitioning={transitioningKey === item.key}
                        onDragStart={startDrag}
                        onDragEnd={() => { setDraggedKey(null); setDragOver(null); }}
                        onOpen={() => store.selectWorkItem(item.key)}
                        onTransition={transition}
                      />
                    ))}
                    {!cards.length && (
                      <button
                        onClick={() => setCreating(true)}
                        className="w-full rounded-lg border border-dashed border-white/10 p-6 text-xs text-gray-600 hover:text-gray-400 cursor-pointer transition-colors"
                      >
                        Tambah task
                      </button>
                    )}
                  </div>
                </section>
              );
            })}
          </div>
        </div>
      )}
      {creating && <CreateModal projectId={activeProjectId} modules={modules} nodes={nodes} members={projectMembers} defaultModuleId={filters.moduleId === 'all' ? activeId || '' : filters.moduleId} onClose={() => setCreating(false)} onCreated={(item) => { setItems((current) => [item, ...current]); setCreating(false); store.selectWorkItem(item.key); }} />}
      {selectedWorkItemKey && <DetailModal item={selected} items={items} nodes={nodes} members={projectMembers} counts={selected ? artifactCounts[selected.id] : undefined} onCounts={(id, counts) => setArtifactCounts((current) => ({ ...current, [id]: counts }))} onClose={() => store.selectWorkItem(null)} onUpdated={(updated) => setItems((current) => current.map((item) => item.id === updated.id ? updated : item))} onNavigate={store.selectWorkItem} onOpenNode={(nodeId) => { store.selectWorkItem(null); store.selectNode(nodeId); store.setView('canvas'); }} />}
    </div>
  );
}

export function BoardFilters({
  filters,
  setFilters,
  modules,
  members,
}: {
  filters: KanbanFilters;
  setFilters: React.Dispatch<React.SetStateAction<KanbanFilters>>;
  modules: Array<{ id: string; name: string }>;
  members: Member[];
}) {
  const change = (key: keyof KanbanFilters, value: string | boolean) =>
    setFilters((current) => ({ ...current, [key]: value }));

  const hasActiveFilters = Boolean(
    filters.search.trim() ||
    filters.moduleId !== 'all' ||
    filters.assigneeId !== 'all' ||
    filters.type !== 'all' ||
    filters.facet !== 'all' ||
    filters.priority !== 'all' ||
    filters.includeCanceled
  );

  const resetFilters = () => {
    setFilters({
      search: '',
      moduleId: 'all',
      assigneeId: 'all',
      type: 'all',
      facet: 'all',
      priority: 'all',
      includeCanceled: false,
    });
  };

  const selectClass = (active: boolean) =>
    `h-8 rounded-lg border px-2.5 text-xs outline-none cursor-pointer transition-all max-w-36 shrink-0 truncate ${
      active
        ? 'border-[#C5A267]/50 bg-[#C5A267]/10 text-[#E7CB93] font-medium'
        : 'border-white/10 bg-[#141416] text-gray-300 hover:border-white/20'
    } focus:border-[#C5A267] focus:ring-1 focus:ring-[#C5A267]/30`;

  return (
    <div className="mt-2.5 flex flex-wrap items-center gap-2">
      {/* Compact Search Input */}
      <div className="relative w-44 sm:w-56 shrink-0">
        <Search className="absolute left-2.5 top-2.5 h-3.5 w-3.5 text-gray-500" />
        <input
          value={filters.search}
          onChange={(event) => change('search', event.target.value)}
          placeholder="Cari task..."
          className="h-8 w-full rounded-lg border border-white/10 bg-[#141416] pl-8 pr-7 text-xs text-gray-200 placeholder-gray-500 outline-none transition-colors focus:border-[#C5A267] focus:ring-1 focus:ring-[#C5A267]/30"
        />
        {filters.search && (
          <button
            onClick={() => change('search', '')}
            className="absolute right-2 top-2 p-0.5 text-gray-400 hover:text-white cursor-pointer"
            aria-label="Hapus pencarian"
          >
            <X className="h-3 w-3" />
          </button>
        )}
      </div>

      <div className="h-4 w-px bg-white/10 hidden sm:block shrink-0" />

      {/* Filter items in single horizontal row */}
      <Select
        value={filters.moduleId}
        onChange={(value) => change('moduleId', value)}
        label="Modul"
        options={[['all', 'Semua modul'], ...modules.map((item) => [item.id, item.name])]}
        className={selectClass(filters.moduleId !== 'all')}
      />

      <Select
        value={filters.assigneeId}
        onChange={(value) => change('assigneeId', value)}
        label="Assignee"
        options={[['all', 'Semua assignee'], ...members.map((item) => [item.id, item.name])]}
        className={selectClass(filters.assigneeId !== 'all')}
      />

      <Select
        value={filters.priority}
        onChange={(value) => change('priority', value)}
        label="Prioritas"
        options={[['all', 'Semua prioritas'], ['low', 'Low'], ['medium', 'Medium'], ['high', 'High'], ['critical', 'Critical']]}
        className={selectClass(filters.priority !== 'all')}
      />

      <Select
        value={filters.type}
        onChange={(value) => change('type', value)}
        label="Tipe"
        options={[['all', 'Semua tipe'], ...types.map((item) => [item, item])]}
        className={selectClass(filters.type !== 'all')}
      />

      <Select
        value={filters.facet}
        onChange={(value) => change('facet', value)}
        label="Facet"
        options={[['all', 'Semua facet'], ['business', 'Bisnis'], ['uiux', 'UI/UX'], ['frontend', 'Frontend'], ['backend', 'Backend']]}
        className={selectClass(filters.facet !== 'all')}
      />

      {/* Compact Canceled Chip */}
      <button
        type="button"
        onClick={() => change('includeCanceled', !filters.includeCanceled)}
        className={`h-8 flex items-center gap-1.5 rounded-lg border px-2.5 text-xs transition-colors cursor-pointer shrink-0 ${
          filters.includeCanceled
            ? 'border-[#C5A267]/50 bg-[#C5A267]/10 text-[#E7CB93] font-medium'
            : 'border-white/10 bg-[#141416] text-gray-400 hover:text-gray-200 hover:border-white/20'
        }`}
      >
        <span className={`w-1.5 h-1.5 rounded-full ${filters.includeCanceled ? 'bg-[#C5A267]' : 'bg-gray-600'}`} />
        <span>Canceled</span>
      </button>

      {/* Reset Filter Action */}
      {hasActiveFilters && (
        <button
          type="button"
          onClick={resetFilters}
          className="h-8 flex items-center gap-1 rounded-lg px-2 text-xs text-gray-400 hover:text-white hover:bg-white/5 transition-colors cursor-pointer shrink-0"
          title="Reset semua filter"
        >
          <RotateCcw className="h-3 w-3" />
          <span>Reset</span>
        </button>
      )}
    </div>
  );
}

function Select({ value, onChange, label, options, className = "w-full rounded-xl border border-white/10 bg-[#111113] px-3 py-2 text-xs text-gray-200 outline-none focus:border-[#C5A267] focus:ring-1 focus:ring-[#C5A267]/30 transition-all cursor-pointer" }: { value: string; onChange: (value: string) => void; label: string; options: string[][]; className?: string }) { return <select aria-label={label} value={value} onChange={(event) => onChange(event.target.value)} className={className}>{options.map(([key,text]) => <option key={key} value={key} className="bg-[#18181B] text-white">{text}</option>)}</select>; }

function WorkItemCard({ item, node, assignee, counts, isDragging, isTransitioning, onDragStart, onDragEnd, onOpen, onTransition }: { key?: React.Key; item: WorkItem; node?: BoardNode; assignee?: string; counts?: { comments: number; attachments: number }; isDragging: boolean; isTransitioning: boolean; onDragStart: (event: React.DragEvent<HTMLElement>, key: string) => void; onDragEnd: () => void; onOpen: () => void; onTransition: (item: WorkItem, status: WorkItemStatus) => void }) {
  const overdue = item.due_date && new Date(item.due_date) < new Date() && item.status !== 'Done';
  const onKeyDown = (event: React.KeyboardEvent<HTMLElement>) => {
    if (event.target !== event.currentTarget) return;
    if (event.key === 'Enter' || event.key === ' ') { event.preventDefault(); onOpen(); return; }
    if (!event.altKey || (event.key !== 'ArrowRight' && event.key !== 'ArrowLeft')) return;
    const status = keyboardTransitionTarget(item.status, event.key === 'ArrowRight' ? 'next' : 'previous');
    if (status) { event.preventDefault(); void onTransition(item, status); }
  };
  return <article tabIndex={0} aria-label={`${item.key}: ${item.title}. Alt+Panah kanan atau kiri untuk memindahkan status.`} draggable={!isTransitioning} onKeyDown={onKeyDown} onDragStart={(event) => onDragStart(event, item.key)} onDragEnd={onDragEnd} onClick={onOpen} className={`rounded-lg border border-white/10 bg-[#18181B] p-3 shadow-md transition-[opacity,border-color,transform] hover:border-[#C5A267]/40 focus:outline-none focus:ring-1 focus:ring-[#C5A267] cursor-grab active:cursor-grabbing ${isDragging ? 'scale-[0.98] opacity-45' : ''} ${isTransitioning ? 'pointer-events-none opacity-60' : ''}`}><div className="flex items-start gap-2"><GripVertical className="mt-0.5 h-4 w-4 shrink-0 text-gray-700" /><div className="min-w-0 flex-1"><div className="flex items-center justify-between gap-2"><span className="text-[10px] font-mono text-[#C5A267]">{item.key}</span><span className="text-[9px] uppercase text-gray-500">{item.type}</span></div><h3 className="mt-1 text-xs font-semibold leading-5 text-white">{item.title}</h3>{node && <p className="mt-1 truncate text-[10px] text-gray-500">{node.moduleName} / {node.label}{item.facet_key ? ` / ${item.facet_key}` : ''}</p>}</div></div><div className="mt-3 flex flex-wrap items-center gap-2 text-[10px] text-gray-500"><span className={`rounded px-1.5 py-0.5 ${item.priority === 'critical' ? 'bg-rose-500/15 text-rose-300' : 'bg-white/5'}`}>{item.priority}</span>{item.points && <span>{item.points} pt</span>}<span className="flex items-center gap-1"><CircleUserRound className="h-3 w-3" />{assignee || 'Unassigned'}</span>{item.due_date && <span className={overdue ? 'text-rose-400' : ''}><CalendarDays className="mr-1 inline h-3 w-3" />{item.due_date.slice(0,10)}</span>}{item.status === 'Blocked' && <AlertTriangle className="h-3.5 w-3.5 text-rose-400" />}{counts && <span className="flex items-center gap-1"><MessageSquare className="h-3 w-3" />{counts.comments} <Paperclip className="ml-1 h-3 w-3" />{counts.attachments}</span>}</div><select aria-label={`Ubah status ${item.key}`} value={item.status} onClick={(event) => event.stopPropagation()} onChange={(event) => void onTransition(item, event.target.value as WorkItemStatus)} className="mt-3 w-full rounded border border-white/10 bg-[#111113] px-2 py-1.5 text-[10px] text-gray-400">{BOARD_STATUSES.map((status) => <option key={status}>{status}</option>)}<option>Canceled</option></select></article>;
}

function CreateModal({ projectId, modules, nodes, members, defaultModuleId, onClose, onCreated }: { projectId: string; modules: Array<{id:string;name:string}>; nodes: BoardNode[]; members: Member[]; defaultModuleId: string; onClose: () => void; onCreated: (item: WorkItem) => void }) {
  const [form, setForm] = useState<WorkItemInput>({ type: 'Task', title: '', priority: 'medium', module_id: defaultModuleId || undefined }); const [error,setError]=useState(''); const [saving,setSaving]=useState(false);
  const set=(key:keyof WorkItemInput,value:unknown)=>setForm((current)=>({...current,[key]:value||undefined}));
  const submit=async(event:React.FormEvent)=>{event.preventDefault();setSaving(true);setError('');try{onCreated(await workItemsApi.create(projectId,form));}catch(requestError){setError(requestError instanceof Error?requestError.message:'Task gagal dibuat.');}finally{setSaving(false);}};
  return <Modal title="Task baru" maxWidth="max-w-2xl" onClose={onClose}><form onSubmit={submit} className="grid gap-4 md:grid-cols-2"><Field label="Judul" wide><input autoFocus required value={form.title} onChange={(event)=>set('title',event.target.value)} placeholder="Judul task..." className="input" /></Field><Field label="Deskripsi" wide><textarea value={form.description||''} onChange={(event)=>set('description',event.target.value)} placeholder="Deskripsi pekerjaan..." className="input min-h-24 leading-relaxed" /></Field><Field label="Tipe"><Select value={form.type} onChange={(value)=>set('type',value)} label="Tipe task" options={types.map((item)=>[item,item])} /></Field><Field label="Prioritas"><Select value={form.priority||'medium'} onChange={(value)=>set('priority',value)} label="Prioritas" options={['low','medium','high','critical'].map((item)=>[item,item])} /></Field><Field label="Modul"><Select value={form.module_id||''} onChange={(value)=>{set('module_id',value);set('node_id','');}} label="Modul" options={[['','Pilih modul'],...modules.map((item)=>[item.id,item.name])]} /></Field><Field label="Node"><Select value={form.node_id||''} onChange={(value)=>set('node_id',value)} label="Node" options={[['','Project umum'],...nodes.filter((node)=>!form.module_id||node.moduleId===form.module_id).map((node)=>[node.id,node.label])]} /></Field><Field label="Facet"><Select value={form.facet_key||''} onChange={(value)=>set('facet_key',value)} label="Facet" options={[['','Tanpa facet'],['business','Bisnis'],['uiux','UI/UX'],['frontend','Frontend'],['backend','Backend']]} /></Field><Field label="Assignee"><Select value={form.assignee_id||''} onChange={(value)=>set('assignee_id',value)} label="Assignee" options={[['','Unassigned'],...members.map((item)=>[item.id,item.name])]} /></Field><Field label="Point"><Select value={form.points?.toString()||''} onChange={(value)=>set('points',value?Number(value):undefined)} label="Point" options={[['','Tanpa point'],...points.map((item)=>[String(item),String(item)])]} /></Field><Field label="Deadline"><input type="date" value={form.due_date||''} onChange={(event)=>set('due_date',event.target.value)} className="input" /></Field>{error&&<p className="md:col-span-2 text-xs text-rose-400 font-medium">{error}</p>}<div className="md:col-span-2 flex justify-end gap-2.5 pt-3 border-t border-white/5"><button type="button" onClick={onClose} className="rounded-xl px-4 py-2 text-xs font-semibold text-gray-400 hover:text-white hover:bg-white/5 transition-colors cursor-pointer">Batal</button><button disabled={saving||!form.title.trim()} className="rounded-xl bg-[#C5A267] hover:bg-[#d8b478] px-5 py-2 text-xs font-bold text-black disabled:opacity-50 transition-all shadow-md shadow-[#C5A267]/20 cursor-pointer">{saving?'Menyimpan...':'Buat task'}</button></div></form></Modal>;
}

function DetailModal({ item, items, nodes, members, counts, onClose, onUpdated, onNavigate, onOpenNode, onCounts }: { item?: WorkItem; items: WorkItem[]; nodes: BoardNode[]; members: Member[]; counts?: { comments: number; attachments: number }; onClose:()=>void; onUpdated:(item:WorkItem)=>void; onNavigate:(key:string|null)=>void; onOpenNode:(id:string)=>void; onCounts:(itemId:string, counts:{comments:number;attachments:number})=>void }) {
  const [draft,setDraft]=useState<WorkItem|undefined>(item); const [comments,setComments]=useState<Array<{id:string;author_id:string;body:string;created_at:string}>>([]); const [activity,setActivity]=useState<WorkItemActivity[]>([]); const [comment,setComment]=useState(''); const [error,setError]=useState(''); const [saving,setSaving]=useState(false); const [copiedLink, setCopiedLink] = useState(false);
  useEffect(()=>{setDraft(item);if(item)Promise.all([workItemsApi.comments(item.key),workItemsApi.activity(item.key)]).then(([nextComments,nextActivity])=>{setComments(nextComments);setActivity(nextActivity.items);}).catch((requestError)=>setError(requestError instanceof Error?requestError.message:'Detail delivery gagal dimuat.'));},[item?.key]);
  const dirty=Boolean(item&&draft&&JSON.stringify(item)!==JSON.stringify(draft));
  const confirmClose = async () => {
    if (!dirty) {
      onClose();
      return;
    }
    const confirmed = await dialog.confirm({
      title: 'Buang Perubahan?',
      message: 'Ada perubahan pada task yang belum disimpan. Yakin ingin membuang perubahan dan menutup?',
      confirmLabel: 'Buang Perubahan',
      cancelLabel: 'Tetap Mengedit',
      tone: 'warning',
      icon: 'warning',
    });
    if (confirmed) onClose();
  };
  useEffect(() => {
    const handler = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        event.preventDefault();
        void confirmClose();
      }
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [dirty, onClose]);
  if(!item||!draft)return <Modal title="Memuat detail" onClose={onClose}><p className="text-sm text-gray-500">Mengambil work item dari server...</p></Modal>;
  const index=items.findIndex((candidate)=>candidate.id===item.id); const node=nodes.find((candidate)=>candidate.id===item.node_id);
  const save=async()=>{setSaving(true);setError('');try{const updated=await workItemsApi.update(item.key,{...draft,parent_id:draft.parent_id||null,assignee_id:draft.assignee_id||null,points:draft.points??null,due_date:draft.due_date||null,row_version:item.row_version});onUpdated(updated);setDraft(updated);}catch(requestError){setError(requestError instanceof Error?requestError.message:'Perubahan gagal disimpan.');}finally{setSaving(false);}};
  const send=async(event:React.FormEvent)=>{event.preventDefault();if(!comment.trim())return;try{await workItemsApi.comment(item.key,comment.trim());setComment('');const updated=await workItemsApi.comments(item.key);setComments(updated);onCounts(item.id,{comments:updated.length,attachments:counts?.attachments||0});}catch(requestError){setError(requestError instanceof Error?requestError.message:'Komentar gagal dikirim.');}};
  const close=()=>{void confirmClose();};
  const handleCopyLink = async () => {
    await navigator.clipboard?.writeText(window.location.href);
    setCopiedLink(true);
    setTimeout(() => setCopiedLink(false), 2000);
  };
  const children=items.filter((candidate)=>candidate.parent_id===item.id); const parent=items.find((candidate)=>candidate.id===item.parent_id); const reporter=members.find((member)=>member.id===item.reporter_id);

  return (
    <Modal
      title={`${item.key} · ${item.type}`}
      maxWidth="max-w-6xl xl:max-w-7xl"
      onClose={close}
      actions={
        <>
          <button
            onClick={handleCopyLink}
            className="flex items-center gap-1.5 rounded-lg border border-white/10 bg-white/5 px-2.5 py-1.5 text-xs font-medium text-gray-300 hover:bg-white/10 hover:text-white transition-colors cursor-pointer"
            aria-label="Salin tautan task"
          >
            {copiedLink ? <Check className="h-3.5 w-3.5 text-emerald-400" /> : <Link2 className="h-3.5 w-3.5 text-gray-400" />}
            <span>{copiedLink ? 'Tersalin!' : 'Salin link'}</span>
          </button>
          <div className="flex items-center border border-white/10 rounded-lg overflow-hidden bg-white/5 ml-1">
            <button
              disabled={index <= 0}
              onClick={() => onNavigate(items[index - 1]?.key)}
              aria-label="Task sebelumnya"
              className="p-1.5 text-gray-400 hover:text-white hover:bg-white/10 disabled:opacity-30 disabled:pointer-events-none transition-colors cursor-pointer border-r border-white/10"
            >
              <ChevronLeft className="h-4 w-4" />
            </button>
            <button
              disabled={index < 0 || index >= items.length - 1}
              onClick={() => onNavigate(items[index + 1]?.key)}
              aria-label="Task berikutnya"
              className="p-1.5 text-gray-400 hover:text-white hover:bg-white/10 disabled:opacity-30 disabled:pointer-events-none transition-colors cursor-pointer"
            >
              <ChevronRight className="h-4 w-4" />
            </button>
          </div>
        </>
      }
    >
      <div className="grid gap-8 lg:grid-cols-[minmax(0,1fr)_340px] xl:grid-cols-[minmax(0,1fr)_360px] items-start">
        {/* Left Column: Primary Content (Title, Description, Artifacts, Comments) */}
        <div className="space-y-6 min-w-0">
          <div>
            <span className="block text-[11px] font-bold uppercase tracking-wider text-gray-400 font-mono mb-1.5">Judul Task</span>
            <input
              value={draft.title}
              onChange={(event) => setDraft({ ...draft, title: event.target.value })}
              placeholder="Judul task..."
              className="w-full rounded-xl border border-transparent hover:border-white/10 focus:border-[#C5A267] bg-transparent focus:bg-[#161619] px-3 py-2 text-2xl font-bold text-white placeholder-gray-600 outline-none transition-all"
            />
          </div>

          <div>
            <span className="block text-[11px] font-bold uppercase tracking-wider text-gray-400 font-mono mb-2">Deskripsi</span>
            <textarea
              value={draft.description || ''}
              onChange={(event) => setDraft({ ...draft, description: event.target.value })}
              placeholder="Deskripsi pekerjaan atau spesifikasi alur..."
              className="input min-h-36 p-4 text-xs leading-relaxed resize-y"
            />
          </div>

          {node && (
            <section className="rounded-xl border border-white/10 bg-[#161619] p-4 space-y-2">
              <div className="flex items-center justify-between">
                <span className="text-[10px] font-bold uppercase tracking-wider text-[#C5A267] font-mono">Node Terkait (Canvas)</span>
                <button
                  type="button"
                  onClick={() => onOpenNode(node.id)}
                  className="text-xs font-semibold text-[#C5A267] hover:underline flex items-center gap-1 cursor-pointer"
                >
                  <span>Buka detail node</span>
                  <ArrowUpRight className="h-3.5 w-3.5" />
                </button>
              </div>
              <p className="text-sm font-bold text-white">
                {node.moduleName} / {node.label}
                {item.facet_key && <span className="ml-2 text-xs font-mono font-normal text-gray-400">({item.facet_key})</span>}
              </p>
              <p className="text-xs text-gray-300 leading-relaxed">{node.doc.outcome || 'Outcome belum ditulis.'}</p>
              {node.doc.actor && (
                <p className="text-[11px] text-gray-400">
                  Aktor: <span className="text-gray-200 font-medium">{node.doc.actor}</span>
                </p>
              )}
            </section>
          )}

          {children.length > 0 && (
            <section className="space-y-3">
              <h3 className="text-xs font-bold uppercase tracking-wider text-gray-300 font-mono">Child Tasks ({children.length})</h3>
              <div className="space-y-2">
                {children.map((child) => (
                  <button
                    key={child.id}
                    onClick={() => onNavigate(child.key)}
                    className="flex w-full items-center justify-between rounded-xl border border-white/10 bg-[#161619] p-3 text-left text-xs hover:border-[#C5A267]/40 hover:bg-white/5 transition-all cursor-pointer"
                  >
                    <span className="font-semibold text-white">
                      <span className="font-mono text-[#C5A267] mr-2">{child.key}</span>
                      {child.title}
                    </span>
                    <span className="text-xs text-gray-400 font-mono">{child.status}</span>
                  </button>
                ))}
              </div>
            </section>
          )}

          <WorkItemArtifacts key={item.key} item={item} members={members} onNavigate={(key) => onNavigate(key)} onCounts={(next) => onCounts(item.id, next)} />

          <section className="space-y-4 pt-2 border-t border-white/5">
            <div className="flex items-center gap-2">
              <MessageSquare className="h-4 w-4 text-[#C5A267]" />
              <h3 className="text-xs font-bold uppercase tracking-wider text-gray-300 font-mono">Komentar ({comments.length})</h3>
            </div>

            <form onSubmit={send} className="flex gap-2">
              <input
                value={comment}
                onChange={(event) => setComment(event.target.value)}
                placeholder="Tulis komentar atau update status..."
                className="input text-xs"
              />
              <button
                type="submit"
                disabled={!comment.trim()}
                className="rounded-xl bg-[#C5A267] hover:bg-[#d8b478] px-4 py-2 text-black font-bold transition-all disabled:opacity-40 disabled:pointer-events-none cursor-pointer shrink-0 flex items-center gap-1.5 text-xs shadow-md shadow-[#C5A267]/20"
                aria-label="Kirim komentar"
              >
                <Send className="h-3.5 w-3.5" />
                <span>Kirim</span>
              </button>
            </form>

            <div className="space-y-2.5">
              {comments.map((entry) => {
                const author = members.find((member) => member.id === entry.author_id)?.name || entry.author_id;
                return (
                  <div key={entry.id} className="rounded-xl border border-white/5 bg-[#161619] p-3.5 space-y-1.5">
                    <div className="flex items-center justify-between text-xs">
                      <div className="flex items-center gap-2">
                        <div className="w-5 h-5 rounded-full bg-[#C5A267]/20 text-[#C5A267] font-bold text-[9px] flex items-center justify-center">
                          {author.slice(0, 2).toUpperCase()}
                        </div>
                        <span className="font-semibold text-white">{author}</span>
                      </div>
                      <span className="text-[10px] text-gray-500 font-mono">{new Date(entry.created_at).toLocaleString('id-ID')}</span>
                    </div>
                    <p className="text-xs text-gray-300 leading-relaxed pl-7">{entry.body}</p>
                  </div>
                );
              })}
              {comments.length === 0 && <p className="text-xs text-gray-500 py-2">Belum ada komentar.</p>}
            </div>
          </section>

          <section className="space-y-3 pt-2 border-t border-white/5">
            <h3 className="text-xs font-bold uppercase tracking-wider text-gray-400 font-mono">Aktivitas Delivery</h3>
            {activity.length ? (
              <ol className="space-y-2">
                {activity.map((entry) => (
                  <li key={entry.id} className="rounded-xl border border-white/5 bg-[#161619] p-3 text-xs text-gray-300 flex items-center justify-between">
                    <span>
                      <strong className="text-white font-semibold">{entry.action}</strong> · {members.find((member) => member.id === entry.actor_id)?.name || entry.actor_id || 'Sistem'}
                    </span>
                    <span className="text-[10px] text-gray-500 font-mono">{new Date(entry.created_at).toLocaleString('id-ID')}</span>
                  </li>
                ))}
              </ol>
            ) : (
              <p className="text-xs text-gray-500">Belum ada aktivitas delivery.</p>
            )}
          </section>
        </div>

        {/* Right Column: Jira-Style Attributes Panel */}
        <aside className="rounded-2xl border border-white/10 bg-[#161619] p-5 space-y-4 lg:sticky lg:top-0">
          <div className="flex items-center justify-between border-b border-white/5 pb-3">
            <h3 className="text-xs font-bold uppercase tracking-wider text-gray-300 font-mono">Detail & Atribut</h3>
            <span className="font-mono text-xs font-bold text-[#C5A267]">{item.key}</span>
          </div>

          <Field label="Status">
            <div className="flex items-center">
              <span className={`inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-bold ${
                draft.status === 'Done' ? 'bg-emerald-500/15 text-emerald-300 border border-emerald-500/30' :
                draft.status === 'Blocked' ? 'bg-rose-500/15 text-rose-300 border border-rose-500/30' :
                draft.status === 'In Progress' ? 'bg-blue-500/15 text-blue-300 border border-blue-500/30' :
                draft.status === 'In Review' ? 'bg-purple-500/15 text-purple-300 border border-purple-500/30' :
                draft.status === 'Ready' ? 'bg-cyan-500/15 text-cyan-300 border border-cyan-500/30' :
                'bg-gray-500/15 text-gray-300 border border-gray-500/30'
              }`}>
                <span className="w-1.5 h-1.5 rounded-full bg-current" />
                {draft.status}
              </span>
            </div>
          </Field>

          <Field label="Assignee">
            <Select
              value={draft.assignee_id || ''}
              onChange={(value) => setDraft({ ...draft, assignee_id: value || undefined })}
              label="Assignee"
              options={[['', 'Belum Ditugaskan (Unassigned)'], ...members.map((member) => [member.id, member.name])]}
            />
          </Field>

          <Field label="Reporter">
            <div className="flex items-center gap-2 px-3 py-2 rounded-xl bg-[#111113] border border-white/5 text-xs text-white">
              <div className="w-5 h-5 rounded-full bg-[#C5A267]/20 text-[#C5A267] font-bold text-[9px] flex items-center justify-center">
                {(reporter?.name || item.reporter_id || 'U').slice(0, 2).toUpperCase()}
              </div>
              <span className="font-medium truncate">{reporter?.name || item.reporter_id}</span>
            </div>
          </Field>

          <Field label="Prioritas">
            <Select
              value={draft.priority}
              onChange={(value) => setDraft({ ...draft, priority: value as WorkItem['priority'] })}
              label="Prioritas"
              options={[['low', 'Low'], ['medium', 'Medium'], ['high', 'High'], ['critical', 'Critical']]}
            />
          </Field>

          <Field label="Tipe Task">
            <Select
              value={draft.type}
              onChange={(value) => setDraft({ ...draft, type: value as WorkItemType })}
              label="Tipe"
              options={types.map((value) => [value, value])}
            />
          </Field>

          <Field label="Story Points">
            <Select
              value={draft.points?.toString() || ''}
              onChange={(value) => setDraft({ ...draft, points: value ? (Number(value) as WorkItem['points']) : undefined })}
              label="Point"
              options={[['', 'Tanpa point'], ...points.map((value) => [String(value), `${value} Points`])]}
            />
          </Field>

          <Field label="Facet / Disiplin">
            <Select
              value={draft.facet_key || ''}
              onChange={(value) => setDraft({ ...draft, facet_key: value || undefined })}
              label="Facet"
              options={[['', 'Tanpa facet'], ['business', 'Bisnis'], ['uiux', 'UI/UX'], ['frontend', 'Frontend'], ['backend', 'Backend']]}
            />
          </Field>

          <Field label="Deadline Target">
            <input
              type="date"
              value={draft.due_date?.slice(0, 10) || ''}
              onChange={(event) => setDraft({ ...draft, due_date: event.target.value })}
              className="w-full rounded-xl border border-white/10 bg-[#111113] px-3 py-2 text-xs text-white outline-none focus:border-[#C5A267]"
            />
          </Field>

          <Field label="Parent Task">
            <Select
              value={draft.parent_id || ''}
              onChange={(value) => setDraft({ ...draft, parent_id: value || undefined })}
              label="Parent"
              options={[['', 'Tanpa parent'], ...items.filter((candidate) => candidate.id !== item.id).map((candidate) => [candidate.id, `${candidate.key} · ${candidate.title}`])]}
            />
          </Field>

          {item.blocked_reason && (
            <div className="rounded-xl border border-rose-500/20 bg-rose-500/10 p-3 space-y-1">
              <span className="text-[10px] font-bold uppercase tracking-wider text-rose-400 font-mono">Alasan Blocked:</span>
              <p className="text-xs text-rose-200">{item.blocked_reason}</p>
            </div>
          )}

          {item.resolution && (
            <div className="rounded-xl border border-emerald-500/20 bg-emerald-500/10 p-3 space-y-1">
              <span className="text-[10px] font-bold uppercase tracking-wider text-emerald-400 font-mono">Resolusi Selesai:</span>
              <p className="text-xs text-emerald-200">{item.resolution}</p>
            </div>
          )}

          <div className="border-t border-white/5 pt-3 space-y-1.5 text-[11px] text-gray-500">
            <div className="flex justify-between items-center">
              <span>Dibuat:</span>
              <span className="text-gray-300 font-mono text-[10px]">{item.created_at ? new Date(item.created_at).toLocaleString('id-ID') : '-'}</span>
            </div>
            <div className="flex justify-between items-center">
              <span>Mulai:</span>
              <span className="text-gray-300">{item.start_date || 'Belum dijadwalkan'}</span>
            </div>
          </div>

          <div className="pt-2">
            {error && <p className="mb-2 text-xs text-rose-400 font-medium">{error}</p>}
            <button
              disabled={!dirty || saving}
              onClick={() => void save()}
              className={`w-full rounded-xl px-4 py-2.5 text-xs font-bold transition-all shadow-md cursor-pointer flex items-center justify-center gap-2 ${
                dirty
                  ? 'bg-gradient-to-r from-[#C5A267] to-[#B38F56] hover:from-[#d8b478] hover:to-[#C5A267] text-black shadow-[#C5A267]/20 active:scale-95'
                  : 'bg-white/5 text-gray-500 border border-white/5 cursor-not-allowed opacity-60'
              }`}
            >
              {saving ? (
                'Menyimpan...'
              ) : dirty ? (
                <>
                  <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
                  Simpan Perubahan
                </>
              ) : (
                'Perubahan Tersimpan'
              )}
            </button>
          </div>
        </aside>
      </div>
    </Modal>
  );
}

export function Modal({
  title,
  maxWidth = 'max-w-6xl xl:max-w-7xl',
  onClose,
  actions,
  children,
}: {
  title: string;
  maxWidth?: string;
  onClose: () => void;
  actions?: React.ReactNode;
  children: React.ReactNode;
}) {
  const ref = React.useRef<HTMLElement>(null);
  const originRef = React.useRef<HTMLElement | null>(null);
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
    originRef.current = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    const first = ref.current?.querySelector<HTMLElement>('input,button,select,textarea');
    first?.focus();
    const trap = (event: KeyboardEvent) => {
      if (event.key !== 'Tab' || !ref.current) return;
      const controls = Array.from(
        ref.current.querySelectorAll('input,button:not([disabled]),select,textarea')
      ) as HTMLElement[];
      if (!controls.length) return;
      const firstControl = controls[0];
      const lastControl = controls[controls.length - 1];
      if (event.shiftKey && document.activeElement === firstControl) {
        event.preventDefault();
        controls[focusTrapIndex(0, controls.length, true)]?.focus();
      } else if (!event.shiftKey && document.activeElement === lastControl) {
        event.preventDefault();
        controls[focusTrapIndex(controls.length - 1, controls.length, false)]?.focus();
      }
    };
    window.addEventListener('keydown', trap);
    return () => {
      window.removeEventListener('keydown', trap);
      window.requestAnimationFrame(() => originRef.current?.isConnected && originRef.current.focus());
    };
  }, []);

  const modalMarkup = (
    <div
      className="fixed inset-0 z-[9999] flex items-center justify-center overflow-y-auto bg-black/80 backdrop-blur-md p-4 sm:p-6 md:p-8 lg:p-10 select-text"
      role="dialog"
      aria-modal="true"
      aria-label={title}
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <section
        ref={ref}
        className={`w-full ${maxWidth} my-auto overflow-hidden rounded-2xl border border-white/10 bg-[#121214] shadow-[0_25px_80px_rgba(0,0,0,0.95)] flex flex-col max-h-[min(88vh,880px)] ring-1 ring-white/10 shrink-0`}
      >
        <header className="flex shrink-0 items-center justify-between border-b border-white/10 bg-[#161619]/90 px-6 py-4 backdrop-blur-sm">
          <div className="flex items-center gap-3 min-w-0 flex-1">
            <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg border border-[#C5A267]/25 bg-[#C5A267]/10 text-[#C5A267]">
              <ClipboardList className="h-4 w-4" />
            </div>
            <h2 className="truncate text-sm font-bold text-white tracking-wide">{title}</h2>
          </div>
          <div className="flex items-center gap-2 text-gray-400">
            {actions}
            <button
              onClick={onClose}
              aria-label="Tutup"
              className="rounded-lg p-1.5 text-gray-400 hover:bg-white/10 hover:text-white transition-colors cursor-pointer ml-1"
            >
              <X className="h-5 w-5" />
            </button>
          </div>
        </header>
        <div className="flex-1 overflow-y-auto p-6 md:p-8 scrollbar-thin">{children}</div>
      </section>
    </div>
  );

  if (!mounted || typeof document === 'undefined') {
    return modalMarkup;
  }

  return createPortal(modalMarkup, document.body);
}

function Field({ label, wide, children }: { label: string; wide?: boolean; children: React.ReactNode }) {
  return (
    <div className={`space-y-1.5 ${wide ? 'md:col-span-2' : ''}`}>
      <span className="block text-[11px] font-bold uppercase tracking-wider text-gray-400 font-mono">{label}</span>
      <div>{children}</div>
    </div>
  );
}

function Empty({ message,action }: { message:string;action?:()=>void }) { return <div className="flex h-full flex-1 flex-col items-center justify-center gap-3 p-8 text-sm text-gray-500"><Archive className="h-8 w-8 text-gray-700" /><p>{message}</p>{action&&<button onClick={()=>void action()} className="text-[#C5A267]">Coba lagi</button>}</div>; }
