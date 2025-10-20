// State management
let currentContextId = null;
let contexts = [];
let isEditMode = false;
let editingContextId = null;
let lastSyncTime = null;
let syncStartTime = null;
let pendingSyncComplete = null;
let autoRefreshInterval = null;
let serverContexts = []; // Track server state for conflict detection

// Color gradient mapping
const colorGradients = {
    purple: 'var(--gradient-purple)',
    blue: 'var(--gradient-blue)',
    green: 'var(--gradient-green)',
    orange: 'var(--gradient-orange)',
    red: 'var(--gradient-red)',
    pink: 'var(--gradient-pink)',
    teal: 'var(--gradient-teal)',
    indigo: 'var(--gradient-indigo)',
    cyan: 'var(--gradient-cyan)',
    emerald: 'var(--gradient-emerald)',
    amber: 'var(--gradient-amber)',
    rose: 'var(--gradient-rose)',
    violet: 'var(--gradient-violet)',
    lime: 'var(--gradient-lime)',
    sky: 'var(--gradient-sky)',
    fuchsia: 'var(--gradient-fuchsia)',
    slate: 'var(--gradient-slate)'
};

// Initialize app on page load
document.addEventListener('DOMContentLoaded', async () => {
    await loadContexts();
    setupEventListeners();
    startAutoRefresh();
});

// Load contexts from API
async function loadContexts(silent = false) {
    try {
        if (!silent) {
            updateSyncStatus('syncing');
        }

        const response = await fetch('/api/contexts');

        if (response.ok) {
            const newContexts = await response.json();

            // Store server state for conflict detection
            serverContexts = JSON.parse(JSON.stringify(newContexts));

            // Update contexts
            contexts = newContexts;
            renderContextCards();

            // Select first context if available and none selected
            if (contexts.length > 0 && !currentContextId) {
                selectContext(contexts[0].id);
            } else if (currentContextId) {
                // Refresh current context view
                const currentContext = contexts.find(c => c.id === currentContextId);
                if (currentContext) {
                    renderNotesList();
                }
            }

            if (!silent) {
                updateSyncStatus('synced');
            }
        } else {
            throw new Error(`HTTP ${response.status}`);
        }
    } catch (error) {
        if (!silent) {
            updateSyncStatus('error', `Backend connection failed: ${error.message}`);
        }
    }
}

// Refresh contexts from server (fetch-only, no writes)
async function refreshContexts() {
    await loadContexts(false);
}

// Start auto-refresh interval
function startAutoRefresh() {
    // Clear any existing interval
    if (autoRefreshInterval) {
        clearInterval(autoRefreshInterval);
    }

    // Refresh every 10 seconds
    autoRefreshInterval = setInterval(async () => {
        await loadContexts(true); // Silent refresh
    }, 10000);
}

