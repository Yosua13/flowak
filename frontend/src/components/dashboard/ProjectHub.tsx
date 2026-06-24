import React, { useState, useEffect } from 'react';
import { useStore } from '../../store/useStore';
import { Plus, LogOut, Folder, Trash2, LayoutGrid, Calendar, Layers, Activity, AlertCircle, Heart } from 'lucide-react';
import { motion, AnimatePresence } from 'motion/react';

export default function ProjectHub() {
  const {
    projects,
    currentUser,
    logoutUser,
    createProject,
    deleteProject,
    selectProject,
    loadProjects,
    dashboardStats
  } = useStore();

  const [isAdding, setIsAdding] = useState(false);
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [errorMsg, setErrorMsg] = useState('');

  useEffect(() => {
    loadProjects();
  }, [loadProjects]);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;
    setErrorMsg('');
    const success = await createProject(name.trim(), description.trim());
    if (success) {
      setName('');
      setDescription('');
      setIsAdding(false);
    } else {
      setErrorMsg('Gagal membuat proyek. Silakan coba lagi.');
    }
  };

  const handleDelete = (id: string, name: string, e: React.MouseEvent) => {
    e.stopPropagation();
    const confirmed = window.confirm(`Apakah Anda yakin ingin menghapus proyek "${name}"? Seluruh modul alur kerja di dalamnya akan dihapus secara permanen.`);
    if (confirmed) {
      deleteProject(id);
    }
  };

  return (
    <div className="min-h-screen w-screen bg-[#0A0A0B] text-gray-100 font-sans p-6 md:p-12 relative overflow-hidden select-none">
      
      {/* Decorative Gold Blurs */}
      <div className="absolute top-0 right-0 w-[500px] h-[500px] rounded-full bg-[#C5A267]/2 blur-[150px] pointer-events-none"></div>
      <div className="absolute bottom-0 left-0 w-[450px] h-[450px] rounded-full bg-[#C5A267]/2 blur-[150px] pointer-events-none"></div>

      <div className="max-w-6xl mx-auto space-y-10 relative z-10">
        
        {/* Header Zone */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-6 pb-6 border-b border-white/5">
          <div className="text-left">
            <div className="flex items-center space-x-2 mb-1.5">
              <div className="w-2.5 h-2.5 rounded-full bg-[#C5A267] shadow-[0_0_8px_#C5A267] animate-pulse"></div>
              <span className="text-[10px] text-gray-500 font-mono tracking-wider uppercase font-bold">Workspace Hub</span>
            </div>
            <h1 className="text-3xl font-light text-white tracking-tight" style={{ fontFamily: 'Georgia, serif' }}>
              Selamat datang kembali, <span className="italic">{currentUser?.name || 'Kontributor'}</span>
            </h1>
            <p className="text-xs text-gray-400 mt-1">Peran: <span className="uppercase text-[#C5A267] font-bold font-mono">{currentUser?.role || '-'}</span></p>
          </div>

          <div className="flex items-center space-x-3">
            {currentUser?.role === 'pm' && (
              <button
                onClick={() => setIsAdding(true)}
                className="flex items-center space-x-1.5 px-4 py-2.5 bg-[#C5A267] hover:bg-[#B38F52] text-black text-xs font-bold uppercase tracking-wider rounded-xl shadow cursor-pointer transition"
              >
                <Plus className="w-4 h-4" />
                <span>Project Baru</span>
              </button>
            )}
            <button
              onClick={logoutUser}
              className="flex items-center space-x-1.5 px-4 py-2.5 bg-white/5 hover:bg-red-950/20 border border-white/10 hover:border-red-900/40 text-gray-300 hover:text-red-400 text-xs font-bold uppercase tracking-wider rounded-xl cursor-pointer transition"
              title="Logout"
            >
              <LogOut className="w-4 h-4" />
              <span>Keluar</span>
            </button>
          </div>
        </div>

        {/* Dashboard Overview Metrics */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <div className="p-4 bg-[#111113] border border-white/5 rounded-2xl text-left">
            <span className="text-[9px] uppercase tracking-wider text-gray-500 font-mono font-bold block mb-1">Total Proyek</span>
            <div className="flex items-center justify-between">
              <span className="text-2xl font-bold text-white">{projects.length}</span>
              <Folder className="w-5 h-5 text-[#C5A267]" />
            </div>
          </div>
          <div className="p-4 bg-[#111113] border border-white/5 rounded-2xl text-left">
            <span className="text-[9px] uppercase tracking-wider text-gray-500 font-mono font-bold block mb-1">Tingkat Penyelesaian</span>
            <div className="flex items-center justify-between">
              <span className="text-2xl font-bold text-emerald-400">
                {dashboardStats ? `${dashboardStats.completionRate}%` : '0%'}
              </span>
              <Activity className="w-5 h-5 text-emerald-500" />
            </div>
          </div>
          <div className="p-4 bg-[#111113] border border-white/5 rounded-2xl text-left">
            <span className="text-[9px] uppercase tracking-wider text-gray-500 font-mono font-bold block mb-1">Peran Aktif</span>
            <div className="flex items-center justify-between">
              <span className="text-xs uppercase font-mono font-bold text-gray-300">
                {currentUser?.role === 'pm' ? 'Project Manager' : currentUser?.role?.toUpperCase()}
              </span>
              <Layers className="w-5 h-5 text-indigo-400" />
            </div>
          </div>
          <div className="p-4 bg-[#111113] border border-white/5 rounded-2xl text-left">
            <span className="text-[9px] uppercase tracking-wider text-gray-500 font-mono font-bold block mb-1">Tugas Saya</span>
            <div className="flex items-center justify-between">
              <span className="text-2xl font-bold text-white">
                {dashboardStats ? dashboardStats.myTasksCount : 0} <span className="text-[10px] text-gray-500 font-normal">tugas aktif</span>
              </span>
              <Heart className="w-5 h-5 text-rose-400" />
            </div>
          </div>
        </div>

        {/* Modal/Form Overlay for New Project */}
        <AnimatePresence>
          {isAdding && (
            <div className="fixed inset-0 bg-[#0A0A0B]/80 backdrop-blur-sm flex items-center justify-center z-50 p-4">
              <motion.form
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
                exit={{ opacity: 0, scale: 0.95 }}
                onSubmit={handleCreate}
                className="w-full max-w-md bg-[#111113] border border-[#C5A267]/20 rounded-2xl p-6 shadow-2xl space-y-4"
              >
                <div className="text-left">
                  <h3 className="text-lg font-bold text-white mb-1">Buat Proyek Baru</h3>
                  <p className="text-[10px] text-gray-400">Hub spesifikasi dan traceability alur kerja tim Anda.</p>
                </div>

                {errorMsg && (
                  <div className="p-2.5 bg-red-950/20 border border-red-900/40 text-red-400 text-xs rounded-xl text-center">
                    {errorMsg}
                  </div>
                )}

                <div className="space-y-1 text-left">
                  <label className="text-[9px] font-bold uppercase tracking-wider text-gray-400 font-mono">Nama Proyek</label>
                  <input
                    type="text"
                    required
                    placeholder="Contoh: Aplikasi Sistem Pengajuan Cuti"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    className="w-full text-xs bg-[#1A1A1D] border border-white/10 rounded-xl p-3 text-white focus:border-[#C5A267]/60 focus:ring-1 focus:ring-[#C5A267]/60 focus:outline-none"
                  />
                </div>

                <div className="space-y-1 text-left">
                  <label className="text-[9px] font-bold uppercase tracking-wider text-gray-400 font-mono">Deskripsi Proyek</label>
                  <textarea
                    placeholder="Tulis ringkasan mengenai tujuan proyek..."
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    className="w-full text-xs bg-[#1A1A1D] border border-white/10 rounded-xl p-3 text-white focus:border-[#C5A267]/60 focus:ring-1 focus:ring-[#C5A267]/60 focus:outline-none h-20 resize-none"
                  />
                </div>

                <div className="flex justify-end space-x-2 pt-2">
                  <button
                    type="button"
                    onClick={() => setIsAdding(false)}
                    className="px-4 py-2 bg-transparent text-gray-400 hover:text-white text-xs uppercase font-mono tracking-wider cursor-pointer"
                  >
                    Batal
                  </button>
                  <button
                    type="submit"
                    className="px-4 py-2 bg-[#C5A267] text-black font-bold text-xs uppercase tracking-wider rounded-xl cursor-pointer hover:bg-[#B38F52]"
                  >
                    Buat Proyek
                  </button>
                </div>
              </motion.form>
            </div>
          )}
        </AnimatePresence>

        {/* Project List / Grid */}
        <div className="space-y-6">
          <div className="flex items-center justify-between text-left">
            <h2 className="text-lg font-bold text-white flex items-center gap-2">
              <LayoutGrid className="w-4 h-4 text-[#C5A267]" />
              Daftar Proyek Anda
            </h2>
            <span className="text-[10px] text-gray-400 font-mono">Menampilkan {projects.length} proyek</span>
          </div>

          {projects.length === 0 ? (
            <div className="p-16 border border-dashed border-white/10 rounded-2xl text-center bg-[#111113]/30">
              <Folder className="w-10 h-10 text-gray-600 mx-auto mb-3" />
              <h3 className="text-sm font-semibold text-gray-400">Belum ada proyek dibuat</h3>
              <p className="text-xs text-gray-500 max-w-xs mx-auto mt-1">Mulai rancang spesifikasi alur kerja dengan membuat proyek baru di tombol kanan atas.</p>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {projects.map((project) => (
                <motion.div
                  key={project.id}
                  onClick={() => selectProject(project.id)}
                  whileHover={{ y: -3, scale: 1.01 }}
                  className="bg-[#111113] border border-white/5 rounded-2xl p-6 text-left cursor-pointer hover:border-[#C5A267]/30 hover:shadow-[0_10px_25px_rgba(0,0,0,0.4)] transition-all flex flex-col justify-between group relative overflow-hidden h-72"
                >
                  {/* Decorative faint grid lines inside card */}
                  <div className="absolute inset-0 bg-[linear-gradient(to_bottom,rgba(255,255,255,0.01)_1px,transparent_1px),linear-gradient(to_right,rgba(255,255,255,0.01)_1px,transparent_1px)] bg-[size:20px_20px] pointer-events-none"></div>

                  <div>
                    <div className="flex items-start justify-between">
                      <div className="w-8 h-8 rounded-lg bg-[#C5A267]/10 flex items-center justify-center text-[#C5A267]">
                        <Folder className="w-4.5 h-4.5" />
                      </div>
                      {currentUser?.id === project.owner_id && (
                        <button
                          onClick={(e) => handleDelete(project.id, project.name, e)}
                          className="opacity-0 group-hover:opacity-100 p-1.5 rounded hover:bg-white/5 text-gray-500 hover:text-red-400 transition cursor-pointer"
                          title="Hapus Proyek"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      )}
                    </div>

                    <h3 className="text-md font-semibold text-white mt-4 group-hover:text-[#C5A267] transition-colors line-clamp-1">
                      {project.name}
                    </h3>
                    <p className="text-xs text-gray-400 mt-2 line-clamp-3 leading-relaxed">
                      {project.description || 'Tidak ada deskripsi proyek.'}
                    </p>
                  </div>

                  <div className="border-t border-white/5 pt-4 mt-4 flex items-center justify-between text-[10px] text-gray-500 font-mono">
                    <span className="flex items-center gap-1">
                      <Calendar className="w-3.5 h-3.5" />
                      {new Date(project.created_at).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })}
                    </span>
                    <span className="px-2 py-0.5 rounded-full bg-[#C5A267]/10 text-[#C5A267] border border-[#C5A267]/25 text-[9px] uppercase font-bold tracking-wider">
                      Masuk Kerja
                    </span>
                  </div>
                </motion.div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
