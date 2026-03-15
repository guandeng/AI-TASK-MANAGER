import { computed, ref } from 'vue';
import { defineStore } from 'pinia';
import {
  createRequirement,
  deleteDocument,
  deleteRequirement,
  fetchRequirementDetail,
  fetchRequirementList,
  fetchRequirementStatistics,
  updateRequirement,
  uploadDocument
} from '@/service/api/requirement';
import type {
  Requirement,
  RequirementDocument,
  RequirementFormData,
  RequirementListParams,
  RequirementStatistics
} from '@/typings/api/requirement';

// 辅助函数：提取后端返回的 data 字段
function extractData(responseData: any): any {
  return responseData?.data || responseData;
}

export const useRequirementStore = defineStore('requirement-store', () => {
  // 状态
  const requirements = ref<Requirement[]>([]);
  const currentRequirement = ref<Requirement | null>(null);
  const statistics = ref<RequirementStatistics | null>(null);
  const loading = ref(false);
  const detailLoading = ref(false);

  // 计算属性
  const requirementsByStatus = computed(() => {
    const grouped: Record<string, Requirement[]> = {
      draft: [],
      active: [],
      completed: [],
      archived: []
    };

    requirements.value.forEach(req => {
      if (grouped[req.status]) {
        grouped[req.status].push(req);
      }
    });

    return grouped;
  });

  const requirementsByPriority = computed(() => {
    const grouped: Record<string, Requirement[]> = {
      high: [],
      medium: [],
      low: []
    };

    requirements.value.forEach(req => {
      if (grouped[req.priority]) {
        grouped[req.priority].push(req);
      }
    });

    return grouped;
  });

  // Actions
  async function loadRequirements(params?: RequirementListParams) {
    loading.value = true;
    try {
      const { data, error } = await fetchRequirementList(params);
      if (!error && data) {
        const responseData = extractData(data);

        if (Array.isArray(responseData)) {
          requirements.value = responseData;
        } else if (responseData && 'list' in responseData) {
          requirements.value = responseData.list || [];
        }
      }
      return { data, error };
    } finally {
      loading.value = false;
    }
  }

  async function loadStatistics() {
    const { data, error } = await fetchRequirementStatistics();
    if (!error && data) {
      statistics.value = extractData(data);
    }
    return { data, error };
  }

  async function loadRequirementDetail(id: number) {
    detailLoading.value = true;
    try {
      const { data, error } = await fetchRequirementDetail(id);
      if (!error && data) {
        currentRequirement.value = extractData(data);
      }
      return { data, error };
    } finally {
      detailLoading.value = false;
    }
  }

  async function createNewRequirement(formData: RequirementFormData) {
    const { data, error } = await createRequirement(formData);
    if (!error && data) {
      const newReq = extractData(data);
      requirements.value.unshift(newReq);
      await loadStatistics();
    }
    return { data, error };
  }

  async function updateRequirementById(id: number, formData: Partial<RequirementFormData>) {
    const { data, error } = await updateRequirement(id, formData);
    if (!error && data) {
      const updatedReq = extractData(data);
      const index = requirements.value.findIndex(r => r.id === id);
      if (index !== -1) {
        requirements.value[index] = updatedReq;
      }
      if (currentRequirement.value?.id === id) {
        currentRequirement.value = updatedReq;
      }
    }
    return { data, error };
  }

  async function deleteRequirementById(id: number) {
    const { data, error } = await deleteRequirement(id);
    if (!error && data) {
      requirements.value = requirements.value.filter(r => r.id !== id);
      if (currentRequirement.value?.id === id) {
        currentRequirement.value = null;
      }
      await loadStatistics();
    }
    return { data, error };
  }

  async function uploadRequirementDocument(id: number, file: File, uploadedBy?: string) {
    const { data, error } = await uploadDocument(id, file, uploadedBy);
    if (!error && data && currentRequirement.value) {
      const doc = extractData(data);
      if (!currentRequirement.value.documents) {
        currentRequirement.value.documents = [];
      }
      currentRequirement.value.documents.unshift(doc);
    }
    return { data, error };
  }

  async function deleteRequirementDocument(requirementId: number, documentId: number) {
    const { data, error } = await deleteDocument(requirementId, documentId);
    if (!error && data && currentRequirement.value) {
      currentRequirement.value.documents = currentRequirement.value.documents?.filter(d => d.id !== documentId);
    }
    return { data, error };
  }

  function clearCurrentRequirement() {
    currentRequirement.value = null;
  }

  return {
    // 状态
    requirements,
    currentRequirement,
    statistics,
    loading,
    detailLoading,
    // 计算属性
    requirementsByStatus,
    requirementsByPriority,
    // Actions
    loadRequirements,
    loadStatistics,
    loadRequirementDetail,
    createNewRequirement,
    updateRequirementById,
    deleteRequirementById,
    uploadRequirementDocument,
    deleteRequirementDocument,
    clearCurrentRequirement
  };
});
