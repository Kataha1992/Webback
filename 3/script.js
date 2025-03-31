const statusBar = document.querySelector('.status-bar');

if (!statusBar.innerHTML.includes('range')) {
    statusBar.classList.remove('hidden');
}
