import { useState } from 'react';
import { render } from 'react-dom';

declare const ConfigParser: { parseConfig(source: string): unknown };

function OfflineConverter() {
  const [source, setSource] = useState('[Bridge] 111\nBindIP=127.0.0.1\nBindPort=30001');
  const [result, setResult] = useState('');
  const [error, setError] = useState('');

  function convert() {
    try {
      setResult(JSON.stringify(ConfigParser.parseConfig(source), null, 2));
      setError('');
    } catch (cause) {
      setResult('');
      setError(cause instanceof Error ? cause.message : '转换失败，请检查输入格式。');
    }
  }

  return <main>
    <h1>配置转 JSON</h1>
    <p>粘贴 INI、键值对、JSON 或 XML。本工具完全在浏览器本地运行。</p>
    <label htmlFor="source">原始内容</label>
    <textarea id="source" rows={10} value={source} onChange={event => {
      setSource(event.target.value);
      setResult('');
      setError('');
    }} />
    <button disabled={!source.trim()} onClick={convert}>转换为 JSON</button>
    {error && <p role="alert">{error}</p>}
    <h2>结果</h2><pre id="result" aria-live="polite">{result}</pre>
  </main>;
}

render(<OfflineConverter />, document.getElementById('root'));
