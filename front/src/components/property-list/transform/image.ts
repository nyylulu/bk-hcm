import { ref } from 'vue';
import { getImages } from '@/api/host/cvm';
export const useImages = () => {
  const imageList = ref([]);
  const fetchDiskTypes = async () => {
    const res = await getImages({ region: [] });
    imageList.value = res?.data?.info || [];
  };
  fetchDiskTypes();
  const getImageName = (id) => imageList.value.find((item) => item.image_id === id)?.image_name || id;
  return { getImageName };
};
