<template>
  <button
    v-show="!visible"
    type="button"
    class="ai-chat-fab"
    :style="{ zIndex: currentZIndex + 1 }"
    aria-label="打开运维中枢 AI 助手"
    title="AI 助手"
    @click="openChat"
  >
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path d="M12 2a2 2 0 0 1 2 2c0 .74-.4 1.39-1 1.73V7h1a7 7 0 0 1 7 7h1a1 1 0 0 1 1 1v3a1 1 0 0 1-1 1h-1v1a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-1H2a1 1 0 0 1-1-1v-3a1 1 0 0 1 1-1h1a7 7 0 0 1 7-7h1V5.73c-.6-.34-1-.99-1-1.73a2 2 0 0 1 2-2M7.5 13A2.5 2.5 0 1 0 10 15.5 2.5 2.5 0 0 0 7.5 13m9 0a2.5 2.5 0 1 0 2.5 2.5 2.5 2.5 0 0 0-2.5-2.5" />
    </svg>
  </button>

  <el-dialog
    v-model="visible"
    class="ai-chat-dialog"
    width="min(1080px, 94vw)"
    :show-close="false"
    :close-on-click-modal="false"
    :close-on-press-escape="!sending"
    append-to-body
    destroy-on-close
  >
    <div class="chat-shell">
      <aside class="chat-sidebar" :class="{ 'is-collapsed': sidebarCollapsed }">
        <div class="sidebar-header">
          <div class="assistant-brand">
            <span class="assistant-mark"><Cpu /></span>
            <div><strong>运维中枢 AI</strong><small>平台智能助手</small></div>
          </div>
          <el-button size="small" type="primary" :icon="Plus" @click="newConversation">
            <span class="new-chat-copy">新建对话</span>
          </el-button>
        </div>

        <div v-loading="loadingConversations" class="conversation-list">
          <div
            v-for="conversation in conversations"
            :key="conversation.id"
            class="conversation-item"
            :class="{ active: conversation.id === activeConversationID }"
            role="button"
            tabindex="0"
            @click="selectConversation(conversation)"
            @keydown.enter.prevent="selectConversation(conversation)"
            @keydown.space.prevent="selectConversation(conversation)"
          >
            <span class="conversation-copy">
              <strong>{{ conversation.title || '新对话' }}</strong>
              <small>{{ formatConversationTime(conversation.updated_at) }}</small>
            </span>
            <el-button
              link
              type="danger"
              :icon="Delete"
              aria-label="删除对话"
              class="delete-conversation"
              @click.stop="removeConversation(conversation)"
            />
          </div>
          <el-empty v-if="!loadingConversations && !conversations.length" :image-size="64" description="暂无历史对话" />
        </div>
      </aside>

      <main class="chat-main">
        <header class="chat-header">
          <el-button
            link
            class="sidebar-toggle"
            :icon="sidebarCollapsed ? Expand : Fold"
            :aria-label="sidebarCollapsed ? '展开对话列表' : '收起对话列表'"
            @click="sidebarCollapsed = !sidebarCollapsed"
          />
          <div class="chat-title">
            <strong>{{ activeTitle }}</strong>
            <small>可查询平台状态，并在授权范围内执行运维操作</small>
          </div>
          <el-button link :icon="Close" aria-label="关闭 AI 助手" @click="visible = false" />
        </header>

        <section v-if="!activeConversationID && !messages.length" class="chat-welcome">
          <span class="welcome-icon"><Cpu /></span>
          <h2>有什么可以帮你？</h2>
          <p>询问告警、拨测、数据源或自动化任务，也可以让我协助完成平台操作。</p>
          <div class="suggestions">
            <button v-for="item in suggestions" :key="item" type="button" @click="quickAsk(item)">{{ item }}</button>
          </div>
        </section>

        <section v-else ref="messageContainer" v-loading="loadingMessages" class="messages-container" aria-live="polite">
          <article v-for="message in messages" :key="message.id" class="message" :class="message.role">
            <span class="message-avatar" aria-hidden="true">
              <User v-if="message.role === 'user'" />
              <Cpu v-else />
            </span>
            <div class="message-body">
              <div class="message-meta">
                <strong>{{ message.role === 'user' ? '您' : 'AI 助手' }}</strong>
                <time>{{ formatMessageTime(message.created_at) }}</time>
              </div>
              <div v-if="message.role === 'assistant'" class="message-content markdown-body" v-html="renderMarkdown(message.content)" />
              <div v-else class="message-content user-content">{{ message.content }}</div>
            </div>
          </article>
          <article v-if="sending" class="message assistant is-thinking">
            <span class="message-avatar"><Cpu /></span>
            <div class="message-body">
              <div class="message-meta"><strong>AI 助手</strong><span>正在处理</span></div>
              <div class="thinking-dots" aria-label="AI 正在思考"><i /><i /><i /></div>
            </div>
          </article>
        </section>

        <footer class="input-area">
          <div class="input-shell">
            <el-input
              ref="inputRef"
              v-model="input"
              type="textarea"
              :rows="3"
              resize="none"
              maxlength="4000"
              show-word-limit
              placeholder="询问任何关于平台和可观测性的问题..."
              :disabled="sending"
              @keydown="handleKeydown"
            />
            <el-button
              class="send-button"
              type="primary"
              circle
              :icon="Promotion"
              :loading="sending"
              :disabled="!input.trim()"
              aria-label="发送消息"
              @click="send"
            />
          </div>
          <small>按 Enter 发送，Shift + Enter 换行 · 对话仅当前用户可见</small>
        </footer>
      </main>
    </div>
  </el-dialog>
