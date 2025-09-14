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
            loadSites();
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

// Проверяем авторизацию при загрузке
if (currentToken) {
    showPage('dashboard-page');
    loadSites();
} else {
    showPage('login-page');
}