// Stop auto-refresh interval
function stopAutoRefresh() {
    if (autoRefreshInterval) {
        clearInterval(autoRefreshInterval);
        autoRefreshInterval = null;
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

    const deleteBtn = document.getElementById('delete-context-btn');
    deleteBtn.addEventListener('click', handleDeleteContext);

    // Refresh button
    const refreshBtn = document.getElementById('refresh-btn');
    refreshBtn.addEventListener('click', async () => {
        refreshBtn.classList.add('spinning');
        await refreshContexts();
        setTimeout(() => {
            refreshBtn.classList.remove('spinning');
        }, 500);
    });

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

    // Setup custom color select dropdown
    setupCustomSelect();
}

// Custom select dropdown functionality
function setupCustomSelect() {
    const customSelect = document.getElementById('context-color-input');
    const trigger = customSelect.querySelector('.custom-select-trigger');
    const options = customSelect.querySelectorAll('.custom-option');

    // Toggle dropdown
    trigger.addEventListener('click', (e) => {
        e.stopPropagation();
        customSelect.classList.toggle('open');
    });

    // Close dropdown when clicking outside
    document.addEventListener('click', (e) => {
        if (!customSelect.contains(e.target)) {
            customSelect.classList.remove('open');
        }
    });

    // Handle option selection
    options.forEach(option => {
        option.addEventListener('click', (e) => {
            e.stopPropagation();

            // Remove selected class from all options
            options.forEach(opt => opt.classList.remove('selected'));

            // Add selected class to clicked option
            option.classList.add('selected');

            // Update trigger display
            const colorValue = option.dataset.value;
            const colorLabel = option.querySelector('.color-label').textContent;
            const triggerPreview = trigger.querySelector('.color-preview');
            const triggerLabel = trigger.querySelector('.color-label');

            triggerPreview.dataset.color = colorValue;
            triggerLabel.textContent = colorLabel;

            // Close dropdown
            customSelect.classList.remove('open');
        });
    });
}

// Get selected color value from custom select
function getSelectedColor() {
    const selectedOption = document.querySelector('.custom-option.selected');
    if (selectedOption) {
        return selectedOption.dataset.value;
    }
    return 'purple'; // default
}

// Set selected color in custom select
function setSelectedColor(colorValue) {
    const customSelect = document.getElementById('context-color-input');
    const trigger = customSelect.querySelector('.custom-select-trigger');
    const options = customSelect.querySelectorAll('.custom-option');

    // Remove selected class from all options
    options.forEach(opt => opt.classList.remove('selected'));

    // Find and select the option with matching value
    const targetOption = Array.from(options).find(opt => opt.dataset.value === colorValue);
    if (targetOption) {
        targetOption.classList.add('selected');

        const colorLabel = targetOption.querySelector('.color-label').textContent;
        const triggerPreview = trigger.querySelector('.color-preview');
        const triggerLabel = trigger.querySelector('.color-label');

        triggerPreview.dataset.color = colorValue;
        triggerLabel.textContent = colorLabel;
    }
}

// Open create context modal
function openCreateContextModal() {
    isEditMode = false;
    editingContextId = null;

    const modal = document.getElementById('context-modal');
    const modalTitle = document.getElementById('modal-title');
    const form = document.getElementById('context-form');
    const deleteBtn = document.getElementById('delete-context-btn');

    modalTitle.textContent = 'Create New Context';
    form.reset();

    // Hide delete button in create mode
    deleteBtn.style.display = 'none';

    // Reset custom select to default
    setSelectedColor('purple');

    modal.classList.add('active');
}

// Open edit context modal
function openEditContextModal(contextId) {
    isEditMode = true;
    editingContextId = contextId;

    const context = contexts.find(c => c.id === contextId);
    if (!context) return;

    const modal = document.getElementById('context-modal');
    const modalTitle = document.getElementById('modal-title');
    const nameInput = document.getElementById('context-name-input');
    const descInput = document.getElementById('context-description-input');
    const deleteBtn = document.getElementById('delete-context-btn');

    modalTitle.textContent = 'Edit Context';
    nameInput.value = context.name;
    setSelectedColor(context.color || 'purple');
    descInput.value = context.description || '';

    // Show delete button in edit mode
    deleteBtn.style.display = 'block';

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
    const descInput = document.getElementById('context-description-input');

    // If editing, preserve existing notes; if creating, start with empty array
    const context = isEditMode && editingContextId ? contexts.find(c => c.id === editingContextId) : null;

    const contextData = {
        name: nameInput.value.trim(),
        color: getSelectedColor(),
        description: descInput.value.trim(),
        notes: context ? context.notes || [] : []
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

            updateSyncStatus('synced');
        } else {
            const error = await response.json();
            throw new Error(error.message || 'Failed to save context');
        }
    } catch (error) {
        updateSyncStatus('error', error.message);
    }
}

// Handle delete context
async function handleDeleteContext() {
    if (!editingContextId) return;

    const context = contexts.find(c => c.id === editingContextId);
    if (!context) return;

    // Confirm deletion
    const confirmed = confirm(`Are you sure you want to delete "${context.name}"? This action cannot be undone.`);
    if (!confirmed) return;

    try {
        updateSyncStatus('syncing');

        const response = await fetch(`/api/contexts/${editingContextId}`, {
            method: 'DELETE'
        });

        if (response.ok) {
            // Remove context from array
            const index = contexts.findIndex(c => c.id === editingContextId);
            if (index !== -1) {
                contexts.splice(index, 1);
            }

            closeModal();
            renderContextCards();

            // Select first available context or clear view
            if (contexts.length > 0) {
                selectContext(contexts[0].id);
            } else {
                currentContextId = null;
                const contextView = document.getElementById('context-view');
                const contextName = document.getElementById('context-name');
                contextName.textContent = 'Select a context';
                contextView.style.background = colorGradients.purple;
                document.getElementById('notes-section').style.display = 'none';
            }

            updateSyncStatus('synced');
        } else {
            const error = await response.json();
            throw new Error(error.message || 'Failed to delete context');
        }
    } catch (error) {
        updateSyncStatus('error', error.message);
    }
}

