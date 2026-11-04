import { mount } from 'svelte'
import './app.css'
import App from './App.svelte'

window.addEventListener('error', (e) => {
  document.title = 'ERR: ' + ((e.error && e.error.message) || e.message)
})
window.addEventListener('unhandledrejection', (e) => {
  document.title = 'REJ: ' + String((e as PromiseRejectionEvent).reason)
})

const app = mount(App, {
  target: document.getElementById('app')!,
})

export default app
