import { userError } from './errorMessages';
import { useContext, useState } from 'react';
import { RequestContext } from './service';
import { FETCH_URLS } from './config.url.tablehandle';

export function ShrimpConverter() {
  const request = useContext(RequestContext);
  const [content, setContent] = useState('');
  const [result, setResult] = useState('');
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
  function edit(value: string) { setContent(value); setResult(''); setError(''); }
  async function convert() {
    setResult(''); setError('');
    if (new TextEncoder().encode(content).length > 65536) { setError('原始内容不能超过 64 KiB'); return; }
    setBusy(true);
    try {
      const data = await request<unknown>(FETCH_URLS.shrimp, {content});
      setResult(JSON.stringify(data, null, 2));
    } catch (err) { setError(userError(err, '转换服务返回异常，请稍后重试。')); }
    finally { setBusy(false); }
  }
  return <main><div className="eyebrow">工厂工作台　/　数据转换</div><h1>元气小虾米转 JSON</h1>
    <p className="lead">将 JSON、XML 或键值对文本发送给元气小虾米。提交前请移除密钥等敏感信息。</p>
    <section className="card"><h2><small>01</small>原始内容<button className="sample" disabled={busy} onClick={()=>edit('[工单]\n编号=0017\n数量=12')}>填入示例</button></h2>
    <textarea aria-label="小虾米原始内容" rows={8} value={content} disabled={busy} onChange={e=>edit(e.target.value)} />
    <button className="primary mt-4" disabled={busy || !content.trim()} onClick={convert}>{busy ? '小虾米正在转换…' : '转换为 JSON'}</button>
    {error && <p role="alert" className="mt-4 text-red-700">{error}</p>}</section>
    {result && <section className="card"><h2><small>02</small>转换结果</h2><details><summary>查看格式化 JSON</summary><pre className="json-preview">{result}</pre></details></section>}</main>;
}
