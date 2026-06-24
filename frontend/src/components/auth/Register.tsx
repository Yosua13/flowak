import React, { useState } from 'react';
import { useStore } from '../../store/useStore';
import { User, Mail, Lock, Briefcase, Loader2, ArrowRight } from 'lucide-react';
import { motion } from 'motion/react';

export default function Register() {
  const { registerUser, setScreen } = useStore();
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [role, setRole] = useState('pm');
  const [isLoading, setIsLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name || !email || !password || !role) return;
    setIsLoading(true);
    setErrorMsg('');

    const success = await registerUser(name, email, password, role);
    setIsLoading(false);
    if (!success) {
      setErrorMsg('Gagal melakukan registrasi. Mohon periksa email atau password Anda.');
    }
  };

  return (
    <div className="min-h-screen w-screen flex items-center justify-center bg-[#0A0A0B] relative overflow-hidden font-sans text-gray-100">
      
      {/* Decorative Gold Blurs */}
      <div className="absolute top-1/4 left-1/4 -translate-x-1/2 -translate-y-1/2 w-96 h-96 rounded-full bg-[#C5A267]/5 blur-[120px] pointer-events-none"></div>
      <div className="absolute bottom-1/4 right-1/4 translate-x-1/2 translate-y-1/2 w-[450px] h-[450px] rounded-full bg-[#C5A267]/3 blur-[150px] pointer-events-none"></div>

      <motion.div 
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, ease: 'easeOut' }}
        className="w-full max-w-md p-8 bg-[#111113] border border-white/5 rounded-2xl shadow-2xl relative z-10 mx-4"
      >
        {/* Brand Header */}
        <div className="text-center mb-6">
          <div className="w-3.5 h-3.5 rounded-full bg-[#C5A267] shadow-[0_0_12px_#C5A267] mx-auto mb-4 animate-pulse"></div>
          <h1 className="text-3xl font-light tracking-tighter text-white" style={{ fontFamily: 'Georgia, serif' }}>
            Daftar <span className="italic">Kontributor</span>
          </h1>
        </div>

        {errorMsg && (
          <motion.div 
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            className="mb-4 p-3 bg-red-950/20 border border-red-900/40 text-red-400 text-xs rounded-xl text-center"
          >
            {errorMsg}
          </motion.div>
        )}

        {/* Formulir */}
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-1.5 text-left">
            <label className="text-[10px] font-bold uppercase tracking-wider text-gray-400 font-mono">Nama Lengkap</label>
            <div className="relative">
              <User className="w-4 h-4 text-gray-500 absolute left-3.5 top-1/2 -translate-y-1/2" />
              <input
                type="text"
                required
                placeholder="Yosia Pratama"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="w-full text-sm pl-11 pr-4 py-3 bg-[#1A1A1D] border border-white/10 rounded-xl focus:border-[#C5A267]/60 focus:ring-1 focus:ring-[#C5A267]/60 focus:outline-none text-white transition-all placeholder-gray-600"
              />
            </div>
          </div>

          <div className="space-y-1.5 text-left">
            <label className="text-[10px] font-bold uppercase tracking-wider text-gray-400 font-mono">Email Instansi</label>
            <div className="relative">
              <Mail className="w-4 h-4 text-gray-500 absolute left-3.5 top-1/2 -translate-y-1/2" />
              <input
                type="email"
                required
                placeholder="developer@flowak.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full text-sm pl-11 pr-4 py-3 bg-[#1A1A1D] border border-white/10 rounded-xl focus:border-[#C5A267]/60 focus:ring-1 focus:ring-[#C5A267]/60 focus:outline-none text-white transition-all placeholder-gray-600"
              />
            </div>
          </div>

          <div className="space-y-1.5 text-left">
            <label className="text-[10px] font-bold uppercase tracking-wider text-gray-400 font-mono">Kata Sandi</label>
            <div className="relative">
              <Lock className="w-4 h-4 text-gray-500 absolute left-3.5 top-1/2 -translate-y-1/2" />
              <input
                type="password"
                required
                placeholder="Minimal 6 karakter"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full text-sm pl-11 pr-4 py-3 bg-[#1A1A1D] border border-white/10 rounded-xl focus:border-[#C5A267]/60 focus:ring-1 focus:ring-[#C5A267]/60 focus:outline-none text-white transition-all placeholder-gray-600"
              />
            </div>
          </div>

          <div className="space-y-1.5 text-left">
            <label className="text-[10px] font-bold uppercase tracking-wider text-gray-400 font-mono">Spesialisasi Peran</label>
            <div className="relative">
              <Briefcase className="w-4 h-4 text-gray-500 absolute left-3.5 top-1/2 -translate-y-1/2" />
              <select
                value={role}
                onChange={(e) => setRole(e.target.value)}
                className="w-full text-sm pl-11 pr-4 py-3 bg-[#1A1A1D] border border-white/10 rounded-xl focus:border-[#C5A267]/60 focus:ring-1 focus:ring-[#C5A267]/60 focus:outline-none text-white transition-all appearance-none cursor-pointer"
              >
                <option value="pm">Project Manager (PM)</option>
                <option value="uiux">UI/UX Designer</option>
                <option value="frontend">Frontend Engineer</option>
                <option value="backend">Backend Engineer</option>
              </select>
            </div>
          </div>

          <button
            type="submit"
            disabled={isLoading}
            className="w-full mt-2 py-3.5 px-4 bg-[#C5A267] hover:bg-[#B38F52] disabled:bg-gray-600 text-black font-bold rounded-xl text-xs uppercase tracking-widest cursor-pointer shadow-lg transition-all duration-200 flex items-center justify-center space-x-2"
          >
            {isLoading ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                <span>Mendaftarkan...</span>
              </>
            ) : (
              <>
                <span>Daftar Kontributor</span>
                <ArrowRight className="w-4 h-4" />
              </>
            )}
          </button>
        </form>

        {/* Footer Redirect */}
        <div className="mt-6 text-center border-t border-white/5 pt-4">
          <p className="text-xs text-gray-500">
            Sudah memiliki akun?{' '}
            <button
              onClick={() => setScreen('login')}
              className="text-[#C5A267] hover:underline font-semibold bg-transparent border-none outline-none cursor-pointer"
            >
              Masuk Sekarang
            </button>
          </p>
        </div>
      </motion.div>
    </div>
  );
}
