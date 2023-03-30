[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# Key-Value List

## Introduction

`key-value list` is a simple HTTP API that allows you to manage lists of pages with associated articles. Each page can have multiple articles, and each article has a title, author, and content.

This project is based on the idea from [here](https://medium.com/dcardlab/de07f45295f6).

## Usage
The API provides the following endpoints:

- `GET /list/head?list_id=<list_id>`: Retrieves the next page ID for the specified list ID.

- `GET /page/get?page_id=<page_id>`: Retrieves the articles and the next page ID for the specified page ID.

- `POST /page/set`: Adds a new article to a page. The request body should be a JSON object with the following fields:
    - title (required): The title of the article.
    - author (required): The author of the article.
    - content (required): The content of the article.
    ```json
    {
        "title": "Article sample",
        "author": "Author sample",
        "content": "Content sample"
    }
    ```
- `PUT /page/update?page_id=<page_id>`: Updates the articles for the specified page ID. The request body can be either a single article or an array of articles, each with the following fields:
    - title (required): The title of the article.
    - author (required): The author of the article.
    - content (required): The content of the article.
- `DELETE /page/delete?list_id=<list_id>`: Deletes all pages and articles for the specified list ID.

## Setup
1. Clone the repository:
```bash
git clone https://github.com/ericlinsechs/key-value-list.git
cd key-value-list
```

2. Run the following command to start the Docker containers:
```bash
docker-compose up
```
This command will start two containers:
- `db`: A PostgreSQL container that will be used to store the data.
- `my-app`: The Go application container.

The containers will be connected to a bridge network called `postgresNetwork`.

3. The database will be automatically created and initialized using the init.sql script, which is mounted as a volume in the db container.

The `init.sql` script creates a database named `my_database` and three tables: 
- `lists`
- `pages`
- `articles`

It also inserts a record into the lists table with id set to 1 and next_page_id set to 1.

## RDBMS vs NoSQL database
A relational database management system (RDBMS) like PostgreSQL is known for its ability to handle complex queries and transactions, while providing strong data consistency and reliability.

On the other hand, a document-oriented NoSQL database like MongoDB is designed for flexibility and scalability. It is optimized for storing and querying unstructured or semi-structured data, and it provides a flexible schema that can adapt to changing data requirements. However, this project's data structure has 

In this project, the use of PostgreSQL as the database management system makes more sense because the project requires the management of relational data, specifically lists of pages, each with a set of related articles. Besides, it provides ACID-compliant transactions, which ensure data consistency, and has strong data integrity features, which can help prevent data corruption.

## License
Key-Value List is released under the MIT License. See [LICENSE](LICENSE) file for details.

