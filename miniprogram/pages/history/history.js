const {
  deleteAnalysis,
  deleteDivination,
  getAnalysisList,
  getDivinationHistory,
  getQimenAnalysisList,
} = require("../../utils/api");
const { buildBaziHistoryListItem } = require("../../utils/bazi");
const { buildQimenHistoryListItem } = require("../../utils/qimen");
const { formatDateTime } = require("../../utils/date");
const { ERROR_TYPES, isBusinessError } = require("../../utils/request");

const PAGE_SIZE = 20;

const FILTERS = Object.freeze([
  { value: "all", label: "全部" },
  { value: "divination", label: "问事起卦" },
  { value: "bazi", label: "八字简析" },
  { value: "qimen", label: "奇门问事" },
]);

function buildDivinationHistoryItem(item) {
  if (!item || !item.id) return null;

  const primaryName = item.primary_hexagram?.name || "未知";
  const changedName = item.changed_hexagram?.name || "未知";
  const categoryName = item.category?.name || "未分类";

  return {
    key: `divination-${item.id}`,
    recordType: "divination",
    id: item.id,
    typeLabel: "问事起卦",
    title: `本卦 ${primaryName} → 变卦 ${changedName}`,
    summary: `事项类型 · ${categoryName}`,
    statusText: Number(item.unlock_status) === 1 ? "已解锁" : "已生成",
    created_at: item.created_at || "",
    createdAtDisplay: formatDateTime(item.created_at) || "—",
    detailUrl: `/pages/result/result?id=${item.id}`,
    canDelete: true,
  };
}

function sortByCreatedAtDesc(items) {
  return items.slice().sort((left, right) => {
    const leftTime = Date.parse(left.created_at || "") || 0;
    const rightTime = Date.parse(right.created_at || "") || 0;
    if (rightTime !== leftTime) return rightTime - leftTime;
    return Number(right.id || 0) - Number(left.id || 0);
  });
}

function filterVisibleItems(filter, divItems, baziItems, qimenItems) {
  const pool = [];
  if (filter === "all" || filter === "divination") {
    pool.push(...divItems);
  }
  if (filter === "all" || filter === "bazi") {
    pool.push(...baziItems);
  }
  if (filter === "all" || filter === "qimen") {
    pool.push(...qimenItems);
  }
  return sortByCreatedAtDesc(pool);
}

function emptyStateCopy(filter) {
  switch (filter) {
    case "divination":
      return {
        title: "还没有问事起卦记录",
        description: "可以从首页进入「问事起卦」开始一次问事。",
        actionUrl: "/pages/ask/ask",
        actionText: "开始问事起卦 →",
      };
    case "bazi":
      return {
        title: "还没有八字简析记录",
        description: "可以从首页进入「八字简析」创建记录。",
        actionUrl: "/pages/bazi/bazi",
        actionText: "开始八字简析 →",
      };
    case "qimen":
      return {
        title: "还没有奇门问事记录",
        description: "可以从首页进入「奇门问事」创建记录。",
        actionUrl: "/pages/qimen/qimen",
        actionText: "开始奇门问事 →",
      };
    default:
      return {
        title: "还没有记录",
        description: "可以从首页开始一次问事、八字简析或奇门问事。",
        actionUrl: "/pages/index/index",
        actionText: "返回首页 →",
      };
  }
}

function mapDeleteError(error) {
  if (!error) return "删除失败，请稍后重试。";
  if (error.type === ERROR_TYPES.NETWORK) {
    return "网络异常，请检查网络后重试。";
  }
  if (isBusinessError(error, 40001)) {
    return "会话已失效，请重新进入页面。";
  }
  if (isBusinessError(error, 40401)) {
    return "记录不存在或已被删除。";
  }
  return "删除失败，请稍后重试。";
}

function shouldFetchDivination(filter) {
  return filter === "all" || filter === "divination";
}

