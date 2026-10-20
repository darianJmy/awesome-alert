<template>
  <main class="login-page">
    <section class="visual-panel" aria-labelledby="platform-title">
      <div class="visual-grid" aria-hidden="true"></div>
      <header class="brand-lockup">
        <img v-if="platform.logoData" :src="platform.logoData" class="brand-image" alt="" />
        <span v-else class="brand-mark">{{ platform.systemName.slice(0, 1) }}</span>
        <div><strong>{{ platform.systemName }}</strong><small>{{ platform.systemSubtitle }}</small></div>
      </header>

      <div class="hero-stage">
        <div class="visual-copy">
          <span class="eyebrow"><i></i> AI 驱动的智能运维</span>
          <h1 id="platform-title">让复杂系统<br><em>始终清晰可控</em></h1>
          <p>让 AI 协同监控、拨测与自动化执行，在一个工作台中发现异常、分析根因并给出处置建议。</p>
        </div>
      </div>

      <div class="visual-footer">
        <div><strong>统一接入</strong><span>多源数据与节点</span></div>
        <div><strong>AI 诊断</strong><span>关联分析与根因定位</span></div>
        <div><strong>闭环处置</strong><span>自动化执行追踪</span></div>
      </div>
    </section>

    <section class="login-panel" aria-labelledby="login-title">
      <div class="login-content">
        <div class="login-heading">
          <span v-if="authType !== 'local'" class="auth-badge">{{ authType === 'oidc' ? 'OIDC' : 'LDAP' }}</span>
          <h2 id="login-title">{{ loginTitle }}</h2>
          <p class="login-intro">{{ loginDescription }}</p>
        </div>

        <div v-if="authType === 'oidc'" class="external-login">
          <div class="provider-mark oidc" aria-hidden="true">
            <svg viewBox="0 0 20 20">
              <path class="oidc-orange" d="M9,16.979V1.985L9,.104l3,1.412V18.445Z" />
              <path class="oidc-gray" d="M9,.104V1.985C5.576,2.414,3,4.285,3,6.527c0,2.126,2.316,3.916,5.473,4.462v1.906C3.644,12.311,0,9.682,0,6.527,0,3.258,3.912.554,9,.104ZM19.804,7.73l-.414,4.3-1.543-.869-.217.13a11.928,11.928,0,0,1-5.1,1.6V10.989a9.256,9.256,0,0,0,3.1-1.08L14,8.992Z" />
            </svg>
          </div>
          <el-button type="primary" size="large" :loading="loading" class="external-login-button" @click="startOIDC">
            使用 OIDC 单点登录
            <el-icon class="button-arrow"><Right /></el-icon>
          </el-button>
          <p>将跳转至统一身份认证平台，认证成功后自动返回。</p>
        </div>

        <el-form v-else ref="formRef" :model="form" :rules="rules" label-position="top" size="large" class="login-form">
          <el-form-item :label="authType === 'ldap' ? 'LDAP 用户名' : '用户名'" prop="username">
            <el-input v-model="form.username" :placeholder="authType === 'ldap' ? '请输入 LDAP 用户名' : '请输入系统用户名'" prefix-icon="User" autocomplete="username" />
          </el-form-item>
          <el-form-item label="密码" prop="password">
            <el-input v-model="form.password" type="password" placeholder="请输入密码" prefix-icon="Lock"
              autocomplete="current-password" @keyup.enter="handleLogin" show-password />
          </el-form-item>
          <el-form-item class="submit-item">
            <el-button type="primary" :loading="loading" class="login-button" @click="handleLogin">
              {{ authType === 'ldap' ? '使用 LDAP 登录' : '使用系统账号登录' }}
              <el-icon class="button-arrow"><Right /></el-icon>
            </el-button>
          </el-form-item>
        </el-form>

        <div class="secure-note"><el-icon><Lock /></el-icon><span>登录过程通过安全连接传输，请勿共享账号凭据</span></div>
        <button v-if="externalAuthType" type="button" class="auth-switch" @click="toggleAuthType">
          {{ authType === 'local' ? externalAuthLabel : '使用系统账号登录' }}
        </button>
      </div>
      <footer class="login-footer">{{ platform.systemName }} · 让运维清晰可控</footer>
    </section>
  </main>
