<template>
  <div class="chat-layout">
    <!-- 左侧对话列表 -->
    <div class="chat-sidebar">
      <div class="sidebar-header">
        <span>对话列表</span>
        <el-button size="small" circle @click="newConversation">+</el-button>
      </div>
      <div class="convo-list">
        <div v-for="c in conversations" :key="c.id" class="convo-item" :class="{ active: c.id === activeConvo }"
          @click="selectConvo(c)">
          <span class="convo-title">{{ c.title }}</span>
          <el-button link size="small" type="danger" @click.stop="deleteConvo(c)"><el-icon><Delete /></el-icon></el-button>
        </div>
      </div>
    </div>

    <!-- 右侧聊天区 -->
    <div class="chat-main">
      <div v-if="!activeConvo" class="chat-welcome">
        <el-icon :size="48" color="#909399"><ChatDotSquare /></el-icon>
        <p>智能运维助手</p>
        <span>你可以问我关于 Prometheus 数据源的问题</span>
        <div class="suggestions">
          <el-tag v-for="s in suggestions" :key="s" class="sugg-tag" @click="quickAsk(s)">{{ s }}</el-tag>
        </div>
      </div>

      <template v-else>
        <div class="chat-messages" ref="msgBox">
          <div v-for="m in messages" :key="m.id" class="msg-row" :class="m.role">
            <div class="msg-avatar">
              <el-icon v-if="m.role === 'user'" :size="20"><User /></el-icon>
              <el-icon v-else :size="20" color="#409EFF"><Cpu /></el-icon>
            </div>
            <div class="msg-bubble" v-html="renderContent(m.content)"/>
          </div>
          <div v-if="sending" class="msg-row assistant">
            <div class="msg-avatar"><el-icon :size="20" color="#409EFF"><Cpu /></el-icon></div>
            <div class="msg-bubble thinking">正在分析中...</div>
          </div>
        </div>

        <div class="chat-input">
          <el-input v-model="input" placeholder="输入你的问题，如：有多少节点在线？" @keyup.enter="handleSend"
            :disabled="sending" clearable>
            <template #append>
              <el-button :loading="sending" @click="handleSend" :icon="Promotion" />
            </template>
          </el-input>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, nextTick, onMounted } from 'vue'
import { getConversations, createConversation, deleteConversation, getMessages, sendMessage } from '../api/chat'
import { ElMessage, ElMessageBox } from 'element-plus'

const conversations = ref([])
const activeConvo = ref(null)
const messages = ref([])
const input = ref('')
const sending = ref(false)
const msgBox = ref(null)

const suggestions = [
  '有多少节点在线？',
  'CPU使用率情况',
  '内存使用情况',
  '磁盘使用率',
]

onMounted(async () => {
  try { conversations.value = (await getConversations()).data } catch {}
})

async function newConversation() {
  activeConvo.value = null; messages.value = []; input.value = ''
}

async function selectConvo(c) {
  activeConvo.value = c.id
  try { messages.value = (await getMessages(c.id)).data } catch {}
  scrollDown()
}

async function deleteConvo(c) {
  await ElMessageBox.confirm('确定删除此对话？', '提示', { type: 'warning' })
  await deleteConversation(c.id)
  if (activeConvo.value === c.id) { activeConvo.value = null; messages.value = [] }
  conversations.value = (await getConversations()).data
}

async function quickAsk(q) { input.value = q; await handleSend() }

async function handleSend() {
  if (!input.value.trim() || sending.value) return
  sending.value = true
  const msg = input.value; input.value = ''

  try {
    const res = await sendMessage({
      conversation_id: activeConvo.value || 0,
      message: msg,
    })
    activeConvo.value = res.data.conversation_id
    messages.value = (await getMessages(activeConvo.value)).data
    conversations.value = (await getConversations()).data
  } catch {}
  sending.value = false
  scrollDown()
}

function scrollDown() {
  nextTick(() => { setTimeout(() => { if (msgBox.value) msgBox.value.scrollTop = msgBox.value.scrollHeight }, 100) })
}

function renderContent(text) {
  return text
    .replace(/## (.*)/g, '<h4>$1</h4>')
    .replace(/`([^`]+)`/g, '<code>$1</code>')
    .replace(/```\n([\s\S]*?)```/g, '<pre>$1</pre>')
    .replace(/\n/g, '<br>')
}
</script>

<style scoped>
.chat-layout { display: flex; height: calc(100vh - 100px); margin: -20px; }
.chat-sidebar { width: 260px; background: #fff; border-right: 1px solid #e6e6e6; display: flex; flex-direction: column; }
.sidebar-header { display: flex; justify-content: space-between; align-items: center; padding: 16px; border-bottom: 1px solid #e6e6e6; font-weight: bold; }
.convo-list { flex: 1; overflow-y: auto; }
.convo-item { display: flex; justify-content: space-between; align-items: center; padding: 12px 16px; cursor: pointer; border-bottom: 1px solid #f0f0f0; }
.convo-item:hover, .convo-item.active { background: #ecf5ff; }
.convo-title { font-size: 14px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; flex: 1; }

.chat-main { flex: 1; display: flex; flex-direction: column; background: #f5f7fa; }
.chat-welcome { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; color: #909399; }
.chat-welcome p { font-size: 20px; margin: 12px 0 4px; color: #303133; font-weight: bold; }
.suggestions { margin-top: 20px; display: flex; gap: 8px; flex-wrap: wrap; justify-content: center; }
.sugg-tag { cursor: pointer; font-size: 13px; }

.chat-messages { flex: 1; overflow-y: auto; padding: 20px; }
.msg-row { display: flex; gap: 10px; margin-bottom: 20px; }
.msg-row.user { flex-direction: row-reverse; }
.msg-avatar { width: 36px; height: 36px; border-radius: 50%; background: #e6e6e6; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.msg-row.user .msg-avatar { background: #409EFF; color: #fff; }
.msg-bubble { max-width: 70%; padding: 10px 16px; border-radius: 8px; font-size: 14px; line-height: 1.6; word-break: break-word; }
.msg-row.user .msg-bubble { background: #409EFF; color: #fff; }
.msg-row.assistant .msg-bubble { background: #fff; box-shadow: 0 1px 2px rgba(0,0,0,0.1); }
.msg-bubble.thinking { color: #909399; font-style: italic; }
.msg-bubble :deep(h4) { margin: 0 0 4px; font-size: 14px; }
.msg-bubble :deep(code) { background: rgba(0,0,0,0.06); padding: 1px 4px; border-radius: 3px; font-size: 13px; }
.msg-bubble :deep(pre) { background: #1e1e1e; color: #d4d4d4; padding: 10px; border-radius: 4px; overflow-x: auto; font-size: 12px; margin: 6px 0; white-space: pre-wrap; }

.chat-input { padding: 12px 20px; background: #fff; border-top: 1px solid #e6e6e6; }
</style>