// Update sync status indicator
function updateSyncStatus(state, errorMessage = '') {
    const syncStatus = document.getElementById('sync-status');
    const syncText = document.getElementById('sync-text');
    const syncTime = document.getElementById('sync-time');

    // Update based on state
    switch(state) {
        case 'syncing':
            // Cancel any pending sync complete
            if (pendingSyncComplete) {
                clearTimeout(pendingSyncComplete);
                pendingSyncComplete = null;
            }

            // Record when sync started
            syncStartTime = Date.now();

            // Remove all state classes
            syncStatus.classList.remove('synced', 'syncing', 'error', 'just-synced');

            syncStatus.classList.add('syncing');
            syncText.textContent = 'Syncing...';
            syncTime.textContent = 'Updating...';
            break;

        case 'synced':
            const completeSyncTransition = () => {
                syncStatus.classList.remove('synced', 'syncing', 'error', 'just-synced');
                syncStatus.classList.add('synced');
                syncText.textContent = 'Synced';
                lastSyncTime = new Date();
                updateSyncTime();

                // Add attention-grabbing animation
                syncStatus.classList.add('just-synced');

                // Remove the animation class after it completes
                setTimeout(() => {
                    syncStatus.classList.remove('just-synced');
                }, 2000);
            };

            // Ensure minimum 1 second of syncing state
            if (syncStartTime) {
                const elapsed = Date.now() - syncStartTime;
                const minSyncDuration = 1000; // 1 second minimum

                if (elapsed < minSyncDuration) {
                    // Wait for remaining time
                    pendingSyncComplete = setTimeout(() => {
                        completeSyncTransition();
                        pendingSyncComplete = null;
                    }, minSyncDuration - elapsed);
                } else {
                    completeSyncTransition();
                }
            } else {
                completeSyncTransition();
            }
            break;

        case 'error':
            // Cancel any pending sync complete
            if (pendingSyncComplete) {
                clearTimeout(pendingSyncComplete);
                pendingSyncComplete = null;
            }

            syncStatus.classList.remove('synced', 'syncing', 'error', 'just-synced');
            syncStatus.classList.add('error');
            syncText.textContent = 'Error';
            syncTime.textContent = errorMessage || 'Sync failed';
            break;
    }
}

// Update sync time display with relative time
function updateSyncTime() {
    if (!lastSyncTime) return;

    const syncTime = document.getElementById('sync-time');
    const now = new Date();
    const diffMs = now - lastSyncTime;
    const diffSec = Math.floor(diffMs / 1000);
    const diffMin = Math.floor(diffSec / 60);
    const diffHour = Math.floor(diffMin / 60);

    if (diffSec < 10) {
        syncTime.textContent = 'Just now';
    } else if (diffSec < 60) {
        syncTime.textContent = `${diffSec} seconds ago`;
    } else if (diffMin < 60) {
        syncTime.textContent = `${diffMin} minute${diffMin > 1 ? 's' : ''} ago`;
    } else {
        syncTime.textContent = `${diffHour} hour${diffHour > 1 ? 's' : ''} ago`;
    }
}

// Update sync time every 10 seconds
setInterval(updateSyncTime, 10000);

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
            updateSyncStatus('synced');
        } else {
            throw new Error('Failed to update context name');
        }
    } catch (error) {
        updateSyncStatus('error', error.message);
        contextName.textContent = context.name; // Revert on error
    }
}

