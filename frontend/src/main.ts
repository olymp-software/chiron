import { createApp } from 'vue'
import App from './App.vue'
import { i18n } from './i18n'
import { createAppRouter, installBackInput } from './router'
import './style.css'

const router = createAppRouter()
installBackInput(router)

createApp(App).use(i18n).use(router).mount('#app')
