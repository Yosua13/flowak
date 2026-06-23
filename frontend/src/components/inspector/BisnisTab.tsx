/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { Node, BusinessFacet } from '../../domain/types';
import { useStore } from '../../store/useStore';

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
    { id: 'actor', label: 'Aktor Pelaku', placeholder: 'Karyawan, HR, Admin...', type: 'text' },
    { id: 'system', label: 'Sistem Terkait', placeholder: 'FlowakPortal, MySQL Server, Oracle...', type: 'text' },
    { id: 'sla', label: 'Resolusi SLA', placeholder: 'Instan, 2 jam, 1 hari...', type: 'text' },
    { id: 'input', label: 'Kebutuhan Input Data', placeholder: 'Formulir data karyawan, id pengajuan...', type: 'textarea' },
    { id: 'process', label: 'Uraian Proses Bisnis', placeholder: 'Mengecek jatah saldo cuti...', type: 'textarea' },
    { id: 'output', label: 'Hasil Output Data', placeholder: 'Konfirmasi kelulusan, penolakan...', type: 'textarea' },
    { id: 'rules', label: 'Aturan Bisnis (Aturan Validasi/Konsistensi)', placeholder: 'Sisa cuti harus > 0...', type: 'textarea' },
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
                 value={doc[f.id] || ''}
                 onChange={(e) => handleFieldChange(f.id, e.target.value)}
                 placeholder={f.placeholder}
                 className="w-full text-xs border border-white/5 rounded-xl px-3 py-2 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267] transition font-medium"
               />
             ) : (
               <textarea
                 value={doc[f.id] || ''}
                 onChange={(e) => handleFieldChange(f.id, e.target.value)}
                 placeholder={f.placeholder}
                 rows={2}
                 className="w-full text-xs border border-white/5 rounded-xl px-3 py-2 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267] transition font-medium h-16 resize-none"
               />
             )}
           </div>
         ))}
       </div>
    </div>
  );
}
