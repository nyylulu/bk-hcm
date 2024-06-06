import { PropType, defineComponent, reactive, ref } from 'vue';
import { Button, Dropdown, Message, TagInput } from 'bkui-vue';
import { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
import { AngleDownLine, Copy, Search } from 'bkui-vue/lib/icon';
import useClipboard from 'vue-clipboard3';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import RemoteTable from '@/components/RemoteTable';
import ExportToExcelButton from '@/components/export-to-excel-button';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
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
    const columnName = props.type === 'online' ? 'scrResourceOnlineHost' : 'scrResourceOfflineHost';
    const { columns } = useColumns(columnName);
    const url =
      props.type === 'online' ? '/api/v1/woa/pool/findmany/launch/host' : '/api/v1/woa/pool/findmany/recall/detail';
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

    const remoteTableRef = ref();
    const filterList = () => {
      remoteTableRef.value.pagination.start = 0;
      remoteTableRef.value.getDataList();
    };

    const copyIps = async () => {
      try {
        await toClipboard(remoteTableRef.value.dataList.map((item: any) => item.labels.ip).join('\n'));
        Message({ theme: 'success', message: '复制成功' });
      } catch (error) {
        Message({ theme: 'success', message: '复制失败: ', error });
      }
    };
    const copyAssetIds = async () => {
      try {
        await toClipboard(remoteTableRef.value.dataList.map((item: any) => item.labels.bk_asset_id).join('\n'));
        Message({ theme: 'success', message: '复制成功' });
      } catch (error) {
        Message({ theme: 'success', message: '复制失败: ', error });
      }
    };

    return () => (
      <div class='scr-resource-manage-detail-page'>
        <DetailHeader>
          <span class='header-title-prefix'>资源上架详情</span>
          <span class='header-title-content'>&nbsp;- ID {props.id}</span>
        </DetailHeader>
        <div class='detail-container'>
          <div class='detail-content'>
            <div class='operation-bar'>
              {props.type === 'online' && (
                <>
                  <BkRadioGroup v-model={filter.phase} class='mr8' onChange={filterList}>
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
                  <Button class='mr8' onClick={filterList}>
                    <Search />
                    查询
                  </Button>
                </>
              )}
              <ExportToExcelButton data={remoteTableRef.value?.dataList} columns={columns} class='mr8' />
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
              <RemoteTable
                ref={remoteTableRef}
                columns={columns}
                path={{ start: 'start', limit: 'limit', count: 'enable_count', data: 'info', total: 'count' }}
                apis={[
                  {
                    url,
                    payload: () => getPayload(),
                  },
                ]}
              />
            </div>
          </div>
        </div>
      </div>
    );
  },
});
