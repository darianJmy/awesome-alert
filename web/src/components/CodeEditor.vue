<template>
  <div ref="container" class="code-editor" :style="{ height }" />
</template>

<script setup>
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import * as monaco from 'monaco-editor/editor/editor.api'
import EditorWorker from 'monaco-editor/editor/editor.worker?worker'
import 'monaco-editor/languages/definitions/python/register'
import 'monaco-editor/languages/definitions/shell/register'

const props = defineProps({
  modelValue: { type: String, default: '' },
  language: { type: String, default: 'shell' },
  height: { type: String, default: '420px' },
  ariaLabel: { type: String, default: '代码编辑器' },
})
const emit = defineEmits(['update:modelValue'])
const container = ref()
let editor
let contentSubscription

self.MonacoEnvironment = { getWorker: () => new EditorWorker() }

if (!monaco.languages.getLanguages().some(item => item.id === 'yaml')) {
  monaco.languages.register({ id: 'yaml', extensions: ['.yaml', '.yml'], aliases: ['YAML', 'yaml'] })
  monaco.languages.setMonarchTokensProvider('yaml', {
    tokenizer: {
      root: [
        [/^\s*#.*/, 'comment'],
        [/^([\w.-]+)(\s*:)/, ['key', 'delimiter']],
        [/\b(true|false|null)\b/, 'keyword'],
        [/-?\d+(\.\d+)?/, 'number'],
        [/"([^"\\]|\\.)*"/, 'string'],
        [/'[^']*'/, 'string'],
      ],
    },
  })
}

onMounted(() => {
  editor = monaco.editor.create(container.value, {
    value: props.modelValue,
    language: props.language,
    theme: 'vs-dark',
    automaticLayout: true,
    ariaLabel: props.ariaLabel,
    fontFamily: "'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace",
    fontSize: 13,
    lineHeight: 21,
    minimap: { enabled: false },
    padding: { top: 12, bottom: 12 },
    scrollBeyondLastLine: false,
    tabSize: 2,
    insertSpaces: true,
    wordWrap: 'on',
  })
  contentSubscription = editor.onDidChangeModelContent(() => emit('update:modelValue', editor.getValue()))
})

watch(() => props.modelValue, value => {
  if (editor && value !== editor.getValue()) editor.setValue(value || '')
})

watch(() => props.language, language => {
  const model = editor?.getModel()
  if (model) monaco.editor.setModelLanguage(model, language || 'plaintext')
})

onBeforeUnmount(() => {
  contentSubscription?.dispose()
  editor?.dispose()
})
</script>

<style scoped>
.code-editor { width: 100%; overflow: hidden; border: 1px solid #d8dee8; border-radius: 7px; }
.code-editor:focus-within { border-color: #409eff; box-shadow: 0 0 0 1px #409eff inset; }
</style>
