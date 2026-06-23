/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import { create } from 'zustand';
import { Module, Node, Edge, ID, NodeType, RoleKey, BusinessFacet, Status } from '../domain/types';
import { canAddEdge, addEdge as domainAddEdge, deleteNode as domainDeleteNode, assertUniqueNodeId, uid } from '../domain/invariants';
import { persistenceAdapter } from '../infra/persistence';
import { INITIAL_MODULES, TeamMember, MOCK_TEAM_MEMBERS } from '../config/seedData';

export type AppView = 'canvas' | 'status' | 'doc' | 'calendar' | 'analytics' | 'kanban' | 'team';

export interface NotificationItem {
  id: string;
  message: string;
  timestamp: string;
  read: boolean;
  type: 'info' | 'success' | 'warning';
  title: string;
}

interface AppStore {
  // State
  modules: Module[];
  activeId: ID | null;
  selectedNodeId: ID | null;
  view: AppView;
  connectFrom: ID | null;
  darkMode: boolean;
  notifications: NotificationItem[];
  teamMembers: TeamMember[];

  // Actions
  initializeStore: () => Promise<void>;
  addModule: (name: string, description?: string) => void;
  renameModule: (id: ID, name: string, description?: string) => void;
  deleteModule: (id: ID) => void;
  selectModule: (id: ID | null) => void;
  setView: (view: AppView) => void;
  
  // Node management
  addNode: (type: NodeType, label: string) => void;
  moveNode: (id: ID, x: number, y: number) => void;
  updateNode: (id: ID, patch: Partial<Omit<Node, 'id' | 'doc' | 'roles'>>) => void;
  updateDoc: (id: ID, fields: Partial<BusinessFacet>) => void;
  updateRole: (id: ID, role: RoleKey, patch: any) => void;
  deleteNode: (id: ID) => void;
  
  // Edge management
  addEdge: (from: ID, to: ID, label?: string) => void;
  deleteEdge: (id: ID) => void;
  updateEdgeLabel: (id: ID, label: string) => void;
  
  // UI Selection
  selectNode: (id: ID | null) => void;
  setConnectFrom: (id: ID | null) => void;
  toggleDarkMode: () => void;
  
  // Notifications
  addNotification: (title: string, message: string, type?: 'info' | 'success' | 'warning') => void;
  markAllNotificationsRead: () => void;
  clearNotifications: () => void;

  // AI & Extra modularity actions
  loadAiGeneratedFlow: (name: string, description: string, nodes: Node[], edges: Edge[]) => void;
  importModule: (mod: Module) => void;
  addTeamMember: (name: string, email: string, role: string) => void;
  deleteTeamMember: (id: string) => void;
}

let saveTimeout: any = null;

const debouncedSave = (modules: Module[]) => {
  if (saveTimeout) clearTimeout(saveTimeout);
  saveTimeout = setTimeout(() => {
    persistenceAdapter.save(modules);
  }, 600);
};

