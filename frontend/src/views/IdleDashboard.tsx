import React, { useState } from 'react';

interface Props {
  onStartSynthesis: (topic: string, provider: string, format: string) => void;
  dailyLimitReached: boolean;
  docsGeneratedToday: number;
  systemStatus: Record<string, string>;
}

const providers = [
  { id: 'gemini', label: 'Gemini 2.5', vendor: 'Google', icon: 'deployed_code' },
  { id: 'kimi', label: 'Kimi K2.5', vendor: 'Currently Unavailable', icon: 'rocket_launch', unavailable: true },
];

export default function IdleDashboard({ onStartSynthesis, dailyLimitReached, docsGeneratedToday, systemStatus }: Props) {
  const [topic, setTopic] = useState('');
  const [provider, setProvider] = useState('gemini');
  const [format, setFormat] = useState('pdf');

  const handleSubmit = (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    if (topic.trim()) onStartSynthesis(topic, provider, format);
  };

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="min-h-full flex flex-col items-center px-6 py-16 md:py-24 w-full max-w-2xl mx-auto gap-10">

        {/* Hero */}
        <div className="text-center flex flex-col items-center gap-4 w-full">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-indigo-500/10 border border-indigo-500/20 mb-1">
            <div className="w-1.5 h-1.5 rounded-full bg-indigo-400 animate-pulse" />
            <span className="text-[11px] font-semibold text-indigo-400 tracking-widest uppercase">AI Research Pipeline</span>
          </div>
          <h1 className="text-4xl md:text-5xl font-semibold tracking-tight text-white leading-[1.1]">
            Deep research,<br />
            <span className="text-indigo-400">synthesized instantly</span>
          </h1>
          <div className="flex items-center gap-3 mt-2">
            <span className="text-[10px] text-slate-500 uppercase tracking-widest font-mono">
              Daily Quota: <span className={docsGeneratedToday >= 2 ? 'text-red-400' : 'text-emerald-400'}>{docsGeneratedToday}</span> / 2
            </span>
          </div>
          <p className="text-slate-400 text-sm max-w-md leading-relaxed">
            Enter a topic and Kyle will research, analyze, and produce a structured document using frontier AI models.
          </p>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className={`w-full rounded-xl bg-indigo-500/[0.04] border border-indigo-500/15 p-6 md:p-8 flex flex-col gap-6 relative overflow-hidden transition-all duration-500 ${dailyLimitReached ? 'opacity-50' : ''}`}>
          <div className="absolute -top-24 -right-24 w-48 h-48 bg-indigo-500/10 blur-[100px] rounded-full"></div>
          
          {dailyLimitReached && (
            <div className="absolute inset-0 z-20 flex flex-col items-center justify-center bg-black/60 backdrop-blur-[2px] rounded-2xl animate-in fade-in duration-700">
              <span className="material-symbols-outlined text-4xl text-indigo-400 mb-2 animate-pulse">snooze</span>
              <h3 className="text-white font-bold text-lg">Daily Quota Reached</h3>
              <p className="text-slate-400 text-sm max-w-xs text-center px-4">Kyle is recalibrating after synthesizing your documents. Discovery limits reset in 24 hours.</p>
            </div>
          )}

          <div className="flex flex-col gap-2">
            <label className="text-[11px] font-semibold text-slate-500 uppercase tracking-widest">Research Topic</label>
            <input
              type="text"
              value={topic}
              onChange={e => setTopic(e.target.value)}
              disabled={dailyLimitReached}
              className="w-full bg-black/30 border border-white/[0.07] rounded-lg px-4 py-3 text-sm text-white placeholder:text-slate-600 focus:outline-none focus:border-indigo-500/50 focus:ring-2 focus:ring-indigo-500/10 transition-all disabled:cursor-not-allowed"
              placeholder="e.g. Neural Architecture Search on Edge Devices"
            />
          </div>

          <div className="flex flex-col gap-2">
            <label className="text-[11px] font-semibold text-slate-500 uppercase tracking-widest">Model Provider</label>
            <div className="grid grid-cols-2 gap-3">
              {providers.map(p => {
                const isLimited = systemStatus[p.id] === 'rate_limited' || p.unavailable;
                const isDisabled = dailyLimitReached || isLimited;
                return (
                  <button
                    key={p.id}
                    type="button"
                    disabled={isDisabled}
                    onClick={() => setProvider(p.id)}
                    className={`flex items-center gap-3 px-4 py-3 rounded-lg border text-left transition-all duration-150 ${
                      provider === p.id
                        ? 'border-indigo-500/50 bg-indigo-500/10 text-white'
                        : 'border-white/[0.06] bg-black/20 text-slate-400 hover:border-white/10 hover:bg-black/30'
                    } ${isDisabled ? 'opacity-40 cursor-not-allowed' : ''}`}
                  >
                    <span
                      className={`material-symbols-outlined text-[18px] ${provider === p.id ? 'text-indigo-400' : 'text-slate-600'}`}
                      style={{ fontVariationSettings: "'FILL' 1" }}
                    >
                      {p.icon}
                    </span>
                    <div className="flex flex-col min-w-0">
                      <span className="text-xs font-semibold">{p.label}</span>
                      <span className={`text-[10px] ${provider === p.id ? 'text-indigo-400/60' : 'text-slate-600'}`}>
                        {p.unavailable ? 'Maintenance' : (isLimited ? 'Rate Limited' : p.vendor)}
                      </span>
                    </div>
                    {provider === p.id && !isLimited && (
                      <div className="ml-auto w-1.5 h-1.5 rounded-full bg-indigo-400 shrink-0" />
                    )}
                    {isLimited && (
                      <span className="material-symbols-outlined text-[14px] text-amber-500 ml-auto">
                        {p.unavailable ? 'construction' : 'error_outline'}
                      </span>
                    )}
                  </button>
                );
              })}
            </div>
          </div>
          <div className="flex flex-col gap-2">
            <label className="text-[11px] font-semibold text-slate-500 uppercase tracking-widest">Synthesis Format</label>
            <div className="grid grid-cols-2 gap-3">
              {[
                { id: 'docx', label: 'Word Doc', icon: 'description' },
                { id: 'pdf', label: 'PDF', icon: 'picture_as_pdf' }
              ].map(f => (
                <button
                  key={f.id}
                  type="button"
                  disabled={dailyLimitReached}
                  onClick={() => setFormat(f.id)}
                  className={`flex items-center gap-3 px-4 py-3 rounded-lg border text-left transition-all duration-150 ${
                    format === f.id
                      ? 'border-indigo-500/50 bg-indigo-500/10 text-white'
                      : 'border-white/[0.06] bg-black/20 text-slate-400 hover:border-white/10 hover:bg-black/30'
                  } ${dailyLimitReached ? 'opacity-40 cursor-not-allowed' : ''}`}
                >
                  <span
                    className={`material-symbols-outlined text-[18px] ${format === f.id ? 'text-indigo-400' : 'text-slate-600'}`}
                    style={{ fontVariationSettings: "'FILL' 1" }}
                  >
                    {f.icon}
                  </span>
                  <span className="text-xs font-semibold">{f.label}</span>
                  {format === f.id && (
                    <div className="ml-auto w-1.5 h-1.5 rounded-full bg-indigo-400 shrink-0" />
                  )}
                </button>
              ))}
            </div>
          </div>

          <button
            type="submit"
            disabled={!topic.trim()}
            className="mt-1 w-full flex items-center justify-center gap-2 px-6 py-3 bg-indigo-600 hover:bg-indigo-500 disabled:opacity-40 disabled:cursor-not-allowed text-white rounded-lg text-sm font-semibold transition-all shadow-lg shadow-indigo-600/20"
          >
            <span className="material-symbols-outlined text-[18px]" style={{ fontVariationSettings: "'FILL' 1" }}>rocket_launch</span>
            Generate Research Document
          </button>
        </form>

        <p className="text-[11px] text-slate-600 text-center">
          Synthesis typically takes 2–5 minutes depending on topic complexity.
        </p>
      </div>
    </div>
  );
}
