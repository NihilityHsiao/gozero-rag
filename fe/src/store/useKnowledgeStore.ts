import { create } from 'zustand';
import { getKnowledgeList } from '@/api/knowledge';
import type { KnowledgeBaseInfo } from '@/types';

interface KnowledgeState {
  list: KnowledgeBaseInfo[];
  total: number;
  loading: boolean;
  page: number;
  pageSize: number;
  hasMore: boolean;
  fetchList: (page?: number, pageSize?: number, status?: number) => Promise<void>;
  reset: () => void;
}

export const useKnowledgeStore = create<KnowledgeState>((set, get) => ({
  list: [],
  total: 0,
  loading: false,
  page: 1,
  pageSize: 10,
  hasMore: true,
  fetchList: async (page = 1, pageSize = 10, status) => {
    // If loading, prevent duplicate requests
    if (get().loading) return;
    
    set({ loading: true });
    try {
      const res = await getKnowledgeList({ page, page_size: pageSize, status });
      const newList = res.list || [];
      const total = res.total;
      
      set((state) => {
        const updatedList = page === 1 ? newList : [...state.list, ...newList];
        return {
          list: updatedList,
          total: total,
          page: page,
          pageSize: pageSize,
          hasMore: updatedList.length < total,
        };
      });
    } catch (error) {
      console.error('Failed to fetch knowledge list', error);
      // In case of error, stop infinite scroll
      set({ hasMore: false });
    } finally {
      set({ loading: false });
    }
  },
  reset: () => {
    set({
      list: [],
      total: 0,
      page: 1,
      hasMore: true,
    });
  },
}));