// Render inline notes list
function renderNotesList() {
    const context = contexts.find(c => c.id === currentContextId);
    if (!context) return;

    const notesList = document.getElementById('notes-list');
    notesList.innerHTML = '';

    // Ensure notes is an array of note objects
    const notes = context.notes || [];

    notes.forEach((note, index) => {
        const noteBox = document.createElement('div');
        noteBox.className = 'note-box';
        noteBox.draggable = true;
        noteBox.dataset.index = index;

        // Drag and drop events
        noteBox.addEventListener('dragstart', handleDragStart);
        noteBox.addEventListener('dragend', handleDragEnd);
        noteBox.addEventListener('dragover', handleDragOver);
        noteBox.addEventListener('drop', handleDrop);
        noteBox.addEventListener('dragenter', handleDragEnter);
        noteBox.addEventListener('dragleave', handleDragLeave);

        // Title input
        const titleInput = document.createElement('input');
        titleInput.className = 'note-title';
        titleInput.type = 'text';
        titleInput.value = typeof note === 'object' ? (note.title || '') : '';
        titleInput.placeholder = 'Note Title';

        // Prevent drag when interacting with input
        titleInput.addEventListener('mousedown', (e) => {
            noteBox.draggable = false;
        });
        titleInput.addEventListener('mouseup', (e) => {
            noteBox.draggable = true;
        });

        // Auto-save title on blur
        titleInput.addEventListener('blur', () => {
            updateNoteTitle(index, titleInput.value);
        });

        // Content textarea
        const textarea = document.createElement('textarea');
        textarea.className = 'note-textarea';
        textarea.value = typeof note === 'object' ? (note.content || '') : (typeof note === 'string' ? note : '');
        textarea.placeholder = 'Type your note here...';

        // Prevent drag when interacting with textarea
        textarea.addEventListener('mousedown', (e) => {
            noteBox.draggable = false;
        });
        textarea.addEventListener('mouseup', (e) => {
            noteBox.draggable = true;
        });

        // Auto-save content on blur
        textarea.addEventListener('blur', () => {
            updateNoteContent(index, textarea.value);
        });

        // Remove button
        const removeBtn = document.createElement('button');
        removeBtn.className = 'btn-remove-inline';
        removeBtn.innerHTML = '&times;';
        removeBtn.onclick = () => removeNote(index);

        noteBox.appendChild(titleInput);
        noteBox.appendChild(textarea);
        noteBox.appendChild(removeBtn);
        notesList.appendChild(noteBox);
    });

    // Add "Add Note" button
    const addBtn = document.createElement('button');
    addBtn.className = 'add-note-btn';
    addBtn.textContent = '+ Add Note';
    addBtn.onclick = addNote;
    notesList.appendChild(addBtn);
}

// Drag and drop handlers
let draggedElement = null;
let draggedIndex = null;

function handleDragStart(e) {
    draggedElement = this;
    draggedIndex = parseInt(this.dataset.index);
    this.classList.add('dragging');
    e.dataTransfer.effectAllowed = 'move';
    e.dataTransfer.setData('text/plain', draggedIndex);
}

function handleDragEnd(e) {
    this.classList.remove('dragging');
    document.querySelectorAll('.note-box').forEach(box => {
        box.classList.remove('drag-over');
    });
    draggedElement = null;
    draggedIndex = null;
}

function handleDragOver(e) {
    if (e.preventDefault) {
        e.preventDefault();
    }
    e.dataTransfer.dropEffect = 'move';
    return false;
}

function handleDragEnter(e) {
    if (this !== draggedElement && this.classList.contains('note-box')) {
        this.classList.add('drag-over');
    }
}

function handleDragLeave(e) {
    this.classList.remove('drag-over');
}

async function handleDrop(e) {
    if (e.stopPropagation) {
        e.stopPropagation();
    }
    if (e.preventDefault) {
        e.preventDefault();
    }

    this.classList.remove('drag-over');

    if (draggedElement !== this && this.classList.contains('note-box')) {
        const context = contexts.find(c => c.id === currentContextId);
        if (!context) return false;

        const fromIndex = draggedIndex;
        const toIndex = parseInt(this.dataset.index);

        // Reorder notes array
        const notes = [...context.notes];
        const [movedNote] = notes.splice(fromIndex, 1);
        notes.splice(toIndex, 0, movedNote);

        // Update context
        context.notes = notes;

        // Update on server and re-render
        await updateContextOnServer(context);
        renderNotesList();
        renderContextCards();
    }

    return false;
}

// Add note inline
async function addNote() {
    const context = contexts.find(c => c.id === currentContextId);
    if (!context) return;

    context.notes = context.notes || [];
    context.notes.push({ title: '', content: '' });

    await updateContextOnServer(context);
    renderNotesList();
    renderContextCards();

    // Focus the newly added title input
    const notesList = document.getElementById('notes-list');
    const titleInputs = notesList.querySelectorAll('.note-title');
    if (titleInputs.length > 0) {
        titleInputs[titleInputs.length - 1].focus();
    }
}

