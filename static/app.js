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
            <a href="${item.html_url}/download" class="download-btn" target="_blank">Download JSON</a>
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
            <a href="/template/${template.id}" class="download-btn">View Details</a>
            <a href="/api/templates/${template.id}/download" class="download-btn" target="_blank">Download JSON</a>
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

// Template Creation Wizard
let currentStep = 1;
let packageData = {
    brews: [],
    casks: [],
    taps: [],
    stow: []
};

// Step Navigation
function nextStep() {
    if (validateCurrentStep()) {
        if (currentStep < 4) {
            currentStep++;
            updateWizardStep();
        }
    }
}

function previousStep() {
    if (currentStep > 1) {
        currentStep--;
        updateWizardStep();
    }
}

function updateWizardStep() {
    // Update step indicators
    document.querySelectorAll('.step').forEach((step, index) => {
        const stepNum = index + 1;
        step.classList.remove('active', 'completed');

        if (stepNum === currentStep) {
            step.classList.add('active');
        } else if (stepNum < currentStep) {
            step.classList.add('completed');
        }
    });

    // Show/hide form steps
    document.querySelectorAll('.form-step').forEach((step, index) => {
        step.classList.remove('active');
        if (index + 1 === currentStep) {
            step.classList.add('active');
        }
    });

    // Update navigation buttons
    const prevBtn = document.getElementById('prev-btn');
    const nextBtn = document.getElementById('next-btn');
    const submitBtn = document.getElementById('submit-btn');

    prevBtn.style.display = currentStep === 1 ? 'none' : 'inline-flex';

    if (currentStep === 4) {
        nextBtn.style.display = 'none';
        submitBtn.style.display = 'inline-flex';
        generatePreview();
    } else {
        nextBtn.style.display = 'inline-flex';
        submitBtn.style.display = 'none';
    }
}

function validateCurrentStep() {
    switch(currentStep) {
        case 1:
            const name = document.getElementById('template-name').value.trim();
            const author = document.getElementById('template-author').value.trim();

            if (!name) {
                showError('Template name is required');
                document.getElementById('template-name').focus();
                return false;
            }
            if (!author) {
                showError('Author name is required');
                document.getElementById('template-author').focus();
                return false;
            }
            return true;
        case 2:
            // Package step is optional, always valid
            return true;
        case 3:
            // Advanced options are optional
            updateOverridesField();
            return true;
        default:
            return true;
    }
}

function showError(message) {
    // Remove existing error messages
    document.querySelectorAll('.error-message').forEach(msg => msg.remove());

    const errorDiv = document.createElement('div');
    errorDiv.className = 'error-message';
    errorDiv.textContent = message;

    const activeStep = document.querySelector('.form-step.active');
    activeStep.insertBefore(errorDiv, activeStep.firstChild);

    setTimeout(() => errorDiv.remove(), 5000);
}

// Package Management
function addTag(tag) {
    const input = document.getElementById('template-tags-input');
    const currentTags = input.value.split(',').map(t => t.trim()).filter(t => t);

    if (!currentTags.includes(tag)) {
        currentTags.push(tag);
        input.value = currentTags.join(', ');
    }
}

function addPackage(type, inputId) {
    const input = document.getElementById(inputId);
    const packageName = input.value.trim();

    if (packageName && !packageData[type].includes(packageName)) {
        packageData[type].push(packageName);
        input.value = '';
        updatePackageDisplay(type);
        updateHiddenField(type);
    }
}

function addPackageFromPill(type, packageName) {
    if (!packageData[type].includes(packageName)) {
        packageData[type].push(packageName);
        updatePackageDisplay(type);
        updateHiddenField(type);
    }
}

function removePackage(type, packageName) {
    const index = packageData[type].indexOf(packageName);
    if (index > -1) {
        packageData[type].splice(index, 1);
        updatePackageDisplay(type);
        updateHiddenField(type);
    }
}

function updatePackageDisplay(type) {
    const container = document.getElementById(`${type}-list`);
    container.innerHTML = '';

    packageData[type].forEach(pkg => {
        const tag = document.createElement('div');
        tag.className = 'package-tag';
        tag.innerHTML = `
            ${pkg}
            <button type="button" class="remove-btn" onclick="removePackage('${type}', '${pkg}')">×</button>
        `;
        container.appendChild(tag);
    });
}

function updateHiddenField(type) {
    const field = document.getElementById(`template-${type}`);
    if (field) {
        field.value = packageData[type].join(', ');
    }
}

