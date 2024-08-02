document.addEventListener('DOMContentLoaded', function() {
    const registerForm = document.getElementById('registerForm');
    const loginForm = document.getElementById(`loginForm`);
    const bookForm = document.getElementById('bookForm');
    const booksTableBody = document.getElementById('booksTable').querySelector('tbody');
    const socket = new WebSocket('ws://localhost:8080/ws');

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

    loginForm.addEventListener('submit', function(event) {
        event.preventDefault();

        const username = document.getElementById('loginUsername').value;
        const password = document.getElementById('loginPassword').value;

        const credentials= {
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
                window.location.href = '/dashboard'; // Redirect or update UI
            } else {
                alert('Failed to login');
            }
        });
    });

    socket.onmessage = function(event) {
        const book = JSON.parse(event.data);
        addBookToTable(book);
    };

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

    function loadBooks() {
        fetch('/api/books')
            .then(response => response.json())
            .then(books => {
                booksTableBody.innerHTML = '';
                books.forEach(book => {
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
                });
            });
    }

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

    loadBooks();
});
