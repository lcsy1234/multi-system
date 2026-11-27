import { defineManifest } from '@pangu/plugin-sdk';
import { TYPE_LAYOUT } from './config';
import './assets/style/global.less';

export default defineManifest({
  name: 'tablehandle',
  title: '工厂数据工作台',
  routes: [
    { component: 'report', layout: TYPE_LAYOUT.SIDEBAR },
    { component: 'ai-script', layout: TYPE_LAYOUT.SIDEBAR },
    { component: 'script-json', layout: TYPE_LAYOUT.SIDEBAR },
    { component: 'shrimp-json', layout: TYPE_LAYOUT.SIDEBAR },
  ],
});
