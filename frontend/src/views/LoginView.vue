<script setup lang="ts">
const githubLoginUrl = '/auth/github/login'
</script>

<template>

  <div class="center">
    <div class="card">
      <div class="card-head">
        <div class="card-icon">
          <svg xmlns="http://www.w3.org/2000/svg" height="24px" viewBox="0 -960 960 960" width="24px" fill="#e3e3e3">
            <path
              d="M367-527q-47-47-47-113t47-113q47-47 113-47t113 47q47 47 47 113t-47 113q-47 47-113 47t-113-47ZM160-160v-112q0-34 17.5-62.5T224-378q62-31 126-46.5T480-440q66 0 130 15.5T736-378q29 15 46.5 43.5T800-272v112H160Zm80-80h480v-32q0-11-5.5-20T700-306q-54-27-109-40.5T480-360q-56 0-111 13.5T260-306q-9 5-14.5 14t-5.5 20v32Zm296.5-343.5Q560-607 560-640t-23.5-56.5Q513-720 480-720t-56.5 23.5Q400-673 400-640t23.5 56.5Q447-560 480-560t56.5-23.5ZM480-640Zm0 400Z" />
          </svg>
        </div>
        <h1 class="card-title">Welcome to JobRadar</h1>
        <p class="card-sub">Sign in to track jobs matched to your skills</p>
      </div>

      <!--
        <button class="oauth-btn" @click="loginWith('google')">
          <svg width="18" height="18" viewBox="0 0 18 18">
            <path d="M17.64 9.2c0-.637-.057-1.251-.164-1.84H9v3.481h4.844c-.209 1.125-.843 2.078-1.796 2.717v2.258h2.908c1.702-1.567 2.684-3.875 2.684-6.615z" fill="#4285F4" />
            <path d="M9 18c2.43 0 4.467-.806 5.956-2.18l-2.908-2.259c-.806.54-1.837.86-3.048.86-2.344 0-4.328-1.584-5.036-3.711H.957v2.332A8.997 8.997 0 009 18z" fill="#34A853" />
            <path d="M3.964 10.71A5.41 5.41 0 013.682 9c0-.593.102-1.17.282-1.71V4.958H.957A8.996 8.996 0 000 9c0 1.452.348 2.827.957 4.042l3.007-2.332z" fill="#FBBC05" />
            <path d="M9 3.58c1.321 0 2.508.454 3.44 1.345l2.582-2.58C13.463.891 11.426 0 9 0A8.997 8.997 0 00.957 4.958L3.964 6.29C4.672 4.163 6.656 3.58 9 3.58z" fill="#EA4335" />
          </svg>
          Continue with Google
        </button>
      -->
      <a :href="githubLoginUrl" class="oauth-btn"> <svg width="18" height="18" viewBox="0 0 18 18" fill="currentColor">
          <path
            d="M9 .5A8.5 8.5 0 00.5 9c0 3.75 2.434 6.929 5.81 8.054.425.078.58-.184.58-.41v-1.437c-2.362.514-2.86-1.138-2.86-1.138-.386-.981-.943-1.242-.943-1.242-.771-.527.058-.516.058-.516.852.06 1.3.875 1.3.875.757 1.298 1.986.923 2.47.706.077-.549.296-.924.539-1.136-1.886-.214-3.868-.943-3.868-4.196 0-.927.33-1.685.875-2.279-.088-.215-.38-1.078.083-2.247 0 0 .713-.228 2.335.87A8.13 8.13 0 019 4.82a8.13 8.13 0 012.128.287c1.62-1.098 2.333-.87 2.333-.87.464 1.169.172 2.032.084 2.247.546.594.874 1.352.874 2.279 0 3.262-1.985 3.98-3.876 4.19.305.263.576.78.576 1.572v2.33c0 .228.154.492.584.409A8.503 8.503 0 0017.5 9 8.5 8.5 0 009 .5z" />
        </svg>
        Continue with GitHub
      </a>

      <p class="terms">
        By continuing, you agree to our
        <RouterLink to="/terms">Terms of Service</RouterLink>
        and
        <RouterLink to="/privacy">Privacy Policy</RouterLink>
      </p>
    </div>

    <div class="feature-pills">
      <div class="pill">
        <span class="pill-dot" />
        Skill-based matching
      </div>
      <div class="pill">
        <span class="pill-dot" />
        Daily job digest
      </div>
      <div class="pill">
        <span class="pill-dot" />
        Track applications
      </div>
    </div>
  </div>
</template>

<style scoped>
.center {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  flex: 1;
  gap: 2rem;
  padding: var(--spacing-8) 0;
}

.card {
  width: 100%;
  max-width: 380px;
  background: #fff;
  border: 1px solid #e8e8e4;
  padding: 32px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.oauth-btn {
  width: 100%;
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1.5rem;
  background: #24292e;
  color: white;
  text-decoration: none;
  font-weight: 500;
  display: flex;
  justify-content: center;
  align-items: center;
  transition: background 0.1s, border-color 0.1s;
  border: var(--muted)
}

.oauth-btn:hover {
  background: #2c3e50;
}

.card-head {
  text-align: center;
  margin-bottom: 1rem;
}

.card-icon {
  width: 40px;
  height: 40px;
  background: #5e6ad2;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 16px;
}

.card-title {
  font-size: 18px;
  font-weight: 500;
  color: #1a1a1a;
  margin-bottom: 6px;
}

.card-sub {
  font-size: 13px;
  color: #888;
  line-height: 1.5;
}

.terms {
  text-align: center;
  font-size: 12px;
  color: #aaa;
  line-height: 1.7;
  margin-top: 20px;
}

.terms a {
  color: #5e6ad2;
  text-decoration: none;
}

.terms a:hover {
  text-decoration: underline;
}

.feature-pills {
  display: flex;
  justify-content: center;
  gap: 8px;
  flex-wrap: wrap;
}

.pill {
  display: flex;
  align-items: center;
  gap: 5px;
  height: 24px;
  padding: 0 10px;
  background: #fff;
  border: 1px solid #e8e8e4;
  font-size: 11.5px;
  color: #666;
}

.pill-dot {
  width: 6px;
  height: 6px;
  background: #5e6ad2;
  flex-shrink: 0;
}
</style>
