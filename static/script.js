document.addEventListener('DOMContentLoaded', function() {
    const bookForm = document.getElementById('bookForm');
    const booksTableBody = document.getElementById('booksTable').querySelector('tbody');
    const socket = new WebSocket('ws://localhosr:8080/ws');

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

        const book = {
            title: title,
            author: author,
            published_date: publishedDate,
            isbn: isbn
        };

        fetch('/books', {
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
        fetch('/books')
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
                        <td>
                            <button onclick="deleteBook(${book.id})">Delete</button>
                        </td>
                    `;
                });
            });
    }

    window.deleteBook = function(id) {
        fetch(`/books/${id}`, {
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
