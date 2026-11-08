<script setup>
import DocumentModeSwitch from "@/components/product/DocumentModeSwitch.vue";
import ReadingToc from "@/components/product/ReadingToc.vue";
import ContentState from "@/components/product/ContentState.vue";
import MarkdownContent from "@/components/product/MarkdownContent.vue";
import ProductIcon from "@/components/product/ProductIcon.vue";
import { useDocumentReader } from "@/composables/useDocumentReader";

const {
  route,
  documentModes,
  modesReady,
  modesError,
  currentMode,
  readableMode,
  modeLocation,
  navigation,
  document,
  html,
  toc,
  titleHeading,
  search,
  loading,
  error,
  errorKind,
  sidebarOpen,
  menuButton,
  copyStatus,
  versionNotice,
  docLocation,
  switchVersion,
  filteredGroups,
  currentGroup,
  previousDocument,
  nextDocument,
  hasToc,
  closeMenu,
  loadModes,
  retry,
  selectDocument,
  copyMarkdown,
} = useDocumentReader();
</script>

<template>
  <div class="docs-page">
    <header v-if="!currentMode.available || !modesReady" class="docs-overview">
      <div>
        <p class="docs-site-title">{{ navigation.title }}</p>
        <p class="docs-description">选择工作模式，查阅对应的使用文档。</p>
      </div>
      <DocumentModeSwitch
        v-if="modesReady"
        :modes="documentModes"
        :current="currentMode"
        :location="modeLocation"
      />
    </header>
    <div v-if="currentMode.available" class="docs-mobile-bar">
      <button
        ref="menuButton"
        type="button"
        :aria-expanded="sidebarOpen"
        aria-controls="docs-sidebar"
        @click="sidebarOpen = !sidebarOpen"
      >
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path d="M4 6h16M4 12h16M4 18h16" /></svg
        >{{ sidebarOpen ? "关闭目录" : "文档目录" }}
      </button>
      <span>{{ document?.title || navigation.title }}</span>
    </div>
    <button
      v-if="sidebarOpen"
      class="docs-backdrop"
      type="button"
      aria-label="关闭文档目录"
      @click="closeMenu(true)"
    />
    <p v-if="versionNotice" class="version-notice" role="status">
      {{ versionNotice }}
    </p>
    <section
      v-if="modesReady && !currentMode.available"
      class="docs-coming-soon"
      aria-labelledby="upcoming-mode-title"
    >
      <span class="docs-upcoming-icon"
        ><ProductIcon :name="currentMode.icon"
      /></span>
      <span class="docs-upcoming-label">文档待推出</span>
      <h1 id="upcoming-mode-title">{{ currentMode.label }}文档</h1>
      <p>此模式的使用文档尚未发布，可以先阅读其他已上线的模式文档。</p>
      <router-link
        v-if="readableMode"
        :to="modeLocation(readableMode)"
        class="site-button site-button--primary"
        >阅读 {{ readableMode.label }}文档
      </router-link>
    </section>
    <ContentState
      v-else-if="!modesReady"
      :title="modesError ? '文档模式加载失败' : '正在加载文档模式'"
      :message="modesError"
      :error="!!modesError"
      :action-label="modesError ? '重新加载' : ''"
      @retry="loadModes"
    />
    <div v-else class="docs-layout" :class="{ 'without-toc': !hasToc }">
      <aside
        id="docs-sidebar"
        class="docs-sidebar"
        :class="{ open: sidebarOpen }"
      >
        <div class="docs-sidebar-controls">
          <DocumentModeSwitch
            :modes="documentModes"
            :current="currentMode"
            :location="modeLocation"
          />
          <div class="docs-sidebar-head">
            <router-link :to="docLocation()"
              ><span class="docs-brand-icon"
                ><ProductIcon name="playbook" /></span
              ><strong>{{ currentMode.label }}文档</strong></router-link
            >
          </div>
          <label v-if="navigation.versions?.length" class="docs-version-picker"
            ><span>产品版本</span
            ><select
              :value="navigation.versionId"
              aria-label="产品文档版本"
              :disabled="loading"
              @change="switchVersion($event.target.value)"
            >
              <option
                v-for="version in navigation.versions"
                :key="version.id"
                :value="version.id"
              >
                {{ version.name
                }}{{
                  version.id === navigation.defaultVersion ? " · 默认" : ""
                }}
              </option>
            </select></label
          >
          <label class="docs-search"
            ><svg viewBox="0 0 24 24" aria-hidden="true">
              <circle cx="10.5" cy="10.5" r="6.5" />
              <path d="m16 16 4 4" /></svg
            ><input
              v-model="search"
              type="search"
              aria-label="搜索文档"
              placeholder="搜索当前模式文档"
          /></label>
          <router-link
            v-if="errorKind === 'navigation' && route.query.version"
            :to="modeLocation(currentMode)"
            class="default-docs-link"
            >返回默认版本
          </router-link>
        </div>
        <nav class="docs-navigation" aria-label="文档目录">
          <section v-for="group in filteredGroups" :key="group.title">
            <h2>
              {{ group.title }}<span>{{ group.items.length }}</span>
            </h2>
            <button
              v-for="item in group.items"
              :key="item.id"
              type="button"
              :class="{ active: String(item.id) === String(document?.id) }"
              :aria-current="
                String(item.id) === String(document?.id) ? 'page' : undefined
              "
              @click="selectDocument(item)"
            >
              {{ item.title }}
            </button>
          </section>
          <div
            v-if="!filteredGroups.length && !loading"
            class="docs-empty-search"
          >
            <p>{{ search ? "没有匹配的文档" : "暂无已发布文档" }}</p>
            <button v-if="search" type="button" @click="search = ''">
              清除搜索
            </button>
          </div>
        </nav>
      </aside>
      <section class="docs-main" aria-label="文档内容" :aria-busy="loading">
        <div
          v-if="loading"
          class="docs-loading"
          role="status"
          aria-label="正在加载文档"
        >
          <i /><i /><i /><i />
        </div>
        <ContentState
          v-else-if="error"
          class="docs-error"
          :error="errorKind !== 'empty'"
          :title="
            errorKind === 'empty'
              ? '文档正在准备中'
              : errorKind === 'missing'
                ? '没有找到这篇文档'
                : '暂时无法显示文档'
          "
          :message="error"
          action-label="重新加载"
          @retry="retry"
          ><router-link to="/" class="site-button site-button--outline"
            >返回首页</router-link
          ></ContentState
        >
        <template v-else-if="document">
          <header class="docs-toolbar">
            <nav class="docs-breadcrumb" aria-label="当前位置">
              <router-link :to="docLocation()">{{
                currentMode.label
              }}</router-link
              ><span aria-hidden="true">/</span
              ><span>{{ currentGroup || document.title }}</span>
            </nav>
            <button type="button" @click="copyMarkdown">复制 Markdown</button>
          </header>
          <p v-if="copyStatus" class="docs-copy-status" role="status">
            {{ copyStatus }}
          </p>
          <header class="docs-document-heading">
            <h1 :id="titleHeading?.slug">
              {{ titleHeading?.text || document.title }}
            </h1>
            <p v-if="document.summary">{{ document.summary }}</p>
          </header>
          <MarkdownContent class="article-content docs-article" :html="html" />
          <p v-if="!document.content.trim()" class="docs-empty-body">
            这篇文档暂时还没有正文。
          </p>
          <nav
            v-if="previousDocument || nextDocument"
            class="docs-pagination"
            aria-label="文档翻页"
          >
            <button
              v-if="previousDocument"
              type="button"
              @click="selectDocument(previousDocument)"
            >
              <small>上一篇</small
              ><strong> {{ previousDocument.title }}</strong></button
            ><span v-else />
            <button
              v-if="nextDocument"
              type="button"
              @click="selectDocument(nextDocument)"
            >
              <small>下一篇</small><strong>{{ nextDocument.title }} </strong>
            </button>
          </nav>
        </template>
      </section>
      <aside v-if="hasToc" class="docs-toc">
        <ReadingToc :items="toc" />
        <div class="docs-toc-help">
          <p>准备开始部署？</p>
          <router-link to="/download">前往下载中心 </router-link>
        </div>
      </aside>
    </div>
  </div>
