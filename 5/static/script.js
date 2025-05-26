document.addEventListener('DOMContentLoaded', function() {
    // Валидация форм перед отправкой
    const forms = document.querySelectorAll('form');
    forms.forEach(form => {
        form.addEventListener('submit', function(e) {
            const inputs = this.querySelectorAll('input[required]');
            let isValid = true;
            
            inputs.forEach(input => {
                if (!input.value.trim()) {
                    input.style.borderColor = '#e74c3c';
                    isValid = false;
                } else {
                    input.style.borderColor = '#ddd';
                }
            });
            
            if (!isValid) {
                e.preventDefault();
                alert('Пожалуйста, заполните все обязательные поля!');
            }
        });
    });
    
    // Подсветка полей при фокусе
    const inputs = document.querySelectorAll('input[type="text"], input[type="password"]');
    inputs.forEach(input => {
        input.addEventListener('focus', function() {
            this.style.borderColor = '#3498db';
        });
        
        input.addEventListener('blur', function() {
            this.style.borderColor = '#ddd';
        });
    });
    
    // Показать/скрыть пароль
    const passwordInput = document.getElementById('password');
    if (passwordInput) {
        const togglePassword = document.createElement('span');
        togglePassword.innerHTML = '👁️';
        togglePassword.style.cursor = 'pointer';
        togglePassword.style.marginLeft = '10px';
        togglePassword.title = 'Показать пароль';
        
        togglePassword.addEventListener('click', function() {
            if (passwordInput.type === 'password') {
                passwordInput.type = 'text';
                this.title = 'Скрыть пароль';
            } else {
                passwordInput.type = 'password';
                this.title = 'Показать пароль';
            }
        });
        
        passwordInput.insertAdjacentElement('afterend', togglePassword);
    }
});
