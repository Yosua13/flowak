import React, { useState } from 'react';
import { useStore } from '../../store/useStore';
import { Mail, Lock, Loader2, ArrowRight } from 'lucide-react';
import { motion } from 'motion/react';

export default function Login() {
  const { loginUser, setScreen } = useStore();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email || !password) return;
    setIsLoading(true);
    setErrorMsg('');
    
    const success = await loginUser(email, password);
    setIsLoading(false);
    if (!success) {
      setErrorMsg('Email atau password Anda salah. Silakan coba kembali.');
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
        <div className="text-center mb-8">
          <div className="w-3.5 h-3.5 rounded-full bg-[#C5A267] shadow-[0_0_12px_#C5A267] mx-auto mb-4 animate-pulse"></div>
          <h1 className="text-3xl font-light tracking-tighter text-white" style={{ fontFamily: 'Georgia, serif' }}>
            Masuk ke <span className="italic">Flowak</span>
          </h1>
          <p className="text-[9px] text-[#C5A267] font-mono tracking-widest uppercase mt-1.5 select-all hover:underline" title="Salin URL halaman">
            {window.location.origin}/login/
          </p>
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
                placeholder="••••••••"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full text-sm pl-11 pr-4 py-3 bg-[#1A1A1D] border border-white/10 rounded-xl focus:border-[#C5A267]/60 focus:ring-1 focus:ring-[#C5A267]/60 focus:outline-none text-white transition-all placeholder-gray-600"
              />
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
                <span>Memproses...</span>
              </>
            ) : (
              <>
                <span>Masuk Sistem</span>
                <ArrowRight className="w-4 h-4" />
              </>
            )}
          </button>
        </form>

        {/* Footer Redirect */}
        <div className="mt-8 text-center border-t border-white/5 pt-5">
          <p className="text-xs text-gray-500">
            Belum memiliki akun kontributor?{' '}
            <button
              onClick={() => setScreen('register')}
              className="text-[#C5A267] hover:underline font-semibold bg-transparent border-none outline-none cursor-pointer"
            >
              Daftar Sekarang
            </button>
          </p>
        </div>
      </motion.div>
    </div>
  );
}
