import { ref } from "vue";
import * as echarts from "echarts/core";
import { LineChart, BarChart, PieChart } from "echarts/charts";
import {
  GridComponent,
  TooltipComponent,
  LegendComponent,
} from "echarts/components";
import { CanvasRenderer } from "echarts/renderers";
echarts.use([
  LineChart,
  BarChart,
  PieChart,
  GridComponent,
  TooltipComponent,
  LegendComponent,
  CanvasRenderer,
]);

export function useStatisticsCharts(days, categoryData, tagData) {
  let resizeFrame;
  let publishChart = null;
  let viewChart = null;
  let categoryChart = null;
  let tagChart = null;

  const publishChartRef = ref(null);
  const viewChartRef = ref(null);
  const categoryChartRef = ref(null);
  const tagChartRef = ref(null);

  const initCharts = (articleTrendRes, viewTrendRes) => {
    const trendDays = days.value || 7;

    // ---- 发布趋势图 ----
    if (publishChartRef.value) {
      publishChart ||= echarts.init(publishChartRef.value);
      const at = articleTrendRes?.data || {};
      const dates = at.dates || generateDates(trendDays);
      const counts = at.values || at.counts || [];

      publishChart.setOption({
        tooltip: { trigger: "axis" },
        grid: { left: "3%", right: "4%", bottom: "3%", containLabel: true },
        xAxis: { type: "category", boundaryGap: false, data: dates },
        yAxis: { type: "value" },
        series: [
          {
            name: "发布数",
            type: "line",
            smooth: true,
            data: counts.length > 0 ? counts : Array(trendDays).fill(0),
            itemStyle: { color: "#67c23a" },
            areaStyle: {
              color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                { offset: 0, color: "rgba(103,194,58,0.25)" },
                { offset: 1, color: "rgba(103,194,58,0.02)" },
              ]),
            },
          },
        ],
      });
    }

    // ---- 浏览量趋势图 ----
    if (viewChartRef.value) {
      viewChart ||= echarts.init(viewChartRef.value);
      const vt = viewTrendRes?.data || {};
      const dates = vt.dates || generateDates(trendDays);
      const views = vt.values || vt.views || [];

      viewChart.setOption({
        tooltip: { trigger: "axis" },
        grid: { left: "3%", right: "4%", bottom: "3%", containLabel: true },
        xAxis: { type: "category", boundaryGap: false, data: dates },
        yAxis: { type: "value" },
        series: [
          {
            name: "浏览量",
            type: "bar",
            barWidth: "40%",
            data: views.length > 0 ? views : Array(trendDays).fill(0),
            itemStyle: {
              color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                { offset: 0, color: "#409eff" },
                { offset: 1, color: "#79bbff" },
              ]),
              borderRadius: [4, 4, 0, 0],
            },
          },
        ],
      });
    }

    // ---- 分类饼图 ----
    if (categoryChartRef.value) {
      categoryChart ||= echarts.init(categoryChartRef.value);
      const pieData = categoryData.value
        .filter((c) => c.articleCount > 0)
        .map((c) => ({ name: c.name, value: c.articleCount }));

      categoryChart.setOption({
        tooltip: { trigger: "item", formatter: "{b}: {c}篇 ({d}%)" },
        legend: { orient: "vertical", left: "left", top: "middle" },
        series: [
          {
            name: "分类文章数",
            type: "pie",
            radius: ["35%", "65%"],
            center: ["58%", "52%"],
            avoidLabelOverlap: true,
            itemStyle: { borderRadius: 8, borderColor: "#fff", borderWidth: 2 },
            label: { show: true, formatter: "{b}\n{c}", fontSize: 11 },
            emphasis: {
              label: { show: true, fontWeight: "bold", fontSize: 14 },
            },
            data:
              pieData.length > 0
                ? pieData
                : [
                    {
                      value: 1,
                      name: "暂无数据",
                      itemStyle: { color: "#ddd" },
                    },
                  ],
          },
        ],
      });
    }

    // ---- 标签柱状图 ----
    if (tagChartRef.value) {
      tagChart ||= echarts.init(tagChartRef.value);
      const top10 = tagData.value
        .slice(0, 10)
        .sort((a, b) => a.articleCount - b.articleCount);

      tagChart.setOption({
        tooltip: { trigger: "axis", axisPointer: { type: "shadow" } },
        grid: { left: "3%", right: "8%", bottom: "3%", containLabel: true },
        xAxis: { type: "value" },
        yAxis: {
          type: "category",
          data: top10.map((t) => t.name),
          axisLabel: { fontSize: 12 },
        },
        series: [
          {
            name: "使用次数",
            type: "bar",
            data: top10.map((t) => t.articleCount || 0),
            barWidth: "60%",
            itemStyle: {
              color: new echarts.graphic.LinearGradient(0, 0, 1, 0, [
                { offset: 0, color: "#409eff" },
                { offset: 1, color: "#79bbff" },
              ]),
              borderRadius: [0, 4, 4, 0],
            },
            label: {
              show: true,
              position: "right",
              fontSize: 11,
              color: "#666",
            },
          },
        ],
      });
    }
  };

  const generateDates = (d) => {
    const arr = [];
    for (let i = d - 1; i >= 0; i--) {
      const dt = new Date();
      dt.setDate(dt.getDate() - i);
      arr.push(
        `${String(dt.getMonth() + 1).padStart(2, "0")}-${String(dt.getDate()).padStart(2, "0")}`,
      );
    }
    return arr;
  };

  const handleResize = () => {
    if (resizeFrame) return;
    resizeFrame = requestAnimationFrame(() => {
      resizeFrame = undefined;
      if (!publishChartRef.value?.getClientRects().length) return;
      publishChart?.resize();
      viewChart?.resize();
      categoryChart?.resize();
      tagChart?.resize();
    });
  };
  const disposeCharts = () => {
    cancelAnimationFrame(resizeFrame);
    publishChart?.dispose();
    viewChart?.dispose();
    categoryChart?.dispose();
    tagChart?.dispose();
    publishChart = null;
    viewChart = null;
    categoryChart = null;
    tagChart = null;
  };

  return {
    publishChartRef,
    viewChartRef,
    categoryChartRef,
    tagChartRef,
    initCharts,
    handleResize,
    disposeCharts,
  };
}