// Update note title
async function updateNoteTitle(index, title) {
    const context = contexts.find(c => c.id === currentContextId);
    if (!context) return;

    // Ensure note is an object
    if (typeof context.notes[index] !== 'object') {
        context.notes[index] = { title: '', content: context.notes[index] || '' };
    }

    context.notes[index].title = title;
    await updateContextOnServer(context);
    renderContextCards();
}

// Update note content
async function updateNoteContent(index, content) {
    const context = contexts.find(c => c.id === currentContextId);
    if (!context) return;

    // Ensure note is an object
    if (typeof context.notes[index] !== 'object') {
        context.notes[index] = { title: '', content: '' };
    }

    context.notes[index].content = content;
    await updateContextOnServer(context);
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

// Check for conflicts before saving
async function checkForConflicts(contextId) {
    try {
        // Fetch latest server state
        const response = await fetch(`/api/contexts/${contextId}`);
        if (!response.ok) {
            return null;
        }

        const serverContext = await response.json();
        const localContext = contexts.find(c => c.id === contextId);
        const cachedServerContext = serverContexts.find(c => c.id === contextId);

        if (!localContext || !cachedServerContext) {
            return null;
        }

        // Check if server version differs from our cached version
        const serverChanged = JSON.stringify(serverContext) !== JSON.stringify(cachedServerContext);

        if (serverChanged) {
            return {
                serverContext,
                localContext,
                hasConflict: true
            };
        }

        return null;
    } catch (error) {
        console.error('Error checking for conflicts:', error);
        return null;
    }
}

// Show conflict warning dialog
function showConflictWarning(conflictInfo) {
    return new Promise((resolve) => {
        const modal = document.createElement('div');
        modal.className = 'modal active';
        modal.style.zIndex = '2000';

        modal.innerHTML = `
            <div class="modal-content" style="max-width: 600px;">
                <div class="modal-header">
                    <h3 style="color: #e53e3e;">⚠️ Conflict Detected</h3>
                </div>
                <div class="modal-body">
                    <p style="margin-bottom: 16px; color: #2d3748; line-height: 1.6;">
                        The context has been modified on the server since you last loaded it.
                        If you continue, your changes will overwrite the server version.
                    </p>
                    <p style="margin-bottom: 16px; color: #4a5568; font-size: 0.9rem;">
                        <strong>Server version:</strong> ${conflictInfo.serverContext.notes?.length || 0} notes<br>
                        <strong>Your version:</strong> ${conflictInfo.localContext.notes?.length || 0} notes
                    </p>
                    <div class="form-actions" style="justify-content: center; gap: 16px;">
                        <button class="btn btn-secondary" id="conflict-cancel">Cancel</button>
                        <button class="btn btn-primary" id="conflict-reload">Reload from Server</button>
                        <button class="btn btn-danger" id="conflict-overwrite">Overwrite Server</button>
                    </div>
                </div>
            </div>
        `;

        document.body.appendChild(modal);

        document.getElementById('conflict-cancel').onclick = () => {
            document.body.removeChild(modal);
            resolve('cancel');
        };

        document.getElementById('conflict-reload').onclick = () => {
            document.body.removeChild(modal);
            resolve('reload');
        };

        document.getElementById('conflict-overwrite').onclick = () => {
            document.body.removeChild(modal);
            resolve('overwrite');
        };

        // Close on background click
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                document.body.removeChild(modal);
                resolve('cancel');
            }
        });
    });
}

// Update context on server
async function updateContextOnServer(context, skipConflictCheck = false) {
    // Check for conflicts before saving
    if (!skipConflictCheck) {
        const conflictInfo = await checkForConflicts(context.id);
        if (conflictInfo) {
            const userChoice = await showConflictWarning(conflictInfo);

            if (userChoice === 'cancel') {
                return;
            } else if (userChoice === 'reload') {
                // Reload from server
                await loadContexts(false);
                return;
            }
            // If 'overwrite', proceed with save
        }
    }

    updateSyncStatus('syncing');

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

            // Update server cache
            const serverIndex = serverContexts.findIndex(c => c.id === context.id);
            if (serverIndex !== -1) {
                serverContexts[serverIndex] = JSON.parse(JSON.stringify(result));
            }

            updateSyncStatus('synced');
        } else {
            throw new Error('Failed to update context');
        }
    } catch (error) {
        updateSyncStatus('error', error.message);
    }
}

// Make functions globally accessible for inline onclick
window.closeModal = closeModal;

console.log('Vibro frontend initialized');
