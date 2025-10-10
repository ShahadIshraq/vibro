// State management
let currentContextId = null;
let contexts = [];
let isEditMode = false;
let editingContextId = null;
let currentNotes = [];

// Color gradient mapping
const colorGradients = {
    purple: 'var(--gradient-purple)',
    blue: 'var(--gradient-blue)',
    green: 'var(--gradient-green)',
    orange: 'var(--gradient-orange)',
    red: 'var(--gradient-red)',
    pink: 'var(--gradient-pink)',
    teal: 'var(--gradient-teal)'
};

// Initialize app on page load
document.addEventListener('DOMContentLoaded', async () => {
    await loadContexts();
    setupEventListeners();
});

// Load contexts from API
async function loadContexts() {
    const statusDiv = document.getElementById('status');

    try {
        const response = await fetch('/api/contexts');

        if (response.ok) {
            contexts = await response.json();
            // Don't show success message on initial load
            statusDiv.textContent = '';
            statusDiv.className = '';

            renderContextCards();

            // Select first context if available
            if (contexts.length > 0) {
                selectContext(contexts[0].id);
            }
        } else {
            throw new Error(`HTTP ${response.status}`);
        }
    } catch (error) {
        statusDiv.textContent = `✗ Backend connection failed: ${error.message}`;
        statusDiv.className = 'error';
    }
}

// Render context cards at the bottom
function renderContextCards() {
    const container = document.getElementById('cards-container');
    container.innerHTML = '';

    contexts.forEach(context => {
        const card = document.createElement('div');
        card.className = 'context-card';
        card.dataset.contextId = context.id;

        // Apply color class
        const color = context.color || 'purple';
        if (color !== 'purple') {
            card.classList.add(`color-${color}`);
        }

        if (context.id === currentContextId) {
            card.classList.add('active');
        }

        card.innerHTML = `
            <div class="context-card-name">${context.name}</div>
            <div class="context-card-info">${(context.notes || []).length} notes</div>
        `;

        card.addEventListener('click', () => selectContext(context.id));

        // Right-click to edit
        card.addEventListener('contextmenu', (e) => {
            e.preventDefault();
            openEditContextModal(context.id);
        });

        container.appendChild(card);
    });
}

// Select and display a context
function selectContext(contextId) {
    currentContextId = contextId;
    const context = contexts.find(c => c.id === contextId);

    if (!context) return;

    // Update context view
    const contextView = document.getElementById('context-view');
    const contextName = document.getElementById('context-name');
    contextName.textContent = context.name;
    contextName.contentEditable = 'true';
    contextName.dataset.contextId = contextId;

    // Apply gradient background based on context color
    const color = context.color || 'purple';
    contextView.style.background = colorGradients[color] || colorGradients.purple;

    // Show notes section
    const notesSection = document.getElementById('notes-section');
    notesSection.style.display = 'block';

    // Render inline notes
    renderNotesList();

    // Update active card
    document.querySelectorAll('.context-card').forEach(card => {
        card.classList.toggle('active', card.dataset.contextId === contextId);
    });
}

// Setup event listeners
function setupEventListeners() {
    const addBtn = document.getElementById('add-context-btn');
    addBtn.addEventListener('click', openCreateContextModal);

    const contextForm = document.getElementById('context-form');
    contextForm.addEventListener('submit', handleContextFormSubmit);

    // Close modal when clicking outside
    const modal = document.getElementById('context-modal');
    modal.addEventListener('click', (e) => {
        if (e.target === modal) {
            closeModal();
        }
    });

    // Keyboard shortcuts
    document.addEventListener('keydown', (e) => {
        // Escape to close modal
        if (e.key === 'Escape') {
            closeModal();
        }
    });

    // Context name editing
    const contextName = document.getElementById('context-name');
    contextName.addEventListener('blur', saveContextName);
    contextName.addEventListener('keydown', (e) => {
        if (e.key === 'Enter') {
            e.preventDefault();
            contextName.blur();
        }
    });
}

// Open create context modal
function openCreateContextModal() {
    isEditMode = false;
    editingContextId = null;
    currentNotes = [];

    const modal = document.getElementById('context-modal');
    const modalTitle = document.getElementById('modal-title');
    const form = document.getElementById('context-form');

    modalTitle.textContent = 'Create New Context';
    form.reset();
    renderModalNotesList();

    modal.classList.add('active');
}

