import { userError } from './errorMessages';
import { useContext, useState } from 'react';
import { RequestContext } from './service';
import { FETCH_URLS } from './config.url.tablehandle';
import { convertSample, readRules, runRules, writeRules, type Rule } from './parserRules';

const examples = { JSON: '{"工单":"0017","数量":12}', XML: '<工单 编号="0017"><工序>灌装</工序><工序>包装</工序></工单>', INI: '[Settings]\nSaveDir=D:\\工厂\\模板\n[Bridge]\nBindIP=127.0.0.1\nBindPort=30001' };

export function JsonConverter({local = false, initial = '', onOptimize}: {local?: boolean; initial?: string; onOptimize?: (sample: string) => void}) {
  const [content, setContent] = useState(initial);
  const request = useContext(RequestContext);
  const [instruction, setInstruction] = useState('');
  const [result, setResult] = useState('');
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
  const [candidate, setCandidate] = useState<Rule>();
  const [saved, setSaved] = useState(false);
  const [rules, setRules] = useState(readRules);
  function reset() { setResult(''); setError(''); setCandidate(undefined); setSaved(false); }
  function edit(value: string) { reset(); setContent(value); }
  async function run() {
    reset(); setBusy(true);
    try {
      if (new TextEncoder().encode(content).length > 65536) throw Error('样本不能超过 64 KiB');
      if (local) { setResult((await convertSample(content)).formatted); return; }
      let actual: string, previous = '';
      try { const current = await convertSample(content); actual = current.formatted; previous = current.rule?.code ?? ''; }
      catch (e) { actual = String(e); }
      const data = await request<Rule>(FETCH_URLS.optimize, {sample:content,instruction,actual:Array.from(actual).slice(0,16000).join(''),previous});
      const trial = await runRules(content, [data]);
      if (trial.json === undefined) throw Error('生成的脚本未识别该样本，请补充格式特征后重试');
      setResult(JSON.stringify(JSON.parse(trial.json), null, 2)); setCandidate(data);
    } catch (e) { setError(userError(e, local ? '解析脚本执行失败，请检查输入格式或优化脚本。' : '脚本生成或试跑失败，请调整处理要求后重试。')); }
    finally { setBusy(false); }
  }
  function update(next: Rule[]) { writeRules(next); setRules(next); }
  function apply() {
    if (!candidate) return;
    try { update([candidate, ...readRules().filter(r => r.code !== candidate.code)]); setSaved(true); }
    catch { setError('浏览器保存失败，请检查本地存储空间'); }
  }
  return <main><div className="eyebrow">工厂工作台　/　数据转换</div><h1>{local ? '脚本转 JSON' : 'AI 优化解析脚本'}</h1>
    <p className="lead">{local ? '在浏览器中运行已应用的解析脚本，支持 XML、JSON 和 INI 等配置。' : '填写遇到的格式和处理要求。点击生成会将样本、当前结果和要求发送给 DeepSeek，请先移除密钥等敏感信息。'}</p>
    <section className="card"><h2>原始内容</h2><div className="example-row">填入示例：{Object.entries(examples).map(([name,value]) => <button key={name} disabled={busy} onClick={() => edit(value)}>{name}</button>)}</div>
      <textarea aria-label="原始内容" rows={8} disabled={busy} value={content} onChange={e => edit(e.target.value)} />
      {!local && <label>希望脚本怎样处理这种格式？<textarea rows={3} disabled={busy} value={instruction} onChange={e => {reset();setInstruction(e.target.value);}} placeholder="例如：每个 [分组名] 建一个对象，端口转数字，目录保留反斜杠。" /></label>}
      <button className="primary mt-4" disabled={busy || !content.trim() || (!local && !instruction.trim())} onClick={run}>{busy ? '处理中…' : local ? '转换为 JSON' : '生成并试跑脚本'}</button>
      {error && <p role="alert" className="mt-4 text-red-700">{error}</p>}
      {local && (error || result) && <button onClick={() => onOptimize?.(content)}>带此样本优化脚本</button>}
    </section>
    {result && <section className="card"><h2>{local ? '转换结果' : '新脚本试跑结果'}</h2><pre className="json-preview">{result}</pre>{candidate && <><details><summary>查看脚本 · {candidate.name}</summary><pre className="json-preview">{candidate.code}</pre></details><button className="primary mt-4" disabled={saved} onClick={apply}>{saved ? '已应用到本浏览器的脚本转 JSON' : '应用到脚本转 JSON'}</button></>}</section>}
    {!local && rules.length > 0 && <section className="card"><h2>本浏览器已应用的脚本</h2>{rules.map(r => <p key={r.code}>{r.name}　<button onClick={() => {try { update(readRules().filter(old => old.code !== r.code));setSaved(false); } catch { setError('无法移除脚本，请检查浏览器的存储权限。'); }}}>移除</button></p>)}</section>}
  </main>;
}
