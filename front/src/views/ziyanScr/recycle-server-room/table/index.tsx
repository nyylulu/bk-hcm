import { defineComponent, type PropType, ref, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useZiyanScrStore } from '@/store/ziyanScr';
import { useBusinessMapStore } from '@/store/useBusinessMap';
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

    const isLoading = ref(false);
    const organizations = ref([]);
    const bkBizIds = ref([]);
    const operators = ref([]);
    const dissloveList = ref<IDissolve[]>([]);
    const currentDialogShow = ref(false);
    const moduleDialogShow = ref(false);
    const originDialogShow = ref(false);
    const searchParams = ref();
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
    const moduleNames = ref<string[]>([]);
    const tableColumns = ref(columns);
    const content = (
      <>
        {t('请勾选相关参数后查询，')}
        <br />
        {t('1.裁撤的机房模块，至少选一个模块，')}
        <br />
        {t('2.组织架构、业务、人员至少选择一个')}
      </>
    );

    const isMeetSearchConditions = computed(
      () => props.moduleNames.length && (organizations.value.length || bkBizIds.value.length || operators.value.length),
    );

    const handleSearch = async () => {
      isLoading.value = true;

      tableColumns.value = columns.concat(
        props.moduleNames.map((moduleName) => ({ label: moduleName, field: `module_host_count.${moduleName}` })),
      );

      moduleNames.value = [...props.moduleNames];
      ziyanScrStore
        .getDissolveList({
          organizations: organizations.value,
          bk_biz_names: bkBizIds.value.map(businessMapStore.getNameFromBusinessMap),
          module_names: moduleNames.value,
          operators: operators.value,
        })
        .then((data) => {
          dissloveList.value = data?.data?.items || [];
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

    const setSearchParams = (row: IDissolve) => {
      searchParams.value = {
        organizations: organizations.value,
        bk_biz_names: [row.bk_biz_name],
        module_names: Object.keys(row?.module_host_count || {}),
        operators: operators.value,
      };
    };

    const handleShowOriginDialog = (row: IDissolve) => {
      originDialogShow.value = true;
      setSearchParams(row);
    };

    const handleShowCurrentDialog = (row: IDissolve) => {
      currentDialogShow.value = true;
      setSearchParams(row);
    };

    const handleShowModuleDialog = (row: IDissolve) => {
      moduleDialogShow.value = true;
      setSearchParams(row);
    };

    return () => (
      <Panel>
        <section class={cssModule.search}>
          <OrganizationSelect class={cssModule['search-item']} v-model={organizations.value}></OrganizationSelect>
          <BusinessSelector
            class={cssModule['search-item']}
            multiple
            isAudit={true}
            v-model={bkBizIds.value}></BusinessSelector>
          <MemberSelect class={cssModule['search-item']} v-model={operators.value}></MemberSelect>
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
          <bk-table show-overflow-tooltip data={dissloveList.value} class={cssModule.table}>
            <bk-table-column label={t('业务')} field='bk_biz_name' fixed='left'></bk-table-column>
            <bk-table-column label={t('裁撤进度')} field='progress'></bk-table-column>
            <bk-table-column label={t('原始数量')} field='total.origin.host_count'>
              {{
                default: ({ row }: { row: IDissolve }) => (
                  <bk-button text theme='primary' onClick={() => handleShowOriginDialog(row)}>
                    {row?.total?.origin?.host_count}
                  </bk-button>
                ),
              }}
            </bk-table-column>
            <bk-table-column label={t('原始CPU')} field='total.origin.cpu_count'></bk-table-column>
            <bk-table-column label={t('当前数量')} field='total.current.host_count'>
              {{
                default: ({ row }: { row: IDissolve }) => (
                  <bk-button text theme='primary' onClick={() => handleShowCurrentDialog(row)}>
                    {row?.total?.current?.host_count}
                  </bk-button>
                ),
              }}
            </bk-table-column>
            <bk-table-column label={t('当前CPU')} field='total.current.cpu_count'></bk-table-column>
            {moduleNames.value.map((moduleName: string) => (
              <bk-table-column label={moduleName} field={moduleName}>
                {{
                  default: ({ row }: { row: IDissolve }) => (
                    <bk-button text theme='primary' onClick={() => handleShowModuleDialog(row)}>
                      {row?.module_host_count?.[moduleName]}
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
