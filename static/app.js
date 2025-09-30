let allConfigs = [];
let allTemplates = [];

// Navigation functionality
function switchTab(tabName) {
    // Remove active class from all tabs and sections
    document.querySelectorAll('.nav-tab').forEach(tab => tab.classList.remove('active'));
    document.querySelectorAll('.content-section').forEach(section => section.classList.remove('active'));

    // Add active class to clicked tab and corresponding section
    event.target.classList.add('active');
    document.getElementById(`${tabName}-section`).classList.add('active');

    // Load data for the active tab
    if (tabName === 'templates') {
        loadTemplates();
    } else if (tabName === 'configs') {
        loadConfigs();
    }
}

// Load stats
async function loadStats() {
    try {
        const response = await fetch('/api/configs/stats');
        const stats = await response.json();

        document.getElementById('total-configs').textContent = stats.total_configs || 0;
        document.getElementById('public-configs').textContent = stats.public_configs || 0;
        document.getElementById('total-downloads').textContent = stats.total_downloads || 0;
    } catch (error) {
        console.error('Failed to load stats:', error);
    }
}

// Load all configurations
async function loadConfigs() {
    try {
        const response = await fetch('/api/configs/search');
        const data = await response.json();
        allConfigs = data.items || [];
        displayConfigs(allConfigs);
    } catch (error) {
        console.error('Failed to load configs:', error);
        document.getElementById('configs-container').innerHTML =
            '<div class="empty"><h3>Failed to load configurations</h3><p>Please try again later.</p></div>';
    }
}

// Load all templates
async function loadTemplates() {
    try {
        const response = await fetch('/api/templates');
        const data = await response.json();
        allTemplates = data.templates || [];
        displayTemplates(allTemplates);
    } catch (error) {
        console.error('Failed to load templates:', error);
        document.getElementById('templates-container').innerHTML =
            '<div class="empty"><h3>Failed to load templates</h3><p>Please try again later.</p></div>';
    }
}

// Display configurations
function displayConfigs(configs) {
    const container = document.getElementById('configs-container');

    if (configs.length === 0) {
        container.innerHTML = '<div class="empty"><h3>No configurations found</h3><p>Be the first to share your dotfiles configuration!</p></div>';
        return;
    }

    const grid = document.createElement('div');
    grid.className = 'configs-grid';

    configs.forEach(async (item) => {
        try {
            // Fetch full config details
            const configResponse = await fetch(item.html_url);
            const config = await configResponse.json();

            const card = createConfigCard(config, item);
            grid.appendChild(card);
        } catch (error) {
            console.error('Failed to load config details:', error);
        }
    });

    container.innerHTML = '';
    container.appendChild(grid);
}

// Display templates
function displayTemplates(templates) {
    const container = document.getElementById('templates-container');

    if (templates.length === 0) {
        container.innerHTML = '<div class="empty"><h3>No templates found</h3><p>Create the first template!</p></div>';
        return;
    }

    const grid = document.createElement('div');
    grid.className = 'templates-grid';

    templates.forEach(template => {
        const card = createTemplateCard(template);
        grid.appendChild(card);
    });

    container.innerHTML = '';
    container.appendChild(grid);
}

