type Props = { label: string; value?: string; onChange: (value: string) => void; placeholder?: string };

/** Edit legacy newline-backed specification lists as explicit repeatable rows. */
export default function TextRows({ label, value = '', onChange, placeholder }: Props) {
  const rows = value === '' ? [''] : value.split('\n');
  const update = (index: number, next: string) => onChange(rows.map((row, position) => position === index ? next.replaceAll('\n', ' ') : row).join('\n'));
  return <div className="space-y-2">
    <span className="block text-[9px] font-bold uppercase tracking-wider text-gray-400">{label}</span>
    {rows.map((row, index) => <div key={index} className="flex gap-2">
      <input aria-label={`${label} ${index + 1}`} value={row} onChange={(event) => update(index, event.target.value)} placeholder={placeholder} className="input min-w-0 flex-1" />
      <button type="button" aria-label={`Hapus ${label} ${index + 1}`} onClick={() => onChange(rows.filter((_, position) => position !== index).join('\n'))} className="text-xs text-gray-400">Hapus</button>
    </div>)}
    <button type="button" onClick={() => onChange(`${value}\n`)} className="text-xs text-[#C5A267]">+ Tambah baris</button>
  </div>;
}
