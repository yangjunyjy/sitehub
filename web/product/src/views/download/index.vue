<script setup>
import { computed, onMounted, onBeforeUnmount, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
  filterDownloadCatalog,
  downloadGroups,
} from "@/utils/download-catalog.mjs";
import { getDownloads } from "@/api";
import { checkMemberDownload } from "@/api/member";
import { useMember } from "@/composables/useMember";
import { useWebsiteContent } from "@/composables/useWebsiteContent";

const route = useRoute(),
  router = useRouter();
const { member, clearMember } = useMember();
const checkingPackage = ref("");
const downloadError = ref("");
const downloadNotice = ref("");
let disposed = false;

const copied = ref("");
const copyError = ref("");
let copyTimer;
onBeforeUnmount(() => {
  disposed = true;
  clearTimeout(copyTimer);
});
const loading = ref(true);
const error = ref("");
const downloads = ref({
  title: "下载中心",
  groups: [],
  version: "",
  releaseDate: "",
  supportArchitectures: "",
  description: "",
  packages: [],
  installOptions: [],
  requirements: [],
  supportTitle: "",
  supportDescription: "",
});
const { website } = useWebsiteContent();
const keyword = ref(typeof route.query.q === "string" ? route.query.q : "");
watch(
  () => route.query.q,
  (value) => {
    keyword.value = typeof value === "string" ? value : "";
  },
);
const filterFields = [
  { key: "product", label: "产品 / 工具" },
  { key: "edition", label: "版本类型" },
  { key: "platform", label: "运行平台" },
  { key: "arch", label: "处理器架构" },
  { key: "version", label: "版本号" },
];
const filters = computed(() =>
  Object.fromEntries(
    ["group", "q", ...filterFields.map((field) => field.key)].map((key) => [
      key,
      typeof route.query[key] === "string" ? route.query[key] : "",
    ]),
  ),
);
const choices = (key) => [
  ...new Set(downloads.value.packages.map((item) => item[key]).filter(Boolean)),
];
const availableGroups = computed(() =>
  downloadGroups(downloads.value.groups || [], downloads.value.packages),
);
const filteredPackages = computed(() =>
  filterDownloadCatalog(downloads.value.packages, filters.value),
);
const catalog = computed(() =>
  downloadGroups(downloads.value.groups || [], filteredPackages.value),
);
const filterActive = computed(() => Object.values(filters.value).some(Boolean));
const setFilter = (key, value) => {
  const query = { ...route.query };
  if (value) query[key] = value;
  else delete query[key];
  router.replace({ path: route.path, query });
};
const clearFilters = () => {
  keyword.value = "";
  router.replace({ path: route.path });
};
const fileSize = (bytes) =>
  bytes > 0
    ? bytes >= 1048576
      ? `${(bytes / 1048576).toFixed(1)} MB`
      : `${Math.ceil(bytes / 1024)} KB`
    : "";

const load = async () => {
  loading.value = true;
  error.value = "";
  try {
    downloads.value = (await getDownloads()).data;
  } catch (loadError) {
    error.value = loadError.message || "下载信息加载失败";
  } finally {
    loading.value = false;
  }
  if (!error.value && !disposed && typeof route.query.package === "string") {
    const selected = downloads.value.packages.find(
      (item) => item.id === route.query.package && item.url,
    );
    // Consume the return destination once so refresh/back cannot trigger another download.
    const query = { ...route.query };
    delete query.package;
    await router.replace({ path: route.path, query, hash: "#packages" });
    if (selected) await downloadPackage(selected);
    else downloadError.value = "所选安装包不存在或已下架，请选择其他版本。";
  }
};
const downloadPackage = async (item) => {
  if (checkingPackage.value || disposed) return;
  checkingPackage.value = item.id;
  downloadError.value = "";
  downloadNotice.value = "";
  const loginForDownload = () =>
    router.push({
      path: "/login",
      query: {
        redirect: router.resolve({
          path: "/download",
          query: { ...route.query, package: item.id },
        }).fullPath,
      },
    });
  try {
    if (member.ready && !member.user) {
      await loginForDownload();
      return;
    }
    const result = await checkMemberDownload(item.id);
    if (disposed) return;
    // The Node frontend forwards its HttpOnly session as a native Bearer credential.
    const link = document.createElement("a");
    link.href =
      "/v1/file/download?" + new URLSearchParams({ id: item.id });
    link.download = result.fileName || "";
    document.body.appendChild(link);
    link.click();
    link.remove();
    downloadNotice.value = "已发起下载，可在浏览器的下载列表中查看。";
  } catch (failure) {
    if (disposed) return;
    if (failure.status === 401) {
      clearMember();
      await loginForDownload();
    } else {
      downloadError.value = failure.message || "暂时无法下载，请稍后重试。";
    }
  } finally {
    checkingPackage.value = "";
  }
};
const copy = async (item) => {
  copyError.value = "";
  try {
    await navigator.clipboard.writeText(item.command);
    copied.value = item.id;
    clearTimeout(copyTimer);
    copyTimer = setTimeout(() => {
      copied.value = "";
    }, 1600);
  } catch {
    copyError.value = "未能访问剪贴板，请选中命令手动复制。";
  }
};
onMounted(load);
</script>

