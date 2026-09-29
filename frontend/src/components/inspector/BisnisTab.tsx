/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { Node, BusinessFacet } from '../../domain/types';
import { useStore } from '../../store/useStore';
import TextRows from './TextRows';

interface BisnisTabProps {
  node: Node;
}

export default function BisnisTab({ node }: BisnisTabProps) {
  const { updateDoc } = useStore();
  const doc = node.doc || {};

  const handleFieldChange = (key: keyof BusinessFacet, val: string) => {
    updateDoc(node.id, { [key]: val });
  };

  const fields: { id: keyof BusinessFacet; label: string; placeholder: string; type: 'text' | 'textarea' }[] = [
    { id: 'outcome', label: 'Outcome', placeholder: 'Nilai bisnis yang dihasilkan...', type: 'textarea' },
    { id: 'actor', label: 'Aktor Pelaku', placeholder: 'Karyawan, HR, Admin...', type: 'text' },
    { id: 'trigger', label: 'Pemicu Langkah', placeholder: 'Submit form, jadwal otomatis, approval diterima...', type: 'text' },
    { id: 'system', label: 'Sistem Terkait', placeholder: 'FlowakPortal, MySQL Server, Oracle...', type: 'text' },
    { id: 'sla', label: 'Resolusi SLA', placeholder: 'Instan, 2 jam, 1 hari...', type: 'text' },
    { id: 'priority', label: 'Prioritas', placeholder: 'low, medium, high, critical...', type: 'text' },
    { id: 'riskLevel', label: 'Level Risiko', placeholder: 'low, medium, high...', type: 'text' },
    { id: 'preconditions', label: 'Prasyarat', placeholder: 'Kondisi sebelum proses dapat dimulai...', type: 'textarea' },
    { id: 'input', label: 'Kebutuhan Input Data', placeholder: 'Formulir data karyawan, id pengajuan...', type: 'textarea' },
    { id: 'process', label: 'Uraian Proses Bisnis', placeholder: 'Mengecek jatah saldo cuti...', type: 'textarea' },
    { id: 'output', label: 'Hasil Output Data', placeholder: 'Konfirmasi kelulusan, penolakan...', type: 'textarea' },
    { id: 'exceptionPaths', label: 'Alur Pengecualian / Gagal', placeholder: 'Jika data tidak valid, arahkan ke revisi atau penolakan...', type: 'textarea' },
    { id: 'acceptanceCriteria', label: 'Kriteria Selesai', placeholder: 'Kondisi yang harus terpenuhi agar langkah dianggap siap.', type: 'textarea' },
  ];

  return (
    <div className="space-y-4 text-left select-text">
       <div>
         <span className="text-[10px] font-bold font-mono tracking-widest text-[#C5A267] uppercase block mb-1">
           Spesifikasi Bisnis
         </span>
         <p className="text-[11px] text-gray-400 mb-4 font-sans">
           Tulis spesifikasi fungsional untuk langkah ini. Perubahan akan direfleksikan secara instan pada semua lensa.
         </p>
       </div>

       <div className="space-y-3.5">
         {fields.map((f) => (
           <div key={f.id} className="space-y-1">
             <label className="block text-[9px] font-bold text-gray-400 font-mono text-left uppercase tracking-wider">
               {f.label}
             </label>
             {f.type === 'text' ? (
               <input
                 type="text"
                value={(doc[f.id] as string) || ''}
                 onChange={(e) => handleFieldChange(f.id, e.target.value)}
                 placeholder={f.placeholder}
                 className="w-full text-xs border border-white/5 rounded-xl px-3 py-2 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267] transition font-medium"
               />
             ) : (
               <textarea
                value={(doc[f.id] as string) || ''}
                 onChange={(e) => handleFieldChange(f.id, e.target.value)}
                 placeholder={f.placeholder}
                 rows={2}
                 className="w-full text-xs border border-white/5 rounded-xl px-3 py-2 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267] transition font-medium h-16 resize-none"
               />
             )}
           </div>
        ))}

        <div className="space-y-2">
          <span className="block text-[9px] font-bold text-gray-400 uppercase tracking-wider">Aturan bisnis</span>
          {(doc.rules || []).map((rule, index) => <div key={index} className="grid gap-2 sm:grid-cols-[90px_100px_1fr_auto]">
            <input aria-label={`Kode aturan ${index + 1}`} className="input" placeholder="Kode" value={rule.code || ''} onChange={(event) => updateDoc(node.id, { rules: (doc.rules || []).map((current, position) => position === index ? { ...current, code: event.target.value } : current) })} />
            <select aria-label={`Severity aturan ${index + 1}`} className="input" value={rule.severity || 'medium'} onChange={(event) => updateDoc(node.id, { rules: (doc.rules || []).map((current, position) => position === index ? { ...current, severity: event.target.value as typeof rule.severity } : current) })}>{['low', 'medium', 'high', 'critical'].map((level) => <option key={level}>{level}</option>)}</select>
            <input aria-label={`Deskripsi aturan ${index + 1}`} className="input" placeholder="Deskripsi aturan" value={rule.description} onChange={(event) => updateDoc(node.id, { rules: (doc.rules || []).map((current, position) => position === index ? { ...current, description: event.target.value } : current) })} />
            <button type="button" className="text-xs text-gray-400" onClick={() => updateDoc(node.id, { rules: (doc.rules || []).filter((_, position) => position !== index) })}>Hapus</button>
          </div>)}
          <button type="button" className="text-xs text-[#C5A267]" onClick={() => updateDoc(node.id, { rules: [...(doc.rules || []), { description: 'Aturan baru', severity: 'medium' }] })}>+ Tambah aturan</button>
        </div>
        <TextRows label="Reference links" value={doc.referenceLinks} onChange={(value) => updateDoc(node.id, { referenceLinks: value })} placeholder="https://..." />
       </div>
    </div>
  );
}
