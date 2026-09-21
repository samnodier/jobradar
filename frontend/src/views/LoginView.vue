<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const email = ref('')
const password = ref('')
const confirmation = ref('')
const setupToken = ref('')
const githubEnabled = ref(false)
const busy = ref(false)
const error = ref(
  route.query.error === 'account_exists'
    ? 'This email already has a Jobradar account. Sign in with your password, then connect GitHub in Settings.'
    : '',
)
onMounted(async () => {
  setupToken.value = new URLSearchParams(window.location.hash.slice(1)).get('setup') || ''
  if (setupToken.value) await router.replace({ hash: '' })
  try {
    const response = await fetch('/auth/methods', { credentials: 'include' })
    if (!response.ok) throw new Error('Could not load sign-in methods. Please refresh.')
    githubEnabled.value = (await response.json()).github
  } catch (e) {
    error.value = String(e)
  }
})
async function submit() {
  error.value = ''
  if (setupToken.value && password.value !== confirmation.value) {
    error.value = 'Passwords do not match'
    return
  }
  busy.value = true
  try {
    const endpoint = setupToken.value ? '/auth/password/setup' : '/auth/password/login'
    const body = setupToken.value
      ? { token: setupToken.value, password: password.value }
      : { email: email.value, password: password.value }
    const response = await fetch(endpoint, {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    })
    if (!response.ok) {
      const data = await response.json().catch(() => null)
      throw new Error(data?.error || 'Sign-in failed')
    }
    setupToken.value = ''
    await auth.fetchMe()
    if (!auth.user)
      throw new Error(
        'Your password was accepted, but the session could not load. Please sign in again.',
      )
    await router.replace('/jobs')
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    busy.value = false
    password.value = ''
    confirmation.value = ''
  }
}
</script>

