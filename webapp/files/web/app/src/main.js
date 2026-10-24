import 'vite/modulepreload-polyfill'
import 'htmx.org'
import { mount } from 'svelte'
import './app.css'
import App from './App.svelte'

const target = document.getElementById('app')
if (target) {
  mount(App, { target })
}
