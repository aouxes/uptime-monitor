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

let currentToken = localStorage.getItem('token');

function showPage(pageId) {
    document.querySelectorAll('.page').forEach(page => {
        page.classList.remove('active');
    });
    document.getElementById(pageId).classList.add('active');
}

// Регистрация
document.getElementById('register-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const formData = {
        username: document.getElementById('register-username').value,
        email: document.getElementById('register-email').value,
        password: document.getElementById('register-password').value
    };

    try {
        const response = await fetch('/api/register', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(formData)
        });

        if (response.ok) {
            showToast('Регистрация успешна! Теперь войдите.', 'success');
            showPage('login-page');
        } else {
            const error = await response.json();
            showToast('Ошибка регистрации: ' + (error.details ? JSON.stringify(error.details) : 'Unknown error'), 'error');
        }
    } catch (error) {
        showToast('Ошибка сети', 'error');
    }
});

// Логин
document.getElementById('login-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const formData = {
        username: document.getElementById('login-username').value,
        password: document.getElementById('login-password').value
    };

    try {
        const response = await fetch('/api/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(formData)
        });

        if (response.ok) {
            const data = await response.json();
            currentToken = data.token;
            localStorage.setItem('token', data.token);
            showToast('Вход выполнен успешно', 'success');
            showPage('dashboard-page');
            if (typeof loadSites === 'function') {
                loadSites();
            }
        } else {
            showToast('Неверный логин или пароль', 'error');
        }
    } catch (error) {
        showToast('Ошибка сети', 'error');
    }
});

// Выход
function logout() {
    currentToken = null;
    localStorage.removeItem('token');
    showPage('login-page');
}

// Проверка токена при загрузке
async function checkAuth() {
    console.log('checkAuth called');
    const token = localStorage.getItem('token');
    console.log('Token from localStorage:', token ? token.substring(0, 20) + '...' : 'null');
    
    if (!token) {
        console.log('No token found, showing login page');
        showPage('login-page');
        return;
    }

    try {
        console.log('Verifying token with server...');
        const response = await fetch('/api/verify-token', {
            headers: {
                'Authorization': 'Bearer ' + token
            }
        });

        console.log('Token verification response:', response.status);
        
        if (response.ok) {
            // Токен валидный, показываем дашборд
            console.log('Token is valid, showing dashboard');
            currentToken = token;
            showPage('dashboard-page');
            if (typeof loadSites === 'function') {
                loadSites();
            }
            showToast('Добро пожаловать!', 'success');
        } else {
            // Токен невалидный, очищаем и показываем логин
            console.log('Token is invalid, clearing and showing login');
            localStorage.removeItem('token');
            showPage('login-page');
        }
    } catch (error) {
        console.error('Auth check failed:', error);
        localStorage.removeItem('token');
        showPage('login-page');
    }
}

// Проверяем авторизацию при загрузке
checkAuth();

async function loadSites() {
    try {
        const response = await fetchWithAuth('/api/sites');

        if (response.ok) {
            const data = await response.json();
            displaySites(data.sites);
        }
    } catch (error) {
        console.error('Error loading sites:', error);
        showToast('Ошибка загрузки сайтов', 'error');
    }
}