function updateOverridesField() {
    const checkboxes = document.querySelectorAll('.checkbox-pills input[type="checkbox"]:checked');
    const overrides = Array.from(checkboxes).map(cb => cb.value);
    document.getElementById('template-overrides').value = overrides.join(', ');
}

// Character counters
function setupCharCounters() {
    const inputs = [
        { id: 'template-name', counterId: 'name-count', max: 50 },
        { id: 'template-author', counterId: 'author-count', max: 30 },
        { id: 'template-description', counterId: 'desc-count', max: 500 }
    ];

    inputs.forEach(({ id, counterId, max }) => {
        const input = document.getElementById(id);
        const counter = document.getElementById(counterId);

        input.addEventListener('input', () => {
            const length = input.value.length;
            counter.textContent = length;
            counter.style.color = length > max * 0.9 ? 'var(--accent-warning)' : 'var(--text-muted)';
        });
    });
}

// Enter key handling for package inputs
function setupPackageInputs() {
    const inputs = [
        { id: 'brew-input', type: 'brews' },
        { id: 'cask-input', type: 'casks' },
        { id: 'tap-input', type: 'taps' },
        { id: 'stow-input', type: 'stow' }
    ];

    inputs.forEach(({ id, type }) => {
        const input = document.getElementById(id);
        input.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                addPackage(type, id);
            }
        });
    });
}

// Preview Generation
function generatePreview() {
    const template = {
        name: document.getElementById('template-name').value,
        author: document.getElementById('template-author').value,
        description: document.getElementById('template-description').value,
        version: document.getElementById('template-version').value || '1.0.0',
        tags: document.getElementById('template-tags-input').value.split(',').map(t => t.trim()).filter(t => t),
        public: document.getElementById('template-public').checked,
        addOnly: document.getElementById('template-add-only').checked,
        extends: document.getElementById('template-extends').value,
        overrides: document.getElementById('template-overrides').value.split(',').map(t => t.trim()).filter(t => t),
        ...packageData
    };

    const totalPackages = template.brews.length + template.casks.length +
                          template.taps.length + template.stow.length;

    const previewContainer = document.getElementById('template-preview');
    previewContainer.innerHTML = `
        <div class="template-header">
            <h3>${template.name}${template.featured ? '<span class="featured-badge">Featured</span>' : ''}</h3>
            <div class="author">by ${template.author}</div>
            <div class="description">${template.description || 'No description provided'}</div>
        </div>

        ${template.tags.length ? `
        <div class="template-tags" style="margin: 15px 0;">
            ${template.tags.map(tag => `<span class="tag">${tag}</span>`).join('')}
        </div>` : ''}

        <div class="template-packages">
            <div style="margin-bottom: 15px; color: var(--text-secondary);">
                <strong>📊 Package Summary: ${totalPackages} total packages</strong>
            </div>

            ${template.brews.length ? `
            <div class="package-group">
                <strong>🍺 Brews (${template.brews.length})</strong>
                <div class="package-list">${template.brews.join(', ')}</div>
            </div>` : ''}

            ${template.casks.length ? `
            <div class="package-group">
                <strong>📦 Casks (${template.casks.length})</strong>
                <div class="package-list">${template.casks.join(', ')}</div>
            </div>` : ''}

            ${template.taps.length ? `
            <div class="package-group">
                <strong>📋 Taps (${template.taps.length})</strong>
                <div class="package-list">${template.taps.join(', ')}</div>
            </div>` : ''}

            ${template.stow.length ? `
            <div class="package-group">
                <strong>🔗 Stow (${template.stow.length})</strong>
                <div class="package-list">${template.stow.join(', ')}</div>
            </div>` : ''}
        </div>

        <div style="margin-top: 20px; padding-top: 15px; border-top: 1px solid var(--border-color); font-size: 0.9em; color: var(--text-muted);">
            <div><strong>Version:</strong> ${template.version}</div>
            <div><strong>Visibility:</strong> ${template.public ? 'Public' : 'Private'}</div>
            ${template.addOnly ? '<div><strong>Mode:</strong> Add-only (won\'t replace existing packages)</div>' : ''}
            ${template.extends ? `<div><strong>Extends:</strong> ${template.extends}</div>` : ''}
            ${template.overrides.length ? `<div><strong>Overrides:</strong> ${template.overrides.join(', ')}</div>` : ''}
        </div>
    `;
}

