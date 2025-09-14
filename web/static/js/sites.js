let lastChecked = null;
let selectedSites = new Set();

// Отображение сайтов
function displaySites(sites) {
    const container = document.getElementById('sites-container');
    
    if (sites.length === 0) {
        container.innerHTML = '<div class="no-sites">Нет добавленных сайтов</div>';
        document.getElementById('bulk-actions').classList.remove('visible');
        return;
    }

    container.innerHTML = sites.map(site => `
        <div class="site-item ${selectedSites.has(site.id) ? 'selected' : ''}">
            <input 
                type="checkbox" 
                class="site-checkbox" 
                data-site-id="${site.id}"
                ${selectedSites.has(site.id) ? 'checked' : ''}
                onchange="handleCheckboxChange(this, event)"
            >
            <div class="site-info">
                <div class="site-url">${site.url}</div>
                <div class="site-status ${site.last_status?.toLowerCase() || 'unknown'}">
                    ${site.last_status || 'UNKNOWN'} • ${new Date(site.last_checked).toLocaleString()}
                </div>
            </div>
            <div class="site-actions">
                <button class="danger-btn" onclick="deleteSite(${site.id})">Удалить</button>
            </div>
        </div>
    `).join('');

    updateBulkActions();
}

// Обработчик чекбоксов с поддержкой Shift
function handleCheckboxChange(checkbox, event) {
    const siteId = parseInt(checkbox.dataset.siteId);
    
    if (event.shiftKey && lastChecked) {
        // Selection с Shift
        const checkboxes = Array.from(document.querySelectorAll('.site-checkbox'));
        const startIndex = checkboxes.indexOf(lastChecked);
        const endIndex = checkboxes.indexOf(checkbox);
        
        const start = Math.min(startIndex, endIndex);
        const end = Math.max(startIndex, endIndex);
        
        const isChecking = checkbox.checked;
        
        for (let i = start; i <= end; i++) {
            const cb = checkboxes[i];
            const id = parseInt(cb.dataset.siteId);
            
            if (isChecking) {
                selectedSites.add(id);
                cb.checked = true;
                cb.closest('.site-item').classList.add('selected');
            } else {
                selectedSites.delete(id);
                cb.checked = false;
                cb.closest('.site-item').classList.remove('selected');
            }
        }
    } else {
        // Одиночный selection
        if (checkbox.checked) {
            selectedSites.add(siteId);
            checkbox.closest('.site-item').classList.add('selected');
        } else {
            selectedSites.delete(siteId);
            checkbox.closest('.site-item').classList.remove('selected');
        }
        lastChecked = checkbox;
    }
    
    updateBulkActions();
}

// Обновление панели массовых действий
function updateBulkActions() {
    const bulkActions = document.getElementById('bulk-actions');
    const selectedCount = document.getElementById('selected-count');
    
    selectedCount.textContent = selectedSites.size;
    
    if (selectedSites.size > 0) {
        bulkActions.classList.add('visible');
    } else {
        bulkActions.classList.remove('visible');
    }
}

// Массовое удаление
async function deleteSelectedSites() {
    if (selectedSites.size === 0) return;
    
    try {
        const response = await fetch('/api/sites/bulk-delete', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': 'Bearer ' + currentToken
            },
            body: JSON.stringify({ site_ids: Array.from(selectedSites) })
        });

        if (response.ok) {
            const result = await response.json();
            showToast(`Удалено ${result.success} из ${result.total} сайтов`, 'success');
            selectedSites.clear();
            lastChecked = null;
            loadSites();
        } else {
            showToast('Ошибка при массовом удалении', 'error');
        }
    } catch (error) {
        showToast('Ошибка сети', 'error');
    }
}

// Удаление одиночного сайта
async function deleteSite(siteId) {
    try {
        const response = await fetch(`/api/sites/${siteId}`, {
            method: 'DELETE',
            headers: {
                'Authorization': 'Bearer ' + currentToken
            }
        });

        if (response.ok) {
            showToast('Сайт удален', 'success');
            loadSites();
        } else {
            showToast('Ошибка при удалении сайта', 'error');
        }
    } catch (error) {
        showToast('Ошибка сети', 'error');
    }
}

// Загрузка сайтов
async function loadSites() {
    try {
        const response = await fetch('/api/sites', {
            headers: {
                'Authorization': 'Bearer ' + currentToken
            }
        });

        if (response.ok) {
            const data = await response.json();
            displaySites(data.sites);
        }
    } catch (error) {
        console.error('Error loading sites:', error);
        showToast('Ошибка загрузки сайтов', 'error');
    }
}

// Массовое добавление сайтов
document.getElementById('bulk-add-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const textarea = document.getElementById('bulk-sites');
    const urlsText = textarea.value.trim();
    
    if (!urlsText) {
        showToast('Введите список сайтов', 'warning');
        return;
    }

    const urls = urlsText.split('\n')
        .map(url => url.trim())
        .filter(url => url.length > 0);

    if (urls.length === 0) {
        showToast('Не найдено валидных URL', 'warning');
        return;
    }

    try {
        const response = await fetch('/api/sites/bulk', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': 'Bearer ' + currentToken
            },
            body: JSON.stringify({ urls })
        });

        if (response.ok) {
            const result = await response.json();
            textarea.value = '';
            showToast(`Добавлено ${result.success} из ${result.total} сайтов`, 'success');
            loadSites();
        } else {
            showToast('Ошибка при массовом добавлении', 'error');
        }
    } catch (error) {
        showToast('Ошибка сети', 'error');
    }
});

// Одиночное добавление сайта
document.getElementById('add-site-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const url = document.getElementById('site-url').value;

    try {
        const response = await fetch('/api/sites', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': 'Bearer ' + currentToken
            },
            body: JSON.stringify({ url })
        });

        if (response.ok) {
            document.getElementById('site-url').value = '';
            showToast('Сайт добавлен', 'success');
            loadSites();
        } else {
            showToast('Ошибка при добавлении сайта', 'error');
        }
    } catch (error) {
        showToast('Ошибка сети', 'error');
    }
});