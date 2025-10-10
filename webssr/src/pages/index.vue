<route lang="yaml">
meta:
  layout: home
</route>
<script setup lang="ts">
import LocaleIndexPage from './[locale]/index.vue'

defineOptions({
  name: 'IndexPage',
})
</script>

<template>
  <LocaleIndexPage />
</template>
