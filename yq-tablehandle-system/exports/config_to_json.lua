-- 配置转 JSON，适用于 Lua 5.4。一次性安装：luarocks install lunajson；luarocks install luaexpat
-- 命令行：lua config_to_json.lua config.ini > result.json（省略文件名则读取标准输入）
-- 模块：local convert = require('config_to_json').convert；返回 JSON 文本，失败返回 nil 和中文错误。
local M = {}
local function fail(message) error(message, 0) end
local function dependency(name, rock)
  local ok, lib = pcall(require, name)
  if not ok then fail('无法加载依赖，请运行 luarocks install ' .. rock) end
  return lib
end
local function trim(s) return s:match('^%s*(.-)%s*$') end
local function scalar(s)
  if s == 'true' then return true end
  if s == 'false' then return false end
  if (s:match('^%-?%d+$') or s:match('^%-?%d+%.%d+$')) and not s:match('^%-?0%d') then
    local n = tonumber(s)
    if n and n >= -9007199254740991 and n <= 9007199254740991 then return n end
  end
  return s
end

-- 只在同一个对象内部归并重名键。字符串键 "0" 和数组长度标记 [0] 不冲突。
local function add(groups, key, value)
  local group = groups[key] or {}
  group[#group + 1] = value
  groups[key] = group
end
local function object(groups)
  local result = {}
  for k, values in pairs(groups) do
    if #values == 1 then result[k] = values[1]
    else values[0] = #values; result[k] = values end
  end
  return result
end
local function comment(s) return s:match('^[#;]') or s:sub(1,2) == '//' end
local function section(s)
  local name, tail = s:match('^%[([^%]]+)%]%s*(.*)$')
  if name and (tail == '' or comment(tail)) then return trim(name) end
end
local function parse_settings(source)
  local root, sections = {}, {}
  local current, number = root, 0
  for raw in (source .. '\n'):gmatch('(.-)\n') do
    number = number + 1
    local line = trim(raw)
    if line ~= '' and not comment(line) then
      local name = section(line)
      if name then
        sections[name] = sections[name] or {}; current = sections[name]
      else
        local key, value = line:match('^([^=:%s]+)%s*[=:]%s*(.*)$')
        if not key then fail(('第 %d 行缺少有效的键值对'):format(number)) end
        local quote = value:sub(1,1)
        if quote == '"' or quote == "'" then
          local body, tail = value:match('^' .. quote .. '(.*)' .. quote .. '%s*(.*)$')
          if not body or (tail ~= '' and not comment(tail)) then fail(('第 %d 行的引号不完整'):format(number)) end
          add(current, key, body)
        else
          local ending = #value + 1
          for _, marker in ipairs({'%s[#;]', '%s//'}) do ending = math.min(ending, value:find(marker) or ending) end
          if comment(value) then ending = 1 end
          add(current, key, scalar(trim(value:sub(1,ending-1))))
        end
      end
    end
  end
  for name, fields in pairs(sections) do add(root, name, object(fields)) end
  if next(root) == nil then fail('未找到键值对') end
  return object(root)
end

local function parse_xml(source)
  local lxp = dependency('lxp', 'luaexpat')
  local stack, root = {}, {}
  local parser = lxp.new({
    StartDoctypeDecl = function() fail('不支持 XML 文档类型声明，请使用普通配置 XML') end,
    StartElement = function(_, name, attrs)
      if #stack >= 256 then fail('XML 嵌套超过 256 层') end
      stack[#stack+1] = {name=name, attrs=attrs, fields={}, text={}, children=0}
    end,
    CharacterData = function(_, text)
      local node = stack[#stack]; if node then table.insert(node.text, text) end
    end,
    EndElement = function()
      local node = table.remove(stack)
      local attrs, text = node.attrs, trim(table.concat(node.text))
      local key, value = node.name, nil
      if #attrs == 0 and node.children == 0 then value = scalar(text)
      elseif node.children == 0 and text == '' and #attrs == 1 and attrs.value then value = scalar(attrs.value)
      elseif #stack > 0 and key == 'param' and #attrs == 2 and attrs.name and attrs.value and node.children == 0 and text == '' then
        key, value = attrs.name, scalar(attrs.value)
      else
        local fields = {}
        for _, attr in ipairs(attrs) do add(fields, attr, scalar(attrs[attr])) end
        for k, values in pairs(node.fields) do for _, v in ipairs(values) do add(fields,k,v) end end
        if text ~= '' then add(fields, '#text', scalar(text)) end
        value = object(fields)
      end
      local parent = stack[#stack]
      if parent then parent.children = parent.children + 1; add(parent.fields,key,value) else root[key] = value end
    end,
  })
  local ok, result = pcall(function()
    local parsed, reason, line, col = parser:parse(source)
    if parsed then parsed, reason, line, col = parser:parse() end
    if not parsed then
      local hint = '请检查标签、属性引号和特殊字符转义'
      if reason == 'mismatched tag' then hint = '起始标签和结束标签不匹配'
      elseif stack[#stack] then hint = '请检查标签 <' .. stack[#stack].name .. '> 是否完整闭合' end
      fail(('XML 格式错误（第 %d 行，第 %d 列）：%s'):format(line, col, hint))
    end
  end)
  pcall(parser.close, parser)
  if not ok then fail(result) end
  return root
end

-- 扫描字符串及转义字符，避免误删 URL、路径或字面量中的注释标记。
local function strip_json_comments(s)
  local chunks, start, i, quoted = {}, 1, 1, false
  while i <= #s do
    local ch, token = s:sub(i,i), s:sub(i,i+1)
    if quoted and ch == '\\' then i = i + 2
    elseif ch == '"' then quoted = not quoted; i = i + 1
    elseif not quoted and (token == '//' or token == '/*') then
      chunks[#chunks+1] = s:sub(start,i-1) .. ' '
      if token == '//' then i = s:find('[\r\n]',i+2) or (#s+1)
      else local stop = s:find('*/',i+2,true); if not stop then fail('JSON 块注释未闭合') end; i = stop+2 end
      start = i
    else i = i + 1 end
  end
  return table.concat(chunks) .. s:sub(start)
end

local function convert(source)
  if type(source) ~= 'string' then fail('请输入配置文本') end
  if #source > 65536 then fail('输入不能超过 64 KiB') end
  source = trim((source:gsub('^\239\187\191','')))
  local json = dependency('lunajson', 'lunajson')
  if source:match('^<') or source:match('^%?xml%s') then
    return json.encode(parse_xml((source:gsub('^%?xml%s','<?xml '))))
  end
  if source:match('^[%[{/]') then
    -- INI 的 [分组] 与 JSON 数组都可以以 [ 开头，先尝试完整 JSON。
    local ok, clean = pcall(strip_json_comments, source)
    if ok and trim(clean):match('^[%[{]') then
      if pcall(json.decode, clean) then return trim(clean) end
      local ini = false
      for line in (source .. '\n'):gmatch('(.-)\n') do if section(trim(line)) then ini = true; break end end
      if not ini then fail('JSON 格式错误：请检查双引号、逗号及括号是否完整') end
    elseif not ok and not section(trim(source:match('^[^\r\n]*'))) then fail(clean) end
  end
  return json.encode(parse_settings(source))
end
function M.convert(source)
  local ok, result = pcall(convert, source)
  if ok then return result end
  return nil, result
end

if ... == 'config_to_json' then return M end
local file = io.stdin
if arg[1] and arg[1] ~= '-' then
  file = io.open(arg[1], 'rb')
  if not file then io.stderr:write('无法读取输入文件，请检查路径和权限\n'); os.exit(1) end
end
local source = file:read(65537) or ''
if file ~= io.stdin then file:close() end
local output, err = M.convert(source)
if not output then io.stderr:write(err, '\n'); os.exit(1) end
io.write(output, '\n')