</template>

<script setup>
import { computed, nextTick, onMounted, reactive, ref } from 'vue'
import { Lock, Right } from '@element-plus/icons-vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import { usePlatformStore } from '../stores/platform'
import { completeOIDCLogin } from '../api/systemSetting'

const router = useRouter()
const route = useRoute()
const store = useUserStore()
const platform = usePlatformStore()
const formRef = ref()
const loading = ref(false)
const authType = ref('local')

const form = reactive({ username: '', password: '' })
const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

const externalAuthType = computed(() => platform.oidcEnabled ? 'oidc' : platform.ldapEnabled ? 'ldap' : '')
const externalAuthLabel = computed(() => externalAuthType.value === 'oidc' ? '使用 OIDC 单点登录' : '使用 LDAP 账号登录')
const loginTitle = computed(() => authType.value === 'oidc' ? '统一身份认证' : authType.value === 'ldap' ? 'LDAP 登录' : '系统账号登录')
const loginDescription = computed(() => authType.value === 'oidc'
  ? `通过统一身份认证进入 ${platform.systemName}`
  : authType.value === 'ldap'
    ? `使用企业目录账号进入 ${platform.systemName}`
    : '使用平台本地管理员或系统账号登录。')

async function handleLogin() {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  loading.value = true
  try {
    await store.loginAction({ ...form, auth_type: authType.value })
    router.push('/dashboard')
  } finally {
    loading.value = false
  }
}

function startOIDC() {
  loading.value = true
  window.location.href = '/api/auth/oidc/start'
}

function toggleAuthType() {
  authType.value = authType.value === 'local' ? externalAuthType.value : 'local'
  nextTick(() => formRef.value?.clearValidate?.())
}

