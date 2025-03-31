const statusBar = document.querySelector('.status-bar');

if (!statusBar.innerHTML.includes('range')) {
    statusBar.classList.remove('hidden');

    notifications = Array.from(document.querySelectorAll('.notification'));

    notifications.forEach((elem, index) => {
        setTimeout(() => {
            elem.classList.add('show');
        }, 250 * (index + 1));
    });

    setTimeout(() => {
        notifications.reverse().forEach((elem, index) => {
            setInterval(() => {
                elem.classList.remove('show');
                elem.classList.add('hide');
            }, 250 * (index + 1));
        });

        setTimeout(() => {
            statusBar.remove();
        }, 2250)
    }, 8000);
}
