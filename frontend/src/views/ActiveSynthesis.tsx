import React, { useEffect, useRef, useState } from 'react';

interface Props {
  taskId: string;
  topic: string;
  onComplete: (resultUrl: string) => void;
  onReset: () => void;
}

interface AgentEvent {
  id: number;
  message: string;
  type: 'init' | 'research' | 'tool' | 'synthesis' | 'write' | 'upload' | 'warn' | 'done';
  ts: string;
}

function classifyEvent(msg: string): AgentEvent['type'] {
  const m = msg.toLowerCase();
  if (m.includes('uploading') || m.includes('cloud')) return 'upload';
  if (m.includes('finalized') || m.includes('complete') || m.includes('done')) return 'done';
  if (m.includes('warning') || m.includes('error') || m.includes('failed')) return 'warn';
  if (m.includes('writing') || m.includes('generating') || m.includes('document') || m.includes('docx')) return 'write';
  if (m.includes('calling') || m.includes('tool') || m.includes('function') || m.includes('invoke')) return 'tool';
  if (m.includes('synthesiz') || m.includes('analyz') || m.includes('processing') || m.includes('reasoning')) return 'synthesis';
  if (m.includes('search') || m.includes('fetch') || m.includes('retriev') || m.includes('query') || m.includes('research')) return 'research';
  return 'init';
}

const EVENT_STYLE: Record<AgentEvent['type'], { icon: string; color: string; bg: string; border: string }> = {
  init:      { icon: 'radio_button_checked', color: 'text-slate-400',  bg: 'bg-slate-800/60',      border: 'border-slate-700/50'  },
  research:  { icon: 'travel_explore',       color: 'text-sky-400',    bg: 'bg-sky-500/8',         border: 'border-sky-500/20'    },
  tool:      { icon: 'settings',             color: 'text-amber-400',  bg: 'bg-amber-500/8',       border: 'border-amber-500/20'  },
  synthesis: { icon: 'psychology',           color: 'text-violet-400', bg: 'bg-violet-500/8',      border: 'border-violet-500/20' },
  write:     { icon: 'edit_document',        color: 'text-indigo-400', bg: 'bg-indigo-500/8',      border: 'border-indigo-500/20' },
  upload:    { icon: 'cloud_upload',         color: 'text-teal-400',   bg: 'bg-teal-500/8',        border: 'border-teal-500/20'   },
  warn:      { icon: 'warning',              color: 'text-red-400',    bg: 'bg-red-500/8',         border: 'border-red-500/20'    },
  done:      { icon: 'task_alt',             color: 'text-emerald-400',bg: 'bg-emerald-500/8',     border: 'border-emerald-500/20'},
};

