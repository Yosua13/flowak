/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useRef, useState } from 'react';
import { motion, AnimatePresence } from 'motion/react';
import {
  AlertOctagon,
  CheckCircle2,
  AlertTriangle,
  HelpCircle,
  X,
  Sparkles,
  Trash2,
  Archive,
  Edit3,
  CornerDownLeft,
} from 'lucide-react';
import { useDialogStore, type DialogTone } from '../../services/dialog';

export default function GlobalDialog() {
  const activeDialog = useDialogStore((state) => state.activeDialog);
  const close = useDialogStore((state) => state.close);

  if (!activeDialog) return null;

  return (
    <AnimatePresence>
      {activeDialog && (
        <motion.div
          key="global-dialog-overlay"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.18 }}
          className="fixed inset-0 z-[100000] flex items-center justify-center p-4 bg-black/75 backdrop-blur-md select-none"
          onClick={close}
        >
          {activeDialog.type === 'prompt' ? (
            <PromptDialogContent dialog={activeDialog} onClose={close} />
          ) : (
            <ConfirmDialogContent dialog={activeDialog} onClose={close} />
          )}
        </motion.div>
      )}
    </AnimatePresence>
  );
}

export function PromptDialogContent({
  dialog,
  onClose,
}: {
  dialog: Extract<import('../../services/dialog').ActiveDialogConfig, { type: 'prompt' }>;
  onClose: () => void;
}) {
  const [val, setVal] = useState(dialog.initialValue || '');
  const [error, setError] = useState('');
  const [shake, setShake] = useState(false);
  const inputRef = useRef<HTMLTextAreaElement>(null);

  const tone = dialog.tone || 'info';

  useEffect(() => {
    setVal(dialog.initialValue || '');
    setError('');
    // Auto-focus input after modal renders
    const timer = setTimeout(() => {
      inputRef.current?.focus();
      inputRef.current?.select();
    }, 50);
    return () => clearTimeout(timer);
  }, [dialog]);

  useEffect(() => {
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        e.preventDefault();
        onClose();
      } else if (e.key === 'Enter' && (e.ctrlKey || e.metaKey || !dialog.multiline)) {
        e.preventDefault();
        handleSubmit();
      }
    };
    window.addEventListener('keydown', handleKey);
    return () => window.removeEventListener('keydown', handleKey);
  }, [val, onClose, dialog]);

  const handleSubmit = (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    const trimmed = val.trim();
    if (dialog.required !== false && !trimmed) {
      setError('Mohon tuliskan keterangan sebelum melanjutkan.');
      setShake(true);
      setTimeout(() => setShake(false), 500);
      inputRef.current?.focus();
      return;
    }
    dialog.resolve(trimmed);
  };

  const handleChipClick = (chip: string) => {
    setError('');
    if (!val.trim()) {
      setVal(chip);
    } else if (val.includes(chip)) {
      // If already contains, keep it or allow toggle
      return;
    } else {
      setVal(`${val.trim()}, ${chip}`);
    }
    inputRef.current?.focus();
  };

  const getToneIcon = () => {
    if (dialog.icon === 'blocked' || tone === 'danger') {
      return <AlertOctagon className="w-5 h-5 text-rose-400" />;
    }
    if (dialog.icon === 'check' || tone === 'success') {
      return <CheckCircle2 className="w-5 h-5 text-emerald-400" />;
    }
    if (dialog.icon === 'warning' || tone === 'warning') {
      return <AlertTriangle className="w-5 h-5 text-amber-400" />;
    }
    if (dialog.icon === 'edit') {
      return <Edit3 className="w-5 h-5 text-[#C5A267]" />;
    }
    return <Sparkles className="w-5 h-5 text-[#C5A267]" />;
  };

  const toneGlowClass: Record<DialogTone, string> = {
    danger: 'shadow-[0_0_60px_-15px_rgba(244,63,94,0.35)] border-rose-500/30',
    success: 'shadow-[0_0_60px_-15px_rgba(16,185,129,0.35)] border-emerald-500/30',
    warning: 'shadow-[0_0_60px_-15px_rgba(245,158,11,0.35)] border-amber-500/30',
    info: 'shadow-[0_0_60px_-15px_rgba(197,162,103,0.35)] border-[#C5A267]/30',
  };

  const toneBadgeClass: Record<DialogTone, string> = {
    danger: 'bg-rose-500/15 border-rose-500/30 text-rose-300',
    success: 'bg-emerald-500/15 border-emerald-500/30 text-emerald-300',
    warning: 'bg-amber-500/15 border-amber-500/30 text-amber-300',
    info: 'bg-[#C5A267]/15 border-[#C5A267]/30 text-[#E7CB93]',
  };

  const toneButtonClass: Record<DialogTone, string> = {
    danger: 'bg-gradient-to-r from-rose-600 to-rose-700 hover:from-rose-500 hover:to-rose-600 text-white shadow-lg shadow-rose-950/60',
    success: 'bg-gradient-to-r from-emerald-600 to-emerald-700 hover:from-emerald-500 hover:to-emerald-600 text-white shadow-lg shadow-emerald-950/60',
    warning: 'bg-gradient-to-r from-amber-500 to-amber-600 hover:from-amber-400 hover:to-amber-500 text-black shadow-lg shadow-amber-950/60',
    info: 'bg-gradient-to-r from-[#C5A267] to-[#B38F56] hover:from-[#d8b478] hover:to-[#C5A267] text-black shadow-lg shadow-[#C5A267]/20',
  };

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.94, y: 14 }}
      animate={{
        opacity: 1,
        scale: 1,
        y: 0,
        x: shake ? [-6, 6, -4, 4, 0] : 0,
        transition: { type: 'spring', damping: 26, stiffness: 380 },
      }}
      exit={{ opacity: 0, scale: 0.96, y: 8, transition: { duration: 0.15 } }}
      onClick={(e) => e.stopPropagation()}
      className={`relative w-full max-w-lg overflow-hidden rounded-2xl bg-[#131315] border ${toneGlowClass[tone]} text-left shadow-2xl transition-all`}
      role="dialog"
      aria-modal="true"
    >
      {/* Ambient Top Glow Line */}
      <div
        className={`h-1 w-full bg-gradient-to-r ${
          tone === 'danger'
            ? 'from-rose-500 via-rose-400 to-transparent'
            : tone === 'success'
            ? 'from-emerald-500 via-emerald-400 to-transparent'
            : tone === 'warning'
            ? 'from-amber-500 via-amber-400 to-transparent'
            : 'from-[#C5A267] via-[#E7CB93] to-transparent'
        }`}
      />

      {/* Header */}
      <div className="flex items-start justify-between p-5 pb-4 border-b border-white/5">
        <div className="flex items-start gap-3.5 min-w-0 flex-1">
          <div
            className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-xl border ${
              tone === 'danger'
                ? 'bg-rose-500/10 border-rose-500/20'
                : tone === 'success'
                ? 'bg-emerald-500/10 border-emerald-500/20'
                : tone === 'warning'
                ? 'bg-amber-500/10 border-amber-500/20'
                : 'bg-[#C5A267]/10 border-[#C5A267]/20'
            }`}
          >
            {getToneIcon()}
          </div>
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <span
                className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-[10px] font-mono font-semibold uppercase tracking-wider border ${toneBadgeClass[tone]}`}
              >
                <span
                  className={`h-1.5 w-1.5 rounded-full ${
                    tone === 'danger'
                      ? 'bg-rose-400 animate-pulse'
                      : tone === 'success'
                      ? 'bg-emerald-400'
                      : tone === 'warning'
                      ? 'bg-amber-400'
                      : 'bg-[#C5A267]'
                  }`}
                />
                {dialog.badge || 'Flowak Prompt'}
              </span>
            </div>
            <h3 className="mt-1.5 text-base font-bold text-white tracking-tight leading-snug">
              {dialog.title}
            </h3>
            {dialog.subtitle && (
              <p className="mt-1 text-xs text-gray-400 leading-relaxed font-normal">
                {dialog.subtitle}
              </p>
            )}
          </div>
        </div>
        <button
          onClick={onClose}
          aria-label="Tutup popup"
          className="ml-2 rounded-lg p-1.5 text-gray-500 hover:text-white hover:bg-white/5 transition-colors cursor-pointer"
        >
          <X className="h-4 w-4" />
        </button>
      </div>

      {/* Body */}
      <form onSubmit={handleSubmit} className="p-5 space-y-4">
        {/* Quick Suggestion Chips */}
        {dialog.quickOptions && dialog.quickOptions.length > 0 && (
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <span className="flex items-center gap-1.5 text-[11px] font-semibold text-gray-400">
                <Sparkles className="h-3 w-3 text-[#C5A267]" />
                Pilihan Cepat:
              </span>
              <span className="text-[10px] text-gray-500">Klik untuk memilih</span>
            </div>
            <div className="flex flex-wrap gap-1.5">
              {dialog.quickOptions.map((chip, idx) => {
                const isSelected = val.includes(chip);
                return (
                  <button
                    key={idx}
                    type="button"
                    onClick={() => handleChipClick(chip)}
                    className={`rounded-full px-3 py-1 text-[11px] font-medium transition-all cursor-pointer border ${
                      isSelected
                        ? `${toneBadgeClass[tone]} shadow-sm scale-[1.02]`
                        : 'bg-white/5 border-white/10 text-gray-300 hover:bg-white/10 hover:border-white/20 hover:text-white'
                    }`}
                  >
                    {chip}
                  </button>
                );
              })}
            </div>
          </div>
        )}

        {/* Input Field */}
        <div className="space-y-1.5">
          <div className="relative">
            <textarea
              ref={inputRef}
              rows={dialog.multiline !== false ? 3 : 2}
              value={val}
              onChange={(e) => {
                setVal(e.target.value);
                if (error) setError('');
              }}
              placeholder={dialog.placeholder || 'Tuliskan keterangan di sini...'}
              className={`w-full rounded-xl bg-[#0D0D0F] p-3 text-xs text-white placeholder-gray-500 border transition-all outline-none resize-none scrollbar-thin ${
                error
                  ? 'border-rose-500/80 ring-2 ring-rose-500/20'
                  : 'border-white/10 focus:border-[#C5A267] focus:ring-1 focus:ring-[#C5A267]/40'
              }`}
            />
            {val && (
              <button
                type="button"
                onClick={() => {
                  setVal('');
                  inputRef.current?.focus();
                }}
                className="absolute right-2.5 top-2.5 text-[10px] text-gray-500 hover:text-gray-300 bg-white/5 hover:bg-white/10 rounded px-1.5 py-0.5 cursor-pointer"
              >
                Hapus
              </button>
            )}
          </div>

          {/* Validation & Meta */}
          <div className="flex items-center justify-between text-[11px]">
            {error ? (
              <span className="text-rose-400 font-medium flex items-center gap-1">
                <AlertOctagon className="h-3 w-3 inline" />
                {error}
              </span>
            ) : (
              <span className="text-gray-500 flex items-center gap-1 font-mono text-[10px]">
                <CornerDownLeft className="h-3 w-3 text-gray-600 inline" />
                {dialog.multiline !== false ? 'Ctrl + Enter kirim • Esc batal' : 'Enter kirim • Esc batal'}
              </span>
            )}
            <span className="font-mono text-[10px] text-gray-500">
              {val.length}/300
            </span>
          </div>
        </div>

        {/* Footer Actions */}
        <div className="flex items-center justify-end gap-2.5 pt-2 border-t border-white/5">
          <button
            type="button"
            onClick={onClose}
            className="rounded-xl border border-white/10 bg-white/5 px-4 py-2 text-xs font-semibold text-gray-300 hover:bg-white/10 hover:text-white transition-all cursor-pointer"
          >
            {dialog.cancelLabel || 'Batal'}
          </button>
          <button
            type="submit"
            className={`rounded-xl px-5 py-2 text-xs font-bold transition-all active:scale-95 cursor-pointer flex items-center gap-1.5 ${toneButtonClass[tone]}`}
          >
            {dialog.confirmLabel || 'Simpan'}
          </button>
        </div>
      </form>
    </motion.div>
  );
}

