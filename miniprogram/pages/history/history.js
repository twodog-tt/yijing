const { getDivinationHistory } = require("../../utils/api");
const { formatDateTime } = require("../../utils/date");

const PAGE_SIZE = 20;

function truncateText(text, maxLength = 48) {
  const characters = [...String(text || "")];
  if (characters.length <= maxLength) return characters.join("");
  return `${characters.slice(0, maxLength).join("")}…`;
}

function prepareItem(item) {
  return {
    ...item,
    category_name: item.category?.name || "未分类",
    question_summary: truncateText(item.question),
    primary_name: item.primary_hexagram?.name || "未知",
    changed_name: item.changed_hexagram?.name || "未知",
    created_at_display: formatDateTime(item.created_at),
  };
}

Page({
  data: {
    items: [],
    page: 0,
    total: 0,
    hasMore: false,
    loading: true,
    loadingMore: false,
    error: "",
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

  loadMore() {
    if (!this.data.hasMore || this.data.loadingMore || this.loadingRequest) return;
    this.loadHistory(false);
  },

  async loadHistory(reset = true, fromPullDown = false) {
    if (this.loadingRequest) {
      if (fromPullDown) wx.stopPullDownRefresh();
      return;
    }

    const page = reset ? 1 : this.data.page + 1;
    this.loadingRequest = true;
    this.setData({
      loading: reset && this.data.items.length === 0,
      loadingMore: !reset,
      error: "",
    });

    try {
      const result = await getDivinationHistory({
        page,
        page_size: PAGE_SIZE,
      });
      const newItems = (Array.isArray(result?.items) ? result.items : []).map(prepareItem);
      const items = reset ? newItems : [...this.data.items, ...newItems];
      const total = Number(result?.total || 0);
      this.setData({
        items,
        page,
        total,
        hasMore: page * PAGE_SIZE < total && newItems.length > 0,
      });
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
});
