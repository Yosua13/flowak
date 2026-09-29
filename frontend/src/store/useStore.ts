/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import { create } from 'zustand';
import { Module, Node, Edge, ID, NodeType, RoleKey, BusinessFacet, Status } from '../domain/types';
import { canAddEdge, addEdge as domainAddEdge, uid } from '../domain/invariants';
import { TeamMember } from '../config/seedData';
import { apiClient } from '../services/apiClient';
import { saveGraph, SaveStatus } from '../services/graphMutation';
import { authService } from '../services/authService';
import { projectsService } from '../services/projects';
import { membersService } from '../services/members';
import { notificationsService, type NotificationItem } from '../services/notifications';
import { persistenceAdapter } from '../infra/persistence';

export type AppView = 'canvas' | 'status' | 'doc' | 'calendar' | 'analytics' | 'kanban' | 'team';
export type AppScreen = 'login' | 'register' | 'dashboard' | 'workspace';

export type { NotificationItem } from '../services/notifications';

export interface ProjectItem {
  id: string;
  name: string;
  description: string;
  owner_id: string;
  status: 'active' | 'archived';
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
  organizationRole: 'owner' | 'member' | null;
  screen: AppScreen;

  // Project Management State
  projects: ProjectItem[];
  archivedProjects: ProjectItem[];
  activeProjectId: string | null;

  // Workspace/Module State
  modules: Module[];
  activeId: ID | null;
  selectedNodeId: ID | null;
  selectedWorkItemKey: string | null;
  view: AppView;
  connectFrom: ID | null;
  darkMode: boolean;
  notifications: NotificationItem[];
  selectedNotif: NotificationItem | null;
  teamMembers: TeamMember[];
  projectMembers: TeamMember[];
  dashboardStats: { myTasksCount: number; completionRate: number } | null;
  saveStatus: SaveStatus;
  saveError: string | null;
  setSaveStatus: (status: SaveStatus) => void;
  retryActiveModuleSave: () => Promise<void>;
  reloadActiveProject: () => Promise<void>;

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
  loadArchivedProjects: () => Promise<void>;
  createProject: (name: string, description: string) => Promise<boolean>;
  deleteProject: (id: string) => Promise<void>;
  restoreProject: (id: string) => Promise<void>;
  selectProject: (id: string | null) => Promise<void>;

  // Actions - Module management
  addModule: (name: string, description?: string) => Promise<ID | null>;
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
  selectWorkItem: (key: string | null) => void;
  setConnectFrom: (id: ID | null) => void;
  
  // Actions - Notifications
  addNotification: (title: string, message: string, type?: 'info' | 'success' | 'warning') => void;
  setSelectedNotif: (notif: NotificationItem | null) => void;
  markAllNotificationsRead: () => void;
  clearNotifications: () => void;
  loadNotifications: () => Promise<void>;

  // Actions - AI & Extra modularity actions
  loadAiGeneratedFlow: (name: string, description: string, nodes: Node[], edges: Edge[]) => Promise<void>;
  importModule: (mod: Module) => void;
  
  // Actions - Team Management API
  loadTeamMembers: () => Promise<void>;
  addTeamMember: (name: string, email: string, role: string) => Promise<void>;
  deleteTeamMember: (id: string) => Promise<void>;

  // Actions - Project Members Management API
  loadProjectMembers: () => Promise<void>;
  addProjectMember: (userId: string) => Promise<boolean>;
  deleteProjectMember: (userId: string) => Promise<boolean>;
  loadDashboardStats: () => Promise<void>;
}

let saveTimeout: ReturnType<typeof setTimeout> | null = null;
let graphRequest: AbortController | null = null;
let projectRequest: AbortController | null = null;
const confirmedGraphs = new Map<ID, Module>();

const persistActiveModule = async (set: any, get: any, moduleId: ID) => {
  const latest = get();
  const activeMod = latest.modules.find((module: Module) => module.id === moduleId);
  if (!activeMod || !latest.token) return;

  set({ saveStatus: 'saving', saveError: null });
  try {
    const result = await saveGraph(activeMod, graphRequest?.signal);
    if (Array.isArray(result.nodes) && Array.isArray(result.edges)) {
      set((state: any) => ({
        modules: state.modules.map((module: Module) => module.id === moduleId
          ? { ...module, nodes: result.nodes, edges: result.edges, deletedNodes: [], deletedEdges: [] }
          : module),
        saveStatus: 'saved'
      }));
    } else {
      set({ saveStatus: 'saved' });
    }
  } catch (err: any) {
    const rollback = confirmedGraphs.get(moduleId);
    if (rollback) set((state: any) => ({ modules: state.modules.map((module: Module) => module.id === moduleId ? rollback : module) }));
    set({ saveStatus: err?.code === 'conflict' ? 'conflict' : typeof navigator !== 'undefined' && !navigator.onLine ? 'offline' : 'failed', saveError: err instanceof Error ? err.message : 'Validasi spesifikasi gagal.' });
    console.error('Failed to sync canvas updates to server:', err);
  }
};