export function ConfirmDialogContent({
  dialog,
  onClose,
}: {
  dialog: Extract<import('../../services/dialog').ActiveDialogConfig, { type: 'confirm' }>;
  onClose: () => void;
}) {
  const tone = dialog.tone || 'warning';

  useEffect(() => {
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        e.preventDefault();
        onClose();
      } else if (e.key === 'Enter') {
        e.preventDefault();
        dialog.resolve(true);
      }
    };
    window.addEventListener('keydown', handleKey);
    return () => window.removeEventListener('keydown', handleKey);
  }, [onClose, dialog]);

  const getToneIcon = () => {
    if (dialog.icon === 'trash' || tone === 'danger') {
      return <Trash2 className="w-5 h-5 text-rose-400" />;
    }
    if (dialog.icon === 'archive') {
      return <Archive className="w-5 h-5 text-amber-400" />;
    }
    if (dialog.icon === 'warning' || tone === 'warning') {
      return <AlertTriangle className="w-5 h-5 text-amber-400" />;
    }
    return <HelpCircle className="w-5 h-5 text-[#C5A267]" />;
  };

  const toneGlowClass: Record<DialogTone, string> = {
    danger: 'shadow-[0_0_60px_-15px_rgba(244,63,94,0.35)] border-rose-500/30',
    success: 'shadow-[0_0_60px_-15px_rgba(16,185,129,0.35)] border-emerald-500/30',
    warning: 'shadow-[0_0_60px_-15px_rgba(245,158,11,0.35)] border-amber-500/30',
    info: 'shadow-[0_0_60px_-15px_rgba(197,162,103,0.35)] border-[#C5A267]/30',
  };

  const toneButtonClass: Record<DialogTone, string> = {
    danger: 'bg-gradient-to-r from-rose-600 to-rose-700 hover:from-rose-500 hover:to-rose-600 text-white shadow-lg shadow-rose-950/60',
    success: 'bg-gradient-to-r from-emerald-600 to-emerald-700 hover:from-emerald-500 hover:to-emerald-600 text-white shadow-lg shadow-emerald-950/60',
    warning: 'bg-gradient-to-r from-amber-500 to-amber-600 hover:from-amber-400 hover:to-amber-500 text-black shadow-lg shadow-amber-950/60',
    info: 'bg-gradient-to-r from-[#C5A267] to-[#B38F56] hover:from-[#d8b478] hover:to-[#C5A267] text-black shadow-lg shadow-[#C5A267]/20',
  };

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.94, y: 14 }}
      animate={{ opacity: 1, scale: 1, y: 0, transition: { type: 'spring', damping: 26, stiffness: 380 } }}
      exit={{ opacity: 0, scale: 0.96, y: 8, transition: { duration: 0.15 } }}
      onClick={(e) => e.stopPropagation()}
      className={`relative w-full max-w-md overflow-hidden rounded-2xl bg-[#131315] border ${toneGlowClass[tone]} text-left shadow-2xl transition-all`}
      role="dialog"
      aria-modal="true"
    >
      <div
        className={`h-1 w-full bg-gradient-to-r ${
          tone === 'danger'
            ? 'from-rose-500 via-rose-400 to-transparent'
            : tone === 'warning'
            ? 'from-amber-500 via-amber-400 to-transparent'
            : 'from-[#C5A267] via-[#E7CB93] to-transparent'
        }`}
      />

      <div className="p-5 space-y-4">
        <div className="flex items-start gap-3.5">
          <div
            className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-xl border ${
              tone === 'danger'
                ? 'bg-rose-500/10 border-rose-500/20'
                : tone === 'warning'
                ? 'bg-amber-500/10 border-amber-500/20'
                : 'bg-[#C5A267]/10 border-[#C5A267]/20'
            }`}
          >
            {getToneIcon()}
          </div>
          <div className="min-w-0 flex-1">
            <h3 className="text-base font-bold text-white tracking-tight leading-snug">
              {dialog.title}
            </h3>
            {dialog.subtitle && (
              <p className="mt-0.5 text-xs text-gray-400 font-normal">
                {dialog.subtitle}
              </p>
            )}
          </div>
          <button
            onClick={onClose}
            aria-label="Tutup konfirmasi"
            className="rounded-lg p-1.5 text-gray-500 hover:text-white hover:bg-white/5 transition-colors cursor-pointer"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        <div className="rounded-xl border border-white/5 bg-[#0D0D0F] p-3.5 text-xs leading-relaxed text-gray-300">
          {dialog.message}
        </div>

        <div className="flex items-center justify-end gap-2.5 pt-2 border-t border-white/5">
          <button
            type="button"
            onClick={onClose}
            className="rounded-xl border border-white/10 bg-white/5 px-4 py-2 text-xs font-semibold text-gray-300 hover:bg-white/10 hover:text-white transition-all cursor-pointer"
          >
            {dialog.cancelLabel || 'Batal'}
          </button>
          <button
            type="button"
            onClick={() => dialog.resolve(true)}
            className={`rounded-xl px-5 py-2 text-xs font-bold transition-all active:scale-95 cursor-pointer ${toneButtonClass[tone]}`}
          >
            {dialog.confirmLabel || 'Lanjutkan'}
          </button>
        </div>
      </div>
    </motion.div>
  );
}
