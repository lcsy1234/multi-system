function message(error: unknown): string {
  return error instanceof Error ? error.message : String(error ?? '');
}

export function userError(error: unknown, fallback: string): string {
  const text = message(error).replace(/^(?:(?:Type|Syntax|Range|Reference)?Error:\s*)+/i, '').trim();
  if (/failed to fetch|networkerror|network request failed|load failed/i.test(text)) return '无法连接服务，请检查网络和后端服务后重试。';
  if (/interrupted|timed? ?out|deadline exceeded/i.test(text)) return '处理超时，请缩短输入后重试。';
  if (/out of memory|allocation failed/i.test(text)) return '处理所需内存过多，请缩短输入后重试。';
  // Keep Chinese business messages and identifiers, discard raw runtime diagnostics.
  if (/[\u3400-\u9fff]/.test(text) && !/[A-Za-z]{2,}\s+[A-Za-z]{2,}|\bat \w+\(/.test(text) && !/^(?:[A-Za-z]|\[)/.test(text.replace(/^(?:XML|JSON|DeepSeek)\s*/, ''))) return text;
  return fallback;
}

function position(text: string) {
  const location = /line(?: number)?\s*[:：]?\s*(\d+)[\s,:]*(?:at\s+)?column\s*[:：]?\s*(\d+)/i.exec(text);
  return location ? `（第 ${location[1]} 行，第 ${location[2]} 列）` : '';
}

export function xmlError(detail: string, root: string): string {
  const lead = `XML 格式错误${position(detail)}：`;
  const open = /premature end of data in tag\s+([^\s<>]+)/i.exec(detail);
  if (open) return `${lead}标签 <${open[1]}> 尚未闭合，请补上 </${open[1]}>。`;
  if (/still inside the root element/i.test(detail) && root && root !== 'parsererror') return `${lead}文档提前结束，请检查未闭合的标签，并补上 </${root}>。`;
  if (/mismatch|unexpected closing tag/i.test(detail)) return lead + '起始标签与结束标签不匹配，请核对标签名称和嵌套顺序。';
  if (/extra content|multiple root|junk after document/i.test(detail)) return lead + '根标签结束后还有内容，请检查是否有多个根标签。';
  if (/entityref|entity reference|entity .*not defined/i.test(detail)) return lead + '实体引用无效，请检查 & 等特殊字符的转义。';
  return lead + '请检查标签是否成对、属性引号是否完整。';
}

export function jsonError(error: unknown): string {
  const text = message(error);
  return `JSON 格式错误${position(text)}：请检查双引号、逗号及括号是否完整。`;
}
