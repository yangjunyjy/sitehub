<template>
  <div class="not-found">
    <div class="content">
      <h1>{{ status }}</h1>
      <p>{{ status === 403 ? '您没有访问此页面的权限' : '抱歉，您访问的页面不存在' }}</p>
      <el-button type="primary" @click="$router.push('/dashboard')">返回首页</el-button>
    </div>
  </div>
</template>

<script setup>
defineProps({ status: { type: Number, default: 404 } })
</script>

<style scoped lang="scss">
.not-found {
  min-height: 60vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f7fa;

  .content {
    text-align: center;

    h1 {
      font-size: 120px;
      color: #409eff;
      margin: 0 0 20px;
    }

    p {
      font-size: 18px;
      color: #303133;
      margin-bottom: 30px;
    }
  }
}
</style>
