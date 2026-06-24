/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import { create } from 'zustand';
import { Module, Node, Edge, ID, NodeType, RoleKey, BusinessFacet, Status } from '../domain/types';
import { canAddEdge, addEdge as domainAddEdge, deleteNode as domainDeleteNode, uid } from '../domain/invariants';
import { TeamMember, MOCK_TEAM_MEMBERS } from '../config/seedData';

export type AppView = 'canvas' | 'status' | 'doc' | 'calendar' | 'analytics' | 'kanban' | 'team';
export type AppScreen = 'login' | 'register' | 'dashboard' | 'workspace';

export interface NotificationItem {
  id: string;
  message: string;
  timestamp: string;
  read: boolean;
  type: 'info' | 'success' | 'warning';
  title: string;
}

export interface ProjectItem {
  id: string;
  name: string;
  description: string;
  owner_id: string;
  created_at: string;
}

interface AppStore {
  // Authentication State
  token: string | null;
  currentUser: {
    id: string;
    name: string;
    email: string;
    role: 'pm' | 'uiux' | 'frontend' | 'backend';
  } | null;
  isAuthenticated: boolean;
  screen: AppScreen;

  // Project Management State
  projects: ProjectItem[];
  activeProjectId: string | null;

  // Workspace/Module State
  modules: Module[];
  activeId: ID | null;
  selectedNodeId: ID | null;
  view: AppView;
  connectFrom: ID | null;
  darkMode: boolean;
  notifications: NotificationItem[];
  teamMembers: TeamMember[];

  // Actions - UI/Screen routing
  setScreen: (screen: AppScreen) => void;
  initializeStore: () => Promise<void>;
  toggleDarkMode: () => void;

  // Actions - Authentication
  loginUser: (email: string, password: string) => Promise<boolean>;
  registerUser: (name: string, email: string, password: string, role: string) => Promise<boolean>;
  logoutUser: () => void;

  // Actions - Project Management
  loadProjects: () => Promise<void>;
  createProject: (name: string, description: string) => Promise<boolean>;
  deleteProject: (id: string) => Promise<void>;
  selectProject: (id: string | null) => Promise<void>;

  // Actions - Module management
  addModule: (name: string, description?: string) => void;
  renameModule: (id: ID, name: string, description?: string) => void;
  deleteModule: (id: ID) => void;
  selectModule: (id: ID | null) => void;
  setView: (view: AppView) => void;
  
  // Actions - Node management
  addNode: (type: NodeType, label: string) => void;
  moveNode: (id: ID, x: number, y: number) => void;
  updateNode: (id: ID, patch: Partial<Omit<Node, 'id' | 'doc' | 'roles'>>) => void;
  updateDoc: (id: ID, fields: Partial<BusinessFacet>) => void;
  updateRole: (id: ID, role: RoleKey, patch: any) => void;
  deleteNode: (id: ID) => void;
  
  // Actions - Edge management
  addEdge: (from: ID, to: ID, label?: string) => void;
  deleteEdge: (id: ID) => void;
  updateEdgeLabel: (id: ID, label: string) => void;
  
  // Actions - UI Selection
  selectNode: (id: ID | null) => void;
  setConnectFrom: (id: ID | null) => void;
  
  // Actions - Notifications
  addNotification: (title: string, message: string, type?: 'info' | 'success' | 'warning') => void;
  markAllNotificationsRead: () => void;
  clearNotifications: () => void;

  // Actions - AI & Extra modularity actions
  loadAiGeneratedFlow: (name: string, description: string, nodes: Node[], edges: Edge[]) => void;
  importModule: (mod: Module) => void;
  addTeamMember: (name: string, email: string, role: string) => void;
  deleteTeamMember: (id: string) => void;
}

let saveTimeout: any = null;