<template>
  <div class="download-page">
    <section class="page-hero">
      <div class="site-shell page-hero__inner">
        <h1>{{ downloads.title || "下载中心" }}</h1>
        <p>{{ downloads.description }}</p>
        <p class="catalog-overview">
          按产品、部署方式或运行平台，选择需要的版本和工具。
        </p>
      </div>
    </section>

    <section id="packages" class="site-section">
      <div class="site-shell">
        <div class="catalog-toolbar">
          <nav class="download-groups" aria-label="下载分组">
            <button
              type="button"
              :class="{ active: !filters.group }"
              :aria-pressed="!filters.group"
              @click="setFilter('group', '')"
            >
              全部下载
            </button>
            <button
              v-for="group in availableGroups"
              :key="group.id"
              type="button"
              :class="{ active: filters.group === group.id }"
              :aria-pressed="filters.group === group.id"
              @click="setFilter('group', group.id)"
            >
              {{ group.name }} <small>{{ group.packages.length }}</small>
            </button>
          </nav>
          <form
            class="catalog-filters"
            @submit.prevent="setFilter('q', keyword.trim())"
          >
            <label v-for="field in filterFields" :key="field.key"
              >{{ field.label
              }}<select
                :value="filters[field.key]"
                @change="setFilter(field.key, $event.target.value)"
              >
                <option value="">全部</option>
                <option
                  v-for="value in choices(field.key)"
                  :key="value"
                  :value="value"
                >
                  {{ value }}
                </option>
              </select></label
            >
            <label class="catalog-search"
              >查找文件
              <div>
                <input
                  v-model="keyword"
                  type="search"
                  placeholder="名称、版本或文件名"
                /><button type="submit">搜索</button>
              </div></label
            >
          </form>
          <div class="catalog-count">
            <span>{{
              loading
                ? "正在读取下载目录…"
                : `共 ${filteredPackages.length} 个文件`
            }}</span
            ><button v-if="filterActive" type="button" @click="clearFilters">
              清除筛选</button
            ><span>登录官网账号后下载</span>
          </div>
        </div>
        <p
          v-if="downloadError"
          class="download-message download-message--error"
          role="alert"
        >
          {{ downloadError }}
        </p>
        <p v-if="downloadNotice" class="download-message" role="status">
          {{ downloadNotice }}
        </p>
        <div v-if="error" class="empty-state">
          <h3>下载服务暂不可用</h3>
          <p>{{ error }}</p>
          <button
            class="site-button site-button--primary site-button--small"
            @click="load"
          >
            重新加载
          </button>
        </div>
        <div v-else-if="loading" class="empty-state" role="status">
          正在获取最新版本…
        </div>
        <div v-else-if="!downloads.packages.length" class="empty-state">
          <h3>新版本正在准备中</h3>
          <p>你可以先阅读部署文档，了解环境要求。</p>
          <router-link class="site-button site-button--outline" to="/docs"
            >阅读部署文档
          </router-link>
        </div>
        <div v-else-if="!filteredPackages.length" class="empty-state">
          <h3>没有匹配的文件</h3>
          <p>调整分组或筛选条件后再试。</p>
          <button
            type="button"
            class="site-button site-button--outline"
            @click="clearFilters"
          >
            清除筛选
          </button>
        </div>
        <div v-else class="download-catalog">
          <section
            v-for="group in catalog"
            :key="group.id"
            class="catalog-group"
          >
            <header>
              <h2>{{ group.name }}</h2>
              <p v-if="group.description">{{ group.description }}</p>
            </header>
            <div class="download-list">
              <article v-for="item in group.packages" :key="item.id">
                <span class="download-icon" aria-hidden="true"
                  ><svg viewBox="0 0 24 24">
                    <path d="M6 3h8l4 4v14H6zM14 3v5h4M9 12h6M9 16h6" /></svg
                ></span>
                <div>
                  <h3>
                    {{ item.product || website.content.brand.name }}
                    <span v-if="item.edition" class="edition-label">{{
                      item.edition
                    }}</span>
                  </h3>
                  <p class="package-meta">
                    {{
                      [
                        item.version,
                        item.platform,
                        item.arch,
                        item.format,
                        fileSize(item.size),
                        item.releaseDate,
                      ]
                        .filter(Boolean)
                        .join(" \u00b7 ")
                    }}
                  </p>
                  <p v-if="item.description" class="package-description">
                    {{ item.description }}
                  </p>
                  <details
                    v-if="item.checksum || item.fileName"
                    class="package-details"
                  >
                    <summary>文件详情</summary>
                    <p v-if="item.fileName">{{ item.fileName }}</p>
                    <code v-if="item.checksum">SHA256 {{ item.checksum }}</code>
                  </details>
                </div>
                <button
                  v-if="item.url"
                  class="site-button site-button--primary site-button--small"
                  type="button"
                  :disabled="Boolean(checkingPackage)"
                  @click="downloadPackage(item)"
                >
                  {{
                    checkingPackage === item.id
                      ? "正在准备…"
                      : member.ready && !member.user
                        ? "登录后下载"
                        : "下载文件"
                  }}</button
                ><button v-else class="pending-button" type="button" disabled>
                  {{ loading ? "加载中" : "待发布" }}
                </button>
              </article>
            </div>
          </section>
        </div>
        <div
          v-if="downloads.packages.some((item) => item.checksum)"
          class="checksum-note"
        >
          <strong>供应链安全提示</strong>
          <p>
            生产环境安装前请核对安装包的 SHA256
            校验值，并始终从本页提供的可信地址下载。
          </p>
        </div>
      </div>
    </section>

    <section
      v-if="downloads.installOptions?.length"
      class="site-section install-section"
    >
      <div class="site-shell">
        <header class="section-header">
          <div>
            <p class="site-kicker">Install options</p>
          </div>
          <h2 class="site-heading">匹配不同团队的部署方式</h2>
          <p>
            从快速体验到离线生产环境，安装方式保持一致的配置模型和升级路径。
          </p>
        </header>
        <p v-if="copyError" class="copy-error" role="status">{{ copyError }}</p>
        <div class="install-grid">
          <article
            v-for="(item, index) in downloads.installOptions"
            :key="item.id"
          >
            <header>
              <span>0{{ index + 1 }}</span
              ><strong>{{ item.title }}</strong
              ><em>{{ item.badge }}</em>
            </header>
            <p>{{ item.description }}</p>
            <code>{{ item.command }}</code
            ><button @click="copy(item)">
              {{ copied === item.id ? "已复制" : "复制命令" }}
            </button>
          </article>
        </div>
      </div>
    </section>

    <section
      v-if="downloads.requirements?.length"
      class="site-section requirements"
    >
      <div class="site-shell requirements__layout">
        <div>
          <p class="site-kicker">Requirements</p>
          <h2 class="site-heading">部署前确认</h2>
          <p class="site-lead">环境基线随发布清单更新，请在安装前核对。</p>
        </div>
        <dl>
          <div v-for="item in downloads.requirements" :key="item.label">
            <dt>{{ item.label }}</dt>
            <dd>{{ item.value }}</dd>
          </div>
        </dl>
      </div>
    </section>

    <section v-if="downloads.supportTitle" class="support-banner">
      <div class="site-shell">
        <div>
          <p class="site-kicker">Enterprise support</p>
          <h2>{{ downloads.supportTitle }}</h2>
          <p>{{ downloads.supportDescription }}</p>
        </div>
        <a
          v-if="website.content.brand.contactEmail"
          class="site-button site-button--primary"
          :href="`mailto:${website.content.brand.contactEmail}?subject=Deploy 企业版咨询`"
          >联系企业支持 </a
        ><span v-else class="support-pending">更多部署帮助请查看文档</span>
      </div>
    </section>
  </div>