</template>

<script setup>
import { computed, nextTick, ref } from 'vue'
import { Close, Cpu, Delete, Expand, Fold, Plus, Promotion, User } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox, useZIndex } from 'element-plus'
import { deleteConversation, getConversations, getMessages, sendMessage } from '../api/chat'

const { currentZIndex } = useZIndex()
const visible = ref(false)
const sidebarCollapsed = ref(false)
const conversations = ref([])
const activeConversationID = ref(0)
const messages = ref([])
const input = ref('')
const sending = ref(false)
const loadingConversations = ref(false)
const loadingMessages = ref(false)
const messageContainer = ref()
const inputRef = ref()

const suggestions = [
  '现在有哪些活跃告警？',
  '最近失败的拨测任务有哪些？',
  '帮我查看自动化任务执行情况',
  '平台目前有哪些数据源？',
]

const activeTitle = computed(() => {
  const active = conversations.value.find(item => item.id === activeConversationID.value)
  return active?.title || '新对话'
})

async function openChat() {
  visible.value = true
  await loadConversations()
  if (!activeConversationID.value && conversations.value.length) {
    await selectConversation(conversations.value[0])
  }
  nextTick(() => inputRef.value?.focus())
}

async function loadConversations() {
  loadingConversations.value = true
  try {
    conversations.value = (await getConversations()).data || []
  } catch {
    ElMessage.error('读取对话列表失败')
  } finally {
    loadingConversations.value = false
  }
}

function newConversation() {
  activeConversationID.value = 0
  messages.value = []
  input.value = ''
  if (window.innerWidth < 760) sidebarCollapsed.value = true
  nextTick(() => inputRef.value?.focus())
}

async function selectConversation(conversation) {
  if (loadingMessages.value || conversation.id === activeConversationID.value) return
  activeConversationID.value = conversation.id
  loadingMessages.value = true
  try {
    messages.value = (await getMessages(conversation.id)).data || []
    await scrollToBottom()
    if (window.innerWidth < 760) sidebarCollapsed.value = true
  } catch {
    ElMessage.error('读取对话消息失败')
  } finally {
    loadingMessages.value = false
  }
}

