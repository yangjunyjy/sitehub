<script setup>
import ResourceIcon from '@shared/icons/ResourceIcon.vue'
import { onBeforeUnmount, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useDownloadFilters } from './useDownloadFilters'
import { useReleaseCatalog } from "./useReleaseCatalog";
import { useWebsiteContent } from "@/composables/useWebsiteContent";

const route = useRoute(),
  router = useRouter();
const { member, downloads, loading, error, downloadError, downloadNotice, checkingPackage, load, downloadPackage } = useReleaseCatalog(route, router);
const copied = ref("");
const copyError = ref("");
let copyTimer;
onBeforeUnmount(() => {
  clearTimeout(copyTimer);
});
const { website } = useWebsiteContent();
const { keyword, filterFields, filters, choices, availableGroups, filteredPackages, catalog, filterActive, setFilter, clearFilters } = useDownloadFilters(downloads, route, router)
const fileSize = (bytes) =>
  bytes > 0
    ? bytes >= 1048576
      ? `${(bytes / 1048576).toFixed(1)} MB`
      : `${Math.ceil(bytes / 1024)} KB`
    : "";

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
</script>

<template>
  <div class="download-page">
    <section class="page-hero">
      <div class="site-shell page-hero__inner">
        <h1>{{ downloads.title || "下载中心" }}</h1>
        <p>{{ downloads.description }}</p>
        <div v-if="downloads.version || downloads.releaseDate" class="release-meta">
          <span v-if="downloads.version"><small>当前版本</small><strong>{{ downloads.version }}</strong></span>
          <span v-if="downloads.releaseDate"><small>发布于</small><time>{{ downloads.releaseDate }}</time></span>
        </div>
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
          <router-link class="site-button site-button--outline" :to="{ name: 'docs', params: { id: '' } }"
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
                  ><ResourceIcon :name="item.icon || 'PKG'" :size="24" /></span>
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
      </div>
    </section>

    <section
      v-if="downloads.installOptions?.length"
      class="site-section install-section"
    >
      <div class="site-shell">
        <header class="section-header">
          <h2 class="site-heading">安装方式</h2>
        </header>
        <p v-if="copyError" class="copy-error" role="status">{{ copyError }}</p>
        <div class="install-grid">
          <article
            v-for="item in downloads.installOptions"
            :key="item.id"
          >
            <header>
              <strong>{{ item.title }}</strong
              ><em v-if="item.badge">{{ item.badge }}</em>
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

<style lang="scss" scoped src="./download.scss"></style>