</template>

<style lang="scss" scoped>
.download-page > .site-section:first-of-type {
  padding-top: 28px;
}
.catalog-overview {
  margin-top: 8px !important;
  font-size: 14px !important;
}
.catalog-toolbar {
  margin-bottom: 26px;
  padding: 20px 22px 0;
  background: var(--deploy-white);
  border: 1px solid var(--deploy-line);
  border-radius: 10px;
}
.download-groups {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  padding-bottom: 18px;
  border-bottom: 1px solid var(--deploy-line);
}
.download-groups button {
  min-height: 38px;
  padding: 7px 14px;
  border: 0;
  border-radius: 6px;
  background: transparent;
  color: var(--deploy-muted);
  font-size: 14px;
}
.download-groups button.active {
  background: var(--deploy-soft);
  color: var(--deploy-primary);
}
.download-groups small {
  margin-left: 6px;
  font-size: 11px;
}
.catalog-filters {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr)) minmax(180px, 1.5fr);
  gap: 12px;
  padding-top: 18px;
}
.catalog-filters label {
  min-width: 0;
  font-size: 12px;
  color: var(--deploy-muted);
}
.catalog-filters select,
.catalog-search > div {
  display: flex;
  width: 100%;
  height: 38px;
  margin-top: 6px;
  border: 1px solid var(--deploy-line);
  border-radius: 6px;
  background: var(--deploy-white);
  color: var(--deploy-ink);
  font: inherit;
}
.catalog-filters select {
  padding: 0 8px;
}
.catalog-search input {
  min-width: 0;
  width: 100%;
  padding: 0 10px;
  border: 0;
  background: transparent;
  font-size: 13px;
}
.catalog-search button,
.catalog-count button {
  flex-shrink: 0;
  border: 0;
  background: transparent;
  color: var(--deploy-primary);
  padding: 0 10px;
  font-size: 13px;
}
.catalog-count {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  padding: 16px 0;
  color: var(--deploy-muted);
  font-size: 12px;
}
.catalog-count > span:last-child {
  margin-left: auto;
}
.catalog-group + .catalog-group {
  margin-top: 32px;
}
.catalog-group > header {
  margin-bottom: 14px;
}
.catalog-group h2 {
  margin: 0;
  font-size: 19px;
  font-weight: 600;
}
.catalog-group header p {
  margin: 4px 0 0;
  color: var(--deploy-muted);
  font-size: 14px;
}
.edition-label {
  display: inline-block;
  margin-left: 8px;
  padding: 2px 8px;
  border-radius: 4px;
  background: var(--deploy-soft);
  color: var(--deploy-primary);
  font-size: 12px;
  font-weight: 500;
  vertical-align: middle;
}
.download-icon svg {
  width: 24px;
  height: 24px;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.5;
}
.package-description {
  margin-top: 6px !important;
}
.package-details {
  margin-top: 8px;
  color: var(--deploy-muted);
  font-size: 12px;
}
.package-details summary {
  cursor: pointer;
  list-style: none;
  width: fit-content;
  text-decoration: underline;
  text-underline-offset: 3px;
}
.package-details summary::-webkit-details-marker {
  display: none;
}
.package-details code {
  overflow-wrap: anywhere;
}
@media (max-width: 1100px) {
  .catalog-filters {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
@media (max-width: 600px) {
  .catalog-filters {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
  .catalog-toolbar {
    padding: 16px 14px 0;
  }
}

.download-message {
  padding: 12px 16px;
  border: 1px solid var(--deploy-line);
  border-radius: 10px;
  background: var(--deploy-soft);
  color: var(--deploy-muted);
  font-size: 14px;
}
.download-message--error {
  background: var(--deploy-paper);
  border-color: #f4dcd6;
  color: #a44f43;
}
.release-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  margin-top: 32px;
}
.release-meta > span {
  min-width: 160px;
  padding: 14px 20px;
  border: 1px solid var(--deploy-line);
  border-radius: 12px;
  background: var(--deploy-soft);
}
.release-meta small,
.release-meta strong {
  display: block;
}
.release-meta small {
  color: var(--deploy-muted);
  font-size: 12px;
}
.release-meta strong {
  margin-top: 5px;
  color: var(--deploy-primary);
  font-size: 15px;
  font-weight: 600;
}
.download-heading {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 40px;
  margin-bottom: 32px;
}
.download-heading > p {
  max-width: 440px;
  margin: 0;
  color: var(--deploy-muted);
  font-size: 15px;
}
.download-list {
  display: grid;
  gap: 14px;
}
.download-list article {
  display: grid;
  grid-template-columns: 56px minmax(0, 1fr) auto;
  align-items: center;
  gap: 22px;
  padding: 20px;
  border: 1px solid var(--deploy-line);
  border-radius: 10px;
  background: var(--deploy-white);
}
.download-icon {
  display: grid;
  width: 54px;
  height: 54px;
  place-items: center;
  border-radius: 8px;
  background: var(--deploy-soft);
  color: var(--deploy-primary);
  font-size: 15px;
}
.download-list h3 {
  margin: 0 0 7px;
  font-size: 18px;
}
.download-list p {
  margin: 0;
  color: var(--deploy-muted);
  font-size: 13px;
  overflow-wrap: anywhere;
}
.download-list .site-button {
  white-space: nowrap;
}
.pending-button {
  min-width: 122px;
  min-height: 42px;
  border: 1px solid var(--deploy-line);
  border-radius: 8px;
  background: var(--deploy-paper);
  color: var(--deploy-muted);
  font-size: 13px;
}
.checksum-note {
  display: grid;
  grid-template-columns: 150px 1fr;
  gap: 20px;
  margin-top: 24px;
  padding: 20px 24px;
  border: 1px solid var(--deploy-line);
  border-radius: 12px;
  background: var(--deploy-paper);
}
.checksum-note strong {
  color: var(--deploy-primary);
  font-size: 14px;
}
.checksum-note p {
  margin: 0;
  color: var(--deploy-muted);
  font-size: 13px;
  line-height: 1.8;
}
.install-section {
  background: var(--deploy-white);
}
.install-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 20px;
}
.install-grid article {
  min-width: 0;
  padding: 26px;
  border: 1px solid var(--deploy-line);
  border-radius: 8px;
  background: var(--deploy-white);
}
.install-grid header {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  align-items: center;
}
.install-grid header span {
  color: var(--deploy-muted);
  font-size: 12px;
}
.install-grid header strong {
  font-size: 17px;
}
.install-grid header em {
  margin-left: auto;
  padding: 3px 8px;
  border-radius: 5px;
  background: var(--deploy-paper);
  color: var(--deploy-primary);
  font-size: 11px;
  font-style: normal;
}
.install-grid p {
  min-height: 60px;
  margin: 24px 0 18px;
  color: var(--deploy-muted);
  font-size: 14px;
}
.install-grid code {
  display: block;
  overflow-x: auto;
  padding: 14px;
  border: 1px solid var(--deploy-line);
  border-radius: 8px;
  background: var(--deploy-paper);
  color: var(--deploy-muted);
  font-size: 12px;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}
.install-grid button {
  min-height: 44px;
  margin-top: 12px;
  padding-inline: 12px;
  border: 1px solid var(--deploy-line);
  border-radius: 8px;
  background: var(--deploy-white);
  color: var(--deploy-primary);
  font-size: 13px;
}
.copy-error {
  color: #af6a41;
}
.requirements__layout {
  display: grid;
  grid-template-columns: 1fr 1.2fr;
  gap: 80px;
}
.requirements dl {
  display: grid;
  grid-template-columns: 1fr 1fr;
  margin: 0;
  overflow: hidden;
  border: 1px solid var(--deploy-line);
  border-radius: 8px;
  background: var(--deploy-white);
}
.requirements dl div {
  padding: 24px;
  border-bottom: 1px solid var(--deploy-line);
}
.requirements dl div:nth-child(even) {
  border-left: 1px solid var(--deploy-line);
}
.requirements dt {
  color: var(--deploy-muted);
  font-size: 13px;
}
.requirements dd {
  margin: 8px 0 0;
  font-size: 14px;
}
.support-banner {
  padding: 60px 0;
  border-top: 1px solid var(--deploy-line);
  background: var(--deploy-paper);
}
.support-banner > div {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 48px;
}
.support-banner h2 {
  margin: 12px 0;
  font-size: 26px;
}
.support-banner p:not(.site-kicker) {
  max-width: 700px;
  margin: 0;
  color: var(--deploy-muted);
  font-size: 15px;
}
.support-banner .site-button {
  flex-shrink: 0;
}
.support-pending {
  color: var(--deploy-muted);
  font-size: 13px;
}
@media (max-width: 900px) {
  .install-grid {
    grid-template-columns: 1fr;
  }
  .requirements__layout {
    grid-template-columns: 1fr;
    gap: 32px;
  }
}
@media (max-width: 760px) {
  .download-heading,
  .support-banner > div {
    align-items: flex-start;
    flex-direction: column;
    gap: 24px;
  }
  .download-list article {
    grid-template-columns: 48px minmax(0, 1fr);
    gap: 16px;
    padding: 20px;
  }
  .download-icon {
    width: 46px;
    height: 46px;
  }
  .download-list article > a,
  .download-list article > button {
    grid-column: 1 / -1;
  }
  .release-meta {
    gap: 10px;
  }
  .release-meta > span {
    flex: 1 1 130px;
    min-width: 0;
    padding: 12px 16px;
  }
}
@media (max-width: 500px) {
  .checksum-note {
    grid-template-columns: 1fr;
    gap: 8px;
  }
  .requirements dl {
    grid-template-columns: 1fr;
  }
  .requirements dl div:nth-child(even) {
    border-left: 0;
  }
}
</style>
