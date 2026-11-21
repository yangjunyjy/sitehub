<script setup>
import ResourceIcon from '@shared/icons/ResourceIcon.vue'
import DocumentModeSwitch from "./components/DocumentModeSwitch.vue";
import ReadingToc from "./components/ReadingToc.vue";
import ContentState from "@/components/product/ContentState.vue";
import MarkdownContent from "./components/MarkdownContent.vue";
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
  loadNavigation,
  retry,
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
      v-if="modesReady && currentMode.id && !currentMode.available"
      class="docs-coming-soon"
      aria-labelledby="upcoming-mode-title"
    >
      <span class="docs-upcoming-icon"
        ><ResourceIcon :size="36" :name="currentMode.icon"
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
    <ContentState v-else-if="modesReady && !documentModes.length" title="暂无公开文档" message="发布文档后会在这里显示。" />
    <ContentState
      v-else-if="!modesReady"
      :title="modesError ? '文档模式加载失败' : '正在加载文档模式'"
      :message="modesError"
      :error="!!modesError"
      :action-label="modesError ? '重新加载' : ''"
      @retry="loadNavigation"
    ><router-link v-if="modesError && (route.query.mode || route.query.version)" class="site-button site-button--outline"
        :to="{ name: 'docs', params: { id: '' } }">返回默认文档</router-link>
    </ContentState>
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
                ><ResourceIcon :name="currentMode.icon || 'playbook'" :size="20" /></span
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
          <details v-for="group in filteredGroups" :key="group.title" open>
            <summary>{{ group.title }}</summary>
            <router-link v-for="item in group.items" :key="item.id" :to="docLocation(item.slug || item.id)"
              :class="{ active: String(item.id) === String(document?.id) }"
              :aria-current="String(item.id) === String(document?.id) ? 'page' : undefined"
              @click="closeMenu()">{{ item.title }}</router-link>
          </details>
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
          ><router-link :to="{ name: 'home' }" class="site-button site-button--outline"
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
            <router-link
              v-if="previousDocument"
              :to="docLocation(previousDocument.slug || previousDocument.id)"
            >
              <small>上一篇</small
              ><strong> {{ previousDocument.title }}</strong></router-link
            ><span v-else />
            <router-link
              v-if="nextDocument"
              :to="docLocation(nextDocument.slug || nextDocument.id)"
            >
              <small>下一篇</small><strong>{{ nextDocument.title }} </strong>
            </router-link>
          </nav>
        </template>
      </section>
      <aside v-if="hasToc" class="docs-toc">
        <ReadingToc :items="toc" />
      </aside>
    </div>
  </div>
</template>

<style lang="scss" scoped src="./docs.scss"></style>
