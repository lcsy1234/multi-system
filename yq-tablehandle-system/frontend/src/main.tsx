import { useEffect, useState } from "react";
import { render } from "react-dom";
import { Drawer, Menu } from "antd";
import "./assets/style/global.less";
import ReportPage from "./components/Report";
import { JsonConverter } from "./JsonConverter";
import { ShrimpConverter } from "./ShrimpConverter";

const menuItems = [
  { key: "reports", label: "智能填报", children: [
    { key: "#report", label: "工厂数据智能填报" },
  ] },
  { key: "scripts", label: "AI 脚本", children: [
    { key: "#json", label: "AI 优化脚本" },
    { key: "#script-json", label: "脚本转 JSON" },
    { key: "#shrimp-json", label: "元气小虾米转 JSON" },
  ] },
];
function App() {
  const [mode, setMode] = useState("连接中");
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [page, setPage] = useState(location.hash);
  const [optimization, setOptimization] = useState({sample:"", revision:0});
  const converter = page === "#json";
  const script = page === "#script-json";
  const shrimp = page === "#shrimp-json";
  const navigation = (
    <nav aria-label="工作台模块">
      <Menu mode="inline" items={menuItems} defaultOpenKeys={["reports", "scripts"]}
        selectedKeys={[converter || script || shrimp ? page : "#report"]}
        onClick={({ key }) => { location.hash = key; setDrawerOpen(false); }} />
    </nav>
  );
  useEffect(() => {
    const navigate = () => setPage(location.hash);
    window.addEventListener("hashchange", navigate);
    return () => window.removeEventListener("hashchange", navigate);
  }, []);
  return (
    <div className="pangu-plugin-tablehandle">
      <header>
        <div className="brand">
          <button className="nav-toggle" aria-label="打开工作台菜单" aria-expanded={drawerOpen} onClick={() => setDrawerOpen(true)}>☰</button>
          <b className="mark">元</b>
          <strong>元气森林</strong>
          <i />
          工厂数据工作台
        </div>
        <span className="online">
          <em />
          {shrimp ? "元气小虾米" : script ? "本地脚本" : converter ? "DeepSeek" : mode}
        </span>
      </header>
      <div className="workbench-layout">
          <div className="module-sidebar">{navigation}</div>
          <Drawer title="工作台菜单" placement="left" width={240} open={drawerOpen}
            bodyStyle={{ padding: 0 }} onClose={() => setDrawerOpen(false)}>{navigation}</Drawer>
      <div className="workbench-content">
      <div hidden={!shrimp}><ShrimpConverter /></div>
      <div hidden={!converter}><JsonConverter key={optimization.revision} initial={optimization.sample} /></div>
      <div hidden={!script}><JsonConverter local onOptimize={sample => { setOptimization(prev => ({sample, revision:prev.revision + 1})); location.hash = "json"; }} /></div>
        <div hidden={converter || script || shrimp}><ReportPage onMode={setMode} /></div>
      </div>
      </div>
    </div>
  );
}
render(<App />, document.getElementById("root"));