</template>

<style scoped>
.docs-page {
  min-height: calc(100dvh - 64px);
  background: var(--deploy-white);
}
.docs-layout {
  display: grid;
  grid-template-columns: 272px minmax(0, 1fr) 224px;
  align-items: start;
  max-width: 1600px;
  margin: auto;
  gap: 0;
}
.docs-layout.without-toc {
  grid-template-columns: 272px minmax(0, 1fr);
}
.docs-sidebar {
  position: sticky;
  top: 64px;
  height: calc(100dvh - 64px);
  display: flex;
  flex-direction: column;
  min-width: 0;
  overflow: hidden;
  background: var(--deploy-paper);
  padding: 24px 24px 32px;
  border-right: 1px solid var(--deploy-line);
}
.docs-sidebar-controls {
  flex-shrink: 0;
  padding-bottom: 18px;
}
.docs-sidebar-head {
  margin-top: 20px;
  font-size: 14px;
}
.docs-sidebar-head > a {
  display: flex;
  align-items: center;
  gap: 8px;
}
.docs-brand-icon {
  display: none;
}
.docs-version-picker {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 16px;
  color: var(--deploy-muted);
  font-size: 11px;
}
.docs-version-picker > span {
  flex-shrink: 0;
}
.docs-version-picker select {
  min-width: 0;
  flex: 1;
  width: 100%;
  border: 1px solid var(--deploy-line);
  background: var(--deploy-white);
  color: var(--deploy-ink);
  padding: 6px 8px;
  border-radius: 5px;
  font-size: 12px;
}
.docs-search {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  border: 1px solid var(--deploy-line);
  border-radius: 5px;
  background: var(--deploy-white);
  margin-top: 14px;
}
.docs-search svg {
  width: 14px;
  height: 14px;
  fill: none;
  stroke: var(--deploy-muted);
  stroke-width: 1.5;
  flex-shrink: 0;
}
.docs-search input {
  width: 100%;
  min-width: 0;
  border: 0;
  outline: 0;
  background: transparent;
  color: var(--deploy-ink);
  font-size: 12px;
}
.docs-search:focus-within {
  border-color: var(--deploy-primary);
}
.docs-navigation {
  overflow: auto;
  min-height: 0;
  scrollbar-width: thin;
  overscroll-behavior: contain;
}
.docs-navigation section {
  padding: 18px 0;
  border-top: 1px solid var(--deploy-line);
}
.docs-navigation h2 {
  display: flex;
  justify-content: space-between;
  gap: 8px;
  font-size: 13px;
  font-weight: 600;
  margin: 0 0 8px;
}
.docs-navigation h2 span {
  font-size: 11px;
  font-weight: 400;
  color: var(--deploy-faint);
}
.docs-navigation button {
  display: block;
  width: 100%;
  padding: 7px 0 7px 12px;
  min-height: 32px;
  border: 0;
  background: transparent;
  text-align: left;
  color: var(--deploy-muted);
  font-size: 13px;
  line-height: 1.5;
  overflow-wrap: anywhere;
}
.docs-navigation button:hover,
.docs-navigation button.active {
  color: var(--deploy-primary);
}
.docs-navigation button.active {
  font-weight: 500;
}
.docs-empty-search {
  padding: 18px 0;
  font-size: 12px;
  color: var(--deploy-muted);
}
.docs-main {
  width: 100%;
  min-width: 0;
  max-width: 860px;
  margin: auto;
  padding: 36px 56px 80px;
}
.without-toc .docs-main {
  max-width: 950px;
}
.docs-toolbar {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: center;
  margin-bottom: 24px;
}
.docs-breadcrumb {
  display: flex;
  gap: 10px;
  font-size: 12px;
  color: var(--deploy-muted);
}
.docs-toolbar > button {
  border: 0;
  background: transparent;
  color: var(--deploy-muted);
  font-size: 11px;
  padding: 5px 0;
}
.docs-toolbar > button:hover {
  color: var(--deploy-primary);
}
.docs-copy-status {
  font-size: 12px;
  color: var(--deploy-primary);
}
.docs-document-heading h1 {
  margin: 0 0 20px;
  font-size: 32px;
  line-height: 1.4;
  letter-spacing: -0.02em;
  scroll-margin-top: 32px;
}
.docs-document-heading > p {
  margin: 0 0 28px;
  font-size: 16px;
  line-height: 1.8;
  color: var(--deploy-muted);
}
.docs-empty-body {
  color: var(--deploy-muted);
  padding: 40px 0;
  font-size: 14px;
}
.docs-pagination {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-top: 56px;
  padding-top: 24px;
  border-top: 1px solid var(--deploy-line);
}
.docs-pagination > button {
  display: grid;
  gap: 6px;
  min-width: 0;
  padding: 14px 16px;
  border: 1px solid var(--deploy-line);
  border-radius: 8px;
  text-align: left;
  background: var(--deploy-white);
}
.docs-pagination > button:last-child {
  text-align: right;
}
.docs-pagination small {
  color: var(--deploy-muted);
  font-size: 11px;
}
.docs-pagination strong {
  color: var(--deploy-primary);
  font-size: 13px;
  font-weight: 500;
  overflow-wrap: anywhere;
}
.docs-pagination > button:hover {
  border-color: var(--deploy-primary);
}
.docs-toc {
  position: sticky;
  top: 96px;
  max-height: calc(100dvh - 124px);
  min-height: 0;
  display: flex;
  flex-direction: column;
  padding: 0 28px 0 0;
  margin-top: 40px;
}
.docs-toc :deep(.reading-toc) {
  flex: 1;
}
.docs-toc-help {
  border-top: 1px solid var(--deploy-line);
  padding-top: 16px;
  margin-top: 28px;
  font-size: 11px;
  color: var(--deploy-muted);
  flex-shrink: 0;
}
.docs-toc-help a {
  color: var(--deploy-primary);
}
.docs-overview {
  max-width: 980px;
  display: flex;
  gap: 32px;
  justify-content: space-between;
  align-items: center;
  margin: 48px auto 24px;
  padding: 0 24px;
}
.docs-site-title {
  font-size: 24px;
  font-weight: 600;
  margin: 0;
}
.docs-description {
  color: var(--deploy-muted);
  font-size: 13px;
}
.docs-coming-soon {
  text-align: center;
  max-width: 720px;
  min-height: 440px;
  margin: 40px auto;
  padding: 60px 24px;
}
.docs-upcoming-icon {
  display: block;
  width: 48px;
  height: 48px;
  padding: 12px;
  border-radius: 8px;
  background: var(--deploy-soft);
  color: var(--deploy-primary);
  margin: 0 auto 20px;
}
.docs-upcoming-icon :deep(svg) {
  width: 24px;
  height: 24px;
  stroke: currentColor;
  stroke-width: 1.5;
}
.docs-upcoming-label {
  font-size: 12px;
  color: var(--deploy-faint);
}
.docs-coming-soon h1 {
  font-size: 28px;
}
.docs-coming-soon p {
  color: var(--deploy-muted);
  font-size: 14px;
  line-height: 1.8;
  margin-bottom: 28px;
}
.version-notice {
  padding: 10px 20px;
  font-size: 12px;
  color: var(--deploy-primary);
  background: var(--deploy-soft);
  margin: 0;
}
.docs-loading {
  display: grid;
  gap: 20px;
  padding: 32px 0;
  min-height: 450px;
}
.docs-loading i {
  height: 24px;
  background: var(--deploy-paper);
  border-radius: 5px;
  animation: site-pulse 1.5s infinite;
}
.docs-loading i:first-child {
  height: 42px;
  width: 65%;
}
.docs-error {
  padding: 48px 0;
}
.docs-mobile-bar,
.docs-backdrop {
  display: none;
}
@media (min-width: 1600px) {
  .docs-sidebar {
    border-left: 1px solid var(--deploy-line);
  }
}
@media (max-width: 1200px) {
  .docs-layout {
    grid-template-columns: 248px minmax(0, 1fr) 180px;
  }
  .docs-layout.without-toc {
    grid-template-columns: 248px minmax(0, 1fr);
  }
  .docs-main {
    padding: 32px;
  }
  .docs-toc {
    padding-right: 20px;
  }
  .docs-sidebar {
    padding-inline: 20px;
  }
}
@media (max-width: 1000px) {
  .docs-layout {
    grid-template-columns: 248px minmax(0, 1fr);
  }
  .docs-toc {
    display: none;
  }
}
@media (max-width: 760px) {
  .docs-layout,
  .docs-layout.without-toc {
    display: block;
  }
  .docs-main {
    padding: 28px 24px 56px;
    max-width: 740px;
  }
  .docs-document-heading h1 {
    font-size: 28px;
  }
  .docs-mobile-bar {
    display: flex;
    position: sticky;
    top: var(--site-header-offset);
    z-index: 45;
    justify-content: space-between;
    align-items: center;
    gap: 16px;
    background: var(--deploy-white);
    border-bottom: 1px solid var(--deploy-line);
    padding: 10px 24px;
    font-size: 12px;
    color: var(--deploy-muted);
  }
  .docs-mobile-bar > span {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .docs-mobile-bar button {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 4px 0;
    border: 0;
    background: transparent;
    color: var(--deploy-muted);
    white-space: nowrap;
  }
  .docs-mobile-bar svg {
    width: 15px;
    height: 15px;
    fill: none;
    stroke: currentColor;
    stroke-width: 1.5;
  }
  .docs-sidebar {
    position: fixed;
    top: 108px;
    bottom: 0;
    left: 0;
    width: min(300px, 85vw);
    height: calc(100dvh - 108px);
    z-index: 70;
    transform: translateX(-100%);
    visibility: hidden;
    transition:
      transform 0.2s,
      visibility 0.2s;
    padding: 24px;
  }
  .docs-sidebar.open {
    transform: none;
    visibility: visible;
  }
  .docs-backdrop {
    display: block;
    position: fixed;
    inset: 108px 0 0;
    background: #0006;
    border: 0;
    z-index: 60;
  }
  .docs-overview {
    margin-top: 28px;
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }
  .docs-toolbar {
    gap: 8px;
  }
  .docs-toolbar > button {
    font-size: 10px;
  }
  .docs-breadcrumb {
    font-size: 11px;
  }
  .docs-pagination {
    gap: 10px;
  }
  .docs-pagination > button {
    padding: 12px;
  }
  .docs-navigation section {
    padding-block: 16px;
  }
}
</style>