async function removeConversation(conversation) {
  try {
    await ElMessageBox.confirm(`确定删除“${conversation.title || '新对话'}”吗？`, '删除对话', { type: 'warning' })
    await deleteConversation(conversation.id)
    if (activeConversationID.value === conversation.id) newConversation()
    await loadConversations()
    ElMessage.success('对话已删除')
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') ElMessage.error('删除对话失败')
  }
}

function quickAsk(question) {
  input.value = question
  send()
}

function handleKeydown(event) {
  if (event.key !== 'Enter' || event.shiftKey || event.isComposing) return
  event.preventDefault()
  send()
}

async function send() {
  const content = input.value.trim()
  if (!content || sending.value) return

  input.value = ''
  sending.value = true
  const optimisticID = `local-${Date.now()}`
  messages.value.push({ id: optimisticID, role: 'user', content, created_at: new Date().toISOString() })
  await scrollToBottom()

  try {
    const response = await sendMessage({ conversation_id: activeConversationID.value || 0, message: content })
    activeConversationID.value = response.data.conversation_id
    const [messageResult, conversationResult] = await Promise.all([
      getMessages(activeConversationID.value),
      getConversations(),
    ])
    messages.value = messageResult.data || []
    conversations.value = conversationResult.data || []
  } catch {
    messages.value = messages.value.filter(item => item.id !== optimisticID)
    input.value = content
    ElMessage.error('发送失败，请稍后重试')
  } finally {
    sending.value = false
    await scrollToBottom()
    nextTick(() => inputRef.value?.focus())
  }
}

async function scrollToBottom() {
  await nextTick()
  if (messageContainer.value) messageContainer.value.scrollTop = messageContainer.value.scrollHeight
}

function formatConversationTime(value) {
  if (!value) return ''
  const date = new Date(value)
  const now = new Date()
  if (date.toDateString() === now.toDateString()) return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  return date.toLocaleDateString('zh-CN', { month: '2-digit', day: '2-digit' })
}

