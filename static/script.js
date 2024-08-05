document.addEventListener('DOMContentLoaded', function() {
    const registerForm = document.getElementById('registerForm');
    const loginForm = document.getElementById('loginForm');
    const bookForm = document.getElementById('bookForm');
    const booksTableBody = document.getElementById('booksTable').querySelector('tbody');
    const socket = new WebSocket('ws://localhost:8080/ws');
    const logoutButton = document.getElementById('logoutButton');
    const importForm = document.getElementById('importForm');
    const exportButton = document.getElementById('exportButton');
    const dashboard = document.getElementById('dashboard');
    const authSection = document.getElementById('authSection');

    // WebSocket event handlers
    socket.onopen = function() {
        console.log('WebSocket connection opened');
    };

    socket.onerror = function(error) {
        console.error('WebSocket error:', error);
    };

    socket.onclose = function() {
        console.log('WebSocket connection closed');
    };

    socket.onmessage = function(event) {
        const book = JSON.parse(event.data);
        addBookToTable(book);
    };

    // Register form submission
    registerForm.addEventListener('submit', function(event) {
        event.preventDefault();

        const username = document.getElementById('registerUsername').value;
        const password = document.getElementById('registerPassword').value;

        const user = {
            username: username,
            password: password
        };

        fetch('/register', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(user)
        }).then(response => {
            if (response.ok) {
                alert('Registration successful');
                registerForm.reset();
            } else {
                alert('Failed to register');
            }
        });
    });

    // Login form submission
    loginForm.addEventListener('submit', function(event) {
        event.preventDefault();

        const username = document.getElementById('loginUsername').value;
        const password = document.getElementById('loginPassword').value;

        const credentials = {
            username: username,
            password: password
        };

        fetch('/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(credentials)
        }).then(response => {
            if (response.ok) {
                alert('Login successful');
                loginForm.reset();
                authSection.style.display = 'none';
                dashboard.style.display = 'block';
                logoutButton.style.display = 'block';
                loadBooks();
            } else {
                alert('Failed to login');
            }
        });
    });

    // Logout button click
    logoutButton.addEventListener('click', function() {
        fetch('/logout', {
            method: 'POST'
        }).then(response => {
            if (response.ok) {
                alert('Logout successful');
                logoutButton.style.display = 'none';
                authSection.style.display = 'block';
                dashboard.style.display = 'none';
                booksTableBody.innerHTML = '';
            } else {
                alert('Failed to logout');
            }
        });
    });

    // Add book to table
    function addBookToTable(book) {
        const row = booksTableBody.insertRow();
        row.innerHTML = `
            <td>${book.id}</td>
            <td>${book.title}</td>
            <td>${book.author}</td>
            <td>${book.published_date}</td>
            <td>${book.isbn}</td>
            <td>${book.categories}</td>
            <td>${book.rating}</td>
            <td>
                <button onclick="deleteBook(${book.id})">Delete</button>
            </td>
        `;
    }

    // Book form submission
    bookForm.addEventListener('submit', function(event) {
        event.preventDefault();

        const title = document.getElementById('title').value;
        const author = document.getElementById('author').value;
        const publishedDate = document.getElementById('publishedDate').value;
        const isbn = document.getElementById('isbn').value;
        const categories = document.getElementById('categories').value;
        const rating = document.getElementById('rating').value;

        const book = {
            title: title,
            author: author,
            published_date: publishedDate,
            isbn: isbn,
            categories: categories,
            rating: rating
        };

        fetch('/api/books', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(book)
        }).then(response => {
            if (response.status === 201) {
                loadBooks();
                bookForm.reset();
            } else {
                alert('Failed to add book');
            }
        });
    });

    // Load books
    function loadBooks() {
        fetch('/api/books')
            .then(response => response.json())
            .then(books => {
                booksTableBody.innerHTML = '';
                books.forEach(book => {
                    addBookToTable(book);
                });
            });
    }

    // Delete book
    window.deleteBook = function(id) {
        fetch(`/api/books/${id}`, {
            method: 'DELETE'
        }).then(response => {
            if (response.status === 200) {
                loadBooks();
            } else {
                alert('Failed to delete book');
            }
        });
    };

    // Import form submission
    importForm.addEventListener('submit', function(event) {
        event.preventDefault();

        const importFile = document.getElementById('importFile').files[0];
        const formData = new FormData();
        formData.append('file', importFile);

        fetch('/api/import-books', {
            method: 'POST',
            body: formData
        }).then(response => {
            if (response.status === 200) {
                alert('Books imported successfully');
                loadBooks();
            } else {
                alert('Failed to import books');
            }
        });
    });

    // Export books
    exportButton.addEventListener('click', function() {
        fetch('/api/export-books')
            .then(response => response.blob())
            .then(blob => {
                const url = window.URL.createObjectURL(blob);
                const a = document.createElement('a');
                a.href = url;
                a.download = 'books.json';
                document.body.appendChild(a);
                a.click();
                a.remove();
            });
    });

    // Initially load books if on the dashboard
    if (window.location.pathname === '/dashboard') {
        loadBooks();
    }
});