function shouldFetchBazi(filter) {
  return filter === "all" || filter === "bazi";
}

function shouldFetchQimen(filter) {
  return filter === "all" || filter === "qimen";
}

Page({
  data: {
    filters: FILTERS,
    activeFilter: "all",
    visibleItems: [],
    divItems: [],
    baziItems: [],
    qimenItems: [],
    divPage: 0,
    baziPage: 0,
    qimenPage: 0,
    hasMoreDiv: false,
    hasMoreBazi: false,
    hasMoreQimen: false,
    hasMore: false,
    loading: true,
    loadingMore: false,
    deletingId: null,
    error: "",
    emptyTitle: "",
    emptyDescription: "",
    emptyActionUrl: "",
    emptyActionText: "",
  },

  onLoad() {
    this.loadHistory(true);
  },

  onShow() {
    if (this.loadedOnce && !this.loadingRequest) {
      this.loadHistory(true);
    }
  },

  onPullDownRefresh() {
    this.loadHistory(true, true);
  },

  onReachBottom() {
    this.loadMore();
  },

  onFilterTap(event) {
    const filter = event.currentTarget.dataset.filter;
    if (!filter || filter === this.data.activeFilter) return;

    this.setData({ activeFilter: filter }, () => {
      this.loadHistory(true);
    });
  },

  loadMore() {
    if (!this.data.hasMore || this.data.loadingMore || this.loadingRequest) return;
    this.loadHistory(false);
  },

  updateVisibleState(extra = {}) {
    const emptyCopy = emptyStateCopy(this.data.activeFilter);
    const visibleItems = filterVisibleItems(
      this.data.activeFilter,
      this.data.divItems,
      this.data.baziItems,
      this.data.qimenItems
    );
    const hasMore =
      (shouldFetchDivination(this.data.activeFilter) && this.data.hasMoreDiv) ||
      (shouldFetchBazi(this.data.activeFilter) && this.data.hasMoreBazi) ||
      (shouldFetchQimen(this.data.activeFilter) && this.data.hasMoreQimen);

    this.setData({
      visibleItems,
      hasMore,
      emptyTitle: emptyCopy.title,
      emptyDescription: emptyCopy.description,
      emptyActionUrl: emptyCopy.actionUrl,
      emptyActionText: emptyCopy.actionText,
      ...extra,
    });
  },

  async loadHistory(reset = true, fromPullDown = false) {
    if (this.loadingRequest) {
      if (fromPullDown) wx.stopPullDownRefresh();
      return;
    }

    const filter = this.data.activeFilter;
    const divPage = reset ? 1 : this.data.divPage + 1;
    const baziPage = reset ? 1 : this.data.baziPage + 1;
    const qimenPage = reset ? 1 : this.data.qimenPage + 1;

    this.loadingRequest = true;
    this.setData({
      loading: reset && this.data.visibleItems.length === 0,
      loadingMore: !reset,
      error: "",
    });

    try {
      const tasks = [];

      if (shouldFetchDivination(filter)) {
        tasks.push(
          getDivinationHistory({ page: divPage, page_size: PAGE_SIZE })
            .then((result) => ({ kind: "divination", result }))
            .catch((error) => ({ kind: "divination", error }))
        );
      }
      if (shouldFetchBazi(filter)) {
        tasks.push(
          getAnalysisList({ page: baziPage, page_size: PAGE_SIZE })
            .then((result) => ({ kind: "bazi", result }))
            .catch((error) => ({ kind: "bazi", error }))
        );
      }
      if (shouldFetchQimen(filter)) {
        tasks.push(
          getQimenAnalysisList({ page: qimenPage, page_size: PAGE_SIZE })
            .then((result) => ({ kind: "qimen", result }))
            .catch((error) => ({ kind: "qimen", error }))
        );
      }

      const responses = await Promise.all(tasks);
      const firstError = responses.find((entry) => entry.error)?.error;
      if (firstError) {
        throw firstError;
      }

      let divItems = reset ? [] : this.data.divItems.slice();
      let baziItems = reset ? [] : this.data.baziItems.slice();
      let qimenItems = reset ? [] : this.data.qimenItems.slice();
      let hasMoreDiv = reset ? false : this.data.hasMoreDiv;
      let hasMoreBazi = reset ? false : this.data.hasMoreBazi;
      let hasMoreQimen = reset ? false : this.data.hasMoreQimen;

      responses.forEach((entry) => {
        const items = Array.isArray(entry.result?.items) ? entry.result.items : [];
        const total = Number(entry.result?.total || 0);

        if (entry.kind === "divination") {
          const mapped = items.map(buildDivinationHistoryItem).filter(Boolean);
          divItems = reset ? mapped : divItems.concat(mapped);
          hasMoreDiv = divPage * PAGE_SIZE < total && mapped.length > 0;
        }
        if (entry.kind === "bazi") {
          const mapped = items.map(buildBaziHistoryListItem).filter(Boolean);
          baziItems = reset ? mapped : baziItems.concat(mapped);
          hasMoreBazi = baziPage * PAGE_SIZE < total && mapped.length > 0;
        }
        if (entry.kind === "qimen") {
          const mapped = items.map(buildQimenHistoryListItem).filter(Boolean);
          qimenItems = reset ? mapped : qimenItems.concat(mapped);
          hasMoreQimen = qimenPage * PAGE_SIZE < total && mapped.length > 0;
        }
      });

      this.setData(
        {
          divItems,
          baziItems,
          qimenItems,
          divPage: shouldFetchDivination(filter) ? divPage : this.data.divPage,
          baziPage: shouldFetchBazi(filter) ? baziPage : this.data.baziPage,
          qimenPage: shouldFetchQimen(filter) ? qimenPage : this.data.qimenPage,
          hasMoreDiv,
          hasMoreBazi,
          hasMoreQimen,
        },
        () => {
          this.updateVisibleState();
        }
      );
      this.loadedOnce = true;
    } catch (error) {
      this.setData({
        error: error?.message || "历史记录加载失败，请稍后重试。",
      });
    } finally {
      this.loadingRequest = false;
      this.setData({ loading: false, loadingMore: false });
      if (fromPullDown) wx.stopPullDownRefresh();
    }
  },

  openRecord(event) {
    const url = event.currentTarget.dataset.url;
    if (!url) return;
    wx.navigateTo({ url });
  },

  confirmDelete(event) {
    const { id, type } = event.currentTarget.dataset;
    if (!id || !type || this.data.deletingId) return;

    wx.showModal({
      title: "确认删除",
      content: "删除后不可恢复，是否确认删除？",
      confirmColor: "#92400e",
      success: (result) => {
        if (result.confirm) {
          this.deleteRecord(id, type);
        }
      },
    });
  },

  async deleteRecord(id, recordType) {
    if (recordType !== "bazi" && recordType !== "qimen" && recordType !== "divination") {
      return;
    }

    this.setData({ deletingId: id });
    try {
      if (recordType === "divination") {
        await deleteDivination(id);
      } else {
        await deleteAnalysis(id);
      }
      wx.showToast({ title: "已删除", icon: "success" });

      const numericId = Number(id);
      const nextState = { deletingId: null };
      if (recordType === "divination") {
        nextState.divItems = this.data.divItems.filter((item) => item.id !== numericId);
      } else if (recordType === "bazi") {
        nextState.baziItems = this.data.baziItems.filter((item) => item.id !== numericId);
      } else {
        nextState.qimenItems = this.data.qimenItems.filter((item) => item.id !== numericId);
      }

      this.setData(nextState, () => {
        this.updateVisibleState();
      });
    } catch (error) {
      this.setData({ deletingId: null });
      wx.showToast({ title: mapDeleteError(error), icon: "none" });
    }
  },
});
