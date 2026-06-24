/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import { Module } from '../domain/types';

export interface TeamMember {
  id: string;
  name: string;
  role: 'pm' | 'uiux' | 'frontend' | 'backend';
  email: string;
  avatar: string; // Tailwind bg color class
}

export const MOCK_TEAM_MEMBERS: TeamMember[] = [
  { id: 'tm_1', name: 'Yosia', role: 'pm', email: 'yosia@flowak.com', avatar: 'bg-indigo-500' },
  { id: 'tm_2', name: 'Rian', role: 'uiux', email: 'rian.ux@flowak.com', avatar: 'bg-pink-500' },
  { id: 'tm_3', name: 'Siti', role: 'frontend', email: 'siti.fe@flowak.com', avatar: 'bg-cyan-500' },
  { id: 'tm_4', name: 'Budi', role: 'backend', email: 'budi.be@flowak.com', avatar: 'bg-orange-500' },
  { id: 'tm_5', name: 'Arief', role: 'backend', email: 'arief.be@flowak.com', avatar: 'bg-purple-500' },
];

export const INITIAL_MODULES: Module[] = [
  {
    id: 'mod_leave_request',
    name: 'Sistem Persetujuan Cuti Karyawan',
    description: 'Modul alur kerja untuk pengajuan cuti tahunan, validasi jatah cuti, dan persetujuan oleh manager.',
    schemaVersion: 1,
    nodes: [
      {
        id: 'node_start',
        type: 'terminator',
        label: 'Mulai Pengajuan Cuti',
        x: 100,
        y: 200,
        doc: {
          actor: 'Karyawan',
          input: 'Formulir detail cuti (tanggal, alasan, tipe cuti)',
          process: 'Karyawan mengisi formulir berbasis web',
          output: 'Data pengajuan cuti yang terbuat',
          rules: 'Tanggal mulai harus lebih besar dari hari ini',
          system: 'Flowak Portal',
          sla: '5 menit',
        },
        roles: {
          uiux: {
            assignee: 'Rian',
            status: 'done',
            screen: 'Form Pengajuan Cuti',
            link: 'https://figma.com/file/leave_request_form',
          },
          frontend: {
            assignee: 'Siti',
            status: 'done',
            page: 'Form Pengajuan Cuti',
            route: '/leave/request',
            interaction: 'Karyawan mengisi tanggal, alasan, tipe cuti, lalu mengirim pengajuan.',
            validation: 'Tanggal mulai harus lebih besar dari hari ini dan alasan wajib diisi.',
            state: 'Loading saat submit, success summary saat berhasil, dan error banner saat validasi gagal.',
            handoffLink: 'https://storybook.flowak.local/leave-request',
          },
          backend: {
            assignee: 'Budi',
            status: 'done',
            method: 'POST',
            endpoint: '/api/leave/request',
            auth: 'Bearer Token REQUIRED',
            request: '{\n  "startDate": "2026-07-01",\n  "endDate": "2026-07-05",\n  "type": "annual",\n  "reason": "Family vacation"\n}',
            response: '{\n  "success": true,\n  "requestId": "req_12345",\n  "status": "pending_approval"\n}',
            statusCode: '201',
          },
        },
      },
      {
        id: 'node_validate_quota',
        type: 'process',
        label: 'Validasi Jatah Cuti',
        x: 400,
        y: 200,
        doc: {
          actor: 'Sistem',
          input: 'Request cuti (user_id, jumlah hari)',
          process: 'Mengecek saldo sisa cuti karyawan di database',
          output: 'Kecukupan saldo jatah cuti (Ya/Tidak)',
          rules: 'Saldo cuti tahunan tidak boleh kurang dari nol',
          system: 'HR Service Database',
          sla: 'Instan',
        },
        roles: {
          uiux: {
            assignee: 'Rian',
            status: 'planned',
            screen: 'Indikator Tunggu Sistem / Loader',
            link: '',
          },
          frontend: {
            assignee: 'Siti',
            status: 'in_progress',
            page: 'Indikator Validasi Kuota',
            route: 'Internal Service',
            interaction: 'Menampilkan progress validasi kuota dan hasil kecukupan saldo cuti.',
            validation: 'Tidak ada input manual; status mengikuti hasil kontrak backend.',
            state: 'Loading, quota available, quota unavailable, dan retry.',
            handoffLink: '',
          },
          backend: {
            assignee: 'Budi',
            status: 'done',
            method: 'GET',
            endpoint: '/api/leave/quota/validate?userId=USR_999&days=5',
            auth: 'API Key Shared Secret',
            request: '',
            response: '{\n  "userId": "USR_999",\n  "hasQuota": true,\n  "remainingQuota": 12\n}',
            statusCode: '200',
          },
        },
      },
      {
        id: 'node_decision',
        type: 'decision',
        label: 'Apakah Kuota Cukup?',
        x: 700,
        y: 200,
        doc: {
          actor: 'Sistem',
          input: 'Hasil pengecekan jatah',
          process: 'Mengarahkan alur berdasarkan ketersediaan saldo',
          output: 'Arah alur (Cukup -> Manager Approval, Tidak Cukup -> Penolakan Otomatis)',
          rules: '',
          system: 'Flowak Workflow Engine',
          sla: 'Instan',
        },
        roles: {
          uiux: {
            assignee: 'Rian',
            status: 'review',
            screen: 'Penanganan Kondisi Error / Kuota Habis',
            link: 'https://figma.com/file/error_states',
          },
          frontend: {
            assignee: 'Siti',
            status: 'planned',
            page: 'Notifikasi Kuota Tidak Cukup',
            route: '/leave/error',
            interaction: 'Memberi pilihan kembali ke form atau melihat detail saldo cuti.',
            validation: 'Pesan harus menjelaskan alasan penolakan tanpa istilah teknis.',
            state: 'Error state dan empty state saldo.',
            handoffLink: '',
          },
          backend: {
            assignee: 'Arief',
            status: 'in_progress',
            method: 'POST',
            endpoint: '/api/leave/routing',
            auth: 'Bearer Token',
            request: '{\n  "requestId": "req_12345",\n  "hasQuota": true\n}',
            response: '{\n  "status": "routed",\n  "nextStep": "manager_approval"\n}',
            statusCode: '200',
          },
        },
      },
      {
        id: 'node_manager_approval',
        type: 'actor',
        label: 'Persetujuan Manager',
        x: 1000,
        y: 100,
        doc: {
          actor: 'Manager Divisi',
          input: 'Notifikasi email / aplikasi berisi link persetujuan',
          process: 'Manager meninjau pengajuan cuti dan melakukan approve/reject',
          output: 'Keputusan status pengajuan (Approved/Rejected)',
          rules: 'Harus diselesaikan dalam waktu maksimal 24 jam kerja',
          system: 'Flowak Manager Dashboard',
          sla: '24 jam',
        },
        roles: {
          uiux: {
            assignee: 'Rian',
            status: 'in_progress',
            screen: 'Manager Dashboard - Tindakan Cuti',
            link: 'https://figma.com/file/manager_approval',
          },
          frontend: {
            assignee: 'Siti',
            status: 'planned',
            page: 'Kartu Persetujuan Manager',
            route: '/manager/approvals',
            interaction: 'Manager dapat approve/reject dengan catatan opsional.',
            validation: 'Reject wajib menyertakan alasan.',
            state: 'Pending, approved, rejected, loading, dan expired SLA.',
            handoffLink: '',
          },
          backend: {
            assignee: 'Arief',
            status: 'planned',
            method: 'PUT',
            endpoint: '/api/leave/approve/req_12345',
            auth: 'Bearer Token (Manager role)',
            request: '{\n  "decision": "approved",\n  "notes": "Have a great holiday!"\n}',
            response: '{\n  "success": true,\n  "requestId": "req_12345",\n  "finalStatus": "approved"\n}',
            statusCode: '200',
          },
        },
      },
      {
        id: 'node_auto_reject',
        type: 'process',
        label: 'Penolakan Otomatis',
        x: 1000,
        y: 350,
        doc: {
          actor: 'Sistem',
          input: 'Ketiadaan jatah cuti',
          process: 'Mengirimkan notifikasi penolakan dan mencatat log penolakan',
          output: 'Status pengajuan berubah menjadu REJECTED',
          rules: 'Sistem langsung menolak tanpa peninjauan manual manager',
          system: 'Mail Notification System',
          sla: 'Instan',
        },
        roles: {
          uiux: {
            assignee: 'Rian',
            status: 'done',
            screen: 'Tampilan Penolakan Cuti',
            link: 'https://figma.com/file/rejection_notif',
          },
          frontend: {
            assignee: 'Siti',
            status: 'done',
            page: 'Banner Penolakan Otomatis',
            route: '/leave/rejected',
            interaction: 'User melihat alasan penolakan dan tautan kembali ke riwayat pengajuan.',
            validation: 'Pesan harus konsisten dengan kode alasan backend.',
            state: 'Rejected final state.',
            handoffLink: '',
          },
          backend: {
            assignee: 'Budi',
            status: 'done',
            method: 'POST',
            endpoint: '/api/leave/reject-automatic',
            auth: 'None (Internal Server Only)',
            request: '{\n  "requestId": "req_12345",\n  "reason": "insufficient_quota"\n}',
            response: '{\n  "success": true,\n  "message": "Notification sent to user"\n}',
            statusCode: '200',
          },
        },
      },
    ],
    edges: [
      { id: 'edge_1', from: 'node_start', to: 'node_validate_quota', label: 'Ajukan' },
      { id: 'edge_2', from: 'node_validate_quota', to: 'node_decision', label: 'Periksa' },
      { id: 'edge_3', from: 'node_decision', to: 'node_manager_approval', label: 'Kuota Cukup (Ya)' },
      { id: 'edge_4', from: 'node_decision', to: 'node_auto_reject', label: 'Kuota Kurang (Tidak)' },
    ],
  },
];
