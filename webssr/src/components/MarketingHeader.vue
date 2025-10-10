<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter, useRoute } from 'vue-router/auto'
import { storeToRefs } from 'pinia'

import { Button } from '~/components/ui/button'
import { LoginAction } from '~/components/login'
import LanguageSwitcher from '~/components/LanguageSwitcher.vue'
import UserMenu from '~/components/UserMenu.vue'
import CarbonIcon from '~/components/CarbonIcon.vue'
import { isDark, toggleDark } from '~/composables/dark'
import { useLocaleNavigation } from '~/composables/useLocaleNavigation'
import { useAuthStore } from '~/stores/auth'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const { isAuthenticated } = storeToRefs(authStore)
const { push } = useLocaleNavigation(router, route)

const navItems = computed(() => [
  { label: t('marketing.nav.features'), href: '#features' },
  { label: t('marketing.nav.pricing'), href: '#pricing' },
  { label: t('marketing.nav.testimonials'), href: '#testimonials' },
  { label: t('marketing.nav.resources'), href: '#resources' },
])

function handleLogoClick() {
  push({ name: '/' })
}
</script>

<template>
  <header class="sticky top-0 z-40 border-b border-border/80 bg-background/80 backdrop-blur">
    <div class="mx-auto flex h-16 max-w-6xl items-center gap-6 px-4 sm:px-6">
      <button
        type="button"
        class="flex min-w-0 items-center gap-2 text-lg font-semibold tracking-tight text-foreground"
        @click="handleLogoClick"
      >
        <span class="text-2xl font-black text-foreground">SaaSwift</span>
        <span class="hidden truncate text-xs text-muted-foreground sm:block">
          {{ t('marketing.brandTagline') }}
        </span>
      </button>

      <nav class="mx-auto hidden items-center gap-8 text-sm font-medium text-muted-foreground lg:flex">
        <a
          v-for="item in navItems"
          :key="item.href"
          :href="item.href"
          class="transition-colors hover:text-foreground"
        >
          {{ item.label }}
        </a>
      </nav>

      <div class="ml-auto flex items-center gap-3">
        <Button
          type="button"
          variant="ghost"
          size="icon-sm"
          class="rounded-full border border-border/60"
          :title="t('marketing.header.toggleTheme')"
          @click="toggleDark()"
        >
          <CarbonIcon :name="isDark ? 'moon' : 'sun'" class="h-4 w-4" />
        </Button>
        <LanguageSwitcher />
        <template v-if="isAuthenticated">
          <UserMenu />
        </template>
        <template v-else>
          <LoginAction>
            <Button size="sm" class="hidden md:inline-flex" variant="outline">
              {{ t('marketing.header.login') }}
            </Button>
          </LoginAction>
        </template>
      </div>
    </div>
  </header>
</template>
