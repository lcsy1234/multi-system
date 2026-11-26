import { parseConfig } from './scriptJson';
export type Rule = { name: string; code: string };
const key = 'factory.parser-rules.v1';
export const readRules = (): Rule[] => JSON.parse(localStorage.getItem(key) ?? '[]');
export const writeRules = (rules: Rule[]) => localStorage.setItem(key, JSON.stringify(rules));
export async function runRules(sample: string, rules: Rule[]): Promise<{json?: string; rule?: Rule}> {
  if (!rules.length) return {};
  return new Promise((resolve, reject) => {
    const worker = new Worker(new URL('./parser.worker.ts', import.meta.url), { type: 'module' });
    const timer = setTimeout(() => { worker.terminate(); reject(Error('脚本运行超时')); }, 5000);
    const stop = () => { clearTimeout(timer); worker.terminate(); };
    worker.onmessage = ({data}) => { stop(); if (data.error) reject(Error(data.error)); else resolve(data); };
    worker.onerror = e => { stop(); reject(Error(e.message)); };
    worker.postMessage({sample, rules});
  });
}
export async function convertSample(sample: string) {
  const match = await runRules(sample, readRules());
  return { formatted: JSON.stringify(match.json === undefined ? parseConfig(sample) : JSON.parse(match.json), null, 2), rule: match.rule };
}