// Create configuration card
function createConfigCard(config, item) {
    const card = document.createElement('div');
    card.className = 'config-card';

    const packages = {
        brews: config.brews || [],
        casks: config.casks || [],
        taps: config.taps || [],
        stow: config.stow || []
    };

    const totalPackages = packages.brews.length + packages.casks.length + packages.taps.length + packages.stow.length;

    card.innerHTML = `
        <div class="config-header">
            <h3>${config.metadata?.name || 'Unnamed Configuration'}</h3>
            <div class="author">by ${config.metadata?.author || item.owner?.login || 'Anonymous'}</div>
            <div class="description">${config.metadata?.description || 'No description provided'}</div>
        </div>
        ${config.metadata?.tags?.length ? `
        <div class="config-tags">
            ${config.metadata.tags.map(tag => `<span class="tag">${tag}</span>`).join('')}
        </div>` : ''}
        <div class="config-packages">
            ${packages.brews.length ? `
            <div class="package-group">
                <strong>🍺 Brews (${packages.brews.length})</strong>
                <div class="package-list">${packages.brews.slice(0, 10).join(', ')}${packages.brews.length > 10 ? '...' : ''}</div>
            </div>` : ''}
            ${packages.casks.length ? `
            <div class="package-group">
                <strong>📦 Casks (${packages.casks.length})</strong>
                <div class="package-list">${packages.casks.slice(0, 10).join(', ')}${packages.casks.length > 10 ? '...' : ''}</div>
            </div>` : ''}
            ${packages.taps.length ? `
            <div class="package-group">
                <strong>📋 Taps (${packages.taps.length})</strong>
                <div class="package-list">${packages.taps.join(', ')}</div>
            </div>` : ''}
            ${packages.stow.length ? `
            <div class="package-group">
                <strong>🔗 Stow (${packages.stow.length})</strong>
                <div class="package-list">${packages.stow.join(', ')}</div>
            </div>` : ''}
            ${totalPackages === 0 ? '<div class="package-list">No packages defined</div>' : ''}
        </div>
        <div class="command-container">
            <strong>📋 Import Command:</strong><br>
            <code id="command-${item.id}">dotfiles clone ${window.location.origin}${item.html_url}</code>
        </div>
        <div class="config-footer">
            <span>${new Date(config.metadata?.created_at || item.created_at).toLocaleDateString()}</span>
            <button class="copy-command" onclick="copyCommand('${item.id}')">Copy Command</button>
            <a href="${item.html_url}" class="download-btn" target="_blank">Raw JSON</a>
        </div>
    `;

    return card;
}

// Create template card
function createTemplateCard(template) {
    const card = document.createElement('div');
    card.className = 'template-card';

    const packages = {
        brews: template.brews || [],
        casks: template.casks || [],
        taps: template.taps || [],
        stow: template.stow || []
    };

    const totalPackages = packages.brews.length + packages.casks.length + packages.taps.length + packages.stow.length;

    card.innerHTML = `
        <div class="template-header">
            <h3>${template.name}${template.featured ? '<span class="featured-badge">Featured</span>' : ''}</h3>
            <div class="author">by ${template.author}</div>
            <div class="description">${template.description || 'No description provided'}</div>
        </div>
        ${template.tags?.length ? `
        <div class="template-tags">
            ${template.tags.map(tag => `<span class="tag">${tag}</span>`).join('')}
        </div>` : ''}
        <div class="template-packages">
            ${packages.brews.length ? `
            <div class="package-group">
                <strong>🍺 Brews (${packages.brews.length})</strong>
                <div class="package-list">${packages.brews.slice(0, 10).join(', ')}${packages.brews.length > 10 ? '...' : ''}</div>
            </div>` : ''}
            ${packages.casks.length ? `
            <div class="package-group">
                <strong>📦 Casks (${packages.casks.length})</strong>
                <div class="package-list">${packages.casks.slice(0, 10).join(', ')}${packages.casks.length > 10 ? '...' : ''}</div>
            </div>` : ''}
            ${packages.taps.length ? `
            <div class="package-group">
                <strong>📋 Taps (${packages.taps.length})</strong>
                <div class="package-list">${packages.taps.join(', ')}</div>
            </div>` : ''}
            ${packages.stow.length ? `
            <div class="package-group">
                <strong>🔗 Stow (${packages.stow.length})</strong>
                <div class="package-list">${packages.stow.join(', ')}</div>
            </div>` : ''}
            ${totalPackages === 0 ? '<div class="package-list">No packages defined</div>' : ''}
        </div>
        <div class="template-footer">
            <span>${template.downloads} downloads • ${new Date(template.updated_at).toLocaleDateString()}</span>
            <a href="/api/templates/${template.id}" class="download-btn" target="_blank">View Details</a>
        </div>
    `;

    return card;
}

// Copy command functionality
function copyCommand(configId) {
    const commandElement = document.getElementById(`command-${configId}`);
    const command = commandElement.textContent;

    navigator.clipboard.writeText(command).then(() => {
        // Find the button and update its appearance
        const button = event.target;
        const originalText = button.textContent;
        button.textContent = '✓ Copied!';
        button.classList.add('copied');

        setTimeout(() => {
            button.textContent = originalText;
            button.classList.remove('copied');
        }, 2000);
    }).catch(err => {
        console.error('Failed to copy command:', err);
        // Fallback for browsers that don't support clipboard API
        const textArea = document.createElement('textarea');
        textArea.value = command;
        document.body.appendChild(textArea);
        textArea.select();
        document.execCommand('copy');
        document.body.removeChild(textArea);

        const button = event.target;
        const originalText = button.textContent;
        button.textContent = '✓ Copied!';
        button.classList.add('copied');

        setTimeout(() => {
            button.textContent = originalText;
            button.classList.remove('copied');
        }, 2000);
    });
}

