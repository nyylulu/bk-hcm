import { defineComponent, type PropType, ref, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useZiyanScrStore } from '@/store/ziyanScr';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { useDepartment } from '@/hooks';
import { useUserStore } from '@/store';
import { getDisplayText } from '@/utils';
import ExportToExcelButton from '@/components/export-to-excel-button';
import Panel from '@/components/panel';
import OrganizationSelect from '@/components/OrganizationSelect/index';
import BusinessSelector from '@/components/business-selector/index.vue';
import MemberSelect from '@/components/MemberSelect';
import CurrentDialog from '../current-dialog';
import ModuleDialog from '../module-dialog';
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
    const businessMapStore = useBusinessMapStore();
    const userStore = useUserStore();

    const { departmentMap } = useDepartment();

    const columns = [
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
    const operators = ref([userStore.username]);
    const bkBizIds = ref([]);
    const dissloveList = ref<IDissolve[]>([]);
    const currentDialogShow = ref(false);
    const moduleDialogShow = ref(false);
    const originDialogShow = ref(false);
    const searchParams = ref();
    const moduleNames = ref<string[]>([]);
    const tableColumns = ref(columns);

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
      organizations.value.forEach((orgId) => {
        const dept = departmentMap.value?.get(orgId);
        if (dept) {
          collectLeafCodes(dept);
        }
      });

      return [...leafIds];
    });

    const handleSearch = async () => {
      isLoading.value = true;

      tableColumns.value = columns.concat(
        props.moduleNames.map((moduleName) => ({ label: moduleName, field: `module_host_count.${moduleName}` })),
      );

      moduleNames.value = [...props.moduleNames];
      ziyanScrStore
        .getDissolveList({
          group_ids: groupIds.value,
          bk_biz_names: bkBizIds.value.map(businessMapStore.getNameFromBusinessMap),
          module_names: moduleNames.value,
          operators: operators.value,
        })
        .then((result) => {
          const list = result?.data?.items || [];
          const fixedBizIds = ['total', 'recycle-progress'];
          dissloveList.value = list.sort((a, b) => {
            const countA = (a?.total?.current?.host_count || 0) as number;
            const countB = (b?.total?.current?.host_count || 0) as number;
            if (fixedBizIds.includes(a.bk_biz_id as string) || fixedBizIds.includes(b.bk_biz_id as string)) {
              return 0;
            }
            return countB - countA;
          });
        })
        .finally(() => {
          isLoading.value = false;
        });
    };

    const handleReset = () => {
      organizations.value = [];
      bkBizIds.value = [];
      operators.value = [];
    };

    const setSearchParams = (bkBizNames: string[], moduleNames: string[]) => {
      searchParams.value = {
        group_ids: groupIds.value,
        bk_biz_names: bkBizNames,
        module_names: moduleNames,
        operators: operators.value,
      };
    };

    const handleShowOriginDialog = (bkBizNames: string[]) => {
      originDialogShow.value = true;
      setSearchParams(bkBizNames, moduleNames.value);
    };

    const handleShowCurrentDialog = (bkBizNames: string[]) => {
      currentDialogShow.value = true;
      setSearchParams(bkBizNames, moduleNames.value);
    };

    const handleShowModuleDialog = (bkBizNames: string[], moduleName: string) => {
      moduleDialogShow.value = true;
      setSearchParams(bkBizNames, [moduleName]);
    };

    return () => (
      <Panel>
        <section class={cssModule.search}>
          <span class={cssModule['search-label']}>{t('组织')}：</span>
          <OrganizationSelect class={cssModule['search-item']} v-model={organizations.value}></OrganizationSelect>
          <span class={cssModule['search-label']}>{t('业务')}：</span>
          <BusinessSelector
            class={cssModule['search-item']}
            multiple
            isAudit={true}
            isShowAll={true}
            autoSelect={true}
            v-model={bkBizIds.value}></BusinessSelector>
          <span class={cssModule['search-label']}>{t('人员')}：</span>
          <MemberSelect
            class={cssModule['search-item']}
            v-model={operators.value}
            defaultUserlist={[
              {
                username: userStore.username,
                display_name: userStore.username,
              },
            ]}></MemberSelect>
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
            columns={tableColumns.value}
            theme='primary'
            filename={t('整体裁撤信息')}
          />
        </section>

        <bk-loading loading={isLoading.value}>
          <bk-table
            show-overflow-tooltip
            virtual-enabled={true}
            height='500px'
            data={dissloveList.value}
            class={cssModule.table}>
            <bk-table-column label={t('业务')} field='bk_biz_name' min-width='150px' fixed='left'></bk-table-column>
            <bk-table-column label={t('裁撤进度')} field='progress' min-width='150px'>
              {{
                default: ({ row }: { row: IDissolve }) => <>{getDisplayText(row?.progress)}</>,
              }}
            </bk-table-column>
            <bk-table-column label={t('原始数量')} field='total.origin.host_count' min-width='150px'>
              {{
                default: ({ row }: { row: IDissolve }) => (
                  <bk-button
                    text
                    theme='primary'
                    onClick={() =>
                      handleShowOriginDialog(['总数', '裁撤进度'].includes(row.bk_biz_name) ? [] : [row.bk_biz_name])
                    }>
                    {getDisplayText(row?.total?.origin?.host_count)}
                  </bk-button>
                ),
              }}
            </bk-table-column>
            <bk-table-column label={t('原始CPU')} field='total.origin.cpu_count' min-width='150px'>
              {{
                default: ({ row }: { row: IDissolve }) => <>{getDisplayText(row?.total?.origin?.cpu_count)}</>,
              }}
            </bk-table-column>
            <bk-table-column label={t('当前数量')} field='total.current.host_count' min-width='150px'>
              {{
                default: ({ row }: { row: IDissolve }) => (
                  <bk-button
                    text
                    theme='primary'
                    onClick={() =>
                      handleShowCurrentDialog(['总数', '裁撤进度'].includes(row.bk_biz_name) ? [] : [row.bk_biz_name])
                    }>
                    {getDisplayText(row?.total?.current?.host_count)}
                  </bk-button>
                ),
              }}
            </bk-table-column>
            <bk-table-column label={t('当前CPU')} field='total.current.cpu_count' min-width='150px'>
              {{
                default: ({ row }: { row: IDissolve }) => <>{getDisplayText(row?.total?.current?.cpu_count)}</>,
              }}
            </bk-table-column>
            {moduleNames.value.map((moduleName: string) => (
              <bk-table-column label={moduleName} field={moduleName} width={`${moduleName.length * 15}px`}>
                {{
                  default: ({ row }: { row: IDissolve }) => (
                    <bk-button
                      text
                      theme='primary'
                      onClick={() =>
                        handleShowModuleDialog(
                          ['总数', '裁撤进度'].includes(row.bk_biz_name) ? [] : [row.bk_biz_name],
                          moduleName,
                        )
                      }>
                      {getDisplayText(row?.module_host_count?.[moduleName])}
                    </bk-button>
                  ),
                }}
              </bk-table-column>
            ))}
          </bk-table>
        </bk-loading>
        <CurrentDialog v-model:isShow={currentDialogShow.value} searchParams={searchParams.value}></CurrentDialog>
        <ModuleDialog v-model:isShow={moduleDialogShow.value} searchParams={searchParams.value}></ModuleDialog>
        <OriginDialog v-model:isShow={originDialogShow.value} searchParams={searchParams.value}></OriginDialog>
      </Panel>
    );
  },
});
