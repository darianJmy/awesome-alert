import { defineStore } from 'pinia'
import { getPublicSettings } from '../api/systemSetting'

const themeStorageKey = 'aiops_theme_preference'
let systemThemeListenerBound = false

function savedThemePreference() {
  const value = localStorage.getItem(themeStorageKey)
  return ['light', 'dark', 'system'].includes(value) ? value : ''
}

export const usePlatformStore = defineStore('platform', {
  state: () => ({ systemName:'观澜运维平台', systemSubtitle:'智能运维控制台', logoData:'', defaultTheme:'system', themePreference:savedThemePreference(), localEnabled:true, ldapEnabled:false, oidcEnabled:false, loaded:false }),
  actions: {
    async load() {
      try {
        const data=(await getPublicSettings()).data||{}
        this.systemName=data.system_name||'观澜运维平台'; this.systemSubtitle=data.system_subtitle||''; this.logoData=data.logo_data||''; this.defaultTheme=data.default_theme||'system'; document.title=this.systemName; this.applyFavicon()
        this.localEnabled=data.local_enabled!==false; this.ldapEnabled=Boolean(data.ldap_enabled); this.oidcEnabled=Boolean(data.oidc_enabled); this.bindSystemTheme(); this.applyTheme()
      } finally { this.loaded=true }
    },
    applyFavicon() {
      const icon = document.querySelector('link[rel="icon"]')
      if (!icon) return
      icon.href = this.logoData || '/favicon.svg'
      if (this.logoData) icon.removeAttribute('type')
      else icon.type = 'image/svg+xml'
    },
    applyTheme() {
      const theme=this.themePreference||this.defaultTheme
      const dark=theme==='dark'||(theme==='system'&&window.matchMedia('(prefers-color-scheme: dark)').matches)
      document.documentElement.classList.toggle('dark',dark); document.documentElement.style.colorScheme=dark?'dark':'light'
    },
    setThemePreference(theme) {
      if (!['light', 'dark', 'system'].includes(theme)) return
      this.themePreference=theme; localStorage.setItem(themeStorageKey,theme); this.applyTheme()
    },
    bindSystemTheme() {
      if (systemThemeListenerBound) return
      window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change',()=>this.applyTheme())
      systemThemeListenerBound=true
    },
    setSettings(data) { this.systemName=data.system_name; this.systemSubtitle=data.system_subtitle; this.logoData=data.logo_data||''; this.defaultTheme=data.default_theme; document.title=this.systemName; this.applyFavicon(); this.applyTheme() },
  },
})
