import http from '@/http';
import { FetchDataType, fetchData as defaultFetchData } from '../useTable';

export const fetchData: FetchDataType = async (params: any) => {
  const { props, BK_HCM_AJAX_URL_PREFIX, pagination } = params;

  let detailsRes;
  let countRes;

  if (typeof props.scrConfig === 'function') {
    const { url, payload } = props.scrConfig();
    [detailsRes, countRes] = await Promise.all(
      [false, true].map((isCount) =>
        http.post(
          BK_HCM_AJAX_URL_PREFIX + url,
          Object.assign(payload, {
            page: {
              start: isCount ? 0 : pagination.start,
              limit: isCount ? 0 : pagination.limit,
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