// Enhanced template submission
async function submitTemplateEnhanced(event) {
    event.preventDefault();

    const submitBtn = document.getElementById('submit-btn');
    const originalText = submitBtn.textContent;
    submitBtn.disabled = true;
    submitBtn.innerHTML = '🚀 Creating...';

    // Clear previous messages
    document.querySelectorAll('.success-message, .error-message').forEach(msg => msg.remove());

    // Update overrides field
    updateOverridesField();

    // Prepare template data
    const templateData = {
        taps: packageData.taps,
        brews: packageData.brews,
        casks: packageData.casks,
        stow: packageData.stow,
        organization_id: document.getElementById('template-organization').value || '',
        metadata: {
            name: document.getElementById('template-name').value,
            description: document.getElementById('template-description').value,
            author: document.getElementById('template-author').value,
            tags: document.getElementById('template-tags-input').value.split(',').map(t => t.trim()).filter(t => t),
            version: document.getElementById('template-version').value || '1.0.0'
        },
        extends: document.getElementById('template-extends').value || '',
        overrides: document.getElementById('template-overrides').value.split(',').map(t => t.trim()).filter(t => t),
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

            // Show success message with animation
            const successDiv = document.createElement('div');
            successDiv.className = 'success-message';
            successDiv.innerHTML = `
                <div style="display: flex; align-items: center; gap: 10px;">
                    <div style="font-size: 1.5em;">🎉</div>
                    <div>
                        <div style="font-weight: 600;">Template created successfully!</div>
                        <div style="font-size: 0.9em; margin-top: 5px;">
                            <a href="/template/${result.id}" style="color: var(--accent-success); text-decoration: underline;">
                                View your template →
                            </a>
                        </div>
                    </div>
                </div>
            `;

            const activeStep = document.querySelector('.form-step.active');
            activeStep.insertBefore(successDiv, activeStep.firstChild);

            // Reset form and go back to step 1
            setTimeout(() => {
                resetWizard();
                switchTab('templates');
                loadTemplates();
            }, 3000);

        } else {
            const error = await response.json();
            showError(`Failed to create template: ${error.error}`);
        }
    } catch (error) {
        console.error('Failed to create template:', error);
        showError('Failed to create template. Please check your connection and try again.');
    } finally {
        submitBtn.disabled = false;
        submitBtn.innerHTML = originalText;
    }
}

function resetWizard() {
    currentStep = 1;
    packageData = { brews: [], casks: [], taps: [], stow: [] };

    document.getElementById('template-form').reset();
    document.getElementById('template-version').value = '1.0.0';

    Object.keys(packageData).forEach(type => {
        updatePackageDisplay(type);
        updateHiddenField(type);
    });

    updateWizardStep();
}

// Copy install command functionality
function copyInstallCommand() {
    const command = document.getElementById('install-cmd').textContent;
    const button = event.target;

    navigator.clipboard.writeText(command).then(() => {
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

        const originalText = button.textContent;
        button.textContent = '✓ Copied!';
        button.classList.add('copied');

        setTimeout(() => {
            button.textContent = originalText;
            button.classList.remove('copied');
        }, 2000);
    });
}

// Initialize when page loads
document.addEventListener('DOMContentLoaded', function() {
    // Load initial data
    loadStats();
    loadTemplates(); // Start with templates instead of configs

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
        templateForm.addEventListener('submit', submitTemplateEnhanced);
        setupCharCounters();
        setupPackageInputs();
        updateWizardStep();
        loadUserOrganizations(); // Load organizations for form
    }
});

// Load user organizations for template form
async function loadUserOrganizations() {
    try {
        // Check if user is authenticated
        const authResponse = await fetch('/auth/user');
        if (!authResponse.ok) {
            return; // User not authenticated
        }

        const user = await authResponse.json();

        // Load user's organization memberships
        const orgsResponse = await fetch('/api/organizations');
        if (!orgsResponse.ok) {
            return;
        }

        const orgData = await orgsResponse.json();
        const organizations = orgData.organizations || [];

        // Filter to only show organizations where the user is a member
        // For now, show all public organizations (we'll enhance this later)
        const orgSelect = document.getElementById('template-organization');
        if (orgSelect && organizations.length > 0) {
            // Clear existing options except the first
            orgSelect.innerHTML = '<option value="">Personal Template</option>';

            organizations.forEach(org => {
                if (org.public) {
                    const option = document.createElement('option');
                    option.value = org.id;
                    option.textContent = org.name;
                    orgSelect.appendChild(option);
                }
            });
        }
    } catch (error) {
        console.error('Failed to load user organizations:', error);
    }
}
