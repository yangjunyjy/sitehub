import { ElMessageBox } from "element-plus";
export const confirmDiscard = async (changed) => {
  if (!changed) return true;
  try {
    await ElMessageBox.confirm("有未保存的修改，关闭后将丢失。", "关闭编辑", {
      confirmButtonText: "放弃修改",
      cancelButtonText: "继续编辑",
      type: "warning",
    });
    return true;
  } catch {
    return false;
  }
};