const debouncedSave = (set: any, get: any) => {
  const { activeId, token } = get();
  if (!activeId || !token) return;
  set({ saveStatus: 'saving', saveError: null });
	if (saveTimeout) clearTimeout(saveTimeout);
	saveTimeout = setTimeout(async () => {
		await persistActiveModule(set, get, activeId);
  }, 600);
};

export const useStore = create<AppStore>((set, get) => ({
  // Initial Auth State
  token: null,
  currentUser: null,
  isAuthenticated: false,
  organizationRole: null,
  screen: 'login',

  // Initial Project State
  projects: [],
  archivedProjects: [],
  activeProjectId: null,

  // Initial Workspace/Module State
  modules: [],
  activeId: null,
  selectedNodeId: null,
  selectedWorkItemKey: null,
  view: 'canvas',
  connectFrom: null,
  darkMode: true,
  teamMembers: [],
  projectMembers: [],
  dashboardStats: null,
  saveStatus: 'idle',
  saveError: null,
  setSaveStatus: (saveStatus) => set({ saveStatus }),
  retryActiveModuleSave: async () => {
    const { activeId } = get();
    if (activeId) await persistActiveModule(set, get, activeId);
  },
  reloadActiveProject: async () => {
    const { activeProjectId } = get();
    if (activeProjectId) await get().selectProject(activeProjectId);
  },
  selectedNotif: null,
  notifications: [],

  setScreen: (screen) => {
    set({ screen });
  },

  initializeStore: async () => {
    document.documentElement.classList.add('dark');
    const storedUser = localStorage.getItem('flowak_user');

    if (storedUser) {
      try {
        const session = await authService.refresh();
        const parsedUser = session.user;
        apiClient.setToken(session.token);
        set({
          token: session.token,
          currentUser: parsedUser,
          organizationRole: session.organization_role,
          isAuthenticated: true,
          screen: 'dashboard'
        });
        await get().loadProjects();
        await get().loadArchivedProjects();
        await get().loadTeamMembers();
        await get().loadNotifications();
      } catch {
        set({ screen: 'login' });
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
      persistenceAdapter.savePreferences({ darkMode: mode });
      return { darkMode: mode };
    });
  },

  // Authentication Actions
  loginUser: async (email, password) => {
    try {
      const data = await authService.login(email, password);

      apiClient.setToken(data.token);
      localStorage.setItem('flowak_user', JSON.stringify(data.user));

      set({
        token: data.token,
        currentUser: data.user,
        organizationRole: data.organization_role,
        isAuthenticated: true,
        screen: 'dashboard'
      });

      get().addNotification('Login Sukses', `Selamat datang kembali, ${data.user.name}!`, 'success');
      await get().loadProjects();
      await get().loadArchivedProjects();
      await get().loadTeamMembers();
      await get().loadNotifications();
      return true;
    } catch (err) {
      get().addNotification('Gagal Masuk', 'Koneksi ke server terputus.', 'warning');
      return false;
    }
  },

  registerUser: async (name, email, password, role) => {
    try {
      await authService.register(name, email, password, role);

      get().addNotification('Registrasi Sukses', 'Akun berhasil dibuat. Silakan masuk.', 'success');
      set({ screen: 'login' });
      return true;
    } catch (err) {
      get().addNotification('Gagal Mendaftar', 'Koneksi ke server terputus.', 'warning');
      return false;
    }
  },

  logoutUser: () => {
    void authService.logout();
    apiClient.setToken(null);
    localStorage.removeItem('flowak_user');

    set({
      token: null,
      currentUser: null,
      isAuthenticated: false,
      organizationRole: null,
      screen: 'login',
      projects: [],
      archivedProjects: [],
      activeProjectId: null,
      modules: [],
      activeId: null,
      selectedNodeId: null,
      teamMembers: [],
      projectMembers: [],
      dashboardStats: null
    });
  },

  // Project Management Actions
  loadProjects: async () => {
    const { token } = get();
    if (!token) return;

    try {
      const data = await projectsService.list();
      if (data) {
        set({ projects: data });
        // Fetch dashboard stats as well
        get().loadDashboardStats();
      }
    } catch (err) {
      console.error('Failed to load projects:', err);
    }
  },

  loadArchivedProjects: async () => {
    const { token } = get();
    if (!token) return;

    try {
      set({ archivedProjects: await projectsService.list('archived') });
    } catch (err) {
      console.error('Failed to load archived projects:', err);
    }
  },

  createProject: async (name, description) => {
    const { token } = get();
    if (!token) return false;

    try {
      await projectsService.create(name, description);
      get().addNotification('Proyek Dibuat', `Proyek "${name}" berhasil ditambahkan.`, 'success');
      await get().loadProjects();
      return true;
    } catch (err) {
      get().addNotification('Gagal Membuat Proyek', 'Koneksi ke server terputus.', 'warning');
      return false;
    }
  },

  deleteProject: async (id) => {
    const { token } = get();
    if (!token) return;

    try {
      await projectsService.archive(id);
		get().addNotification('Proyek Diarsipkan', 'Proyek dapat dipulihkan dari daftar arsip.', 'warning');
        await get().loadProjects();
		await get().loadArchivedProjects();
    } catch (err) {
      console.error('Failed to delete project:', err);
    }
  },

  restoreProject: async (id) => {
    const { token } = get();
    if (!token) return;

    try {
      await projectsService.restore(id);
      get().addNotification('Proyek Dipulihkan', 'Proyek kembali tersedia di workspace.', 'success');
      await get().loadProjects();
      await get().loadArchivedProjects();
    } catch (err) {
      get().addNotification('Gagal Memulihkan Proyek', 'Koneksi ke server terputus.', 'warning');
    }
  },

  selectProject: async (projectId) => {
    const { token } = get();
    if (!token) return;

    if (!projectId) {
      graphRequest?.abort();
      projectRequest?.abort();
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
      projectRequest?.abort();
      projectRequest = apiClient.abortable();
      const data = await projectsService.detail(projectId, projectRequest.signal);
      if (data) {
        const parsedModules = data.modules;
        parsedModules.forEach((module: Module) => confirmedGraphs.set(module.id, module));

        set({
          activeProjectId: projectId,
          modules: parsedModules,
          activeId: parsedModules.length > 0 ? parsedModules[0].id : null,
          selectedNodeId: null,
          screen: 'workspace'
        });

        // Sync team list from DB
        await get().loadTeamMembers();
        await get().loadProjectMembers();
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
    if (!token || !activeProjectId) return null;

    try {
      const data = await projectsService.createModule(activeProjectId, { name, description });
      get().addNotification('Modul Ditambahkan', `Modul "${name}" berhasil dibuat.`, 'success');
      const parsedModules = await projectsService.listModules(activeProjectId);
      set({ modules: parsedModules, activeId: data.module_id, selectedNodeId: null });
      return data.module_id;
    } catch (err) {
      console.error('Failed to add module:', err);
      return null;
    }
  },

  renameModule: async (id, name, description) => {
    const { token, activeProjectId } = get();
    if (!token || !activeProjectId) return;

    try {
      await projectsService.renameModule(id, name, description);
      set((state) => {
        const updated = state.modules.map((m) =>
          m.id === id ? { ...m, name, description: description !== undefined ? description : m.description } : m
        );
        return { modules: updated };
      });
    } catch (err) {
      console.error('Failed to rename module:', err);
    }
  },

  deleteModule: async (id) => {
    const { token, activeProjectId } = get();
    if (!token || !activeProjectId) return;

    try {
      await projectsService.deleteModule(id);
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
    } catch (err) {
      console.error('Failed to delete module:', err);
    }
  },

  selectModule: (id) => {
    graphRequest?.abort();
    set({
      activeId: id,
      selectedNodeId: null,
      connectFrom: null,
    });
  },

  setView: (view) => {
    set((state) => ({ view, selectedWorkItemKey: view === 'kanban' ? state.selectedWorkItemKey : null }));
  },

  // Node Management Actions
  addNode: (type, label) => {
    const { activeId, modules, teamMembers } = get();
    if (!activeId) return;

    // Helper to find assignees dynamically from database users list
    const findAssignee = (r: 'pm' | 'uiux' | 'frontend' | 'backend', fallback: string) => {
      const match = teamMembers.find(m => m.role === r);
      return match ? match.name : fallback;
    };

    // Retrieve default assignee based on node type
    let defaultAssignee = findAssignee('uiux', 'Rian');
    if (type === 'system' || type === 'decision') {
      defaultAssignee = findAssignee('backend', 'Budi');
    } else if (type === 'process') {
      defaultAssignee = findAssignee('frontend', 'Siti');
    }

    const newNode: Node = {
      id: `node_${uid()}`,
      type,
      label,
      x: 150 + Math.random() * 100,
      y: 150 + Math.random() * 100,
      doc: {
        actor: type === 'actor' ? 'Petugas / User' : 'Sistem',
        trigger: '',
        input: '',
        process: '',
        output: '',
        rules: [],
        exceptionPath: '',
        system: type === 'system' ? 'Aplikasi Gateway' : 'FlowakPortal',
        sla: 'Instan',
        priority: 'medium',
        riskLevel: 'medium',
        acceptanceCriteria: '',
      },
      roles: {
        uiux: {
          assignee: findAssignee('uiux', 'Rian'),
          status: 'planned',
          screen: '',
          link: '',
        },
        frontend: {
          assignee: findAssignee('frontend', 'Siti'),
          status: 'planned',
          page: '',
          route: '',
          interaction: '',
          validation: '',
          state: '',
          handoffLink: '',
        },
        backend: {
          assignee: type === 'system'
            ? (teamMembers.filter(m => m.role === 'backend')[1]?.name || findAssignee('backend', 'Arief'))
            : findAssignee('backend', 'Budi'),
          status: 'planned',
          method: 'POST',
          endpoint: '',
          auth: '',
          request: '',
          response: '',
          statusCode: '200',
          errorCodes: '',
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
    debouncedSave(set, get);

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
    debouncedSave(set, get);
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
    debouncedSave(set, get);
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
    debouncedSave(set, get);
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
    debouncedSave(set, get);
  },

  deleteNode: (id) => {
    const { activeId, modules } = get();
    if (!activeId) return;

    const activeMod = modules.find((m) => m.id === activeId);
    if (!activeMod) return;

    const deletingNode = activeMod.nodes.find((n) => n.id === id);
    const deletingNodeName = deletingNode?.label || 'Langkah';
    
    // Delete logic helper
    const filteredNodes = activeMod.nodes.filter((n) => n.id !== id);
    const filteredEdges = activeMod.edges.filter((e) => e.from !== id && e.to !== id);
    
    const updatedMod = {
      ...activeMod,
      nodes: filteredNodes,
      edges: filteredEdges,
      deletedNodes: deletingNode?.rowVersion ? [...(activeMod.deletedNodes || []), { id, rowVersion: deletingNode.rowVersion }] : activeMod.deletedNodes,
      deletedEdges: [...(activeMod.deletedEdges || []), ...activeMod.edges
        .filter((edge) => edge.from === id || edge.to === id)
        .filter((edge) => edge.rowVersion)
        .map((edge) => ({ id: edge.id, rowVersion: edge.rowVersion! }))]
    };

    const updated = modules.map((m) => (m.id === activeId ? updatedMod : m));

    set({
      modules: updated,
      selectedNodeId: null,
      connectFrom: null,
    });
    debouncedSave(set, get);

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
    debouncedSave(set, get);

    get().addNotification(
      'Koneksi Berhasil',
      `Menghubungkan "${fromNodeLabel}" ke "${toNodeLabel}".`,
      'success'
    );
  },

  deleteEdge: (id) => {
    const { activeId, modules } = get();
    if (!activeId) return;

    const deletedEdge = modules.find((m) => m.id === activeId)?.edges.find((edge) => edge.id === id);
    const updated = modules.map((m) => {
      if (m.id === activeId) {
        return {
          ...m,
          edges: m.edges.filter((e) => e.id !== id),
          deletedEdges: deletedEdge?.rowVersion ? [...(m.deletedEdges || []), { id, rowVersion: deletedEdge.rowVersion }] : m.deletedEdges,
        };
      }
      return m;
    });

    set({ modules: updated });
    debouncedSave(set, get);

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
    debouncedSave(set, get);
  },

  selectNode: (id) => {
    set({ selectedNodeId: id });
  },
  selectWorkItem: (key) => {
    set({ selectedWorkItemKey: key });
  },

  setConnectFrom: (id) => {
    set({ connectFrom: id });
  },

  // Notification Management Actions
  addNotification: (title, message, type: NotificationItem['type'] = 'info') => {
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
    void notificationsService.markAllRead();
  },

  setSelectedNotif: (notif) => {
    set({ selectedNotif: notif });
  },

  clearNotifications: () => {
    set({ notifications: [] });
  },
  loadNotifications: async () => {
    try { set({ notifications: await notificationsService.list() }); }
    catch { /* Notification availability must not block the workspace. */ }
  },

  // AI & Extra modularity actions
  loadAiGeneratedFlow: async (name, description, nodes, edges) => {
    const { activeProjectId, token } = get();
    if (!activeProjectId || !token) return;

    const moduleId = await get().addModule(name, description);
    if (!moduleId) return;

    const { modules } = get();
    const updated = modules.map((m) => {
      if (m.id === moduleId) {
        return {
          ...m,
          nodes,
          edges
        };
      }
      return m;
    });

    set({ modules: updated, activeId: moduleId });
    debouncedSave(set, get);

    get().addNotification(
      'Alur AI Dimuat',
      `Alur kerja "${name}" dari asisten AI sukses dimuat.`,
      'success'
    );
  },

  importModule: async (mod) => {
    const { token, activeProjectId } = get();
    if (!token || !activeProjectId) return;

    if (!mod || !mod.name || !Array.isArray(mod.nodes)) {
      get().addNotification('Impor Gagal', 'Format file JSON modul tidak valid.', 'warning');
      return;
    }

    try {
      const data = await projectsService.createModule(activeProjectId, { name: mod.name, description: mod.description || '', nodes: mod.nodes, edges: mod.edges || [] });
      get().addNotification('Modul Diimpor', `Modul "${mod.name}" berhasil diimpor.`, 'success');
      const parsedModules = await projectsService.listModules(activeProjectId);
      set({ modules: parsedModules, activeId: data.module_id, selectedNodeId: null });
    } catch (err) {
      console.error('Failed to import module:', err);
    }
  },

  // Team Management Actions connected to Backend DB API
  loadTeamMembers: async () => {
    const { token } = get();
    if (!token) return;

    try {
      set({ teamMembers: await membersService.listTeam() });
    } catch (err) {
      console.error('Failed to load team members:', err);
    }
  },

  addTeamMember: async (name, email, role) => {
    const { token } = get();
    if (!token) return;

    try {
      const data = await membersService.createTeamMember(name, email, role);
        const tempPassword = data.temporary_password ? ` Password sementara: ${data.temporary_password}` : '';
        get().addNotification('Anggota Tim Ditambahkan', `${name} dimasukkan ke daftar kontributor.${tempPassword}`, 'success');
        await get().loadTeamMembers();
    } catch (err) {
      get().addNotification('Gagal Menambahkan Anggota', 'Koneksi ke server terputus.', 'warning');
    }
  },

  deleteTeamMember: async (id) => {
    const { token, currentUser } = get();
    if (!token) return;

    if (currentUser && currentUser.id === id) {
      get().addNotification('Tindakan Dicegah', 'Anda tidak dapat menghapus akun Anda sendiri.', 'warning');
      return;
    }

    try {
      await membersService.deleteTeamMember(id);
        get().addNotification('Anggota Tim Dihentikan', 'Kontributor telah dihapus.', 'warning');
        await get().loadTeamMembers();
    } catch (err) {
      get().addNotification('Gagal Mengeluarkan Anggota', 'Koneksi ke server terputus.', 'warning');
    }
  },

  loadProjectMembers: async () => {
    const { token, activeProjectId } = get();
    if (!token || !activeProjectId) return;

    try {
      set({ projectMembers: await membersService.listProjectMembers(activeProjectId) });
    } catch (err) {
      console.error('Failed to load project members:', err);
    }
  },

  addProjectMember: async (userId) => {
    const { token, activeProjectId } = get();
    if (!token || !activeProjectId) return false;

    try {
      await membersService.addProjectMember(activeProjectId, userId);
        get().addNotification('Anggota Ditambahkan', 'Anggota tim berhasil ditambahkan ke proyek.', 'success');
        await get().loadProjectMembers();
        return true;
    } catch (err) {
      get().addNotification('Gagal Menambahkan Anggota', 'Koneksi ke server terputus.', 'warning');
      return false;
    }
  },

  deleteProjectMember: async (userId) => {
    const { token, activeProjectId } = get();
    if (!token || !activeProjectId) return false;

    try {
      await membersService.deleteProjectMember(activeProjectId, userId);
        get().addNotification('Anggota Dihapus', 'Anggota tim telah dihapus dari proyek.', 'warning');
        await get().loadProjectMembers();
        return true;
    } catch (err) {
      get().addNotification('Gagal Menghapus Anggota', 'Koneksi ke server terputus.', 'warning');
      return false;
    }
  },

  loadDashboardStats: async () => {
    const { token } = get();
    if (!token) return;

    try {
      const data = await membersService.dashboard();
      set({ dashboardStats: { myTasksCount: data.my_tasks_count, completionRate: data.completion_rate } });
    } catch (err) {
      console.error('Failed to load dashboard stats:', err);
    }
  }
}));
