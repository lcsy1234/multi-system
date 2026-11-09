import { createRoot } from "react-dom/client";
import { App as AntApp, ConfigProvider } from "antd";
import zhCN from "antd/locale/zh_CN";
import dayjs from "dayjs";
import "dayjs/locale/zh-cn";
import "antd/dist/reset.css";
import App from "./App";
import "./styles.css";

dayjs.locale("zh-cn");

createRoot(document.getElementById("root")!).render(
  <ConfigProvider
    locale={zhCN}
    theme={{
      token: {
        colorPrimary: "#1760e8",
        borderRadius: 5,
        fontSize: 13,
        colorBgLayout: "#f5f7fa",
        colorBorderSecondary: "#e8ecf2",
        fontFamily:
          '-apple-system, BlinkMacSystemFont, "Segoe UI", "Microsoft YaHei", sans-serif',
      },
      components: {
        Table: {
          headerBg: "#f7f9fc",
          headerColor: "#637188",
          cellPaddingBlock: 20,
        },
        Menu: { itemSelectedBg: "#edf4ff", itemSelectedColor: "#1760e8" },
      },
    }}
  >
    <AntApp>
      <App />
    </AntApp>
  </ConfigProvider>,
);
