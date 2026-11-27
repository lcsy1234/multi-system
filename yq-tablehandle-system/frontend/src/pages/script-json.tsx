import { getMessage, go } from '@pangu/plugin-sdk';
import { JsonConverter } from '../JsonConverter';

export default function ScriptJson() {
  return <JsonConverter local onOptimize={sample => {
    if (!go('ai-script', {}, { state: { sample } })) getMessage().warning('未找到 AI 优化脚本菜单，请检查菜单配置或访问权限');
  }} />;
}
