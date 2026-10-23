<template>
  <router-view />
</template>

<script setup>
import { onMounted } from 'vue'
import { usePlatformStore } from './stores/platform'

const platform = usePlatformStore()
onMounted(() => platform.load())
</script>
