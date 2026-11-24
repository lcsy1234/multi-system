import React, { useEffect, useState } from 'react';
import { createRoot } from 'react-dom/client';

function App() {
  const [status, setStatus] = useState('正在连接…');
  useEffect(() => {
    async function checkHealth() {
      try {
        const response = await fetch('/api/health');
        if (!response.ok || (await response.json()).status !== 'ok') throw new Error();
        setStatus('已连接');
      } catch {
        setStatus('未连接，请启动后端后刷新页面');
      }
    }
    checkHealth();
  }, []);
  return <main><h1>表格处理系统</h1><p role="status">后端服务：{status}</p></main>;
}

createRoot(document.getElementById('root')).render(<App />);
