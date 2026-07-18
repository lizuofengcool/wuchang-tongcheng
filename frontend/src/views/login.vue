<template>
  <div class="login-container">
    <div class="login-box">
      <div class="login-header">
        <h1>近享同城</h1>
        <p>本地生活服务管理后台</p>
      </div>
      <el-tabs v-model="activeTab" class="login-tabs" stretch>
        <el-tab-pane label="密码登录" name="password">
          <el-form
            ref="pwdFormRef"
            :model="pwdForm"
            :rules="pwdRules"
            class="login-form"
            size="large"
            @keyup.enter="handlePasswordLogin"
          >
            <el-form-item prop="username">
              <el-input
                v-model="pwdForm.username"
                placeholder="请输入用户名"
                :prefix-icon="User"
                clearable
              />
            </el-form-item>
            <el-form-item prop="password">
              <el-input
                v-model="pwdForm.password"
                type="password"
                placeholder="请输入密码"
                :prefix-icon="Lock"
                show-password
                clearable
              />
            </el-form-item>
            <el-form-item>
              <el-button
                type="primary"
                class="login-btn"
                :loading="loading"
                @click="handlePasswordLogin"
              >
                登 录
              </el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <el-tab-pane label="短信登录" name="sms">
          <el-form
            ref="smsFormRef"
            :model="smsForm"
            :rules="smsRules"
            class="login-form"
            size="large"
            @keyup.enter="handleSmsLogin"
          >
            <el-form-item prop="phone">
              <el-input
                v-model="smsForm.phone"
                placeholder="请输入手机号"
                :prefix-icon="Iphone"
                clearable
                maxlength="11"
              />
            </el-form-item>
            <el-form-item prop="code">
              <div class="sms-code-row">
                <el-input
                  v-model="smsForm.code"
                  placeholder="请输入验证码"
                  :prefix-icon="Key"
                  clearable
                  maxlength="6"
                />
                <el-button
                  type="primary"
                  plain
                  :disabled="smsCountdown > 0 || sendingCode"
                  :loading="sendingCode"
                  class="sms-code-btn"
                  @click="handleSendCode"
                >
                  {{ smsCountdown > 0 ? `${smsCountdown}s 后重发` : '获取验证码' }}
                </el-button>
              </div>
            </el-form-item>
            <el-form-item>
              <el-button
                type="primary"
                class="login-btn"
                :loading="loading"
                @click="handleSmsLogin"
              >
                登 录
              </el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>
      <div class="login-footer">
        <span>© {{ year }} 近享同城 · 管理后台</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onBeforeUnmount } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock, Iphone, Key } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { sendSmsCode } from '@/api/user'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const activeTab = ref('password')
const loading = ref(false)
const sendingCode = ref(false)
const smsCountdown = ref(0)
let countdownTimer = null

const pwdFormRef = ref(null)
const smsFormRef = ref(null)

const pwdForm = reactive({
  username: '',
  password: ''
})
const smsForm = reactive({
  phone: '',
  code: ''
})

// 手机号格式校验：中国大陆 11 位、1 开头
const phoneValidator = (rule, value, callback) => {
  if (!value) {
    callback(new Error('请输入手机号'))
    return
  }
  if (!/^1[3-9]\d{9}$/.test(value)) {
    callback(new Error('请输入正确的手机号'))
    return
  }
  callback()
}

const pwdRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 50, message: '密码长度 6-50 位', trigger: 'blur' }
  ]
}
const smsRules = {
  phone: [{ required: true, validator: phoneValidator, trigger: 'blur' }],
  code: [
    { required: true, message: '请输入验证码', trigger: 'blur' },
    { min: 4, max: 6, message: '验证码长度 4-6 位', trigger: 'blur' }
  ]
}

const year = computed(() => new Date().getFullYear())

// 启动验证码倒计时
function startCountdown() {
  smsCountdown.value = 60
  countdownTimer = setInterval(() => {
    smsCountdown.value -= 1
    if (smsCountdown.value <= 0) {
      clearInterval(countdownTimer)
      countdownTimer = null
    }
  }, 1000)
}

onBeforeUnmount(() => {
  if (countdownTimer) {
    clearInterval(countdownTimer)
    countdownTimer = null
  }
})

// 跳转登录后页面
function redirectAfterLogin() {
  const redirect = route.query.redirect || '/'
  router.push(redirect)
}

// 密码登录
const handlePasswordLogin = async () => {
  if (!pwdFormRef.value) return
  await pwdFormRef.value.validate(async (valid) => {
    if (!valid) return
    loading.value = true
    try {
      await userStore.login({ username: pwdForm.username, password: pwdForm.password })
      ElMessage.success('登录成功')
      redirectAfterLogin()
    } catch (e) {
      // 错误已在拦截器中提示
    } finally {
      loading.value = false
    }
  })
}

// 发送短信验证码
const handleSendCode = async () => {
  if (!smsFormRef.value) return
  // 仅校验手机号字段，避免 code 未填阻断发送
  try {
    await smsFormRef.value.validateField('phone')
  } catch (e) {
    return
  }
  sendingCode.value = true
  try {
    const res = await sendSmsCode(smsForm.phone)
    // mock + dev_return_code=true 时后端回 dev_code 明文，联调自动回填方便测试
    const devCode = res?.data?.dev_code
    if (devCode) {
      smsForm.code = devCode
      ElMessage.success(`验证码已发送（联调回填：${devCode}）`)
    } else {
      ElMessage.success('验证码已发送，请注意查收')
    }
    startCountdown()
  } catch (e) {
    // 错误已在拦截器中提示
  } finally {
    sendingCode.value = false
  }
}

// 短信验证码登录
const handleSmsLogin = async () => {
  if (!smsFormRef.value) return
  await smsFormRef.value.validate(async (valid) => {
    if (!valid) return
    loading.value = true
    try {
      await userStore.loginBySms({ phone: smsForm.phone, code: smsForm.code })
      ElMessage.success('登录成功')
      redirectAfterLogin()
    } catch (e) {
      // 错误已在拦截器中提示
    } finally {
      loading.value = false
    }
  })
}
</script>

<style scoped>
.login-container {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-box {
  width: 400px;
  padding: 40px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15);
}

.login-header {
  text-align: center;
  margin-bottom: 24px;
}

.login-header h1 {
  font-size: 28px;
  color: #303133;
  margin-bottom: 8px;
}

.login-header p {
  color: #909399;
  font-size: 14px;
}

.login-tabs {
  margin-bottom: 8px;
}

.login-form {
  margin-bottom: 16px;
}

.login-btn {
  width: 100%;
}

.sms-code-row {
  display: flex;
  gap: 8px;
  width: 100%;
}

.sms-code-row .el-input {
  flex: 1;
}

.sms-code-btn {
  flex-shrink: 0;
  min-width: 120px;
}

.login-footer {
  text-align: center;
  color: #c0c4cc;
  font-size: 12px;
  margin-top: 8px;
}
</style>
