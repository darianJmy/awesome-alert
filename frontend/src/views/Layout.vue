<template>
  <el-container class="layout" :style="{ '--sidebar-width': isCollapse ? '72px' : '220px' }">
    <el-aside :width="isCollapse ? '72px' : '220px'" class="sidebar">
      <router-link class="logo" :class="{ 'is-collapsed': isCollapse }" to="/dashboard" aria-label="返回首页">
        <img v-if="platform.logoData" class="logo-image" :src="platform.logoData" alt="" />
        <div v-else class="logo-mark">枢</div>
      </router-link>
      <div class="sidebar-menu">
        <el-menu :default-active="activeMenu" router :collapse="isCollapse">
          <template v-for="menu in menus" :key="menu.id">
            <el-sub-menu v-if="menu.children && menu.children.length" :index="String(menu.id)">
              <template #title>
                <el-icon><component :is="menu.icon" /></el-icon>
                <span>{{ menu.name }}</span>
              </template>
              <template v-for="child in menu.children" :key="child.id">
                <el-menu-item v-if="child.type === 2" :index="child.path">
                  <span>{{ child.name }}</span>
                </el-menu-item>
              </template>
            </el-sub-menu>
            <el-menu-item v-else :index="menu.path">
              <el-icon><component :is="menu.icon" /></el-icon>
              <span>{{ menu.name }}</span>
            </el-menu-item>
          </template>
        </el-menu>
      </div>
      <div class="sidebar-footer">
        <button class="collapse-btn" type="button" @click="isCollapse = !isCollapse" aria-label="切换导航">
          <ArrowLeftBold v-if="!isCollapse" />
          <ArrowRightBold v-else />
        </button>
      </div>
    </el-aside>

    <el-container>
      <el-header class="header">
        <div class="header-left">
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/dashboard' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item v-if="currentTitle">{{ currentTitle }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="header-right">
          <el-dropdown trigger="click" placement="bottom-end" @command="handleThemeCommand">
            <button type="button" class="theme-trigger" aria-label="切换主题" title="主题">
              <el-icon><component :is="themeIcon" /></el-icon>
            </button>
            <template #dropdown>
              <el-dropdown-menu class="theme-menu">
                <el-dropdown-item command="light">
                  <el-icon><Sunny /></el-icon><span>浅色</span><el-icon v-if="activeTheme === 'light'" class="theme-check"><Check /></el-icon>
                </el-dropdown-item>
                <el-dropdown-item command="dark">
                  <el-icon><Moon /></el-icon><span>深色</span><el-icon v-if="activeTheme === 'dark'" class="theme-check"><Check /></el-icon>
                </el-dropdown-item>
                <el-dropdown-item command="system">
                  <el-icon><Monitor /></el-icon><span>跟随系统</span><el-icon v-if="activeTheme === 'system'" class="theme-check"><Check /></el-icon>
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <el-dropdown @command="handleCommand">
            <span class="user-info">
              <span class="user-avatar">{{ userInitial }}</span>
              {{ store.userInfo?.nickname || store.userInfo?.username }}
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">个人信息</el-dropdown-item>
                <el-dropdown-item command="password">修改密码</el-dropdown-item>
                <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </el-container>

  <AIChatAssistant />

  <!-- 修改密码弹窗 -->
  <el-dialog v-model="passwordVisible" title="修改密码" width="400px" :close-on-click-modal="false">
    <el-form ref="pwdFormRef" :model="pwdForm" :rules="pwdRules" label-width="80px">
      <el-form-item label="原密码" prop="old_password">
        <el-input v-model="pwdForm.old_password" type="password" show-password />
      </el-form-item>
      <el-form-item label="新密码" prop="new_password">
        <el-input v-model="pwdForm.new_password" type="password" show-password />
      </el-form-item>
      <el-form-item label="确认密码" prop="confirm_password">
        <el-input v-model="pwdForm.confirm_password" type="password" show-password />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="passwordVisible = false">取消</el-button>
      <el-button type="primary" @click="handleChangePassword">确认</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import { usePlatformStore } from '../stores/platform'
import { changePassword } from '../api/auth'
import { ElMessage } from 'element-plus'
import AIChatAssistant from '../components/AIChatAssistant.vue'

const route = useRoute()
const router = useRouter()
const store = useUserStore()
const platform = usePlatformStore()

const isCollapse = ref(false)
const menus = computed(() => store.menus)
const currentTitle = computed(() => route.meta.title)
const activeMenu = computed(() => route.meta.activeMenu || route.path)
const activeTheme = computed(() => platform.themePreference || platform.defaultTheme)
const themeIcon = computed(() => ({ light: 'Sunny', dark: 'Moon', system: 'Monitor' }[activeTheme.value] || 'Monitor'))
const userInitial = computed(() => {
  const name = store.userInfo?.nickname || store.userInfo?.username || 'U'
  return String(name).slice(0, 1).toUpperCase()
})

const passwordVisible = ref(false)
const pwdFormRef = ref()
const pwdForm = ref({ old_password: '', new_password: '', confirm_password: '' })

const validateConfirm = (rule, value, callback) => {
  if (value !== pwdForm.value.new_password) {
    callback(new Error('两次密码不一致'))
  } else {
    callback()
  }
}

