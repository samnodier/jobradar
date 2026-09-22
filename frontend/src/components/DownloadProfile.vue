<script setup lang="ts">
import { ref } from 'vue'
import { fetchProfileDownload } from '@/utils/profileDownload'

const downloading = ref(false)
const error = ref('')
const downloaded = ref(false)
async function download() {
  if (downloading.value) return
  downloading.value = true
  downloaded.value = false
  error.value = ''
  try {
    const data = await fetchProfileDownload()
    const url = URL.createObjectURL(
      new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' }),
    )
    const link = document.createElement('a')
    link.href = url
    link.download = 'jobradar-profile.json'
    document.body.appendChild(link)
    link.click()
    link.remove()
    // Keep the object URL alive until the browser starts handling the download.
    setTimeout(() => URL.revokeObjectURL(url), 1000)
    downloaded.value = true
  } catch (err) {
    error.value =
      err instanceof Error ? err.message : 'Could not download your profile. Please retry.'
  } finally {
    downloading.value = false
  }
}
</script>

<template>
  <div>
    <button type="button" class="button button-secondary" :disabled="downloading" @click="download">
      {{ downloading ? 'Preparing download…' : 'Download profile' }}
    </button>
    <p class="mt-2 text-sm text-gray-600">
      JSON copy of your saved profile, resume details and preferences. Includes contact information;
      excludes credentials, saved jobs and applications. This is not a full account backup.
    </p>
    <p v-if="error" role="alert" class="mt-2 text-sm text-red-600">{{ error }}</p>
    <p v-else-if="downloaded" role="status" class="mt-2 text-sm text-gray-600">
      Profile download prepared.
    </p>
  </div>
</template>
