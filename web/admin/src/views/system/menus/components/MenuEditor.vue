<script setup>
import { ref } from "vue";
import AdminFormDialog from "@/components/admin/AdminFormDialog.vue";
import IconPicker from "@/components/icons/IconPicker.vue";

const dialogVisible = defineModel({ type: Boolean });
defineProps({
  dialogTitle: String,
  submitting: Boolean,
  form: Object,
  rules: Object,
  systems: Array,
  menuTreeSelectData: Array,
  isExternal: Boolean,
});
defineEmits(["submit", "system-change"]);
const formRef = ref(null);
defineExpose({ validate: (callback) => formRef.value?.validate(callback) });
</script>
<template>
  <AdminFormDialog
    :loading="submitting"
    confirm-text="确定"
    @submit="$emit('submit')"
    v-model="dialogVisible"
    :title="dialogTitle"
    class="menu-editor-dialog"
    width="min(760px, calc(100vw - 32px))"
    top="6vh"
  >
    <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
      <el-form-item label="所属系统" prop="systemCode"
        ><el-select v-model="form.systemCode" @change="$emit('system-change')"
          ><el-option
            v-for="system in systems"
            :key="system.code"
            :value="system.code"
            :label="system.name" /></el-select
      ></el-form-item>
      <el-form-item label="菜单类型" prop="type">
        <el-radio-group v-model="form.type">
          <el-radio :value="1">目录</el-radio>
          <el-radio :value="2">菜单</el-radio>
          <el-radio :value="3">按钮</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="上级菜单" prop="parent_id">
        <el-tree-select
          v-model="form.parent_id"
          :data="menuTreeSelectData"
          :props="{ label: 'menu_name', value: 'id' }"
          clearable
          placeholder="请选择上级菜单"
          check-strictly
        />
      </el-form-item>
      <el-form-item label="菜单名称" prop="name">
        <el-input v-model="form.name" placeholder="请输入菜单名称" />
      </el-form-item>
      <el-form-item label="图标" prop="icon" v-if="form.type !== 3">
        <IconPicker v-model="form.icon" />
      </el-form-item>
      <el-form-item v-if="form.type === 2" label="页面类型">
        <el-radio-group v-model="form.pageType">
          <el-radio value="internal">内部页面</el-radio
          ><el-radio value="external">外链页面</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item
        v-if="isExternal"
        label="链接地址"
        prop="externalUrl"
        required
      >
        <el-input
          v-model.trim="form.externalUrl"
          maxlength="2048"
          placeholder="https://example.com"
        />
      </el-form-item>
      <el-form-item
        label="路径"
        prop="path"
        v-if="form.type !== 3 && !isExternal"
      >
        <el-input v-model="form.path" placeholder="请输入路由地址" />
      </el-form-item>
      <el-form-item
        label="组件路径"
        prop="component"
        v-if="form.type === 2 && !isExternal"
      >
        <el-input
          v-model="form.component"
          placeholder="例如 system/users/index（相对 views）"
        />
      </el-form-item>
      <el-form-item label="是否隐藏" prop="isVisible">
        <el-radio-group v-model="form.isVisible">
          <el-radio :value="true">否</el-radio>
          <el-radio :value="false">是</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="菜单编码" prop="code">
        <el-input
          v-model="form.code"
          placeholder="菜单唯一编码，例如 admin-system-users"
        />
      </el-form-item>
      <el-form-item
        label="权限标识"
        prop="permission"
        v-if="form.type !== 1"
        :required="form.type === 3"
      >
        <el-input
          v-model="form.permission"
          placeholder="例如 system:user:list；页面可留空"
        />
      </el-form-item>
      <el-form-item label="菜单描述" prop="description">
        <el-input
          v-model="form.description"
          placeholder="请输入描述信息"
          type="textarea"
          :rows="3"
        />
      </el-form-item>
      <el-form-item label="排序" prop="sort">
        <el-input-number v-model="form.sort" :min="0" :max="999" />
      </el-form-item>
      <el-form-item label="状态" prop="status">
        <el-radio-group v-model="form.status">
          <el-radio :value="1">启用</el-radio>
          <el-radio :value="0">禁用</el-radio>
        </el-radio-group>
      </el-form-item>
    </el-form>
  </AdminFormDialog>
</template>
