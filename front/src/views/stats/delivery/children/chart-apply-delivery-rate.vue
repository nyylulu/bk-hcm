<script setup lang="ts">
import { onMounted, ref, watch, onBeforeUnmount } from 'vue';
import http from '@/http';
import * as echarts from 'echarts/core';
import { LineChart } from 'echarts/charts';
import {
  LegendComponent,
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
  LegendComponentOption,
  TitleComponentOption,
  TooltipComponentOption,
  GridComponentOption,
  DatasetComponentOption,
} from 'echarts/components';
import type { ComposeOption } from 'echarts/core';
import { getMonthRange } from '../../utils';

// 通过 ComposeOption 来组合出一个只有必须组件和图表的 Option 类型
type ECOption = ComposeOption<
  | LineSeriesOption
  | TitleComponentOption
  | TooltipComponentOption
  | GridComponentOption
  | DatasetComponentOption
  | LegendComponentOption
>;

const props = defineProps<{
  daterange: Date[];
}>();

// 注册必须的组件
echarts.use([
  TitleComponent,
  LegendComponent,
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
  tooltip: {
    trigger: 'axis',
  },
  legend: {
    show: true,
    data: [
      {
        name: '平均耗时',
        textStyle: {
          color: '#4D4F56',
        },
      },
    ],
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
      name: '平均耗时',
      color: '#699DF4',
      symbol: 'none',
    },
  ],
};

const fetchChartData = async () => {
  const res = await http.post('/api/v1/woa/task/apply/delivery-rate/statistics', {
    start_time: getMonthRange(props.daterange[0]).startTime,
    end_time: getMonthRange(props.daterange[1]).endTime,
  });
  const chartData = res.data?.details || [];
  chartInstance.setOption({
    dataset: {
      source: chartData.map((item: any) => [item.year_month, item.delivery_rate]),
    },
  });
};

watch(
  () => props.daterange,
  () => {
    fetchChartData();
  },
  { immediate: true },
);

const handleChartResize = () => {
  chartInstance?.resize();
};

onMounted(() => {
  chartInstance = echarts.init(chartRef.value);
  chartInstance.setOption(option);
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
  <div class="chart-apply-delivery-rate" ref="chartRef"></div>
</template>

<style lang="scss" scoped>
.chart-apply-delivery-rate {
  width: 100%;
  height: 100%;
}
</style>
