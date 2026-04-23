import React from 'react';

export default function TopAppBar() {
  return (
    <header className="fixed top-0 w-full z-50 flex justify-between items-center px-6 md:px-10 h-16 border-b border-white/[0.05] bg-[#080A0F]/90 backdrop-blur-xl">
      <div className="flex items-center gap-3">
        <div className="w-7 h-7 rounded-md bg-indigo-600 flex items-center justify-center shadow-lg shadow-indigo-600/30">
          <span className="text-white text-xs font-bold tracking-tight">K</span>
        </div>
        <span className="text-sm font-semibold text-white/90 tracking-tight">Kyle</span>
        <span className="hidden sm:block text-xs text-slate-600 font-medium">/ Research Intelligence</span>
      </div>

      <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-white/[0.03] border border-white/[0.06]">
        <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse shadow-sm shadow-emerald-500/50" />
        <span className="text-[10px] font-mono text-slate-400 tracking-wide">ONLINE</span>
      </div>
    </header>
  );
}
