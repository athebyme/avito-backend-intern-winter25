CREATE TABLE merch(
    id Serial PRIMARY KEY,

    --на самом деле как будто бы можно не unique сделать, но оставил для автозаполнения
    name VARCHAR(255) UNIQUE NOT NULL,
    price INT
);