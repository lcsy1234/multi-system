import variant from '@jitl/quickjs-singlefile-browser-release-sync';
import { newQuickJSWASMModuleFromVariant, shouldInterruptAfterDeadline } from 'quickjs-emscripten-core';

self.onmessage = async ({ data }) => {
  try {
    const vm = await newQuickJSWASMModuleFromVariant(variant);
    for (const rule of data.rules) {
      const expression = `JSON.stringify((${rule.code})(${JSON.stringify(data.sample)}))`;
      const json = vm.evalCode(expression, {
        memoryLimitBytes: 8 * 1024 * 1024,
        shouldInterrupt: shouldInterruptAfterDeadline(Date.now() + 300),
      });
      if (json !== undefined) {
        if (typeof json !== 'string' || json.length > 512 * 1024) throw Error('脚本输出无效或过大');
        self.postMessage({ json, rule });
        return;
      }
    }
    self.postMessage({});
  } catch (error) { self.postMessage({ error: String(error) }); }
};
