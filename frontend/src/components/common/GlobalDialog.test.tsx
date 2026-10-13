import { renderToStaticMarkup } from 'react-dom/server';
import { describe, expect, it } from 'vitest';
import GlobalDialog, { PromptDialogContent, ConfirmDialogContent } from './GlobalDialog';
import { useDialogStore } from '../../services/dialog';

describe('GlobalDialog markup and accessibility', () => {
  it('renders nothing when activeDialog is null', () => {
    useDialogStore.setState({ activeDialog: null });
    const html = renderToStaticMarkup(<GlobalDialog />);
    expect(html).toBe('');
  });

  it('renders prompt dialog with accessible role and quick options', () => {
    const dialogConfig = {
      type: 'prompt' as const,
      title: 'Alasan Task Diblokir',
      subtitle: 'Keterangan kendala',
      placeholder: 'Tuliskan alasan...',
      badge: 'Status: Blocked',
      tone: 'danger' as const,
      quickOptions: ['Menunggu API Backend', 'Bug Kritis'],
      resolve: () => undefined,
    };

    const html = renderToStaticMarkup(<PromptDialogContent dialog={dialogConfig} onClose={() => undefined} />);
    expect(html).toContain('role="dialog"');
    expect(html).toContain('aria-modal="true"');
    expect(html).toContain('Alasan Task Diblokir');
    expect(html).toContain('Status: Blocked');
    expect(html).toContain('Menunggu API Backend');
    expect(html).toContain('Bug Kritis');
  });

  it('renders confirmation dialog with message and action buttons', () => {
    const dialogConfig = {
      type: 'confirm' as const,
      title: 'Hapus Modul',
      message: 'Apakah Anda yakin ingin menghapus modul ini?',
      confirmLabel: 'Hapus Modul',
      tone: 'danger' as const,
      resolve: () => undefined,
    };

    const html = renderToStaticMarkup(<ConfirmDialogContent dialog={dialogConfig} onClose={() => undefined} />);
    expect(html).toContain('role="dialog"');
    expect(html).toContain('Hapus Modul');
    expect(html).toContain('Apakah Anda yakin ingin menghapus modul ini?');
  });
});
