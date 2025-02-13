import { defineComponent, type PropType, ref, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useZiyanScrStore } from '@/store/ziyanScr';
import { useBusinessGlobalStore } from '@/store/business-global';
import { useUserStore } from '@/store';
import { getDisplayText } from '@/utils';
import ExportToExcelButton from '@/components/export-to-excel-button';
import Panel from '@/components/panel';
import HcmSearchOrg from '@/components/search/org.vue';
import HcmSearchBusiness from '@/components/search/business.vue';
import HcmSearchUser from '@/components/search/user.vue';
import CurrentDialog from '../current-dialog';
import OriginDialog from '../origin-dialog';
import cssModule from './index.module.scss';

import type { IDissolve } from '@/typings/ziyanScr';
import type { Department } from '@/typings';

interface IDepartmentWithExtras extends Department {
  extras?: {
    code: string;
  };
}

export default defineComponent({
  components: {
    ExportToExcelButton,
  },
  props: {
    moduleNames: Array as PropType<string[]>,
  },
  setup(props) {
    const { t } = useI18n();
    const ziyanScrStore = useZiyanScrStore();
    const businessGlobalStore = useBusinessGlobalStore();
    const userStore = useUserStore();

    const tableColumns = [
      {
        label: '业务',
        field: 'bk_biz_name',
      },
      {
        label: '裁撤进度',
        field: 'progress',
      },
      {
        label: '原始数量',
        field: 'total.origin.host_count',
      },
      {
        label: '原始CPU',
        field: 'total.origin.cpu_count',
      },
      {
        label: '当前数量',
        field: 'total.current.host_count',
      },
      {
        label: '当前CPU',
        field: 'total.current.cpu_count',
      },
    ];
    const content = (
      <>
        {t('请勾选相关参数后查询，')}
        <br />
        {t('1.裁撤的机房模块，至少选一个模块，')}
      </>
    );
    const isLoading = ref(false);
    const organizations = ref([]);
    const orgChecked = ref([]);
    const operators = ref([]);
    const bkBizIds = ref([]);
    const dissloveList = ref<IDissolve[]>([]);
    const currentDialogShow = ref(false);
    const originDialogShow = ref(false);
    const searchParams = ref();
    const moduleNames = ref<string[]>([]);
    const exportColumns = ref(tableColumns);

    const currentRowData = ref<IDissolve>();

    const orgRef = ref(null);

    const isMeetSearchConditions = computed(() => props.moduleNames.length);

    const groupIds = computed(() => {
      const leafIds = new Set<string>();

      const collectLeafCodes = (dept: IDepartmentWithExtras) => {
        if (!dept?.has_children && dept?.extras?.code) {
          leafIds.add(dept.extras.code);
        }
        if (dept?.has_children && dept?.children) {
          dept.children.forEach((child) => collectLeafCodes(child));
        }
      };

      // 收集所选节点下的所有叶子节点的code数据
      orgChecked.value.forEach((org) => {
        collectLeafCodes(org);
      });

      return [...leafIds];
    });

    const handleSearch = async () => {
      isLoading.value = true;

      exportColumns.value = tableColumns.concat(
        props.moduleNames.map((moduleName) => ({ label: moduleName, field: `module_host_count.${moduleName}` })),
      );

      moduleNames.value = [...props.moduleNames];
      ziyanScrStore
        .getDissolveList({
          group_ids: groupIds.value,
          bk_biz_names: !bkBizIds.value?.[0] ? [] : businessGlobalStore.getBusinessNames(bkBizIds.value),
          module_names: moduleNames.value,
          operators: operators.value,
        })
        .then((result) => {
          const list = result?.data?.items || [];
          const fixedBizIds = ['total'];

          // “裁撤进度”合并到“总数中”
          const totalIndex = list.findIndex((item) => item.bk_biz_id === 'total');
          const processIndex = list.findIndex((item) => item.bk_biz_id === 'recycle-progress');
          if (totalIndex > -1) {
            list[totalIndex].progress = list[processIndex]?.progress;
          }
          if (processIndex > -1) {
            list.splice(processIndex, 1);
          }

          dissloveList.value = list.sort((a, b) => {
            const countA = (a?.total?.current?.host_count || 0) as number;
            const countB = (b?.total?.current?.host_count || 0) as number;
            // 置顶
            if (fixedBizIds.includes(a.bk_biz_id as string) || fixedBizIds.includes(b.bk_biz_id as string)) {
              return -1;
            }
            return countB - countA;
          });
        })
        .finally(() => {
          isLoading.value = false;
        });
    };

    const handleReset = () => {
      orgRef.value.clear();
      bkBizIds.value = [];
      operators.value = [userStore.username];
    };

    const setSearchParams = (bkBizNames: string[], moduleNames: string[]) => {
      searchParams.value = {
        group_ids: groupIds.value,
        bk_biz_names: bkBizNames,
        module_names: moduleNames,
        operators: operators.value,
      };
    };

    const firstLevelDataFilter = (data: IDepartmentWithExtras[]) => {
      // 正式环境第一层级只保留指定id的节点
      const found = data.filter((item) => item.id === 2874);
      if (found.length) {
        return found;
      }
      return data;
    };

    const handleShowOriginDialog = (bkBizNames: string[], row: IDissolve) => {
      originDialogShow.value = true;
      setSearchParams(bkBizNames, moduleNames.value);
      currentRowData.value = row;
    };

    const handleShowCurrentDialog = (bkBizNames: string[], row: IDissolve) => {
      currentDialogShow.value = true;
      setSearchParams(bkBizNames, moduleNames.value);
      currentRowData.value = row;
    };

    return () => (
      <Panel>
        <section class={cssModule.search}>
          <span class={cssModule['search-label']}>{t('组织')}：</span>
          <HcmSearchOrg
            ref={orgRef}
            class={cssModule['search-item']}
            {...{
              placeholder: '请选择或输入名称查找',
              searchPlaceholder: '请输入2个以上字符搜索',
              firstLevelDataFilter,
            }}
            v-model={organizations.value}
            v-model:checked={orgChecked.value}
          />
          <span class={cssModule['search-label']}>{t('业务')}：</span>
          <HcmSearchBusiness
            class={cssModule['search-item']}
            v-model={bkBizIds.value}
            multiple
            {...{ scope: 'auth', showAll: true }}
          />
          <span class={cssModule['search-label']}>{t('人员')}：</span>
          <HcmSearchUser class={cssModule['search-item']} v-model={operators.value} />
          <bk-button
            theme='primary'
            class={cssModule['search-button']}
            onClick={handleSearch}
            v-bk-tooltips={{
              content,
              disabled: isMeetSearchConditions.value,
            }}
            disabled={!isMeetSearchConditions.value}>
            {t('查询')}
          </bk-button>
          <bk-button class={cssModule['search-button']} onClick={handleReset}>
            {t('重置')}
          </bk-button>
          <export-to-excel-button
            data={dissloveList.value}
            text={t('导出')}
            columns={exportColumns.value}
            theme='primary'
            filename={t('整体裁撤信息')}
          />
        </section>

        <bk-loading loading={isLoading.value}>
          <bk-table
            show-overflow-tooltip
            data={dissloveList.value}
            max-height={'calc(100vh - 531px)'}
            class={cssModule.table}>
            <bk-table-column label={t('业务')} field='bk_biz_name' min-width='150px' fixed='left'></bk-table-column>
            <bk-table-column label={t('裁撤进度')} field='progress' min-width='150px'>
              {{
                default: ({ row }: { row: IDissolve }) => <>{getDisplayText(row?.progress)}</>,
              }}
            </bk-table-column>
            <bk-table-column label={t('原始数量')} field='total.origin.host_count' min-width='150px'>
              {{
                default: ({ row }: { row: IDissolve }) => {
                  return row.bk_biz_name !== '裁撤进度' ? (
                    <bk-button
                      text
                      theme='primary'
                      onClick={() => handleShowOriginDialog(row.bk_biz_name === '总数' ? [] : [row.bk_biz_name], row)}>
                      {getDisplayText(row?.total?.origin?.host_count)}
                    </bk-button>
                  ) : (
                    getDisplayText(row?.total?.origin?.host_count)
                  );
                },
              }}
            </bk-table-column>
            <bk-table-column label={t('原始CPU')} field='total.origin.cpu_count' min-width='150px'>
              {{
                default: ({ row }: { row: IDissolve }) => <>{getDisplayText(row?.total?.origin?.cpu_count)}</>,
              }}
            </bk-table-column>
            <bk-table-column label={t('当前数量')} field='total.current.host_count' min-width='150px'>
              {{
                default: ({ row }: { row: IDissolve }) => {
                  return row.bk_biz_name !== '裁撤进度' ? (
                    <bk-button
                      text
                      theme='primary'
                      onClick={() => handleShowCurrentDialog(row.bk_biz_name === '总数' ? [] : [row.bk_biz_name], row)}>
                      {getDisplayText(row?.total?.current?.host_count)}
                    </bk-button>
                  ) : (
                    getDisplayText(row?.total?.current?.host_count)
                  );
                },
              }}
            </bk-table-column>
            <bk-table-column label={t('当前CPU')} field='total.current.cpu_count' min-width='150px'>
              {{
                default: ({ row }: { row: IDissolve }) => <>{getDisplayText(row?.total?.current?.cpu_count)}</>,
              }}
            </bk-table-column>
          </bk-table>
        </bk-loading>
        <CurrentDialog
          v-model:isShow={currentDialogShow.value}
          searchParams={searchParams.value}
          rowData={currentRowData.value}></CurrentDialog>
        <OriginDialog
          v-model:isShow={originDialogShow.value}
          searchParams={searchParams.value}
          rowData={currentRowData.value}></OriginDialog>
      </Panel>
    );
  },
});
