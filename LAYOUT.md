# Frontend Layout Design

## Main Application Layout

```
┌────────────────────────────────────────┐
│ 🟦 Work Project        ⌘K  ← →     [≡] │  <- Minimal header
├────────────────────────────────────────┤
│                                        │
│  [ Add Text ] [ Bullet ] [ Image ]     │
│                                        │
│  +----------------------------------+  │
│  | Item 1 (text/bullet/image)       |  │
│  | [edit] [delete]                  |  │
│  +----------------------------------+  │
│                                        │
│  +----------------------------------+  │
│  | Item 2                           |  │
│  +----------------------------------+  │
│                                        │
└────────────────────────────────────────┘
```

**Components:**
- **Header (Sticky):** Shows active context name with colored dot, keyboard shortcut hints, and hamburger menu button
- **Action Bar:** Three buttons for adding different item types
- **Content Area:** Scrollable list of items with inline edit/delete controls

---

## Quick Switcher (Ctrl+K)

**Modal overlay activated by Ctrl+K:**

```
┌─────────────────────────────────┐
│ > Switch context...             │
├─────────────────────────────────┤
│ 🟦 Work Project                 │
│ 🟩 Personal Tasks               │
│ 🟨 Research Notes               │
│ 🟥 Meeting Notes                │
└─────────────────────────────────┘
```

**Features:**
- Fuzzy search filter as you type
- Keyboard navigation (arrow keys + Enter)
- Colored dot indicators for each context
- Close with Escape key
- "Create new context" option at bottom (when no results match)

---

## Sidebar Menu ([≡])

**Slide-in panel from left side:**

```
┌────┬────────────────────────────┐
│ 🟦 │ Main Content              │
│ 🟩 │                           │
│ 🟨 │ (Active context shown)    │
│ 🟥 │                           │
│ 🟪 │                           │
│ +  │                           │
└────┴────────────────────────────┘
```

**Features:**
- Narrow vertical bar with colored context dots
- Click any dot to switch to that context
- Hover shows context name as tooltip
- "+" button at bottom to create new context
- Click outside or press Escape to close
- Smooth slide-in/out animation

---

## Context Switching Methods

Multiple ways to switch contexts for maximum flexibility:

1. **Command Palette:** `Ctrl+K` opens quick switcher with fuzzy search
2. **Arrow Navigation:** `Ctrl+Left/Right` cycles through contexts
3. **Number Shortcuts:** `Ctrl+1-9` jumps to specific context (1st, 2nd, 3rd, etc.)
4. **Sidebar Menu:** Click `[≡]` button to show all contexts, click colored dot to switch
5. **Visual Indicator:** Current context shown with color dot + name in header

---

## Item Types & Display

### Text Item
```
+----------------------------------+
| 📝 This is a text block          |
|                   [edit] [delete] |
+----------------------------------+
```

### Bullet Point Item
```
+----------------------------------+
| • Bullet point text here         |
|                   [edit] [delete] |
+----------------------------------+
```

### Image Item
```
+----------------------------------+
| [Image Preview]                  |
|                        [delete]  |
+----------------------------------+
```

**Edit Behavior:**
- Click [edit] to enter inline editing mode
- Text converts to textarea/input
- Save on Enter or click outside
- Cancel with Escape key

---

## Color Scheme

### Context Colors
Each context gets a unique color for visual distinction:
- 🟦 Blue
- 🟩 Green
- 🟨 Yellow
- 🟥 Red
- 🟪 Purple
- 🟧 Orange
- 🩷 Pink
- 🩵 Cyan

Colors are assigned automatically when creating new contexts, cycling through the available options.

---

## Responsive Behavior

**Desktop (Primary):**
- Full sidebar and command palette
- All keyboard shortcuts active
- Hover states for edit/delete buttons

**Mobile/Tablet (Optional):**
- Hamburger menu becomes primary navigation
- Touch-friendly button sizes
- Edit/delete buttons always visible (not on hover)
- Command palette still accessible but may be less prominent

---

## Keyboard Shortcuts Summary

| Shortcut | Action |
|----------|--------|
| `Ctrl+K` (or `Cmd+K`) | Open command palette |
| `Ctrl+Left Arrow` | Previous context |
| `Ctrl+Right Arrow` | Next context |
| `Ctrl+1` through `Ctrl+9` | Jump to context by index |
| `Escape` | Close command palette or sidebar |
| `Enter` | Select context in command palette |
| `Arrow Up/Down` | Navigate command palette |

---

## UI/UX Notes

- **Minimal Design:** Clean, spacious interface with lots of white space
- **Fast Switching:** All switching methods should feel instant (< 100ms)
- **Visual Feedback:** Subtle animations for context changes and item additions
- **Auto-save:** Debounced saves with brief "Saving..." indicator
- **Drag-and-drop:** Support for image uploads (drop anywhere in content area)
- **Accessibility:** Proper focus management, ARIA labels, keyboard navigation
