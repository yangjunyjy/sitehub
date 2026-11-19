// The API owns mode metadata. Loading states must not invent public modes.
export const getDocumentMode = (modes, id, defaultID, versionID) =>
  modes.find((mode) => mode.id === id) ||
  (!id &&
    modes.find(
      (mode) => mode.available && mode.versionIds?.includes(versionID),
    )) ||
  modes.find((mode) => mode.id === defaultID) ||
  modes[0] || { id: "", label: "", icon: "playbook", available: false };
