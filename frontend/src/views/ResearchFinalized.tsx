import React from 'react';

interface Props {
  taskId: string;
  topic: string;
  resultUrl: string | null;
  onReset: () => void;
}

export default function ResearchFinalized({ taskId, topic, resultUrl, onReset }: Props) {
  const downloadHref = resultUrl || `/api/v1/download/${taskId}`;
  const pdfHref = `/api/v1/download/${taskId}/pdf`;
  const baseName = `kyle_research_${topic.replace(/\s+/g, '_').toLowerCase().slice(0, 60)}`;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="min-h-full max-w-3xl mx-auto px-6 py-10 flex flex-col gap-6">

        {/* Header */}
        <div className="flex flex-col gap-3">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/20 w-fit">
            <span className="material-symbols-outlined text-[14px] text-emerald-400" style={{ fontVariationSettings: "'FILL' 1" }}>check_circle</span>
            <span className="text-[11px] font-semibold text-emerald-400 tracking-widest uppercase">Research Complete</span>
          </div>
          <h1 className="text-2xl md:text-3xl font-semibold text-white tracking-tight leading-snug">{topic || 'Research Document'}</h1>
          <p className="text-sm text-slate-400 leading-relaxed">
            Analysis finalized. Your document has been synthesized and is ready for download.
          </p>
        </div>

        {/* Actions */}
        <div className="flex flex-wrap items-center gap-3">
          <a
            href={downloadHref}
            download={`${baseName}.docx`}
            target="_blank"
            rel="noreferrer"
            className="inline-flex items-center gap-2 px-5 py-2.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-semibold transition-all shadow-lg shadow-indigo-600/20"
          >
            <span className="material-symbols-outlined text-[18px]" style={{ fontVariationSettings: "'FILL' 1" }}>description</span>
            Download DOCX
          </a>
          <a
            href={pdfHref}
            download={`${baseName}.pdf`}
            className="inline-flex items-center gap-2 px-5 py-2.5 rounded-lg bg-white/[0.04] border border-white/[0.07] text-slate-300 hover:text-white hover:bg-white/[0.07] text-sm font-medium transition-all"
          >
            <span className="material-symbols-outlined text-[18px]" style={{ fontVariationSettings: "'FILL' 1" }}>picture_as_pdf</span>
            Download PDF
          </a>
          <button
            onClick={onReset}
            className="flex items-center gap-2 px-5 py-2.5 rounded-lg bg-white/[0.04] border border-white/[0.07] text-slate-300 hover:text-white hover:bg-white/[0.07] text-sm font-medium transition-all"
          >
            <span className="material-symbols-outlined text-[18px]">add</span>
            New Research
          </button>
        </div>

        <div className="h-px w-full bg-white/[0.05]" />

        {/* Meta */}
        <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
          {[
            { label: 'Task ID', value: taskId.slice(0, 10) + '…', icon: 'tag' },
            { label: 'Status', value: 'Finalized', icon: 'check_circle' },
            { label: 'Delivery', value: resultUrl ? 'Cloud' : 'Local', icon: 'cloud_done' },
          ].map(item => (
            <div key={item.label} className="flex items-center gap-3 px-4 py-3 rounded-lg bg-white/[0.02] border border-white/[0.05]">
              <span className="material-symbols-outlined text-[16px] text-slate-500" style={{ fontVariationSettings: "'FILL' 1" }}>{item.icon}</span>
              <div>
                <p className="text-[10px] font-semibold text-slate-600 uppercase tracking-widest">{item.label}</p>
                <p className="text-xs font-mono text-slate-300 mt-0.5">{item.value}</p>
              </div>
            </div>
          ))}
        </div>

        {/* Summary section */}
        <div className="rounded-xl bg-white/[0.02] border border-white/[0.06] p-6 flex flex-col gap-5">
          <h2 className="flex items-center gap-2 text-sm font-semibold text-white">
            <span className="material-symbols-outlined text-[18px] text-indigo-400" style={{ fontVariationSettings: "'FILL' 1" }}>auto_awesome</span>
            What was produced
          </h2>
          <p className="text-sm text-slate-400 leading-relaxed">
            Kyle ran a multi-step agentic research pipeline on <span className="text-slate-200 font-medium">"{topic}"</span>.
            The agent gathered primary sources, reasoned over findings, synthesized conclusions,
            and produced a structured Word document ready for download above.
          </p>

          <div className="h-px w-full bg-white/[0.05]" />

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            {[
              { icon: 'travel_explore', color: 'text-sky-400',    bg: 'bg-sky-500/8 border-sky-500/15',     label: 'Research',  desc: 'Sources gathered & indexed' },
              { icon: 'psychology',     color: 'text-violet-400', bg: 'bg-violet-500/8 border-violet-500/15',label: 'Analysis',  desc: 'Reasoning & synthesis complete' },
              { icon: 'edit_document',  color: 'text-indigo-400', bg: 'bg-indigo-500/8 border-indigo-500/15',label: 'Document',  desc: 'DOCX generated & delivered' },
            ].map(step => (
              <div key={step.label} className={`flex flex-col gap-2 rounded-lg border p-4 ${step.bg}`}>
                <span className={`material-symbols-outlined text-[20px] ${step.color}`} style={{ fontVariationSettings: "'FILL' 1" }}>{step.icon}</span>
                <p className="text-xs font-semibold text-white">{step.label}</p>
                <p className="text-[11px] text-slate-500 leading-snug">{step.desc}</p>
              </div>
            ))}
          </div>
        </div>

        {resultUrl && (
          <div className="flex items-start gap-3 px-4 py-3 rounded-lg bg-emerald-500/5 border border-emerald-500/15">
            <span className="material-symbols-outlined text-[16px] text-emerald-400 mt-0.5" style={{ fontVariationSettings: "'FILL' 1" }}>cloud_done</span>
            <div>
              <p className="text-xs font-semibold text-emerald-400 mb-0.5">Document uploaded to cloud</p>
              <p className="text-[11px] text-slate-500 font-mono break-all">{resultUrl}</p>
            </div>
          </div>
        )}

      </div>
    </div>
  );
}
