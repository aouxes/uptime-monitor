// Добавление сайта
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
            loadSites();
        } else {
            alert('Ошибка при добавлении сайта');
        }
    } catch (error) {
        alert('Ошибка сети');
    }
});

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
    }
}

// Отображение сайтов с кнопкой удаления
function displaySites(sites) {
    const container = document.getElementById('sites-container');
    
    if (sites.length === 0) {
        container.innerHTML = '<p>Нет добавленных сайтов</p>';
        return;
    }

    container.innerHTML = sites.map(site => `
        <div class="site-item ${site.last_status === 'DOWN' ? 'down' : 'up'}">
            <div class="site-info">
                <strong>${site.url}</strong>
                <div class="site-status ${site.last_status?.toLowerCase() || 'unknown'}">
                    Статус: ${site.last_status || 'UNKNOWN'}
                </div>
                <small>Последняя проверка: ${new Date(site.last_checked).toLocaleString()}</small>
            </div>
            <div class="site-actions">
                <button class="delete-btn" onclick="deleteSite(${site.id})">Удалить</button>
            </div>
        </div>
    `).join('');
}

// Удаление сайта
async function deleteSite(siteId) {
    if (!confirm('Удалить этот сайт из мониторинга?')) {
        return;
    }

    try {
        const response = await fetch(`/api/sites/${siteId}`, {
            method: 'DELETE',
            headers: {
                'Authorization': 'Bearer ' + currentToken
            }
        });

        if (response.ok) {
            alert('Сайт удален');
            loadSites(); // Перезагружаем список
        } else {
            alert('Ошибка при удалении сайта');
        }
    } catch (error) {
        alert('Ошибка сети');
    }
}

if (currentToken) {
    loadSites();
}

document.getElementById('bulk-add-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const textarea = document.getElementById('bulk-sites');
    const urlsText = textarea.value.trim();
    
    if (!urlsText) {
        alert('Введите список сайтов');
        return;
    }

    const urls = urlsText.split('\n')
        .map(url => url.trim())
        .filter(url => url.length > 0);

    if (urls.length === 0) {
        alert('Не найдено валидных URL');
        return;
    }

    if (urls.length > 50) {
        alert('Максимум 50 сайтов за раз');
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
            alert(`Добавлено ${result.success} из ${result.total} сайтов`);
            loadSites(); // Обновляем список
        } else {
            alert('Ошибка при массовом добавлении сайтов');
        }
    } catch (error) {
        alert('Ошибка сети');
    }
});