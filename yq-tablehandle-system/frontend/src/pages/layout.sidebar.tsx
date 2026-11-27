import { useEffect, useState, type ReactNode } from 'react';
import { Drawer, Menu, type MenuProps } from 'antd';
import { useLocation } from 'react-router-dom';
import { getPageMeta, getRouter, pickPluginBranches, request, useMenu, type PluginMenuNode } from '@pangu/plugin-sdk';
import { RequestContext, type JsonRequest } from '../service';

const hostRequest: JsonRequest = async <Result,>(url: string, data?: unknown) => {
  const response = await request<Result>({ url, method: data === undefined ? 'get' : 'post', data });
  return response.data;
};

function items(nodes: PluginMenuNode[]): MenuProps['items'] {
  return nodes.filter(node => !node.hide).map(node => ({ key: node.path || node.key, label: node.title,
    children: node.children?.some(child => !child.hide) ? items(node.children) : undefined,
  }));
}

export default function Sidebar({ children }: { children?: ReactNode }) {
  const { tree, ready } = useMenu();
  const { pathname } = useLocation();
  const [open, setOpen] = useState<string[]>([]);
  const [drawer, setDrawer] = useState(false);
  const branches = ready ? pickPluginBranches(tree, 'tablehandle') : [];
  const meta = ready ? getPageMeta(pathname) : null;
  useEffect(() => { setOpen(ready ? getPageMeta(pathname)?.openKeys ?? [] : []); setDrawer(false); }, [ready, pathname, tree]);
  const menu = <nav aria-label="工作台模块"><Menu mode="inline" openKeys={open} onOpenChange={setOpen}
    selectedKeys={meta?.selectedKeys ?? []} items={items(branches.length === 1 && !branches[0].path ? branches[0].children ?? [] : branches)}
    onClick={({ key }) => { setDrawer(false); if (key !== pathname) getRouter().push(key); }} /></nav>;
  return <RequestContext.Provider value={hostRequest}><div className="pangu-plugin-tablehandle">
    <header><div className="brand"><button className="nav-toggle" aria-label="打开工作台菜单" onClick={() => setDrawer(true)}>☰</button><b className="mark">元</b><strong>元气森林</strong><i />工厂数据工作台</div></header>
    <div className="workbench-layout"><div className="module-sidebar">{menu}</div>
      <Drawer title="工作台菜单" placement="left" open={drawer} width={240} onClose={() => setDrawer(false)}>{menu}</Drawer>
      <div id="yqsl-content" className="workbench-content">{children}</div>
    </div></div></RequestContext.Provider>;
}
