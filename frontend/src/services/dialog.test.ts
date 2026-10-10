import { describe, expect, it } from 'vitest';
import { dialog, useDialogStore } from './dialog';

describe('dialog service', () => {
  it('opens a prompt dialog and resolves with the inputted value', async () => {
    const promise = dialog.prompt({
      title: 'Alasan Task Diblokir',
      badge: 'Status: Blocked',
      tone: 'danger',
    });

    const active = useDialogStore.getState().activeDialog;
    expect(active).not.toBeNull();
    expect(active?.type).toBe('prompt');
    if (active?.type === 'prompt') {
      expect(active.title).toBe('Alasan Task Diblokir');
      expect(active.tone).toBe('danger');
      active.resolve('Menunggu API Backend selesai');
    }

    const result = await promise;
    expect(result).toBe('Menunggu API Backend selesai');
    expect(useDialogStore.getState().activeDialog).toBeNull();
  });

  it('resolves prompt with null on close/cancel', async () => {
    const promise = dialog.prompt({
      title: 'Resolusi Penyelesaian',
      badge: 'Status: Done',
      tone: 'success',
    });

    dialog.close();
    const result = await promise;
    expect(result).toBeNull();
    expect(useDialogStore.getState().activeDialog).toBeNull();
  });

  it('opens a confirm dialog and resolves boolean values', async () => {
    const promise = dialog.confirm({
      title: 'Hapus Modul',
      message: 'Yakin ingin menghapus modul?',
      tone: 'danger',
    });

    const active = useDialogStore.getState().activeDialog;
    expect(active?.type).toBe('confirm');
    if (active?.type === 'confirm') {
      active.resolve(true);
    }

    const result = await promise;
    expect(result).toBe(true);
    expect(useDialogStore.getState().activeDialog).toBeNull();
  });

  it('resolves confirm with false on close/cancel', async () => {
    const promise = dialog.confirm({
      title: 'Arsipkan Proyek',
      message: 'Yakin arsipkan proyek?',
    });

    dialog.close();
    const result = await promise;
    expect(result).toBe(false);
    expect(useDialogStore.getState().activeDialog).toBeNull();
  });
});
