/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from 'react';
import { Node, Status, HttpMethod } from '../../domain/types';
import { useStore } from '../../store/useStore';
import { generateCurl } from '../../services/curl';
import { Copy, Terminal, Play, Check, AlertCircle, RefreshCw, Sparkles, Code, Loader2 } from 'lucide-react';

interface BackendTabProps {
  node: Node;
}

export default function BackendTab({ node }: BackendTabProps) {
  const { updateRole, teamMembers, addNotification } = useStore();
  const be = node.roles?.backend || {
    assignee: '',
    status: 'planned',
    method: 'GET',
    endpoint: '',
    auth: '',
    request: '',
    response: '',
    statusCode: '200'
  };

  const [copied, setCopied] = useState(false);
  const [testing, setTesting] = useState(false);
  const [testResult, setTestResult] = useState<{ code: string; body: string } | null>(null);

  // AI Mock & Boilerplate states
  const [aiGeneratingMock, setAiGeneratingMock] = useState(false);
  const [aiGeneratingCode, setAiGeneratingCode] = useState(false);
  const [selectedLanguage, setSelectedLanguage] = useState('Express TS');
  const [generatedCode, setGeneratedCode] = useState('');
  const [copiedCode, setCopiedCode] = useState(false);

  const handleFieldChange = (key: string, val: string) => {
    updateRole(node.id, 'backend', { [key]: val });
    setTestResult(null);
  };

  const handleCopyCurl = () => {
    const command = generateCurl(be);
    navigator.clipboard.writeText(command);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleTestEndpoint = () => {
    setTesting(true);
    setTestResult(null);
    setTimeout(() => {
      setTesting(false);
      setTestResult({
        code: be.statusCode || '200',
        body: be.response || '{\n  "message": "Success",\n  "simulated": true\n}',
      });
    }, 850);
  };

  // AI Generates Mock Request/Response Contracts
  const handleGenerateMockData = async () => {
    setAiGeneratingMock(true);
    try {
      const res = await fetch('/api/ai/mock-payload', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          actor: node.doc?.actor || 'Sistem',
          label: node.label,
          endpoint: be.endpoint || '/api/resource',
          process: node.doc?.process || '',
          method: be.method || 'GET',
        }),
      });

      const data = await res.json();
      if (data.error) throw new Error(data.error);

      // Save generated mockup database formats
      updateRole(node.id, 'backend', {
        request: be.method !== 'GET' ? data.request : '',
        response: data.response || '{\n  "status": "success"\n}',
      });

      addNotification('Kontrak AI Sukses', `Payload request & response langkah "${node.label}" terisi otomatis.`, 'success');
    } catch (err: any) {
      console.error(err);
      addNotification('Gagal mengambil AI payload', 'Menggunakan template data default lokal.', 'warning');
    } finally {
      setAiGeneratingMock(false);
    }
  };

  // AI Generates Backend Code files
  const handleGenerateBoilerplate = async () => {
    setAiGeneratingCode(true);
    setGeneratedCode('');
    try {
      const res = await fetch('/api/ai/generate-code', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          type: 'backend',
          nodeLabel: node.label,
          nodeType: node.type,
          detail: be,
          framework: selectedLanguage,
        }),
      });

      const data = await res.json();
      if (data.error) throw new Error(data.error);

      setGeneratedCode(data.code);
      addNotification('Kode Terbuat', `Boilerplate backend controller ${selectedLanguage} berhasil disusun.`, 'success');
    } catch (err: any) {
      console.error(err);
      setGeneratedCode(`// Gagal menghasilkan kode:\n// ${err.message}\n\ntry {\n  // Mohon hubungkan GEMINI_API_KEY\n}`);
    } finally {
      setAiGeneratingCode(false);
    }
  };

  const beMembers = teamMembers.filter((m) => m.role === 'backend' || m.role === 'pm');
  const methods: HttpMethod[] = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE'];

  const getBadgeColors = (m: string) => {
    switch (m) {
      case 'GET': return 'bg-emerald-500/15 text-emerald-400 border-emerald-400/40';
      case 'POST': return 'bg-blue-500/15 text-blue-400 border-blue-400/40';
      case 'PUT': return 'bg-amber-500/15 text-amber-400 border-amber-400/40';
      case 'DELETE': return 'bg-red-500/15 text-red-400 border-red-400/40';
      default: return 'bg-slate-500/15 text-slate-400 border-slate-400/40';
    }
  };

  const handleCopyCode = () => {
    navigator.clipboard.writeText(generatedCode);
    setCopiedCode(true);
    setTimeout(() => setCopiedCode(false), 2000);
  };

  return (
    <div className="space-y-4 text-left select-text pb-6">
      <div className="flex justify-between items-start">
        <div>
          <span className="text-[10px] font-bold font-mono tracking-widest text-[#C5A267] uppercase block mb-1">
            Kontrak API Backend & Explorer
          </span>
          <p className="text-[11px] text-gray-400 font-sans">
            Arsitektur kontrak endpoint server untuk pengerjaan modul. Menghasilkan perintah cURL dan simulator uji.
          </p>
        </div>

        {/* AI Auto fill Mock button */}
        <button
          onClick={handleGenerateMockData}
          disabled={aiGeneratingMock || !be.endpoint}
          className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-[#C5A267]/10 hover:bg-[#C5A267]/20 border border-[#C5A267]/35 text-[#C5A267] text-[10px] font-bold transition-all disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
          title="Isi otomatis request dan response menggunakan AI"
        >
          {aiGeneratingMock ? (
            <Loader2 className="w-3 h-3 animate-spin text-[#C5A267]" />
          ) : (
            <Sparkles className="w-3 h-3 text-[#C5A267] animate-pulse" />
          )}
          <span>MOCK VIA AI</span>
        </button>
      </div>

      <div className="space-y-3.5">
        {/* Assignee selection */}
        <div className="space-y-1">
          <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">
            Penanggung Jawab (BE Engineer)
          </label>
          <select
            value={be.assignee || ''}
            onChange={(e) => handleFieldChange('assignee', e.target.value)}
            className="w-full text-xs border border-white/5 rounded-xl px-3 py-2 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267] transition font-medium"
          >
            <option value="" className="bg-[#131315]">-- Pilih Anggota Tim --</option>
            {beMembers.map((m) => (
              <option key={m.id} value={m.name} className="bg-[#131315]">
                {m.name} ({m.email})
              </option>
            ))}
          </select>
        </div>

        {/* Status Selection */}
        <div className="space-y-1">
          <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">
            Status Kesiapan Kontrak/API
          </label>
          <select
            value={be.status || 'planned'}
            onChange={(e) => handleFieldChange('status', e.target.value as Status)}
            className="w-full text-xs border border-white/5 rounded-xl px-3 py-2 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267] transition font-medium"
          >
            <option value="planned" className="bg-[#131315]">Planned (Terencana)</option>
            <option value="in_progress" className="bg-[#131315]">In Progress (Pengerjaan)</option>
            <option value="review" className="bg-[#131315]">In Review (Peninjauan)</option>
            <option value="done" className="bg-[#131315]">Done (Selesai)</option>
          </select>
        </div>

        {/* API Endpoint Input row with Method selector */}
        <div className="space-y-1">
          <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">
            Definisi Endpoint
          </label>
          <div className="flex space-x-1.5 font-mono">
            <select
              value={be.method || 'GET'}
              onChange={(e) => handleFieldChange('method', e.target.value as HttpMethod)}
              className={`text-xs font-mono font-bold border border-white/5 rounded-xl px-2.5 bg-[#1A1A1D] outline-none transition cursor-pointer ${getBadgeColors(be.method || 'GET')}`}
            >
              {methods.map((m) => (
                <option key={m} value={m} className="bg-[#131315] text-white">{m}</option>
              ))}
            </select>
            
            <input
              type="text"
              value={be.endpoint || ''}
              onChange={(e) => handleFieldChange('endpoint', e.target.value)}
              placeholder="/api/v1/leave/request"
              className="flex-1 text-xs font-mono border border-white/5 rounded-xl px-3 py-2 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267] transition font-medium"
            />
          </div>
        </div>

        {/* Auth token & Expected status code */}
        <div className="grid grid-cols-2 gap-2">
          <div className="space-y-1">
            <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">
              Otorisasi token
            </label>
            <input
              type="text"
              value={be.auth || ''}
              onChange={(e) => handleFieldChange('auth', e.target.value)}
              placeholder="Bearer Token..."
              className="w-full text-xs font-mono border border-white/5 rounded-xl px-3 py-2 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267] transition"
            />
          </div>
          <div className="space-y-1">
            <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">
              Kode Status Respon
            </label>
            <input
              type="text"
              value={be.statusCode || '200'}
              onChange={(e) => handleFieldChange('statusCode', e.target.value)}
              placeholder="200, 201..."
              className="w-full text-xs font-mono border border-white/5 rounded-xl px-3 py-2 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267] transition"
            />
          </div>
        </div>

        {/* Request Response Area */}
        <div className="grid grid-cols-1 gap-2.5">
          {be.method !== 'GET' && (
            <div className="space-y-1">
              <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">
                Unggahan Body Permintaan (JSON Request Body)
              </label>
              <textarea
                value={be.request || ''}
                onChange={(e) => handleFieldChange('request', e.target.value)}
                placeholder={'{\n  "payload": "key"\n}'}
                rows={3}
                className="w-full text-[10px] font-mono border border-[#C5A267]/10 bg-[#0A0A0B] text-gray-300 rounded-xl p-2.5 resize-y focus:outline-none focus:ring-1 focus:ring-[#C5A267] font-medium leading-normal h-20"
              />
            </div>
          )}

          <div className="space-y-1">
            <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">
              Kerangka Respon Hasil (JSON Response Body)
            </label>
            <textarea
              value={be.response || ''}
              onChange={(e) => handleFieldChange('response', e.target.value)}
              placeholder={'{\n  "success": true\n}'}
              rows={3}
              className="w-full text-[10px] font-mono border border-[#C5A267]/10 bg-[#0A0A0B] text-gray-300 rounded-xl p-2.5 resize-y focus:outline-none focus:ring-1 focus:ring-[#C5A267] font-medium leading-normal h-20"
            />
          </div>
        </div>

        {/* cURL Generation view */}
        <div className="pt-2">
          <div className="flex items-center justify-between mb-1.5">
            <label className="block text-[9px] font-bold text-gray-400 font-mono uppercase tracking-wider">
              Perintah cURL Klien Tergenerasi
            </label>
            <button
              onClick={handleCopyCurl}
              className="flex items-center space-x-1.5 text-[10px] font-bold font-mono tracking-wide text-[#C5A267] hover:text-[#B08E56] transition cursor-pointer"
            >
              {copied ? (
                <>
                  <Check className="w-3 h-3 text-emerald-500" />
                  <span className="text-emerald-500 uppercase">Tersalin!</span>
                </>
              ) : (
                <>
                  <Copy className="w-3 h-3" />
                  <span>SALIN CURL</span>
                </>
              )}
            </button>
          </div>
          <div className="bg-[#0A0A0B] border border-white/5 p-3 rounded-xl max-h-24 overflow-y-auto whitespace-pre-wrap word-break font-mono text-[9px] text-[#C5A267] select-all leading-normal">
            {generateCurl(be)}
          </div>
        </div>

        {/* Endpoint Tester (Simulasi Interaktif) */}
        <div className="pt-2 border-t border-white/5">
          <button
            onClick={handleTestEndpoint}
            disabled={testing || !be.endpoint}
            className="w-full flex items-center justify-center space-x-2 py-2.5 rounded-xl bg-[#C5A267] hover:bg-[#B08E56] disabled:bg-white/5 disabled:text-gray-500 text-black text-xs font-bold font-mono uppercase cursor-pointer disabled:cursor-not-allowed tracking-wider shadow-sm transition"
          >
            {testing ? (
              <>
                <RefreshCw className="w-3.5 h-3.5 animate-spin" />
                <span>MENGIRIM SIMULASI...</span>
              </>
            ) : (
              <>
                <Play className="w-3.5 h-3.5 text-black" />
                <span>UJI ENDPOINT (SIMULASI)</span>
              </>
            )}
          </button>

          {testResult && (
            <div className="mt-3 bg-[#0A0A0B] border border-white/10 rounded-xl p-3 text-[10px] font-mono leading-relaxed select-text relative">
              <span className="absolute right-2.5 top-2 bg-[#C5A267] text-black font-mono text-[8px] font-bold px-1.5 py-0.5 rounded shadow">
                SIMULASI UJI
              </span>
              <div className="flex items-center space-x-1.5 text-gray-400 border-b border-white/5 pb-1.5 mb-2">
                <AlertCircle className="w-3.5 h-3.5 text-emerald-400" />
                <span>HTTP/{be.method || 'GET'} Status:</span>
                <span className="font-bold text-emerald-400 pr-2">{testResult.code} OK</span>
              </div>
              <div className="text-gray-300 max-h-36 overflow-y-auto leading-normal whitespace-pre-wrap font-mono">
                {testResult.body}
              </div>
            </div>
          )}
        </div>

        {/* AI Boilerplate Code Generator Section C */}
        <div className="pt-4 mt-2 border-t border-white/5">
          <span className="text-[10px] font-bold font-mono tracking-widest text-[#C5A267] uppercase block mb-2">
            AI Code Boilerplate Generator
          </span>
          <div className="flex space-x-2 mb-3">
            <select
              value={selectedLanguage}
              onChange={(e) => setSelectedLanguage(e.target.value)}
              className="flex-1 text-xs border border-white/5 rounded-xl px-3 py-2 bg-[#1A1A1D] text-white outline-none focus:ring-1 focus:ring-[#C5A267]"
            >
              <option value="Express TS">Node.js (Express TS)</option>
              <option value="NestJS Controller">NestJS (TypeScript)</option>
              <option value="Go Fiber">GoLang (Fiber Router)</option>
              <option value="FastAPI Python">FastAPI (Python)</option>
            </select>
            <button
              onClick={handleGenerateBoilerplate}
              disabled={aiGeneratingCode || !be.endpoint}
              className="flex items-center justify-center space-x-1.5 px-4 bg-white/5 border border-white/10 hover:bg-white/15 rounded-xl text-xs font-semibold text-white cursor-pointer transition disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {aiGeneratingCode ? (
                <Loader2 className="w-3.5 h-3.5 animate-spin whitespace-nowrap" />
              ) : (
                <Code className="w-3.5 h-3.5" />
              )}
              <span className="whitespace-nowrap">BUAT KODE</span>
            </button>
          </div>

          {generatedCode && (
            <div className="mt-3 bg-[#0A0A0B] border border-[#C5A267]/20 rounded-xl p-3 text-[10px] font-mono leading-relaxed relative">
              <div className="flex items-center justify-between border-b border-white/5 pb-2 mb-2">
                <span className="text-white font-mono text-[9px] uppercase font-bold text-gray-400">{selectedLanguage} Controller</span>
                <button
                  onClick={handleCopyCode}
                  className="text-[#C5A267] font-mono text-[9px] hover:underline"
                >
                  {copiedCode ? 'TERSALIN' : 'SALIN'}
                </button>
              </div>
              <pre className="text-emerald-400 overflow-x-auto whitespace-pre leading-relaxed max-h-56 leading-normal p-1 scrollbar-thin">
                {generatedCode}
              </pre>
            </div>
          )}
        </div>

      </div>
    </div>
  );
}
