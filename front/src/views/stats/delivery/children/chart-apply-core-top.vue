<script setup lang="ts">
import { onMounted, ref, watch } from 'vue';
import http from '@/http';
import * as echarts from 'echarts/core';
import { BarChart } from 'echarts/charts';
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
  BarSeriesOption,
} from 'echarts/charts';
import type {
  // 组件类型的定义后缀都为 ComponentOption
  TitleComponentOption,
  TooltipComponentOption,
  GridComponentOption,
  DatasetComponentOption,
} from 'echarts/components';
import type { ComposeOption } from 'echarts/core';
import { useBusinessGlobalStore } from '@/store/business-global';
import { getMonthRange } from '../../utils';

// 通过 ComposeOption 来组合出一个只有必须组件和图表的 Option 类型
type ECOption = ComposeOption<
  BarSeriesOption | TitleComponentOption | TooltipComponentOption | GridComponentOption | DatasetComponentOption
>;

const props = defineProps<{
  daterange: Date[];
}>();

// 注册必须的组件
echarts.use([
  TitleComponent,
  TooltipComponent,
  GridComponent,
  DatasetComponent,
  TransformComponent,
  BarChart,
  LabelLayout,
  UniversalTransition,
  CanvasRenderer,
]);

let chartInstance: echarts.ECharts | null = null;
const chartRef = ref<HTMLElement | null>(null);
const option: ECOption = {
  grid: {
    left: 10,
    right: 10,
    bottom: 10,
    top: 10,
  },
  dataset: {
    source: [],
  },
  xAxis: {
    type: 'value',
  },
  yAxis: {
    type: 'category',
    axisLine: {
      show: false,
    },
  },
  series: [
    {
      type: 'bar',
      label: { show: true, position: 'insideRight', color: '#fff' },
      color: '#F27051',
    },
  ],
};

const businessGlobalStore = useBusinessGlobalStore();

const fetchChartData = async () => {
  const res = await http.post('/api/v1/woa/task/apply/findmany/bizs/cpucores/statistics', {
    start_time: getMonthRange(props.daterange[0]).startTime,
    end_time: getMonthRange(props.daterange[1]).endTime,
  });
  const chartData = res.data?.details || [];
  chartInstance.setOption({
    dataset: {
      source: chartData.map((item: any) => [
        businessGlobalStore.businessFullList.find((business) => business.id === item.bk_biz_id)?.name,
        item.delivered_core_count,
      ]),
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
  if (!chartInstance) {
    return;
  }
  chartInstance.resize();
};

onMounted(() => {
  chartInstance = echarts.init(chartRef.value);
  chartInstance?.setOption(option);
  window.addEventListener('resize', handleChartResize);
  handleChartResize();
});
</script>

<template>
  <div class="chart-apply-core-top" ref="chartRef"></div>
</template>

<style lang="scss" scoped>
.chart-apply-core-top {
  width: 100%;
  height: 100%;
}
</style>
