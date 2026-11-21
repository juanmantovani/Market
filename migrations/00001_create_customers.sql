-- +goose Up
CREATE TABLE customers (
    id text PRIMARY KEY,
    name text NOT NULL,
    document_type text NOT NULL CHECK (document_type IN ('none', 'dni', 'cuit', 'cuil')),
    document_number text NOT NULL DEFAULT '',
    tax_condition text NOT NULL CHECK (tax_condition IN ('final_consumer', 'monotax', 'registered', 'exempt')),
    email text,
    phone text,
    address text,
    notes text,
    is_system boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

-- A customer without a document (document_type 'none') never collides; the
-- partial index mirrors domain.Customer.DocumentKey.
CREATE UNIQUE INDEX customers_document_key ON customers (document_type, document_number)
    WHERE document_type <> 'none';

INSERT INTO customers (id, name, document_type, tax_condition, is_system, created_at, updated_at)
VALUES ('00000000-0000-7000-8000-000000000001', 'Consumidor Final', 'none', 'final_consumer', true, now(), now());

-- +goose Down
DROP TABLE customers;
