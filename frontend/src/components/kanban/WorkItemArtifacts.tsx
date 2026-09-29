import { useEffect, useState } from 'react';
import { CheckSquare2, Link2, Paperclip, Trash2, Users } from 'lucide-react';
import type { WorkItem, WorkItemArtifacts as ArtifactData } from '../../domain/types';
import { workItemsApi } from '../../services/workItems';

type Member = { id: string; name: string };
type Props = { key?: string; item: WorkItem; members: Member[]; onCounts: (counts: { comments: number; attachments: number }) => void; onNavigate: (key: string) => void };
type LinkType = 'relates_to' | 'blocks' | 'depends_on';

export default function WorkItemArtifacts({ item, members, onCounts, onNavigate }: Props) {
  const [data, setData] = useState<ArtifactData | null>(null);
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
  const [checklistBody, setChecklistBody] = useState('');
  const [editingChecklist, setEditingChecklist] = useState<string | null>(null);
  const [editingBody, setEditingBody] = useState('');
  const [watcherId, setWatcherId] = useState('');
  const [linkedKey, setLinkedKey] = useState('');
  const [linkType, setLinkType] = useState<LinkType>('relates_to');
  const [fileName, setFileName] = useState('');
  const [artifactURL, setArtifactURL] = useState('');
  const [contentType, setContentType] = useState('');

  const refresh = async () => {
    const updated = await workItemsApi.artifacts(item.key);
    setData(updated);
    onCounts({ comments: updated.comment_count, attachments: updated.attachment_count });
  };

  useEffect(() => {
    let active = true;
    workItemsApi.artifacts(item.key).then((updated) => {
      if (!active) return;
      setData(updated);
      onCounts({ comments: updated.comment_count, attachments: updated.attachment_count });
    }).catch((reason) => { if (active) setError(reason instanceof Error ? reason.message : 'Artifact gagal dimuat.'); });
    return () => { active = false; };
  }, [item.key]);

  const mutate = async (action: () => Promise<unknown>) => {
    setBusy(true); setError('');
    try { await action(); await refresh(); }
    catch (reason) { setError(reason instanceof Error ? reason.message : 'Artifact tidak dapat disimpan.'); }
    finally { setBusy(false); }
  };

  if (!data) return <section className="rounded-lg border border-white/10 p-3 text-sm text-gray-400">{error || 'Memuat checklist dan evidence…'}</section>;

  return <div className="space-y-5">
    <section className="rounded-lg border border-white/10 bg-[#151517] p-3">
      <h3 className="mb-3 flex items-center gap-2 text-sm font-bold"><CheckSquare2 className="h-4 w-4" />Checklist ({data.checklist.filter((entry) => entry.is_complete).length}/{data.checklist.length})</h3>
      <ul className="space-y-2">{data.checklist.map((entry) => <li key={entry.id} className="flex items-start gap-2 text-xs">
        <input type="checkbox" aria-label={`Selesai ${entry.body}`} checked={entry.is_complete} disabled={busy} onChange={(event) => void mutate(() => workItemsApi.updateArtifact(item.key, 'checklist', entry.id, { is_complete: event.target.checked }))} />
        {editingChecklist === entry.id ? <form className="flex min-w-0 flex-1 gap-2" onSubmit={(event) => { event.preventDefault(); void mutate(async () => { await workItemsApi.updateArtifact(item.key, 'checklist', entry.id, { body: editingBody.trim() }); setEditingChecklist(null); }); }}><input aria-label="Ubah checklist" className="input" value={editingBody} onChange={(event) => setEditingBody(event.target.value)} required /><button disabled={busy || !editingBody.trim()} className="text-[#C5A267]">Simpan</button></form> : <button className={`min-w-0 flex-1 text-left ${entry.is_complete ? 'text-gray-500 line-through' : 'text-gray-200'}`} onClick={() => { setEditingChecklist(entry.id); setEditingBody(entry.body); }}>{entry.body}</button>}
        <button disabled={busy} aria-label={`Hapus checklist ${entry.body}`} onClick={() => void mutate(() => workItemsApi.deleteArtifact(item.key, 'checklist', entry.id))}><Trash2 className="h-3.5 w-3.5 text-gray-500" /></button>
      </li>)}</ul>
      {data.checklist.length === 0 && <p className="text-xs text-gray-500">Belum ada kriteria selesai.</p>}
      <form className="mt-3 flex gap-2" onSubmit={(event) => { event.preventDefault(); void mutate(async () => { await workItemsApi.createArtifact(item.key, 'checklist', { body: checklistBody.trim() }); setChecklistBody(''); }); }}><input aria-label="Kriteria baru" className="input" placeholder="Tambah kriteria selesai" value={checklistBody} onChange={(event) => setChecklistBody(event.target.value)} /><button disabled={busy || !checklistBody.trim()} className="text-xs text-[#C5A267]">Tambah</button></form>
    </section>

    <section className="rounded-lg border border-white/10 bg-[#151517] p-3">
      <h3 className="mb-3 flex items-center gap-2 text-sm font-bold"><Users className="h-4 w-4" />Watcher ({data.watchers.length})</h3>
      <ul className="space-y-2">{data.watchers.map((entry) => <li key={entry.user_id} className="flex justify-between text-xs text-gray-200"><span>{entry.name}</span><button disabled={busy} onClick={() => void mutate(() => workItemsApi.deleteArtifact(item.key, 'watchers', entry.user_id))} aria-label={`Hapus watcher ${entry.name}`}><Trash2 className="h-3.5 w-3.5 text-gray-500" /></button></li>)}</ul>
      <form className="mt-3 flex gap-2" onSubmit={(event) => { event.preventDefault(); void mutate(async () => { await workItemsApi.createArtifact(item.key, 'watchers', { user_id: watcherId }); setWatcherId(''); }); }}><select aria-label="Pilih watcher" className="input" value={watcherId} onChange={(event) => setWatcherId(event.target.value)}><option value="">Pilih anggota</option>{members.filter((member) => !data.watchers.some((entry) => entry.user_id === member.id)).map((member) => <option key={member.id} value={member.id}>{member.name}</option>)}</select><button disabled={busy || !watcherId} className="text-xs text-[#C5A267]">Tambah</button></form>
    </section>

    <section className="rounded-lg border border-white/10 bg-[#151517] p-3">
      <h3 className="mb-3 flex items-center gap-2 text-sm font-bold"><Link2 className="h-4 w-4" />Work item terkait ({data.links.length})</h3>
      <ul className="space-y-2">{data.links.map((entry) => <li key={entry.id} className="flex items-center gap-2 text-xs"><button className="flex-1 text-left text-[#C5A267]" onClick={() => onNavigate(entry.linked_key)}>{entry.linked_key}</button><select aria-label={`Jenis link ${entry.linked_key}`} className="rounded bg-[#111113] text-gray-300" value={entry.link_type} disabled={busy} onChange={(event) => void mutate(() => workItemsApi.updateArtifact(item.key, 'links', entry.id, { link_type: event.target.value }))}><option value="relates_to">Terkait</option><option value="blocks">Memblokir</option><option value="depends_on">Bergantung pada</option></select><button disabled={busy} aria-label={`Hapus link ${entry.linked_key}`} onClick={() => void mutate(() => workItemsApi.deleteArtifact(item.key, 'links', entry.id))}><Trash2 className="h-3.5 w-3.5 text-gray-500" /></button></li>)}</ul>
      <form className="mt-3 flex flex-wrap gap-2" onSubmit={(event) => { event.preventDefault(); void mutate(async () => { await workItemsApi.createArtifact(item.key, 'links', { linked_key: linkedKey.trim(), link_type: linkType }); setLinkedKey(''); }); }}><input aria-label="Key work item terkait" className="input min-w-32 flex-1" placeholder="KEY-123" value={linkedKey} onChange={(event) => setLinkedKey(event.target.value)} /><select aria-label="Jenis link baru" className="input" value={linkType} onChange={(event) => setLinkType(event.target.value as LinkType)}><option value="relates_to">Terkait</option><option value="blocks">Memblokir</option><option value="depends_on">Bergantung pada</option></select><button disabled={busy || !linkedKey.trim()} className="text-xs text-[#C5A267]">Tambah</button></form>
    </section>

    <section className="rounded-lg border border-white/10 bg-[#151517] p-3">
      <h3 className="mb-2 flex items-center gap-2 text-sm font-bold"><Paperclip className="h-4 w-4" />Evidence ({data.attachment_count})</h3>
      <p className="mb-3 text-xs text-gray-500">Simpan tautan HTTPS ke berkas yang sudah ada. Flowak menyimpan metadata saja.</p>
      <ul className="space-y-2">{data.attachments.map((entry) => <li key={entry.id} className="flex items-center gap-2 text-xs"><a className="min-w-0 flex-1 truncate text-[#C5A267] underline" href={entry.storage_key} target="_blank" rel="noopener noreferrer">{entry.file_name}</a><button disabled={busy} onClick={() => { const next = window.prompt('Nama evidence', entry.file_name); if (next?.trim()) void mutate(() => workItemsApi.updateArtifact(item.key, 'attachments', entry.id, { file_name: next.trim() })); }} className="text-gray-400">Ubah</button><button disabled={busy} aria-label={`Hapus evidence ${entry.file_name}`} onClick={() => void mutate(() => workItemsApi.deleteArtifact(item.key, 'attachments', entry.id))}><Trash2 className="h-3.5 w-3.5 text-gray-500" /></button></li>)}</ul>
      <form className="mt-3 grid gap-2 sm:grid-cols-2" onSubmit={(event) => { event.preventDefault(); void mutate(async () => { await workItemsApi.createArtifact(item.key, 'attachments', { file_name: fileName.trim(), storage_key: artifactURL.trim(), content_type: contentType.trim() }); setFileName(''); setArtifactURL(''); setContentType(''); }); }}><input aria-label="Nama evidence" className="input" placeholder="Nama berkas" value={fileName} onChange={(event) => setFileName(event.target.value)} required /><input aria-label="Tautan HTTPS evidence" className="input" type="url" placeholder="https://…" value={artifactURL} onChange={(event) => setArtifactURL(event.target.value)} required /><input aria-label="Tipe konten evidence" className="input" placeholder="application/pdf (opsional)" value={contentType} onChange={(event) => setContentType(event.target.value)} /><button disabled={busy || !fileName.trim() || !artifactURL.trim()} className="rounded bg-[#C5A267] px-3 py-2 text-xs font-bold text-black disabled:opacity-50">Tambah evidence</button></form>
    </section>
    {error && <p role="alert" className="text-xs text-rose-400">{error}</p>}
  </div>;
}
