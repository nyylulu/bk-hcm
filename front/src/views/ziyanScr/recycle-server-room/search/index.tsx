import { defineComponent, ref, onBeforeMount, type PropType, computed } from 'vue';

import { useI18n } from 'vue-i18n';
import { useZiyanScrStore } from '@/store/ziyanScr';

import Panel from '@/components/panel';

import cssModule from './index.module.scss';

import type { IRecycleArea } from '@/typings/ziyanScr';

type RecycleArea = {
  start_time: string;
  end_time: string;
  which_stages: number;
  children: IRecycleArea[];
};

export default defineComponent({
  props: {
    moduleNames: Array as PropType<string[]>,
  },

  emits: {
    'update:moduleNames'(moduleNames: string[]) {
      return moduleNames;
    },
  },

  setup(props, { emit }) {
    const { t } = useI18n();
    const ziyanScrStore = useZiyanScrStore();

    const recycleAreaGroups = ref<RecycleArea[]>([]);
    const loading = ref(false);

    const isGroupChecked = (areaGroup: RecycleArea) => {
      return areaGroup.children.every((group) => props.moduleNames.includes(group.name));
    };
    const isGroupHalfChecked = (areaGroup: RecycleArea) => {
      return areaGroup.children.some((group) => props.moduleNames.includes(group.name));
    };

    const isAllChecked = computed(() => {
      const total = recycleAreaGroups.value.reduce((acc, cur) => acc + cur.children.length, 0);
      return props.moduleNames.length === total;
    });
    const isAllHalfChecked = computed(() => props.moduleNames.length > 0);

    // 选择
    const handleCheck = (moduleName: string) => {
      const moduleIndex = props.moduleNames.findIndex((item) => item === moduleName);
      const moduleNames = [...props.moduleNames];
      if (moduleIndex > -1) {
        moduleNames.splice(moduleIndex, 1);
      } else {
        moduleNames.push(moduleName);
      }

      emit('update:moduleNames', moduleNames);
    };

    // 选择所有
    const handleCheckAll = (isChecked: boolean) => {
      const moduleNames = isChecked
        ? recycleAreaGroups.value.reduce((acc, cur) => {
            acc.push(...cur.children.map((item) => item.name));
            return acc;
          }, [])
        : [];
      emit('update:moduleNames', moduleNames);
    };

    // 选择区间所有
    const handleCheckGroupAll = (recycleAreaGroup: RecycleArea, isChecked: boolean) => {
      const moduleNames = [...props.moduleNames];
      recycleAreaGroup.children.forEach((item) => {
        const index = moduleNames.findIndex((moduleName) => moduleName === item.name);
        if (index > -1 && !isChecked) {
          moduleNames.splice(index, 1);
        }
        if (index < 0 && isChecked) {
          moduleNames.push(item.name);
        }
      });
      emit('update:moduleNames', moduleNames);
    };

    const handleInitRecycleArea = async () => {
      try {
        loading.value = true;
        const data = await ziyanScrStore.getRecycleAreas({
          page: {
            count: false,
            start: 0,
            limit: 500,
          },
        });

        recycleAreaGroups.value =
          data?.data?.details?.reduce((acc, cur) => {
            let currentRecycleAreaGroup = acc.find(
              (recycleAreaGroup) => recycleAreaGroup.which_stages === cur.which_stages,
            );

            if (!currentRecycleAreaGroup) {
              currentRecycleAreaGroup = {
                start_time: cur.start_time,
                end_time: cur.end_time,
                which_stages: cur.which_stages,
                children: [],
              };

              acc.push(currentRecycleAreaGroup);
            }

            currentRecycleAreaGroup.children.push(cur);

            return acc;
          }, [] as RecycleArea[]) || [];
      } catch (error) {
        recycleAreaGroups.value = [];
      } finally {
        loading.value = false;
      }
    };

    onBeforeMount(handleInitRecycleArea);

    return () => (
      <bk-loading loading={loading.value}>
        <Panel class={cssModule.gap}>
          <bk-checkbox
            onChange={handleCheckAll}
            checked={isAllChecked.value}
            indeterminate={!isAllChecked.value && isAllHalfChecked.value}
            immediateEmitChange={false}>
            {t('全选所有区间')}
          </bk-checkbox>
        </Panel>

        {recycleAreaGroups.value.map((recycleAreaGroup) => (
          <Panel class={cssModule.gap}>
            <section class={cssModule.title}>
              {t('日期区间')}：{recycleAreaGroup.start_time} {t('至')} {recycleAreaGroup.end_time}
              <bk-checkbox
                class={cssModule.chooseAll}
                checked={isGroupChecked(recycleAreaGroup)}
                indeterminate={!isGroupChecked(recycleAreaGroup) && isGroupHalfChecked(recycleAreaGroup)}
                immediateEmitChange={false}
                onChange={(isChecked: boolean) => handleCheckGroupAll(recycleAreaGroup, isChecked)}>
                {t('全选')}
              </bk-checkbox>
            </section>
            {recycleAreaGroup.children.map((recycleArea) => (
              <bk-checkbox
                class={cssModule.choose}
                modelValue={props.moduleNames.includes(recycleArea.name)}
                onChange={() => handleCheck(recycleArea.name)}>
                {recycleArea.name}
              </bk-checkbox>
            ))}
          </Panel>
        ))}
      </bk-loading>
    );
  },
});
