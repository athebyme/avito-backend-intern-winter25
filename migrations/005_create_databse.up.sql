SELECT 'CREATE DATABASE shop'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'shop')