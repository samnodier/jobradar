<script setup lang="ts"></script>

async function test () { const res = await fetch('/api/jobs'); console.log(await res.json()); }

<template>
  <h1>Jobradar</h1>
  <button @click="test">Test API</button>
</template>

<style scoped></style>