export const useStore = create<AppStore>((set, get) => ({
  // Initial values
  modules: [],
  activeId: null,
  selectedNodeId: null,
  view: 'canvas',
  connectFrom: null,
  darkMode: true,
  teamMembers: MOCK_TEAM_MEMBERS,
  notifications: [
    {
      id: 'notif_1',
      title: 'Selamat Datang!',
      message: 'Aplikasi Flowak Anda siap digunakan. Cek modul contoh di modul manager.',
      timestamp: new Date().toLocaleTimeString(),
      read: false,
      type: 'info',
    }
  ],

  initializeStore: async () => {
    const loadedModules = await persistenceAdapter.load();
    document.documentElement.classList.add('dark');
    if (loadedModules && loadedModules.length > 0) {
      set({
        modules: loadedModules,
        activeId: loadedModules[0].id,
        darkMode: true,
      });
    } else {
      // Seed initial module
      set({
        modules: INITIAL_MODULES,
        activeId: INITIAL_MODULES[0].id,
        darkMode: true,
      });
      await persistenceAdapter.save(INITIAL_MODULES);
    }
  },

  addModule: (name, description) => {
    const newMod: Module = {
      id: `mod_${uid()}`,
      name,
      description: description || '',
      nodes: [],
      edges: [],
      schemaVersion: 1,
    };
    set((state) => {
      const updated = [...state.modules, newMod];
      debouncedSave(updated);
      state.addNotification(
        'Modul Baru Dibuat',
        `Modul "${name}" berhasil dibuat secara sukses dan siap dirancang.`,
        'success'
      );
      return {
        modules: updated,
        activeId: newMod.id,
        selectedNodeId: null,
        connectFrom: null,
      };
    });
  },

  renameModule: (id, name, description) => {
    set((state) => {
      const updated = state.modules.map((m) =>
        m.id === id ? { ...m, name, description: description !== undefined ? description : m.description } : m
      );
      debouncedSave(updated);
      return { modules: updated };
    });
  },

  deleteModule: (id) => {
    set((state) => {
      const activeMod = state.modules.find((m) => m.id === id);
      const updated = state.modules.filter((m) => m.id !== id);
      debouncedSave(updated);
      const nextActiveId = updated.length > 0 ? updated[0].id : null;
      state.addNotification(
        'Modul Dihapus',
        `Modul "${activeMod?.name || ''}" telah berhasil dihapus secara permanen.`,
        'warning'
      );
      return {
        modules: updated,
        activeId: nextActiveId,
        selectedNodeId: null,
        connectFrom: null,
      };
    });
  },

  selectModule: (id) => {
    set({
      activeId: id,
      selectedNodeId: null,
      connectFrom: null,
    });
  },

  setView: (view) => {
    set({ view });
  },

  addNode: (type, label) => {
    const { activeId, modules, addNotification } = get();
    if (!activeId) return;

    // Determine default assignee based on node type
    let defaultAssignee = MOCK_TEAM_MEMBERS[1].name; // Rian (UI/UX)
    if (type === 'system' || type === 'decision') {
      defaultAssignee = MOCK_TEAM_MEMBERS[3].name; // Budi (BE)
    } else if (type === 'process') {
      defaultAssignee = MOCK_TEAM_MEMBERS[2].name; // Siti (FE)
    }

    const newNode: Node = {
      id: `node_${uid()}`,
      type,
      label,
      x: 150 + Math.random() * 100, // Small random offset to avoid overlapping
      y: 150 + Math.random() * 100,
      doc: {
        actor: type === 'actor' ? 'Petugas / User' : 'Sistem',
        input: '',
        process: '',
        output: '',
        rules: '',
        system: type === 'system' ? 'Aplikasi Gateway' : 'FlowakPortal',
        sla: 'Instan',
      },
      roles: {
        uiux: {
          assignee: MOCK_TEAM_MEMBERS[1].name,
          status: 'planned',
          screen: '',
          link: '',
        },
        frontend: {
          assignee: MOCK_TEAM_MEMBERS[2].name,
          status: 'planned',
          component: '',
          route: '',
          framework: 'React',
          link: '',
        },
        backend: {
          assignee: type === 'system' ? MOCK_TEAM_MEMBERS[4].name : MOCK_TEAM_MEMBERS[3].name,
          status: 'planned',
          method: 'POST',
          endpoint: '',
          auth: '',
          request: '',
          response: '',
          statusCode: '200',
        },
      },
    };

    const updated = modules.map((m) => {
      if (m.id === activeId) {
        return {
          ...m,
          nodes: [...m.nodes, newNode],
        };
      }
      return m;
    });

    set({ modules: updated, selectedNodeId: newNode.id });
    debouncedSave(updated);

    addNotification(
      'Langkah Baru Ditambahkan',
      `Langkah dengan tipe "${type}" bernama "${label}" telah ditambahkan ke alur kerja.`,
      'success'
    );
  },

  moveNode: (id, x, y) => {
    const { activeId, modules } = get();
    if (!activeId) return;

    const updated = modules.map((m) => {
      if (m.id === activeId) {
        return {
          ...m,
          nodes: m.nodes.map((n) => (n.id === id ? { ...n, x: Math.max(0, x), y: Math.max(0, y) } : n)),
        };
      }
      return m;
    });

    set({ modules: updated });
    debouncedSave(updated);
  },

  updateNode: (id, patch) => {
    const { activeId, modules } = get();
    if (!activeId) return;

    const updated = modules.map((m) => {
      if (m.id === activeId) {
        return {
          ...m,
          nodes: m.nodes.map((n) => (n.id === id ? { ...n, ...patch } : n)),
        };
      }
      return m;
    });

    set({ modules: updated });
    debouncedSave(updated);
  },

  updateDoc: (id, fields) => {
    const { activeId, modules } = get();
    if (!activeId) return;

    const updated = modules.map((m) => {
      if (m.id === activeId) {
        return {
          ...m,
          nodes: m.nodes.map((n) => (n.id === id ? { ...n, doc: { ...n.doc, ...fields } } : n)),
        };
      }
      return m;
    });

    set({ modules: updated });
    debouncedSave(updated);
  },

  updateRole: (id, role, patch) => {
    const { activeId, modules, addNotification } = get();
    if (!activeId) return;

    let preStatus: Status | undefined;
    let postStatus: Status | undefined;

    const updated = modules.map((m) => {
      if (m.id === activeId) {
        return {
          ...m,
          nodes: m.nodes.map((n) => {
            if (n.id === id) {
              const currentRoleData = n.roles[role] || {};
              preStatus = currentRoleData.status;
              postStatus = patch.status !== undefined ? patch.status : preStatus;
              
              const updatedRoles = {
                ...n.roles,
                [role]: {
                  ...currentRoleData,
                  ...patch,
                },
              };
              return { ...n, roles: updatedRoles };
            }
            return n;
          }),
        };
      }
      return m;
    });

    // Notify on status change
    if (preStatus !== postStatus && postStatus) {
      const activeNode = modules.find((m) => m.id === activeId)?.nodes.find((n) => n.id === id);
      if (activeNode) {
        addNotification(
          'Ubah Progres Tugas',
          `Status peran ${role.toUpperCase()} untuk langkah "${activeNode.label}" diperbarui dari [${preStatus || 'Belum Ada'}] ke [${postStatus.toUpperCase()}].`,
          postStatus === 'done' ? 'success' : 'info'
        );
      }
    }

    set({ modules: updated });
    debouncedSave(updated);
  },

  deleteNode: (id) => {
    const { activeId, modules, addNotification } = get();
    if (!activeId) return;

    const activeMod = modules.find((m) => m.id === activeId);
    if (!activeMod) return;

    const deletingNodeName = activeMod.nodes.find((n) => n.id === id)?.label || 'Langkah';
    const updatedMod = domainDeleteNode(activeMod, id);

    const updated = modules.map((m) => (m.id === activeId ? updatedMod : m));

    set({
      modules: updated,
      selectedNodeId: null,
      connectFrom: null,
    });
    debouncedSave(updated);

    addNotification(
      'Hapus Langkah Alur',
      `Langkah "${deletingNodeName}" beserta relasi sambungannya berhasil dihapus.`,
      'warning'
    );
  },

  addEdge: (from, to, label) => {
    const { activeId, modules, addNotification } = get();
    if (!activeId) return;

    const activeMod = modules.find((m) => m.id === activeId);
    if (!activeMod) return;

    const fromNodeLabel = activeMod.nodes.find((n) => n.id === from)?.label || 'Asal';
    const toNodeLabel = activeMod.nodes.find((n) => n.id === to)?.label || 'Tujuan';

    if (!canAddEdge(activeMod, from, to)) {
      addNotification(
        'Sambungan Ditolak',
        `Koneksi dari "${fromNodeLabel}" ke "${toNodeLabel}" melanggar batas aturan (duplikat atau putaran mandiri).`,
        'warning'
      );
      set({ connectFrom: null });
      return;
    }

    const updatedMod = domainAddEdge(activeMod, from, to, label);
    const updated = modules.map((m) => (m.id === activeId ? updatedMod : m));

    set({
      modules: updated,
      connectFrom: null,
    });
    debouncedSave(updated);

    addNotification(
      'Koneksi Berhasil',
      `Menghubungkan "${fromNodeLabel}" ke "${toNodeLabel}" dengan sukses.`,
      'success'
    );
  },

  deleteEdge: (id) => {
    const { activeId, modules, addNotification } = get();
    if (!activeId) return;

    const updated = modules.map((m) => {
      if (m.id === activeId) {
        return {
          ...m,
          edges: m.edges.filter((e) => e.id !== id),
        };
      }
      return m;
    });

    set({ modules: updated });
    debouncedSave(updated);

    addNotification(
      'Koneksi Dihapus',
      `Hubungan alur kerja sukses dicabut.`,
      'warning'
    );
  },

  updateEdgeLabel: (id, label) => {
    const { activeId, modules, addNotification } = get();
    if (!activeId) return;

    const updated = modules.map((m) => {
      if (m.id === activeId) {
        return {
          ...m,
          edges: m.edges.map((e) => (e.id === id ? { ...e, label } : e)),
        };
      }
      return m;
    });

    set({ modules: updated });
    debouncedSave(updated);

    addNotification(
      'Label Koneksi Diperbarui',
      `Label hubungan alur kerja diperbarui menjadi "${label || '(Tanpa Label)'}".`,
      'success'
    );
  },

  selectNode: (id) => {
    set({ selectedNodeId: id });
  },

  setConnectFrom: (id) => {
    set({ connectFrom: id });
  },

  toggleDarkMode: () => {
    set((state) => {
      const mode = !state.darkMode;
      if (mode) {
        document.documentElement.classList.add('dark');
      } else {
        document.documentElement.classList.remove('dark');
      }
      return { darkMode: mode };
    });
  },

  addNotification: (title, message, type = 'info') => {
    const newNotif: NotificationItem = {
      id: `notif_${uid()}`,
      title,
      message,
      timestamp: new Date().toLocaleTimeString(),
      read: false,
      type,
    };
    set((state) => ({
      notifications: [newNotif, ...state.notifications].slice(0, 50), // Cap at 50
    }));

    // Trigger standard Notification API simulation
    if (Notification.permission === 'granted') {
      try {
        new Notification(`Flowak: ${title}`, {
          body: message,
          tag: 'flowak-notif',
        });
      } catch (err) {
        console.log('Push simulated via web notification');
      }
    }
  },

  markAllNotificationsRead: () => {
    set((state) => ({
      notifications: state.notifications.map((n) => ({ ...n, read: true })),
    }));
  },

  clearNotifications: () => {
    set({ notifications: [] });
  },

  loadAiGeneratedFlow: (name, description, nodes, edges) => {
    const { modules, addNotification } = get();
    const newModuleId = `mod_${uid()}`;
    const newMod: Module = {
      id: newModuleId,
      name,
      description,
      nodes,
      edges,
      schemaVersion: 1,
    };
    
    const updated = [...modules, newMod];
    set({
      modules: updated,
      activeId: newModuleId,
      selectedNodeId: null,
      connectFrom: null,
    });
    debouncedSave(updated);

    addNotification(
      'Alur Kerja AI Berhasil',
      `Modul "${name}" berhasil dirancang menggunakan asisten AI.`,
      'success'
    );
  },

  importModule: (mod) => {
    const { modules, addNotification } = get();
    // Validate module
    if (!mod || !mod.name || !Array.isArray(mod.nodes)) {
      addNotification('Impor Gagal', 'Format file JSON modul tidak valid.', 'warning');
      return;
    }
    
    // Assign generic ID if not matching format to prevent collisions
    const importedId = mod.id && !modules.find(m => m.id === mod.id) ? mod.id : `mod_imported_${uid()}`;
    const newMod: Module = {
      ...mod,
      id: importedId,
    };
    
    const updated = [...modules, newMod];
    set({
      modules: updated,
      activeId: importedId,
      selectedNodeId: null,
      connectFrom: null,
    });
    debouncedSave(updated);

    addNotification(
      'Modul Berhasil Diimpor',
      `Modul "${mod.name}" berhasil dimuat ke dalam workspace.`,
      'success'
    );
  },

  addTeamMember: (name, email, role) => {
    const { teamMembers, addNotification } = get();
    const newMember: TeamMember = {
      id: `member_${uid()}`,
      name,
      email,
      role: role as 'pm' | 'uiux' | 'frontend' | 'backend',
      avatar: `https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=100&fit=crop`,
    };
    const updated = [...teamMembers, newMember];
    set({ teamMembers: updated });
    addNotification('Anggota Tim Ditambahkan', `${name} dimasukkan ke daftar kontributor sebagai ${role.toUpperCase()}.`, 'success');
  },

  deleteTeamMember: (id) => {
    const { teamMembers, addNotification } = get();
    const target = teamMembers.find(m => m.id === id);
    if (!target) return;
    const updated = teamMembers.filter(m => m.id !== id);
    set({ teamMembers: updated });
    addNotification('Anggota Tim Dihentikan', `${target.name} telah dikeluarkan dari panel kontributor.`, 'warning');
  }
}));
