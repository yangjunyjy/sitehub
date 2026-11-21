<script setup>
import { computed } from "vue";
import { externalAddress } from "@shared/navigation/external-address.mjs";
import ContentState from "@/components/product/ContentState.vue";
import { useMember } from "@/composables/useMember";
import { useWebsiteContent } from "@/composables/useWebsiteContent";
import HomeHero from "./components/HomeHero.vue";
import ProductShowcase from "./components/ProductShowcase.vue";
import ProductCapabilities from "./components/ProductCapabilities.vue";
import HomeDelivery from "./components/HomeDelivery.vue";

const { website, loadWebsiteContent } = useWebsiteContent();
const { member } = useMember();
const home = computed(() => website.content.home);
const blogURL = externalAddress(window.__SITEHUB_CONFIG__?.blogURL);
</script>

<template>
  <ContentState
    v-if="website.error && !website.loaded"
    class="site-shell home-state"
    error
    title="产品内容加载失败"
    :message="website.error"
    action-label="重新加载"
    :loading="website.loading"
    @retry="loadWebsiteContent(true).catch(() => {})"
  />
  <div v-else-if="!website.loaded" class="site-shell home-state" role="status">
    正在加载产品信息…
  </div>
  <div v-else class="home-page">
    <HomeHero :home="home" :brand="website.content.brand.name" />
    <section class="home-workbench" aria-labelledby="workbench-heading">
      <div class="site-shell home-workbench__layout">
        <div class="home-workbench__preview">
          <h2 id="workbench-heading">产品工作台</h2>
          <ProductShowcase :brand="website.content.brand.name" />
        </div>
        <div class="home-workbench__details">
          <details
            v-if="
              website.content.capabilities.length ||
              home.capabilityIntro ||
              home.eyebrow ||
              home.meta?.length
            "
            id="capabilities"
          >
            <summary>{{ home.capabilityTitle || "产品功能" }}</summary>
            <p v-if="home.eyebrow">{{ home.eyebrow }}</p>
            <p v-if="home.capabilityIntro">{{ home.capabilityIntro }}</p>
            <p v-if="home.meta?.length" class="home-specs">
              {{ home.meta.join(" · ") }}
            </p>
            <ProductCapabilities :items="website.content.capabilities" />
          </details>
          <details
            v-if="website.content.deliverySteps.length || home.flowIntro"
            id="delivery"
          >
            <summary>{{ home.flowTitle || "交付流程" }}</summary>
            <p v-if="home.flowIntro">{{ home.flowIntro }}</p>
            <HomeDelivery :steps="website.content.deliverySteps" />
          </details>
          <details
            v-if="website.content.audiences.length || home.audienceIntro"
            id="teams"
          >
            <summary>{{ home.audienceTitle || "适用团队" }}</summary>
            <p v-if="home.audienceIntro">{{ home.audienceIntro }}</p>
            <dl class="home-audiences">
              <template
                v-for="item in website.content.audiences"
                :key="item.title"
              >
                <dt>{{ item.title }}</dt>
                <dd v-if="item.headline">{{ item.headline }}</dd>
                <dd v-if="item.description">{{ item.description }}</dd>
              </template>
            </dl>
          </details>
          <a
            v-if="home.blogTitle && blogURL"
            class="home-configured-link"
            :href="blogURL"
            >{{ home.blogTitle }} <span aria-hidden="true">↗</span></a
          >
          <router-link
            v-if="home.cta"
            class="home-configured-link"
            :to="{ name: member.user ? 'account' : 'register' }"
            >{{ home.cta }} <span aria-hidden="true">→</span></router-link
          >
        </div>
      </div>
    </section>
  </div>
</template>

<style lang="scss" scoped src="./home.scss"></style>
