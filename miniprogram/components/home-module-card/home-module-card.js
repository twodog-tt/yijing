Component({
  properties: {
    title: {
      type: String,
      value: "",
    },
    subtitle: {
      type: String,
      value: "",
    },
    tags: {
      type: Array,
      value: [],
    },
    buttonText: {
      type: String,
      value: "进入",
    },
    url: {
      type: String,
      value: "",
    },
    accent: {
      type: String,
      value: "default",
    },
  },
});
