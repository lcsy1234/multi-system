import { useEffect, useRef, useState } from "react";
import { createRoot } from "react-dom/client";
import "./style.css";

type Options = {
  factories: string[];
  lines: string[];
  skus: string[];
  start: string;
  end: string;
  mode: string;
};
type Filters = {
  Factory: string;
  Start: string;
  End: string;
  Line: string;
  SKU: string;
  Note: string;
};
type Upload = { id: string; sheets: string[]; name: string };
type Step = { tool: string; message: string };
type Report = {
  Mapping: Record<string, string>;
  Rows: number;
  Source: string;
  Sheet: string;
  Headers: string[];
  Filters: Filters;
  Warnings: string[] | null;
  Preview: Record<string, string>[] | null;
  Download: string;
};
const tools = [
  "parse_excel_template",
  "list_data_sources",
  "query_data_warehouse",
  "fill_excel_template",
];
const stages = ["解析模板", "匹配字段", "查询数据", "回填工作簿"];
async function checked(response: Response) {
  if (!response.ok) {
    const body = await response.json().catch(() => ({}));
    throw Error(body.error || "服务连接失败，请检查后端");
  }
  return response;
}
async function readEvents(
  res: Response,
  handle: (kind: string, data: string) => void,
) {
  if (!res.body) throw Error("浏览器未提供数据流");
  const reader = res.body.pipeThrough(new TextDecoderStream()).getReader();
  let buffer = "";
  try {
    while (true) {
      const { value, done } = await reader.read();
      if (done) break;
      buffer += value;
      let boundary: number;
      while ((boundary = buffer.indexOf("\n\n")) >= 0) {
        const block = buffer.slice(0, boundary);
        buffer = buffer.slice(boundary + 2);
        const lines = block.split("\n");
        handle(
          lines[0].slice(7),
          lines
            .slice(1)
            .map((s) => s.slice(6))
            .join("\n"),
        );
      }
    }
  } finally {
    reader.releaseLock();
  }
}
function App() {
  const [opts, setOpts] = useState<Options>(),
    [filters, setFilters] = useState<Filters>({
      Factory: "",
      Start: "",
      End: "",
      Line: "",
      SKU: "",
      Note: "",
    });
  const [file, setFile] = useState<Upload>(),
    [sheet, setSheet] = useState(""),
    [busy, setBusy] = useState(false),
    [loading, setLoading] = useState(false),
    [drag, setDrag] = useState(false);
  const [steps, setSteps] = useState<Step[]>([]),
    [result, setResult] = useState<Report>(),
    [error, setError] = useState("");
  const input = useRef<HTMLInputElement>(null);
  useEffect(() => {
    fetch("/api/options")
      .then(checked)
      .then((r) => r.json())
      .then((o: Options) => {
        setOpts(o);
        setFilters((f) => ({
          ...f,
          Factory: o.factories[0],
          Start: o.start,
          End: o.end,
        }));
      })
      .catch((e) => setError(e.message));
  }, []);
  function reset() {
    setResult(undefined);
    setSteps([]);
    setError("");
  }
  function change(k: keyof Filters, v: string) {
    reset();
    setFilters((f) => ({ ...f, [k]: v }));
  }
  async function upload(f?: File) {
    if (!f || busy || loading) return;
    reset();
    setFile(undefined);
    setLoading(true);
    try {
      const form = new FormData();
      form.append("file", f);
      const r = await checked(
        await fetch("/api/uploads", { method: "POST", body: form }),
      );
      const data = await r.json();
      setFile({ ...data, name: f.name });
      setSheet(data.sheets[0]);
    } catch (e) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  }
  async function run() {
    if (!file) return;
    reset();
    setBusy(true);
    let finished = false;
    try {
      const res = await checked(
        await fetch("/api/run", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ id: file.id, sheet, filters }),
        }),
      );
      await readEvents(res, (kind, raw) => {
        const d = JSON.parse(raw);
        if (kind === "step") setSteps((s) => [...s, d]);
        if (kind === "failure") throw Error(d.message);
        if (kind === "result") {
          finished = true;
          setResult(d);
        }
      });
      if (!finished) throw Error("连接提前结束，请重新运行");
    } catch (e) {
      setError(String(e));
    } finally {
      setBusy(false);
    }
  }
  const locked = busy || loading;
  function select(
    label: string,
    key: "Factory" | "Line" | "SKU",
    values: string[],
    all?: string,
  ) {
    return (
      <label>
        {label}
        <select
          disabled={locked}
          value={filters[key]}
          onChange={(e) => change(key, e.target.value)}
        >
          {all && <option value="">{all}</option>}
          {values.map((v) => (
            <option key={v}>{v}</option>
          ))}
        </select>
      </label>
    );
  }
  return (
    <>
      <header>
        <div className="brand">
          <b className="mark">元</b>
          <strong>元气森林</strong>
          <i />
          工厂数据工作台
        </div>
        <span className="online">
          <em />
          {opts ? opts.mode : "连接中"}
        </span>
      </header>
      <main>
        <div className="eyebrow">工厂工作台　/　智能填报</div>
        <h1>工厂数据智能填报</h1>
        <p className="lead">
          上传业务模板，按所选范围提取工厂数据，生成可下载的 Excel 报表。
        </p>
        <div className="workspace">
          <section>
            <div className="card">
              <h2>
                <small>01</small>上传 Excel 模板
                <a
                  className="sample"
                  href="/api/sample"
                  download="元气森林工厂生产日报模板.xlsx"
                >
                  ↓ 下载示例模板
                </a>
              </h2>
              <input
                hidden
                type="file"
                accept=".xlsx"
                ref={input}
                onChange={(e) => {
                  upload(e.target.files?.[0]);
                  e.target.value = "";
                }}
              />
              <button
                disabled={locked}
                className={"drop " + (drag ? "drag" : "")}
                onDragOver={(e) => {
                  e.preventDefault();
                  setDrag(true);
                }}
                onDragLeave={() => setDrag(false)}
                onDrop={(e) => {
                  e.preventDefault();
                  setDrag(false);
                  upload(e.dataTransfer.files[0]);
                }}
                onClick={() => input.current?.click()}
              >
                <span className="file-icon">XLSX</span>
                <strong>
                  {loading
                    ? "正在读取模板…"
                    : file
                      ? file.name
                      : "点击选择或拖入 Excel 模板"}
                </strong>
                <span>
                  {file ? "点击可更换模板" : "支持 .xlsx 格式，最大 8 MB"}
                </span>
              </button>
              {file && (
                <label className="mt-4">
                  目标工作表
                  <select
                    disabled={locked}
                    value={sheet}
                    onChange={(e) => {
                      reset();
                      setSheet(e.target.value);
                    }}
                  >
                    {file.sheets.map((s) => (
                      <option key={s}>{s}</option>
                    ))}
                  </select>
                </label>
              )}
            </div>
            <form
              className="card"
              onSubmit={(e) => {
                e.preventDefault();
                run();
              }}
            >
              <h2>
                <small>02</small>选择数据范围
              </h2>
              <div className="fields">
                {select("工厂", "Factory", opts?.factories || [])}
                {select("产线", "Line", opts?.lines || [], "全部产线")}
                <label>
                  开始日期
                  <input
                    required
                    disabled={locked}
                    type="date"
                    value={filters.Start}
                    max={filters.End}
                    onChange={(e) => change("Start", e.target.value)}
                  />
                </label>
                <label>
                  结束日期
                  <input
                    required
                    disabled={locked}
                    type="date"
                    value={filters.End}
                    min={filters.Start}
                    onChange={(e) => change("End", e.target.value)}
                  />
                </label>
                <div className="col-span-full">
                  {select("产品 / SKU", "SKU", opts?.skus || [], "全部 SKU")}
                </div>
              </div>
              <label className="mt-4">
                补充说明 <span>选填</span>
                <textarea
                  maxLength={600}
                  rows={2}
                  disabled={locked}
                  placeholder="例如：只填实际产量和良品率，汇总行保留公式"
                  value={filters.Note}
                  onChange={(e) => change("Note", e.target.value)}
                />
              </label>
              <button
                className="primary mt-5"
                disabled={!opts || !file || locked}
              >
                {busy ? "正在智能填报…" : "开始智能填报　→"}
              </button>
            </form>
          </section>
          <aside>
            <div className="card">
              <h2>
                <small>03</small>Agent 执行轨迹
              </h2>
              <ol>
                {tools.map((t, i) => {
                  const x = steps.find((s) => s.tool === t);
                  return (
                    <li key={t} className={x ? "active" : ""}>
                      <b>{x ? "●" : i + 1}</b>
                      <div>
                        <strong>{stages[i]}</strong>
                        <p>
                          {x?.message ||
                            [
                              "识别表头与公式区域",
                              "查看来源表与业务口径",
                              "按工厂和日期筛选",
                              "生成保留样式的 Excel",
                            ][i]}
                        </p>
                        {x && <code>{t}</code>}
                      </div>
                    </li>
                  );
                })}
              </ol>
            </div>
            <div className="card delivery">
              <h2>
                <small>04</small>填报结果
              </h2>
              {result ? (
                <>
                  <strong className="count">{result.Rows}</strong> 行已填报
                  <p>
                    {result.Sheet} · {result.Filters.Factory}
                    <br />
                    {result.Filters.Start} 至 {result.Filters.End}
                    <br />
                    {result.Filters.Line || "全部产线"} ·{" "}
                    {result.Filters.SKU || "全部 SKU"}
                  </p>
                  <code>{result.Source}</code>
                  <a className="primary mt-4" href={result.Download}>
                    ↓ 下载填报结果
                  </a>
                </>
              ) : (
                <p>
                  {busy
                    ? "处理中，完成后可下载文件。"
                    : "等待填报完成，可在此下载工作簿。"}
                </p>
              )}
            </div>
          </aside>
        </div>
        {error && (
          <div role="alert" className="alert">
            {error}
          </div>
        )}
        {result && (
          <section className="card">
            <h2>
              数据预览 <span>前 20 行</span>
            </h2>
            {result.Preview?.length ? (
              <div className="scroll">
                <table>
                  <thead>
                    <tr>
                      {result.Headers.map((h, i) => (
                        <th key={i}>{h}</th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {result.Preview.map((row, i) => (
                      <tr key={i}>
                        {result.Headers.map((h, k) => (
                          <td key={k}>{row[h] || "—"}</td>
                        ))}
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <p>没有可回填的数据行。</p>
            )}
            <details className="mapping">
              <summary>查看字段映射</summary>
              <div className="scroll">
                <table>
                  <thead>
                    <tr>
                      <th>模板表头</th>
                      <th>来源字段</th>
                    </tr>
                  </thead>
                  <tbody>
                    {Object.entries(result.Mapping).map(([header, field]) => (
                      <tr key={header}>
                        <td>{header}</td>
                        <td>{field}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </details>
            {!!result.Warnings?.length && (
              <details open>
                <summary>缺失与填报说明（{result.Warnings.length}）</summary>
                {result.Warnings.map((s, i) => (
                  <p key={i}>{s}</p>
                ))}
              </details>
            )}
          </section>
        )}
      </main>
    </>
  );
}
createRoot(document.getElementById("root")!).render(<App />);
