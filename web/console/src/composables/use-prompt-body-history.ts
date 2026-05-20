import {
  expandBodyForDisplay,
  mapDisplaySelectionToDraft,
  mapDraftSelectionToDisplay,
} from "@/lib/prompt-placeholders";
import { nextTick, ref } from "vue";

export type EditorHistoryEntry = {
  body: string;
  selectionStart: number;
  selectionEnd: number;
};

export type PromptBodyHistoryOptions = {
  getBody: () => string;
  setBody: (body: string) => void;
  getTextarea: () => HTMLTextAreaElement | null;
  hasDebugPreview: () => boolean;
  getSampleValues: () => Record<string, string>;
  maxDepth?: number;
  coalesceMs?: number;
};

export function usePromptBodyHistory(options: PromptBodyHistoryOptions) {
  const undoStack = ref<EditorHistoryEntry[]>([]);
  const redoStack = ref<EditorHistoryEntry[]>([]);
  const applying = ref(false);
  const composing = ref(false);

  const maxDepth = options.maxDepth ?? 100;
  const coalesceMs = options.coalesceMs ?? 400;

  let coalesceTimer: ReturnType<typeof setTimeout> | null = null;
  let coalesceActive = false;

  function clearCoalesceTimer() {
    if (coalesceTimer) {
      clearTimeout(coalesceTimer);
      coalesceTimer = null;
    }
  }

  function resetCoalesce() {
    clearCoalesceTimer();
    coalesceActive = false;
  }

  function captureSelection(): Pick<EditorHistoryEntry, "selectionStart" | "selectionEnd"> {
    const ta = options.getTextarea();
    if (!ta) return { selectionStart: 0, selectionEnd: 0 };
    const body = options.getBody();
    if (!options.hasDebugPreview()) {
      return { selectionStart: ta.selectionStart, selectionEnd: ta.selectionEnd };
    }
    return mapDisplaySelectionToDraft(
      body,
      options.getSampleValues(),
      ta.selectionStart,
      ta.selectionEnd,
    );
  }

  function capture(): EditorHistoryEntry {
    return {
      body: options.getBody(),
      ...captureSelection(),
    };
  }

  function pushUndo(entry: EditorHistoryEntry) {
    undoStack.value.push(entry);
    if (undoStack.value.length > maxDepth) {
      undoStack.value.shift();
    }
    redoStack.value = [];
  }

  function reset(body?: string) {
    undoStack.value = [];
    redoStack.value = [];
    resetCoalesce();
    if (body !== undefined) {
      // Anchor future undo to loaded content without pushing empty state.
      void body;
    }
  }

  /** Coalesce rapid typing into a single undo step. */
  function noteBeforeInput() {
    if (applying.value || composing.value) return;
    if (!coalesceActive) {
      pushUndo(capture());
      coalesceActive = true;
    }
    clearCoalesceTimer();
    coalesceTimer = setTimeout(resetCoalesce, coalesceMs);
  }

  /** One undo step per programmatic edit (Tab, Enter, variable insert, atomic delete). */
  function noteBeforeProgrammaticEdit() {
    if (applying.value || composing.value) return;
    resetCoalesce();
    pushUndo(capture());
  }

  function restore(entry: EditorHistoryEntry) {
    applying.value = true;
    options.setBody(entry.body);
    void nextTick(() => {
      const ta = options.getTextarea();
      if (ta) {
        const sel = options.hasDebugPreview()
          ? mapDraftSelectionToDisplay(
              entry.body,
              options.getSampleValues(),
              entry.selectionStart,
              entry.selectionEnd,
            )
          : {
              start: entry.selectionStart,
              end: entry.selectionEnd,
            };
        const displayLen = options.hasDebugPreview()
          ? expandBodyForDisplay(entry.body, options.getSampleValues()).length
          : entry.body.length;
        ta.selectionStart = Math.min(sel.start, displayLen);
        ta.selectionEnd = Math.min(sel.end, displayLen);
        ta.focus();
      }
      applying.value = false;
    });
  }

  function undo(): boolean {
    if (applying.value || undoStack.value.length === 0) return false;
    resetCoalesce();
    const current = capture();
    const prev = undoStack.value.pop()!;
    redoStack.value.push(current);
    restore(prev);
    return true;
  }

  function redo(): boolean {
    if (applying.value || redoStack.value.length === 0) return false;
    resetCoalesce();
    const current = capture();
    const next = redoStack.value.pop()!;
    undoStack.value.push(current);
    restore(next);
    return true;
  }

  function onCompositionStart() {
    composing.value = true;
    resetCoalesce();
    pushUndo(capture());
  }

  function onCompositionEnd() {
    composing.value = false;
    resetCoalesce();
  }

  function onBeforeInput(ev: InputEvent) {
    if (applying.value || composing.value) return;
    const t = ev.inputType;
    if (t === "historyUndo" || t === "historyRedo") {
      ev.preventDefault();
      return;
    }
    if (t.startsWith("delete") || t === "insertText" || t === "insertFromPaste" || t === "insertFromDrop") {
      noteBeforeInput();
    }
  }

  const canUndo = () => undoStack.value.length > 0;
  const canRedo = () => redoStack.value.length > 0;

  return {
    undo,
    redo,
    reset,
    noteBeforeProgrammaticEdit,
    onBeforeInput,
    onCompositionStart,
    onCompositionEnd,
    canUndo,
    canRedo,
  };
}
