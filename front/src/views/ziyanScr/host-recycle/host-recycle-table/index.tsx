import { defineComponent, ref, computed, watch, onMounted } from 'vue';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import { useTable } from '@/hooks/useTable/useTable';
import { getRecycleStageOpts, retryOrder, submitOrder, stopOrder } from '@/api/host/recycle';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import { Button, Message, Dropdown } from 'bkui-vue';
import GridContainer from '@/components/layout/grid-container/grid-container.vue';
import GridItemFormElement from '@/components/layout/grid-container/grid-item-form-element.vue';
import GridItem from '@/components/layout/grid-container/grid-item.vue';
import ExportToExcelButton from '@/components/export-to-excel-button';
import { Plus } from 'bkui-vue/lib/icon';
import './index.scss';
import { useRoute, useRouter } from 'vue-router';
import useSearchQs from '@/hooks/use-search-qs';
import { useBusinessGlobalStore } from '@/store/business-global';
import { getDateRange, transformFlatCondition } from '@/utils/search';
import type { ModelProperty } from '@/model/typings';
import { getModel } from '@/model/manager';
import HocSearch from '@/model/hoc-search.vue';
import { HostRecycleSearchNonBusiness } from '@/model/order/host-recycle-search';
import { serviceShareBizSelectedKey } from '@/constants/storage-symbols';

const { DropdownMenu, DropdownItem } = Dropdown;

