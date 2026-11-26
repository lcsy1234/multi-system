import { xmlError, jsonError } from './errorMessages';

function scalar(text: string): unknown {
  if (text === 'true' || text === 'false') return text === 'true';
  if (/^-?(0|[1-9]\d*)(\.\d+)?$/.test(text) && Number.isFinite(Number(text)) && Math.abs(Number(text)) <= Number.MAX_SAFE_INTEGER) return Number(text);
  return text;
}

// Group repeated keys without overwriting earlier values or inheriting prototype keys.
function pairs(entries: [string, unknown][]) {
  const groups = new Map<string, unknown[]>();
  for (const [key, value] of entries) {
    const group = groups.get(key);
    if (group) group.push(value);
    else groups.set(key, [value]);
  }
  return Object.fromEntries([...groups].map(([key, values]) => [key, values.length === 1 ? values[0] : values]));
}

function xmlValue(node: Element): unknown {
  const children = Array.from(node.children);
  const attrs = Array.from(node.attributes);
  const text = Array.from(node.childNodes).filter(n => n.nodeType === Node.TEXT_NODE || n.nodeType === Node.CDATA_SECTION_NODE).map(n => n.textContent).join('').trim();
  if (!children.length && !attrs.length) return scalar(text);
  if (!children.length && !text && attrs.length === 1 && node.hasAttribute('value')) return scalar(node.getAttribute('value')!);
  const entries: [string, unknown][] = attrs.map(a => [a.name, scalar(a.value)]);
  for (const child of children) {
    if (child.tagName === 'param' && child.attributes.length === 2 && child.hasAttribute('name') && child.hasAttribute('value') && !child.children.length && !child.textContent?.trim()) {
      entries.push([child.getAttribute('name')!, scalar(child.getAttribute('value')!)]);
    } else entries.push([child.tagName, xmlValue(child)]);
  }
  if (text) entries.push(['#text', scalar(text)]);
  return pairs(entries);
}

function settings(source: string) {
  const entries: [string, unknown][] = [];
  const sections = new Map<string, [string, unknown][]>();
  let current = entries;
  for (const [index, raw] of source.split(/\r?\n/).entries()) {
    const line = raw.trim();
    if (!line || /^(#|;|\/\/)/.test(line)) continue;
    const section = /^\[([^\]\r\n]+)\]\s*(?:(?:#|;|\/\/).*)?$/.exec(line);
    if (section) {
      const name = section[1].trim();
      current = sections.get(name) ?? [];
      sections.set(name, current);
      continue;
    }
    const match = /^([^=:\s]+)\s*[=:]\s*(.*)$/.exec(line);
    if (!match) throw Error(`第 ${index + 1} 行缺少有效的键值对`);
    const value = match[2].trim();
    const quote = value[0];
    if (quote === '"' || quote === "'") {
      const last = value.match(quote === '"' ? /^"(.*)"\s*(?:(?:#|;|\/\/).*)?$/ : /^'(.*)'\s*(?:(?:#|;|\/\/).*)?$/);
      if (!last) throw Error(`第 ${index + 1} 行的引号不完整`);
      current.push([match[1], last[1]]);
    } else current.push([match[1], scalar(value.replace(/(?:^|\s+)(?:#|;|\/\/).*$/, '').trim())]);
  }
  for (const [name, fields] of sections) entries.push([name, pairs(fields)]);
  if (!entries.length) throw Error('未找到键值对');
  return pairs(entries);
}

export function parseConfig(input: string) {
  const source = input.trim();
  if (/^(<|\?xml\s)/.test(source)) {
    const doc = new DOMParser().parseFromString(source.replace(/^\?xml\s/, '<?xml '), 'application/xml');
    const error = doc.getElementsByTagNameNS('*', 'parsererror')[0];
    if (error) throw Error(xmlError(error.textContent ?? '', doc.documentElement.tagName));
    return { [doc.documentElement.tagName]: xmlValue(doc.documentElement) };
  }
  // Tokenize strings before comments so URLs and comment markers inside strings survive.
  const clean = source.replace(/"(?:\\.|[^"\\])*"|\/\/[^\r\n]*|\/\*[\s\S]*?\*\//g, token => token.startsWith('/') ? ' ' : token).trim();
  if (/^[{\[]/.test(clean)) {
    try { return JSON.parse(clean); }
    catch (error) {
      if (!/^\s*\[[^\]\r\n]+\]\s*(?:[#;]|\/\/|$)/m.test(source)) throw Error(jsonError(error));
    }
  }
  return settings(source);
}
