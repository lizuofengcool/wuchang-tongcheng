<template>
  <div class="profile app-container">
    <el-row :gutter="20">
      <el-col :xs="24" :md="8">
        <el-card shadow="never" class="info-card">
          <div class="avatar-wrap">
            <el-avatar :size="100" :src="form.avatar">
              {{ (form.nickname || form.username || '?').charAt(0) }}
            </el-avatar>
            <h3>{{ form.nickname || form.username }}</h3>
            <p>{{ form.username }}</p>
          </div>
        </el-card>
      </el-col>

      <el-col :xs="24" :md="16">
        <el-card shadow="never">
          <template #header>
            <div class="card-header">
              <span>个人资料</span>
            </div>
          </template>
          <el-form
            ref="profileRef"
            :model="form"
            :rules="profileRules"
            label-width="80px"
          >
            <el-form-item label="用户名">
              <el-input v-model="form.username" disabled />
            </el-form-item>
            <el-form-item label="昵称" prop="nickname">
              <el-input v-model="form.nickname" placeholder="请输入昵称" />
            </el-form-item>
            <el-form-item label="手机号" prop="phone">
              <el-input v-model="form.phone" placeholder="请输入手机号" />
            </el-form-item>
            <el-form-item label="邮箱" prop="email">
              <el-input v-model="form.email" placeholder="请输入邮箱" />
            </el-form-item>
            <el-form-item label="头像" prop="avatar">
              <el-input v-model="form.avatar" placeholder="头像URL" />
            </el-form-item>
            <el-form-item label="性别" prop="gender">
              <el-radio-group v-model="form.gender">
                <el-radio :value="0">未知</el-radio>
                <el-radio :value="1">男</el-radio>
                <el-radio :value="2">女</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="saving" @click="handleSaveProfile">
                保存
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <el-card shadow="never" class="pwd-card">
          <template #header>
            <div class="card-header">
              <span>修改密码</span>
            </div>
          </template>
          <el-form
            ref="pwdRef"
            :model="pwdForm"
            :rules="pwdRules"
            label-width="100px"
          >
            <el-form-item label="原密码" prop="old_password">
              <el-input v-model="pwdForm.old_password" type="password" show-password />
            </el-form-item>
            <el-form-item label="新密码" prop="new_password">
              <el-input v-model="pwdForm.new_password" type="password" show-password />
            </el-form-item>
            <el-form-item label="确认密码" prop="confirm_password">
              <el-input v-model="pwdForm.confirm_password" type="password" show-password />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="changingPwd" @click="handleChangePwd">
                修改密码
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { updateProfile, changePassword } from '@/api/user'

const userStore = useUserStore()
const profileRef = ref(null)
const pwdRef = ref(null)
const saving = ref(false)
const changingPwd = ref(false)

const form = reactive({
  username: '',
  nickname: '',
  phone: '',
  email: '',
  avatar: '',
  gender: 0
})

const pwdForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

const profileRules = {
  nickname: [{ max: 50, message: '昵称不能超过50个字符', trigger: 'blur' }],
  phone: [{ pattern: /^1\d{10}$/, message: '请输入正确的手机号', trigger: 'blur' }],
  email: [{ type: 'email', message: '请输入正确的邮箱', trigger: 'blur' }]
}

const validateConfirm = (rule, value, callback) => {
  if (value !== pwdForm.new_password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const pwdRules = {
  old_password: [{ required: true, message: '请输入原密码', trigger: 'blur' }],
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, max: 50, message: '密码长度 6-50 位', trigger: 'blur' }
  ],
  confirm_password: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    { validator: validateConfirm, trigger: 'blur' }
  ]
}

const initForm = () => {
  const u = userStore.userInfo || {}
  form.username = u.username || ''
  form.nickname = u.nickname || ''
  form.phone = u.phone || ''
  form.email = u.email || ''
  form.avatar = u.avatar || ''
  form.gender = u.gender ?? 0
}

const handleSaveProfile = async () => {
  if (!profileRef.value) return
  await profileRef.value.validate(async (valid) => {
    if (!valid) return
    saving.value = true
    try {
      await updateProfile({
        nickname: form.nickname,
        phone: form.phone,
        email: form.email,
        avatar: form.avatar,
        gender: form.gender
      })
      ElMessage.success('保存成功')
      await userStore.fetchProfile()
      initForm()
    } catch (e) {
      // ignore
    } finally {
      saving.value = false
    }
  })
}

const handleChangePwd = async () => {
  if (!pwdRef.value) return
  await pwdRef.value.validate(async (valid) => {
    if (!valid) return
    changingPwd.value = true
    try {
      await changePassword({
        old_password: pwdForm.old_password,
        new_password: pwdForm.new_password
      })
      ElMessage.success('密码修改成功，请重新登录')
      setTimeout(() => {
        userStore.logout()
        location.href = '/login'
      }, 1500)
    } catch (e) {
      // ignore
    } finally {
      changingPwd.value = false
    }
  })
}

onMounted(async () => {
  initForm()
  try {
    await userStore.fetchProfile()
    initForm()
  } catch (e) {
    // ignore
  }
})
</script>

<style scoped>
.info-card {
  text-align: center;
}
.avatar-wrap {
  padding: 20px 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}
.avatar-wrap h3 {
  margin-top: 12px;
  color: #303133;
}
.avatar-wrap p {
  color: #909399;
  font-size: 13px;
}
.card-header {
  font-weight: 600;
}
.pwd-card {
  margin-top: 20px;
}
</style>
