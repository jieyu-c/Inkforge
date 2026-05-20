/**
 * Placeholder segments for prompt body preview.
 * Regex aligned with services/console/internal/pkg/promptdraft/validate.go
 */
export type PlaceholderSegment = { kind: "text" | "placeholder"; text: string };

const placeholderRe = /\{\{\s*([a-zA-Z0-9_]+)\s*\}\}/g;

export const VAR_NAME_RE = /^[a-zA-Z0-9_]+$/;

export function isValidVarName(name: string): boolean {
  const trimmed = name.trim();
  return trimmed.length > 0 && VAR_NAME_RE.test(trimmed);
}

/** Normalize free text into a valid placeholder name. */
export function sanitizeVarName(raw: string): string {
  const cleaned = raw.trim().replace(/[^a-zA-Z0-9_]/g, "_");
  if (!cleaned) return "";
  return /^[0-9]/.test(cleaned) ? `_${cleaned}` : cleaned;
}

export function formatPlaceholder(name: string): string {
  return `{{${name}}}`;
}

export type PlaceholderTrigger = { start: number; query: string };

/** True when the cursor is immediately after an incomplete `{{name` segment. */
export function getPlaceholderTrigger(body: string, cursor: number): PlaceholderTrigger | null {
  const before = body.slice(0, cursor);
  const match = before.match(/\{\{([a-zA-Z0-9_]*)$/);
  if (!match) return null;
  return { start: cursor - match[0].length, query: match[1] ?? "" };
}

export type PlaceholderSpan = { start: number; end: number };

const debugPlaceholderRe = /\{\{\s*([a-zA-Z0-9_]+)\s*(?::([^}]*))?\s*\}\}/g;

function placeholderPattern(withDebugValues: boolean): RegExp {
  return withDebugValues
    ? new RegExp(debugPlaceholderRe.source, "g")
    : new RegExp(placeholderRe.source, "g");
}

/** Ranges of complete `{{name}}` (or `{{name:value}}` in debug preview) tokens. */
export function findPlaceholderSpans(body: string, withDebugValues = false): PlaceholderSpan[] {
  const spans: PlaceholderSpan[] = [];
  const re = placeholderPattern(withDebugValues);
  let m: RegExpExecArray | null;
  while ((m = re.exec(body)) !== null) {
    spans.push({ start: m.index, end: m.index + m[0].length });
  }
  return spans;
}

function spanContains(sp: PlaceholderSpan, pos: number): boolean {
  return pos >= sp.start && pos < sp.end;
}

function spansOverlap(aStart: number, aEnd: number, bStart: number, bEnd: number): boolean {
  return aStart < bEnd && bStart < aEnd;
}

/**
 * When deleting a placeholder, return the full span to remove.
 * Handles cursor-inside, Backspace/Delete at edges, and partial selections.
 */
export function getAtomicPlaceholderDeleteRange(
  body: string,
  selStart: number,
  selEnd: number,
  key: "Backspace" | "Delete",
  withDebugValues = false,
): PlaceholderSpan | null {
  const spans = findPlaceholderSpans(body, withDebugValues);

  if (selStart !== selEnd) {
    let start = selStart;
    let end = selEnd;
    let expanded = false;
    for (const sp of spans) {
      if (!spansOverlap(selStart, selEnd, sp.start, sp.end)) continue;
      const fullySelected = selStart <= sp.start && selEnd >= sp.end;
      if (fullySelected) continue;
      start = Math.min(start, sp.start);
      end = Math.max(end, sp.end);
      expanded = true;
    }
    return expanded ? { start, end } : null;
  }

  const pos = selStart;
  if (key === "Backspace") {
    if (pos === 0) return null;
    const hit = spans.find((sp) => spanContains(sp, pos - 1));
    return hit ?? null;
  }
  if (key === "Delete") {
    if (pos >= body.length) return null;
    const hit = spans.find((sp) => spanContains(sp, pos));
    return hit ?? null;
  }
  return null;
}

/** Unique placeholder names in first-seen order (aligned with API validate). */
export function extractPlaceholderNames(body: string): string[] {
  const names: string[] = [];
  const seen = new Set<string>();
  const r = new RegExp(placeholderRe.source, "g");
  let m: RegExpExecArray | null;
  while ((m = r.exec(body)) !== null) {
    const n = m[1];
    if (!seen.has(n)) {
      seen.add(n);
      names.push(n);
    }
  }
  return names;
}

export function parsePlaceholderSegments(src: string): PlaceholderSegment[] {
  const out: PlaceholderSegment[] = [];
  let last = 0;
  let m: RegExpExecArray | null;
  const r = new RegExp(placeholderRe.source, "g");
  while ((m = r.exec(src)) !== null) {
    if (m.index > last) {
      out.push({ kind: "text", text: src.slice(last, m.index) });
    }
    out.push({ kind: "placeholder", text: m[0] });
    last = m.index + m[0].length;
  }
  if (last < src.length) {
    out.push({ kind: "text", text: src.slice(last) });
  }
  return out;
}

function escapeHtml(raw: string): string {
  return raw
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

const displayPlaceholderRe = /\{\{\s*([a-zA-Z0-9_]+)\s*(?::([^}]*))?\s*\}\}/g;

export function hasDebugSampleValues(sampleValues: Record<string, string>): boolean {
  return Object.values(sampleValues).some((v) => v.trim().length > 0);
}

/** Expand placeholders to {{name:value}} in overlay when debug value exists. */
export function expandBodyForDisplay(body: string, sampleValues: Record<string, string>): string {
  return body.replace(/\{\{\s*([a-zA-Z0-9_]+)\s*\}\}/g, (match, name: string) => {
    const val = sampleValues[name]?.trim();
    return val ? `{{${name}:${val}}}` : match;
  });
}

type BodyDisplaySegment = {
  kind: "text" | "placeholder";
  draftStart: number;
  draftLen: number;
  displayStart: number;
  displayLen: number;
  draftText: string;
  displayText: string;
};

function eachBodyDisplaySegment(
  body: string,
  sampleValues: Record<string, string>,
  visit: (seg: BodyDisplaySegment) => boolean | void,
): void {
  let last = 0;
  const re = new RegExp(placeholderRe.source, "g");
  let m: RegExpExecArray | null;
  while ((m = re.exec(body)) !== null) {
    if (m.index > last) {
      const text = body.slice(last, m.index);
      const seg: BodyDisplaySegment = {
        kind: "text",
        draftStart: last,
        draftLen: text.length,
        displayStart: 0,
        displayLen: text.length,
        draftText: text,
        displayText: text,
      };
      if (visit(seg) === true) return;
      last = m.index;
    }
    const canonical = m[0];
    const name = m[1] ?? "";
    const val = sampleValues[name]?.trim();
    const display = val ? `{{${name}:${val}}}` : canonical;
    const seg: BodyDisplaySegment = {
      kind: "placeholder",
      draftStart: m.index,
      draftLen: canonical.length,
      displayStart: 0,
      displayLen: display.length,
      draftText: canonical,
      displayText: display,
    };
    if (visit(seg) === true) return;
    last = m.index + canonical.length;
  }
  if (last < body.length) {
    const text = body.slice(last);
    visit({
      kind: "text",
      draftStart: last,
      draftLen: text.length,
      displayStart: 0,
      displayLen: text.length,
      draftText: text,
      displayText: text,
    });
  }
}

function offsetWithinExpandedPlaceholder(draftText: string, displayText: string, offsetInDraft: number): number {
  const nameMatch = draftText.match(/\{\{\s*([a-zA-Z0-9_]+)\s*\}\}/);
  if (!nameMatch || nameMatch.index === undefined) return offsetInDraft;
  const name = nameMatch[1];
  const nameStart = nameMatch.index + draftText.indexOf(name, 2);
  const nameEnd = nameStart + name.length;
  if (offsetInDraft <= nameStart) return offsetInDraft;
  if (offsetInDraft <= nameEnd) return offsetInDraft;
  return displayText.length - (draftText.length - offsetInDraft);
}

function buildDisplaySegmentOffsets(
  body: string,
  sampleValues: Record<string, string>,
): BodyDisplaySegment[] {
  const segments: BodyDisplaySegment[] = [];
  let displayIdx = 0;
  eachBodyDisplaySegment(body, sampleValues, (raw) => {
    const seg = { ...raw, displayStart: displayIdx };
    displayIdx += seg.displayLen;
    segments.push(seg);
    return false;
  });
  return segments;
}

/** Map a draft-body offset to the expanded debug-preview textarea offset. */
export function mapDraftOffsetToDisplay(
  body: string,
  sampleValues: Record<string, string>,
  draftOffset: number,
): number {
  if (!hasDebugSampleValues(sampleValues)) return draftOffset;
  const segments = buildDisplaySegmentOffsets(body, sampleValues);
  for (const seg of segments) {
    const draftEnd = seg.draftStart + seg.draftLen;
    if (draftOffset < seg.draftStart) break;
    if (draftOffset <= draftEnd) {
      if (seg.kind === "text") {
        return seg.displayStart + (draftOffset - seg.draftStart);
      }
      return seg.displayStart + offsetWithinExpandedPlaceholder(
        seg.draftText,
        seg.displayText,
        draftOffset - seg.draftStart,
      );
    }
  }
  const last = segments.at(-1);
  return last ? last.displayStart + last.displayLen : draftOffset;
}

/** Map an expanded debug-preview textarea offset back to draft-body space. */
export function mapDisplayOffsetToDraft(
  body: string,
  sampleValues: Record<string, string>,
  displayOffset: number,
): number {
  if (!hasDebugSampleValues(sampleValues)) return displayOffset;
  const segments = buildDisplaySegmentOffsets(body, sampleValues);
  for (const seg of segments) {
    const displayEnd = seg.displayStart + seg.displayLen;
    if (displayOffset < seg.displayStart) break;
    if (displayOffset <= displayEnd) {
      if (seg.kind === "text") {
        return seg.draftStart + (displayOffset - seg.displayStart);
      }
      const offsetInDisplay = displayOffset - seg.displayStart;
      const nameMatch = seg.draftText.match(/\{\{\s*([a-zA-Z0-9_]+)\s*\}\}/);
      if (!nameMatch || nameMatch.index === undefined) {
        return seg.draftStart + offsetInDisplay;
      }
      const name = nameMatch[1];
      const draftNameStart = nameMatch.index + seg.draftText.indexOf(name, 2);
      const displayNameStart = seg.displayText.indexOf(name, 2);
      const displayNameEnd = displayNameStart + name.length;
      if (offsetInDisplay <= displayNameStart) return seg.draftStart + offsetInDisplay;
      if (offsetInDisplay <= displayNameEnd) {
        return seg.draftStart + (draftNameStart + (offsetInDisplay - displayNameStart));
      }
      return seg.draftStart + seg.draftLen - (seg.displayLen - offsetInDisplay);
    }
  }
  return body.length;
}

export function mapDraftSelectionToDisplay(
  body: string,
  sampleValues: Record<string, string>,
  start: number,
  end: number,
): { start: number; end: number } {
  return {
    start: mapDraftOffsetToDisplay(body, sampleValues, start),
    end: mapDraftOffsetToDisplay(body, sampleValues, end),
  };
}

export function mapDisplaySelectionToDraft(
  body: string,
  sampleValues: Record<string, string>,
  start: number,
  end: number,
): { selectionStart: number; selectionEnd: number } {
  return {
    selectionStart: mapDisplayOffsetToDraft(body, sampleValues, start),
    selectionEnd: mapDisplayOffsetToDraft(body, sampleValues, end),
  };
}

function renderPlaceholderTag(name: string, debugVal?: string): string {
  let html = `<span class="var-ph-tag${debugVal ? " has-val" : ""}">`;
  html += `<span class="p-open">{{</span>`;
  html += `<span class="p-name">${escapeHtml(name)}</span>`;
  if (debugVal) {
    html += `<span class="p-val">:${escapeHtml(debugVal)}</span>`;
  }
  html += `<span class="p-close">}}</span></span>`;
  return html;
}

/**
 * Highlight overlay. With debug values, renders expanded {{name:value}} in-flow
 * so following text is pushed; textarea keeps canonical {{name}} for save.
 */
export function renderBodyHighlightHtml(body: string, sampleValues: Record<string, string>): string {
  const debugPreview = hasDebugSampleValues(sampleValues);
  const src = debugPreview ? expandBodyForDisplay(body, sampleValues) : body;
  const re = debugPreview
    ? new RegExp(displayPlaceholderRe.source, "g")
    : new RegExp(placeholderRe.source, "g");

  let html = "";
  let last = 0;
  let m: RegExpExecArray | null;
  while ((m = re.exec(src)) !== null) {
    if (m.index > last) {
      html += `<span class="hl-plain">${escapeHtml(src.slice(last, m.index))}</span>`;
    }
    const name = m[1] ?? "";
    if (debugPreview) {
      const inlineVal = m[2]?.trim();
      html += renderPlaceholderTag(name, inlineVal || undefined);
    } else {
      html += `<mark class="var-ph">${escapeHtml(m[0])}</mark>`;
    }
    last = m.index + m[0].length;
  }
  if (last < src.length) {
    html += `<span class="hl-plain">${escapeHtml(src.slice(last))}</span>`;
  }
  return html;
}

/** Remove debug `:value` suffixes from placeholders before persisting draft body. */
export function normalizeDraftBodyForSave(body: string): string {
  return body.replace(/\{\{\s*([a-zA-Z0-9_]+)\s*:([^}]*)\}\}/g, "{{$1}}");
}

/** Canonical JSON string for dirty-checking schema drafts. */
export function normalizeSchemaJson(raw: string): string {
  const trimmed = raw.trim() || "[]";
  try {
    return JSON.stringify(JSON.parse(trimmed));
  } catch {
    return trimmed;
  }
}

/** Unique variable names from schema JSON array entries with "name". */
export function schemaVariableNames(schemaJson: string): string[] {
  const raw = schemaJson.trim();
  if (!raw) return [];
  try {
    const v = JSON.parse(raw) as unknown;
    if (!Array.isArray(v)) return [];
    const names: string[] = [];
    for (const item of v) {
      if (item && typeof item === "object" && "name" in item) {
        const n = String((item as { name: unknown }).name).trim();
        if (n) names.push(n);
      }
    }
    return [...new Set(names)];
  } catch {
    return [];
  }
}