// Open edit context modal
function openEditContextModal(contextId) {
    isEditMode = true;
    editingContextId = contextId;

    const context = contexts.find(c => c.id === contextId);
    if (!context) return;

    currentNotes = [...(context.notes || [])];

    const modal = document.getElementById('context-modal');
    const modalTitle = document.getElementById('modal-title');
    const nameInput = document.getElementById('context-name-input');
    const colorInput = document.getElementById('context-color-input');
    const descInput = document.getElementById('context-description-input');

    modalTitle.textContent = 'Edit Context';
    nameInput.value = context.name;
    colorInput.value = context.color || 'purple';
    descInput.value = context.description || '';

    renderModalNotesList();

    modal.classList.add('active');
}

// Close modal
function closeModal() {
    const modal = document.getElementById('context-modal');
    modal.classList.remove('active');
    isEditMode = false;
    editingContextId = null;
}

// Handle context form submission
async function handleContextFormSubmit(e) {
    e.preventDefault();

    const nameInput = document.getElementById('context-name-input');
    const colorInput = document.getElementById('context-color-input');
    const descInput = document.getElementById('context-description-input');

    const contextData = {
        name: nameInput.value.trim(),
        color: colorInput.value,
        description: descInput.value.trim(),
        notes: currentNotes
    };

    try {
        let response;

        if (isEditMode && editingContextId) {
            // Update existing context
            response = await fetch(`/api/contexts/${editingContextId}`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(contextData)
            });
        } else {
            // Create new context
            response = await fetch('/api/contexts', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(contextData)
            });
        }

        if (response.ok) {
            const result = await response.json();

            if (isEditMode) {
                // Update context in array
                const index = contexts.findIndex(c => c.id === editingContextId);
                if (index !== -1) {
                    contexts[index] = result;
                }
            } else {
                // Add new context to array
                contexts.push(result);
            }

            renderContextCards();
            selectContext(result.id);
            closeModal();

            showStatus(`Context ${isEditMode ? 'updated' : 'created'} successfully!`, 'success');
        } else {
            const error = await response.json();
            throw new Error(error.message || 'Failed to save context');
        }
    } catch (error) {
        showStatus(`Error: ${error.message}`, 'error');
    }
}

// Show status message
function showStatus(message, type) {
    const statusDiv = document.getElementById('status');
    statusDiv.textContent = message;
    statusDiv.className = type;

    // Auto-hide after 3 seconds
    setTimeout(() => {
        statusDiv.textContent = '';
        statusDiv.className = '';
    }, 3000);
}

// Utility function for API calls
async function apiCall(endpoint, options = {}) {
    const response = await fetch(`/api${endpoint}`, {
        headers: {
            'Content-Type': 'application/json',
            ...options.headers,
        },
        ...options,
    });

    if (!response.ok) {
        const error = await response.json();
        throw new Error(error.message || 'API request failed');
    }

    return response.json();
}

// Export for use in frontend development
window.api = {
    call: apiCall,
};

// Render modal notes list
function renderModalNotesList() {
    const notesList = document.getElementById('modal-notes-list');
    notesList.innerHTML = '';

    currentNotes.forEach((note, index) => {
        const item = document.createElement('div');
        item.className = 'note-item-wrapper';

        const indentLevel = note.indent || 0;

        // Create tree structure
        let treeHTML = '';
        for (let i = 0; i < indentLevel; i++) {
            treeHTML += '<span class="tree-indent"></span>';
        }
        if (indentLevel > 0) {
            treeHTML += '<span class="tree-branch">└─</span>';
        }

        item.innerHTML = `
            <div class="note-tree-modal">${treeHTML}</div>
            <div class="note-item">
                <button type="button" class="btn-indent" onclick="indentNote(${index}, -1)" ${note.indent === 0 ? 'disabled' : ''}>&larr;</button>
                <button type="button" class="btn-indent" onclick="indentNote(${index}, 1)">&rarr;</button>
                <span class="note-text">${note.text}</span>
                <button type="button" class="btn-remove" onclick="removeModalNote(${index})">&times;</button>
            </div>
        `;
        notesList.appendChild(item);
    });
}

// Add note in modal
function addModalNote() {
    const input = document.getElementById('modal-note-input');
    const text = input.value.trim();

    if (!text) {
        showStatus('Note text is required', 'error');
        return;
    }

    currentNotes.push({ text, indent: 0 });
    renderModalNotesList();

    input.value = '';
}

