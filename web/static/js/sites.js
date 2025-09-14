async function fetchWithAuth(url, options = {}) {
    const response = await fetch(url, {
        ...options,
        headers: {
            ...options.headers,
            'Authorization': 'Bearer ' + currentToken
        }
    });

    if (response.status === 401) {
        localStorage.removeItem('token');
        showToast('Сессия истекла', 'error');
        showPage('login-page');
        throw new Error('Authentication failed');
    }

    return response;
}

let allSites = [];
let lastChecked = null;
let selectedSites = new Set();

// Функция "Выбрать все"
function toggleSelectAll(selectAllCheckbox) {
    const isChecked = selectAllCheckbox.checked;
    
    if (isChecked) {
        // Выбираем все сайты
        selectedSites = new Set(allSites.map(site => site.id));
    } else {
        // Снимаем выделение со всех сайтов
        selectedSites.clear();
    }
    
    // Обновляем чекбоксы и стили
    document.querySelectorAll('.site-checkbox').forEach(cb => {
        cb.checked = isChecked;
    });
    
    document.querySelectorAll('.site-item').forEach(item => {
        item.classList.toggle('selected', isChecked);
    });
    
    updateBulkActions();
    updateSelectAllCheckbox();
}

// Обновление состояния чекбокса "Выбрать все"
function updateSelectAllCheckbox() {
    const selectAllCheckbox = document.getElementById('select-all-checkbox');
    if (!selectAllCheckbox) return;
    
    if (selectedSites.size === 0) {
        selectAllCheckbox.checked = false;
        selectAllCheckbox.indeterminate = false;
    } else if (selectedSites.size === allSites.length) {
        selectAllCheckbox.checked = true;
        selectAllCheckbox.indeterminate = false;
    } else {
        selectAllCheckbox.checked = false;
        selectAllCheckbox.indeterminate = true;
    }
}

// Отображение сайтов
function displaySites(sites) {
    allSites = sites;
    const container = document.getElementById('sites-container');
    
    if (sites.length === 0) {
        container.innerHTML = '<div class="no-sites">Нет добавленных сайтов</div>';
        document.getElementById('bulk-actions').classList.remove('visible');
        document.getElementById('select-all-checkbox').disabled = true;
        return;
    }

    document.getElementById('select-all-checkbox').disabled = false;

    container.innerHTML = sites.map(site => `
        <div class="site-item ${selectedSites.has(site.id) ? 'selected' : ''}" 
             onclick="handleSiteClick(${site.id}, event)">
            <input 
                type="checkbox" 
                class="site-checkbox" 
                data-site-id="${site.id}"
                ${selectedSites.has(site.id) ? 'checked' : ''}
            >
            <div class="site-info">
                <div class="site-url">${site.url}</div>
                <div class="site-status ${site.last_status?.toLowerCase() || 'unknown'}">
                    ${site.last_status || 'UNKNOWN'} • ${new Date(site.last_checked).toLocaleString()}
                </div>
            </div>
            <div class="site-actions">
                <button class="danger-btn" onclick="deleteSite(${site.id}, event)">Удалить</button>
            </div>
        </div>
    `).join('');

    updateBulkActions();
    updateSelectAllCheckbox();
}

function handleSiteClick(siteId, event) {
    // Игнорируем клики по кнопкам внутри строки
    if (event.target.tagName === 'BUTTON' || event.target.closest('button')) {
        return;
    }
    
    const checkbox = document.querySelector(`.site-checkbox[data-site-id="${siteId}"]`);
    if (!checkbox) return;
    
    // Эмулируем изменение чекбокса
    checkbox.checked = !checkbox.checked;
    
    // Создаем искусственное событие для обработки
    const fakeEvent = {
        shiftKey: event.shiftKey,
        target: checkbox
    };
    
    handleCheckboxChange(checkbox, fakeEvent);
}

// Обработчик чекбоксов с поддержкой Shift
function handleCheckboxChange(checkbox, event) {
    const siteId = parseInt(checkbox.dataset.siteId);
    const siteItem = checkbox.closest('.site-item');
    
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
            siteItem.classList.add('selected');
        } else {
            selectedSites.delete(siteId);
            siteItem.classList.remove('selected');
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
    
    updateSelectAllCheckbox();
}

// Массовое удаление ВСЕХ выбранных сайтов
async function deleteSelectedSites() {
    if (selectedSites.size === 0) return;
    
    try {
        const siteIds = Array.from(selectedSites);
        const response = await fetchWithAuth('/api/sites/bulk-delete', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ site_ids: siteIds })
        });

        if (response.ok) {
            const result = await response.json();
            showToast(`Удалено ${result.success} из ${result.total} сайтов`, 'success');
            
            // Сбрасываем selection
            selectedSites.clear();
            lastChecked = null;
            
            // Перезагружаем список
            loadSites();
        } else {
            showToast('Ошибка при массовом удалении', 'error');
        }
    } catch (error) {
        if (error.message !== 'Authentication failed') {
            showToast('Ошибка сети', 'error');
        }
    }
}

// Удаление одиночного сайта
async function deleteSite(siteId) {
    event.stopPropagation();
    
    if (!confirm('Удалить этот сайт из мониторинга?')) {
        return;
    }

    try {
        const response = await fetchWithAuth(`/api/sites/${siteId}`, {
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
        const response = await fetchWithAuth('/api/sites');
        
        if (response.ok) {
            const data = await response.json();
            displaySites(data.sites);
        }
    } catch (error) {
        if (error.message !== 'Authentication failed') {
            showToast('Ошибка загрузки сайтов', 'error');
        }
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
        const response = await fetchWithAuth('/api/sites/bulk', {
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
        const response = await fetchWithAuth('/api/sites', {
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