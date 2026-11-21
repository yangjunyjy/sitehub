import { createApp } from 'vue'
import { ElDialog, ElIcon } from 'element-plus'
import 'element-plus/es/components/dialog/style/css'
import 'element-plus/es/components/icon/style/css'
import App from './App.vue'
import './console.css'
createApp(App).use(ElDialog).use(ElIcon).mount('#app')