<template>
  <div class="flex flex-col items-center justify-center flex-1 gap-8 py-8">
    <div class="w-full max-w-95 bg-white border border-[#e8e8e4] p-8 flex flex-col justify-center">
      <!-- Head -->
      <div class="text-center mb-4">
        <div class="w-10 h-10 bg-[#5e6ad2] flex items-center justify-center mx-auto mb-4">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            height="24px"
            viewBox="0 -960 960 960"
            width="24px"
            fill="#e3e3e3"
          >
            <path
              d="M367-527q-47-47-47-113t47-113q47-47 113-47t113 47q47 47 47 113t-47 113q-47 47-113 47t-113-47ZM160-160v-112q0-34 17.5-62.5T224-378q62-31 126-46.5T480-440q66 0 130 15.5T736-378q29 15 46.5 43.5T800-272v112H160Zm80-80h480v-32q0-11-5.5-20T700-306q-54-27-109-40.5T480-360q-56 0-111 13.5T260-306q-9 5-14.5 14t-5.5 20v32Zm296.5-343.5Q560-607 560-640t-23.5-56.5Q513-720 480-720t-56.5 23.5Q400-673 400-640t23.5 56.5Q447-560 480-560t56.5-23.5ZM480-640Zm0 400Z"
            />
          </svg>
        </div>
        <h1 class="text-lg font-medium text-[#1a1a1a] mb-1.5">Welcome to JobRadar</h1>
        <p class="text-[13px] text-[#888] leading-relaxed">
          Sign in to track jobs matched to your skills
        </p>
      </div>

      <form class="space-y-4 mb-5" @submit.prevent="submit">
        <h2 v-if="setupToken" class="font-medium">Choose your password</h2>
        <p v-if="error" role="alert" class="text-sm text-red-600">{{ error }}</p>
        <label v-if="!setupToken" class="block text-sm"
          >Email
          <input
            v-model="email"
            type="email"
            autocomplete="username"
            required
            maxlength="254"
            class="mt-1 w-full border border-gray-300 rounded p-2"
          />
        </label>
        <label class="block text-sm"
          >Password
          <input
            v-model="password"
            type="password"
            :autocomplete="setupToken ? 'new-password' : 'current-password'"
            :minlength="setupToken ? 15 : undefined"
            required
            maxlength="1024"
            class="mt-1 w-full border border-gray-300 rounded p-2"
          />
        </label>
        <label v-if="setupToken" class="block text-sm"
          >Confirm password
          <input
            v-model="confirmation"
            type="password"
            autocomplete="new-password"
            required
            maxlength="1024"
            class="mt-1 w-full border border-gray-300 rounded p-2"
          />
        </label>
        <p v-if="setupToken" class="text-xs text-gray-600">
          Use a passphrase of at least 15 characters. Saving it signs out older sessions and revokes
          workflow tokens.
        </p>
        <button
          type="submit"
          :disabled="busy"
          class="w-full bg-[#5e6ad2] text-white p-2 disabled:opacity-50"
        >
          {{
            busy ? 'Please wait…' : setupToken ? 'Save password and sign in' : 'Sign in with email'
          }}
        </button>
      </form>

      <a
        v-if="githubEnabled && !setupToken"
        href="/auth/github/login"
        class="w-full inline-flex items-center justify-center gap-2 px-6 py-2 bg-[#24292e] text-white no-underline font-medium transition-colors hover:bg-[#2c3e50]"
      >
        <svg width="18" height="18" viewBox="0 0 18 18" fill="currentColor">
          <path
            d="M9 .5A8.5 8.5 0 00.5 9c0 3.75 2.434 6.929 5.81 8.054.425.078.58-.184.58-.41v-1.437c-2.362.514-2.86-1.138-2.86-1.138-.386-.981-.943-1.242-.943-1.242-.771-.527.058-.516.058-.516.852.06 1.3.875 1.3.875.757 1.298 1.986.923 2.47.706.077-.549.296-.924.539-1.136-1.886-.214-3.868-.943-3.868-4.196 0-.927.33-1.685.875-2.279-.088-.215-.38-1.078.083-2.247 0 0 .713-.228 2.335.87A8.13 8.13 0 019 4.82a8.13 8.13 0 012.128.287c1.62-1.098 2.333-.87 2.333-.87.464 1.169.172 2.032.084 2.247.546.594.874 1.352.874 2.279 0 3.262-1.985 3.98-3.876 4.19.305.263.576.78.576 1.572v2.33c0 .228.154.492.584.409A8.503 8.503 0 0017.5 9 8.5 8.5 0 009 .5z"
          />
        </svg>
        Continue with GitHub
      </a>

      <p v-if="!setupToken" class="text-xs text-gray-600 mt-4">
        Already use GitHub? Sign in with GitHub and add a password under Profile → Settings →
        Sign-in methods.
      </p>
      <details v-if="!setupToken" class="text-xs text-gray-600 mt-3">
        <summary class="cursor-pointer">First-time setup or forgotten password</summary>
        <p class="mt-2">
          Ask the installation owner for a private password setup link. For your own server, follow
          the owner setup instructions in the README. Public email registration is disabled.
        </p>
      </details>

      <p class="text-center text-[12px] text-[#aaa] leading-[1.7] mt-5">
        By continuing, you agree to our
        <RouterLink to="/terms" class="text-[#5e6ad2] no-underline hover:underline"
          >Terms of Service</RouterLink
        >
        and
        <RouterLink to="/privacy" class="text-[#5e6ad2] no-underline hover:underline"
          >Privacy Policy</RouterLink
        >
      </p>
    </div>

    <!-- Feature pills -->
    <div class="flex justify-center gap-2 flex-wrap">
      <div
        v-for="label in ['Skill-based matching', 'Daily job digest', 'Track applications']"
        :key="label"
        class="flex items-center gap-1.25 h-6 px-2.5 bg-white border border-[#e8e8e4] text-[11.5px] text-[#666]"
      >
        <span class="w-1.5 h-1.5 bg-[#5e6ad2] shrink-0" />
        {{ label }}
      </div>
    </div>
  </div>
</template>