const debouncedSave = (get: any) => {
  const { activeId, modules, token } = get();
  if (!activeId || !token) return;
  const activeMod = modules.find((m: any) => m.id === activeId);
  if (!activeMod) return;

  if (saveTimeout) clearTimeout(saveTimeout);
  saveTimeout = setTimeout(async () => {
    try {
      await fetch(`/api/modules/${activeId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({
          nodes: activeMod.nodes,
          edges: activeMod.edges
        })
      });
    } catch (err) {
      console.error('Failed to sync canvas updates to server:', err);
    }
  }, 600);
};

export const useStore = create<AppStore>((set, get) => ({
  // Initial Auth State
  token: null,
  currentUser: null,
  isAuthenticated: false,
  screen: 'login',

  // Initial Project State
  projects: [],
  activeProjectId: null,

  // Initial Workspace/Module State
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
      message: 'Aplikasi Flowak Anda siap digunakan. Silakan kelola alur kerja tim Anda.',
      timestamp: new Date().toLocaleTimeString(),
      read: false,
      type: 'info',
    }
  ],

  setScreen: (screen) => {
    set({ screen });
  },

  initializeStore: async () => {
    document.documentElement.classList.add('dark');
    const storedToken = localStorage.getItem('flowak_token');
    const storedUser = localStorage.getItem('flowak_user');

    if (storedToken && storedUser) {
      try {
        const parsedUser = JSON.parse(storedUser);
        set({
          token: storedToken,
          currentUser: parsedUser,
          isAuthenticated: true,
          screen: 'dashboard'
        });
        await get().loadProjects();
      } catch (err) {
        get().logoutUser();
      }
    } else {
      set({ screen: 'login' });
    }
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

  // Authentication Actions
  loginUser: async (email, password) => {
    try {
      const res = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password })
      });

      const data = await res.json();
      if (!res.ok) {
        get().addNotification('Gagal Masuk', data.error || 'Autentikasi gagal', 'warning');
        return false;
      }

      localStorage.setItem('flowak_token', data.token);
      localStorage.setItem('flowak_user', JSON.stringify(data.user));

      set({
        token: data.token,
        currentUser: data.user,
        isAuthenticated: true,
        screen: 'dashboard'
      });

      get().addNotification('Login Sukses', `Selamat datang kembali, ${data.user.name}!`, 'success');
      await get().loadProjects();
      return true;
    } catch (err) {
      get().addNotification('Gagal Masuk', 'Koneksi ke server terputus.', 'warning');
      return false;
    }
  },

  registerUser: async (name, email, password, role) => {
    try {
      const res = await fetch('/api/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, email, password, role })
      });

      const data = await res.json();
      if (!res.ok) {
        get().addNotification('Gagal Mendaftar', data.error || 'Registrasi gagal', 'warning');
        return false;
      }

      get().addNotification('Registrasi Sukses', 'Akun berhasil dibuat. Silakan masuk.', 'success');
      set({ screen: 'login' });
      return true;
    } catch (err) {
      get().addNotification('Gagal Mendaftar', 'Koneksi ke server terputus.', 'warning');
      return false;
    }
  },

  logoutUser: () => {
    localStorage.removeItem('flowak_token');
    localStorage.removeItem('flowak_user');

    set({
      token: null,
      currentUser: null,
      isAuthenticated: false,
      screen: 'login',
      projects: [],
      activeProjectId: null,
      modules: [],
      activeId: null,
      selectedNodeId: null
    });
  },

  // Project Management Actions
  loadProjects: async () => {
    const { token } = get();
    if (!token) return;

    try {
      const res = await fetch('/api/projects', {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      if (res.ok) {
        const data = await res.json();
        set({ projects: data });
      }
    } catch (err) {
      console.error('Failed to load projects:', err);
    }
  },

  createProject: async (name, description) => {
    const { token } = get();
    if (!token) return false;

    try {
      const res = await fetch('/api/projects', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({ name, description })
      });

      const data = await res.json();
      if (res.ok) {
        get().addNotification('Proyek Dibuat', `Proyek "${name}" berhasil ditambahkan.`, 'success');
        await get().loadProjects();
        return true;
      } else {
        get().addNotification('Gagal Membuat Proyek', data.error || 'Terjadi kesalahan', 'warning');
        return false;
      }
    } catch (err) {
      get().addNotification('Gagal Membuat Proyek', 'Koneksi ke server terputus.', 'warning');
      return false;
    }
  },

  deleteProject: async (id) => {
    const { token } = get();
    if (!token) return;

    try {
      const res = await fetch(`/api/projects/${id}`, {
        method: 'DELETE',
        headers: { 'Authorization': `Bearer ${token}` }
      });
      if (res.ok) {
        get().addNotification('Proyek Dihapus', 'Proyek berhasil dihapus.', 'warning');
        await get().loadProjects();
      }
    } catch (err) {
      console.error('Failed to delete project:', err);
    }
  },

  selectProject: async (projectId) => {
    const { token } = get();
    if (!token) return;

    if (!projectId) {
      set({
        activeProjectId: null,
        modules: [],
        activeId: null,
        selectedNodeId: null,
        screen: 'dashboard'
      });
      return;
    }

    try {
      const res = await fetch(`/api/projects/${projectId}`, {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      if (res.ok) {
        const data = await res.json();
        
        // Parse modules nodes and edges from string back to JSON objects
        const parsedModules = data.modules.map((m: any) => ({
          ...m,
          nodes: typeof m.nodes === 'string' ? JSON.parse(m.nodes) : m.nodes,
          edges: typeof m.edges === 'string' ? JSON.parse(m.edges) : m.edges
        }));

        set({
          activeProjectId: projectId,
          modules: parsedModules,
          activeId: parsedModules.length > 0 ? parsedModules[0].id : null,
          selectedNodeId: null,
          screen: 'workspace'
        });
      } else {
        get().addNotification('Akses Ditolak', 'Gagal memuat proyek ini.', 'warning');
      }
    } catch (err) {
      console.error('Failed to select project:', err);
    }
  },

  // Module Management Actions
  addModule: async (name, description) => {
    const { token, activeProjectId } = get();
    if (!token || !activeProjectId) return;

    try {
      const res = await fetch(`/api/projects/${activeProjectId}/modules`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({ name, description })
      });

      const data = await res.json();
      if (res.ok) {
        get().addNotification('Modul Ditambahkan', `Modul "${name}" berhasil dibuat.`, 'success');
        
        // Reload modules
        const modRes = await fetch(`/api/projects/${activeProjectId}/modules`, {
          headers: { 'Authorization': `Bearer ${token}` }
        });
        if (modRes.ok) {
          const modData = await modRes.json();
          const parsedModules = modData.map((m: any) => ({
            ...m,
            nodes: typeof m.nodes === 'string' ? JSON.parse(m.nodes) : m.nodes,
            edges: typeof m.edges === 'string' ? JSON.parse(m.edges) : m.edges
          }));
          set({
            modules: parsedModules,
            activeId: data.module_id,
            selectedNodeId: null
          });
        }
      }
    } catch (err) {
      console.error('Failed to add module:', err);
    }
  },

  renameModule: async (id, name, description) => {
    const { token, activeProjectId } = get();
    if (!token || !activeProjectId) return;

    try {
      const res = await fetch(`/api/modules/${id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({ name, description })
      });

      if (res.ok) {
        set((state) => {
          const updated = state.modules.map((m) =>
            m.id === id ? { ...m, name, description: description !== undefined ? description : m.description } : m
          );
          return { modules: updated };
        });
      }
    } catch (err) {
      console.error('Failed to rename module:', err);
    }
  },

  deleteModule: async (id) => {
    const { token, activeProjectId } = get();
    if (!token || !activeProjectId) return;

    try {
      const res = await fetch(`/api/modules/${id}`, {
        method: 'DELETE',
        headers: { 'Authorization': `Bearer ${token}` }
      });

      if (res.ok) {
        get().addNotification('Modul Dihapus', 'Modul berhasil dihapus secara permanen.', 'warning');
        
        set((state) => {
          const updated = state.modules.filter((m) => m.id !== id);
          const nextActiveId = updated.length > 0 ? updated[0].id : null;
          return {
            modules: updated,
            activeId: nextActiveId,
            selectedNodeId: null
          };
        });
      }
    } catch (err) {
      console.error('Failed to delete module:', err);
    }
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

  // Node Management Actions
  addNode: (type, label) => {
    const { activeId, modules } = get();
    if (!activeId) return;

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
      x: 150 + Math.random() * 100,
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
    debouncedSave(get);

    get().addNotification(
      'Langkah Baru Ditambahkan',
      `Langkah dengan tipe "${type}" bernama "${label}" telah ditambahkan.`,
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
    debouncedSave(get);
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
    debouncedSave(get);
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
    debouncedSave(get);
  },

  updateRole: (id, role, patch) => {
    const { activeId, modules } = get();
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

    // Notify status change
    if (preStatus !== postStatus && postStatus) {
      const activeNode = modules.find((m) => m.id === activeId)?.nodes.find((n) => n.id === id);
      if (activeNode) {
        get().addNotification(
          'Ubah Progres Tugas',
          `Status peran ${role.toUpperCase()} untuk "${activeNode.label}" diperbarui ke [${postStatus.toUpperCase()}].`,
          postStatus === 'done' ? 'success' : 'info'
        );
      }
    }

    set({ modules: updated });
    debouncedSave(get);
  },

  deleteNode: (id) => {
    const { activeId, modules } = get();
    if (!activeId) return;

    const activeMod = modules.find((m) => m.id === activeId);
    if (!activeMod) return;

    const deletingNodeName = activeMod.nodes.find((n) => n.id === id)?.label || 'Langkah';
    
    // Delete logic helper
    const filteredNodes = activeMod.nodes.filter((n) => n.id !== id);
    const filteredEdges = activeMod.edges.filter((e) => e.from !== id && e.to !== id);
    
    const updatedMod = {
      ...activeMod,
      nodes: filteredNodes,
      edges: filteredEdges
    };

    const updated = modules.map((m) => (m.id === activeId ? updatedMod : m));

    set({
      modules: updated,
      selectedNodeId: null,
      connectFrom: null,
    });
    debouncedSave(get);

    get().addNotification(
      'Hapus Langkah Alur',
      `Langkah "${deletingNodeName}" berhasil dihapus.`,
      'warning'
    );
  },

  // Edge Management Actions
  addEdge: (from, to, label) => {
    const { activeId, modules } = get();
    if (!activeId) return;

    const activeMod = modules.find((m) => m.id === activeId);
    if (!activeMod) return;

    const fromNodeLabel = activeMod.nodes.find((n) => n.id === from)?.label || 'Asal';
    const toNodeLabel = activeMod.nodes.find((n) => n.id === to)?.label || 'Tujuan';

    if (!canAddEdge(activeMod, from, to)) {
      get().addNotification(
        'Sambungan Ditolak',
        `Koneksi dari "${fromNodeLabel}" ke "${toNodeLabel}" melanggar aturan.`,
        'warning'
      );
      set({ connectFrom: null });
      return;
    }

    const newEdge: Edge = {
      id: `edge_${uid()}`,
      from,
      to,
      label: label || '',
    };

    const updatedMod = {
      ...activeMod,
      edges: [...activeMod.edges, newEdge]
    };

    const updated = modules.map((m) => (m.id === activeId ? updatedMod : m));

    set({
      modules: updated,
      connectFrom: null,
    });
    debouncedSave(get);

    get().addNotification(
      'Koneksi Berhasil',
      `Menghubungkan "${fromNodeLabel}" ke "${toNodeLabel}".`,
      'success'
    );
  },

  deleteEdge: (id) => {
    const { activeId, modules } = get();
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
    debouncedSave(get);

    get().addNotification(
      'Koneksi Dihapus',
      `Hubungan alur kerja sukses dicabut.`,
      'warning'
    );
  },

  updateEdgeLabel: (id, label) => {
    const { activeId, modules } = get();
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
    debouncedSave(get);
  },

  selectNode: (id) => {
    set({ selectedNodeId: id });
  },

  setConnectFrom: (id) => {
    set({ connectFrom: id });
  },

  // Notification Management Actions
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
      notifications: [newNotif, ...state.notifications].slice(0, 50),
    }));

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

  // AI & Extra modularity actions
  loadAiGeneratedFlow: (name, description, nodes, edges) => {
    const { activeProjectId, token } = get();
    if (!activeProjectId || !token) return;

    // AI flows generated on canvas need to be saved to database
    // We create a new module with the generated flow
    get().addModule(name, description);
    
    // Once activeId is set to new module, we load it in state
    setTimeout(() => {
      const { activeId, modules } = get();
      if (!activeId) return;

      const updated = modules.map((m) => {
        if (m.id === activeId) {
          return {
            ...m,
            nodes,
            edges
          };
        }
        return m;
      });

      set({ modules: updated });
      debouncedSave(get);

      get().addNotification(
        'Alur AI Dimuat',
        `Alur kerja "${name}" dari asisten AI sukses dimuat.`,
        'success'
      );
    }, 1000);
  },

  importModule: async (mod) => {
    const { token, activeProjectId } = get();
    if (!token || !activeProjectId) return;

    if (!mod || !mod.name || !Array.isArray(mod.nodes)) {
      get().addNotification('Impor Gagal', 'Format file JSON modul tidak valid.', 'warning');
      return;
    }

    try {
      const res = await fetch(`/api/projects/${activeProjectId}/modules`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({
          name: mod.name,
          description: mod.description || '',
          nodes: mod.nodes,
          edges: mod.edges || []
        })
      });

      const data = await res.json();
      if (res.ok) {
        get().addNotification('Modul Diimpor', `Modul "${mod.name}" berhasil diimpor.`, 'success');
        
        // Reload modules
        const modRes = await fetch(`/api/projects/${activeProjectId}/modules`, {
          headers: { 'Authorization': `Bearer ${token}` }
        });
        if (modRes.ok) {
          const modData = await modRes.json();
          const parsedModules = modData.map((m: any) => ({
            ...m,
            nodes: typeof m.nodes === 'string' ? JSON.parse(m.nodes) : m.nodes,
            edges: typeof m.edges === 'string' ? JSON.parse(m.edges) : m.edges
          }));
          set({
            modules: parsedModules,
            activeId: data.module_id,
            selectedNodeId: null
          });
        }
      }
    } catch (err) {
      console.error('Failed to import module:', err);
    }
  },

  addTeamMember: (name, email, role) => {
    const { teamMembers } = get();
    const newMember: TeamMember = {
      id: `member_${uid()}`,
      name,
      email,
      role: role as 'pm' | 'uiux' | 'frontend' | 'backend',
      avatar: `https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=100&fit=crop`,
    };
    const updated = [...teamMembers, newMember];
    set({ teamMembers: updated });
    get().addNotification('Anggota Tim Ditambahkan', `${name} dimasukkan ke daftar kontributor sebagai ${role.toUpperCase()}.`, 'success');
  },

  deleteTeamMember: (id) => {
    const { teamMembers } = get();
    const target = teamMembers.find(m => m.id === id);
    if (!target) return;
    const updated = teamMembers.filter(m => m.id !== id);
    set({ teamMembers: updated });
    get().addNotification('Anggota Tim Dihentikan', `${target.name} telah dikeluarkan dari panel kontributor.`, 'warning');
  }
}));