onMounted(async () => {
  if (!platform.loaded) await platform.load()
  authType.value = externalAuthType.value || 'local'
  if (route.query.oidc !== 'complete') return
  loading.value = true
  try {
    const res = await completeOIDCLogin()
    store.acceptToken(res.data.token)
    await router.replace('/dashboard')
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.login-page {
  --navy-950: #071321; --navy-900: #0a1a2d; --surface: #f7f9fc; --ink: #17243a; --muted: #738097;
  --line: #dfe6ef; --accent: #276ee8; --cyan: #61d8d0; --online: #61d899;
  --text-xs: 12px; --text-sm: 13px; --text-md: 15px; --text-xl: clamp(38px, 4vw, 64px);
  --radius-sm: 8px; --ease-out: cubic-bezier(.2, .8, .2, 1);
  display: grid; grid-template-columns: minmax(600px, 1.08fr) minmax(480px, .92fr); min-height: 100vh; min-height: 100dvh;
  color: var(--ink); background: var(--surface);
}
.visual-panel {
  position: relative; display: flex; min-height: 100vh; padding: 42px 52px 34px; overflow: hidden; flex-direction: column;
  color: #f5f9ff; isolation: isolate; background: radial-gradient(circle at 73% 38%, rgba(42,113,223,.22), transparent 34%), linear-gradient(145deg, var(--navy-950), var(--navy-900) 57%, #0c2239);
}
.visual-grid { position: absolute; z-index: -1; inset: 0; opacity: .28; background-image: linear-gradient(rgba(126,166,211,.08) 1px, transparent 1px), linear-gradient(90deg, rgba(126,166,211,.08) 1px, transparent 1px); background-size: 44px 44px; mask-image: linear-gradient(to bottom right, #000, transparent 78%); }
.brand-lockup { display: flex; align-items: center; gap: 11px; }
.brand-mark { display: grid; width: 38px; height: 38px; place-items: center; color: #081726; font-size: 20px; font-weight: 850; background: var(--cyan); border-radius: 11px 4px 11px 4px; box-shadow: 0 0 30px rgba(97,216,208,.2); }
.brand-image { width: 148px; height: 32px; object-fit: contain; }
.brand-lockup > div { display: flex; flex-direction: column; gap: 1px; }.brand-lockup strong { font-size: 17px; letter-spacing: .04em; }.brand-lockup small { color: #8fa6c0; font-size: 10px; letter-spacing: .12em; }
.hero-stage { display: flex; min-height: 0; align-items: center; flex: 1; }
.visual-copy { position: relative; z-index: 3; width: min(680px,82%); margin-right: auto; margin-left: auto; }
.eyebrow { display: inline-flex; align-items: center; gap: 9px; color: #9bb2cb; font-size: var(--text-xs); letter-spacing: .08em; }.eyebrow i { width: 24px; height: 1px; background: var(--cyan); box-shadow: 0 0 8px var(--cyan); }
.visual-copy h1 { margin: 18px 0; color: #f7fbff; font-size: var(--text-xl); font-weight: 650; line-height: 1.12; letter-spacing: -.035em; }.visual-copy h1 em { color: var(--cyan); font-style: normal; }
.visual-copy p { max-width: 560px; margin: 0; color: #9eb0c5; font-size: var(--text-md); line-height: 1.8; }
.visual-footer { z-index: 3; display: grid; width: 100%; grid-template-columns: repeat(3,1fr); margin-top: auto; border-top: 1px solid rgba(130,164,199,.15); }.visual-footer div { display: flex; padding: 19px 18px 0 0; flex-direction: column; gap: 4px; }.visual-footer strong { color: #d7e6f5; font-size: var(--text-xs); }.visual-footer span { color: #68839e; font-size: 10px; }
.login-panel { display: flex; min-height: 100vh; min-height: 100dvh; padding: 48px clamp(42px,5vw,86px) 30px; flex-direction: column; background: #fbfcfe; }.login-content { width: 100%; max-width: 430px; margin: auto; }
.login-heading { margin-bottom: 26px; }.login-content h2 { margin: 0; color: var(--ink); font-size: 34px; font-weight: 700; letter-spacing: -.03em; }.login-intro { margin: 8px 0 0; color: var(--muted); font-size: var(--text-sm); line-height: 1.7; }.auth-badge { display: inline-flex; margin-bottom: 11px; padding: 4px 8px; color: #245fae; font-size: 10px; font-weight: 750; letter-spacing: .08em; background: #edf5ff; border: 1px solid #d4e6fc; border-radius: 4px; }
.login-form :deep(.el-form-item) { margin-bottom: 22px; }.login-form :deep(.el-form-item__label) { height: auto; padding-bottom: 8px; color: #3d4a5e; font-size: var(--text-sm); font-weight: 650; line-height: 1.4; }
.login-form :deep(.el-input__wrapper) { min-height: 50px; padding: 0 15px; background: #fff; border: 1px solid var(--line); border-radius: var(--radius-sm); box-shadow: none; transition: border-color .2s,box-shadow .2s; }.login-form :deep(.el-input__wrapper:hover) { border-color: #b8c5d7; }.login-form :deep(.el-input__wrapper.is-focus) { border-color: var(--accent); box-shadow: 0 0 0 3px rgba(39,110,232,.11); }
.login-form :deep(.el-input__inner) { color: var(--ink); font-size: 14px; }.login-form :deep(.el-input__prefix) { color: #8997aa; }.submit-item { margin-top: 30px; }
.login-button { width: 100%; min-height: 50px; margin: 0; color: #fff; font-size: 14px; font-weight: 650; background: var(--accent); border-color: var(--accent); border-radius: var(--radius-sm); box-shadow: 0 10px 24px rgba(39,110,232,.18); transition: transform .2s var(--ease-out),box-shadow .2s var(--ease-out); }.login-button:hover { background: #1f62d3; border-color: #1f62d3; box-shadow: 0 13px 28px rgba(39,110,232,.24); transform: translateY(-1px); }.button-arrow { margin-left: 8px; }
.external-login { display: flex; min-height: 252px; padding: 30px 28px 24px; align-items: center; flex-direction: column; background: #fff; border: 1px solid var(--line); border-radius: 12px; }.provider-mark { display: grid; width: 58px; height: 58px; margin-bottom: 25px; place-items: center; background: #f8fafc; border: 1px solid #e3e9f1; border-radius: 14px; box-shadow: 0 8px 22px rgba(16,35,61,.07); }.provider-mark svg { width: 30px; height: 30px; }.oidc-orange { fill: #f8931e; }.oidc-gray { fill: #929aa5; }.external-login-button { width: 100%; min-height: 50px; font-size: 14px; font-weight: 650; border-radius: var(--radius-sm); box-shadow: 0 10px 24px rgba(39,110,232,.18); }.external-login p { margin: 16px 0 0; color: var(--muted); font-size: 11px; line-height: 1.6; text-align: center; }
.secure-note { display: flex; align-items: center; justify-content: center; gap: 7px; margin-top: 22px; color: #94a0b1; font-size: 10px; }.secure-note .el-icon { color: #8090a5; }.auth-switch { display: block; margin: 18px auto 0; padding: 4px 8px; cursor: pointer; color: var(--accent); font: inherit; font-size: 12px; font-weight: 600; background: transparent; border: 0; border-radius: 4px; }.auth-switch:hover { color: #1f62d3; text-decoration: underline; text-underline-offset: 3px; }.auth-switch:focus-visible { outline: 2px solid rgba(39,110,232,.4); outline-offset: 3px; }.login-footer { color: #a2acba; font-size: 10px; letter-spacing: .04em; text-align: center; }
:global(html.dark) .login-panel { background: #111827; }:global(html.dark) .login-content h2 { color: #e5e7eb; }:global(html.dark) .login-intro { color: #a3adbd; }:global(html.dark) .auth-badge { color: #8ec1ff; background: #142a45; border-color: #29496e; }:global(html.dark) .login-form :deep(.el-form-item__label) { color: #cbd5e1; }:global(html.dark) .login-form :deep(.el-input__wrapper) { background: #182230; border-color: #344054; }:global(html.dark) .login-form :deep(.el-input__inner) { color: #e5e7eb; }:global(html.dark) .external-login { background: #182230; border-color: #344054; }:global(html.dark) .provider-mark { background: #111827; border-color: #344054; }
@media (max-width: 1100px) { .login-page { grid-template-columns: minmax(480px,1fr) minmax(400px,.82fr); }.visual-panel { padding-inline: 36px; }.visual-copy { width: 92%; } }
@media (max-width: 860px) { .login-page { display: block; }.visual-panel { min-height: 290px; padding: 28px 32px; }.visual-copy { width: 72%; }.visual-copy h1 { margin: 12px 0; font-size: 34px; }.visual-copy p { font-size: 12px; }.visual-footer { display: none; }.login-panel { min-height: auto; padding: 48px 28px 26px; }.login-content { margin: 0 auto 48px; } }
@media (max-width: 560px) { .visual-panel { min-height: 220px; padding: 24px; }.brand-lockup small { display: none; }.visual-copy { width: 90%; margin: 30px auto 0; }.eyebrow,.visual-copy p { display: none; }.visual-copy h1 { font-size: 30px; line-height: 1.18; }.login-panel { padding: 38px 22px 24px; }.login-content h2 { font-size: 30px; }.login-intro { margin-bottom: 20px; }.secure-note { align-items: flex-start; line-height: 1.5; text-align: center; } }
</style>
