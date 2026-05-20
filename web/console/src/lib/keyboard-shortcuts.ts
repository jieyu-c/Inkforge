export type ShortcutKeyToken =
  | "mod"
  | "shift"
  | "alt"
  | "enter"
  | "tab"
  | "esc"
  | "up"
  | "down"
  | "backspace"
  | "delete"
  | string;

export type ShortcutItem = {
  labelKey: string;
  /** Each inner array is one key chord (e.g. redo: Mod+Shift+Z or Mod+Y). */
  chords: ShortcutKeyToken[][];
};

export type ShortcutGroup = {
  titleKey: string;
  items: ShortcutItem[];
};

/** Console keyboard shortcuts shown in the FAB dialog. */
export const CONSOLE_SHORTCUT_CATALOG: ShortcutGroup[] = [
  {
    titleKey: "shortcuts.groups.editor",
    items: [
      { labelKey: "shortcuts.save", chords: [["mod", "S"]] },
      { labelKey: "shortcuts.undo", chords: [["mod", "Z"]] },
      { labelKey: "shortcuts.redo", chords: [["mod", "shift", "Z"], ["mod", "Y"]] },
      { labelKey: "shortcuts.addVariable", chords: [["mod", "shift", "V"]] },
      {
        labelKey: "shortcuts.autocomplete",
        chords: [["{{"], ["up"], ["down"], ["enter"], ["tab"], ["esc"]],
      },
    ],
  },
];

function isApplePlatform(): boolean {
  if (typeof navigator === "undefined") return false;
  return /Mac|iPhone|iPad|iPod/i.test(navigator.platform || navigator.userAgent);
}

const TOKEN_LABELS: Record<string, { mac: string; win: string }> = {
  mod: { mac: "⌘", win: "Ctrl" },
  shift: { mac: "⇧", win: "Shift" },
  alt: { mac: "⌥", win: "Alt" },
  enter: { mac: "↵", win: "Enter" },
  tab: { mac: "⇥", win: "Tab" },
  esc: { mac: "Esc", win: "Esc" },
  up: { mac: "↑", win: "↑" },
  down: { mac: "↓", win: "↓" },
  backspace: { mac: "⌫", win: "Backspace" },
  delete: { mac: "⌦", win: "Delete" },
};

/** Format one chord as human-readable key labels (platform-aware). */
export function formatShortcutChord(tokens: ShortcutKeyToken[]): string[] {
  const apple = isApplePlatform();
  return tokens.map((token) => {
    const known = TOKEN_LABELS[token.toLowerCase()];
    if (known) return apple ? known.mac : known.win;
    if (token.length === 1) return token.toUpperCase();
    return token;
  });
}
