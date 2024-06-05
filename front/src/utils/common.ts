/**
 * 获取实例的ip地址
 * @param inst 实例
 * @returns 实例的ip地址
 */
const getInstVip = (inst: any) => {
  const {
    private_ipv4_addresses,
    private_ipv6_addresses,
    public_ipv4_addresses,
    public_ipv6_addresses,
    private_ip_address,
    public_ip_address,
  } = inst;
  if (private_ipv4_addresses || private_ipv6_addresses || public_ipv4_addresses || public_ipv6_addresses) {
    if (public_ipv4_addresses.length > 0) return public_ipv4_addresses.join(',');
    if (public_ipv6_addresses.length > 0) return public_ipv6_addresses.join(',');
    if (private_ipv4_addresses.length > 0) return private_ipv4_addresses.join(',');
    if (private_ipv6_addresses.length > 0) return private_ipv6_addresses.join(',');
  }
  if (private_ip_address || public_ip_address) {
    if (private_ip_address.length > 0) return private_ip_address.join(',');
    if (public_ip_address.length > 0) return public_ip_address.join(',');
  }

  return '--';
};

/**
 * 清洗请求载荷，去除空值
 * @param payload 请求载荷
 * @returns 返回新的请求载荷
 */
const cleanPayload = (payload: any) => {
  const newPayload = {};
  Object.keys(payload).forEach((key) => {
    if (Object.prototype.hasOwnProperty.call(payload, key)) {
      const value = payload[key];
      if (value !== '' && !(Array.isArray(value) && value.length === 0)) {
        newPayload[key] = value;
      }
    }
  });
  return newPayload;
};

/**
 * 导出表格数据为 Excel
 * @param {Array} list 表格数据
 * @param {Array} columns 表格列
 * @param {String} filename 文件名，自动添加时间戳
 */
const exportTableToExcel = (list, columns, filename) => {
  import('@/vendor/Export2Excel').then((excel) => {
    const header = columns.map((col) => col.label);
    const data = list.map((item) =>
      columns.map((col) => {
        if (col.formatter) {
          return col.formatter({ [col.field]: item[col.field] });
        }

        if (col.exportFormatter) {
          return col.exportFormatter(item);
        }

        return item[col.field];
      }),
    );
    excel.export_json_to_excel({
      header,
      data,
      filename: `${filename}${getDate('yyyyMMddhhmmss')}`,
    });
  });
};
const getDate = (fmt, n) => {
  let d;
  if (n) {
    let nd = Date.parse(new Date());
    nd = nd + n * 86400000;
    d = new Date(nd);
  } else {
    d = new Date();
  }
  const o = {
    'M+': d.getMonth() + 1, // 月份
    'd+': d.getDate(), // 日
    'h+': d.getHours(), // 小时
    'm+': d.getMinutes(), // 分
    's+': d.getSeconds(), // 秒
    'q+': Math.floor((d.getMonth() + 3) / 3), // 季度
    S: d.getMilliseconds(), // 毫秒
  };

  if (/(y+)/.test(fmt)) {
    fmt = fmt.replace(RegExp.$1, `${d.getFullYear()}`.substr(4 - RegExp.$1.length));
  }
  Object.keys(o).forEach((k) => {
    if (new RegExp(`(${k})`).test(fmt)) {
      fmt = fmt.replace(RegExp.$1, RegExp.$1.length === 1 ? o[k] : `00${o[k]}`.substr(`${o[k]}`.length));
    }
  });
  return fmt;
};

// 拼接 接口 路径
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;
const getEntirePath = (tailPath: string, interfacePrefix = '/api/v1/woa/') => {
  return `${BK_HCM_AJAX_URL_PREFIX + interfacePrefix + tailPath}`;
};

export { getInstVip, exportTableToExcel, getEntirePath, cleanPayload };