const pwdRules = {
  old_password: [{ required: true, message: '请输入原密码', trigger: 'blur' }],
  new_password: [{ required: true, min: 6, message: '密码至少6位', trigger: 'blur' }],
  confirm_password: [{ required: true, validator: validateConfirm, trigger: 'blur' }],
}

function handleCommand(cmd) {
  if (cmd === 'logout') {
    store.logout()
    router.push('/login')
  } else if (cmd === 'password') {
    pwdForm.value = { old_password: '', new_password: '', confirm_password: '' }
    passwordVisible.value = true
  }
}

function handleThemeCommand(theme) {
  platform.setThemePreference(theme)
}

async function handleChangePassword() {
  const valid = await pwdFormRef.value.validate().catch(() => false)
  if (!valid) return
  try {
    await changePassword({
      old_password: pwdForm.value.old_password,
      new_password: pwdForm.value.new_password,
    })
    ElMessage.success('密码修改成功')
    passwordVisible.value = false
  } catch {}
}
</script>

<style scoped>
.layout {
  height: 100vh;
  background: var(--ops-bg);
}

.sidebar {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  overflow-x: hidden;
  background: var(--ops-surface);
  border-right: 1px solid var(--ops-border);
  box-shadow: 8px 0 28px rgba(16, 24, 40, 0.04);
}

.logo {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 68px;
  padding: 0 18px;
  color: inherit;
  text-decoration: none;
  background: var(--ops-surface);
  border-bottom: 1px solid var(--ops-border-soft);
  cursor: pointer;
  transition: background-color 0.18s ease;
}

.logo:hover {
  background: var(--ops-primary-soft);
}

.logo:focus-visible {
  outline: 2px solid var(--ops-primary);
  outline-offset: -3px;
}

.logo-mark {
  display: grid;
  place-items: center;
  width: 38px;
  height: 38px;
  flex: 0 0 38px;
  border-radius: 8px;
  color: #fff;
  font-size: 15px;
  font-weight: 800;
  background: linear-gradient(135deg, var(--ops-primary), var(--ops-accent));
}
.logo-image { width: min(154px, 100%); height: 36px; object-fit: contain; }
.logo.is-collapsed { padding: 0 10px; }
.logo.is-collapsed .logo-image { width: 36px; height: 36px; object-fit: cover; object-position: left; }

.sidebar-menu {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  overflow-x: hidden;
}

.sidebar :deep(.el-menu) {
  border-right: 0;
  padding: 10px;
}

.sidebar :deep(.el-menu--collapse) {
  width: 72px;
}

.sidebar :deep(.el-menu-item),
.sidebar :deep(.el-sub-menu__title) {
  height: 42px;
  margin: 4px 0;
  border-radius: 8px;
  color: var(--ops-muted);
  font-weight: 650;
}

.sidebar :deep(.el-menu-item:hover),
.sidebar :deep(.el-sub-menu__title:hover) {
  color: var(--ops-primary);
  background: var(--ops-primary-soft);
}

.sidebar :deep(.el-menu-item.is-active) {
  color: var(--ops-primary);
  background: var(--ops-primary-soft);
}

.sidebar-footer {
  flex: 0 0 auto;
  display: flex;
  justify-content: center;
  padding: 8px 0 10px;
  border-top: 1px solid var(--ops-border-soft);
}

.sidebar-footer .collapse-btn {
  width: 30px;
  height: 30px;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 64px;
  padding: 0 24px;
  background: color-mix(in srgb, var(--ops-surface) 86%, transparent);
  border-bottom: 1px solid var(--ops-border);
  backdrop-filter: blur(12px);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 14px;
  min-width: 0;
}

.collapse-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  padding: 0;
  color: #98a2b3;
  cursor: pointer;
  background: transparent;
  border: 0;
  border-radius: 8px;
  transition: color 0.16s ease, background 0.16s ease;
}

.collapse-btn svg {
  width: 13px;
  height: 13px;
}

.collapse-btn:hover {
  color: var(--ops-primary);
  background: var(--ops-primary-soft);
}

.collapse-btn:active {
  transform: translateY(1px);
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.theme-trigger {
  display: inline-grid;
  width: 32px;
  height: 32px;
  place-items: center;
  padding: 0;
  color: var(--ops-muted);
  background: transparent;
  border: 0;
  border-radius: 8px;
  cursor: pointer;
  font-size: 17px;
  transition: color .16s ease, background-color .16s ease;
}

.theme-trigger:hover,
.theme-trigger:focus-visible {
  color: var(--ops-primary);
  background: var(--ops-primary-soft);
}

.theme-trigger:focus-visible {
  outline: 2px solid var(--ops-primary);
  outline-offset: 2px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--ops-text);
  cursor: pointer;
  font-weight: 650;
}

.user-avatar {
  display: inline-grid;
  place-items: center;
  width: 30px;
  height: 30px;
  color: #fff;
  background: #172033;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 800;
}

.el-main {
  min-width: 0;
  padding: 22px 24px;
  background: var(--ops-bg);
}

</style>

<style>
.theme-menu .el-dropdown-menu__item {
  display: grid;
  min-width: 148px;
  grid-template-columns: 20px 1fr 18px;
  gap: 8px;
}

.theme-menu .theme-check {
  color: var(--ops-primary);
}
</style>
