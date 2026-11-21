<script setup>
import { onMounted, reactive, ref } from "vue";
import { ElMessage } from "element-plus";
import AdminListPage from "@/components/admin/AdminListPage.vue";
import AdminDataTable from "@/components/admin/AdminDataTable.vue";
import { useAdminList } from "@/composables/useAdminList";
import { getReviewQueue, reviewItem } from "@/api/community";

const props = defineProps({
  kind: { type: String, required: true },
  title: String,
  fields: Array,
});
const status = ref("pending"),
  rows = ref([]),
  saving = ref(false);
const pagination = reactive({ page: 1, pageSize: 20, total: 0 });
const columns = [
  ...props.fields,
  { key: "actions", label: "操作", width: 150 },
];
const { loadList, loading, loadError } = useAdminList({
  fetcher: getReviewQueue,
  rows,
  pagination,
  params: () => ({
    kind: props.kind,
    status: status.value,
    page: pagination.page,
  }),
  selectRows: (data) => data.items || [],
  errorMessage: "审核记录加载失败",
});
function changeStatus(value) {
  status.value = value;
  pagination.page = 1;
  rows.value = [];
  loadList();
}
async function review(row, next) {
  if (saving.value) return;
  saving.value = true;
  try {
    await reviewItem(props.kind, row.id, next);
    ElMessage.success("审核已保存");
    if (pagination.page > 1 && rows.value.length === 1) pagination.page--;
    await loadList();
  } catch (failure) {
    loadError.value = failure.message;
  } finally {
    saving.value = false;
  }
}
onMounted(loadList);
</script>
<template>
  <AdminListPage
    :title="title"
    :columns="columns"
    :preference-key="'review-' + kind"
    :loading="loading"
    :error="loadError"
    :total="pagination.total"
    :page-sizes="[20]"
    v-model:page-size="pagination.pageSize"
    v-model:current-page="pagination.page"
    @refresh="loadList"
    @page-change="loadList"
  >
    <template #actions>
      <el-radio-group
        :model-value="status"
        aria-label="审核状态"
        :disabled="saving"
        @change="changeStatus"
      >
        <el-radio-button value="pending">待审核</el-radio-button>
        <el-radio-button value="approved">已公开</el-radio-button>
        <el-radio-button value="rejected">未通过</el-radio-button>
      </el-radio-group>
    </template>
    <template #default="{ tableProps, tableHeight }">
      <AdminDataTable
        v-bind="tableProps"
        :max-height="tableHeight"
        :data="rows"
        row-key="id"
      >
        <template #actions="{ row }">
          <div v-permission="'blog:community:review'">
            <el-button
              v-if="status !== 'approved'"
              link
              type="primary"
              :disabled="saving || loading"
              @click="review(row, 'approved')"
              >通过</el-button
            >
            <el-button
              v-if="status !== 'rejected'"
              link
              type="danger"
              :disabled="saving || loading"
              @click="review(row, 'rejected')"
              >{{ status === "approved" ? "撤下" : "拒绝" }}</el-button
            >
          </div>
        </template>
      </AdminDataTable>
    </template>
  </AdminListPage>
</template>
