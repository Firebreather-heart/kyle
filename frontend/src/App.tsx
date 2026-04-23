/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */
import { useState, useEffect, useRef } from 'react';
import { getFingerprint } from './utils/fingerprint';
import TopAppBar from './components/TopAppBar';
import IdleDashboard from './views/IdleDashboard';
import ActiveSynthesis from './views/ActiveSynthesis';
import ResearchFinalized from './views/ResearchFinalized';

export default function App() {
  const [currentView, setCurrentView] = useState<'idle' | 'synthesis' | 'finalized'>('idle');
  const [taskId, setTaskId] = useState<string | null>(null);
  const [fingerprint, setFingerprint] = useState<string | null>(null);
  const [dailyLimitReached, setDailyLimitReached] = useState(false);
  const [docsGeneratedToday, setDocsGeneratedToday] = useState(0);
  const [resultUrl, setResultUrl] = useState<string | null>(null);
  const [topic, setTopic] = useState<string>('');
  const [systemStatus, setSystemStatus] = useState<Record<string, string>>({});

  const initializationRef = useRef(false);

  useEffect(() => {
    if (initializationRef.current) return;
    initializationRef.current = true;

    async function init() {
      const fp = await getFingerprint();
      const res = await fetch('/api/v1/obtain-token', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        cache: 'no-store',
        body: JSON.stringify({ fingerprint: fp }),
      });
      if (res.ok) {
        const data = await res.json();
        setDailyLimitReached(data.user.daily_limit_reached);
        setDocsGeneratedToday(data.user.docs_generated_today);
        setSystemStatus(data.system_status || {});
      } else if (res.status === 429) {
        setDailyLimitReached(true);
      }
    }
    init();
  }, []);

  const handleStartSynthesis = async (t: string, provider: string, format: string) => {
    try {
      const res = await fetch('/api/v1/generate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ topic: t, provider, format })
      });
      if (!res.ok) throw new Error(`HTTP error! status: ${res.status}`);
      const data = await res.json();
      if (data.task_id) {
        setTaskId(data.task_id);
        setTopic(t);
        setCurrentView('synthesis');
      }
    } catch (error: any) {
      console.error("Failed to start synthesis", error);
      alert(`API Error: ${error.message}`);
    }
  };

  const handleSynthesisComplete = (url: string) => {
    setResultUrl(url);
    setCurrentView('finalized');
  };

  const handleReset = () => {
    setCurrentView('idle');
    setTaskId(null);
    setResultUrl(null);
    setTopic('');
  };

  return (
    <div className="flex h-screen w-full relative bg-[#080A0F] text-slate-300 font-sans overflow-hidden">
      <div className="ambient-mesh" aria-hidden="true" />
      <TopAppBar />

      <main className="flex-1 pt-16 h-screen overflow-hidden relative z-10 flex flex-col">
        {currentView === 'idle' && (
          <IdleDashboard 
            onStartSynthesis={(t, p, f) => handleStartSynthesis(t, p, f)} 
            dailyLimitReached={dailyLimitReached}
            docsGeneratedToday={docsGeneratedToday}
            systemStatus={systemStatus}
          />
        )}
        {currentView === 'synthesis' && taskId && (
          <ActiveSynthesis 
            taskId={taskId} 
            topic={topic} 
            onComplete={(url: string) => handleSynthesisComplete(url)} 
            onReset={handleReset}
          />
        )}
        {currentView === 'finalized' && taskId && (
          <ResearchFinalized 
            taskId={taskId} 
            topic={topic} 
            resultUrl={resultUrl || ''} 
            onReset={handleReset} 
          />
        )}
      </main>
    </div>
  );
}