export default defineComponent({
  components: {
    ExportToExcelButton,
  },
  setup() {
    const currentOperateRowIndex = ref(-1);
    const router = useRouter();
    const route = useRoute();

    const businessGlobalStore = useBusinessGlobalStore();

    const searchFields = getModel(HostRecycleSearchNonBusiness).getProperties();
    const searchQs = useSearchQs({ key: 'filter', properties: searchFields });

    const condition = ref<Record<string, any>>({});
    const searchValues = ref<Record<string, any>>({});

    const stageList = ref([]);
    const fetchStageList = async () => {
      const data = await getRecycleStageOpts();
      stageList.value = data?.info || [];
    };

    const goToPrecheck = () => {
      router.push({
        path: '/service/hostRecycling/preDetail',
        query: {
          suborder_id: getBatchSuborderId().join('\n'),
        },
      });
    };
    const getBatchSuborderId = () => {
      return selections.value.map((item) => {
        return item.suborder_id;
      });
    };
    const loadHostRecycle = () => {
      getListData();
    };
    const textTip = (text, theme) => {
      const themeDes = {
        error: '失败',
        success: '成功',
      };
      Message({
        message: `${text}${themeDes[theme]}`,
        theme,
        duration: 1500,
      });
    };
    const retryOrderFunc = (id: string, disabled) => {
      if (disabled) return;
      const suborderId = id === 'isBatch' ? getBatchSuborderId() : [id];
      retryOrder(
        {
          suborderId,
        },
        {},
      ).then((res) => {
        if (res.code === 0) {
          textTip('重试', 'success');
          loadHostRecycle();
        }
      });
    };
    const stopOrderFunc = (id: string, disabled) => {
      if (disabled) return;
      stopOrder(
        {
          suborderId: [id],
        },
        {},
      ).then((res) => {
        if (res.code === 0) {
          textTip('终止', 'success');
          loadHostRecycle();
        }
      });
    };
    const submitOrderFunc = (id: string, disabled) => {
      if (disabled) return;
      const suborderId = id === 'isBatch' ? getBatchSuborderId() : [id];
      submitOrder(
        {
          suborderId,
        },
        {},
      ).then((res) => {
        if (res.code === 0) {
          textTip('去除预检失败IP提交', 'success');
          loadHostRecycle();
        }
      });
    };
    const { selections, handleSelectionChange } = useSelection();
    const { columns } = useColumns('hostRecycleApplication');
    const enterDetail = (row) => {
      router.push({
        path: '/service/hostRecycling/docDetail',
        query: {
          suborderId: row.suborder_id,
          bkBizId: row.bk_biz_id,
        },
      });
    };

    const opBtnDisabled = computed(() => {
      return (status) => {
        if (
          ['UNCOMMIT', 'COMMITTED', 'DETECTING', 'FOR_AUDIT', 'TRANSITING', 'RETURNING', 'DONE', 'TERMINATE'].includes(
            status,
          )
        ) {
          return true;
        }
        return false;
      };
    });
    const operateColList = [
      {
        label: '操作',
        width: 120,
        showOverflowTooltip: false,
        render: ({ row, index }) => {
          return (
            <div class='operation-column'>
              <Button text theme='primary' class='mr10' onClick={() => returnPreDetails(row)}>
                预检详情
              </Button>
              <Dropdown
                trigger='click'
                popoverOptions={{
                  renderType: 'shown',
                  onAfterShow: () => (currentOperateRowIndex.value = index),
                  onAfterHidden: () => (currentOperateRowIndex.value = -1),
                }}>
                {{
                  default: () => (
                    <div class={`more-action${currentOperateRowIndex.value === index ? ' current-operate-row' : ''}`}>
                      <i class='hcm-icon bkhcm-icon-more-fill' />
                    </div>
                  ),
                  content: () => (
                    <DropdownMenu>
                      <DropdownItem
                        key='retry'
                        onClick={() => retryOrderFunc(row.suborder_id, opBtnDisabled.value(row.status))}
                        extCls={`more-action-item${opBtnDisabled.value(row.status) ? ' disabled' : ''}`}>
                        全部重试
                      </DropdownItem>
                      <DropdownItem
                        key='stop'
                        onClick={() => stopOrderFunc(row.suborder_id, opBtnDisabled.value(row.status))}
                        extCls={`more-action-item${opBtnDisabled.value(row.status) ? ' disabled' : ''}`}>
                        全部终止
                      </DropdownItem>
                      <DropdownItem
                        key='submit'
                        onClick={() => submitOrderFunc(row.suborder_id, opBtnDisabled.value(row.status))}
                        extCls={`more-action-item${opBtnDisabled.value(row.status) ? ' disabled' : ''}`}>
                        剔除预检失败IP重试
                      </DropdownItem>
                    </DropdownMenu>
                  ),
                }}
              </Dropdown>
            </div>
          );
        },
      },
    ];
    // 在第三个加子单号，需要跳转到单据详情，未用到路由
    columns.splice(1, 0, {
      label: '单号/子单号',
      width: 100,
      render: ({ row }) => {
        return (
          <div>
            <div>
              <p>{row.order_id}</p>
            </div>
            <div>
              <Button theme='primary' text onClick={() => enterDetail(row)}>
                {row.suborder_id}
              </Button>
            </div>
          </div>
        );
      },
    });
    const tableColumns = [...columns, ...operateColList];
    const { CommonTable, getListData, dataList } = useTable({
      tableOptions: {
        columns: tableColumns,
        extra: {
          onSelect: (selections: any) => {
            handleSelectionChange(selections, () => true, false);
          },
          onSelectAll: (selections: any) => {
            handleSelectionChange(selections, () => true, true);
          },
        },
      },
      requestOption: {
        dataPath: 'data.info',
        sortOption: {
          sort: 'create_at',
          order: 'DESC',
        },
        immediate: false,
      },
      scrConfig: () => {
        const payload = transformFlatCondition(condition.value, searchFields);
        if (payload.bk_biz_id?.[0] === 0) {
          payload.bk_biz_id = businessGlobalStore.businessAuthorizedList.map((item: any) => item.id);
        }
        return {
          url: '/api/v1/woa/task/findmany/recycle/order',
          payload,
        };
      },
    });

    const getSearchCompProps = (field: ModelProperty) => {
      if (field.type === 'business') {
        return {
          scope: 'auth',
          showAll: true,
          emptySelectAll: true,
          cacheKey: serviceShareBizSelectedKey,
        };
      }
      if (field.id === 'create_at') {
        return {
          type: 'daterange',
          format: 'yyyy-MM-dd',
        };
      }
      if (field.id === 'order_id') {
        return {
          pasteFn: (value: string) =>
            value
              .split(/\r\n|\n|\r/)
              .filter((tag) => /^\d+$/.test(tag))
              .map((tag) => ({ id: tag, name: tag })),
        };
      }
      if (field.id === 'suborder_id') {
        return {
          pasteFn: (value: string) => value.split(/\r\n|\n|\r/).map((tag) => ({ id: tag, name: tag })),
        };
      }
      if (field.id === 'stage') {
        const stages = stageList.value.reduce((acc, { stage, description }) => {
          acc[stage] = description;
          return acc;
        }, {});
        return {
          option: stages,
        };
      }
      if (field.id === 'recycle_type') {
        return {
          useNameValue: true,
        };
      }
      return {
        option: field.option,
      };
    };

    const handleSearch = () => {
      searchQs.set(searchValues.value);
    };

    const handleReset = () => {
      searchQs.clear();
    };

    watch(
      () => route.query,
      async (query) => {
        const defaultCondition = {
          create_at: getDateRange('last30d', true),
          bk_biz_id: businessGlobalStore.getCacheSelected(serviceShareBizSelectedKey) ?? [0],
        };
        condition.value = searchQs.get(query, defaultCondition);

        searchValues.value = condition.value;

        getListData();
      },
      { immediate: true },
    );

    const returnPreDetails = (row: { suborder_id: any; bk_biz_id: any }) => {
      router.push({
        path: '/service/hostRecycling/preDetail',
        query: {
          suborder_id: row.suborder_id,
        },
      });
    };
    const returnRecyclingResources = () => {
      router.push({
        path: '/service/hostRecycling/resources',
        query: { ...route.query },
      });
    };

    const renderNodes = () => {
      return (
        <div class={'apply-list-container'}>
          <div class={'filter-container'} style={{ margin: '0 24px 20px 24px' }}>
            <GridContainer layout='vertical' column={4} content-min-width={'1fr'} gap={[16, 60]}>
              {searchFields.map((field) => (
                <GridItemFormElement key={field.id} label={field.name}>
                  <HocSearch
                    is={field.type}
                    display={field.meta?.display}
                    v-model={searchValues.value[field.id]}
                    {...getSearchCompProps(field)}
                  />
                </GridItemFormElement>
              ))}
              <GridItem span={4}>
                <div style={{ display: 'flex', gap: '8px' }}>
                  <bk-button theme='primary' style={{ minWidth: '86px' }} onClick={handleSearch}>
                    查询
                  </bk-button>
                  <bk-button style={{ minWidth: '86px' }} onClick={handleReset}>
                    重置
                  </bk-button>
                </div>
              </GridItem>
            </GridContainer>
          </div>
          <div class='btn-container oper-btn-pad'>
            <Button theme='primary' onClick={returnRecyclingResources}>
              <Plus />
              回收资源
            </Button>
            <export-to-excel-button data={dataList.value} columns={tableColumns} filename='回收单据列表' />
            <Button disabled={!selections.value.length} onClick={goToPrecheck}>
              批量查看预检详情
            </Button>
            <Button disabled={!selections.value.length} onClick={() => retryOrderFunc('isBatch', false)}>
              批量重试
            </Button>
            <Button disabled={!selections.value.length} onClick={() => submitOrderFunc('isBatch', false)}>
              批量去除预检失败IP提交
            </Button>
          </div>
          <CommonTable class={'filter-common-table'} />
        </div>
      );
    };

    onMounted(() => {
      fetchStageList();
    });

    return renderNodes;
  },
});
