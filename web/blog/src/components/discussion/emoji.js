const artwork = import.meta.glob("../../assets/emoji/*.png", {
  eager: true,
  import: "default",
});
export const emoji = [
  ["😄", "开心"],
  ["👍", "赞同"],
  ["👏", "鼓掌"],
  ["😍", "喜欢"],
  ["🎉", "庆祝"],
  ["🤔", "思考"],
  ["😂", "笑哭"],
  ["🙌", "太棒了"],
  ["🚀", "起飞"],
  ["👀", "围观"],
  ["🌟", "闪亮"],
  ["💚", "爱心"],
].map(([symbol, name]) => ({
  symbol,
  name,
  src: artwork[
    "../../assets/emoji/" + symbol.codePointAt(0).toString(16) + ".png"
  ],
}));
const pattern = new RegExp(
  "(" + emoji.map((item) => item.symbol).join("|") + ")",
  "gu",
);
export function commentParts(value) {
  return String(value || "")
    .split(pattern)
    .filter(Boolean)
    .map((text) => emoji.find((item) => item.symbol === text) || { text });
}