// Search functionality for configs
document.addEventListener('DOMContentLoaded', function() {
    const searchInput = document.getElementById('search');
    if (searchInput) {
        searchInput.addEventListener('input', (e) => {
            const query = e.target.value.toLowerCase();

            if (!query) {
                displayConfigs(allConfigs);
                return;
            }

            const filtered = allConfigs.filter(item => {
                const searchText = (item.description || '').toLowerCase();
                return searchText.includes(query);
            });

            displayConfigs(filtered);
        });
    }
});

// Template search and filtering
async function searchTemplates() {
    const search = document.getElementById('template-search').value;
    const tags = document.getElementById('template-tags').value;
    const featured = document.getElementById('template-featured').value;

    try {
        const params = new URLSearchParams();
        if (search) params.append('search', search);
        if (tags) params.append('tags', tags);
        if (featured) params.append('featured', featured);

        const response = await fetch(`/api/templates?${params.toString()}`);
        const data = await response.json();
        displayTemplates(data.templates || []);
    } catch (error) {
        console.error('Failed to search templates:', error);
    }
}

// Template form submission
async function submitTemplate(event) {
    event.preventDefault();

    const submitBtn = document.querySelector('.submit-btn');
    const originalText = submitBtn.textContent;
    submitBtn.disabled = true;
    submitBtn.textContent = 'Creating...';

    // Clear previous messages
    document.querySelectorAll('.success-message, .error-message').forEach(msg => msg.remove());

    const templateData = {
        taps: parseCommaSeparated(document.getElementById('template-taps').value),
        brews: parseCommaSeparated(document.getElementById('template-brews').value),
        casks: parseCommaSeparated(document.getElementById('template-casks').value),
        stow: parseCommaSeparated(document.getElementById('template-stow').value),
        metadata: {
            name: document.getElementById('template-name').value,
            description: document.getElementById('template-description').value,
            author: document.getElementById('template-author').value,
            tags: parseCommaSeparated(document.getElementById('template-tags-input').value),
            version: document.getElementById('template-version').value || '1.0.0'
        },
        extends: document.getElementById('template-extends').value || '',
        overrides: parseCommaSeparated(document.getElementById('template-overrides').value),
        addOnly: document.getElementById('template-add-only').checked,
        public: document.getElementById('template-public').checked,
        featured: false
    };

    try {
        const response = await fetch('/api/templates', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(templateData)
        });

        if (response.ok) {
            const result = await response.json();
            showMessage('success', `Template created successfully! ID: ${result.id}`);
            document.getElementById('template-form').reset();
            // Switch to templates tab to show the new template
            switchTab('templates');
        } else {
            const error = await response.json();
            showMessage('error', `Failed to create template: ${error.error}`);
        }
    } catch (error) {
        console.error('Failed to create template:', error);
        showMessage('error', 'Failed to create template. Please try again.');
    } finally {
        submitBtn.disabled = false;
        submitBtn.textContent = originalText;
    }
}

// Utility functions
function parseCommaSeparated(value) {
    return value ? value.split(',').map(item => item.trim()).filter(item => item) : [];
}

function showMessage(type, message) {
    const messageDiv = document.createElement('div');
    messageDiv.className = `${type}-message`;
    messageDiv.textContent = message;

    const form = document.querySelector('.template-form');
    form.insertBefore(messageDiv, form.firstChild);

    // Auto-remove success messages after 5 seconds
    if (type === 'success') {
        setTimeout(() => {
            messageDiv.remove();
        }, 5000);
    }
}

// Initialize when page loads
document.addEventListener('DOMContentLoaded', function() {
    // Load initial data
    loadStats();
    loadConfigs();

    // Add event listeners for template search
    const templateSearch = document.getElementById('template-search');
    const templateTags = document.getElementById('template-tags');
    const templateFeatured = document.getElementById('template-featured');

    if (templateSearch) {
        templateSearch.addEventListener('input', searchTemplates);
    }
    if (templateTags) {
        templateTags.addEventListener('input', searchTemplates);
    }
    if (templateFeatured) {
        templateFeatured.addEventListener('change', searchTemplates);
    }

    // Add event listener for template form
    const templateForm = document.getElementById('template-form');
    if (templateForm) {
        templateForm.addEventListener('submit', submitTemplate);
    }
});