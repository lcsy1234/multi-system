const root = document.getElementById('root');
if (root && !('__pangu' in window)) root.textContent = `工厂数据工作台插件已启动。请在盘古 /pd 中将 tablehandle 映射到 ${location.origin}，再打开对应菜单。`;