function formatMessageTime(value) {
  if (!value) return ''
  return new Date(value).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

function escapeHTML(value) {
  return String(value ?? '').replace(/[&<>"']/g, char => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[char]))
}

function renderInline(value) {
  let html = escapeHTML(value)
  html = html.replace(/`([^`]+)`/g, '<code>$1</code>')
  html = html.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
  html = html.replace(/\[([^\]]+)\]\((https?:\/\/[^\s)]+)\)/g, '<a href="$2" target="_blank" rel="noopener noreferrer">$1</a>')
  return html
}

function isTableSeparator(line) {
  return /^\s*\|?\s*:?-{3,}:?\s*(\|\s*:?-{3,}:?\s*)+\|?\s*$/.test(line)
}

function tableCells(line) {
  return line.trim().replace(/^\||\|$/g, '').split('|').map(cell => cell.trim())
}

function renderMarkdown(value) {
  const lines = String(value ?? '').replace(/\r\n/g, '\n').split('\n')
  const output = []
  let index = 0
  while (index < lines.length) {
    const line = lines[index]
    if (line.trim().startsWith('```')) {
      const code = []
      index += 1
      while (index < lines.length && !lines[index].trim().startsWith('```')) code.push(lines[index++])
      output.push(`<pre><code>${escapeHTML(code.join('\n'))}</code></pre>`)
      index += 1
      continue
    }
    if (index + 1 < lines.length && line.includes('|') && isTableSeparator(lines[index + 1])) {
      const headers = tableCells(line)
      index += 2
      const rows = []
      while (index < lines.length && lines[index].includes('|') && lines[index].trim()) rows.push(tableCells(lines[index++]))
      output.push(`<div class="markdown-table-wrap"><table><thead><tr>${headers.map(cell => `<th>${renderInline(cell)}</th>`).join('')}</tr></thead><tbody>${rows.map(row => `<tr>${row.map(cell => `<td>${renderInline(cell)}</td>`).join('')}</tr>`).join('')}</tbody></table></div>`)
      continue
    }
    if (/^\s*[-*]\s+/.test(line)) {
      const items = []
      while (index < lines.length && /^\s*[-*]\s+/.test(lines[index])) items.push(lines[index++].replace(/^\s*[-*]\s+/, ''))
      output.push(`<ul>${items.map(item => `<li>${renderInline(item)}</li>`).join('')}</ul>`)
      continue
    }
    if (/^\s*\d+\.\s+/.test(line)) {
      const items = []
      while (index < lines.length && /^\s*\d+\.\s+/.test(lines[index])) items.push(lines[index++].replace(/^\s*\d+\.\s+/, ''))
      output.push(`<ol>${items.map(item => `<li>${renderInline(item)}</li>`).join('')}</ol>`)
      continue
    }
    const heading = line.match(/^(#{1,4})\s+(.+)$/)
    if (heading) {
      const level = Math.min(heading[1].length + 2, 6)
      output.push(`<h${level}>${renderInline(heading[2])}</h${level}>`)
    } else if (/^\s*---+\s*$/.test(line)) {
      output.push('<hr>')
    } else if (/^\s*>\s?/.test(line)) {
      output.push(`<blockquote>${renderInline(line.replace(/^\s*>\s?/, ''))}</blockquote>`)
    } else if (line.trim()) {
      output.push(`<p>${renderInline(line)}</p>`)
    }
    index += 1
  }
  return output.join('')
}
</script>

<style scoped>
.ai-chat-fab{position:fixed;right:28px;bottom:88px;z-index:1900;display:grid;width:54px;height:54px;place-items:center;padding:0;color:#fff;background:linear-gradient(145deg,var(--ops-primary),#315f95);border:0;border-radius:16px;box-shadow:0 12px 28px color-mix(in srgb,var(--ops-primary) 34%,transparent);cursor:pointer;transition:transform .18s ease,box-shadow .18s ease}.ai-chat-fab:hover{transform:translateY(-2px);box-shadow:0 16px 34px color-mix(in srgb,var(--ops-primary) 42%,transparent)}.ai-chat-fab:focus-visible{outline:3px solid color-mix(in srgb,var(--ops-primary) 36%,#fff);outline-offset:3px}.ai-chat-fab svg{width:31px;height:31px;fill:currentColor}.chat-shell{display:flex;height:min(80vh,760px);min-height:540px;overflow:hidden;background:var(--ops-surface)}.chat-sidebar{display:flex;width:270px;min-width:270px;flex-direction:column;background:var(--ops-surface-soft);border-right:1px solid var(--ops-border);transition:width .2s ease,min-width .2s ease}.chat-sidebar.is-collapsed{width:0;min-width:0;overflow:hidden;border-right:0}.sidebar-header{display:flex;height:68px;box-sizing:border-box;align-items:center;justify-content:space-between;gap:10px;padding:12px 14px;background:var(--ops-surface);border-bottom:1px solid var(--ops-border-soft)}.assistant-brand{display:flex;min-width:0;align-items:center;gap:9px}.assistant-mark,.welcome-icon{display:grid;place-items:center;color:#fff;background:linear-gradient(145deg,var(--ops-primary),#315f95)}.assistant-mark{width:34px;height:34px;flex:0 0 34px;border-radius:10px}.assistant-mark svg{width:20px}.assistant-brand>div{display:flex;min-width:0;flex-direction:column}.assistant-brand strong{color:var(--ops-text);font-size:12px;white-space:nowrap}.assistant-brand small{color:var(--ops-muted);font-size:10px;white-space:nowrap}.sidebar-header>.el-button{padding-inline:9px}.conversation-list{min-height:0;flex:1;overflow-y:auto;padding:8px}.conversation-item{display:flex;width:100%;min-height:58px;align-items:center;gap:8px;margin-bottom:4px;padding:9px 8px 9px 11px;color:inherit;text-align:left;background:transparent;border:1px solid transparent;border-radius:7px;cursor:pointer}.conversation-item:hover,.conversation-item.active{background:var(--ops-surface);border-color:var(--ops-border-soft)}.conversation-item.active{box-shadow:inset 3px 0 var(--ops-primary)}.conversation-copy{display:flex;min-width:0;flex:1;flex-direction:column;gap:5px}.conversation-copy strong{overflow:hidden;color:var(--ops-text);font-size:12px;font-weight:650;text-overflow:ellipsis;white-space:nowrap}.conversation-copy small{color:var(--ops-muted);font-size:10px}.delete-conversation{flex:none;opacity:0}.conversation-item:hover .delete-conversation,.conversation-item:focus-within .delete-conversation{opacity:1}.chat-main{display:flex;min-width:0;flex:1;flex-direction:column;background:var(--ops-bg)}.chat-header{display:flex;height:68px;box-sizing:border-box;align-items:center;gap:12px;padding:0 17px;background:var(--ops-surface);border-bottom:1px solid var(--ops-border-soft)}.sidebar-toggle{flex:none}.chat-title{display:flex;min-width:0;flex:1;flex-direction:column;gap:3px}.chat-title strong{overflow:hidden;color:var(--ops-text);font-size:14px;text-overflow:ellipsis;white-space:nowrap}.chat-title small{overflow:hidden;color:var(--ops-muted);font-size:10px;text-overflow:ellipsis;white-space:nowrap}.chat-welcome{display:flex;min-height:0;flex:1;align-items:center;justify-content:center;flex-direction:column;padding:36px;text-align:center}.welcome-icon{width:58px;height:58px;border-radius:18px;box-shadow:0 14px 30px color-mix(in srgb,var(--ops-primary) 20%,transparent)}.welcome-icon svg{width:31px}.chat-welcome h2{margin:18px 0 8px;color:var(--ops-text);font-size:20px}.chat-welcome p{max-width:520px;margin:0;color:var(--ops-muted);font-size:12px;line-height:1.7}.suggestions{display:grid;width:min(580px,100%);grid-template-columns:1fr 1fr;gap:9px;margin-top:24px}.suggestions button{min-height:42px;padding:8px 12px;color:var(--ops-text);text-align:left;background:var(--ops-surface);border:1px solid var(--ops-border);border-radius:7px;cursor:pointer;font:inherit;font-size:11px;transition:border-color .16s ease,color .16s ease}.suggestions button:hover{color:var(--ops-primary);border-color:var(--ops-primary)}.messages-container{min-height:0;flex:1;overflow-y:auto;padding:22px 24px}.message{display:flex;align-items:flex-start;gap:10px;margin-bottom:22px}.message.user{flex-direction:row-reverse}.message-avatar{display:grid;width:34px;height:34px;flex:0 0 34px;place-items:center;color:var(--ops-primary);background:var(--ops-primary-soft);border:1px solid color-mix(in srgb,var(--ops-primary) 22%,var(--ops-border));border-radius:10px}.message.user .message-avatar{color:#fff;background:#27354d;border-color:#27354d}.message-avatar svg{width:18px}.message-body{min-width:0;max-width:min(760px,82%)}.message.user .message-body{display:flex;align-items:flex-end;flex-direction:column}.message-meta{display:flex;align-items:center;gap:8px;margin-bottom:6px;color:var(--ops-muted);font-size:10px}.message-meta strong{color:var(--ops-text);font-size:11px}.message-content{box-sizing:border-box;padding:11px 14px;color:var(--ops-text);background:var(--ops-surface);border:1px solid var(--ops-border-soft);border-radius:4px 12px 12px 12px;font-size:12px;line-height:1.7;box-shadow:0 2px 7px color-mix(in srgb,var(--ops-text) 4%,transparent)}.message.user .message-content{color:#fff;background:#315f95;border-color:#315f95;border-radius:12px 4px 12px 12px;white-space:pre-wrap}.thinking-dots{display:flex;width:58px;gap:5px;padding:14px 16px;background:var(--ops-surface);border:1px solid var(--ops-border-soft);border-radius:4px 12px 12px 12px}.thinking-dots i{width:6px;height:6px;background:var(--ops-primary);border-radius:50%;animation:chat-dot 1.2s infinite ease-in-out}.thinking-dots i:nth-child(2){animation-delay:.15s}.thinking-dots i:nth-child(3){animation-delay:.3s}@keyframes chat-dot{0%,60%,100%{opacity:.25;transform:translateY(0)}30%{opacity:1;transform:translateY(-3px)}}.input-area{padding:12px 16px 10px;background:var(--ops-surface);border-top:1px solid var(--ops-border-soft)}.input-shell{position:relative}.input-shell :deep(.el-textarea__inner){padding:11px 48px 24px 12px;border-radius:8px;box-shadow:0 0 0 1px var(--ops-border) inset!important}.input-shell :deep(.el-input__count){right:47px;bottom:5px;background:transparent;font-size:9px}.send-button{position:absolute;right:10px;bottom:10px;width:31px;height:31px}.input-area>small{display:block;margin-top:6px;color:var(--ops-muted);font-size:9px}.markdown-body :deep(p){margin:0 0 8px}.markdown-body :deep(p:last-child){margin-bottom:0}.markdown-body :deep(h3),.markdown-body :deep(h4),.markdown-body :deep(h5),.markdown-body :deep(h6){margin:12px 0 6px;font-size:13px}.markdown-body :deep(code){padding:1px 4px;color:#315f95;background:var(--ops-primary-soft);border-radius:3px;font:11px ui-monospace,SFMono-Regular,Menlo,monospace}.markdown-body :deep(pre){overflow:auto;margin:8px 0;padding:11px;color:#dbeafe;background:#111827;border-radius:6px;font:10px/1.6 ui-monospace,SFMono-Regular,Menlo,monospace;white-space:pre-wrap}.markdown-body :deep(pre code){padding:0;color:inherit;background:transparent}.markdown-body :deep(ul),.markdown-body :deep(ol){margin:7px 0;padding-left:20px}.markdown-body :deep(blockquote){margin:8px 0;padding:7px 10px;color:var(--ops-muted);background:var(--ops-surface-soft);border-left:3px solid var(--ops-primary)}.markdown-body :deep(hr){margin:12px 0;border:0;border-top:1px solid var(--ops-border)}.markdown-body :deep(a){color:var(--ops-primary)}.markdown-table-wrap{max-width:100%;overflow-x:auto;margin:8px 0}.markdown-body :deep(table){width:100%;border-collapse:collapse}.markdown-body :deep(th),.markdown-body :deep(td){padding:6px 8px;text-align:left;border:1px solid var(--ops-border);font-size:10px}.markdown-body :deep(th){background:var(--ops-surface-soft);font-weight:650}.conversation-list :deep(.el-empty){padding:52px 0}.conversation-list :deep(.el-empty__description p){font-size:11px}@media(max-width:760px){.chat-shell{height:88vh;min-height:0}.chat-sidebar{position:absolute;z-index:3;height:100%;box-shadow:10px 0 24px rgba(16,24,40,.12)}.chat-sidebar.is-collapsed{box-shadow:none}.chat-header{padding-inline:10px}.chat-title small{display:none}.messages-container{padding:16px 12px}.message-body{max-width:88%}.suggestions{grid-template-columns:1fr}.new-chat-copy{display:none}.ai-chat-fab{right:18px;bottom:88px}}
</style>

<style>
.ai-chat-dialog .markdown-table-wrap{max-width:100%;overflow-x:auto;margin:8px 0}
.ai-chat-dialog{overflow:hidden;border-radius:16px}.ai-chat-dialog .el-dialog__header{display:none}.ai-chat-dialog .el-dialog__body{height:auto;padding:0}.ai-chat-dialog .el-loading-mask{background-color:color-mix(in srgb,var(--ops-surface) 72%,transparent)}
</style>
