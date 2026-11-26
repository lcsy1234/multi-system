var ConfigParser = (() => {
  var __defProp = Object.defineProperty;
  var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
  var __getOwnPropNames = Object.getOwnPropertyNames;
  var __hasOwnProp = Object.prototype.hasOwnProperty;
  var __export = (target, all) => {
    for (var name in all)
      __defProp(target, name, { get: all[name], enumerable: true });
  };
  var __copyProps = (to, from, except, desc) => {
    if (from && typeof from === "object" || typeof from === "function") {
      for (let key of __getOwnPropNames(from))
        if (!__hasOwnProp.call(to, key) && key !== except)
          __defProp(to, key, { get: () => from[key], enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable });
    }
    return to;
  };
  var __toCommonJS = (mod) => __copyProps(__defProp({}, "__esModule", { value: true }), mod);

  // scriptJson.ts
  var scriptJson_exports = {};
  __export(scriptJson_exports, {
    parseConfig: () => parseConfig
  });

  // errorMessages.ts
  function message(error) {
    return error instanceof Error ? error.message : String(error ?? "");
  }
  function position(text) {
    const location = /line(?: number)?\s*[:：]?\s*(\d+)[\s,:]*(?:at\s+)?column\s*[:：]?\s*(\d+)/i.exec(text);
    return location ? `\uFF08\u7B2C ${location[1]} \u884C\uFF0C\u7B2C ${location[2]} \u5217\uFF09` : "";
  }
  function xmlError(detail, root) {
    const lead = `XML \u683C\u5F0F\u9519\u8BEF${position(detail)}\uFF1A`;
    const open = /premature end of data in tag\s+([^\s<>]+)/i.exec(detail);
    if (open) return `${lead}\u6807\u7B7E <${open[1]}> \u5C1A\u672A\u95ED\u5408\uFF0C\u8BF7\u8865\u4E0A </${open[1]}>\u3002`;
    if (/still inside the root element/i.test(detail) && root && root !== "parsererror") return `${lead}\u6587\u6863\u63D0\u524D\u7ED3\u675F\uFF0C\u8BF7\u68C0\u67E5\u672A\u95ED\u5408\u7684\u6807\u7B7E\uFF0C\u5E76\u8865\u4E0A </${root}>\u3002`;
    if (/mismatch|unexpected closing tag/i.test(detail)) return lead + "\u8D77\u59CB\u6807\u7B7E\u4E0E\u7ED3\u675F\u6807\u7B7E\u4E0D\u5339\u914D\uFF0C\u8BF7\u6838\u5BF9\u6807\u7B7E\u540D\u79F0\u548C\u5D4C\u5957\u987A\u5E8F\u3002";
    if (/extra content|multiple root|junk after document/i.test(detail)) return lead + "\u6839\u6807\u7B7E\u7ED3\u675F\u540E\u8FD8\u6709\u5185\u5BB9\uFF0C\u8BF7\u68C0\u67E5\u662F\u5426\u6709\u591A\u4E2A\u6839\u6807\u7B7E\u3002";
    if (/entityref|entity reference|entity .*not defined/i.test(detail)) return lead + "\u5B9E\u4F53\u5F15\u7528\u65E0\u6548\uFF0C\u8BF7\u68C0\u67E5 & \u7B49\u7279\u6B8A\u5B57\u7B26\u7684\u8F6C\u4E49\u3002";
    return lead + "\u8BF7\u68C0\u67E5\u6807\u7B7E\u662F\u5426\u6210\u5BF9\u3001\u5C5E\u6027\u5F15\u53F7\u662F\u5426\u5B8C\u6574\u3002";
  }
  function jsonError(error) {
    const text = message(error);
    return `JSON \u683C\u5F0F\u9519\u8BEF${position(text)}\uFF1A\u8BF7\u68C0\u67E5\u53CC\u5F15\u53F7\u3001\u9017\u53F7\u53CA\u62EC\u53F7\u662F\u5426\u5B8C\u6574\u3002`;
  }

  // scriptJson.ts
  function scalar(text) {
    if (text === "true" || text === "false") return text === "true";
    if (/^-?(0|[1-9]\d*)(\.\d+)?$/.test(text) && Number.isFinite(Number(text)) && Math.abs(Number(text)) <= Number.MAX_SAFE_INTEGER) return Number(text);
    return text;
  }
  function pairs(entries) {
    const groups = /* @__PURE__ */ new Map();
    for (const [key, value] of entries) {
      const group = groups.get(key);
      if (group) group.push(value);
      else groups.set(key, [value]);
    }
    return Object.fromEntries([...groups].map(([key, values]) => [key, values.length === 1 ? values[0] : values]));
  }
  function xmlValue(node) {
    const children = Array.from(node.children);
    const attrs = Array.from(node.attributes);
    const text = Array.from(node.childNodes).filter((n) => n.nodeType === Node.TEXT_NODE || n.nodeType === Node.CDATA_SECTION_NODE).map((n) => n.textContent).join("").trim();
    if (!children.length && !attrs.length) return scalar(text);
    if (!children.length && !text && attrs.length === 1 && node.hasAttribute("value")) return scalar(node.getAttribute("value"));
    const entries = attrs.map((a) => [a.name, scalar(a.value)]);
    for (const child of children) {
      if (child.tagName === "param" && child.attributes.length === 2 && child.hasAttribute("name") && child.hasAttribute("value") && !child.children.length && !child.textContent?.trim()) {
        entries.push([child.getAttribute("name"), scalar(child.getAttribute("value"))]);
      } else entries.push([child.tagName, xmlValue(child)]);
    }
    if (text) entries.push(["#text", scalar(text)]);
    return pairs(entries);
  }
  function settings(source) {
    const entries = [];
    const sections = /* @__PURE__ */ new Map();
    let current = entries;
    for (const [index, raw] of source.split(/\r?\n/).entries()) {
      const line = raw.trim();
      if (!line || /^(#|;|\/\/)/.test(line)) continue;
      const section = /^\[([^\]\r\n]+)\]/.exec(line);
      if (section) {
        const name = section[1].trim();
        current = sections.get(name) ?? [];
        sections.set(name, current);
        continue;
      }
      const match = /^([^=:\s]+)\s*[=:]\s*(.*)$/.exec(line);
      if (!match) throw Error(`\u7B2C ${index + 1} \u884C\u7F3A\u5C11\u6709\u6548\u7684\u952E\u503C\u5BF9`);
      const value = match[2].trim();
      const quote = value[0];
      if (quote === '"' || quote === "'") {
        const last = value.match(quote === '"' ? /^"(.*)"\s*(?:(?:#|;|\/\/).*)?$/ : /^'(.*)'\s*(?:(?:#|;|\/\/).*)?$/);
        if (!last) throw Error(`\u7B2C ${index + 1} \u884C\u7684\u5F15\u53F7\u4E0D\u5B8C\u6574`);
        current.push([match[1], last[1]]);
      } else current.push([match[1], scalar(value.replace(/(?:^|\s+)(?:#|;|\/\/).*$/, "").trim())]);
    }
    for (const [name, fields] of sections) entries.push([name, pairs(fields)]);
    if (!entries.length) throw Error("\u672A\u627E\u5230\u952E\u503C\u5BF9");
    return pairs(entries);
  }
  function parseConfig(input) {
    const source = input.trim();
    if (/^(<|\?xml\s)/.test(source)) {
      const doc = new DOMParser().parseFromString(source.replace(/^\?xml\s/, "<?xml "), "application/xml");
      const error = doc.getElementsByTagNameNS("*", "parsererror")[0];
      if (error) throw Error(xmlError(error.textContent ?? "", doc.documentElement.tagName));
      return { [doc.documentElement.tagName]: xmlValue(doc.documentElement) };
    }
    const clean = source.replace(/"(?:\\.|[^"\\])*"|\/\/[^\r\n]*|\/\*[\s\S]*?\*\//g, (token) => token.startsWith("/") ? " " : token).trim();
    if (/^[{\[]/.test(clean)) {
      try {
        return JSON.parse(clean);
      } catch (error) {
        if (!/^\s*\[[^\]\r\n]+\][^\r\n]*\r?\n\s*[^\s=:\[\]]+\s*[=:]/m.test(source) && !/^\s*\[[^\]\r\n]+\]\s*(?:[#;]|\/\/|$)/m.test(source)) throw Error(jsonError(error));
      }
    }
    return settings(source);
  }
  return __toCommonJS(scriptJson_exports);
})();
