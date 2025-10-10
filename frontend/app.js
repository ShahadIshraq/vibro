// Simple API health check on page load
document.addEventListener('DOMContentLoaded', async () => {
    const statusDiv = document.getElementById('status');

    try {
        const response = await fetch('/api/contexts');

        if (response.ok) {
            const contexts = await response.json();
            statusDiv.textContent = `✓ Backend connected! Found ${contexts.length} context(s).`;
            statusDiv.className = 'success';
        } else {
            throw new Error(`HTTP ${response.status}`);
        }
    } catch (error) {
        statusDiv.textContent = `✗ Backend connection failed: ${error.message}`;
        statusDiv.className = 'error';
    }
});

// Utility function for API calls (to be used by frontend implementation)
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

console.log('Vibro frontend initialized');
