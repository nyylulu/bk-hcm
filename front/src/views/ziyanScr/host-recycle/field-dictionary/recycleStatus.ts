import { h } from 'vue';
// import { getRecycleStatusOpts } from '@/api/host/recycle';
// import { Spinner } from 'bkui-vue/lib/icon';
const recycleTaskStatus: any[] = ['DONE'];

// getRecycleStatusOpts().then((res: { data: { info: any[] } }) => {
//   recycleTaskStatus = res.data.info.map((item: { description: any; status: any }) => ({
//     label: item.description,
//     value: item.status,
//   }));
// });

export const getRecycleTaskStatusLabel = (value: string) =>
  recycleTaskStatus.find((status) => status.value === value)?.label || value || '-';

export const getRecycleTaskStatusView = (value: string) => {
  const label = getRecycleTaskStatusLabel(value);
  let statusNodes = h('span', label);
  if (value === 'DONE') {
    statusNodes = h(
      'span',
      {
        class: 'c-success',
      },
      label,
    );
  }
  if (value.includes('ING')) {
    statusNodes = h('span', [h('Spinner'), h('span', label)]);
  }
  if (value === 'DETECT_FAILED') {
    statusNodes = h(
      'span',
      {
        class: 'c-danger',
      },
      [
        h(
          'bk-badge',
          {
            dot: true,
            'v-bk-tooltips': {
              content: '请到“预检详情”查看失败原因，或者点击“去除预检失败IP提交”',
            },
          },
          label,
        ),
      ],
    );
  }
  if (value.includes('FAILED')) {
    statusNodes = h(
      'span',
      {
        class: 'c-danger',
      },
      label,
    );
  }
  return statusNodes;
};
