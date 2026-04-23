/**
 * Generates a stable client-side fingerprint based on 3 factors:
 * 1. Canvas Fingerprinting (Rendering quirks)
 * 2. Audio Context (Hardware dynamics)
 * 3. Hardware/Screen resolution
 */
export async function getFingerprint(): Promise<string> {
  const factors: string[] = [];

  // 1. Canvas Fingerprinting
  try {
    const canvas = document.createElement('canvas');
    const ctx = canvas.getContext('2d');
    if (ctx) {
      ctx.textBaseline = "top";
      ctx.font = "14px 'Arial'";
      ctx.textBaseline = "alphabetic";
      ctx.fillStyle = "#f60";
      ctx.fillRect(125, 1, 62, 20);
      ctx.fillStyle = "#069";
      ctx.fillText("KyleResearch", 2, 15);
      ctx.fillStyle = "rgba(102, 204, 0, 0.7)";
      ctx.fillText("KyleResearch", 4, 17);
      factors.push(canvas.toDataURL());
    }
  } catch (e) {
    factors.push('canvas-fail');
  }

  // 2. Hardware/Screen quirks
  factors.push(`${window.screen.width}x${window.screen.height}x${window.screen.colorDepth}`);
  factors.push(`${navigator.hardwareConcurrency || 1}`);

  // 3. Audio Context (Simplified)
  try {
    const audioCtx = new (window.AudioContext || (window as any).webkitAudioContext)();
    factors.push(`${audioCtx.sampleRate}`);
    await audioCtx.close();
  } catch (e) {
    factors.push('audio-fail');
  }

  const combined = factors.join('|');
  
  // Simple hash function (Murmur or similar logic)
  let hash = 0;
  for (let i = 0; i < combined.length; i++) {
    const char = combined.charCodeAt(i);
    hash = ((hash << 5) - hash) + char;
    hash |= 0; // Convert to 32bit integer
  }
  
  return Math.abs(hash).toString(16);
}
