Component({
  properties: {
    visible: {
      type: Boolean,
      value: false,
    },
    message: {
      type: String,
      value: "加载失败，请稍后重试。",
    },
    showRetry: {
      type: Boolean,
      value: true,
    },
  },

  methods: {
    handleRetry() {
      this.triggerEvent("retry");
    },
  },
});
