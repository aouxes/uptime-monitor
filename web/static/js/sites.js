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

// Отображение сайтов
function displaySites(sites) {
    const container = document.getElementById('sites-container');
    
    if (sites.length === 0) {
        container.innerHTML = '<p>Нет добавленных сайтов</p>';
        return;
    }

    container.innerHTML = sites.map(site => `
        <div class="site-item ${site.last_status === 'DOWN' ? 'down' : 'up'}">
            <div>
                <strong>${site.url}</strong>
                <div class="site-status ${site.last_status?.toLowerCase() || 'unknown'}">
                    Статус: ${site.last_status || 'UNKNOWN'}
                </div>
            </div>
            <div>${new Date(site.last_checked).toLocaleString()}</div>
        </div>
    `).join('');
}