// Remove note from modal
function removeModalNote(index) {
    currentNotes.splice(index, 1);
    renderModalNotesList();
}

// Indent/outdent note
function indentNote(index, direction) {
    const note = currentNotes[index];
    const newIndent = Math.max(0, (note.indent || 0) + direction);
    currentNotes[index].indent = newIndent;
    renderModalNotesList();
}

// Save context name when editing inline
async function saveContextName() {
    const contextName = document.getElementById('context-name');
    const newName = contextName.textContent.trim();
    const contextId = contextName.dataset.contextId;

    if (!contextId || !newName) return;

    const context = contexts.find(c => c.id === contextId);
    if (!context || context.name === newName) return;

    try {
        const response = await fetch(`/api/contexts/${contextId}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                name: newName,
                color: context.color,
                description: context.description,
                notes: context.notes
            })
        });

        if (response.ok) {
            const result = await response.json();
            const index = contexts.findIndex(c => c.id === contextId);
            if (index !== -1) {
                contexts[index] = result;
            }
            renderContextCards();
            showStatus('Context name updated!', 'success');
        } else {
            throw new Error('Failed to update context name');
        }
    } catch (error) {
        showStatus(`Error: ${error.message}`, 'error');
        contextName.textContent = context.name; // Revert on error
    }
}

// Render inline notes list
function renderNotesList() {
    const context = contexts.find(c => c.id === currentContextId);
    if (!context) return;

    const notesList = document.getElementById('notes-list');
    notesList.innerHTML = '';

    (context.notes || []).forEach((note, index) => {
        const item = document.createElement('div');
        item.className = 'note-item-wrapper';

        const indentLevel = note.indent || 0;

        // Create tree structure
        let treeHTML = '';
        for (let i = 0; i < indentLevel; i++) {
            treeHTML += '<span class="tree-indent"></span>';
        }
        if (indentLevel > 0) {
            treeHTML += '<span class="tree-branch">└─</span>';
        }

        item.innerHTML = `
            <div class="note-tree">${treeHTML}</div>
            <div class="note-item-inline">
                <button type="button" class="btn-indent-inline" onclick="indentNoteInline(${index}, -1)" ${note.indent === 0 ? 'disabled' : ''}>&larr;</button>
                <button type="button" class="btn-indent-inline" onclick="indentNoteInline(${index}, 1)">&rarr;</button>
                <span class="note-text-inline">${note.text}</span>
                <button type="button" class="btn-remove-inline" onclick="removeNote(${index})">&times;</button>
            </div>
        `;
        notesList.appendChild(item);
    });
}

// Add note inline
async function addNote() {
    const input = document.getElementById('note-input');
    const text = input.value.trim();

    if (!text) {
        showStatus('Note text is required', 'error');
        return;
    }

    const context = contexts.find(c => c.id === currentContextId);
    if (!context) return;

    context.notes = context.notes || [];
    context.notes.push({ text, indent: 0 });

    await updateContextOnServer(context);

    input.value = '';
    renderNotesList();
    renderContextCards();
}

// Remove note inline
async function removeNote(index) {
    const context = contexts.find(c => c.id === currentContextId);
    if (!context) return;

    context.notes.splice(index, 1);
    await updateContextOnServer(context);
    renderNotesList();
    renderContextCards();
}

// Indent/outdent note inline
async function indentNoteInline(index, direction) {
    const context = contexts.find(c => c.id === currentContextId);
    if (!context) return;

    const note = context.notes[index];
    const newIndent = Math.max(0, (note.indent || 0) + direction);
    context.notes[index].indent = newIndent;

    await updateContextOnServer(context);
    renderNotesList();
}

// Update context on server
async function updateContextOnServer(context) {
    try {
        const response = await fetch(`/api/contexts/${context.id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(context)
        });

        if (response.ok) {
            const result = await response.json();
            const index = contexts.findIndex(c => c.id === context.id);
            if (index !== -1) {
                contexts[index] = result;
            }
            showStatus('Updated successfully!', 'success');
        } else {
            throw new Error('Failed to update context');
        }
    } catch (error) {
        showStatus(`Error: ${error.message}`, 'error');
    }
}

// Make functions globally accessible for inline onclick
window.closeModal = closeModal;
window.addModalNote = addModalNote;
window.removeModalNote = removeModalNote;
window.indentNote = indentNote;
window.addNote = addNote;
window.removeNote = removeNote;
window.indentNoteInline = indentNoteInline;

console.log('Vibro frontend initialized');
