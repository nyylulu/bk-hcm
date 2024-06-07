import { PropType, defineComponent, reactive, computed, onMounted } from 'vue';
import { Button, Dropdown, Message, TagInput } from 'bkui-vue';
import { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
import { AngleDownLine, Copy, Search } from 'bkui-vue/lib/icon';
import useClipboard from 'vue-clipboard3';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import ExportToExcelButton from '@/components/export-to-excel-button';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useTable } from '@/hooks/useTable/useTable';
import './index.scss';

interface ResourceManageDetailFilterType {
  phase: string;
  ip: string[];
  assetId: string[];
}

export default defineComponent({
  name: 'ScrResourceManageDetail',
  props: { id: String, type: String as PropType<'online' | 'offline'> },
  setup(props) {
    const { toClipboard } = useClipboard();

    const { columns: scrResourceOnlineHostColumns } = useColumns('scrResourceOnlineHost');
    const { columns: scrResourceOfflineHostColumns } = useColumns('scrResourceOfflineHost');

    const {
      CommonTable: ScrResourceOnlineHostTable,
      getListData: reloadScrResourceOnlineHostTable,
      dataList: scrResourceOnlineHostDataList,
      pagination: scrResourceOnlineHostPagination,
    } = useTable({
      tableOptions: { columns: scrResourceOnlineHostColumns },
      requestOption: { dataPath: 'data.info' },
      scrConfig: () => ({
        url: '/api/v1/woa/pool/findmany/launch/host',
        payload: {
          ...getPayload(),
        },
      }),
    });

    const {
      CommonTable: ScrResourceOfflineHostTable,
      getListData: reloadScrResourceOfflineHostTable,
      pagination: scrResourceOfflineHostPagination,
      dataList: scrResourceOfflineHostDataList,
    } = useTable({
      tableOptions: { columns: scrResourceOfflineHostColumns },
      requestOption: { dataPath: 'data.info' },
      scrConfig: () => ({
        url: '/api/v1/woa/pool/findmany/recall/detail',
        payload: {
          ...getPayload(),
        },
      }),
    });

    const renderDataList = computed(() => {
      return props.type === 'online' ? scrResourceOnlineHostDataList.value : scrResourceOfflineHostDataList.value;
    });

    const renderColumns = computed(() => {
      return props.type === 'online' ? scrResourceOnlineHostColumns : scrResourceOfflineHostColumns;
    });

    const getDefaultFilter = (): ResourceManageDetailFilterType => ({
      phase: 'ALL',
      ip: [],
      assetId: [],
    });
    const filter = reactive(getDefaultFilter());
    const getPayload = () => {
      const { phase, ip, assetId } = filter;
      const phaseRule = phase !== 'ALL' ? { field: 'phase', operator: 'equal', value: phase } : undefined;
      const ipRule = ip.length ? { field: 'labels.ip', operator: 'in', value: ip } : undefined;
      const assetIdRule = assetId.length ? { field: 'labels.bk_asset_id', operator: 'in', value: assetId } : undefined;

      return {
        filter:
          phaseRule || ipRule || assetIdRule
            ? {
                condition: 'AND',
                rules: [phaseRule, ipRule, assetIdRule].filter(Boolean),
              }
            : undefined,
        id: Number(props.id),
      };
    };

    const reloadDataList = () => {
      if (props.type === 'online') {
        scrResourceOnlineHostPagination.start = 0;
        reloadScrResourceOnlineHostTable();
      } else {
        scrResourceOfflineHostPagination.start = 0;
        reloadScrResourceOfflineHostTable();
      }
    };

    const copyIps = async () => {
      try {
        await toClipboard(renderDataList.value.map((item: any) => item.labels.ip).join('\n'));
        Message({ theme: 'success', message: '复制成功' });
      } catch (error) {
        Message({ theme: 'success', message: '复制失败: ', error });
      }
    };
    const copyAssetIds = async () => {
      try {
        await toClipboard(renderDataList.value.map((item: any) => item.labels.bk_asset_id).join('\n'));
        Message({ theme: 'success', message: '复制成功' });
      } catch (error) {
        Message({ theme: 'success', message: '复制失败: ', error });
      }
    };

    onMounted(() => {
      reloadDataList();
    });

    return () => (
      <div class='scr-resource-manage-detail-page'>
        <DetailHeader>
          <span class='header-title-prefix'>资源上架详情</span>
          <span class='header-title-content'>&nbsp;- ID {props.id}</span>
        </DetailHeader>
        <div class='common-sub-main-container'>
          <div class='sub-main-content'>
            <div class='operation-bar'>
              {props.type === 'online' && (
                <>
                  <BkRadioGroup v-model={filter.phase} class='mr8' onChange={reloadDataList}>
                    <BkRadioButton label='ALL'>全部</BkRadioButton>
                    <BkRadioButton label='SUCCESS'>成功</BkRadioButton>
                    <BkRadioButton label='FAILED'>失败</BkRadioButton>
                    <BkRadioButton label='RUNNING'>执行中</BkRadioButton>
                  </BkRadioGroup>
                  <TagInput
                    class='w200 mr8'
                    v-model={filter.ip}
                    allow-create
                    collapse-tags
                    placeholder='请输入内网 IP'
                  />
                  <TagInput
                    class='w200 mr8'
                    v-model={filter.assetId}
                    allow-create
                    collapse-tags
                    placeholder='请输入固资号'
                  />
                  <Button class='mr8' onClick={reloadDataList}>
                    <Search />
                    查询
                  </Button>
                </>
              )}
              <ExportToExcelButton data={renderDataList.value} columns={renderColumns.value} class='mr8' />
              <Dropdown>
                {{
                  default: () => (
                    <Button theme='primary' outline>
                      <Copy />
                      复制
                      <AngleDownLine />
                    </Button>
                  ),
                  content: () => (
                    <Dropdown.DropdownMenu>
                      <Dropdown.DropdownItem onClick={copyIps}>IP</Dropdown.DropdownItem>
                      <Dropdown.DropdownItem onClick={copyAssetIds}>固资号</Dropdown.DropdownItem>
                    </Dropdown.DropdownMenu>
                  ),
                }}
              </Dropdown>
            </div>
            <div class='table-container'>
              {props.type === 'online' ? <ScrResourceOnlineHostTable /> : <ScrResourceOfflineHostTable />}
            </div>
          </div>
        </div>
      </div>
    );
  },
});
