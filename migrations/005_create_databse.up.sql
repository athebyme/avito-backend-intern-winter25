DO $$
BEGIN 
    IF NOT EXISTS (SELECT FROM pg_database WHERE datname = 'shop') THEN 
        CREATE DATABASE shop;
END IF;
END $$;
