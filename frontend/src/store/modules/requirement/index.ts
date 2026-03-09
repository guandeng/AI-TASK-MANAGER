import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import type {
  Requirement,
  RequirementDocument,
  RequirementListParams,
  RequirementFormData,
  RequirementStatistics
} from '@/typings/api/requirement';
import {
  fetchRequirementList,
  fetchRequirementStatistics,
  fetchRequirementDetail,
  createRequirement,
  updateRequirement,
  deleteRequirement,
  uploadDocument,
  deleteDocument
} from '@/service/api/requirement';

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
        requirements.value = data;
      }
      return { data, error };
    } finally {
      loading.value = false;
    }
  }

  async function loadStatistics() {
    const { data, error } = await fetchRequirementStatistics();
    if (!error && data) {
      statistics.value = data;
    }
    return { data, error };
  }

  async function loadRequirementDetail(id: number) {
    detailLoading.value = true;
    try {
      const { data, error } = await fetchRequirementDetail(id);
      if (!error && data) {
        currentRequirement.value = data;
      }
      return { data, error };
    } finally {
      detailLoading.value = false;
    }
  }

  async function createNewRequirement(formData: RequirementFormData) {
    const { data, error } = await createRequirement(formData);
    if (!error && data) {
      requirements.value.unshift(data);
      await loadStatistics();
    }
    return { data, error };
  }

  async function updateRequirementById(id: number, formData: Partial<RequirementFormData>) {
    const { data, error } = await updateRequirement(id, formData);
    if (!error && data) {
      const index = requirements.value.findIndex(r => r.id === id);
      if (index !== -1) {
        requirements.value[index] = data;
      }
      if (currentRequirement.value?.id === id) {
        currentRequirement.value = data;
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
      if (!currentRequirement.value.documents) {
        currentRequirement.value.documents = [];
      }
      currentRequirement.value.documents.unshift(data);
    }
    return { data, error };
  }

  async function deleteRequirementDocument(requirementId: number, documentId: number) {
    const { data, error } = await deleteDocument(requirementId, documentId);
    if (!error && data && currentRequirement.value) {
      currentRequirement.value.documents = currentRequirement.value.documents?.filter(
        d => d.id !== documentId
      );
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
