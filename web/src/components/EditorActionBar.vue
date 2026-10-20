<template>
  <footer class="editor-action-bar">
    <div class="editor-action-leading"><slot name="leading" /></div>
    <div class="editor-action-primary"><slot /></div>
  </footer>
</template>

<style scoped>
.editor-action-bar {
  position: fixed;
  right: 0;
  bottom: 0;
  left: var(--sidebar-width, 248px);
  z-index: 20;
  display: flex;
  min-height: 62px;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 0 24px;
  border-top: 1px solid #dfe6ee;
  background: rgba(255, 255, 255, .96);
  box-shadow: 0 -7px 22px rgba(31, 50, 81, .07);
  backdrop-filter: blur(8px);
}
.editor-action-leading,
.editor-action-primary,
.editor-action-leading :deep(.editor-action-copy) {
  display: flex;
  align-items: center;
  gap: 9px;
}
.editor-action-leading { min-width: 0; color: #8a96a8; font-size: 11px; }
.editor-action-leading :deep(strong) { color: #475467; font-size: 12px; }
.editor-action-primary :deep(.el-button + .el-button) { margin-left: 0; }
@media (max-width: 700px) {
  .editor-action-bar { left: 0; padding: 0 14px; }
  .editor-action-leading :deep(.editor-action-hint) { display: none; }
}
</style>
