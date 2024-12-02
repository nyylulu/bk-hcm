import { defineStore } from 'pinia';
import { ref } from 'vue';
import Cookies from 'js-cookie';

export default defineStore('usePagePermissionStore', () => {
  const hasPagePermission = ref(true);
  const permissionMsg = ref('');
  const setHasPagePermission = (val: boolean) => (hasPagePermission.value = val);
  const setPermissionMsg = (val: string) => (permissionMsg.value = val);

  const logout = () => {
    Cookies.remove('bk_token', { path: '/' });
    Cookies.remove('bk_ticket', { path: '/' });
    window.location.href = `${window.PROJECT_CONFIG.BK_LOGIN_URL}/?is_from_logout=1&c_url=${window.location.href}`;
  };

  return {
    hasPagePermission,
    setHasPagePermission,
    permissionMsg,
    setPermissionMsg,
    logout,
  };
});
