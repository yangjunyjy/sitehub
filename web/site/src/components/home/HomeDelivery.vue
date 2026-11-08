<script setup>
defineProps({
  home: { type: Object, required: true },
  deliverySteps: { type: Array, default: () => [] },
});
</script>
<template>
  <section id="delivery" class="site-section flow-section">
    <div class="site-shell">
      <header class="flow-heading">
        <div>
          <span class="site-index">02</span>
          <p class="site-kicker">Delivery flow</p>
          <h2 class="site-heading">{{ home.flowTitle }}</h2>
        </div>
        <p>{{ home.flowIntro }}</p>
      </header>
      <div class="delivery-flow">
        <article v-for="step in deliverySteps" :key="step.key">
          <div>
            <span>{{ step.index }}</span
            ><em>{{ step.key }}</em>
          </div>
          <div class="flow-line"><i></i><b></b></div>
          <h3>{{ step.title }}</h3>
          <p>{{ step.description }}</p>
        </article>
      </div>
      <div class="flow-proof">
        <div>
          <p class="site-kicker">Policy in the loop</p>
          <h3>效率和安全，不必二选一</h3>
          <p>
            审批、权限与审计成为交付流程的默认规则，在不打断协作的前提下，为生产环境建立清晰边界。
          </p>
        </div>
        <div class="approval">
          <header>
            <span><i></i> 生产发布审批</span><small>流程演示</small>
          </header>
          <div class="approval__body">
            <div class="approval__app">
              <span>YL</span>
              <p>
                <strong>api-service v2.8.4</strong
                ><small>申请发布至 production</small>
              </p>
              <em>待审批</em>
            </div>
            <dl>
              <div>
                <dt>目标集群</dt>
                <dd>prod-k8s-01</dd>
              </div>
              <div>
                <dt>发布策略</dt>
                <dd>滚动更新</dd>
              </div>
              <div>
                <dt>变更内容</dt>
                <dd>3 commits</dd>
              </div>
            </dl>
            <router-link class="approval-action" to="/docs"
              >了解审批流程
            </router-link>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>
<style scoped>
.flow-section {
  background: var(--deploy-white);
  border-bottom: 1px solid var(--deploy-line);
}
.flow-heading {
  display: grid;
  grid-template-columns: 1.2fr 1fr;
  gap: 60px;
  align-items: end;
}
.flow-heading > p {
  margin: 0;
  color: var(--deploy-muted);
  font-size: 14px;
  line-height: 1.9;
}
.delivery-flow {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 28px;
  margin: 40px 0;
}
.delivery-flow article {
  padding: 0;
  min-width: 0;
}
.delivery-flow article > div:first-child {
  display: flex;
  justify-content: space-between;
  font: 11px var(--site-mono);
  color: var(--deploy-faint);
}
.delivery-flow em {
  font-style: normal;
}
.flow-line {
  display: flex;
  gap: 6px;
  align-items: center;
  height: 36px;
}
.flow-line i {
  width: 8px;
  height: 8px;
  background: var(--deploy-primary);
  border-radius: 50%;
}
.flow-line b {
  flex: 1;
  height: 1px;
  background: var(--deploy-line);
}
.delivery-flow h3 {
  font-size: 16px;
  margin: 0 0 12px;
}
.delivery-flow p {
  font-size: 13px;
  line-height: 1.9;
  color: var(--deploy-muted);
  margin: 0;
}
.flow-proof {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 64px;
  align-items: center;
  padding-top: 32px;
}
.flow-proof h3 {
  font-size: 24px;
  margin: 16px 0;
}
.flow-proof > div > p:last-child {
  color: var(--deploy-muted);
  font-size: 14px;
  line-height: 1.9;
}
.approval {
  background: var(--deploy-paper);
  border: 1px solid var(--deploy-line);
  border-radius: 8px;
  font-size: 12px;
}
.approval > header {
  display: flex;
  justify-content: space-between;
  padding: 12px 18px;
  border-bottom: 1px solid var(--deploy-line);
  color: var(--deploy-muted);
}
.approval small {
  color: var(--deploy-faint);
  font-size: 10px;
}
.approval__body {
  padding: 18px;
}
.approval__app {
  display: flex;
  gap: 12px;
  align-items: center;
}
.approval__app > span {
  background: var(--deploy-soft);
  color: var(--deploy-primary);
  padding: 10px;
  border-radius: 5px;
}
.approval__app p {
  display: grid;
  gap: 4px;
  margin: 0;
}
.approval__app em {
  margin-left: auto;
  font-style: normal;
  background: var(--deploy-soft);
  color: var(--deploy-primary);
  padding: 3px 8px;
  border-radius: 4px;
  font-size: 11px;
}
.approval dl {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  padding-block: 10px;
}
.approval dt {
  color: var(--deploy-faint);
  font-size: 10px;
}
.approval dd {
  margin: 5px 0 0;
  font-size: 11px;
}
.approval-action {
  display: block;
  text-align: right;
  color: var(--deploy-primary);
  border-top: 1px solid var(--deploy-line);
  padding-top: 12px;
}
@media (max-width: 760px) {
  .flow-heading,
  .flow-proof {
    grid-template-columns: 1fr;
    gap: 20px;
  }
  .delivery-flow {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 24px;
  }
  .flow-proof {
    padding-top: 16px;
  }
}
</style>