export default function ActiveSynthesis({ taskId, topic, onComplete, onReset }: Props) {
  const [events, setEvents] = useState<AgentEvent[]>([]);
  const [status, setStatus] = useState<'running' | 'done' | 'error'>('running');
  const [eventCount, setEventCount] = useState(0);
  const logsEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    logsEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [events]);

  useEffect(() => {
    let evtSource: EventSource;
    try {
      evtSource = new EventSource(`/api/v1/tasks/${taskId}`);

      evtSource.onmessage = (event) => {
        const data = JSON.parse(event.data);
        const msgs: string[] = data.newLogs && data.newLogs.length > 0
          ? data.newLogs
          : data.progress
            ? [data.progress]
            : [];

        if (msgs.length > 0) {
          setEvents(prev => {
            let id = prev.length;
            const next = [...prev];
            for (const msg of msgs) {
              next.push({ id: id++, message: msg, type: classifyEvent(msg), ts: new Date().toLocaleTimeString([], { hour12: false }) });
            }
            return next;
          });
          setEventCount(c => c + msgs.length);
        }

        if (data.complete) {
          setStatus('done');
          evtSource.close();
          setTimeout(() => onComplete(data.result ?? ''), 1200);
        }
        if (data.status === 'error') {
          setStatus('error');
          evtSource.close();
        }
      };

      evtSource.onerror = () => {
        setStatus('error');
        evtSource.close();
      };
    } catch (e) {
      console.error(e);
    }
    return () => evtSource?.close();
  }, [taskId, onComplete]);

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="min-h-full max-w-3xl mx-auto px-6 py-10 flex flex-col gap-6">

        {/* Header */}
        <div className="flex items-start justify-between gap-4">
          <div>
            <p className="text-[11px] font-semibold text-slate-500 uppercase tracking-widest mb-1">Agent Run</p>
            <h1 className="text-2xl font-semibold text-white tracking-tight truncate max-w-sm md:max-w-lg">{topic}</h1>
          </div>
          <div className={`shrink-0 flex items-center gap-2 px-3 py-1.5 rounded-full border text-[11px] font-mono tracking-wide ${
            status === 'done'  ? 'bg-emerald-500/10 border-emerald-500/25 text-emerald-400' :
            status === 'error' ? 'bg-red-500/10 border-red-500/25 text-red-400' :
                                 'bg-white/[0.04] border-white/[0.06] text-slate-400'
          }`}>
            <div className={`w-1.5 h-1.5 rounded-full ${
              status === 'done'  ? 'bg-emerald-400' :
              status === 'error' ? 'bg-red-400' :
                                   'bg-indigo-400 animate-pulse'
            }`} />
            {status === 'done' ? 'COMPLETE' : status === 'error' ? 'FAILED' : 'RUNNING'}
          </div>
        </div>

        {/* Stats bar */}
        <div className="grid grid-cols-3 gap-3">
          {[
            { label: 'Events', value: eventCount },
            { label: 'Task ID', value: taskId.slice(0, 8) + '…' },
            { label: 'Status', value: status === 'running' ? 'Active' : status === 'done' ? 'Done' : 'Error' },
          ].map(stat => (
            <div key={stat.label} className="rounded-lg bg-white/[0.02] border border-white/[0.05] px-4 py-3">
              <p className="text-[10px] font-semibold text-slate-600 uppercase tracking-widest mb-0.5">{stat.label}</p>
              <p className="text-sm font-mono text-slate-300">{stat.value}</p>
            </div>
          ))}
        </div>

        {/* Event feed */}
        <div className="flex flex-col gap-2">
          {events.length === 0 && (
            <div className="flex items-center gap-3 px-4 py-3 rounded-lg bg-white/[0.02] border border-white/[0.04]">
              <div className="w-1.5 h-1.5 rounded-full bg-indigo-400 animate-pulse shrink-0" />
              <span className="text-xs text-slate-500 font-mono">Awaiting agent stream…</span>
            </div>
          )}
          {events.map(ev => {
            const s = EVENT_STYLE[ev.type];
            return (
              <div
                key={ev.id}
                className={`flex items-start gap-3 px-4 py-3 rounded-lg border ${s.bg} ${s.border} transition-all`}
              >
                <span
                  className={`material-symbols-outlined text-[16px] shrink-0 mt-0.5 ${s.color}`}
                  style={{ fontVariationSettings: "'FILL' 1" }}
                >
                  {s.icon}
                </span>
                <p className="flex-1 text-[13px] text-slate-300 leading-relaxed font-mono">{ev.message}</p>
                <span className="text-[10px] text-slate-600 font-mono shrink-0 mt-0.5">{ev.ts}</span>
              </div>
            );
          })}
          <div ref={logsEndRef} />
        </div>

        {status === 'running' && (
          <div className="flex items-center gap-2 px-4 py-2.5 rounded-lg bg-indigo-500/5 border border-indigo-500/15">
            <span className="material-symbols-outlined text-[16px] text-indigo-400 animate-spin">progress_activity</span>
            <span className="text-xs text-slate-400">Agent is running… this may take a few minutes.</span>
          </div>
        )}

        {status === 'error' && (
          <div className="flex flex-col items-center gap-4 py-6 px-4 rounded-xl bg-red-500/[0.03] border border-red-500/10 mt-2">
            <div className="flex items-center gap-3">
              <span className="material-symbols-outlined text-red-500 text-xl">error_outline</span>
              <p className="text-sm text-slate-400">The synthesis pipeline encountered a terminal error.</p>
            </div>
            <button
              onClick={onReset}
              className="px-6 py-2 bg-red-500/10 hover:bg-red-500/15 text-red-400 border border-red-500/20 rounded-lg text-xs font-semibold transition-all"
            >
              Back to Dashboard
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
