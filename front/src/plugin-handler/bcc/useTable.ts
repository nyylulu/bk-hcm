import http from '@/http';
import { FetchDataType, fetchData as defaultFetchData } from '../useTable';

export const fetchData: FetchDataType = async (params: any) => {
  const { props, BK_HCM_AJAX_URL_PREFIX, pagination, sort, order } = params;

  let detailsRes;
  let countRes;

  if (typeof props.scrConfig === 'function') {
    const { url, payload } = props.scrConfig();
    const sortVal = `${sort.value}:${order.value === 'ASC' ? 1 : -1}`;
    [detailsRes, countRes] = await Promise.all(
      [false, true].map((isCount) =>
        http.post(
          BK_HCM_AJAX_URL_PREFIX + url,
          Object.assign(payload?.filter?.rules.length === 0 ? {} : payload, {
            page: {
              start: isCount ? 0 : pagination.start,
              limit: isCount ? 0 : pagination.limit,
              sort: isCount ? undefined : sortVal,
              enable_count: isCount,
            },
          }),
        ),
      ),
    );
  } else {
    [detailsRes, countRes] = await defaultFetchData(params);
  }

  return [detailsRes, countRes];
};
