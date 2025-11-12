<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue';
import http from '@/http';
import * as echarts from 'echarts/core';
import { LineChart } from 'echarts/charts';
import {
  TitleComponent,
  TooltipComponent,
  GridComponent,
  // 数据集组件
  DatasetComponent,
  // 内置数据转换器组件 (filter, sort)
  TransformComponent,
} from 'echarts/components';
import { LabelLayout, UniversalTransition } from 'echarts/features';
import { CanvasRenderer } from 'echarts/renderers';
import type {
  // 系列类型的定义后缀都为 SeriesOption
  LineSeriesOption,
} from 'echarts/charts';
import type {
  // 组件类型的定义后缀都为 ComponentOption
  TitleComponentOption,
  TooltipComponentOption,
  GridComponentOption,
  DatasetComponentOption,
} from 'echarts/components';
import type { ComposeOption } from 'echarts/core';
import { getMonthRange } from '../../utils';
import type { IChartCompareProps } from '../../typings';

// 通过 ComposeOption 来组合出一个只有必须组件和图表的 Option 类型
type ECOption = ComposeOption<
  LineSeriesOption | TitleComponentOption | TooltipComponentOption | GridComponentOption | DatasetComponentOption
>;

const props = defineProps<IChartCompareProps>();

// 注册必须的组件
echarts.use([
  TitleComponent,
  TooltipComponent,
  GridComponent,
  DatasetComponent,
  TransformComponent,
  LineChart,
  LabelLayout,
  UniversalTransition,
  CanvasRenderer,
]);

let chartInstance: echarts.ECharts | null = null;
const chartRef = ref<HTMLElement | null>(null);
const option: ECOption = {
  legend: {
    show: true,
  },
  grid: {
    left: 10,
    right: 10,
    top: 10,
  },
  dataset: {
    source: [],
  },
  xAxis: {
    type: 'category',
    axisLine: {
      show: false,
    },
  },
  yAxis: {},
  series: [
    {
      type: 'line',
      name: '当前结单率',
      color: '#699DF4',
      symbol: 'none',
    },
    {
      type: 'line',
      name: '对比结单率',
      color: '#699DF4',
      symbol: 'none',
    },
  ],
};

const fetchChartData = async () => {
  const _res = await Promise.all([
    http.post('/api/v1/woa/task/apply/completion-rate/statistics', {
      start_time: getMonthRange(props.option.daterange[0]).startTime,
      end_time: getMonthRange(props.option.daterange[1]).endTime,
    }),
    http.post('/api/v1/woa/task/apply/completion-rate/statistics', {
      start_time: getMonthRange(props.option.daterange[0]).startTime,
      end_time: getMonthRange(props.option.daterange[1]).endTime,
    }),
  ]);
  // console.log(res);
  // const chartData = res.data?.details || [];
  // chartInstance.setOption({
  //   dataset: {
  //     source: chartData.map((item: any) => [
  //       businessGlobalStore.businessFullList.find((business) => business.id === item.bk_biz_id)?.name,
  //       item.completion_rate,
  //     ]),
  //   },
  // });
};

watch(
  () => props.option,
  () => {
    fetchChartData();
  },
  { deep: true, immediate: true },
);

const handleChartResize = () => {
  chartInstance?.resize();
};

onMounted(() => {
  chartInstance = echarts.init(chartRef.value);
  chartInstance?.setOption(option);
  window.addEventListener('resize', handleChartResize);
  handleChartResize();
});

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleChartResize);
  chartInstance?.dispose();
  chartInstance = null;
});
</script>

<template>
  <div class="chart-apply-completion-rate" ref="chartRef"></div>
</template>

<style lang="scss" scoped>
.chart-apply-completion-rate {
  width: 100%;
  height: 100%;
}
</style>
